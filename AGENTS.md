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
