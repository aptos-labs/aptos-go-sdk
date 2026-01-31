// Package crypto provides cryptographic primitives for the Aptos blockchain.
//
// This internal package implements all cryptographic operations required by the SDK:
//   - Key generation and management
//   - Digital signatures
//   - Authentication key derivation
//   - BCS serialization of cryptographic types
//
// # Supported Key Types
//
// Ed25519:
//   - Ed25519PrivateKey: Standard Ed25519 signing key
//   - Ed25519PublicKey: Ed25519 verification key
//   - Ed25519Signature: 64-byte Ed25519 signature
//
// Secp256k1:
//   - Secp256k1PrivateKey: ECDSA signing key (Bitcoin-style)
//   - Secp256k1PublicKey: Uncompressed public key (65 bytes)
//   - Secp256k1Signature: ECDSA signature (r || s, 64 bytes)
//
// Secp256r1 (P-256):
//   - Secp256r1PrivateKey: ECDSA signing key (WebAuthn-compatible)
//   - Secp256r1PublicKey: Uncompressed public key (65 bytes)
//   - Secp256r1Signature: ECDSA signature (r || s, 64 bytes)
//
// Post-Quantum:
//   - SlhDsaPrivateKey: SPHINCS+ SLH-DSA-SHA2-128s signing key
//   - SlhDsaPublicKey: SLH-DSA public key (32 bytes)
//   - SlhDsaSignature: SLH-DSA signature (7856 bytes)
//
// # Authentication Schemes
//
// SingleKey (scheme 0x02):
//   - AnyPublicKey wraps any supported key type
//   - AnySignature wraps the corresponding signature type
//   - Used for modern single-signer accounts
//
// MultiKey (scheme 0x03):
//   - Combines multiple AnyPublicKey instances
//   - K-of-N threshold signatures
//   - Supports heterogeneous key types
//
// Legacy schemes:
//   - Ed25519Scheme (0x00): Legacy Ed25519 accounts
//   - MultiEd25519Scheme (0x01): Legacy multi-sig Ed25519
//
// # WebAuthn Support
//
// The package supports WebAuthn/Passkey authentication:
//   - PartialAuthenticatorAssertionResponse: WebAuthn assertion
//   - AssertionSignature: Wraps Secp256r1 signature
//   - Verification of client data and authenticator data
//
// # Keyless Authentication
//
// Types for OIDC-based keyless accounts:
//   - KeylessPublicKey: OIDC identity commitment
//   - FederatedKeylessPublicKey: With JWK address
//   - KeylessSignature: ZK proof or OpenID signature
//
// # Thread Safety
//
// Private key types (Ed25519PrivateKey, Secp256k1PrivateKey, Secp256r1PrivateKey)
// are thread-safe. Cached public keys and authentication keys are protected by
// sync.RWMutex using double-checked locking.
//
// Public key and signature types are immutable after creation and safe to share.
//
// # Security Considerations
//
//   - Private keys are redacted in String() methods to prevent accidental logging
//   - Signature malleability is prevented by enforcing low-s values
//   - WebAuthn uses constant-time comparison for challenge verification
//   - All pooled resources clear sensitive data before reuse
package crypto
