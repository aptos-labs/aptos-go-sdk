package crypto

import (
	"fmt"
	"sort"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// MultiKey constants.
const (
	MaxMultiKeySignatures  = uint8(32)
	MaxMultiKeyBitmapBytes = MaxMultiKeySignatures / 8
)

// MultiKey is an off-chain multi-sig public key with mixed key types.
//
// MultiKey allows combining different key types (Ed25519, Secp256k1) for
// threshold signatures. This is useful for organizations that want to
// require M-of-N signatures from different key types.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type MultiKey struct {
	PubKeys            []*AnyPublicKey
	SignaturesRequired uint8
}

// Verify verifies a multi-key signature against a message.
//
// Implements [VerifyingKey].
func (key *MultiKey) Verify(msg []byte, signature Signature) bool {
	sig, ok := signature.(*MultiKeySignature)
	if !ok {
		return false
	}

	numSigs, err := util.IntToU8(len(sig.Signatures))
	if err != nil || key.SignaturesRequired > numSigs {
		return false
	}

	// Verify each signature in the bitmap
	for sigIndex, keyIndex := range sig.Bitmap.Indices() {
		auth := &AccountAuthenticator{}
		err := auth.FromKeyAndSignature(key.PubKeys[keyIndex], sig.Signatures[sigIndex])
		if err != nil || !auth.Verify(msg) {
			return false
		}
	}
	return true
}

// AuthKey returns the AuthenticationKey for this public key.
//
// Implements [PublicKey].
func (key *MultiKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns MultiKeyScheme.
//
// Implements [PublicKey].
func (key *MultiKey) Scheme() uint8 {
	return MultiKeyScheme
}

// Bytes returns the BCS-serialized public key.
//
// Implements [CryptoMaterial].
func (key *MultiKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

// FromBytes deserializes from BCS bytes.
//
// Implements [CryptoMaterial].
func (key *MultiKey) FromBytes(bytes []byte) error {
	return bcs.Deserialize(key, bytes)
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *MultiKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *MultiKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *MultiKey) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(ser, key.PubKeys)
	ser.U8(key.SignaturesRequired)
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *MultiKey) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	key.PubKeys = make([]*AnyPublicKey, length)
	for i := range length {
		key.PubKeys[i] = &AnyPublicKey{}
		des.Struct(key.PubKeys[i])
	}
	key.SignaturesRequired = des.U8()
}

// IndexedAnySignature pairs a signature with its key index in a MultiKey.
type IndexedAnySignature struct {
	Index     uint8
	Signature *AnySignature
}

// MarshalBCS serializes the indexed signature to BCS.
func (e *IndexedAnySignature) MarshalBCS(ser *bcs.Serializer) {
	ser.U8(e.Index)
	ser.Struct(e.Signature)
}

// UnmarshalBCS deserializes the indexed signature from BCS.
func (e *IndexedAnySignature) UnmarshalBCS(des *bcs.Deserializer) {
	e.Index = des.U8()
	e.Signature = &AnySignature{}
	des.Struct(e.Signature)
}

// MultiKeySignature is a collection of signatures for a MultiKey.
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type MultiKeySignature struct {
	Signatures []*AnySignature
	Bitmap     MultiKeyBitmap
}

// NewMultiKeySignature creates a MultiKeySignature from indexed signatures.
// Signatures are automatically sorted by index.
func NewMultiKeySignature(signatures []IndexedAnySignature) (*MultiKeySignature, error) {
	// Sort by index
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].Index < signatures[j].Index
	})

	sig := &MultiKeySignature{}
	for _, s := range signatures {
		sig.Signatures = append(sig.Signatures, s.Signature)
		if err := sig.Bitmap.AddKey(s.Index); err != nil {
			return nil, err
		}
	}
	return sig, nil
}

// Bytes returns the BCS-serialized signature.
//
// Implements [CryptoMaterial].
func (e *MultiKeySignature) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

// FromBytes deserializes from BCS bytes.
//
// Implements [CryptoMaterial].
func (e *MultiKeySignature) FromBytes(bytes []byte) error {
	return bcs.Deserialize(e, bytes)
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (e *MultiKeySignature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (e *MultiKeySignature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (e *MultiKeySignature) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(ser, e.Signatures)
	e.Bitmap.MarshalBCS(ser)
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (e *MultiKeySignature) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	e.Signatures = make([]*AnySignature, length)
	for i := range length {
		e.Signatures[i] = &AnySignature{}
		des.Struct(e.Signatures[i])
	}
	e.Bitmap.UnmarshalBCS(des)
}

// MultiKeyAuthenticator is an authenticator for MultiKey signatures.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type MultiKeyAuthenticator struct {
	PubKey *MultiKey
	Sig    *MultiKeySignature
}

// PublicKey returns the multi-key public key.
//
// Implements [AccountAuthenticatorImpl].
func (ea *MultiKeyAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

// Signature returns the multi-key signature.
//
// Implements [AccountAuthenticatorImpl].
func (ea *MultiKeyAuthenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the signature is valid.
//
// Implements [AccountAuthenticatorImpl].
func (ea *MultiKeyAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (ea *MultiKeyAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PubKey)
	ser.Struct(ea.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (ea *MultiKeyAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &MultiKey{}
	des.Struct(ea.PubKey)
	if des.Error() != nil {
		return
	}
	ea.Sig = &MultiKeySignature{}
	des.Struct(ea.Sig)
}

// MultiKeyBitmap tracks which keys in a MultiKey have signed.
//
// Implements [bcs.Struct].
type MultiKeyBitmap struct {
	inner []byte
}

// ContainsKey returns true if the key at index has signed.
func (bm *MultiKeyBitmap) ContainsKey(index uint8) bool {
	if index >= MaxMultiKeySignatures {
		return false
	}
	numByte, numBit := keyIndices(index)
	if int(numByte) >= len(bm.inner) {
		return false
	}
	return (bm.inner[numByte] & (128 >> numBit)) != 0
}

// AddKey marks the key at index as having signed.
func (bm *MultiKeyBitmap) AddKey(index uint8) error {
	if index >= MaxMultiKeySignatures {
		return fmt.Errorf("index %d exceeds maximum %d", index, MaxMultiKeySignatures)
	}
	if bm.ContainsKey(index) {
		return fmt.Errorf("index %d already in bitmap", index)
	}

	numByte, numBit := keyIndices(index)

	// Expand bitmap if needed
	if int(numByte) >= len(bm.inner) {
		newInner := make([]byte, numByte+1)
		copy(newInner, bm.inner)
		bm.inner = newInner
	}

	bm.inner[numByte] |= 128 >> numBit
	return nil
}

// Indices returns the key indices that have signed.
func (bm *MultiKeyBitmap) Indices() []uint8 {
	var indices []uint8
	for i := uint8(0); i < MaxMultiKeySignatures; i++ {
		if bm.ContainsKey(i) {
			indices = append(indices, i)
		}
	}
	return indices
}

// MarshalBCS serializes the bitmap to BCS.
func (bm *MultiKeyBitmap) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(bm.inner)
}

// UnmarshalBCS deserializes the bitmap from BCS.
func (bm *MultiKeyBitmap) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	if length > uint32(MaxMultiKeyBitmapBytes) {
		des.SetError(fmt.Errorf("bitmap must be at most %d bytes, got %d", MaxMultiKeyBitmapBytes, length))
		return
	}
	bm.inner = make([]byte, length)
	des.ReadFixedBytesInto(bm.inner)
}

func keyIndices(index uint8) (byte uint8, bit uint8) {
	return index / 8, index % 8
}
