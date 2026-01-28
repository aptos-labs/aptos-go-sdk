package crypto

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// MultiEd25519BitmapLen is the fixed bitmap length for multi-ed25519 signatures.
const MultiEd25519BitmapLen = 4

// Compile-time assertion: ensure bitmap indices fit in uint8.
// Maximum index is (MultiEd25519BitmapLen * 8 - 1) = 31 for the current value of 4.
const _ = uint8(MultiEd25519BitmapLen*8 - 1)

// MultiEd25519PublicKey is an off-chain multi-sig public key using only Ed25519 keys.
//
// This is the legacy multi-sig format. For new implementations, consider using
// [MultiKey] which supports mixed key types.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type MultiEd25519PublicKey struct {
	PubKeys            []*Ed25519PublicKey
	SignaturesRequired uint8
}

// Verify verifies a multi-ed25519 signature against a message.
//
// The bitmap in the signature indicates which public keys were used to sign.
// Each bit corresponds to a public key index - if set, that key signed.
//
// Implements [VerifyingKey].
func (key *MultiEd25519PublicKey) Verify(msg []byte, signature Signature) bool {
	sig, ok := signature.(*MultiEd25519Signature)
	if !ok {
		return false
	}

	verified := uint8(0)
	sigIndex := 0

	// Use bitmap to determine which public keys signed and map them to signatures.
	for i, pubKey := range key.PubKeys {
		byteIndex := i / 8
		bitIndex := uint(7 - (i % 8)) // Bitmap is big-endian within each byte

		// If the bitmap does not have a byte for this index, treat as not signed.
		if byteIndex >= len(sig.Bitmap) {
			continue
		}

		if ((sig.Bitmap[byteIndex] >> bitIndex) & 0x1) == 0 {
			// This public key did not participate in signing.
			continue
		}

		// This public key is indicated as a signer; use the next signature.
		if sigIndex >= len(sig.Signatures) {
			// Malformed: bitmap indicates more signers than provided signatures.
			return false
		}

		if pubKey.Verify(msg, sig.Signatures[sigIndex]) {
			verified++
		}
		sigIndex++
	}

	return verified >= key.SignaturesRequired
}

// AuthKey returns the AuthenticationKey for this public key.
//
// Implements [PublicKey].
func (key *MultiEd25519PublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns MultiEd25519Scheme.
//
// Implements [PublicKey].
func (key *MultiEd25519PublicKey) Scheme() uint8 {
	return MultiEd25519Scheme
}

// Bytes returns the concatenated public keys plus threshold byte.
//
// Implements [CryptoMaterial].
func (key *MultiEd25519PublicKey) Bytes() []byte {
	bytes := make([]byte, len(key.PubKeys)*ed25519.PublicKeySize+1)
	for i, pk := range key.PubKeys {
		start := i * ed25519.PublicKeySize
		copy(bytes[start:start+ed25519.PublicKeySize], pk.Bytes())
	}
	bytes[len(bytes)-1] = key.SignaturesRequired
	return bytes
}

// FromBytes deserializes from concatenated keys + threshold byte.
//
// Implements [CryptoMaterial].
func (key *MultiEd25519PublicKey) FromBytes(bytes []byte) error {
	if len(bytes) < ed25519.PublicKeySize+1 {
		return errors.New("multi-ed25519 public key too short")
	}

	numKeys := (len(bytes) - 1) / ed25519.PublicKeySize
	key.PubKeys = make([]*Ed25519PublicKey, numKeys)

	for i := range numKeys {
		start := i * ed25519.PublicKeySize
		key.PubKeys[i] = &Ed25519PublicKey{}
		if err := key.PubKeys[i].FromBytes(bytes[start : start+ed25519.PublicKeySize]); err != nil {
			return fmt.Errorf("failed to parse key %d: %w", i, err)
		}
	}

	key.SignaturesRequired = bytes[len(bytes)-1]
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *MultiEd25519PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *MultiEd25519PublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *MultiEd25519PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *MultiEd25519PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := key.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// MultiEd25519Signature is a multi-sig signature using Ed25519.
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type MultiEd25519Signature struct {
	Signatures []*Ed25519Signature
	Bitmap     [MultiEd25519BitmapLen]byte
}

// Bytes returns the concatenated signatures plus bitmap.
//
// Implements [CryptoMaterial].
func (e *MultiEd25519Signature) Bytes() []byte {
	bytes := make([]byte, len(e.Signatures)*ed25519.SignatureSize+MultiEd25519BitmapLen)
	for i, sig := range e.Signatures {
		start := i * ed25519.SignatureSize
		copy(bytes[start:start+ed25519.SignatureSize], sig.Bytes())
	}
	copy(bytes[len(bytes)-MultiEd25519BitmapLen:], e.Bitmap[:])
	return bytes
}

// FromBytes deserializes from concatenated signatures + bitmap.
//
// Implements [CryptoMaterial].
func (e *MultiEd25519Signature) FromBytes(bytes []byte) error {
	if len(bytes) < ed25519.SignatureSize+MultiEd25519BitmapLen {
		return errors.New("multi-ed25519 signature too short")
	}

	numSigs := (len(bytes) - MultiEd25519BitmapLen) / ed25519.SignatureSize
	e.Signatures = make([]*Ed25519Signature, numSigs)

	for i := range numSigs {
		start := i * ed25519.SignatureSize
		e.Signatures[i] = &Ed25519Signature{}
		if err := e.Signatures[i].FromBytes(bytes[start : start+ed25519.SignatureSize]); err != nil {
			return fmt.Errorf("failed to parse signature %d: %w", i, err)
		}
	}

	copy(e.Bitmap[:], bytes[len(bytes)-MultiEd25519BitmapLen:])
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (e *MultiEd25519Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (e *MultiEd25519Signature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (e *MultiEd25519Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(e.Bytes())
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (e *MultiEd25519Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := e.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// MultiEd25519Authenticator is an authenticator for MultiEd25519 signatures.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type MultiEd25519Authenticator struct {
	PubKey *MultiEd25519PublicKey
	Sig    *MultiEd25519Signature
}

// PublicKey returns the multi-ed25519 public key.
//
// Implements [AccountAuthenticatorImpl].
func (ea *MultiEd25519Authenticator) PublicKey() PublicKey {
	return ea.PubKey
}

// Signature returns the multi-ed25519 signature.
//
// Implements [AccountAuthenticatorImpl].
func (ea *MultiEd25519Authenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the signature is valid.
//
// Implements [AccountAuthenticatorImpl].
func (ea *MultiEd25519Authenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (ea *MultiEd25519Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PubKey)
	ser.Struct(ea.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (ea *MultiEd25519Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &MultiEd25519PublicKey{}
	des.Struct(ea.PubKey)
	if des.Error() != nil {
		return
	}
	ea.Sig = &MultiEd25519Signature{}
	des.Struct(ea.Sig)
}
