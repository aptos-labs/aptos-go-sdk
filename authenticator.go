package aptos

import (
	"crypto/ed25519"
	"fmt"
)

type AuthenticatorType int

const (
	AuthenticatorEd25519      AuthenticatorType = 0
	AuthenticatorMultiEd25519 AuthenticatorType = 1
	AuthenticatorMultiAgent   AuthenticatorType = 2
	AuthenticatorFeePayer     AuthenticatorType = 3
	AuthenticatorSingleSender AuthenticatorType = 4
)

type Authenticator struct {
	Kind AuthenticatorType
	Auth AuthenticatorImpl
}

func (ea *Authenticator) MarshalBCS(bcs *Serializer) {
	bcs.Uleb128(uint64(ea.Kind))
	bcs.Struct(ea.Auth)
}
func (ea *Authenticator) UnmarshalBCS(bcs *Deserializer) {
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

type AuthenticatorImpl interface {
	BCSStruct
	Verify(data []byte)
}

type Ed25519Authenticator struct {
	PublicKey [ed25519.PublicKeySize]byte
	Signature [ed25519.SignatureSize]byte
}

func (ea *Ed25519Authenticator) MarshalBCS(*Serializer) {

}
func (ea *Ed25519Authenticator) UnmarshalBCS(*Deserializer) {
}
func (ea *Ed25519Authenticator) Verify(data []byte) {
}
