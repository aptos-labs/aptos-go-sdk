package crypto

import (
	"crypto/ed25519"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

type MultiEd25519PublicKey struct {
	PublicKeys []Ed25519PublicKey
	Threshold  uint8
}

func (key *MultiEd25519PublicKey) Bytes() []byte {
	bytes, _ := bcs.Serialize(key)
	return bytes
}

func (key *MultiEd25519PublicKey) Scheme() uint8 {
	return MultiEd25519Scheme
}

func (key *MultiEd25519PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

func (key *MultiEd25519PublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

func (key *MultiEd25519PublicKey) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(key, bytes)
}

func (key *MultiEd25519PublicKey) Verify(_ []byte, _ Signature) bool {
	panic("TODO implement")
}

func (key *MultiEd25519PublicKey) MarshalBCS(bcs *bcs.Serializer) {
	// This is a weird one, we need to serialize in set bytes
	keyBytes := make([]byte, len(key.PublicKeys)*ed25519.PublicKeySize+1)
	for i, publicKey := range key.PublicKeys {
		start := i * ed25519.PublicKeySize
		end := start + ed25519.PublicKeySize
		copy(keyBytes[start:end], publicKey.Bytes()[:])
	}
	keyBytes[len(keyBytes)-1] = key.Threshold
	bcs.WriteBytes(keyBytes)
}

func (key *MultiEd25519PublicKey) UnmarshalBCS(bcs *bcs.Deserializer) {
	keyBytes := bcs.ReadBytes()
	key.Threshold = keyBytes[len(keyBytes)-1]

	key.PublicKeys = make([]Ed25519PublicKey, len(keyBytes)/ed25519.PublicKeySize)
	for i := 0; i < len(keyBytes); i++ {
		start := i * ed25519.PublicKeySize
		end := start + ed25519.PublicKeySize
		key.PublicKeys[i] = Ed25519PublicKey{}
		err := key.PublicKeys[i].FromBytes(keyBytes[start:end])
		if err != nil {
			bcs.SetError(fmt.Errorf("failed to deserialize multi ed25519 public key sub key %d: %w", i, err))
			return
		}
	}
}

type MultiEd25519Authenticator struct {
	PubKey *MultiEd25519PublicKey
	Sig    *MultiEd25519Signature
}

func (ea *MultiEd25519Authenticator) PublicKey() PublicKey {
	return ea.PubKey
}

func (ea *MultiEd25519Authenticator) Signature() Signature {
	return ea.Sig
}

func (ea *MultiEd25519Authenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Struct(ea.PublicKey())
	bcs.Struct(ea.Signature())
}

func (ea *MultiEd25519Authenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.PubKey = &MultiEd25519PublicKey{}
	bcs.Struct(ea.PubKey)
	err := bcs.Error()
	if err != nil {
		return
	}
	ea.Sig = &MultiEd25519Signature{}
	bcs.Struct(ea.Sig)
}

// Verify Return true if the data was well signed
func (ea *MultiEd25519Authenticator) Verify(msg []byte) bool {
	return ea.Sig.Verify(ea.PubKey, msg)
}

type MultiEd25519Signature struct {
	Signatures []Ed25519Signature
	Bitmap     []byte
}

func (e *MultiEd25519Signature) Bytes() []byte {
	bytes, _ := bcs.Serialize(e)
	return bytes
}

func (e *MultiEd25519Signature) MarshalBCS(bcs *bcs.Serializer) {
	// This is a weird one, we need to serialize in set bytes
	sigBytes := make([]byte, len(e.Signatures)*ed25519.SignatureSize+4) // TODO: move to a constant
	for i, signature := range e.Signatures {
		start := i * ed25519.SignatureSize
		end := start + ed25519.SignatureSize
		copy(sigBytes[start:end], signature.Bytes()[:])
	}
	copy(sigBytes[len(sigBytes)-4:], e.Bitmap)
	bcs.WriteBytes(sigBytes)
}

func (e *MultiEd25519Signature) UnmarshalBCS(bcs *bcs.Deserializer) {
	sigBytes := bcs.ReadBytes()
	e.Bitmap = sigBytes[len(sigBytes)-4:]

	e.Signatures = make([]Ed25519Signature, len(sigBytes)/ed25519.SignatureSize)
	for i := 0; i < len(sigBytes); i++ {
		start := i * ed25519.PublicKeySize
		end := start + ed25519.PublicKeySize
		e.Signatures[i] = Ed25519Signature{}
		err := e.Signatures[i].FromBytes(sigBytes[start:end])
		if err != nil {
			bcs.SetError(fmt.Errorf("failed to deserialize multi ed25519 signature sub signature %d: %w", i, err))
			return
		}
	}
}

func (e *MultiEd25519Signature) Verify(publicKey *MultiEd25519PublicKey, msg []byte) bool {
	return publicKey.Verify(msg, e)
}

func (e *MultiEd25519Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

func (e *MultiEd25519Signature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

func (e *MultiEd25519Signature) FromBytes(bytes []byte) (err error) {
	return bcs.Deserialize(e, bytes)
}
