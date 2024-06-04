package crypto

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"io"
)

//region Ed25519PrivateKey

// Ed25519PrivateKey represents an Ed25519Private key
// Implements Signer, MessageSigner, CryptoMaterial
type Ed25519PrivateKey struct {
	Inner ed25519.PrivateKey
}

// GenerateEd25519PrivateKey generates a random Ed25519 private key
func GenerateEd25519PrivateKey(rand ...io.Reader) (privateKey *Ed25519PrivateKey, err error) {
	var priv ed25519.PrivateKey
	if len(rand) > 0 {
		_, priv, err = ed25519.GenerateKey(rand[0])
	} else {
		_, priv, err = ed25519.GenerateKey(nil)
	}
	if err != nil {
		return nil, err
	}
	return &Ed25519PrivateKey{priv}, nil
}

//region Ed25519PrivateKey Signer Implementation

func (key *Ed25519PrivateKey) Sign(msg []byte) (authenticator *AccountAuthenticator, err error) {
	signature, err := key.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	publicKeyBytes := key.PubKey().Bytes()

	return &AccountAuthenticator{
		Variant: AccountAuthenticatorEd25519,
		Auth: &Ed25519Authenticator{
			PubKey: &Ed25519PublicKey{Inner: publicKeyBytes},
			Sig:    signature.(*Ed25519Signature),
		},
	}, nil
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

//endregion

//region Ed25519PrivateKey MessageSigner Implementation

func (key *Ed25519PrivateKey) SignMessage(msg []byte) (sig Signature, err error) {
	sigBytes := ed25519.Sign(key.Inner, msg)
	return &Ed25519Signature{Inner: [64]byte(sigBytes)}, nil
}

func (key *Ed25519PrivateKey) VerifyingKey() VerifyingKey {
	return key.PubKey()
}

//endregion

//region Ed25519PrivateKey CryptoMaterial Implementation

func (key *Ed25519PrivateKey) Bytes() []byte {
	return key.Inner.Seed()
}

func (key *Ed25519PrivateKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.SeedSize {
		return fmt.Errorf("invalid ed25519 private key size %d", len(bytes))
	}
	key.Inner = ed25519.NewKeyFromSeed(bytes)
	return nil
}

func (key *Ed25519PrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

func (key *Ed25519PrivateKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion

//endregion

//region Ed25519PublicKey

// Ed25519PublicKey is a Ed25519PublicKey which can be used to verify signatures
// Implements VerifyingKey, PublicKey, CryptoMaterial, bcs.Struct
type Ed25519PublicKey struct {
	Inner ed25519.PublicKey
}

//region Ed25519PublicKey VerifyingKey implementation

func (key *Ed25519PublicKey) Verify(msg []byte, sig Signature) bool {
	switch sig.(type) {
	case *Ed25519Signature:
		return ed25519.Verify(key.Inner, msg, sig.Bytes())
	default:
		return false
	}
}

//endregion

//region Ed25519PublicKey PublicKey implementation

func (key *Ed25519PublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

func (key *Ed25519PublicKey) Scheme() uint8 {
	return Ed25519Scheme
}

//endregion

//region Ed25519PublicKey CryptoMaterial implementation

func (key *Ed25519PublicKey) Bytes() []byte {
	return key.Inner[:]
}

func (key *Ed25519PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

func (key *Ed25519PublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

func (key *Ed25519PublicKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.PublicKeySize {
		return errors.New("invalid ed25519 public key size")
	}
	key.Inner = bytes
	return nil
}

//endregion

//region Ed25519PublicKey bcs.Struct implementation

func (key *Ed25519PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Inner)
}
func (key *Ed25519PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	kb := des.ReadBytes()
	if len(kb) != ed25519.PublicKeySize {
		des.SetError(fmt.Errorf("bad ed25519 public key, expected %d bytes but got %d", ed25519.PublicKeySize, len(kb)))
		return
	}
	key.Inner = kb
}

//endregion
//endregion

//region Ed25519Authenticator

// Ed25519Authenticator represents a verifiable signature with it's accompanied public key
// Implements AccountAuthenticatorImpl
type Ed25519Authenticator struct {
	PubKey *Ed25519PublicKey
	Sig    *Ed25519Signature
}

//region Ed25519Authenticator AccountAuthenticatorImpl implementation

func (ea *Ed25519Authenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *Ed25519Authenticator) Signature() Signature {
	return ea.Sig
}

func (ea *Ed25519Authenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

//endregion

//region Ed25519Authenticator bcs.Struct implementation

func (ea *Ed25519Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PublicKey())
	ser.Struct(ea.Signature())
}

func (ea *Ed25519Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &Ed25519PublicKey{}
	des.Struct(ea.PubKey)
	err := des.Error()
	if err != nil {
		return
	}
	ea.Sig = &Ed25519Signature{}
	des.Struct(ea.Sig)
}

//endregion
//endregion

//region Ed25519Signature

// Ed25519Signature a wrapper for serialization of Ed25519 signatures
// Implements Signature, CryptoMaterial, bcs.Struct
type Ed25519Signature struct {
	Inner [ed25519.SignatureSize]byte
}

//region Ed25519Signature CryptoMaterial implementation

func (e *Ed25519Signature) Bytes() []byte {
	return e.Inner[:]
}

func (e *Ed25519Signature) FromBytes(bytes []byte) (err error) {
	if len(bytes) != ed25519.SignatureSize {
		return errors.New("invalid ed25519 signature size")
	}
	copy(e.Inner[:], bytes)
	return nil
}

func (e *Ed25519Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

func (e *Ed25519Signature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

//endregion

//region Ed25519Signature bcs.Struct implementation

func (e *Ed25519Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(e.Bytes())
}

func (e *Ed25519Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if len(bytes) != ed25519.SignatureSize {
		des.SetError(fmt.Errorf("cannot deserialize ed25519 signature, expected %d bytes but got %d", ed25519.SignatureSize, len(bytes)))
		return
	}
	copy(e.Inner[:], bytes)
}

//endregion
//endregion
