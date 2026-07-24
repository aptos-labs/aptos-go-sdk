# AGENTS.md

## Cursor Cloud specific instructions

This is a Go SDK (library) for the Aptos blockchain. There are two separate Go modules:

- **v1** (root `/`): `github.com/aptos-labs/aptos-go-sdk`
- **v2** (`/v2/`): `github.com/aptos-labs/aptos-go-sdk/v2`

Each has its own `go.mod` and must be managed independently (`go mod tidy`, `go test`, etc.).

### Running tests

- **Unit tests only** (no network needed): `go test -short ./...` (run from root for v1, from `v2/` for v2)
- **Full v1 tests** require a local Aptos testnet via `aptos node run-localnet --with-indexer-api` (Aptos CLI not pre-installed in cloud VMs)
- **v2 integration tests** hit public Aptos networks (no local infra needed): `go test -v -run "TestIntegration_" -timeout 120s ./v2/...`
- **v2 faucet tests** require env var: `APTOS_TEST_FAUCET=1 go test -v -run "TestIntegrationFaucet_" ./v2/...`
- **v2 confidential asset** (packages `confidentialasset` + `confidentialasset/native`): root `go test ./...` from repo **root does not include v2**. **TS parity map:** `v2/confidentialasset/doc/TS_GO_MAP.md`. FFI APIs live in **`native`** (`//go:build cgo`); importing `native` with `CGO_ENABLED=0` fails at compile time. Use:
  - **CI**: workflow **Confidential asset (v2)** — job `confidentialasset-nocgo` runs `CGO_ENABLED=0 go test -short` on `./confidentialasset/...` **excluding** `native`; job `confidentialasset-cgo-ffi` builds Rust crate **`aptos_confidential_asset_ffi`** then `CGO_ENABLED=1 go test -short ./confidentialasset/...`. **Code Coverage** for v2 runs `go test ./...` under `v2/` with **`CGO_ENABLED=1`** after the same FFI build + bindings checkout.
  - **Local**: `cd v2 && pkgs=$(go list ./confidentialasset/... | grep -v '/native$' | grep -v '/internal/rangeproof$') && CGO_ENABLED=0 go test -short $pkgs`; with FFI built: `CGO_ENABLED=1 go test -short ./confidentialasset/...`.
  - **Optional live views** (not default CI): `APTOS_CONFIDENTIAL_INTEGRATION=1 go test -run TestIntegration_Confidential -count=1 ./confidentialasset/...` (omit `-short`; hits public testnet).
  - **Optional simulate gate (placeholder)**: `APTOS_CONFIDENTIAL_SIMULATE=1` enables `TestGate_ConfidentialSimulate_envOnly` (currently only checks devnet `ChainID`; extend for real simulate later).

### Lint and format

Both must pass before committing (see `CLAUDE.md`):

```bash
gofumpt -l -w .
golangci-lint run
```

`golangci-lint` has pre-existing warnings in both v1 and v2; these are not regressions.

### Build

Standard `go build ./...` from root and from `v2/`.

### Gotchas

- `~/go/bin` must be on `PATH` for `gofumpt` and `golangci-lint` to work. The update script ensures this via `~/.bashrc`.
- The v2 module's `go.mod` references v1 as a published module (`github.com/aptos-labs/aptos-go-sdk v1.11.0`), not a local replace. Changes to v1 types do not automatically propagate to v2.
- v2 network presets are `aptos.Testnet`, `aptos.Mainnet`, `aptos.Devnet`, `aptos.Localnet` (not `TestnetConfig`).
