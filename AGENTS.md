# AGENTS.md

## Cursor Cloud specific instructions

This is a Go SDK (library) for the Aptos blockchain. There are two separate Go modules:

- **v1** (root `/`): `github.com/aptos-labs/aptos-go-sdk`
- **v2** (`/v2/`): `github.com/aptos-labs/aptos-go-sdk/v2`

Each has its own `go.mod` and must be managed independently (`go mod tidy`, `go test`, etc.). Both modules currently declare Go 1.25.0 with toolchain `go1.25.10`.

### Running tests

- **Unit tests only** (no network needed): `go test -short ./...` (run from root for v1, from `v2/` for v2)
- **Full v1 tests** require a local Aptos testnet via `aptos node run-localnet --with-indexer-api` (Aptos CLI and Docker are not pre-installed in cloud VMs)
- **v2 integration tests** hit public Aptos networks (no local infra needed): from `v2/`, run `go test -v -run "TestIntegration_" -timeout 120s ./...`
- **v2 faucet tests** require env var: from `v2/`, run `APTOS_TEST_FAUCET=1 go test -v -run "TestIntegrationFaucet_" ./...`

### Lint and format

Use the commands in `CLAUDE.md`/`mise.toml`:

```bash
gofumpt -l -w .
golangci-lint run
```

Root `golangci-lint run` passes. `golangci-lint run` from `v2/` currently reports a pre-existing baseline of lint warnings; these are not regressions.

### Build

Standard `go build ./...` from root and from `v2/`.

### Gotchas

- Cloud VMs may not have `mise`; the startup update script installs `gofumpt` and `golangci-lint` into `$(go env GOPATH)/bin`, so export that directory on `PATH` (or call the tools by full path) before lint/format commands.
- The root `examples/transfer_coin` demo submits two devnet transfers after one faucet funding. If devnet gas is high, the second transfer can fail with `INSUFFICIENT_BALANCE_FOR_TRANSACTION_FEE`; for a smoke test, use a one-transfer variant or increase funding in a throwaway copy.
- The v2 module's `go.mod` references v1 as a published module (`github.com/aptos-labs/aptos-go-sdk v1.13.0`), not a local replace. Changes to v1 types do not automatically propagate to v2.
- v2 network presets are `aptos.Testnet`, `aptos.Mainnet`, `aptos.Devnet`, `aptos.Localnet` (not `TestnetConfig`).
