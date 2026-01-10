// error_handling demonstrates v2 structured error handling.
//
// This example shows:
//   - Using sentinel errors (errors.Is)
//   - Extracting detailed error info (errors.As)
//   - Different error types and when they occur
package main

import (
	"context"
	"errors"
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

	fmt.Println("=== Aptos Go SDK v2 - Error Handling Example ===")
	fmt.Println()

	// Example 1: Not Found Error
	fmt.Println("1. Handling 'Not Found' errors:")
	nonExistentAddr := aptos.MustParseAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	_, err = client.AccountResource(ctx, nonExistentAddr, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	if err != nil {
		if errors.Is(err, aptos.ErrNotFound) {
			fmt.Println("  ✓ Correctly identified as ErrNotFound")
		}

		var apiErr *aptos.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("  Status Code: %d\n", apiErr.StatusCode)
			fmt.Printf("  Error Code: %s\n", apiErr.ErrorCode)
			fmt.Printf("  Message: %s\n", apiErr.Message)
		}
	}
	fmt.Println()

	// Example 2: Checking error types with switch
	fmt.Println("2. Using error type checking:")
	_, err = client.Transaction(ctx, "0xinvalid_hash_that_does_not_exist")
	if err != nil {
		switch {
		case errors.Is(err, aptos.ErrNotFound):
			fmt.Println("  → Transaction not found")
		case errors.Is(err, aptos.ErrRateLimited):
			fmt.Println("  → Rate limited, should retry later")
		case errors.Is(err, aptos.ErrTimeout):
			fmt.Println("  → Request timed out")
		default:
			fmt.Printf("  → Other error: %v\n", err)
		}
	}
	fmt.Println()

	// Example 3: Context cancellation
	fmt.Println("3. Handling context cancellation:")
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc() // Cancel immediately

	_, err = client.Info(cancelCtx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Println("  ✓ Correctly identified as context.Canceled")
		}
	}
	fmt.Println()

	// Example 4: Timeout handling
	fmt.Println("4. Handling timeouts:")
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer shortCancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout expires

	_, err = client.Info(shortCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("  ✓ Correctly identified as context.DeadlineExceeded")
		}
	}
	fmt.Println()

	// Example 5: Best practices for error handling
	fmt.Println("5. Best practices summary:")
	fmt.Println("  - Use errors.Is() for sentinel error checking")
	fmt.Println("  - Use errors.As() to extract detailed error info")
	fmt.Println("  - Always handle context.Canceled and context.DeadlineExceeded")
	fmt.Println("  - Log APIError details for debugging")
	fmt.Println("  - Implement retry logic for ErrRateLimited and ErrTimeout")
}
