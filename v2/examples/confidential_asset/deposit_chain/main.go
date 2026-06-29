//go:build cgo

// # Prerequisites
//
// 1. Download the pre-compiled FFI static library (no Rust toolchain needed):
//
//	go run github.com/aptos-labs/confidential-asset-bindings/bindings/go/aptosconfidential/tools/download@v1.1.2
//
// 2. Build / run with the library path (replace <triple> with your platform):
//
//	macOS M-series:  CGO_LDFLAGS="-L$(pwd)/native/aarch64-apple-darwin"
//	Linux amd64:     CGO_LDFLAGS="-L$(pwd)/native/x86_64-unknown-linux-gnu"
//	Linux arm64:     CGO_LDFLAGS="-L$(pwd)/native/aarch64-unknown-linux-gnu"
//	Windows amd64:   set CGO_LDFLAGS=-L%cd%\native\x86_64-pc-windows-msvc
//
//	CGO_LDFLAGS="-L$(pwd)/native/<triple>" go run .
//
// # Environment variables
//
//	APTOS_PRIVATE_KEY       Ed25519 account private key (required)
//	TWISTED_PRIVATE_KEY_HEX Decryption key hex; derived from Ed25519 key when omitted (optional)
//	APTOS_NETWORK           testnet or mainnet (required)
//
// deposit_chain runs deposit → normalize (if needed) → rollover_pending_balance → prints confidential balance.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	confidentialasset "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/native"
)

const (
	tokenMetadataLong   = "0x000000000000000000000000000000000000000000000000000000000000000a"
	defaultDepositOctas = uint64(800_000)
	minPublicFAOctasGas = uint64(1_000_000) // 0.01 APT on FA 0xa for gas headroom
)

func networkFromEnv() aptos.NetworkConfig {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APTOS_NETWORK"))) {
	case "mainnet":
		return aptos.Mainnet
	case "devnet":
		return aptos.Devnet
	case "localnet", "local":
		return aptos.Localnet
	default:
		return aptos.Testnet
	}
}

func restBaseURL(net aptos.NetworkConfig) string {
	if u := strings.TrimSpace(os.Getenv("APTOS_NODE_URL")); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return strings.TrimSuffix(net.NodeURL, "/")
}

func loadAccount() (*account.Account, error) {
	pk := strings.TrimSpace(os.Getenv("APTOS_PRIVATE_KEY"))
	if pk == "" {
		pk = strings.TrimSpace(os.Getenv("FIXED_ED25519_PRIVATE_KEY"))
	}
	if pk == "" {
		return nil, errors.New("set APTOS_PRIVATE_KEY or FIXED_ED25519_PRIVATE_KEY")
	}
	if len(pk) > 12 && pk[:13] == "ed25519-priv-" {
		return account.FromAIP80(pk)
	}
	return account.FromPrivateKeyHex(pk)
}

func newClientPair() (aptos.Client, *confidentialasset.Client, aptos.NetworkConfig, error) {
	net := networkFromEnv()
	client, err := aptos.NewClient(net)
	if err != nil {
		return nil, nil, net, err
	}
	cc := confidentialasset.NewClient(client, confidentialasset.WithRESTBaseURL(restBaseURL(net)))
	return client, cc, net, nil
}

func defaultAPTToken() aptos.AccountAddress {
	return aptos.MustParseAddress(tokenMetadataLong)
}

func envTruthy(name string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	return v == "1" || v == "true" || v == "yes"
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()

	_, cc, net, err := newClientPair()
	if err != nil {
		log.Fatalf("client: %v", err)
	}
	nc := native.Wrap(cc)
	fmt.Printf("deposit_chain — Network: %s\n\n", net.Name)

	acct, err := loadAccount()
	if err != nil {
		log.Fatalf("account: %v", err)
	}
	fmt.Printf("Address: %s\n\n", acct.Address().String())

	pubOctas, err := cc.FetchPublicFABalanceOctas(ctx, acct.Address(), tokenMetadataLong)
	if err != nil {
		log.Fatalf("public FA APT balance: %v", err)
	}
	if pubOctas < minPublicFAOctasGas {
		log.Fatalf("public APT (FA 0xa) octas=%d — need at least %d (0.01 APT) for gas; fund this address on %s then re-run",
			pubOctas, minPublicFAOctasGas, net.Name)
	}
	fmt.Printf("Public APT (FA 0xa) octas: %d\n", pubOctas)

	token := defaultAPTToken()
	registered, err := cc.HasUserRegistered(ctx, acct.Address(), token)
	if err != nil {
		log.Fatalf("has_confidential_store: %v", err)
	}
	if !registered {
		log.Fatalf("No confidential store for FA 0xa — register first:\n  CGO_ENABLED=1 go run ./examples/confidential_asset/register")
	}

	fmt.Printf("\nDepositing %d octas into confidential balance…\n", defaultDepositOctas)
	if _, err := cc.Deposit(ctx, acct, token, defaultDepositOctas, tokenMetadataLong); err != nil {
		log.Fatalf("deposit: %v", err)
	}
	fmt.Println("deposit: ok")

	norm, err := cc.IsBalanceNormalized(ctx, acct.Address(), token)
	if err != nil {
		log.Fatalf("is_normalized: %v", err)
	}
	if !norm {
		if envTruthy("ATTEMPT_ROLLOVER_IF_NOT_NORMALIZED") {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "is_normalized == false — ATTEMPT_ROLLOVER_IF_NOT_NORMALIZED=1: skipping normalize_raw, attempting rollover anyway…")
		} else {
			fmt.Println("\nBalance not normalized after deposit — submitting normalize_raw…")
			twistedHex := strings.TrimSpace(os.Getenv("TWISTED_PRIVATE_KEY_HEX"))
			if _, err := nc.NormalizeBalance(ctx, acct, token, twistedHex, tokenMetadataLong); err != nil {
				log.Fatalf("normalize_raw: %v", err)
			}
			fmt.Println("normalize_raw: ok")
			norm, err = cc.IsBalanceNormalized(ctx, acct.Address(), token)
			if err != nil {
				log.Fatalf("is_normalized (after normalize): %v", err)
			}
			if !norm {
				log.Fatalf("is_normalized still false after normalize_raw — investigate on-chain state or proof mismatch")
			}
		}
	}

	fmt.Println("\nRolling over pending balance…")
	if _, err := cc.RolloverPendingBalance(ctx, acct, token, false, tokenMetadataLong); err != nil {
		log.Fatalf("rollover_pending_balance: %v", err)
	}
	fmt.Println("rollover_pending_balance: ok")

	if strings.TrimSpace(os.Getenv("SKIP_CONFIDENTIAL_BALANCE_READ")) == "1" {
		fmt.Println("\nSkipped balance print: SKIP_CONFIDENTIAL_BALANCE_READ=1")
	} else {
		fmt.Println("\n=== Confidential balance (SDK GetBalance) ===")
		twistedHex := strings.TrimSpace(os.Getenv("TWISTED_PRIVATE_KEY_HEX"))
		bal, err := nc.GetBalance(ctx, acct, token, twistedHex)
		if err != nil {
			log.Fatalf("get balance: %v", err)
		}
		fmt.Printf("available (octas): %d\n", bal.AvailableOctas)
		fmt.Printf("pending (octas): %d\n", bal.PendingOctas)
	}

	fmt.Println("\nDone.")
}
