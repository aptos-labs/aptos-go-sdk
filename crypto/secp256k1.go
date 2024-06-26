package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

//region Secp256k1PrivateKey

// Secp256k1PrivateKeyLength is the [Secp256k1PrivateKey] length in bytes
const Secp256k1PrivateKeyLength = 32

// Secp256k1PublicKeyLength is the [Secp256k1PublicKey] length in bytes.  We use the uncompressed version.
const Secp256k1PublicKeyLength = 65

// Secp256k1SignatureLength is the [Secp256k1Signature] length in bytes.  It is a signature without the recovery bit.
const Secp256k1SignatureLength = ethCrypto.SignatureLength - 1

// Secp256k1PrivateKey is a private key that can be used with [SingleSigner].  It cannot stand on its own.
//
// Implements:
//   - [MessageSigner]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type Secp256k1PrivateKey struct {
	Inner *ecdsa.PrivateKey // Inner is the actual private key
}

// GenerateSecp256k1Key generates a new [Secp256k1PrivateKey]
func GenerateSecp256k1Key() (*Secp256k1PrivateKey, error) {
	priv, err := ethCrypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return &Secp256k1PrivateKey{priv}, nil
}

//region Secp256k1PrivateKey MessageSigner

// VerifyingKey returns the corresponding public key for the private key
//
// Implements:
//   - [MessageSigner]
func (key *Secp256k1PrivateKey) VerifyingKey() VerifyingKey {
	return &Secp256k1PublicKey{
		&key.Inner.PublicKey,
	}
}

// SignMessage signs a message and returns the raw [Signature] without a [PublicKey] for verification
//
// Implements:
//   - [MessageSigner]
func (key *Secp256k1PrivateKey) SignMessage(msg []byte) (sig Signature, err error) {
	hash := util.Sha3256Hash([][]byte{msg})
	// TODO: The eth library doesn't protect against malleability issues, so we need to handle those.
	signature, err := ethCrypto.Sign(hash, key.Inner)
	if err != nil {
		return nil, err
	}

	// Strip the recovery bit
	return &Secp256k1Signature{
		signature[0:64],
	}, nil
}

//endregion

//region Secp256k1PrivateKey CryptoMaterial

// Bytes outputs the raw byte representation of the [Secp256k1PrivateKey]
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PrivateKey) Bytes() []byte {
	return ethCrypto.FromECDSA(key.Inner)
}

// FromBytes populates the [Secp256k1PrivateKey] from bytes
//
// Returns an error if the bytes length is not [Secp256k1PrivateKeyLength]
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PrivateKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != Secp256k1PrivateKeyLength {
		return fmt.Errorf("invalid secp256k1 private key size %d", len(bytes))
	}
	key.Inner, err = ethCrypto.ToECDSA(bytes)
	return err
}

// ToHex serializes the private key to a hex string
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

//endregion

// FromHex populates the [Secp256k1PrivateKey] from a hex string
//
// Returns an error if the hex string is invalid or is not [Secp256k1PrivateKeyLength] bytes
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PrivateKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion
//endregion

//region Secp256k1PublicKey

// Secp256k1PublicKey is the corresponding public key for [Secp256k1PrivateKey], it cannot be used on its own
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type Secp256k1PublicKey struct {
	Inner *ecdsa.PublicKey // Inner is the actual public key
}

//region Secp256k1PublicKey VerifyingKey

// Verify verifies the signature of a message
//
// Returns true if the signature is valid and a [Secp256k1Signature], false otherwise
//
// Implements:
//   - [VerifyingKey]
func (key *Secp256k1PublicKey) Verify(msg []byte, sig Signature) bool {
	switch sig.(type) {
	case *Secp256k1Signature:
		typedSig := sig.(*Secp256k1Signature)

		// Verification requires to pass the SHA-256 hash of the message
		msg = util.Sha3256Hash([][]byte{msg})
		return ethCrypto.VerifySignature(key.Bytes(), msg, typedSig.Bytes())
	default:
		return false
	}
}

//endregion

//region Secp256k1PublicKey CryptoMaterial

// Bytes returns the raw bytes of the [Secp256k1PublicKey]
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PublicKey) Bytes() []byte {
	return ethCrypto.FromECDSAPub(key.Inner)
}

// FromBytes sets the [Secp256k1PublicKey] to the given bytes
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PublicKey) FromBytes(bytes []byte) (err error) {
	key.Inner, err = ethCrypto.UnmarshalPubkey(bytes)
	return err
}

// ToHex returns the hex string representation of the [Secp256k1PublicKey], with a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex sets the [Secp256k1PublicKey] to the bytes represented by the hex string, with or without a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (key *Secp256k1PublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion

//region Secp256k1PublicKey bcs.Struct

// MarshalBCS serializes the [Secp256k1PublicKey] to BCS bytes
//
// Implements:
//   - [bcs.Marshaler]
func (key *Secp256k1PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}

// UnmarshalBCS deserializes the [Secp256k1PublicKey] from BCS bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (key *Secp256k1PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	kb := des.ReadBytes()
	pubKey, err := ethCrypto.UnmarshalPubkey(kb)
	if err != nil {
		des.SetError(err)
		return
	}
	key.Inner = pubKey
}

//endregion
//endregion

//region Secp256k1Authenticator

// Secp256k1Authenticator is the authenticator for Secp256k1, but it cannot stand on its own and must be used with SingleKeyAuthenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type Secp256k1Authenticator struct {
	PubKey *Secp256k1PublicKey // PubKey is the public key
	Sig    *Secp256k1Signature // Sig is the signature
}

//region Secp256k1Authenticator AccountAuthenticatorImpl

// PublicKey returns the [VerifyingKey] for the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *Secp256k1Authenticator) PublicKey() VerifyingKey {
	return ea.PubKey
}

// Signature returns the [Signature] for the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *Secp256k1Authenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the authenticator can be cryptographically verified
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *Secp256k1Authenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

//endregion

//region Secp256k1Authenticator bcs.Struct

// MarshalBCS serializes the [Secp256k1Authenticator] to BCS bytes
//
// Implements:
//   - [bcs.Marshaler]
func (ea *Secp256k1Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PublicKey())
	ser.Struct(ea.Signature())
}

// UnmarshalBCS deserializes the [Secp256k1Authenticator] from BCS bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (ea *Secp256k1Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &Secp256k1PublicKey{}
	des.Struct(ea.PubKey)
	err := des.Error()
	if err != nil {
		return
	}
	ea.Sig = &Secp256k1Signature{}
	des.Struct(ea.Sig)
}

//endregion
//endregion

//region Secp256k1Signature

// Secp256k1Signature a wrapper for serialization of Secp256k1 signatures
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type Secp256k1Signature struct {
	Inner []byte // Inner is the actual signature
}

//region Secp256k1Signature CryptoMaterial

// Bytes returns the raw bytes of the [Secp256k1Signature]
//
// Implements:
//   - [CryptoMaterial]
func (e *Secp256k1Signature) Bytes() []byte {
	return e.Inner
}

// FromBytes sets the [Secp256k1Signature] to the given bytes
//
// Returns an error if the bytes length is not [Secp256k1SignatureLength]
//
// Implements:
// - [CryptoMaterial]
func (e *Secp256k1Signature) FromBytes(bytes []byte) (err error) {
	if len(bytes) != Secp256k1SignatureLength {
		return fmt.Errorf("invalid secp256k1 signature size %d, expected %d", len(bytes), Secp256k1SignatureLength)
	}
	e.Inner = bytes
	return nil
}

// ToHex returns the hex string representation of the [Secp256k1Signature], with a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (e *Secp256k1Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex sets the [Secp256k1Signature] to the bytes represented by the hex string, with or without a leading 0x
//
// Returns an error if the hex string is invalid or is not [Secp256k1SignatureLength] bytes
//
// Implements:
//   - [CryptoMaterial]
func (e *Secp256k1Signature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

//endregion

//region Secp256k1Signature bcs.Struct

// MarshalBCS serializes the [Secp256k1Signature] to BCS bytes
//
// Implements:
//   - [bcs.Marshaler]
func (e *Secp256k1Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(e.Bytes())
}

// UnmarshalBCS deserializes the [Secp256k1Signature] from BCS bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (e *Secp256k1Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	err := e.FromBytes(bytes)
	if err != nil {
		des.SetError(err)
	}
}

//endregion
//endregion
