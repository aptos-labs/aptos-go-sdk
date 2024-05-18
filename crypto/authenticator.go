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
	AuthenticatorMultiAgent   AuthenticatorType = 2
	AuthenticatorFeePayer     AuthenticatorType = 3
	AuthenticatorSingleSender AuthenticatorType = 4
)

// AuthenticatorImpl an implementation of an authenticator to provide generic verification across multiple types
type AuthenticatorImpl interface {
	bcs.Struct

	// Verify Return true if this Authenticator approves
	Verify(data []byte) bool
}

// Authenticator a generic authenticator type for a transaction
type Authenticator struct {
	Kind AuthenticatorType
	Auth AuthenticatorImpl
}

func (ea *Authenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(ea.Kind))
	ea.Auth.MarshalBCS(bcs)
}

func (ea *Authenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	kindu := bcs.Uleb128()
	if bcs.Error() != nil {
		return
	}
	kind := AuthenticatorType(kindu)
	switch kind {
	case AuthenticatorEd25519:
		auth := &Ed25519Authenticator{}
		auth.UnmarshalBCS(bcs)
		ea.Auth = auth
	default:
		bcs.SetError(fmt.Errorf("unknown Authenticator kind: %d", kindu))
	}
}

// Verify verifies a message with the public key and signature
func (ea *Authenticator) Verify(data []byte) bool {
	return ea.Auth.Verify(data)
}

type AuthenticationKey [32]byte

func (ak *AuthenticationKey) FromPublicKey(pubkey PublicKey) {
	bytes := util.SHA3_256Hash([][]byte{
		pubkey.Bytes(),
		{pubkey.Scheme()},
	})
	copy((*ak)[:], bytes)
}

// TODO: FeePayerAuthenticator, MultiAgentAuthenticator, MultiEd25519Authenticator, SingleSenderAuthenticator, SingleKeyAuthenticator
