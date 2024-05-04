package aptos

import (
	"crypto/ed25519"
	"fmt"
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

type Authenticator struct {
	Kind AuthenticatorType
	Auth AuthenticatorImpl
}

func (ea *Authenticator) MarshalBCS(bcs *Serializer) {
	bcs.Uleb128(uint64(ea.Kind))
	ea.Auth.MarshalBCS(bcs)
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

func (ea *Authenticator) Verify(data []byte) bool {
	return ea.Auth.Verify(data)
}

type AuthenticatorImpl interface {
	BCSStruct

	// Verify Return true if this Authenticator approves
	Verify(data []byte) bool
}

type Ed25519Authenticator struct {
	PublicKey [ed25519.PublicKeySize]byte
	Signature [ed25519.SignatureSize]byte
}

func (ea *Ed25519Authenticator) MarshalBCS(bcs *Serializer) {
	bcs.WriteBytes(ea.PublicKey[:])
	bcs.WriteBytes(ea.Signature[:])
}
func (ea *Ed25519Authenticator) UnmarshalBCS(bcs *Deserializer) {
	kb := bcs.ReadBytes()
	if len(kb) != ed25519.PublicKeySize {
		bcs.SetError(fmt.Errorf("bad ed25519 public key, expected %d bytes but got %d", ed25519.PublicKeySize, len(kb)))
		return
	}
	sb := bcs.ReadBytes()
	if len(sb) != ed25519.SignatureSize {
		bcs.SetError(fmt.Errorf("bad ed25519 signature, expected %d bytes but got %d", ed25519.SignatureSize, len(sb)))
		return
	}
	copy(ea.PublicKey[:], kb)
	copy(ea.Signature[:], sb)
}

// Verify Return true if the data was well signed
func (ea *Ed25519Authenticator) Verify(data []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(ea.PublicKey[:]), data, ea.Signature[:])
}

// TODO: FeePayerAuthenticator, MultiAgentAuthenticator, MultiEd25519Authenticator, SingleSenderAuthenticator, SingleKeyAuthenticator
