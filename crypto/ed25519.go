package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"github.com/aptos-labs/aptos-go-sdk/util"
)

type Ed25519PrivateKey struct {
	inner ed25519.PrivateKey
}

func GenerateEd5519Keys() (privkey Ed25519PrivateKey, pubkey Ed25519PublicKey, err error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return
	}
	privkey = Ed25519PrivateKey{priv}
	pubkey = Ed25519PublicKey{pub}
	return
}

func (key *Ed25519PrivateKey) PubKey() PublicKey {
	pubkey := key.inner.Public()
	return &Ed25519PublicKey{
		pubkey.(ed25519.PublicKey),
	}
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
