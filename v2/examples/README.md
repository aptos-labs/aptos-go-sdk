# Aptos Go SDK v2 Examples

This directory contains example programs demonstrating the Aptos Go SDK v2.

## Prerequisites

- Go 1.23 or later
- Network access to Aptos testnet

## Examples

### Basic Examples

| Example | Description |
|---------|-------------|
| [transfer_coin](./transfer_coin/) | Basic APT transfer between accounts |
| [view_function](./view_function/) | Calling read-only view functions |
| [error_handling](./error_handling/) | Structured error handling patterns |

### Advanced Examples

| Example | Description |
|---------|-------------|
| [sponsored_transaction](./sponsored_transaction/) | Fee payer (sponsored) transactions |
| [iterators](./iterators/) | Go 1.23 iterators for streaming data |

## Running Examples

Each example is a standalone Go program:

```bash
cd examples/transfer_coin
go run main.go
```

Or run from the v2 directory:

```bash
go run ./examples/transfer_coin
```

## Example Walkthrough

### transfer_coin

Demonstrates the basic flow of:
1. Creating accounts
2. Funding via faucet
3. Building a transaction
4. Signing and submitting
5. Waiting for confirmation

```go
// Create client
client, _ := aptos.NewClient(aptos.Testnet)
ctx := context.Background()

// Create accounts
alice, _ := aptos.NewEd25519Account()
bob, _ := aptos.NewEd25519Account()

// Fund sender
client.Fund(ctx, alice.Address, 100_000_000)

// Build transaction
payload, _ := aptos.APTTransferPayload(bob.Address, 1000)
rawTxn, _ := client.BuildTransaction(ctx, alice.Address, 
    aptos.TransactionPayload{Payload: payload},
    aptos.WithGasEstimation(),
)

// Sign and submit
signedTxn, _ := rawTxn.SignedTransaction(alice.Signer)
result, _ := client.SubmitTransaction(ctx, signedTxn)

// Wait for confirmation
client.WaitForTransaction(ctx, result.Hash)
```

### error_handling

Shows how to handle errors idiomatically:

```go
_, err := client.Account(ctx, address)
if err != nil {
    switch {
    case errors.Is(err, aptos.ErrNotFound):
        // Handle not found
    case errors.Is(err, aptos.ErrRateLimited):
        // Handle rate limit, retry later
    default:
        // Handle other errors
    }
    
    // Get detailed error info
    var apiErr *aptos.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API Error %d: %s", apiErr.StatusCode, apiErr.Message)
    }
}
```

### iterators

Demonstrates Go 1.23 iterators for memory-efficient streaming:

```go
// Stream transactions without loading all into memory
for txn, err := range client.TransactionsIter(ctx, nil) {
    if err != nil {
        break
    }
    fmt.Println(txn.Hash)
}

// Use iter utilities
first10 := iter.Take(client.TransactionsIter(ctx, nil), 10)
txns, _ := iter.Collect(first10)
```

## Key Differences from v1

1. **Context parameter**: All methods require `context.Context`
2. **Functional options**: Use `WithXxx()` functions for configuration
3. **Structured errors**: Use `errors.Is()` and `errors.As()`
4. **Iterators**: New Go 1.23 iterators for streaming

## More Resources

- [Migration Guide](../MIGRATION.md)
- [API Documentation](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk/v2)
- [Aptos Developer Docs](https://aptos.dev)

