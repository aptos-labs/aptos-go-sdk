package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
)

var (
	materials map[string]CryptoMaterial
	structs   map[string]bcs.Struct
)

func init() {
	msg := []byte("hello")
	var err error
	materials = make(map[string]CryptoMaterial)
	structs = make(map[string]bcs.Struct)
	ed25519PrivateKey, err := GenerateEd25519PrivateKey()
	if err != nil {
		panic("Failed to create ed25519 private key" + err.Error())
	}
	materials["ed25519PrivateKey"] = ed25519PrivateKey
	materials["ed25519PublicKey"] = ed25519PrivateKey.PubKey()
	structs["ed25519PublicKey"] = ed25519PrivateKey.PubKey()
	materials["ed25519VerifyingKey"] = ed25519PrivateKey.VerifyingKey()
	structs["ed25519VerifyingKey"] = ed25519PrivateKey.VerifyingKey()
	materials["ed25519AuthKey"] = ed25519PrivateKey.AuthKey()
	structs["ed25519AuthKey"] = ed25519PrivateKey.AuthKey()
	ed25519Sig, _ := ed25519PrivateKey.SignMessage(msg)
	materials["ed25519Signature"] = ed25519Sig
	structs["ed25519Signature"] = ed25519Sig
	ed25519Authenticator, _ := ed25519PrivateKey.Sign(msg)
	structs["ed25519Authenticator"] = ed25519Authenticator

	// Wrap in a single sender
	ed25519SingleSender := NewSingleSigner(ed25519PrivateKey)
	materials["singleSenderEd25519PublicKey"] = ed25519SingleSender.PubKey()
	structs["singleSenderEd25519PublicKey"] = ed25519SingleSender.PubKey()
	materials["singleSenderEd25519AuthKey"] = ed25519SingleSender.AuthKey()
	structs["singleSenderEd25519AuthKey"] = ed25519SingleSender.AuthKey()
	ed25519SingleSenderSig, _ := ed25519SingleSender.SignMessage(msg)
	materials["singleSenderEd25519Signature"] = ed25519SingleSenderSig
	structs["singleSenderEd25519Signature"] = ed25519SingleSenderSig
	ed25519SingleSenderAuthenticator, _ := ed25519SingleSender.Sign(msg)
	structs["singleSenderEd25519Authenticator"] = ed25519SingleSenderAuthenticator

	secp256k1PrivateKey, err := GenerateSecp256k1Key()
	if err != nil {
		panic("Failed to create secp256k1 private key" + err.Error())
	}
	materials["secp256k1PrivateKey"] = secp256k1PrivateKey
	materials["secp256k1VerifyingKey"] = secp256k1PrivateKey.VerifyingKey()
	structs["secp256k1VerifyingKey"] = secp256k1PrivateKey.VerifyingKey()

	// Wrap in a single sender
	secp256k1SingleSender := NewSingleSigner(secp256k1PrivateKey)
	materials["singleSenderSecp256k1PublicKey"] = secp256k1SingleSender.PubKey()
	structs["singleSenderSecp256k1PublicKey"] = secp256k1SingleSender.PubKey()
	materials["singleSenderSecp256k1AuthKey"] = secp256k1SingleSender.AuthKey()
	structs["singleSenderSecp256k1AuthKey"] = secp256k1SingleSender.AuthKey()
	secp256k1SingleSenderSig, _ := secp256k1SingleSender.SignMessage(msg)
	materials["singleSenderEd25519Signature"] = secp256k1SingleSenderSig
	structs["singleSenderEd25519Signature"] = secp256k1SingleSenderSig
	secp256k1SingleSenderAuthenticator, _ := secp256k1SingleSender.Sign(msg)
	structs["singleSenderEd25519Authenticator"] = secp256k1SingleSenderAuthenticator

	// generate 2 keys
	key2, err := GenerateEd25519PrivateKey()
	if err != nil {
		panic("Failed to create ed25519 private key" + err.Error())
	}

	multiEd25519 := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{ed25519PrivateKey.PubKey().(*Ed25519PublicKey), key2.PubKey().(*Ed25519PublicKey)},
		SignaturesRequired: 1,
	}
	materials["multied25519"] = multiEd25519
	structs["multied25519"] = multiEd25519
	materials["multied25519AuthKey"] = multiEd25519.AuthKey()
	structs["multied25519AuthKey"] = multiEd25519.AuthKey()

	multiKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{ed25519SingleSender.PubKey().(*AnyPublicKey), secp256k1SingleSender.PubKey().(*AnyPublicKey)},
		SignaturesRequired: 1,
	}
	materials["multiKey"] = multiKey
	structs["multiKey"] = multiKey
	materials["multiKeyAuthkey"] = multiKey.AuthKey()
	structs["multiKeyAuthKey"] = multiKey.AuthKey()
}

func TestCryptoMaterial(t *testing.T) {
	t.Parallel()
	for name, material := range materials {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, material.FromBytes(material.Bytes()))
			assert.NoError(t, material.FromHex(material.ToHex()))
		})
	}
}

func TestStructs(t *testing.T) {
	t.Parallel()
	for name, str := range structs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			bytes, err := bcs.Serialize(str)
			assert.NoError(t, err)
			assert.NoError(t, bcs.Deserialize(str, bytes))
		})
	}
}
