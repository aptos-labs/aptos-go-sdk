package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

type AnyPublicKeyVariant uint32

const (
	AnyPublicKeyVariantEd25519   AnyPublicKeyVariant = 0
	AnyPublicKeyVariantSecp256k1 AnyPublicKeyVariant = 1
)

type AnySignatureVariant uint32

const (
	AnySignatureVariantEd25519   AnySignatureVariant = 0
	AnySignatureVariantSecp256k1 AnySignatureVariant = 1
)

type AnyPublicKey struct {
	Variant AnyPublicKeyVariant
	PubKey  PublicKey
}

func (key *AnyPublicKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

func (key *AnyPublicKey) Scheme() uint8 {
	return SingleKeyScheme
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

func (key *AnyPublicKey) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(key, bytes)
}

func (key *AnyPublicKey) Verify(msg []byte, sig Signature) bool {
	return key.PubKey.Verify(msg, sig)
}

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
		panic("TODO Implement")
	default:
		bcs.SetError(fmt.Errorf("unknown public key variant: %d", key.Variant))
		return
	}
	bcs.Struct(key.PubKey)
}

type AnySignature struct {
	Variant   AnySignatureVariant
	Signature Signature
}

func (e *AnySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

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
		panic("TODO Implement")
	default:
		bcs.SetError(fmt.Errorf("unknown signature variant: %d", e.Variant))
		return
	}
	bcs.Struct(e.Signature)
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

func (e *AnySignature) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(e, bytes)
}

type SingleKeyAuthenticator struct {
	PubKey *AnyPublicKey
	Sig    *AnySignature
}

func (ea *SingleKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *SingleKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

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

// Verify Return true if the data was well signed
func (ea *SingleKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}
