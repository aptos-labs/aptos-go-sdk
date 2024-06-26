package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// AccountAuthenticatorImpl an implementation of an authenticator to provide generic verification across multiple types.
//
// Types:
//   - [Ed25519Authenticator]
//   - [MultiEd25519Authenticator]
//   - [SingleKeyAuthenticator]
//   - [MultiKeyAuthenticator]
type AccountAuthenticatorImpl interface {
	bcs.Struct

	// PublicKey is the public key that can be used to verify the signature.  It must be a valid on-chain representation
	// and cannot be something like [Secp256k1PublicKey] on its own.
	PublicKey() PublicKey

	// Signature is a typed signature that can be verified by the public key. It must be a valid on-chain representation
	// and cannot be something like [Secp256k1Signature] on its own.
	Signature() Signature

	// Verify Return true if the [AccountAuthenticator] can be cryptographically verified
	Verify(data []byte) bool
}

//region AccountAuthenticator

// AccountAuthenticatorType single byte representing the spot in the enum from the Rust implementation
type AccountAuthenticatorType uint8

const (
	AccountAuthenticatorEd25519      AccountAuthenticatorType = 0 // AccountAuthenticatorEd25519 is the authenticator type for ed25519 accounts
	AccountAuthenticatorMultiEd25519 AccountAuthenticatorType = 1 // AccountAuthenticatorMultiEd25519 is the authenticator type for multi-ed25519 accounts
	AccountAuthenticatorSingleSender AccountAuthenticatorType = 2 // AccountAuthenticatorSingleSender is the authenticator type for single-key accounts
	AccountAuthenticatorMultiKey     AccountAuthenticatorType = 3 // AccountAuthenticatorMultiKey is the authenticator type for multi-key accounts
)

// AccountAuthenticator a generic authenticator type for a transaction
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type AccountAuthenticator struct {
	Variant AccountAuthenticatorType // Variant is the type of authenticator
	Auth    AccountAuthenticatorImpl // Auth is the actual authenticator
}

//region AccountAuthenticator AccountAuthenticatorImpl implementation

// PubKey returns the public key of the authenticator
func (ea *AccountAuthenticator) PubKey() PublicKey {
	return ea.Auth.PublicKey()
}

// Signature returns the signature of the authenticator
func (ea *AccountAuthenticator) Signature() Signature {
	return ea.Auth.Signature()
}

// Verify returns true if the authenticator can be cryptographically verified
func (ea *AccountAuthenticator) Verify(data []byte) bool {
	return ea.Auth.Verify(data)
}

//endregion

//region AccountAuthenticator bcs.Struct implementation

// MarshalBCS serializes the [AccountAuthenticator] to the BCS format
//
// Implements:
//   - [bcs.Marshaler]
func (ea *AccountAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(ea.Variant))
	ea.Auth.MarshalBCS(ser)
}

// UnmarshalBCS deserializes the [AccountAuthenticator] from the BCS format
//
// Implements:
//   - [bcs.Unmarshaler]
func (ea *AccountAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	kindNum := des.Uleb128()
	if des.Error() != nil {
		return
	}
	ea.Variant = AccountAuthenticatorType(kindNum)
	switch ea.Variant {
	case AccountAuthenticatorEd25519:
		ea.Auth = &Ed25519Authenticator{}
	case AccountAuthenticatorMultiEd25519:
		ea.Auth = &MultiEd25519Authenticator{}
	case AccountAuthenticatorSingleSender:
		ea.Auth = &SingleKeyAuthenticator{}
	case AccountAuthenticatorMultiKey:
		ea.Auth = &MultiKeyAuthenticator{}
	default:
		des.SetError(fmt.Errorf("unknown AccountAuthenticator kind: %d", kindNum))
		return
	}
	ea.Auth.UnmarshalBCS(des)
}

//endregion
//endregion
