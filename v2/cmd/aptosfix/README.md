# aptosfix

`aptosfix` is a migration tool for upgrading Go code from Aptos Go SDK v1 to v2.

## Installation

```bash
go install github.com/aptos-labs/aptos-go-sdk/v2/cmd/aptosfix@latest
```

Or build from source:

```bash
cd v2/cmd/aptosfix
go build -o aptosfix .
```

## Usage

```bash
aptosfix [flags] [path ...]
```

### Flags

| Flag | Description |
|------|-------------|
| `-w` | Write result to (source) file instead of stdout |
| `-d` | Display diffs instead of rewriting files |
| `-l` | List files whose formatting differs from aptosfix's |
| `-v` | Verbose mode: print files being processed |
| `-imports` | Only update import paths (skip other transformations) |
| `-dry-run` | Show what would be changed without modifying files |

### Examples

```bash
# Preview changes to a single file
aptosfix -d myfile.go

# Fix all files in current directory
aptosfix -w .

# Only update imports (safer first step)
aptosfix -w -imports .

# List files that need updating
aptosfix -l ./...

# Dry run - see what would change
aptosfix -dry-run -d ./...

# Process specific package
aptosfix -w ./internal/...
```

## What It Does

### Import Path Updates

The tool automatically updates import paths from v1 to v2:

| v1 Import | v2 Import |
|-----------|-----------|
| `github.com/aptos-labs/aptos-go-sdk` | `github.com/aptos-labs/aptos-go-sdk/v2` |
| `github.com/aptos-labs/aptos-go-sdk/bcs` | `github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs` |
| `github.com/aptos-labs/aptos-go-sdk/crypto` | `github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto` |

### Context Parameter Addition

Client methods in v2 require `context.Context` as the first parameter. The tool:

1. Detects method calls that need context
2. Adds `ctx` as the first argument to empty parameter lists
3. Adds a TODO comment reminding you to add the context import

Methods that require context include:
- `Info`, `Account`, `AccountResource`, `AccountResources`
- `BlockByHeight`, `BlockByVersion`
- `TransactionByHash`, `TransactionByVersion`, `Transactions`
- `SubmitTransaction`, `SimulateTransaction`
- `BuildTransaction`, `BuildSignAndSubmitTransaction`
- `View`, `EstimateGasPrice`, `Fund`
- And more...

### Function Renames

Some functions have been renamed for consistency:

| v1 Name | v2 Name |
|---------|---------|
| `BCSSerialize` | `BCSMarshal` |
| `BCSDeserialize` | `BCSUnmarshal` |

## Migration Steps

We recommend a staged migration approach:

### Step 1: Update Imports Only

```bash
aptosfix -w -imports ./...
```

This is the safest first step - it only updates import paths without changing any code logic.

### Step 2: Preview Full Changes

```bash
aptosfix -d ./...
```

Review the diff output to understand what changes will be made.

### Step 3: Apply Full Migration

```bash
aptosfix -w ./...
```

### Step 4: Fix Compilation Errors

After running aptosfix, you'll likely need to:

1. Add `"context"` to your imports
2. Create or pass `context.Context` to client method calls
3. Handle any v2-specific API changes not covered by the tool

### Step 5: Update go.mod

Update your `go.mod` to use the v2 SDK:

```bash
go get github.com/aptos-labs/aptos-go-sdk/v2@latest
go mod tidy
```

## Limitations

The tool handles common migration patterns but cannot automatically fix:

1. **Complex context threading**: You may need to manually thread `context.Context` through your call stack
2. **Error handling changes**: v2 uses structured errors - review error handling code
3. **Custom BCS implementations**: If you've implemented custom BCS marshaling, review for compatibility
4. **Removed APIs**: Some v1 APIs may not have direct v2 equivalents

## Manual Review Checklist

After running aptosfix, review:

- [ ] All `context.Context` parameters are properly passed
- [ ] Error handling uses `errors.Is` and `errors.As` for v2 error types
- [ ] BCS serialization uses the new `Marshal`/`Unmarshal` naming
- [ ] Functional options are used for client configuration
- [ ] Any custom HTTP clients implement the `HTTPDoer` interface

## Getting Help

If you encounter issues:

1. Run with `-v` for verbose output
2. Use `-d` to see what changes would be made
3. Check the [migration guide](../../MIGRATION.md) for detailed information
4. Open an issue at https://github.com/aptos-labs/aptos-go-sdk/issues

## Example Migration

### Before (v1)

```go
package main

import (
    "fmt"
    "github.com/aptos-labs/aptos-go-sdk"
)

func main() {
    client, _ := aptos.NewClient(aptos.Testnet)
    info, _ := client.Info()
    fmt.Println(info.ChainId)
}
```

### After (v2)

```go
package main

import (
    "context"
    "fmt"
    "github.com/aptos-labs/aptos-go-sdk/v2"
)

func main() {
    ctx := context.Background()
    client, _ := aptos.NewClient(aptos.Testnet)
    info, _ := client.Info(ctx)
    fmt.Println(info.ChainId)
}
```
