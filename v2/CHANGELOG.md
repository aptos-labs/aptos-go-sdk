# Aptos Go SDK v2 Changelog

All notable changes to the Aptos Go SDK v2 will be captured in this file. This changelog is written by hand for now. It
adheres to the format set out by [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# Unreleased

- [`Fix`] Increase default max gas amount by 10x from 200,000 to 2,000,000

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
