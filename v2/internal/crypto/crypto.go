// Package crypto provides cryptographic primitives for the Aptos blockchain.
//
// This package supports multiple signature schemes used by Aptos:
//   - Ed25519: The default signature scheme
//   - Secp256k1: Used for Ethereum compatibility
//   - MultiEd25519: Off-chain multi-sig with Ed25519 keys
//   - MultiKey: Off-chain multi-sig with mixed key types
//
// # Key Interfaces
//
// The package defines several key interfaces:
//   - [Signer]: Signs transactions and produces AccountAuthenticators
//   - [MessageSigner]: Signs raw messages (used by underlying key types)
//   - [PublicKey]: A public key that can verify signatures and derive addresses
//   - [VerifyingKey]: A key that can only verify signatures
//   - [Signature]: A signature that can be verified
//   - [CryptoMaterial]: Common interface for hex/bytes conversion
//
// # Usage
//
//	// Generate a new Ed25519 key
//	key, err := crypto.GenerateEd25519PrivateKey()
//
//	// Sign a transaction
//	auth, err := key.Sign(txnBytes)
//
//	// Get the account address
//	address := key.AuthKey().AccountAddress()
//
// # Single Signer vs Direct Keys
//
// Ed25519PrivateKey implements Signer directly and uses the legacy Ed25519Scheme.
// For newer key types like Secp256k1, use SingleSigner which wraps any MessageSigner
// and uses the SingleKeyScheme:
//
//	secpKey, _ := crypto.GenerateSecp256k1Key()
//	signer := crypto.NewSingleSigner(secpKey)
//	auth, _ := signer.Sign(txnBytes)
package crypto

import "github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"

// Signer signs transactions and produces AccountAuthenticators.
// This is the primary interface for transaction signing.
type Signer interface {
	// Sign signs a message and returns an AccountAuthenticator.
	Sign(msg []byte) (*AccountAuthenticator, error)

	// SignMessage signs a message and returns the raw Signature.
	SignMessage(msg []byte) (Signature, error)

	// SimulationAuthenticator creates an authenticator for simulation.
	// Simulation authenticators use empty signatures.
	SimulationAuthenticator() *AccountAuthenticator

	// AuthKey returns the AuthenticationKey for this signer.
	AuthKey() *AuthenticationKey

	// PubKey returns the PublicKey for signature verification.
	PubKey() PublicKey
}

// MessageSigner signs raw messages. This is implemented by underlying
// key types and wrapped by SingleSigner for transaction signing.
//
// Note: MessageSigner does not implement Signer because some key types
// (like Secp256k1) cannot be used directly for account authentication
// and must be wrapped with SingleSigner.
type MessageSigner interface {
	// SignMessage signs a message and returns the raw Signature.
	SignMessage(msg []byte) (Signature, error)

	// EmptySignature returns an empty signature for simulation.
	EmptySignature() Signature

	// VerifyingKey returns the public key for verification.
	VerifyingKey() VerifyingKey
}

// PublicKey is a public key that can verify signatures and derive addresses.
// All PublicKeys are also VerifyingKeys.
type PublicKey interface {
	VerifyingKey

	// AuthKey returns the AuthenticationKey for this public key.
	AuthKey() *AuthenticationKey

	// Scheme returns the DeriveScheme used for address derivation.
	Scheme() DeriveScheme
}

// VerifyingKey verifies signatures. Not all verifying keys can be used
// directly as PublicKeys (e.g., Secp256k1PublicKey needs to be wrapped).
type VerifyingKey interface {
	bcs.Struct
	CryptoMaterial

	// Verify returns true if the signature is valid for the message.
	Verify(msg []byte, sig Signature) bool
}

// Signature is a cryptographic signature that can be serialized.
type Signature interface {
	bcs.Struct
	CryptoMaterial
}

// CryptoMaterial provides common serialization methods for cryptographic objects.
type CryptoMaterial interface {
	// Bytes returns the raw byte representation.
	Bytes() []byte

	// FromBytes loads from raw bytes.
	FromBytes(input []byte) error

	// ToHex returns the hex representation with "0x" prefix.
	ToHex() string

	// FromHex parses a hex string with optional "0x" prefix.
	FromHex(input string) error
}
