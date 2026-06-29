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
//	RECIPIENT_ADDRESS       Recipient account address (required)
//	TRANSFER_AMOUNT_OCTAS   Amount to transfer in octas (required)
//
// Submits confidential_transfer_raw (TS confidentialAsset.transfer).
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
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/native"
)

const tokenMetadataLong = "0x000000000000000000000000000000000000000000000000000000000000000a"

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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	toHex := strings.TrimSpace(os.Getenv("APTOS_SEND_TO"))
	if toHex == "" {
		log.Fatal("set APTOS_SEND_TO (recipient account address)")
	}
	recipient, err := aptos.ParseAddress(toHex)
	if err != nil {
		log.Fatalf("APTOS_SEND_TO: %v", err)
	}

	amount := uint64(1)
	if s := strings.TrimSpace(os.Getenv("APTOS_SEND_OCTAS")); s != "" {
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || n == 0 {
			log.Fatalf("APTOS_SEND_OCTAS must be a positive integer: %q", s)
		}
		amount = n
	}

	_, cc, net, err := newClientPair()
	if err != nil {
		log.Fatalf("client: %v", err)
	}
	acct, err := loadAccount()
	if err != nil {
		log.Fatalf("account: %v", err)
	}

	token := defaultAPTToken()
	twistedHex := strings.TrimSpace(os.Getenv("TWISTED_PRIVATE_KEY_HEX"))

	fmt.Printf("transfer — Network: %s\n", net.Name)
	fmt.Printf("Same entry as TS confidentialAsset.transfer → confidential_transfer_raw\n\n")
	fmt.Printf("From:     %s\n", acct.Address().String())
	fmt.Printf("To:       %s\n", recipient.String())
	fmt.Printf("Token:    %s (FA metadata / confidential asset)\n", token.String())
	fmt.Printf("Amount:   %d octas (from available balance)\n\n", amount)

	okStore, err := cc.HasUserRegistered(ctx, acct.Address(), token)
	if err != nil {
		log.Fatalf("has_confidential_store (sender): %v", err)
	}
	if !okStore {
		log.Fatalf("sender has no confidential store for this token — run: CGO_ENABLED=1 go run ./examples/confidential_asset/register")
	}
	okRecv, err := cc.HasUserRegistered(ctx, recipient, token)
	if err != nil {
		log.Fatalf("has_confidential_store (recipient): %v", err)
	}
	if !okRecv {
		log.Fatalf("recipient has no confidential store for this token — they must register before receiving a confidential transfer")
	}

	nc := native.Wrap(cc)
	tx, err := nc.Transfer(ctx, acct, token, amount, recipient, twistedHex, tokenMetadataLong)
	if err != nil {
		log.Fatalf("confidential_transfer_raw: %v", err)
	}
	fmt.Printf("confidential_transfer_raw: ok version=%d hash=%s\n", tx.Version, tx.Hash)
}
