# Confidential asset examples (Go)

Runnable examples for confidential fungible assets, located in `v2/examples/confidential_asset/`.
On-chain capabilities are in `github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset`;
decryption and range proofs live in **`confidentialasset/native`** (requires CGO + FFI).

## CGO requirements

| Example | Requires CGO |
|---------|--------------|
| **`register`** | No (sigma registration + optional deposit only) |
| **`balance`**, **`transfer`**, **`withdraw`**, **`deposit_chain`** | Yes (imports `native`) |

`native` **cannot compile** with `CGO_ENABLED=0`; the root `confidentialasset` package compiles without CGO.

**No automatic faucet**: fund the account with public APT (FA metadata `0xa`) beforehand.
**`deposit_chain`** checks that the public balance is at least ~0.01 APT.

Environment variables: **`APTOS_NETWORK`**, **`APTOS_NODE_URL`**, **`APTOS_PRIVATE_KEY`** or **`FIXED_ED25519_PRIVATE_KEY`**.

Run from **`v2/`**:

```bash
cd v2
go run ./examples/confidential_asset/register
CGO_ENABLED=1 go run ./examples/confidential_asset/balance
CGO_ENABLED=1 go run ./examples/confidential_asset/transfer
CGO_ENABLED=1 go run ./examples/confidential_asset/withdraw
CGO_ENABLED=1 go run ./examples/confidential_asset/deposit_chain
```

| Command | What it does |
|---------|--------------|
| **`balance`** | Public FA APT + decrypted confidential available/pending (`native.GetBalance`) |
| **`register`** | `register_raw` + optional deposit |
| **`transfer`** | `confidential_transfer_raw` |
| **`withdraw`** | `withdraw_to_raw` |
| **`deposit_chain`** | deposit → normalize → rollover → print balance |

## Prerequisites (FFI examples)

Build **`aptos_confidential_asset_ffi`** from
[`confidential-asset-bindings`](https://github.com/aptos-labs/confidential-asset-bindings);
see `bindings/go/README.md`. `v2/go.mod` can `replace` to a sibling checkout.
FFI smoke runs in **`go test -short ./confidentialasset/native/...`** (`CGO_ENABLED=1`);
there is no separate example command.

```bash
cd confidential-asset-bindings/rust
cargo build -p aptos_confidential_asset_ffi --release
cd ../../aptos-go-sdk/v2
CGO_ENABLED=1 go test -short ./confidentialasset/native/...
```

## Differences from TypeScript

See [`confidentialasset/README.md`](../../confidentialasset/README.md) and
[`doc/SPEC.md`](../../confidentialasset/doc/SPEC.md).
