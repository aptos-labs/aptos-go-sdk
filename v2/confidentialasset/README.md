# Confidential Asset (Go)

Go SDK for Aptos **confidential fungible assets**, aligned with [`@aptos-labs/confidential-asset`](https://github.com/aptos-labs/aptos-ts-sdk/tree/main/confidential-asset).

Import path: `github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset`

## Features

- **Views**: `HasUserRegistered`, `IsBalanceNormalized`, cipher balance fetches, encryption keys, auditor hints
- **Transactions**: `Deposit`, `RolloverPendingBalance`, `RegisterBalance`, `NormalizeBalance`, `Withdraw`, `Transfer`, `RotateEncryptionKey`
- **Orchestration**: `ConfidentialAsset.RolloverPendingBalance` (normalize then rollover when needed)
- **Gas**: `SubmitWithSimulatedGas` simulates, caps `max_gas` from public FA APT balance (metadata `0xa`)

## Requirements

- `github.com/aptos-labs/aptos-go-sdk/v2` (same module as core SDK)
- **CGO + FFI** for decrypt and proof generation: build [`confidential-asset-bindings`](https://github.com/aptos-labs/confidential-asset-bindings) `libaptos_confidential_asset_ffi` (see bindings `bindings/go/README.md`)
- Set `WithRESTBaseURL` to your fullnode base including `/v1` when using simulated-gas submit helpers

## Usage

```go
import (
    aptos "github.com/aptos-labs/aptos-go-sdk/v2"
    confidentialasset "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
)

client, _ := aptos.NewClient(aptos.Testnet)
ca := confidentialasset.NewClient(client,
    confidentialasset.WithRESTBaseURL("https://fullnode.testnet.aptoslabs.com/v1"),
)

// Views + deposit (no CGO required for deposit)
tx, err := ca.Deposit(ctx, signer, token, amountOctas, faMetadataHex)

// Transfer / normalize / register require CGO_ENABLED=1 and FFI linked
tx, err = ca.Transfer(ctx, signer, token, amountOctas, recipient, twistedHex, faMetadataHex)
```

## Examples

From the `v2` module directory (each subdirectory is its own `main`; requires CGO + FFI):

```bash
cd v2
export APTOS_PRIVATE_KEY=...
CGO_ENABLED=1 go run ./examples/confidential_asset/balance
CGO_ENABLED=1 go run ./examples/confidential_asset/register
CGO_ENABLED=1 go run ./examples/confidential_asset/transfer
CGO_ENABLED=1 go run ./examples/confidential_asset/withdraw
CGO_ENABLED=1 go run ./examples/confidential_asset/deposit_chain
CGO_ENABLED=1 go run ./examples/confidential_asset/ffismoke
```

See [`examples/confidential_asset/README.md`](../examples/confidential_asset/README.md) for environment variables.

## Specification

Move entry/view argument layout: [`doc/SPEC.md`](doc/SPEC.md).

## Local FFI (development)

`v2/go.mod` may `replace` the bindings module to a sibling checkout:

```text
replace github.com/aptos-labs/confidential-asset-bindings/bindings/go => ../../confidential-asset-bindings/bindings/go
```

Build the static library before `CGO_ENABLED=1 go test ./confidentialasset/...`.
