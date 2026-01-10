// view_function demonstrates calling read-only view functions.
//
// This example shows:
//   - Calling view functions on Move modules
//   - Parsing view function results
//   - No transaction/gas required for views
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := aptos.NewClient(aptos.Testnet)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Aptos Go SDK v2 - View Function Example ===")
	fmt.Println()

	// Example 1: Check if an account exists
	fmt.Println("1. Checking if account 0x1 exists...")
	result, err := client.View(ctx, &aptos.ViewPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "account"},
		Function: "exists_at",
		TypeArgs: nil,
		Args:     []any{aptos.AccountOne.String()},
	})
	if err != nil {
		log.Fatalf("View failed: %v", err)
	}
	fmt.Printf("Account 0x1 exists: %v\n", result[0])
	fmt.Println()

	// Example 2: Get chain ID
	fmt.Println("2. Getting chain ID...")
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	fmt.Printf("Chain ID: %d\n", chainID)
	fmt.Println()

	// Example 3: Get node info
	fmt.Println("3. Getting node info...")
	info, err := client.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get info: %v", err)
	}
	fmt.Printf("Chain ID: %d\n", info.ChainID)
	fmt.Printf("Ledger Version: %d\n", info.LedgerVersion)
	fmt.Printf("Block Height: %d\n", info.BlockHeight)
	fmt.Printf("Node Role: %s\n", info.NodeRole)
	fmt.Println()

	// Example 4: Get gas estimate
	fmt.Println("4. Getting gas estimates...")
	gas, err := client.EstimateGasPrice(ctx)
	if err != nil {
		log.Fatalf("Failed to get gas: %v", err)
	}
	fmt.Printf("Normal gas price: %d\n", gas.GasEstimate)
	fmt.Printf("Prioritized gas price: %d\n", gas.PrioritizedGasEstimate)
	fmt.Printf("Deprioritized gas price: %d\n", gas.DeprioritizedGasEstimate)
}
