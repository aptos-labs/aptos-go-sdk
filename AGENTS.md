# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, Cursor, Codex, etc.) when working with code in this repository.

## Overview

This is a Go SDK (library) for the Aptos blockchain. There are two separate Go modules:

- **v1** (root `/`): `github.com/aptos-labs/aptos-go-sdk`
- **v2** (`/v2/`): `github.com/aptos-labs/aptos-go-sdk/v2`

Each has its own `go.mod` and must be managed independently (`go mod tidy`, `go test`, etc.).

## Development Commands

### Testing

- **Unit tests only** (no network needed): `go test -short ./...` (run from root for v1, from `v2/` for v2)
- Run all tests: `go test ./...`
- **Full v1 tests** require a local Aptos testnet via `aptos node run-localnet --with-indexer-api` (Aptos CLI not pre-installed in Cursor Cloud VMs)
- Run tests with race detection: `go test -race ./...`
- Run a specific test: `go test ./... -run TestName`
- Run tests in specific directories: `go test ./examples/...`
- All examples run as unit tests in CI to ensure they work correctly
- **v2 integration tests** hit public Aptos networks (no local infra needed): `go test -v -run "TestIntegration_" -timeout 120s ./v2/...`
- **v2 faucet tests** require env var: `APTOS_TEST_FAUCET=1 go test -v -run "TestIntegrationFaucet_" ./v2/...`
- **v2 confidential asset** (packages `confidentialasset` + `confidentialasset/native`): root `go test ./...` from repo **root does not include v2**. **TS parity map:** `v2/confidentialasset/doc/TS_GO_MAP.md`. FFI APIs live in **`native`** (`//go:build cgo`); importing `native` with `CGO_ENABLED=0` fails at compile time. Use:
  - **CI**: workflow **Confidential asset (v2)** — job `confidentialasset-nocgo` runs `CGO_ENABLED=0 go test -short` on `./confidentialasset/...` **excluding** `native`; job `confidentialasset-cgo-ffi` builds Rust crate **`aptos_confidential_asset_ffi`** then `CGO_ENABLED=1 go test -short ./confidentialasset/...`. **Code Coverage** for v2 runs `go test ./...` under `v2/` with **`CGO_ENABLED=1`** after the same FFI build + bindings checkout.
  - **Local**: `cd v2 && pkgs=$(go list ./confidentialasset/... | grep -v '/native$' | grep -v '/internal/rangeproof$') && CGO_ENABLED=0 go test -short $pkgs`; with FFI built: `CGO_ENABLED=1 go test -short ./confidentialasset/...`.
  - **Optional live views** (not default CI): `APTOS_CONFIDENTIAL_INTEGRATION=1 go test -run TestIntegration_Confidential -count=1 ./confidentialasset/...` (omit `-short`; hits public testnet).
  - **Optional simulate gate (placeholder)**: `APTOS_CONFIDENTIAL_SIMULATE=1` enables `TestGate_ConfidentialSimulate_envOnly` (currently only checks devnet `ChainID`; extend for real simulate later).

### Code Quality (must pass before committing)

- Format code: `gofumpt -l -w .`
- Run linter: `golangci-lint run`
- **Both checks must be run and pass before creating any commit.**
- `golangci-lint` has pre-existing warnings in both v1 and v2; these are not regressions.

### Build

- Standard Go build commands work: `go build ./...` (from the repo root for v1 and from `v2/` for v2)
- Install dependencies: `go mod tidy` (per module)

## Code Architecture (v1, root module)

### Core Structure

- **Root package (aptos)**: Main SDK interface with re-exported types from internal packages
- **internal/types/**: Core blockchain types (Account, AccountAddress) - separated to avoid circular dependencies
- **internal/util/**: Shared utility functions
- **internal/testutil/**: Test helper functions
- **examples/**: Standalone runnable examples that also serve as integration tests

### Key Components

#### Client Architecture

- **NodeClient**: Low-level HTTP client for Aptos node API interactions (`nodeClient.go`)
- **Client**: High-level client wrapper combining NodeClient with indexer and faucet clients (`client.go`)
- **IndexerClient**: GraphQL client for querying blockchain data (`indexerClient.go`)
- **FungibleAssetClient**: Specialized client for fungible asset operations (`fungible_asset_client.go`)

#### Network Configuration

Pre-configured network settings available:

- `LocalnetConfig`: For local development (requires `aptos node run-localnet --with-indexer-api`)
- `DevnetConfig`: Development network (resets weekly)
- `TestnetConfig`: Stable test network
- `MainnetConfig`: Production network

#### Account Management

- **Account types**: Ed25519 (legacy and single-sender), Secp256k1
- **Address handling**: 32-byte AccountAddress with relaxed parsing
- **Signing**: Abstracted through crypto.Signer interface

#### Transaction Handling

- **Raw transactions**: Core transaction building (`rawTransaction.go`)
- **Payloads**: Various transaction payload types (script, entry function, multisig)
- **Submission**: BCS-encoded transaction submission with proper content types
- **Authentication**: Multi-signature and single signature support

### Package Organization

The SDK uses type re-exports to maintain a clean public API while avoiding circular dependencies. Core types are defined in `internal/types/` and re-exported from the main package.

### Move Integration

- **View functions**: Query Move module functions without gas costs
- **Entry functions**: Execute Move module functions on-chain
- **Script transactions**: Execute Move scripts with type arguments
- **ABI support**: Both local and remote ABI fetching for type safety

### Crypto Support

- Ed25519 signatures (legacy and consensus variants)
- Secp256k1 signatures
- Multi-signature schemes (on-chain and off-chain)
- BCS serialization for all cryptographic operations

## Gotchas

- `~/go/bin` must be on `PATH` for `gofumpt` and `golangci-lint` to work. In Cursor Cloud VMs the update script ensures this via `~/.bashrc`.
- The v2 module's `go.mod` references v1 (`github.com/aptos-labs/aptos-go-sdk`) as a published module (version pinned in `v2/go.mod`), not a local replace. Changes to v1 types do not automatically propagate to v2.
- v2 network presets are `aptos.Testnet`, `aptos.Mainnet`, `aptos.Devnet`, `aptos.Localnet` (not `TestnetConfig`).
