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
// Registers confidential store when needed, then optionally deposits into confidential balance.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	confidentialasset "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
)

const (
	tokenMetadataLong   = "0x000000000000000000000000000000000000000000000000000000000000000a"
	defaultDepositOctas = uint64(800_000)
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

func envSkipConfidentialDeposit() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SKIP_CONFIDENTIAL_DEPOSIT")))
	return v == "1" || v == "true" || v == "yes"
}

func depositOctasFromEnv() (uint64, error) {
	s := strings.TrimSpace(os.Getenv("CONFIDENTIAL_DEPOSIT_OCTAS"))
	if s == "" {
		return defaultDepositOctas, nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("CONFIDENTIAL_DEPOSIT_OCTAS: %w", err)
	}
	return n, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	_, cc, net, err := newClientPair()
	if err != nil {
		log.Fatalf("client: %v", err)
	}
	fmt.Printf("register — Network: %s\n\n", net.Name)

	acct, err := loadAccount()
	if err != nil {
		log.Fatalf("account: %v", err)
	}
	token := defaultAPTToken()
	twistedHex := strings.TrimSpace(os.Getenv("TWISTED_PRIVATE_KEY_HEX"))

	fmt.Printf("Address: %s\n", acct.Address().String())
	pubOctas, err := cc.FetchPublicFABalanceOctas(ctx, acct.Address(), tokenMetadataLong)
	if err != nil {
		log.Fatalf("public FA APT balance: %v", err)
	}
	fmt.Printf("Public APT (FA 0xa) octas: %d (need some for gas)\n", pubOctas)

	registered, err := cc.HasUserRegistered(ctx, acct.Address(), token)
	if err != nil {
		log.Fatalf("has_confidential_store: %v", err)
	}

	if !registered {
		tx, err := cc.RegisterBalance(ctx, acct, token, twistedHex, tokenMetadataLong)
		if err != nil {
			log.Fatalf("register_raw: %v", err)
		}
		fmt.Printf("register_raw: success version=%d hash=%s\n", tx.Version, tx.Hash)
	} else {
		fmt.Println("Confidential store already exists for FA 0xa — skipping register_raw.")
	}

	if envSkipConfidentialDeposit() {
		fmt.Println("SKIP_CONFIDENTIAL_DEPOSIT=1 — skipping deposit.")
		return
	}

	depOctas, err := depositOctasFromEnv()
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("\nDepositing %d octas into confidential balance (confidential_asset::deposit)…\n", depOctas)
	dtx, err := cc.Deposit(ctx, acct, token, depOctas, tokenMetadataLong)
	if err != nil {
		log.Fatalf("deposit: %v", err)
	}
	fmt.Printf("deposit: success version=%d hash=%s\n", dtx.Version, dtx.Hash)
}
