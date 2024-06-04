package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// AuthenticatorImpl an implementation of an authenticator to provide generic verification across multiple types
type AuthenticatorImpl interface {
	bcs.Struct

	// PublicKey is the public key that can be used to verify the signature.  It must be a valid on-chain representation
	// and cannot be something like Secp256k1PublicKey on its own.
	PublicKey() PublicKey

	// Signature is a typed signature that can be verified by the public key. It must be a valid on-chain representation
	//	// and cannot be something like Secp256k1Signature on its own.
	Signature() Signature

	// Verify Return true if the Authenticator can be cryptographically verified
	Verify(data []byte) bool
}

//region Authenticator

// AuthenticatorType single byte representing the spot in the enum from the Rust implementation
type AuthenticatorType uint8

const (
	AuthenticatorEd25519      AuthenticatorType = 0
	AuthenticatorMultiEd25519 AuthenticatorType = 1
	AuthenticatorSingleSender AuthenticatorType = 2
	AuthenticatorMultiKey     AuthenticatorType = 3
)

// Authenticator a generic authenticator type for a transaction
// Implements AuthenticatorImpl, bcs.Struct
type Authenticator struct {
	Variant AuthenticatorType
	Auth    AuthenticatorImpl
}

//region Authenticator AuthenticatorImpl implementation

func (ea *Authenticator) PubKey() PublicKey {
	return ea.Auth.PublicKey()
}

func (ea *Authenticator) Signature() Signature {
	return ea.Auth.Signature()
}

func (ea *Authenticator) Verify(data []byte) bool {
	return ea.Auth.Verify(data)
}

//endregion

//region Authenticator bcs.Struct implementation

func (ea *Authenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(ea.Variant))
	ea.Auth.MarshalBCS(bcs)
}

func (ea *Authenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	kindNum := bcs.Uleb128()
	if bcs.Error() != nil {
		return
	}
	ea.Variant = AuthenticatorType(kindNum)
	switch ea.Variant {
	case AuthenticatorEd25519:
		ea.Auth = &Ed25519Authenticator{}
	case AuthenticatorMultiEd25519:
		ea.Auth = &MultiEd25519Authenticator{}
	case AuthenticatorSingleSender:
		ea.Auth = &SingleKeyAuthenticator{}
	case AuthenticatorMultiKey:
		ea.Auth = &MultiKeyAuthenticator{}
	default:
		bcs.SetError(fmt.Errorf("unknown Authenticator kind: %d", kindNum))
		return
	}
	ea.Auth.UnmarshalBCS(bcs)
}

//endregion
//endregion
