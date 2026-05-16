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
- **CGO + FFI** for decrypt and proof generation: build [`confidential-asset-bindings`](https://github.com/aptos-labs/confidential-asset-bindings) Rust workspace member **`aptos_confidential_asset_ffi`** → static library `libaptos_confidential_asset_ffi.a` (see bindings `bindings/go/aptosconfidential` CGO flags). Go package `aptosconfidential` links that archive on Linux amd64.
- Set `WithRESTBaseURL` to your fullnode base including `/v1` when using simulated-gas submit helpers

## CI and local testing (what actually runs)

| 场景 | 命令 / 位置 | 覆盖内容 |
|------|----------------|----------|
| 仓库根 **Go Unit Tests** workflow | `go test ./...`（仅 v1 模块） | **不包含** `v2/confidentialasset` |
| **Confidential asset (v2)** workflow | `confidentialasset-nocgo`：`CGO_ENABLED=0 go test -short ./confidentialasset/...` | `cipherparse`、`movearg`、`sigbcs`、`views` 单测；**`!cgo` stub** 返回 `ErrCGODisabled` 契约 |
| 同上 | `confidentialasset-cgo-ffi`：`cargo build -p aptos_confidential_asset_ffi --release` 后 `CGO_ENABLED=1 go test -short ./confidentialasset/...` | **FFI smoke**、`rangeproof` CGO 往返等 |
| **Code Coverage** workflow | `v2/` 下 `go test -race ./...`，`CGO_ENABLED=1`，且先构建同一 FFI | 全 v2 模块覆盖率（含机密子包 CGO 路径） |
| 可选链上只读 | `APTOS_CONFIDENTIAL_INTEGRATION=1`，无 `-short`，跑 `TestIntegration_Confidential*` | 对公共网络的 `HasUserRegistered` 等（**默认 CI 不启用**） |

临时 CI：checkout bindings 分支由 workflow 内 `CONFIDENTIAL_ASSET_BINDINGS_REF` 固定（与 codecov 一致）；**正式发布前**应删除该临时逻辑并改为已发布的 `bindings/go` 版本（见 workflow 内 TODO 注释）。

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
