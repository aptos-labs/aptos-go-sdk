package crypto

import (
	"errors"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// AccountAuthenticatorImpl is implemented by all authenticator types.
type AccountAuthenticatorImpl interface {
	bcs.Struct

	// PublicKey returns the public key for verification.
	PublicKey() PublicKey

	// Signature returns the signature.
	Signature() Signature

	// Verify returns true if the signature is valid for the message.
	Verify(data []byte) bool
}

// AccountAuthenticatorType identifies the type of authenticator.
type AccountAuthenticatorType uint8

const (
	AccountAuthenticatorEd25519      AccountAuthenticatorType = 0
	AccountAuthenticatorMultiEd25519 AccountAuthenticatorType = 1
	AccountAuthenticatorSingleSender AccountAuthenticatorType = 2
	AccountAuthenticatorMultiKey     AccountAuthenticatorType = 3
	AccountAuthenticatorNone         AccountAuthenticatorType = 4 // Simulation only
)

// AccountAuthenticator is a generic container for different authenticator types.
//
// Implements:
//   - [bcs.Struct]
type AccountAuthenticator struct {
	Variant AccountAuthenticatorType
	Auth    AccountAuthenticatorImpl
}

// PubKey returns the public key from the inner authenticator.
func (ea *AccountAuthenticator) PubKey() PublicKey {
	return ea.Auth.PublicKey()
}

// Signature returns the signature from the inner authenticator.
func (ea *AccountAuthenticator) Signature() Signature {
	return ea.Auth.Signature()
}

// Verify returns true if the authenticator can be verified.
func (ea *AccountAuthenticator) Verify(data []byte) bool {
	return ea.Auth.Verify(data)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (ea *AccountAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(ea.Variant))
	ea.Auth.MarshalBCS(ser)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (ea *AccountAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	kindNum := des.Uleb128()
	if des.Error() != nil {
		return
	}
	index, err := util.Uint32ToU8(kindNum)
	if err != nil {
		des.SetError(err)
		return
	}
	ea.Variant = AccountAuthenticatorType(index)

	switch ea.Variant {
	case AccountAuthenticatorEd25519:
		ea.Auth = &Ed25519Authenticator{}
	case AccountAuthenticatorMultiEd25519:
		ea.Auth = &MultiEd25519Authenticator{}
	case AccountAuthenticatorSingleSender:
		ea.Auth = &SingleKeyAuthenticator{}
	case AccountAuthenticatorMultiKey:
		ea.Auth = &MultiKeyAuthenticator{}
	case AccountAuthenticatorNone:
		ea.Auth = &NoAuthenticator{}
	default:
		des.SetError(fmt.Errorf("unknown authenticator type: %d", kindNum))
		return
	}
	ea.Auth.UnmarshalBCS(des)
}

// FromKeyAndSignature creates an AccountAuthenticator from a public key and signature.
func (ea *AccountAuthenticator) FromKeyAndSignature(key PublicKey, sig Signature) error {
	switch k := key.(type) {
	case *Ed25519PublicKey:
		s, ok := sig.(*Ed25519Signature)
		if !ok {
			return errors.New("expected Ed25519Signature for Ed25519PublicKey")
		}
		ea.Variant = AccountAuthenticatorEd25519
		ea.Auth = &Ed25519Authenticator{PubKey: k, Sig: s}

	case *MultiEd25519PublicKey:
		s, ok := sig.(*MultiEd25519Signature)
		if !ok {
			return errors.New("expected MultiEd25519Signature for MultiEd25519PublicKey")
		}
		ea.Variant = AccountAuthenticatorMultiEd25519
		ea.Auth = &MultiEd25519Authenticator{PubKey: k, Sig: s}

	case *AnyPublicKey:
		s, ok := sig.(*AnySignature)
		if !ok {
			return errors.New("expected AnySignature for AnyPublicKey")
		}
		ea.Variant = AccountAuthenticatorSingleSender
		ea.Auth = &SingleKeyAuthenticator{PubKey: k, Sig: s}

	case *MultiKey:
		s, ok := sig.(*MultiKeySignature)
		if !ok {
			return errors.New("expected MultiKeySignature for MultiKey")
		}
		ea.Variant = AccountAuthenticatorMultiKey
		ea.Auth = &MultiKeyAuthenticator{PubKey: k, Sig: s}

	default:
		return fmt.Errorf("unknown key type: %T", key)
	}
	return nil
}

// NoAuthenticator is used for simulation only. It has no real signature.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type NoAuthenticator struct{}

// PublicKey returns nil (no public key).
func (ea *NoAuthenticator) PublicKey() PublicKey { return nil }

// Signature returns nil (no signature).
func (ea *NoAuthenticator) Signature() Signature { return nil }

// Verify always returns false.
func (ea *NoAuthenticator) Verify([]byte) bool { return false }

// MarshalBCS is a no-op.
func (ea *NoAuthenticator) MarshalBCS(*bcs.Serializer) {}

// UnmarshalBCS is a no-op.
func (ea *NoAuthenticator) UnmarshalBCS(*bcs.Deserializer) {}

// NoAccountAuthenticator creates an AccountAuthenticator for simulation.
func NoAccountAuthenticator() *AccountAuthenticator {
	return &AccountAuthenticator{
		Variant: AccountAuthenticatorNone,
		Auth:    &NoAuthenticator{},
	}
}
