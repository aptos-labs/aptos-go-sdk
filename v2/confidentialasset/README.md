# Confidential Asset (Go)

Go SDK for Aptos **confidential fungible assets**, aligned with [`@aptos-labs/confidential-asset`](https://github.com/aptos-labs/aptos-ts-sdk/tree/main/confidential-asset).

Import paths:

- `github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset` — views, deposit, register (sigma), rotate key
- `github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/native` — decrypt, normalize, withdraw, transfer (**requires CGO + FFI**)

## Features

- **Views** (no CGO): `HasUserRegistered`, `IsBalanceNormalized`, cipher balance fetches, encryption keys, auditor hints
- **Transactions** (no CGO): `Deposit`, `RolloverPendingBalance`, `RegisterBalance`, `RotateEncryptionKey`
- **FFI** (`native` package): `GetBalance`, `NormalizeBalance`, `Withdraw`, `Transfer`, `ConfidentialAsset.RolloverPendingBalance`
- **Gas**: `SubmitWithSimulatedGas` simulates, caps `max_gas` from public FA APT balance (metadata `0xa`)

## Requirements

- `github.com/aptos-labs/aptos-go-sdk/v2` (same module as core SDK)
- **CGO + FFI** only when importing `confidentialasset/native`: build [`confidential-asset-bindings`](https://github.com/aptos-labs/confidential-asset-bindings) Rust member **`aptos_confidential_asset_ffi`** → `libaptos_confidential_asset_ffi.a`
- Set `WithRESTBaseURL` to your fullnode base including `/v1` when using simulated-gas submit helpers

Importing `native` with `CGO_ENABLED=0` fails at compile time (`build constraints exclude all Go files`).

## TS SDK parity

File-level and API mapping vs **`@aptos-labs/confidential-asset`**: **[doc/TS_GO_MAP.md](doc/TS_GO_MAP.md)** (for porting, debugging proofs, and AI-assisted changes). Move argument layouts remain in **[doc/SPEC.md](doc/SPEC.md)**.

## CI and local testing

| Scenario | Command | Coverage |
|------|------|----------|
| **Confidential asset (v2)** nocgo job | `CGO_ENABLED=0 go test -short` on `confidentialasset/...` **excluding** `native` and `internal/rangeproof` | views, movearg, sigbcs, compile constraint tests |
| **Confidential asset (v2)** cgo job | `CGO_ENABLED=1 go test -short ./confidentialasset/...` | `native` FFI smoke, `rangeproof`, etc. |
| **Code Coverage** | `v2/` with `CGO_ENABLED=1` + FFI build | full v2 including `native` |

## Usage

```go
import (
    aptos "github.com/aptos-labs/aptos-go-sdk/v2"
    confidentialasset "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
    "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/native"
)

client, _ := aptos.NewClient(aptos.Testnet)
ca := confidentialasset.NewClient(client,
    confidentialasset.WithRESTBaseURL("https://fullnode.testnet.aptoslabs.com/v1"),
)

// No CGO
tx, err := ca.Deposit(ctx, signer, token, amountOctas, faMetadataHex)

// CGO + FFI (import native)
nc := native.Wrap(ca)
tx, err = nc.Transfer(ctx, signer, token, amountOctas, recipient, twistedHex, faMetadataHex)
```

## Examples

From `v2` (FFI examples need `CGO_ENABLED=1` and built static lib):

```bash
cd v2
export APTOS_PRIVATE_KEY=...
go run ./examples/confidential_asset/register          # no CGO required
CGO_ENABLED=1 go run ./examples/confidential_asset/balance
CGO_ENABLED=1 go run ./examples/confidential_asset/transfer
CGO_ENABLED=1 go run ./examples/confidential_asset/withdraw
CGO_ENABLED=1 go run ./examples/confidential_asset/deposit_chain
```

FFI linkage is covered by `CGO_ENABLED=1 go test -short ./confidentialasset/native/...` (see **Confidential asset (v2)** CI), not a separate example.

See [`examples/confidential_asset/README.md`](../examples/confidential_asset/README.md).

## Specification

- [`doc/TS_GO_MAP.md`](doc/TS_GO_MAP.md) — TypeScript ↔ Go mapping  
- [`doc/SPEC.md`](doc/SPEC.md) — Move views and entry functions

## Local FFI (development)

```text
replace github.com/aptos-labs/confidential-asset-bindings/bindings/go => ../../confidential-asset-bindings/bindings/go
```

Build the static library before `CGO_ENABLED=1 go test ./confidentialasset/...`.
