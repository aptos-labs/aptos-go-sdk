# Aptos Go SDK v2

[![codecov](https://codecov.io/gh/aptos-labs/aptos-go-sdk/graph/badge.svg?token=YOUR_TOKEN&flag=v2)](https://codecov.io/gh/aptos-labs/aptos-go-sdk)
[![Go Reference](https://pkg.go.dev/badge/github.com/aptos-labs/aptos-go-sdk/v2.svg)](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk/v2)

A modern, idiomatic Go SDK for the [Aptos blockchain](https://aptos.dev).

## Features

- **Idiomatic Go**: Designed with Go best practices from the ground up
- **Go 1.24+**: Leverages modern Go features including generics and iterators
- **Context-aware**: All operations accept `context.Context` for cancellation and timeouts
- **Functional options**: Flexible configuration using the functional options pattern
- **Comprehensive error handling**: Sentinel errors and typed errors with `errors.Is/As` support
- **Type-safe BCS**: Binary Canonical Serialization with reflection support
- **Iterator support**: Go 1.23 iterators for streaming large datasets
- **Testable**: Mock client and test helpers included

## Installation

```bash
go get github.com/aptos-labs/aptos-go-sdk/v2
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

func main() {
    ctx := context.Background()

    // Create a client for devnet
    client, err := aptos.NewClient(aptos.Devnet)
    if err != nil {
        log.Fatal(err)
    }

    // Get account info
    address := aptos.MustParseAddress("0x1")
    info, err := client.Account(ctx, address)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Sequence number: %d\n", info.SequenceNumber)
}
```

## Core Concepts

### Client

The `Client` interface is the main entry point for interacting with Aptos:

```go
// Create client with options
client, err := aptos.NewClient(aptos.Mainnet,
    aptos.WithTimeout(30 * time.Second),
    aptos.WithRetry(3, 100*time.Millisecond),
    aptos.WithAPIKey("your-api-key"),
)
```

### Accounts and Signers

```go
// Generate a new account
signer, err := crypto.GenerateEd25519PrivateKey()
if err != nil {
    return err
}

// Get the address
address := signer.AuthKey().Address()

// Fund the account (devnet/testnet only)
err = client.Fund(ctx, address, 100_000_000)
```

### Building Transactions

Use the fluent transaction builder:

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/transaction"

// Build a transfer transaction
txn, err := transaction.New().
    Sender(senderAddress).
    EntryFunction("0x1::aptos_account::transfer",
        nil, // no type args
        recipientAddress.Bytes(),
        uint64(1_000_000),
    ).
    MaxGas(2000).
    GasPrice(100).
    Build(ctx, client)
```

Or use convenience builders:

```go
// Transfer APT
txn, err := transaction.TransferAPT(sender, recipient, 1_000_000).
    Build(ctx, client)

// Transfer other coins
txn, err := transaction.TransferCoins(sender, recipient, coinType, amount).
    Build(ctx, client)
```

### Iterators

Stream large datasets efficiently:

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/iter"

// Iterate over transactions
for txn, err := range client.TransactionsIter(ctx, nil) {
    if err != nil {
        return err
    }
    fmt.Printf("Transaction: %s\n", txn.Hash)
}

// Use iterator utilities
successfulTxns := iter.Filter(txns, func(t *aptos.Transaction) bool {
    return t.Success
})

first10 := iter.Take(successfulTxns, 10)
```

### Error Handling

```go
info, err := client.Account(ctx, address)
if err != nil {
    // Check for specific error types
    if errors.Is(err, aptos.ErrNotFound) {
        fmt.Println("Account not found")
    } else if errors.Is(err, aptos.ErrRateLimited) {
        fmt.Println("Rate limited, try again later")
    }
    
    // Get detailed API error info
    var apiErr *aptos.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
    }
    return err
}
```

### Keyless Authentication

Authenticate users with their OIDC identities:

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/keyless"

// Generate ephemeral key pair
ephemeralKeyPair, err := keyless.GenerateEphemeralKeyPair(time.Hour)

// Get nonce for OIDC redirect
nonce := ephemeralKeyPair.Nonce()

// After OIDC auth, derive the account
claims, err := keyless.ParseJWT(jwtToken)
address, err := keyless.DeriveAddress(claims, "sub", pepper)
```

### ANS (Aptos Names Service)

Resolve human-readable names:

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/ans"

ansClient := ans.NewClient(client)

// Resolve name to address
address, err := ansClient.Resolve(ctx, "alice.apt")

// Get primary name for address
name, err := ansClient.GetPrimaryName(ctx, address)

// Register a name (requires signing)
payload, err := ansClient.RegisterPayload("myname.apt", ans.RegisterOptions{Years: 1})
```

## Package Structure

| Package | Description |
|---------|-------------|
| `aptos` | Main client, types, and errors |
| `aptos/transaction` | Transaction builder with fluent API |
| `aptos/iter` | Go 1.23 iterator utilities |
| `aptos/keyless` | Keyless (ZK) authentication |
| `aptos/ans` | Aptos Names Service integration |
| `aptos/testutil` | Test helpers and mock client |
| `aptos/internal/bcs` | BCS serialization |
| `aptos/internal/crypto` | Cryptographic primitives |
| `aptos/internal/types` | Core type definitions |
| `aptos/internal/http` | HTTP client with retry |

## Testing

Use the included test utilities:

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/testutil"

func TestMyFunction(t *testing.T) {
    // Create a fake client
    client := testutil.NewFakeClient().
        WithAccount(addr, &aptos.AccountInfo{SequenceNumber: 5}).
        WithBalance(addr, 100_000_000).
        WithRecording()

    // Test your code
    result, err := myFunction(client, addr)
    require.NoError(t, err)

    // Verify interactions
    calls := client.RecordedCalls()
    assert.Len(t, calls, 2)
}
```

## Migration from v1

### Import Path

```go
// v1
import "github.com/aptos-labs/aptos-go-sdk"

// v2
import aptos "github.com/aptos-labs/aptos-go-sdk/v2"
```

### Client Creation

```go
// v1
client, err := aptos.NewClient(aptos.DevnetConfig)

// v2
client, err := aptos.NewClient(aptos.Devnet,
    aptos.WithTimeout(30*time.Second),
)
```

### Context Usage

```go
// v1
info, err := client.GetAccount(address)

// v2 - all methods require context
ctx := context.Background()
info, err := client.Account(ctx, address)
```

### Error Handling

```go
// v1
if err != nil {
    // String comparison or type assertion
}

// v2
if errors.Is(err, aptos.ErrNotFound) {
    // Proper error checking
}
```

### BCS Serialization

```go
// v2 - reflection-based serialization
type MyStruct struct {
    Value uint64  `bcs:"value"`
    Name  string  `bcs:"name"`
}

data, err := bcs.Marshal(&MyStruct{Value: 42, Name: "test"})
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 - see [LICENSE](../LICENSE)

