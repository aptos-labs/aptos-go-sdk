# Confidential asset examples (Go)

在 **aptos-go-sdk v2** 模块内（`v2/examples/confidential_asset/`）。链上与解密能力见 SDK：`github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset`。

## CGO-only

各子目录为 **独立 `main` 包**，源文件均带 `//go:build cgo`。需 **`CGO_ENABLED=1`** 且已按下文构建 **`libaptos_confidential_asset_ffi`**。`CGO_ENABLED=0` 时这些包不会参与 `go test ./...`。

**不自动领水**：示例不调用网络水龙头；请你在目标网络上自行给发送地址准备足够的 **公开 APT**（FA metadata `0xa`）。**`deposit_chain`** 在链上操作前会检查公开余额是否至少约 **0.01 APT**（`1_000_000` octas），不足则退出并提示充值。

每个 `main.go` 内联读取环境变量（无共享 `support` / `exampleenv` 包）：**`APTOS_NETWORK`**（默认 testnet）、**`APTOS_NODE_URL`**（可选，覆盖 fullnode REST `/v1` 基址）、**`APTOS_PRIVATE_KEY`** 或 **`FIXED_ED25519_PRIVATE_KEY`**。

从 **`v2/`** 目录执行（与 `go run ./examples/transfer_coin` 相同风格）：

```bash
cd v2
CGO_ENABLED=1 go run ./examples/confidential_asset/balance
CGO_ENABLED=1 go run ./examples/confidential_asset/register
CGO_ENABLED=1 go run ./examples/confidential_asset/transfer
CGO_ENABLED=1 go run ./examples/confidential_asset/deposit_chain
CGO_ENABLED=1 go run ./examples/confidential_asset/withdraw
CGO_ENABLED=1 go run ./examples/confidential_asset/ffismoke
```

| 命令 | 作用 |
|------|------|
| **`balance`** | 打印公开 FA APT（0xa）余额 + 解密后的机密 **available** / **pending**（与 TS `getBalance` 一致）。需 `APTOS_PRIVATE_KEY` 或 `FIXED_ED25519_PRIVATE_KEY`；可选 `TWISTED_PRIVATE_KEY_HEX`。未注册机密余额时会提示先跑 **`register`**。 |
| **`register`** | 若无 `has_confidential_store` 则 **`register_raw`**，然后默认 **`deposit`**（默认 **800000** octas，可用 **`CONFIDENTIAL_DEPOSIT_OCTAS`** 覆盖）。**`SKIP_CONFIDENTIAL_DEPOSIT=1`** 只注册不存。 |
| **`transfer`** | 与 TS **`confidentialAsset.transfer`** 一致：提交 **`confidential_transfer_raw`**。必填 **`APTOS_SEND_TO`**；**`APTOS_SEND_OCTAS`** 默认 **1**。双方须已注册该 token 的机密 store；可选 **`TWISTED_PRIVATE_KEY_HEX`**。 |
| **`withdraw`** | 与 TS **`confidentialAsset.withdraw`** 一致：提交 **`withdraw_to_raw`**（机密 **available** → 收款方 **公开** FA）。必填 **`APTOS_WITHDRAW_OCTAS`**；**`APTOS_WITHDRAW_TO`** 可选（默认收款方为 signer 的公开 APT）；可选 **`TWISTED_PRIVATE_KEY_HEX`**。需已有机密 store 且 available 余额充足。 |
| **`deposit_chain`** | **deposit** → 若未 normalized 则 **NormalizeBalance**（与 TS `normalizeBalance` 等价）→ **RolloverPendingBalance** → 打印解密余额。需已有机密 store；需事先自备公开 APT。可选：`TWISTED_PRIVATE_KEY_HEX`、`SKIP_CONFIDENTIAL_BALANCE_READ`、`ATTEMPT_ROLLOVER_IF_NOT_NORMALIZED`（调试）。 |
| **`ffismoke`** | 无链、无私钥：跑一轮 **`BatchRangeProof` → `BatchVerifyProof`** 验证 FFI 链接。**`SKIP_CONFIDENTIAL_BINDINGS=1`** 跳过。 |

## 与 TypeScript `@aptos-labs/confidential-asset` 的差异

Go 示例与 [`confidentialasset`](../../confidentialasset) 对齐 TS 的 **单笔** 上链接路；下列能力在 TS 中存在或更强，Go 侧尚未等同：

- **编排**：TS 有 **`withdrawWithTotalBalance`** / **`transferWithTotalBalance`**（多笔、可涉及 pending 等）；Go 仅有单笔 **`Withdraw`** / **`Transfer`**，无同等自动编排。
- **Deposit**：TS **`deposit`** 可选 **`recipient`**（代他人公开余额转入机密）；Go **`Deposit`** 仅 `token + amount`（见 SDK `txsubmit.go`）。
- **Transfer 扩展**：TS 支持 **`memo`**、**`additionalAuditorEncryptionKeys`**；Go 内部有 `transferWithMemo` 但未导出，且无额外 auditor 参数。
- **Indexer**：TS **`getActivities`** 为完整 GraphQL；Go **`GetActivities`** 仍为占位实现。

命名上 TS 的 **`getAssetAuditorEncryptionKey`** 与 Go **`GetEffectiveAuditorEncryptionKeyHex`** 等 view 可大致对应，返回形态不同。

## 先决条件（FFI）

与 [`confidential-asset-bindings/bindings/go/README.md`](https://github.com/aptos-labs/confidential-asset-bindings/blob/main/bindings/go/README.md)：`v2/go.mod` 的 **`replace`** 指向 sibling **`confidential-asset-bindings/bindings/go`**，并在该仓库中产出 **`rust/target/<triple>/release/libaptos_confidential_asset_ffi.a`**。

```bash
cd confidential-asset-bindings/rust
cargo build -p aptos_confidential_asset_ffi --release
```

或使用 release 安装脚本（见 bindings 仓库 README）。

## `go test`

FFI 烟测与 `Solver` 生命周期测试在 SDK 包内（`confidentialasset`，需 CGO）：

```bash
cd v2 && CGO_ENABLED=1 go test -short ./confidentialasset/...
```

另可编译校验示例子包（无单元测试时仅编译）：

```bash
CGO_ENABLED=1 go test -short ./examples/confidential_asset/...
```
