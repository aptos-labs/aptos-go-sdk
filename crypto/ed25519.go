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
	inner ed25519.PrivateKey
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
	pubKey := key.inner.Public()
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
	return key.inner[:]
}

func (key *Ed25519PrivateKey) ToHex() string {
	return hex.EncodeToString(key.Bytes())
}

func (key *Ed25519PrivateKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

func (key *Ed25519PrivateKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.PrivateKeySize {
		return errors.New("invalid ed25519 private key size")
	}
	key.inner = bytes
	return nil
}

func (key *Ed25519PrivateKey) Sign(msg []byte) (authenticator Authenticator, err error) {
	publicKeyBytes := key.PubKey().Bytes()
	signature := ed25519.Sign(key.inner, msg)

	auth := &Ed25519Authenticator{}
	copy(auth.PublicKey[:], publicKeyBytes[:])
	copy(auth.Signature[:], signature[:]) // TODO: Signature type?
	authenticator = Authenticator{
		AuthenticatorEd25519,
		auth,
	}
	return
}

type Ed25519PublicKey struct {
	inner ed25519.PublicKey
}

func (key *Ed25519PublicKey) Bytes() []byte {
	return key.inner[:]
}

func (key *Ed25519PublicKey) Scheme() uint8 {
	return Ed25519Scheme
}

func (key *Ed25519PublicKey) ToHex() string {
	return hex.EncodeToString(key.Bytes())
}

func (key *Ed25519PublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	if len(bytes) != ed25519.PublicKeySize {
		return errors.New("invalid ed25519 public key size")
	}
	key.inner = bytes
	return nil
}

type Ed25519Authenticator struct {
	PublicKey [ed25519.PublicKeySize]byte
	Signature [ed25519.SignatureSize]byte
}

func (ea *Ed25519Authenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.WriteBytes(ea.PublicKey[:])
	bcs.WriteBytes(ea.Signature[:])
}
func (ea *Ed25519Authenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
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
