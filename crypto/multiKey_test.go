package crypto

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMultiKey(t *testing.T) {
	key1, key2, _, _, publicKey := createMultiKey(t)

	message := []byte("hello world")

	signature := createMultiKeySignature(t, key1, key2, message)

	// Test verification of signature
	assert.True(t, publicKey.Verify(message, signature))

	// Test serialization / deserialization authenticator
	auth := &MultiKeyAuthenticator{
		PubKey: publicKey,
		Sig:    signature,
	}
	assert.True(t, auth.Verify(message))
}

func TestMultiKeySerialization(t *testing.T) {
	key1, key2, _, _, publicKey := createMultiKey(t)

	// Test serialization / deserialization public key
	keyBytes, err := bcs.Serialize(publicKey)
	assert.NoError(t, err)
	publicKeyDeserialized := &MultiKey{}
	err = bcs.Deserialize(publicKeyDeserialized, keyBytes)
	assert.NoError(t, err)
	assert.Equal(t, publicKey, publicKeyDeserialized)

	// Test serialization / deserialization signature
	signature := createMultiKeySignature(t, key1, key2, []byte("test message"))
	sigBytes, err := bcs.Serialize(signature)
	assert.NoError(t, err)
	signatureDeserialized := &MultiKeySignature{}
	err = bcs.Deserialize(signatureDeserialized, sigBytes)
	assert.NoError(t, err)
	assert.Equal(t, signature, signatureDeserialized)

	// Test serialization / deserialization authenticator
	auth := &AccountAuthenticator{
		Variant: AccountAuthenticatorMultiKey,
		Auth: &MultiKeyAuthenticator{
			PubKey: publicKey,
			Sig:    signature,
		},
	}
	authBytes, err := bcs.Serialize(auth)
	assert.NoError(t, err)
	authDeserialized := &AccountAuthenticator{}
	err = bcs.Deserialize(authDeserialized, authBytes)
	assert.NoError(t, err)
	assert.Equal(t, auth, authDeserialized)

}

func createMultiKey(t *testing.T) (
	*SingleSigner,
	*SingleSigner,
	*AnyPublicKey,
	*AnyPublicKey,
	*MultiKey,
) {
	// TODO: Add secp256k1 as well
	key1, err := GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	pubkey1 := ToAnyPublicKey(key1.PubKey())
	key2, err := GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	pubkey2 := ToAnyPublicKey(key2.PubKey())

	publicKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{pubkey1, pubkey2},
		SignaturesRequired: 2,
	}

	return &SingleSigner{key1}, &SingleSigner{key2}, ToAnyPublicKey(pubkey1), ToAnyPublicKey(pubkey2), publicKey
}

func createMultiKeySignature(t *testing.T, key1 *SingleSigner, key2 *SingleSigner, message []byte) *MultiKeySignature {
	sig1, err := key1.SignMessage(message)
	assert.NoError(t, err)
	sig2, err := key2.SignMessage(message)
	assert.NoError(t, err)

	// TODO: This signature should be built easier, ergonomics to fix this late
	return &MultiKeySignature{
		Signatures: []*AnySignature{
			sig1.(*AnySignature),
			sig2.(*AnySignature),
		},
		Bitmap: MultiKeyBitmap([]byte("c0000000")),
	}
}
