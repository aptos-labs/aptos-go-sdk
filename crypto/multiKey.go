package crypto

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/util"
)

//region MultiKey

// MultiKey is an off-chain multi-sig, where multiple different keys can be used together to create an account
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type MultiKey struct {
	PubKeys            []*AnyPublicKey // The public keys of the sub-keys
	SignaturesRequired uint8           // The number of signatures required to pass verification
}

//region MultiKey VerifyingKey implementation

// Verify verifies the signature against the message
// This function will return true if the number of verified signatures is greater than or equal to the number of required signatures
//
// Implements:
//   - [VerifyingKey]
func (key *MultiKey) Verify(msg []byte, signature Signature) bool {
	switch sig := signature.(type) {
	case *MultiKeySignature:
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

// AuthKey converts the public key to an authentication key
//
// Implements:
//   - [PublicKey]
func (key *MultiKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns the scheme for the public key
//
// Implements:
//   - [PublicKey]
func (key *MultiKey) Scheme() uint8 {
	return MultiKeyScheme
}

//endregion

//region MultiKey CryptoMaterial implementation

// Bytes converts the public key to bytes
//
// Implements:
//   - [CryptoMaterial]
func (key *MultiKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

// FromBytes converts the public key from bytes
//
// Implements:
//   - [CryptoMaterial]
func (key *MultiKey) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(key, bytes)
}

// ToHex converts the public key to a hex string
//
// Implements:
//   - [CryptoMaterial]
func (key *MultiKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex converts the public key from a hex string
//
// Implements:
//   - [CryptoMaterial]
func (key *MultiKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion

//region MultiKey bcs.Struct implementation

// MarshalBCS converts the public key to BCS
//
// Implements:
//   - [bcs.Marshaler]
func (key *MultiKey) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(key.PubKeys, ser)
	ser.U8(key.SignaturesRequired)
}

// UnmarshalBCS converts the public key from BCS
//
// Implements:
//   - [bcs.Unmarshaler]
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
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type MultiKeySignature struct {
	Signatures []*AnySignature // The signatures of the sub-keys
	Bitmap     MultiKeyBitmap  // The bitmap of the signatures
}

//region MultiKeySignature CryptoMaterial implementation

// Bytes converts the signature to bytes
//
// Implements:
//   - [CryptoMaterial]
func (e *MultiKeySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

// FromBytes converts the signature from bytes
//
// Implements:
//   - [CryptoMaterial]
func (e *MultiKeySignature) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(e, bytes)
}

// ToHex converts the signature to a hex string
//
// Implements:
//   - [CryptoMaterial]
func (e *MultiKeySignature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex converts the signature from a hex string
//
// Implements:
//   - [CryptoMaterial]
func (e *MultiKeySignature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

//endregion

//region MultiKeySignature bcs.Struct implementation

// MarshalBCS converts the signature to BCS
//
// Implements:
//   - [bcs.Marshaler]
func (e *MultiKeySignature) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(e.Signatures, ser)
	e.Bitmap.MarshalBCS(ser)
}

// UnmarshalBCS converts the signature from BCS
//
// Implements:
//   - [bcs.Unmarshaler]
func (e *MultiKeySignature) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	e.Signatures = make([]*AnySignature, length)

	for i := uint32(0); i < length; i++ {
		e.Signatures[i] = &AnySignature{}
		des.Struct(e.Signatures[i])
	}

	e.Bitmap.UnmarshalBCS(des)
}

//endregion

//endregion

//region MultiKeyAuthenticator

// MultiKeyAuthenticator is an on-chain authenticator for a MultiKeySignature
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type MultiKeyAuthenticator struct {
	PubKey *MultiKey          // The public key of the authenticator
	Sig    *MultiKeySignature // The signature of the authenticator
}

//region MultiKeyAuthenticator AccountAuthenticatorImpl implementation

// PublicKey returns the public key of the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *MultiKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

// Signature returns the signature of the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *MultiKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

// Verify verifies the signature against the message
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *MultiKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

//endregion

//region MultiKeyAuthenticator bcs.Struct implementation

// MarshalBCS serializes the authenticator to bytes
//
// Implements:
//   - [bcs.Marshaler]
func (ea *MultiKeyAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PublicKey())
	ser.Struct(ea.Signature())
}

// UnmarshalBCS deserializes the authenticator from bytes
//
// Implements:
//   - [bcs.Unmarshaler]
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

//region MultiKeyBitmap

// MultiKeyBitmapSize represents the 4 bytes needed to make a 32-bit bitmap
const MultiKeyBitmapSize = uint32(4)

// MultiKeyBitmap represents a bitmap of signatures in a MultiKey public key that signed the transaction
// There are a maximum of 32 possible values in MultiKeyBitmapSize, starting from the leftmost bit representing
// index 0 of the public key
type MultiKeyBitmap [MultiKeyBitmapSize]byte

// ContainsKey tells us if the current index is in the map
func (bm *MultiKeyBitmap) ContainsKey(index uint8) bool {
	numByte, numBit := KeyIndices(index)
	return (bm[numByte] & (128 >> numBit)) == 1
}

// AddKey adds the value to the map, returning an error if it is already added
func (bm *MultiKeyBitmap) AddKey(index uint8) error {
	if bm.ContainsKey(index) {
		return fmt.Errorf("index %d already in bitmap", index)
	}
	numByte, numBit := KeyIndices(index)
	bm[numByte] = bm[numByte] | (128 >> numBit)
	return nil
}

// KeyIndices determines the byte and bit set in the bitmap
func KeyIndices(index uint8) (numByte uint8, numBit uint8) {
	// Bytes and bits are counted from left
	return index / 8, index % 8
}

//region MultiKeyBitmap bcs.Struct

// MarshalBCS serializes the bitmap to bytes
//
// Implements:
//   - [bcs.Marshaler]
func (bm *MultiKeyBitmap) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(bm[:])
}

// UnmarshalBCS deserializes the bitmap from bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (bm *MultiKeyBitmap) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	if length != MultiKeyBitmapSize {
		des.SetError(fmt.Errorf("MultiKeyBitmap must be %d bytes, got %d", MultiKeyBitmapSize, length))
		return
	}
	des.ReadFixedBytesInto(bm[:])
}

//endregion
//endregion
