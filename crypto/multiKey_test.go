package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiKey(t *testing.T) {
	t.Parallel()
	key1, key2, key3, publicKey := createMultiKey(t)

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
	t.Parallel()
	key1, _, key3, publicKey := createMultiKey(t)

	// Test serialization / deserialization public key
	keyBytes, err := bcs.Serialize(publicKey)
	require.NoError(t, err)
	publicKeyDeserialized := &MultiKey{}
	err = bcs.Deserialize(publicKeyDeserialized, keyBytes)
	require.NoError(t, err)
	assert.Equal(t, publicKey, publicKeyDeserialized)

	// Test serialization / deserialization signature
	signature := createMultiKeySignature(t, 0, key1, 2, key3, []byte("test message"))
	sigBytes, err := bcs.Serialize(signature)
	require.NoError(t, err)
	signatureDeserialized := &MultiKeySignature{}
	err = bcs.Deserialize(signatureDeserialized, sigBytes)
	require.NoError(t, err)
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
	require.NoError(t, err)
	authDeserialized := &AccountAuthenticator{}
	err = bcs.Deserialize(authDeserialized, authBytes)
	require.NoError(t, err)
	assert.Equal(t, auth, authDeserialized)
}

func TestMultiKey_Serialization_CrossPlatform(t *testing.T) {
	t.Parallel()
	serialized := "020140118d6ebe543aaf3a541453f98a5748ab5b9e3f96d781b8c0a43740af2b65c03529fdf62b7de7aad9150770e0994dc4e0714795fdebf312be66cd0550c607755e00401a90421453aa53fa5a7aa3dfe70d913823cbf087bf372a762219ccc824d3a0eeecccaa9d34f22db4366aec61fb6c204d2440f4ed288bc7cc7e407b766723a60901c0"
	serializedBytes, err := hex.DecodeString(serialized)
	require.NoError(t, err)
	signature := &MultiKeySignature{}
	require.NoError(t, bcs.Deserialize(signature, serializedBytes))

	reserialized, err := bcs.Serialize(signature)
	require.NoError(t, err)
	assert.Equal(t, serializedBytes, reserialized)
}

func createMultiKey(t *testing.T) (
	*SingleSigner,
	*SingleSigner,
	*SingleSigner,
	*MultiKey,
) {
	t.Helper()
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	pubkey1, err := ToAnyPublicKey(key1.PubKey())
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	pubkey2, err := ToAnyPublicKey(key2.PubKey())
	require.NoError(t, err)
	key3, err := GenerateSecp256k1Key()
	require.NoError(t, err)
	signer3 := NewSingleSigner(key3)
	pubkey3, err := ToAnyPublicKey(signer3.PubKey())
	require.NoError(t, err)

	publicKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{pubkey1, pubkey2, pubkey3},
		SignaturesRequired: 2,
	}

	return &SingleSigner{key1}, &SingleSigner{key2}, &SingleSigner{key3}, publicKey
}

func createMultiKeySignature(t *testing.T, index1 uint8, key1 *SingleSigner, index2 uint8, key2 *SingleSigner, message []byte) *MultiKeySignature {
	t.Helper()
	sig1, err := key1.SignMessage(message)
	require.NoError(t, err)
	sig2, err := key2.SignMessage(message)
	require.NoError(t, err)

	bitmap := MultiKeyBitmap{}
	err = bitmap.AddKey(index1)
	require.NoError(t, err)
	err = bitmap.AddKey(index2)
	require.NoError(t, err)

	anySig1, ok := sig1.(*AnySignature)
	require.True(t, ok)
	anySig2, ok := sig2.(*AnySignature)
	require.True(t, ok)

	sig, err := NewMultiKeySignature([]IndexedAnySignature{
		{Index: index1, Signature: anySig1},
		{Index: index2, Signature: anySig2},
	})
	require.NoError(t, err)
	return sig
}
