# Migrating from Aptos Go SDK v1 to v2

This guide helps you upgrade from Aptos Go SDK v1 to v2. The v2 SDK is a major rewrite that brings idiomatic Go patterns, improved performance, and better developer experience.

## Quick Start

### Automated Migration

Use the `aptosfix` tool to automate most of the migration:

```bash
# Install the migration tool
go install github.com/aptos-labs/aptos-go-sdk/cmd/aptosfix@latest

# Preview changes (recommended first step)
aptosfix -d ./...

# Apply changes
aptosfix -w ./...

# Or step-by-step: first imports only
aptosfix -w -imports ./...
```

### Manual Migration Steps

1. Update your `go.mod`:
   ```bash
   go get github.com/aptos-labs/aptos-go-sdk/v2@latest
   go mod tidy
   ```

2. Update imports in your code
3. Add `context.Context` to client method calls
4. Update error handling
5. Adjust for API changes

## Import Changes

| v1 Import | v2 Import |
|-----------|-----------|
| `github.com/aptos-labs/aptos-go-sdk` | `github.com/aptos-labs/aptos-go-sdk/v2` |
| `github.com/aptos-labs/aptos-go-sdk/bcs` | `github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs` |
| `github.com/aptos-labs/aptos-go-sdk/crypto` | `github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto` |
| `github.com/aptos-labs/aptos-go-sdk/api` | Types moved to main `aptos` package |

### Example

```go
// v1
import (
    "github.com/aptos-labs/aptos-go-sdk"
    "github.com/aptos-labs/aptos-go-sdk/crypto"
)

// v2
import (
    "context"
    "github.com/aptos-labs/aptos-go-sdk/v2"
)
```

## Context Parameter

The biggest API change is that all client methods now require `context.Context` as the first parameter. This enables:
- Request cancellation
- Timeouts
- Deadline propagation
- Tracing integration

### Before (v1)

```go
client, _ := aptos.NewClient(aptos.Testnet)

// No context
info, err := client.Info()
account, err := client.Account(address)
balance, err := client.AccountAPTBalance(address)
```

### After (v2)

```go
client, _ := aptos.NewClient(aptos.Testnet)
ctx := context.Background()

// With context
info, err := client.Info(ctx)
account, err := client.Account(ctx, address)
balance, err := client.AccountBalance(ctx, address)

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
info, err := client.Info(ctx)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // Cancel after some condition
    cancel()
}()
info, err := client.Info(ctx)
```

## Client Configuration

v2 uses functional options for client configuration instead of struct fields.

### Before (v1)

```go
client, err := aptos.NewClient(aptos.NetworkConfig{
    Name:    "testnet",
    NodeUrl: "https://fullnode.testnet.aptoslabs.com/v1",
})
client.SetTimeout(10 * time.Second)
client.SetHeader("Authorization", "Bearer token")
```

### After (v2)

```go
client, err := aptos.NewClient(
    aptos.Testnet,
    aptos.WithTimeout(10*time.Second),
    aptos.WithHeader("Authorization", "Bearer token"),
    aptos.WithRetry(3, time.Second),
    aptos.WithRateLimitHandling(true, 30*time.Second),
)
```

### Available Options

| Option | Description |
|--------|-------------|
| `WithTimeout(d)` | Set HTTP client timeout |
| `WithHTTPClient(client)` | Use custom HTTP client |
| `WithLogger(logger)` | Set structured logger |
| `WithRetry(n, backoff)` | Configure retry behavior |
| `WithRetryConfig(cfg)` | Advanced retry configuration |
| `WithRateLimitHandling(wait, max)` | Handle rate limits |
| `WithHeader(k, v)` | Add custom header |
| `WithAPIKey(key)` | Set API key header |

## Error Handling

v2 introduces structured errors with sentinel values and typed error structs.

### Before (v1)

```go
info, err := client.Info()
if err != nil {
    if httpErr, ok := err.(*aptos.HttpError); ok {
        if httpErr.StatusCode == 404 {
            // Handle not found
        }
    }
}
```

### After (v2)

```go
info, err := client.Info(ctx)
if err != nil {
    // Check for specific error types
    if errors.Is(err, aptos.ErrNotFound) {
        // Handle not found
    }
    if errors.Is(err, aptos.ErrRateLimited) {
        // Handle rate limit
    }
    
    // Get detailed API error info
    var apiErr *aptos.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status: %d, Code: %s, Message: %s\n",
            apiErr.StatusCode, apiErr.ErrorCode, apiErr.Message)
    }
}
```

### Sentinel Errors

| Error | Description |
|-------|-------------|
| `ErrNotFound` | Resource not found (404) |
| `ErrRateLimited` | Rate limit exceeded (429) |
| `ErrBadRequest` | Invalid request (400) |
| `ErrUnauthorized` | Authentication failed (401) |
| `ErrForbidden` | Access denied (403) |
| `ErrTimeout` | Request timed out (408) |
| `ErrInternalServer` | Server error (500) |
| `ErrUnavailable` | Service unavailable (503) |
| `ErrTransactionFailed` | Transaction execution failed |

## BCS Serialization

BCS function names have been updated to match Go conventions.

### Before (v1)

```go
import "github.com/aptos-labs/aptos-go-sdk/bcs"

// Serialize
data, err := bcs.Serialize(&value)
ser := bcs.Serializer{}
ser.SerializeU64(123)

// Deserialize
err := bcs.Deserialize(&result, data)
des := bcs.Deserializer{data}
val := des.DeserializeU64()
```

### After (v2)

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"

// Marshal (renamed from Serialize)
data, err := bcs.Marshal(&value)
ser := bcs.NewSerializer()
ser.U64(123)

// Unmarshal (renamed from Deserialize)
err := bcs.Unmarshal(data, &result)
des := bcs.NewDeserializer(data)
val := des.U64()
```

## Transaction Building

v2 provides functional options for transaction configuration.

### Before (v1)

```go
rawTxn, err := client.BuildTransaction(
    sender.AccountAddress(),
    payload,
    aptos.MaxGasAmount(100000),
    aptos.GasUnitPrice(100),
    aptos.ExpirationSeconds(600),
)
```

### After (v2)

```go
rawTxn, err := client.BuildTransaction(
    ctx,
    sender.AccountAddress(),
    payload,
    aptos.WithMaxGas(100000),
    aptos.WithGasPrice(100),
    aptos.WithExpiration(10*time.Minute),
    aptos.WithGasEstimation(), // Auto-estimate gas
)
```

## Iterators (New in v2)

v2 introduces Go 1.23 iterators for streaming large result sets.

```go
// Iterate over transactions without loading all into memory
for txn, err := range client.TransactionsIter(ctx, nil) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(txn.Hash)
}

// With iter package utilities
import "github.com/aptos-labs/aptos-go-sdk/v2/iter"

// Take first 10
first10 := iter.Take(client.TransactionsIter(ctx, nil), 10)

// Filter
userTxns := iter.Filter(client.TransactionsIter(ctx, nil), func(t *Transaction, _ error) bool {
    return t.Type == "user_transaction"
})

// Collect to slice
txns, err := iter.Collect(iter.Take(client.TransactionsIter(ctx, nil), 100))
```

## Account Creation

### Before (v1)

```go
account, err := aptos.NewEd25519Account()
singleSigner, err := aptos.NewEd25519SingleSenderAccount()
```

### After (v2)

```go
// Same API, just different import
account, err := aptos.NewEd25519Account()
singleSigner, err := aptos.NewEd25519SingleSignerAccount()
```

## Crypto Changes

Crypto types are now in the internal package but re-exported from the main package.

### Before (v1)

```go
import "github.com/aptos-labs/aptos-go-sdk/crypto"

key, _ := crypto.GenerateEd25519PrivateKey()
pubKey := key.PubKey()
sig, _ := key.Sign(message)
```

### After (v2)

```go
// Crypto is now internal, use main package exports
import "github.com/aptos-labs/aptos-go-sdk/v2"

// Or import internal if needed
import "github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"

key, _ := crypto.GenerateEd25519PrivateKey()
pubKey := key.PublicKey()  // Note: PubKey -> PublicKey
sig, _ := key.Sign(message)
```

## Type Changes

### AccountAddress

```go
// v1 & v2 - Same API
var addr aptos.AccountAddress
err := addr.ParseStringRelaxed("0x1")
str := addr.String()
```

### TypeTag

```go
// v1 & v2 - Same API
tag, err := aptos.ParseTypeTag("0x1::aptos_coin::AptosCoin")
str := tag.String()
```

## New Features in v2

### 1. Keyless Authentication

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/keyless"

// Create keyless account (ZK-based, social login)
config := keyless.Config{
    // Configuration
}
account, err := keyless.NewAccount(config)
```

### 2. ANS (Aptos Names Service)

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/ans"

ansClient := ans.NewClient(client)
address, err := ansClient.Resolve(ctx, "myname.apt")
name, err := ansClient.ReverseLookup(ctx, address)
```

### 3. Structured Logging

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
client, _ := aptos.NewClient(aptos.Testnet, aptos.WithLogger(logger))
```

### 4. Test Utilities

```go
import "github.com/aptos-labs/aptos-go-sdk/v2/testutil"

// Create fake client for testing
fake := testutil.NewFakeClient()
fake.SetAccount(addr, &aptos.AccountInfo{...})

// Use in tests
result, err := myFunction(fake)
```

## Method Renames

| v1 Method | v2 Method |
|-----------|-----------|
| `AccountAPTBalance` | `AccountBalance` |
| `NodeAPIHealthCheck` | `HealthCheck` |

## Removed/Changed APIs

### Removed

- `api` package - Types moved to main package
- Direct HTTP client access - Use `WithHTTPClient` option

### Changed

- All client methods require `context.Context`
- Error types are now structured
- Configuration uses functional options

## Migration Checklist

- [ ] Update `go.mod` to use v2
- [ ] Run `aptosfix -w ./...` to auto-migrate
- [ ] Add `"context"` import where needed
- [ ] Update error handling to use `errors.Is`/`errors.As`
- [ ] Review client configuration for functional options
- [ ] Update BCS calls (`Serialize` → `Marshal`)
- [ ] Test thoroughly

## Getting Help

- [v2 Documentation](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk/v2)
- [v2 Examples](./examples/)
- [GitHub Issues](https://github.com/aptos-labs/aptos-go-sdk/issues)
- [Aptos Discord](https://discord.gg/aptosnetwork)

## Troubleshooting

### "undefined: context"

Add the context import:
```go
import "context"
```

### "too many arguments"

You probably need to add `ctx` as the first argument:
```go
// Wrong
client.Info()

// Correct
client.Info(ctx)
```

### "cannot use ... as HTTPDoer"

If you have a custom HTTP client, ensure it implements `HTTPDoer`:
```go
type HTTPDoer interface {
    Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
```

### BCS errors

Update function names:
```go
// Old
bcs.Serialize(&v)
bcs.Deserialize(&v, data)

// New
bcs.Marshal(&v)
bcs.Unmarshal(data, &v)
```

