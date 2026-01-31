package crypto

import (
	"errors"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// SingleSigner wraps a MessageSigner to implement the Signer interface.
//
// SingleSigner uses the SingleKeyScheme for address derivation, which supports
// multiple key types including Ed25519 and Secp256k1.
//
// Example:
//
//	secpKey, _ := crypto.GenerateSecp256k1Key()
//	signer := crypto.NewSingleSigner(secpKey)
//	auth, _ := signer.Sign(txnBytes)
//
// Implements [Signer].
type SingleSigner struct {
	inner MessageSigner
}

// NewSingleSigner creates a new SingleSigner from a MessageSigner.
func NewSingleSigner(signer MessageSigner) *SingleSigner {
	return &SingleSigner{inner: signer}
}

// Sign signs a message and returns an AccountAuthenticator.
//
// Implements [Signer].
func (s *SingleSigner) Sign(msg []byte) (*AccountAuthenticator, error) {
	sig, err := s.SignMessage(msg)
	if err != nil {
		return nil, err
	}

	pubKey, ok := s.PubKey().(*AnyPublicKey)
	if !ok {
		return nil, errors.New("expected AnyPublicKey")
	}
	anySig, ok := sig.(*AnySignature)
	if !ok {
		return nil, errors.New("expected AnySignature")
	}

	auth := &SingleKeyAuthenticator{
		PubKey: pubKey,
		Sig:    anySig,
	}
	return &AccountAuthenticator{
		Variant: AccountAuthenticatorSingleSender,
		Auth:    auth,
	}, nil
}

// SignMessage signs a message and returns an AnySignature.
//
// Implements [Signer].
func (s *SingleSigner) SignMessage(msg []byte) (Signature, error) {
	sig, err := s.inner.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	return &AnySignature{
		Variant:   s.signatureVariant(),
		Signature: sig,
	}, nil
}

// SimulationAuthenticator creates an authenticator with an empty signature for simulation.
//
// Implements [Signer].
func (s *SingleSigner) SimulationAuthenticator() *AccountAuthenticator {
	pubKey, ok := s.PubKey().(*AnyPublicKey)
	if !ok {
		return nil
	}
	return &AccountAuthenticator{
		Variant: AccountAuthenticatorSingleSender,
		Auth: &SingleKeyAuthenticator{
			PubKey: pubKey,
			Sig:    s.EmptySignature(),
		},
	}
}

// AuthKey returns the AuthenticationKey derived from this signer.
//
// Implements [Signer].
func (s *SingleSigner) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(s.PubKey())
	return out
}

// PubKey returns the AnyPublicKey for this signer.
//
// Implements [Signer].
func (s *SingleSigner) PubKey() PublicKey {
	innerPubKey := s.inner.VerifyingKey()
	return &AnyPublicKey{
		Variant: s.publicKeyVariant(),
		PubKey:  innerPubKey,
	}
}

// EmptySignature returns an empty AnySignature for simulation.
func (s *SingleSigner) EmptySignature() *AnySignature {
	return &AnySignature{
		Variant:   s.signatureVariant(),
		Signature: s.inner.EmptySignature(),
	}
}

func (s *SingleSigner) signatureVariant() AnySignatureVariant {
	switch s.inner.(type) {
	case *Ed25519PrivateKey:
		return AnySignatureVariantEd25519
	case *Secp256k1PrivateKey:
		return AnySignatureVariantSecp256k1
	default:
		return AnySignatureVariantEd25519
	}
}

func (s *SingleSigner) publicKeyVariant() AnyPublicKeyVariant {
	switch s.inner.(type) {
	case *Ed25519PrivateKey:
		return AnyPublicKeyVariantEd25519
	case *Secp256k1PrivateKey:
		return AnyPublicKeyVariantSecp256k1
	default:
		return AnyPublicKeyVariantEd25519
	}
}

// AnyPublicKeyVariant identifies the type of public key in AnyPublicKey.
type AnyPublicKeyVariant uint32

const (
	AnyPublicKeyVariantEd25519          AnyPublicKeyVariant = 0
	AnyPublicKeyVariantSecp256k1        AnyPublicKeyVariant = 1
	AnyPublicKeyVariantSecp256r1        AnyPublicKeyVariant = 2
	AnyPublicKeyVariantKeyless          AnyPublicKeyVariant = 3
	AnyPublicKeyVariantFederatedKeyless AnyPublicKeyVariant = 4
	AnyPublicKeyVariantSlhDsaSha2_128s  AnyPublicKeyVariant = 5
)

// AnyPublicKey wraps different public key types for use with SingleSigner.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type AnyPublicKey struct {
	Variant AnyPublicKeyVariant
	PubKey  VerifyingKey
}

// ToAnyPublicKey converts a VerifyingKey to an AnyPublicKey.
func ToAnyPublicKey(key VerifyingKey) (*AnyPublicKey, error) {
	switch k := key.(type) {
	case *Ed25519PublicKey:
		return &AnyPublicKey{Variant: AnyPublicKeyVariantEd25519, PubKey: k}, nil
	case *Secp256k1PublicKey:
		return &AnyPublicKey{Variant: AnyPublicKeyVariantSecp256k1, PubKey: k}, nil
	case *Secp256r1PublicKey:
		return &AnyPublicKey{Variant: AnyPublicKeyVariantSecp256r1, PubKey: k}, nil
	case *SlhDsaPublicKey:
		return &AnyPublicKey{Variant: AnyPublicKeyVariantSlhDsaSha2_128s, PubKey: k}, nil
	case *AnyPublicKey:
		return k, nil
	default:
		return nil, fmt.Errorf("unknown public key type: %T", key)
	}
}

// Verify verifies a signature against a message.
//
// Implements [VerifyingKey].
func (key *AnyPublicKey) Verify(msg []byte, sig Signature) bool {
	anySig, ok := sig.(*AnySignature)
	if !ok {
		return false
	}
	return key.PubKey.Verify(msg, anySig.Signature)
}

// AuthKey returns the AuthenticationKey for this public key.
//
// Implements [PublicKey].
func (key *AnyPublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns SingleKeyScheme.
//
// Implements [PublicKey].
func (key *AnyPublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

// Bytes returns the BCS-serialized public key.
//
// Implements [CryptoMaterial].
func (key *AnyPublicKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

// FromBytes deserializes from BCS bytes.
//
// Implements [CryptoMaterial].
func (key *AnyPublicKey) FromBytes(bytes []byte) error {
	return bcs.Deserialize(key, bytes)
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *AnyPublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *AnyPublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *AnyPublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(key.Variant))
	ser.Struct(key.PubKey)
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *AnyPublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	key.Variant = AnyPublicKeyVariant(des.Uleb128())
	switch key.Variant {
	case AnyPublicKeyVariantEd25519:
		key.PubKey = &Ed25519PublicKey{}
	case AnyPublicKeyVariantSecp256k1:
		key.PubKey = &Secp256k1PublicKey{}
	case AnyPublicKeyVariantSecp256r1:
		key.PubKey = &Secp256r1PublicKey{}
	case AnyPublicKeyVariantKeyless:
		des.SetError(fmt.Errorf("Keyless public key deserialization not yet implemented"))
		return
	case AnyPublicKeyVariantFederatedKeyless:
		des.SetError(fmt.Errorf("FederatedKeyless public key deserialization not yet implemented"))
		return
	case AnyPublicKeyVariantSlhDsaSha2_128s:
		key.PubKey = &SlhDsaPublicKey{}
	default:
		des.SetError(fmt.Errorf("unknown public key variant: %d", key.Variant))
		return
	}
	des.Struct(key.PubKey)
}

// AnySignatureVariant identifies the type of signature in AnySignature.
type AnySignatureVariant uint32

const (
	AnySignatureVariantEd25519         AnySignatureVariant = 0
	AnySignatureVariantSecp256k1       AnySignatureVariant = 1
	AnySignatureVariantWebAuthn        AnySignatureVariant = 2
	AnySignatureVariantKeyless         AnySignatureVariant = 3
	AnySignatureVariantSlhDsaSha2_128s AnySignatureVariant = 4
)

// AnySignature wraps different signature types for use with SingleSigner.
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type AnySignature struct {
	Variant   AnySignatureVariant
	Signature Signature
}

// Bytes returns the BCS-serialized signature.
//
// Implements [CryptoMaterial].
func (e *AnySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

// FromBytes deserializes from BCS bytes.
//
// Implements [CryptoMaterial].
func (e *AnySignature) FromBytes(bytes []byte) error {
	return bcs.Deserialize(e, bytes)
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (e *AnySignature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (e *AnySignature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (e *AnySignature) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(e.Variant))
	ser.Struct(e.Signature)
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (e *AnySignature) UnmarshalBCS(des *bcs.Deserializer) {
	e.Variant = AnySignatureVariant(des.Uleb128())
	switch e.Variant {
	case AnySignatureVariantEd25519:
		e.Signature = &Ed25519Signature{}
	case AnySignatureVariantSecp256k1:
		e.Signature = &Secp256k1Signature{}
	case AnySignatureVariantWebAuthn:
		des.SetError(fmt.Errorf("WebAuthn signature deserialization not yet implemented"))
		return
	case AnySignatureVariantKeyless:
		des.SetError(fmt.Errorf("Keyless signature deserialization not yet implemented"))
		return
	case AnySignatureVariantSlhDsaSha2_128s:
		e.Signature = &SlhDsaSignature{}
	default:
		des.SetError(fmt.Errorf("unknown signature variant: %d", e.Variant))
		return
	}
	des.Struct(e.Signature)
}

// SingleKeyAuthenticator is the authenticator for SingleSigner.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type SingleKeyAuthenticator struct {
	PubKey *AnyPublicKey
	Sig    *AnySignature
}

// PublicKey returns the public key.
//
// Implements [AccountAuthenticatorImpl].
func (ea *SingleKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

// Signature returns the signature.
//
// Implements [AccountAuthenticatorImpl].
func (ea *SingleKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the signature is valid for the message.
//
// Implements [AccountAuthenticatorImpl].
func (ea *SingleKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (ea *SingleKeyAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PubKey)
	ser.Struct(ea.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (ea *SingleKeyAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &AnyPublicKey{}
	des.Struct(ea.PubKey)
	if des.Error() != nil {
		return
	}
	ea.Sig = &AnySignature{}
	des.Struct(ea.Sig)
}
