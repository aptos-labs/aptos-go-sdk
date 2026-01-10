package crypto

import (
	"errors"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// DeriveScheme identifies the key type used to derive an AuthenticationKey.
// The scheme byte is appended to the public key bytes before hashing.
type DeriveScheme = uint8

// Derive schemes for different key types and address derivation methods.
const (
	Ed25519Scheme         DeriveScheme = 0   // Standard Ed25519 accounts
	MultiEd25519Scheme    DeriveScheme = 1   // Multi-sig Ed25519 accounts
	SingleKeyScheme       DeriveScheme = 2   // Single-key accounts (supports multiple key types)
	MultiKeyScheme        DeriveScheme = 3   // Multi-key accounts (mixed key types)
	DeriveObjectScheme    DeriveScheme = 252 // Object address derivation
	NamedObjectScheme     DeriveScheme = 254 // Named object address derivation
	ResourceAccountScheme DeriveScheme = 255 // Resource account address derivation
)

// AuthenticationKeyLength is the length of an authentication key (SHA3-256 hash).
const AuthenticationKeyLength = 32

// AuthenticationKey is a SHA3-256 hash that identifies how an account is authenticated.
// It is derived from a public key and a scheme identifier.
//
// The authentication key equals the account address for most accounts,
// but can differ if the account's authentication key has been rotated.
//
// Implements:
//   - [CryptoMaterial]
//   - [bcs.Struct]
type AuthenticationKey [AuthenticationKeyLength]byte

// FromPublicKey derives an AuthenticationKey from a PublicKey.
func (ak *AuthenticationKey) FromPublicKey(publicKey PublicKey) {
	ak.FromBytesAndScheme(publicKey.Bytes(), publicKey.Scheme())
}

// FromBytesAndScheme derives an AuthenticationKey by hashing the bytes with the scheme.
func (ak *AuthenticationKey) FromBytesAndScheme(bytes []byte, scheme DeriveScheme) {
	hash := util.Sha3256Hash([][]byte{bytes, {scheme}})
	copy(ak[:], hash)
}

// Bytes returns the raw bytes of the AuthenticationKey.
//
// Implements [CryptoMaterial].
func (ak *AuthenticationKey) Bytes() []byte {
	return ak[:]
}

// FromBytes sets the AuthenticationKey from raw bytes.
// Returns an error if bytes is not 32 bytes.
//
// Implements [CryptoMaterial].
func (ak *AuthenticationKey) FromBytes(bytes []byte) error {
	if len(bytes) != AuthenticationKeyLength {
		return errors.New("authentication key must be 32 bytes")
	}
	copy(ak[:], bytes)
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (ak *AuthenticationKey) ToHex() string {
	return util.BytesToHex(ak[:])
}

// FromHex parses a hex string with optional "0x" prefix.
//
// Implements [CryptoMaterial].
func (ak *AuthenticationKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return ak.FromBytes(bytes)
}

// MarshalBCS serializes the AuthenticationKey to BCS.
//
// Implements [bcs.Marshaler].
func (ak *AuthenticationKey) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(AuthenticationKeyLength)
	ser.FixedBytes(ak[:])
}

// UnmarshalBCS deserializes the AuthenticationKey from BCS.
//
// Implements [bcs.Unmarshaler].
func (ak *AuthenticationKey) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	if length != AuthenticationKeyLength {
		des.SetError(fmt.Errorf("authentication key has wrong length %d", length))
		return
	}
	des.ReadFixedBytesInto(ak[:])
}
