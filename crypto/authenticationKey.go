package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

//region AuthenticationKey

// DeriveScheme is the key type for deriving the AuthenticationKey
type DeriveScheme = uint8

// Seeds for deriving addresses from addresses
const (
	Ed25519Scheme         DeriveScheme = 0
	MultiEd25519Scheme    DeriveScheme = 1
	SingleKeyScheme       DeriveScheme = 2
	MultiKeyScheme        DeriveScheme = 3
	DeriveObjectScheme    DeriveScheme = 252
	NamedObjectScheme     DeriveScheme = 254
	ResourceAccountScheme DeriveScheme = 255
)

// AuthenticationKeyLength is the length of a SHA3-256 Hash
const AuthenticationKeyLength = 32

// AuthenticationKey a hash representing the method for authorizing an account
type AuthenticationKey [AuthenticationKeyLength]byte

// FromPublicKey for private / public key pairs, the authentication key is derived from the public key directly
func (ak *AuthenticationKey) FromPublicKey(publicKey PublicKey) {
	bytes := util.Sha3256Hash([][]byte{
		publicKey.Bytes(),
		{publicKey.Scheme()},
	})
	copy((*ak)[:], bytes)
}

//region AuthenticationKey CryptoMaterial

func (ak *AuthenticationKey) Bytes() []byte {
	return ak[:]
}

func (ak *AuthenticationKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != AuthenticationKeyLength {
		return fmt.Errorf("invalid authentication key, not 32 bytes")
	}
	copy((*ak)[:], bytes)
	return nil
}

func (ak *AuthenticationKey) ToHex() string {
	return util.BytesToHex(ak[:])
}

func (ak *AuthenticationKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return ak.FromBytes(bytes)
}

//endregion

//region AuthenticationKey bcs.Struct

func (ak *AuthenticationKey) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(AuthenticationKeyLength)
	ser.FixedBytes(ak[:])
}

func (ak *AuthenticationKey) UnmarshalBCS(des *bcs.Deserializer) {
	length := des.Uleb128()
	if length != AuthenticationKeyLength {
		des.SetError(fmt.Errorf("authentication key has wrong length %d", length))
		return
	}
	des.ReadFixedBytesInto(ak[:])
}

//endregion
//endregion
