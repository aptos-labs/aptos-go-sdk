# Aptos Go SDK v2 Changelog

All notable changes to the Aptos Go SDK v2 will be captured in this file. This changelog is written by hand for now. It
adheres to the format set out by [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# Unreleased

# v2.2.0 (7/2/2026)

- [`Feature`] Add BIP-39 mnemonic and BIP-44 derivation path support for Ed25519 accounts
  - `account.FromMnemonic` and `account.FromDerivationPath` for wallet import
  - Optional BIP-39 passphrase and SingleKey authentication scheme via `account.DerivationConfig`
  - `aptos.ValidateMnemonic`, `aptos.DefaultDerivationPath`, and `aptos.Ed25519PrivateKeyFromDerivationPath`
  - SLIP-0010 hardened derivation on the standard Aptos path `m/44'/637'/0'/0'/0'`
- [`Feature`] Add automatic retry middleware to the node client with configurable backoff, rate-limit (429) handling, and `MaxRetries`
- [`Feature`] Support legacy `MultiEd25519` transaction authenticators alongside fee-payer and multi-agent flows
- [`Fix`] Fix AIP-80 private key export for `SingleSigner`-wrapped keys
- [`Fix`] Guard `MultiEd25519TransactionAuthenticator` against nil sender authenticator
- [`Fix`] Validate inner authenticator type in `MultiEd25519` BCS marshaling
- [`Fix`] Make `FakeClient` view-result stubs return defensive copies to prevent mutation
- [`Fix`] Honor explicit `MaxRetries == 0` in retry middleware to disable retries
- [`Fix`] Keep `MaxRetries` as a strict cap on error/status retries
- [`Fix`] Fix 32-bit overflow in BCS length checks and `IntToU32` conversion on 32-bit systems
- [`Perf`] Use a stoppable timer for retry backoff to avoid goroutine leaks on context cancellation
- [`Security`] Upgrade Go toolchain from go1.25.0 to go1.25.10, resolving 18 stdlib CVEs
- [`Security`] Upgrade `golang.org/x/crypto` v0.46.0 → v0.52.0
- [`Security`] Upgrade `github.com/decred/dcrd/dcrec/secp256k1/v4` v4.4.0 → v4.4.1
- [`Dependency`] Upgrade `filippo.io/edwards25519` v1.1.1 → v1.2.0
- [`Dependency`] Upgrade `github.com/aptos-labs/aptos-go-sdk` v1.13.0 → v1.14.0
- [`Dependency`] Upgrade `github.com/hasura/go-graphql-client` v0.14.4 → v0.16.0
- [`Dependency`] Upgrade `golang.org/x/sys` v0.42.0 → v0.45.0

# v2.1.0 (5/21/2026)

- [`Fix`] Fix critical signing bug in transaction authentication
- [`Fix`] Fix entry-function argument BCS encoding and ANS payloads
- [`Fix`] Stringify `uint64` view function arguments correctly
- [`Fix`] Increase default max gas amount by 10x from 200,000 to 2,000,000
- [`Security`] Upgrade OpenTelemetry SDK from 1.39.0 to 1.43.0

# v2.0.0 (2/25/2026)

## New Features
- [`Feature`] Add Keyless authentication support (OpenID-based)
  - `KeylessPublicKey`: Contains issuer (`iss_val`) and identity commitment (`idc`)
  - `FederatedKeylessPublicKey`: For accounts using on-chain JWK addresses
  - `KeylessSignature`: Contains ephemeral certificate, JWT header, expiry, ephemeral key/signature
  - `EphemeralCertificate`: Either `ZeroKnowledgeSig` (Groth16 proof) or `OpenIdSig`
  - `EphemeralPublicKey` and `EphemeralSignature` for short-lived key management
  - Full BCS serialization/deserialization for all types
  - Integration with `AnyPublicKey` (variants 3, 4) and `AnySignature` (variant 3)
- [`Feature`] Add WebAuthn signature support
  - `PartialAuthenticatorAssertionResponse` and `AssertionSignature` types
  - Properly handles WebAuthn's SHA-256 verification flow
  - Extracts and validates challenge from `clientDataJSON`
  - Full BCS serialization/deserialization support
  - Integration with `AnySignature` (variant 2)
  - Security hardening: constant-time challenge comparison, bounds validation
- [`Feature`] Add Secp256r1 (P-256/prime256v1) signature support
  - `Secp256r1PrivateKey`, `Secp256r1PublicKey`, `Secp256r1Signature` types
  - Full integration with `SingleSigner` and `AnyPublicKey`
  - Uses Go standard library (`crypto/ecdsa`, `crypto/elliptic`)
  - Commonly used with WebAuthn/passkeys
- [`Feature`] Add SLH-DSA-SHA2-128s post-quantum signature support
  - `SlhDsaPrivateKey`, `SlhDsaPublicKey`, `SlhDsaSignature` types
  - Full integration with `SingleSigner` and `AnyPublicKey`/`AnySignature`
  - Uses Cloudflare CIRCL library (FIPS 205 compliant)
  - 32-byte public keys, 64-byte private keys, 7,856-byte signatures
- [`Feature`] Add `codegen` package for generating type-safe Go bindings from Move ABIs
  - Generate Go structs from Move struct definitions
  - Generate entry function wrappers with automatic argument encoding
  - Generate view function helpers with typed return values
  - New `aptosgen` CLI tool for code generation
- [`Feature`] Add OpenTelemetry `telemetry` package for tracing and metrics instrumentation
  - Distributed tracing for all API calls
  - Request duration, count, and error metrics
  - Customizable tracer and meter providers
- [`Feature`] Add ANS (Aptos Names Service) client for `.apt` domain management
  - Name resolution (name → address)
  - Reverse lookup (address → primary name)
  - Registration, renewal, and subdomain management payloads
  - Support for mainnet and testnet router addresses
- [`Feature`] Support signed integers (i8, i16, i32, i64, i128, i256) as arguments

## Security
- [`Security`] Add thread-safety to all private key types with proper mutex protection
  - `Ed25519PrivateKey`, `Secp256k1PrivateKey`, `Secp256r1PrivateKey` now use `sync.RWMutex`
  - `SingleSigner` now uses `sync.RWMutex` for cached values
  - Double-checked locking pattern prevents race conditions (documented)
- [`Security`] Prevent sensitive data leakage in pooled resources
  - `Serializer.Reset()` now zeroes buffer memory before reuse
  - Buffer pools validate both length AND capacity before accepting returns
  - Prevents cryptographic data from leaking between pooled operations
- [`Security`] Redact private keys in `String()` methods to prevent accidental logging
  - `Ed25519PrivateKey.String()` and `Secp256k1PrivateKey.String()` now return redacted output
  - Use `ToAIP80()` to get actual private key string when needed
- [`Security`] Add bounds checking in `MultiKey.Verify()` to prevent index out-of-bounds panics
  - Validates bitmap indices against public key array length
  - Checks for duplicate indices and bitmap/signature count consistency
- [`Security`] Add weak key rejection in Secp256r1 private key parsing
  - Rejects d=1 and d=n-1 as weak keys
- [`Security`] Add bounds checking in `Secp256r1Signature.setRS()` to prevent buffer underflow
- [`Security`] Fix AIP-80 parsing to validate split results before accessing array elements
- [`Security`] Improve `MultiKeyBitmap.Indices()` to use `numBits` as iteration limit

## Bug Fixes
- [`Fix`][`Breaking`] Fix `MultiKeyBitmap` serialization to comply with Aptos BitVec specification
  - Bitmap now correctly serializes `num_bits` as u16 little-endian before the bytes
  - `NewMultiKeySignature` now requires `numKeys` parameter to ensure correct serialization
  - Added `NewMultiKeyBitmap(numKeys)` and `SetNumKeys(numKeys)` helpers
- [`Fix`] Add missing `AnyPublicKey` variants: `Secp256r1` (2), `Keyless` (3), `FederatedKeyless` (4), `SlhDsaSha2_128s` (5)
- [`Fix`] Add missing `AnySignature` variants: `WebAuthn` (2), `Keyless` (3), `SlhDsaSha2_128s` (4)

## Performance
- [`Perf`] Reduce allocations in BCS serialization
  - Use stack-allocated arrays for U16/U32/U64/I16/I32/I64 serialization
  - Add `sync.Pool` for `Serializer` reuse in hot paths
- [`Perf`] Add `sync.Pool` for SHA3-256 hashers in `util.Sha3256Hash`
- [`Perf`] Pre-allocate exact buffer sizes in `Secp256k1Signature.Bytes()`
- [`Perf`] Optimize WebAuthn verification data generation
- [`Perf`] Cache public keys in private key structs (Ed25519, Secp256k1, Secp256r1)
  - Avoid repeated derivation on `PubKey()` / `VerifyingKey()` calls
  - Thread-safe with proper mutex protection
- [`Perf`] Cache authentication keys in `Ed25519PrivateKey` and `SingleSigner`
  - Avoid repeated SHA3-256 hashing on `AuthKey()` calls
- [`Perf`] Optimize `AccountAddress.String()` for special addresses
  - Pre-computed string table for addresses 0x0-0xf avoids `fmt.Sprintf`
- [`Perf`] Optimize `ParseAddress` to avoid string concatenation for odd-length hex
- [`Perf`] Optimize `BytesToHex` with direct encoding (avoids `hex.EncodeToString` overhead)
- [`Perf`] Add byte buffer pools (`GetBuffer32`, `GetBuffer64`) for common sizes

## Improvements
- [`Doc`] Comprehensive documentation improvements across all packages
  - Enhanced main package docs with examples
  - Updated README with feature highlights and usage examples
  - Improved godoc comments for all public APIs
  - Added doc.go files for all internal packages (bcs, crypto, http, types, util)
  - Internal crypto package documents all key types, authentication schemes, and security considerations
  - Internal BCS package documents serialization patterns and performance optimizations
  - Internal HTTP package documents middleware composition patterns
