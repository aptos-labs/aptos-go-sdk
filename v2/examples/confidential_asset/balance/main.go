//go:build cgo

// Prints public FA APT balance and decrypted confidential available/pending (TS getBalance).
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, cc, net, err := newClientPair()
	if err != nil {
		log.Fatalf("client: %v", err)
	}
	fmt.Printf("Network: %s\n\n", net.Name)

	acct, err := loadAccount()
	if err != nil {
		log.Fatalf("account: %v", err)
	}
	token := defaultAPTToken()

	fmt.Printf("Address: %s\n", acct.Address().String())
	pubOctas, err := cc.FetchPublicFABalanceOctas(ctx, acct.Address(), tokenMetadataLong)
	if err != nil {
		log.Fatalf("public FA APT balance: %v", err)
	}
	fmt.Printf("Public APT (FA 0xa) octas: %d\n", pubOctas)

	registered, err := cc.HasUserRegistered(ctx, acct.Address(), token)
	if err != nil {
		log.Fatalf("has_confidential_store: %v", err)
	}
	if !registered {
		fmt.Fprintln(os.Stderr, "No confidential store for 0xa — nothing to decrypt. Register with:\n  CGO_ENABLED=1 go run ./examples/confidential_asset/register")
		os.Exit(1)
	}

	if strings.TrimSpace(os.Getenv("SKIP_CONFIDENTIAL_BALANCE_READ")) == "1" {
		fmt.Println("\nSkipped: SKIP_CONFIDENTIAL_BALANCE_READ=1")
		return
	}

	fmt.Println("\n=== Confidential balance (SDK GetBalance) ===")
	twistedHex := strings.TrimSpace(os.Getenv("TWISTED_PRIVATE_KEY_HEX"))
	bal, err := cc.GetBalance(ctx, acct, token, twistedHex)
	if err != nil {
		log.Fatalf("get balance: %v", err)
	}
	fmt.Printf("available (octas): %d\n", bal.AvailableOctas)
	fmt.Printf("pending (octas): %d\n", bal.PendingOctas)
}
