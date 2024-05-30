package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

type Ed25519PrivateKey struct {
	Inner ed25519.PrivateKey
}

func GenerateEd5519Keys() (privateKey Ed25519PrivateKey, publicKey Ed25519PublicKey, err error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return
	}
	privateKey = Ed25519PrivateKey{priv}
	publicKey = Ed25519PublicKey{pub}
	return
}

func (key *Ed25519PrivateKey) PubKey() PublicKey {
	pubKey := key.Inner.Public()
	return &Ed25519PublicKey{
		pubKey.(ed25519.PublicKey),
	}
}

func (key *Ed25519PrivateKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key.PubKey())
	return out
}

func (key *Ed25519PrivateKey) Bytes() []byte {
	return key.Inner.Seed()
}

func (key *Ed25519PrivateKey) ToHex() string {
	return "0x" + hex.EncodeToString(key.Bytes())
}

func (key *Ed25519PrivateKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

func (key *Ed25519PrivateKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.SeedSize {
		return fmt.Errorf("invalid ed25519 private key size %d", len(bytes))
	}
	key.Inner = ed25519.NewKeyFromSeed(bytes)
	return nil
}

func (key *Ed25519PrivateKey) Sign(msg []byte) (authenticator *Authenticator, err error) {
	publicKeyBytes := key.PubKey().Bytes()
	signature := ed25519.Sign(key.Inner, msg)

	auth := &Ed25519Authenticator{}
	auth.PubKey = &Ed25519PublicKey{Inner: publicKeyBytes}

	var sigBytes [ed25519.SignatureSize]byte
	copy(sigBytes[:], signature[:])
	auth.Sig = &Ed25519Signature{Inner: sigBytes}
	authenticator = &Authenticator{
		AuthenticatorEd25519,
		auth,
	}
	return
}

type Ed25519PublicKey struct {
	Inner ed25519.PublicKey
}

func (key *Ed25519PublicKey) Bytes() []byte {
	return key.Inner[:]
}

func (key *Ed25519PublicKey) Scheme() uint8 {
	return Ed25519Scheme
}

func (key *Ed25519PublicKey) ToHex() string {
	return "0x" + hex.EncodeToString(key.Bytes())
}

func (key *Ed25519PublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	if len(bytes) != ed25519.PublicKeySize {
		return errors.New("invalid ed25519 public key size")
	}
	key.Inner = bytes
	return nil
}

func (key *Ed25519PublicKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.PublicKeySize {
		return errors.New("invalid ed25519 public key size")
	}
	key.Inner = bytes
	return nil
}

func (key *Ed25519PublicKey) Verify(msg []byte, sig Signature) bool {
	return ed25519.Verify(key.Inner, msg, sig.Bytes())
}

func (key *Ed25519PublicKey) MarshalBCS(bcs *bcs.Serializer) {
	bcs.WriteBytes(key.Inner)
}
func (key *Ed25519PublicKey) UnmarshalBCS(bcs *bcs.Deserializer) {
	kb := bcs.ReadBytes()
	if len(kb) != ed25519.PublicKeySize {
		bcs.SetError(fmt.Errorf("bad ed25519 public key, expected %d bytes but got %d", ed25519.PublicKeySize, len(kb)))
		return
	}
	key.Inner = kb
}

type Ed25519Authenticator struct {
	PubKey *Ed25519PublicKey
	Sig    *Ed25519Signature
}

func (ea *Ed25519Authenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *Ed25519Authenticator) Signature() Signature {
	return ea.Sig
}

func (ea *Ed25519Authenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Struct(ea.PublicKey())
	bcs.Struct(ea.Signature())
}

func (ea *Ed25519Authenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.PubKey = &Ed25519PublicKey{}
	bcs.Struct(ea.PubKey)
	err := bcs.Error()
	if err != nil {
		return
	}
	ea.Sig = &Ed25519Signature{}
	bcs.Struct(ea.Sig)
}

// Verify Return true if the data was well signed
func (ea *Ed25519Authenticator) Verify(msg []byte) bool {
	return ea.Sig.Verify(ea.PubKey, msg)
}

// Ed25519Signature a wrapper for serialization of Ed25519 signatures
type Ed25519Signature struct {
	Inner [ed25519.SignatureSize]byte
}

func (e *Ed25519Signature) Bytes() []byte {
	return e.Inner[:]
}

func (e *Ed25519Signature) MarshalBCS(bcs *bcs.Serializer) {
	bcs.WriteBytes(e.Bytes())
}

func (e *Ed25519Signature) UnmarshalBCS(bcs *bcs.Deserializer) {
	bytes := bcs.ReadBytes()
	if bcs.Error() != nil {
		return
	}
	if len(bytes) != ed25519.SignatureSize {
		bcs.SetError(fmt.Errorf("cannot deserialize ed25519 signature, expected %d bytes but got %d", ed25519.SignatureSize, len(bytes)))
		return
	}
	copy(e.Inner[:], bytes)
}

func (e *Ed25519Signature) Verify(publicKey *Ed25519PublicKey, msg []byte) bool {
	return publicKey.Verify(msg, e)
}

func (e *Ed25519Signature) ToHex() string {
	return "0x" + hex.EncodeToString(e.Bytes())
}

func (e *Ed25519Signature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	if len(bytes) != ed25519.SignatureSize {
		return errors.New("invalid ed25519 signature size")
	}
	copy(e.Inner[:], bytes)
	return nil
}

func (e *Ed25519Signature) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.SignatureSize {
		return errors.New("invalid ed25519 signature size")
	}
	copy(e.Inner[:], bytes)
	return nil
}
