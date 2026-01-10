[![Go Reference](https://pkg.go.dev/badge/github.com/aptos-labs/aptos-go-sdk.svg)](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptos-labs/aptos-go-sdk)](https://goreportcard.com/report/github.com/aptos-labs/aptos-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aptos-labs/aptos-go-sdk)
[![GitHub Tag](https://img.shields.io/github/v/tag/aptos-labs/aptos-go-sdk?label=Latest%20Version)](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk)

# Aptos Go SDK

A comprehensive Go SDK for building applications on the [Aptos blockchain](https://aptos.dev).

## Features

- 🔐 **Multiple Key Types** - Ed25519, Secp256k1, MultiKey, and MultiEd25519 support
- 💸 **Transaction Building** - Simple and advanced transaction construction
- 🎫 **Sponsored Transactions** - Fee payer and multi-agent transaction support
- 📊 **Fungible Assets** - Full FA standard support
- 🏷️ **ANS Integration** - Aptos Names Service (.apt domain) resolution
- 📈 **Telemetry** - OpenTelemetry tracing and metrics
- 🔧 **Code Generation** - Generate type-safe Go bindings from Move ABIs
- ⚡ **Concurrent Operations** - Batch transaction submission

## Installation

```bash
go get github.com/aptos-labs/aptos-go-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/aptos-labs/aptos-go-sdk"
)

func main() {
    // Create a client
    client, err := aptos.NewClient(aptos.TestnetConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Create an account
    account, err := aptos.NewEd25519Account()
    if err != nil {
        log.Fatal(err)
    }

    // Fund it (testnet only)
    err = client.Fund(account.AccountAddress(), 100_000_000)
    if err != nil {
        log.Fatal(err)
    }

    // Check balance
    balance, err := client.AccountAPTBalance(account.AccountAddress())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Balance: %d octas\n", balance)
}
```

## Examples

The `examples/` directory contains runnable examples:

| Example | Description |
|---------|-------------|
| [`transfer_coin`](examples/transfer_coin) | Basic APT transfer |
| [`sponsored_transaction`](examples/sponsored_transaction) | Fee payer transactions |
| [`multi_agent`](examples/multi_agent) | Multi-signer transactions |
| [`fungible_asset`](examples/fungible_asset) | Fungible asset operations |
| [`ans`](examples/ans) | ANS domain resolution |
| [`telemetry_example`](examples/telemetry_example) | OpenTelemetry integration |
| [`codegen`](examples/codegen) | Code generation from ABIs |
| [`sending_concurrent_transactions`](examples/sending_concurrent_transactions) | Batch operations |

Run an example:
```bash
cd examples/transfer_coin
go run main.go
```

## Network Configuration

```go
// Testnet (recommended for development)
client, _ := aptos.NewClient(aptos.TestnetConfig)

// Devnet (resets weekly)
client, _ := aptos.NewClient(aptos.DevnetConfig)

// Mainnet
client, _ := aptos.NewClient(aptos.MainnetConfig)

// Local node
client, _ := aptos.NewClient(aptos.LocalnetConfig)
```

## Building Transactions

### Simple Transfer

```go
txn, err := aptos.APTTransferTransaction(client, sender, receiver, amount)
signedTxn, err := txn.SignedTransaction(sender)
resp, err := client.SubmitTransaction(signedTxn)
```

### With Options

```go
rawTxn, err := client.BuildTransaction(
    sender.AccountAddress(),
    payload,
    aptos.MaxGasAmount(10_000),
    aptos.GasUnitPrice(100),
    aptos.ExpirationSeconds(300),
)
```

### Sponsored Transaction

```go
rawTxn, err := client.BuildTransactionMultiAgent(
    sender.AccountAddress(),
    payload,
    aptos.FeePayer(&feePayerAddress),
)
```

## Aptos Names Service (ANS)

```go
ansClient := aptos.NewANSClient(client)

// Resolve name to address
address, err := ansClient.Resolve("alice.apt")

// Get primary name for address
name, err := ansClient.GetPrimaryName(address)

// Check availability
available, err := ansClient.IsAvailable("myname")
```

## Telemetry

Add OpenTelemetry observability:

```go
import "github.com/aptos-labs/aptos-go-sdk/telemetry"

httpClient, _ := telemetry.NewInstrumentedHTTPClient(
    telemetry.WithServiceName("my-app"),
)
client, _ := aptos.NewNodeClientWithHttpClient(
    aptos.MainnetConfig.NodeUrl,
    aptos.MainnetConfig.ChainId,
    httpClient,
)
```

## Code Generation

Generate type-safe Go code from Move ABIs:

```bash
# Install the generator
go install github.com/aptos-labs/aptos-go-sdk/cmd/aptosgen@latest

# Generate from on-chain module
aptosgen -address 0x1 -module coin -package coin -output coin/

# Generate from local ABI file
aptosgen -abi coin.json -package coin -output coin/
```

## Documentation

- [Go Package Documentation](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk)
- [Aptos Developer Docs](https://aptos.dev)
- [Go SDK Guide](https://aptos.dev/sdks/go-sdk/)

## Development

1. Clone the repository
2. Make your changes
3. Update `CHANGELOG.md`
4. Run formatters and linters:
   ```bash
   gofumpt -l -w .
   golangci-lint run
   ```
5. Run tests:
   ```bash
   go test ./...
   ```
6. Submit a PR

## Publishing

1. Update `CHANGELOG.md` with a PR
2. Create a new tag (e.g., `v1.12.0`) with the list of changes

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.
