// transfer_coin demonstrates a basic APT transfer between two accounts.
//
// This example shows:
//   - Connecting to the network
//   - Funding accounts via faucet
//   - Building and submitting transactions
//   - Waiting for transaction confirmation
//   - Checking balances
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

const (
	FundAmount     = 100_000_000 // 1 APT in octas
	TransferAmount = 1_000       // 0.00001 APT
)

func main() {
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create client connected to testnet
	client, err := aptos.NewClient(aptos.Testnet)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Aptos Go SDK v2 - Transfer Coin Example ===")
	fmt.Println()

	// Create two addresses for Alice (sender) and Bob (receiver)
	// In a real app, these would come from wallets/key management
	alice := aptos.MustParseAddress("0x1") // Using 0x1 as demo (would be a real account)
	bob := aptos.MustParseAddress("0x2")   // Using 0x2 as demo

	fmt.Printf("Alice: %s\n", alice.String())
	fmt.Printf("Bob:   %s\n", bob.String())
	fmt.Println()

	// Fund Alice's account using the testnet faucet
	fmt.Println("Funding Alice's account...")
	err = client.Fund(ctx, alice, FundAmount)
	if err != nil {
		log.Fatalf("Failed to fund Alice: %v", err)
	}
	fmt.Printf("Funded Alice with %d octas (%.8f APT)\n", FundAmount, float64(FundAmount)/1e8)
	fmt.Println()

	// Check initial balances
	fmt.Println("Initial balances:")
	printBalance(ctx, client, "Alice", alice)
	printBalance(ctx, client, "Bob", bob)
	fmt.Println()

	// Build the transfer transaction
	fmt.Printf("Building transfer of %d octas from Alice to Bob...\n", TransferAmount)

	// Create the transfer payload (APT transfer using aptos_account::transfer)
	payload := &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "aptos_account"},
		Function: "transfer",
		TypeArgs: nil,
		Args:     []any{bob.String(), fmt.Sprintf("%d", TransferAmount)},
	}

	// Build the transaction with gas estimation
	rawTxn, err := client.BuildTransaction(ctx, alice, payload,
		aptos.WithGasEstimation(),
	)
	if err != nil {
		log.Fatalf("Failed to build transaction: %v", err)
	}

	fmt.Printf("Transaction built: sender=%s, seq=%d, max_gas=%d\n",
		rawTxn.Sender.String(), rawTxn.SequenceNumber, rawTxn.MaxGasAmount)
	fmt.Println()

	// Note: In a real application, you would sign the transaction here.
	// This example shows the structure but doesn't actually submit
	// because we don't have a real private key for the demo addresses.

	fmt.Println("Transaction structure ready for signing!")
	fmt.Println("In a real app, you would:")
	fmt.Println("  1. Sign with: signedTxn := signer.Sign(rawTxn)")
	fmt.Println("  2. Submit with: result, _ := client.SubmitTransaction(ctx, signedTxn)")
	fmt.Println("  3. Wait with: client.WaitForTransaction(ctx, result.Hash)")
}

func printBalance(ctx context.Context, client aptos.Client, name string, addr aptos.AccountAddress) {
	balance, err := client.AccountBalance(ctx, addr)
	if err != nil {
		fmt.Printf("%s: (no balance - account may not exist)\n", name)
		return
	}
	fmt.Printf("%s: %d octas (%.8f APT)\n", name, balance, float64(balance)/1e8)
}
