package crypto

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

//region MultiKey

// MultiKey is an off-chain multisig, where multiple different keys can be used together to create an account
// Implements VerifyingKey, PublicKey, CryptoMaterial, bcs.Struct
type MultiKey struct {
	PubKeys            []*AnyPublicKey
	SignaturesRequired uint8
}

//region MultiKey VerifyingKey implementation

func (key *MultiKey) Verify(msg []byte, signature Signature) bool {
	switch signature.(type) {
	case *MultiKeySignature:
		sig := signature.(*MultiKeySignature)
		verified := uint8(0)
		for i, pub := range key.PubKeys {
			if pub.Verify(msg, sig.Signatures[i]) {
				verified++
			}
		}
		return key.SignaturesRequired <= verified
	default:
		return false
	}
}

//endregion

//region MultiKey PublicKey implementation

func (key *MultiKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

func (key *MultiKey) Scheme() uint8 {
	return MultiKeyScheme
}

//endregion

//region MultiKey CryptoMaterial implementation

func (key *MultiKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

func (key *MultiKey) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(key, bytes)
}

func (key *MultiKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

func (key *MultiKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion

//region MultiKey bcs.Struct implementation

func (key *MultiKey) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(key.PubKeys, ser)
	ser.U8(key.SignaturesRequired)
}

func (key *MultiKey) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	key.PubKeys = make([]*AnyPublicKey, length)

	for i := uint32(0); i < length; i++ {
		key.PubKeys[i] = &AnyPublicKey{}
		des.Struct(key.PubKeys[i])
	}
	key.SignaturesRequired = des.U8()
}

//endregion
//endregion

//region MultiKeySignature

// MultiKeySignature is an off-chain multi-sig signature that can be verified by a MultiKey
// Implements Signature, CryptoMaterial, bcs.Struct
type MultiKeySignature struct {
	Signatures []*AnySignature
	Bitmap     []byte
}

//region MultiKeySignature CryptoMaterial implementation

func (e *MultiKeySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

func (e *MultiKeySignature) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(e, bytes)
}

func (e *MultiKeySignature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

func (e *MultiKeySignature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

//endregion

//region MultiKeySignature bcs.Struct implementation

func (e *MultiKeySignature) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(e.Signatures, ser)
	ser.WriteBytes(e.Bitmap)
}

func (e *MultiKeySignature) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	e.Signatures = make([]*AnySignature, length)

	for i := uint32(0); i < length; i++ {
		e.Signatures[i] = &AnySignature{}
		des.Struct(e.Signatures[i])
	}
	e.Bitmap = des.ReadBytes()
}

//endregion

//endregion

//region MultiKeyAuthenticator

// MultiKeyAuthenticator is an on-chain authenticator for a MultiKeySignature
// Implements AccountAuthenticatorImpl, bcs.Struct
type MultiKeyAuthenticator struct {
	PubKey *MultiKey
	Sig    *MultiKeySignature
}

//region MultiKeyAuthenticator AccountAuthenticatorImpl implementation

func (ea *MultiKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *MultiKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

func (ea *MultiKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

//endregion

//region MultiKeyAuthenticator bcs.Struct implementation

func (ea *MultiKeyAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PublicKey())
	ser.Struct(ea.Signature())
}

func (ea *MultiKeyAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &MultiKey{}
	des.Struct(ea.PubKey)
	err := des.Error()
	if err != nil {
		return
	}
	ea.Sig = &MultiKeySignature{}
	des.Struct(ea.Sig)
}

//endregion
//endregion
