package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

//region SingleSigner

// SingleSigner is a wrapper around different types of MessageSigners to allow for many types of keys
// Implements Signer
type SingleSigner struct {
	Signer MessageSigner
}

func NewSingleSigner(input MessageSigner) *SingleSigner {
	return &SingleSigner{Signer: input}
}

// SignMessage similar, but doesn't implement MessageSigner so there's no circular usage
func (key *SingleSigner) SignMessage(msg []byte) (*AnySignature, error) {
	signature, err := key.Signer.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	sigType := AnySignatureVariantEd25519
	switch key.Signer.(type) {
	case *Ed25519PrivateKey:
		sigType = AnySignatureVariantEd25519
	case *Secp256k1PrivateKey:
		sigType = AnySignatureVariantSecp256k1
	}

	return &AnySignature{
		Variant:   sigType,
		Signature: signature,
	}, nil
}

// region SingleSigner Signer implementation

func (key *SingleSigner) Sign(msg []byte) (authenticator *Authenticator, err error) {
	signature, err := key.SignMessage(msg)
	if err != nil {
		return nil, err
	}

	auth := &SingleKeyAuthenticator{}
	auth.PubKey = key.PubKey().(*AnyPublicKey)
	auth.Sig = signature
	return &Authenticator{Variant: AuthenticatorSingleSender, Auth: auth}, nil
}

func (key *SingleSigner) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key.PubKey())
	return out
}

func (key *SingleSigner) PubKey() PublicKey {
	innerPubKey := key.Signer.VerifyingKey()
	keyType := AnyPublicKeyVariantEd25519
	switch key.Signer.(type) {
	case *Ed25519PrivateKey:
		keyType = AnyPublicKeyVariantEd25519
	case *Secp256k1PrivateKey:
		keyType = AnyPublicKeyVariantSecp256k1
	}
	return &AnyPublicKey{
		Variant: keyType,
		PubKey:  innerPubKey,
	}
}

//endregion
//endregion

//region AnyPublicKey

// AnyPublicKeyVariant is an enum ID for the public key used in AnyPublicKey
type AnyPublicKeyVariant uint32

const (
	AnyPublicKeyVariantEd25519   AnyPublicKeyVariant = 0
	AnyPublicKeyVariantSecp256k1 AnyPublicKeyVariant = 1
)

// AnyPublicKey is used by SingleSigner and MultiKey to allow for using different keys with the same structs
// Implements VerifyingKey, PublicKey, bcs.Struct, CryptoMaterial
type AnyPublicKey struct {
	Variant AnyPublicKeyVariant
	PubKey  VerifyingKey
}

func ToAnyPublicKey(key VerifyingKey) *AnyPublicKey {
	out := &AnyPublicKey{}
	switch key.(type) {
	case *Ed25519PublicKey:
		out.Variant = AnyPublicKeyVariantEd25519
	case *Secp256k1PublicKey:
		out.Variant = AnyPublicKeyVariantSecp256k1
	}
	out.PubKey = key
	return out
}

//region AnyPublicKey VerifyingKey implementation

func (key *AnyPublicKey) Verify(msg []byte, sig Signature) bool {
	switch sig.(type) {
	case *AnySignature:
		anySig := sig.(*AnySignature)
		return key.PubKey.Verify(msg, anySig.Signature)
	default:
		return false
	}
}

//endregion

//region AnyPublicKey PublicKey implementation

func (key *AnyPublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

func (key *AnyPublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

//endregion

//region AnyPublicKey CryptoMaterial implementation

func (key *AnyPublicKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

func (key *AnyPublicKey) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(key, bytes)
}

func (key *AnyPublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

func (key *AnyPublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion

//region AnyPublicKey bcs.Struct implementation

func (key *AnyPublicKey) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(key.Variant))
	bcs.Struct(key.PubKey)
}

func (key *AnyPublicKey) UnmarshalBCS(bcs *bcs.Deserializer) {
	key.Variant = AnyPublicKeyVariant(bcs.Uleb128())
	switch key.Variant {
	case AnyPublicKeyVariantEd25519:
		key.PubKey = &Ed25519PublicKey{}
	case AnyPublicKeyVariantSecp256k1:
		key.PubKey = &Secp256k1PublicKey{}
	default:
		bcs.SetError(fmt.Errorf("unknown public key variant: %d", key.Variant))
		return
	}
	bcs.Struct(key.PubKey)
}

//endregion
//endregion

//region AnySignature

// AnySignatureVariant is an enum ID for the signature used in AnySignature
type AnySignatureVariant uint32

const (
	AnySignatureVariantEd25519   AnySignatureVariant = 0
	AnySignatureVariantSecp256k1 AnySignatureVariant = 1
)

// AnySignature is a wrapper around signatures signed with SingleSigner and verified with AnyPublicKey
// Implements Signature, CryptoMaterial, bcs.Struct
type AnySignature struct {
	Variant   AnySignatureVariant
	Signature Signature
}

// region AnySignature CryptoMaterial implementation

func (e *AnySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

func (e *AnySignature) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(e, bytes)
}

func (e *AnySignature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

func (e *AnySignature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

//endregion

//region AnySignature bcs.Struct implementation

func (e *AnySignature) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(e.Variant))
	bcs.Struct(e.Signature)
}

func (e *AnySignature) UnmarshalBCS(bcs *bcs.Deserializer) {
	e.Variant = AnySignatureVariant(bcs.Uleb128())
	switch e.Variant {
	case AnySignatureVariantEd25519:
		e.Signature = &Ed25519Signature{}
	case AnySignatureVariantSecp256k1:
		e.Signature = &Secp256k1Signature{}
	default:
		bcs.SetError(fmt.Errorf("unknown signature variant: %d", e.Variant))
		return
	}
	bcs.Struct(e.Signature)
}

//endregion
//endregion

//region SingleKeyAuthenticator

// SingleKeyAuthenticator is an authenticator for a SingleSigner
// Implements AuthenticatorImpl
// TODO: Converge on names for all these types
type SingleKeyAuthenticator struct {
	PubKey *AnyPublicKey
	Sig    *AnySignature
}

//region SingleKeyAuthenticator AuthenticatorImpl implementation

func (ea *SingleKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *SingleKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

func (ea *SingleKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

//endregion

//region SingleKeyAuthenticator bcs.Struct implementation

func (ea *SingleKeyAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Struct(ea.PublicKey())
	bcs.Struct(ea.Signature())
}

func (ea *SingleKeyAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.PubKey = &AnyPublicKey{}
	bcs.Struct(ea.PubKey)
	err := bcs.Error()
	if err != nil {
		return
	}
	ea.Sig = &AnySignature{}
	bcs.Struct(ea.Sig)
}

//endregion
//endregion
