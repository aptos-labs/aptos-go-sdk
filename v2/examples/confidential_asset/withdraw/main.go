//go:build cgo

// Submits withdraw_to_raw (TS confidentialAsset.withdraw): confidential → public FA for recipient.
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

	octasStr := strings.TrimSpace(os.Getenv("APTOS_WITHDRAW_OCTAS"))
	if octasStr == "" {
		log.Fatal("set APTOS_WITHDRAW_OCTAS (positive integer, from available confidential balance)")
	}
	amount, err := strconv.ParseUint(octasStr, 10, 64)
	if err != nil || amount == 0 {
		log.Fatalf("APTOS_WITHDRAW_OCTAS must be a positive integer: %q", octasStr)
	}

	var recipient aptos.AccountAddress
	if toHex := strings.TrimSpace(os.Getenv("APTOS_WITHDRAW_TO")); toHex != "" {
		recipient, err = aptos.ParseAddress(toHex)
		if err != nil {
			log.Fatalf("APTOS_WITHDRAW_TO: %v", err)
		}
	}
	// If recipient is zero, Client.Withdraw uses signer address (same as TS default).

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

	fmt.Printf("withdraw — Network: %s\n", net.Name)
	fmt.Printf("Same entry as TS confidentialAsset.withdraw → withdraw_to_raw\n\n")
	fmt.Printf("Signer:   %s\n", acct.Address().String())
	if recipient == (aptos.AccountAddress{}) {
		fmt.Printf("Recipient (public FA): %s (default: signer)\n", acct.Address().String())
	} else {
		fmt.Printf("Recipient (public FA): %s\n", recipient.String())
	}
	fmt.Printf("Token:    %s\n", token.String())
	fmt.Printf("Amount:   %d octas (from available confidential balance)\n\n", amount)

	ok, err := cc.HasUserRegistered(ctx, acct.Address(), token)
	if err != nil {
		log.Fatalf("has_confidential_store: %v", err)
	}
	if !ok {
		log.Fatalf("no confidential store for this token — run: CGO_ENABLED=1 go run ./examples/confidential_asset/register")
	}

	tx, err := cc.Withdraw(ctx, acct, token, amount, recipient, twistedHex, tokenMetadataLong)
	if err != nil {
		if errors.Is(err, confidentialasset.ErrCGODisabled) {
			fmt.Fprintln(os.Stderr, "withdraw requires CGO_ENABLED=1 and libaptos_confidential_asset_ffi (see confidential-asset-bindings/bindings/go/README.md).")
			os.Exit(1)
		}
		log.Fatalf("withdraw_to_raw: %v", err)
	}
	fmt.Printf("withdraw_to_raw: ok version=%d hash=%s\n", tx.Version, tx.Hash)
}
