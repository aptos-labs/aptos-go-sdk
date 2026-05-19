# Confidential asset examples (Go)

在 **aptos-go-sdk v2** 模块内（`v2/examples/confidential_asset/`）。链上能力见 `github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset`；解密与 range proof 见 **`confidentialasset/native`**（需 CGO + FFI）。

## CGO 要求

| 示例 | CGO |
|------|-----|
| **`register`** | 否（仅 sigma 注册 + 可选 deposit） |
| **`balance`**, **`transfer`**, **`withdraw`**, **`deposit_chain`** | 是（import `native`） |

`native` 在 `CGO_ENABLED=0` 时 **无法编译**；根包 `confidentialasset` 在无 CGO 时可编。

**不自动领水**：请自备公开 APT（FA metadata `0xa`）。**`deposit_chain`** 会检查公开余额至少约 **0.01 APT**。

环境变量：**`APTOS_NETWORK`**、**`APTOS_NODE_URL`**、**`APTOS_PRIVATE_KEY`** 或 **`FIXED_ED25519_PRIVATE_KEY`**。

从 **`v2/`** 执行：

```bash
cd v2
go run ./examples/confidential_asset/register
CGO_ENABLED=1 go run ./examples/confidential_asset/balance
CGO_ENABLED=1 go run ./examples/confidential_asset/transfer
CGO_ENABLED=1 go run ./examples/confidential_asset/withdraw
CGO_ENABLED=1 go run ./examples/confidential_asset/deposit_chain
```

| 命令 | 作用 |
|------|------|
| **`balance`** | 公开 FA APT + 解密机密 available/pending（`native.GetBalance`） |
| **`register`** | `register_raw` + 可选 deposit |
| **`transfer`** | `confidential_transfer_raw` |
| **`withdraw`** | `withdraw_to_raw` |
| **`deposit_chain`** | deposit → normalize → rollover → 打印余额 |

## 先决条件（FFI 示例）

构建 [`confidential-asset-bindings`](https://github.com/aptos-labs/confidential-asset-bindings) 的 **`aptos_confidential_asset_ffi`**，见 bindings `bindings/go/README.md`。`v2/go.mod` 可 `replace` 到 sibling checkout。FFI 链路 smoke 在 **`go test -short ./confidentialasset/native/...`**（`CGO_ENABLED=1`）中运行，无单独示例命令。

```bash
cd confidential-asset-bindings/rust
cargo build -p aptos_confidential_asset_ffi --release
cd ../../aptos-go-sdk/v2
CGO_ENABLED=1 go test -short ./confidentialasset/native/...
```

## 与 TypeScript 的差异

见 [`confidentialasset/README.md`](../../confidentialasset/README.md) 与 [`doc/SPEC.md`](../../confidentialasset/doc/SPEC.md)。
