// iterators demonstrates Go 1.23 iterators for streaming data.
//
// This example shows:
//   - Using TransactionsIter for streaming transactions
//   - Using iter utilities (Take, Filter, Map, Collect)
//   - Memory-efficient processing of large datasets
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/iter"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := aptos.NewClient(aptos.Testnet)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Aptos Go SDK v2 - Iterators Example ===")
	fmt.Println()

	// Example 1: Get first 5 transactions using iterator
	fmt.Println("1. First 5 transactions (using iterator):")
	count := 0
	for txn, err := range client.TransactionsIter(ctx, nil) {
		if err != nil {
			log.Fatalf("Iterator error: %v", err)
		}
		fmt.Printf("  %d. Version %d: %s\n", count+1, txn.Version, txn.Type)
		count++
		if count >= 5 {
			break
		}
	}
	fmt.Println()

	// Example 2: Using iter.Take utility
	fmt.Println("2. Using iter.Take to get exactly 3 transactions:")
	first3 := iter.Take(client.TransactionsIter(ctx, nil), 3)
	txns, err := iter.Collect(first3)
	if err != nil {
		log.Fatalf("Collect error: %v", err)
	}
	for i, txn := range txns {
		fmt.Printf("  %d. Version %d, Hash: %s...\n", i+1, txn.Version, txn.Hash[:16])
	}
	fmt.Println()

	// Example 3: Using iter.Map to transform data
	fmt.Println("3. Extracting just transaction hashes:")
	hashIter := iter.Map(
		iter.Take(client.TransactionsIter(ctx, nil), 5),
		func(txn *aptos.Transaction) string {
			return txn.Hash
		},
	)
	hashes, err := iter.Collect(hashIter)
	if err != nil {
		log.Fatalf("Collect error: %v", err)
	}
	for i, hash := range hashes {
		fmt.Printf("  %d. %s\n", i+1, hash)
	}
	fmt.Println()

	// Example 4: Count transactions of each type
	fmt.Println("4. Counting first 20 transaction types:")
	typeCounts := make(map[string]int)
	typeIter := iter.Take(client.TransactionsIter(ctx, nil), 20)
	for txn, err := range typeIter {
		if err != nil {
			log.Fatalf("Iterator error: %v", err)
		}
		typeCounts[txn.Type]++
	}
	for txnType, cnt := range typeCounts {
		fmt.Printf("  %s: %d\n", txnType, cnt)
	}
	fmt.Println()

	fmt.Println("Done! Iterators allow efficient streaming without loading all data into memory.")
}
