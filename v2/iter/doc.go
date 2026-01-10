// Package iter provides Go 1.23 iterator utilities for streaming Aptos blockchain data.
//
// This package offers iterator types and functions for efficiently processing
// large datasets from the Aptos blockchain without loading everything into memory.
//
// # Basic Usage
//
// Use iterators to stream data:
//
//	for txn, err := range client.TransactionsIter(ctx, nil) {
//		if err != nil {
//			return err
//		}
//		fmt.Printf("Transaction: %s\n", txn.Hash)
//	}
//
// # Early Termination
//
// Break out of iteration early:
//
//	for txn, err := range client.TransactionsIter(ctx, nil) {
//		if err != nil {
//			return err
//		}
//		if txn.Version > 1000 {
//			break // Stop iterating
//		}
//	}
//
// # Utility Functions
//
// Transform and filter iterators:
//
//	// Filter only successful transactions
//	successfulTxns := iter.Filter(txns, func(t *Transaction) bool {
//		return t.Success
//	})
//
//	// Take first 10 transactions
//	first10 := iter.Take(txns, 10)
//
//	// Collect into slice
//	slice, err := iter.Collect(txns)
package iter
