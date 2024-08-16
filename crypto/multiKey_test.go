package crypto

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMultiKey(t *testing.T) {
	key1, key2, key3, _, _, _, publicKey := createMultiKey(t)

	message := []byte("hello world")

	signature := createMultiKeySignature(t, 0, key1, 1, key2, message)

	// Test verification of signature
	assert.True(t, publicKey.Verify(message, signature))

	// Test serialization / deserialization authenticator
	auth := &MultiKeyAuthenticator{
		PubKey: publicKey,
		Sig:    signature,
	}
	assert.True(t, auth.Verify(message))

	signature = createMultiKeySignature(t, 2, key3, 1, key2, message)

	// Test verification of signature
	assert.True(t, publicKey.Verify(message, signature))

	// Test serialization / deserialization authenticator
	auth = &MultiKeyAuthenticator{
		PubKey: publicKey,
		Sig:    signature,
	}
	assert.True(t, auth.Verify(message))

	signature = createMultiKeySignature(t, 2, key3, 0, key1, message)

	// Test verification of signature
	assert.True(t, publicKey.Verify(message, signature))

	// Test serialization / deserialization authenticator
	auth = &MultiKeyAuthenticator{
		PubKey: publicKey,
		Sig:    signature,
	}
	assert.True(t, auth.Verify(message))
}

func TestMultiKeySerialization(t *testing.T) {
	key1, _, key3, _, _, _, publicKey := createMultiKey(t)

	// Test serialization / deserialization public key
	keyBytes, err := bcs.Serialize(publicKey)
	assert.NoError(t, err)
	publicKeyDeserialized := &MultiKey{}
	err = bcs.Deserialize(publicKeyDeserialized, keyBytes)
	assert.NoError(t, err)
	assert.Equal(t, publicKey, publicKeyDeserialized)

	// Test serialization / deserialization signature
	signature := createMultiKeySignature(t, 0, key1, 2, key3, []byte("test message"))
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
	*SingleSigner,
	*AnyPublicKey,
	*AnyPublicKey,
	*AnyPublicKey,
	*MultiKey,
) {
	key1, err := GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	pubkey1, err := ToAnyPublicKey(key1.PubKey())
	assert.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	pubkey2, err := ToAnyPublicKey(key2.PubKey())
	assert.NoError(t, err)
	key3, err := GenerateSecp256k1Key()
	assert.NoError(t, err)
	signer3 := NewSingleSigner(key3)
	pubkey3, err := ToAnyPublicKey(signer3.PubKey())
	assert.NoError(t, err)

	publicKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{pubkey1, pubkey2, pubkey3},
		SignaturesRequired: 2,
	}

	return &SingleSigner{key1}, &SingleSigner{key2}, &SingleSigner{key3}, pubkey1, pubkey2, pubkey3, publicKey
}

func createMultiKeySignature(t *testing.T, index1 uint8, key1 *SingleSigner, index2 uint8, key2 *SingleSigner, message []byte) *MultiKeySignature {
	sig1, err := key1.SignMessage(message)
	assert.NoError(t, err)
	sig2, err := key2.SignMessage(message)
	assert.NoError(t, err)

	bitmap := MultiKeyBitmap{}
	err = bitmap.AddKey(index1)
	assert.NoError(t, err)
	err = bitmap.AddKey(index2)
	assert.NoError(t, err)

	sig, err := NewMultiKeySignature([]IndexedAnySignature{
		{Index: index1, Signature: sig1.(*AnySignature)},
		{Index: index2, Signature: sig2.(*AnySignature)},
	})
	assert.NoError(t, err)
	return sig
}
