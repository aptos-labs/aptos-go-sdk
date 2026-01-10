# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Testing
- Run all tests: `go test ./...`
- Run tests with race detection: `go test -race ./...`
- Run a specific test: `go test ./... -run TestName`
- Run tests in specific directories: `go test ./examples/...`

### Code Quality
- Format code: `gofumpt -l -w .`
- Run linter: `golangci-lint run`
- All examples run as unit tests in CI to ensure they work correctly

### Build
- Standard Go build commands work: `go build`
- Install dependencies: `go mod tidy`

## Code Architecture

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