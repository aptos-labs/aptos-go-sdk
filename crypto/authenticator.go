package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// Seeds for deriving addresses from addresses
const (
	Ed25519Scheme         = uint8(0)
	MultiEd25519Scheme    = uint8(1)
	SingleKeyScheme       = uint8(2)
	MultiKeyScheme        = uint8(3)
	DeriveObjectScheme    = uint8(252)
	NamedObjectScheme     = uint8(254)
	ResourceAccountScheme = uint8(255)
)

// AuthenticatorType single byte representing the spot in the enum from the Rust implementation
type AuthenticatorType uint8

const (
	AuthenticatorEd25519      AuthenticatorType = 0
	AuthenticatorMultiEd25519 AuthenticatorType = 1
	AuthenticatorSingleKey    AuthenticatorType = 2
	AuthenticatorMultiKey     AuthenticatorType = 3
)

// AuthenticatorImpl an implementation of an authenticator to provide generic verification across multiple types
type AuthenticatorImpl interface {
	bcs.Struct

	PublicKey() PublicKey

	// Signature return the signature bytes
	Signature() Signature

	// Verify Return true if this Authenticator approves
	Verify(data []byte) bool
}

// Authenticator a generic authenticator type for a transaction
type Authenticator struct {
	Kind AuthenticatorType
	Auth AuthenticatorImpl
}

func (ea *Authenticator) PublicKey() PublicKey {
	return ea.Auth.PublicKey()
}

func (ea *Authenticator) Signature() Signature {
	return ea.Auth.Signature()
}

func (ea *Authenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(ea.Kind))
	ea.Auth.MarshalBCS(bcs)
}

func (ea *Authenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	kindNum := bcs.Uleb128()
	if bcs.Error() != nil {
		return
	}
	ea.Kind = AuthenticatorType(kindNum)
	switch ea.Kind {
	case AuthenticatorEd25519:
		ea.Auth = &Ed25519Authenticator{}
	case AuthenticatorMultiEd25519:
		ea.Auth = &MultiEd25519Authenticator{}
	case AuthenticatorSingleKey:
		ea.Auth = &SingleKeyAuthenticator{}
	case AuthenticatorMultiKey:
		ea.Auth = &MultiKeyAuthenticator{}
	default:
		bcs.SetError(fmt.Errorf("unknown Authenticator kind: %d", kindNum))
		return
	}
	ea.Auth.UnmarshalBCS(bcs)
}

// Verify verifies a message with the public key and signature
func (ea *Authenticator) Verify(data []byte) bool {
	return ea.Auth.Verify(data)
}

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

func (ak *AuthenticationKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return ak.FromBytes(bytes)
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

func (ak *AuthenticationKey) Bytes() []byte {
	return ak[:]
}

func (ak *AuthenticationKey) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(AuthenticationKeyLength)
	bcs.FixedBytes(ak[:])
}

func (ak *AuthenticationKey) UnmarshalBCS(bcs *bcs.Deserializer) {
	length := bcs.Uleb128()
	if length != AuthenticationKeyLength {
		bcs.SetError(fmt.Errorf("authentication key has wrong length %d", length))
	}
	bcs.ReadFixedBytesInto(ak[:])
}
