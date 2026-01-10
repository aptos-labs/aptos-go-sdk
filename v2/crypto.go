// crypto.go re-exports cryptographic types from internal/crypto for public use.
package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
)

// TransactionSigner combines Signer with an account address for transaction building.
// Use this interface when you need both signing capability and an address.
//
// Most account types implement this interface. If you have just a Signer,
// you can create an Account to get an address.
type TransactionSigner interface {
	Signer

	// Address returns the account address for this signer.
	Address() AccountAddress
}

// Re-export crypto types for public API.
type (
	// Signer signs transactions and produces AccountAuthenticators.
	// This is the primary interface for transaction signing.
	//
	// Implementations:
	//   - [Ed25519PrivateKey]: Direct Ed25519 signing (legacy scheme)
	//   - [SingleSigner]: Wraps any MessageSigner for SingleKey scheme
	Signer = crypto.Signer

	// MessageSigner signs raw messages without producing an authenticator.
	// Used by underlying key types and wrapped by SingleSigner.
	MessageSigner = crypto.MessageSigner

	// PublicKey is a public key that can verify signatures and derive addresses.
	PublicKey = crypto.PublicKey

	// VerifyingKey can verify signatures but may not derive addresses directly.
	VerifyingKey = crypto.VerifyingKey

	// Signature is a cryptographic signature.
	Signature = crypto.Signature

	// CryptoMaterial provides common serialization methods for crypto objects.
	CryptoMaterial = crypto.CryptoMaterial

	// AuthenticationKey is a 32-byte key derived from a public key.
	// It identifies how an account is authenticated.
	AuthenticationKey = crypto.AuthenticationKey

	// AccountAuthenticator wraps different authenticator types for transactions.
	AccountAuthenticator = crypto.AccountAuthenticator

	// AccountAuthenticatorType identifies the type of authenticator.
	AccountAuthenticatorType = crypto.AccountAuthenticatorType
)

// Re-export Ed25519 types.
type (
	// Ed25519PrivateKey is an Ed25519 private key.
	// It implements Signer directly using the legacy Ed25519Scheme.
	Ed25519PrivateKey = crypto.Ed25519PrivateKey

	// Ed25519PublicKey is an Ed25519 public key for signature verification.
	Ed25519PublicKey = crypto.Ed25519PublicKey

	// Ed25519Signature is an Ed25519 signature.
	Ed25519Signature = crypto.Ed25519Signature

	// Ed25519Authenticator combines an Ed25519 public key and signature.
	Ed25519Authenticator = crypto.Ed25519Authenticator
)

// Re-export Secp256k1 types.
type (
	// Secp256k1PrivateKey is a Secp256k1 private key.
	// Must be wrapped with SingleSigner for transaction signing.
	Secp256k1PrivateKey = crypto.Secp256k1PrivateKey

	// Secp256k1PublicKey is a Secp256k1 public key.
	Secp256k1PublicKey = crypto.Secp256k1PublicKey

	// Secp256k1Signature is a Secp256k1 signature.
	Secp256k1Signature = crypto.Secp256k1Signature
)

// Re-export SingleKey types for unified key handling.
type (
	// SingleSigner wraps a MessageSigner for transaction signing.
	// Use this to wrap Secp256k1 keys or any other MessageSigner.
	SingleSigner = crypto.SingleSigner

	// AnyPublicKey wraps different public key types for SingleKey scheme.
	AnyPublicKey = crypto.AnyPublicKey

	// AnySignature wraps different signature types for SingleKey scheme.
	AnySignature = crypto.AnySignature

	// SingleKeyAuthenticator combines an AnyPublicKey and AnySignature.
	SingleKeyAuthenticator = crypto.SingleKeyAuthenticator
)

// Re-export MultiEd25519 types.
type (
	// MultiEd25519PublicKey is a multi-sig Ed25519 public key.
	MultiEd25519PublicKey = crypto.MultiEd25519PublicKey

	// MultiEd25519Signature is a multi-sig Ed25519 signature.
	MultiEd25519Signature = crypto.MultiEd25519Signature

	// MultiEd25519Authenticator combines a MultiEd25519 key and signature.
	MultiEd25519Authenticator = crypto.MultiEd25519Authenticator
)

// Re-export MultiKey types.
type (
	// MultiKey is a multi-sig key that can contain mixed key types.
	MultiKey = crypto.MultiKey

	// MultiKeySignature is a signature for MultiKey.
	MultiKeySignature = crypto.MultiKeySignature

	// MultiKeyAuthenticator combines a MultiKey and signature.
	MultiKeyAuthenticator = crypto.MultiKeyAuthenticator
)

// Authenticator type constants.
const (
	AccountAuthenticatorEd25519      = crypto.AccountAuthenticatorEd25519
	AccountAuthenticatorMultiEd25519 = crypto.AccountAuthenticatorMultiEd25519
	AccountAuthenticatorSingleSender = crypto.AccountAuthenticatorSingleSender
	AccountAuthenticatorMultiKey     = crypto.AccountAuthenticatorMultiKey
)

// Derive scheme constants.
const (
	Ed25519Scheme         = crypto.Ed25519Scheme
	MultiEd25519Scheme    = crypto.MultiEd25519Scheme
	SingleKeyScheme       = crypto.SingleKeyScheme
	MultiKeyScheme        = crypto.MultiKeyScheme
	DeriveObjectScheme    = crypto.DeriveObjectScheme
	NamedObjectScheme     = crypto.NamedObjectScheme
	ResourceAccountScheme = crypto.ResourceAccountScheme
)

// Key generation functions.

// GenerateEd25519PrivateKey generates a new random Ed25519 private key.
var GenerateEd25519PrivateKey = crypto.GenerateEd25519PrivateKey

// GenerateSecp256k1Key generates a new random Secp256k1 private key.
var GenerateSecp256k1Key = crypto.GenerateSecp256k1Key

// NewSingleSigner wraps a MessageSigner for transaction signing.
// Use this for Secp256k1 keys or when you want to use SingleKeyScheme.
var NewSingleSigner = crypto.NewSingleSigner

// NoAccountAuthenticator creates an authenticator for simulation.
var NoAccountAuthenticator = crypto.NoAccountAuthenticator

// AIP-80 private key formatting.
var (
	// FormatPrivateKey formats a private key as an AIP-80 compliant string.
	FormatPrivateKey = crypto.FormatPrivateKey

	// ParsePrivateKey parses an AIP-80 formatted private key string.
	ParsePrivateKey = crypto.ParsePrivateKey
)
