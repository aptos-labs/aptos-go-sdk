package crypto

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

type MultiKey struct {
	PubKeys            []AnyPublicKey
	SignaturesRequired uint8
}

func (key *MultiKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

func (key *MultiKey) Scheme() uint8 {
	return MultiKeyScheme
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

func (key *MultiKey) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(key, bytes)
}

func (key *MultiKey) Verify(msg []byte, sig Signature) bool {
	for _, pub := range key.PubKeys {
		if !pub.Verify(msg, sig) {
			return false
		}
	}
	return true
}

func (key *MultiKey) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(key.PubKeys, ser)
	ser.U8(key.SignaturesRequired)
}

func (key *MultiKey) UnmarshalBCS(des *bcs.Deserializer) {
	key.PubKeys = bcs.DeserializeSequence[AnyPublicKey](des)
	key.SignaturesRequired = des.U8()
}

type MultiKeySignature struct {
	Signatures []AnySignature
	Bitmap     []byte
}

func (e *MultiKeySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

func (e *MultiKeySignature) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(e.Signatures, ser)
	ser.WriteBytes(e.Bitmap)
}

func (e *MultiKeySignature) UnmarshalBCS(des *bcs.Deserializer) {
	e.Signatures = bcs.DeserializeSequence[AnySignature](des)
	e.Bitmap = des.ReadBytes()
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

func (e *MultiKeySignature) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(e, bytes)
}

type MultiKeyAuthenticator struct {
	PubKey *MultiKey
	Sig    *MultiKeySignature
}

func (ea *MultiKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *MultiKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

func (ea *MultiKeyAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Struct(ea.PublicKey())
	bcs.Struct(ea.Signature())
}

func (ea *MultiKeyAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.PubKey = &MultiKey{}
	bcs.Struct(ea.PubKey)
	err := bcs.Error()
	if err != nil {
		return
	}
	ea.Sig = &MultiKeySignature{}
	bcs.Struct(ea.Sig)
}

// Verify Return true if the data was well signed
func (ea *MultiKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}
