package crypto

import (
	"crypto/ed25519"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationKey_FromPublicKey(t *testing.T) {
	t.Parallel()
	// Ed25519
	privateKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	publicKey := privateKey.PubKey()

	authKey := AuthenticationKey{}
	authKey.FromPublicKey(publicKey)

	hash := util.Sha3256Hash([][]byte{
		publicKey.Bytes(),
		{Ed25519Scheme},
	})

	assert.Equal(t, hash, authKey[:])
}

func Test_AuthenticationKeySerialization(t *testing.T) {
	t.Parallel()
	bytesWithLength := []byte{
		32,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
	}
	bytes := []byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
	}
	authKey := AuthenticationKey(bytes)
	serialized, err := bcs.Serialize(&authKey)
	require.NoError(t, err)
	assert.Equal(t, bytesWithLength, serialized)

	newAuthKey := AuthenticationKey{}
	err = bcs.Deserialize(&newAuthKey, serialized)
	require.NoError(t, err)
	assert.Equal(t, authKey, newAuthKey)
}

func Test_AuthenticatorSerialization(t *testing.T) {
	t.Parallel()
	msg := []byte{0x01, 0x02}
	privateKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	authenticator, err := privateKey.Sign(msg)
	require.NoError(t, err)

	serialized, err := bcs.Serialize(authenticator)
	require.NoError(t, err)
	assert.Equal(t, uint8(AccountAuthenticatorEd25519), serialized[0])
	assert.Len(t, serialized, 1+(1+ed25519.PublicKeySize)+(1+ed25519.SignatureSize))

	newAuthenticator := &AccountAuthenticator{}
	err = bcs.Deserialize(newAuthenticator, serialized)
	require.NoError(t, err)
	assert.Equal(t, authenticator.Variant, newAuthenticator.Variant)
	assert.Equal(t, authenticator.Auth, newAuthenticator.Auth)
}

func Test_AuthenticatorVerification(t *testing.T) {
	t.Parallel()
	msg := []byte{0x01, 0x02}
	privateKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	authenticator, err := privateKey.Sign(msg)
	require.NoError(t, err)

	assert.True(t, authenticator.Verify(msg))
}

func TestAccountAuthenticator_Ed25519_RoundTrip(t *testing.T) {
	t.Parallel()
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("authenticator round trip")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	assert.Equal(t, AccountAuthenticatorEd25519, auth.Variant)
	assert.True(t, auth.Verify(msg))

	// BCS round trip
	authBytes, err := bcs.Serialize(auth)
	require.NoError(t, err)
	auth2 := &AccountAuthenticator{}
	err = bcs.Deserialize(auth2, authBytes)
	require.NoError(t, err)
	assert.Equal(t, auth.Variant, auth2.Variant)
	assert.True(t, auth2.Verify(msg))
}

func TestAccountAuthenticator_SingleKey_RoundTrip(t *testing.T) {
	t.Parallel()
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	signer := NewSingleSigner(key)

	msg := []byte("single key round trip")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)

	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)
	assert.True(t, auth.Verify(msg))

	// BCS round trip
	authBytes, err := bcs.Serialize(auth)
	require.NoError(t, err)
	auth2 := &AccountAuthenticator{}
	err = bcs.Deserialize(auth2, authBytes)
	require.NoError(t, err)
	assert.Equal(t, auth.Variant, auth2.Variant)
	assert.True(t, auth2.Verify(msg))
}

func Test_InvalidAuthenticatorDeserialization(t *testing.T) {
	t.Parallel()
	serialized := []byte{0xFF}
	newAuthenticator := &AccountAuthenticator{}
	err := bcs.Deserialize(newAuthenticator, serialized)
	require.Error(t, err)
	serialized = []byte{0x4F}
	newAuthenticator = &AccountAuthenticator{}
	err = bcs.Deserialize(newAuthenticator, serialized)
	require.Error(t, err)
}

func Test_InvalidAuthenticationKeyDeserialization(t *testing.T) {
	t.Parallel()
	serialized := []byte{0xFF}
	newAuthkey := AuthenticationKey{}
	err := bcs.Deserialize(&newAuthkey, serialized)
	require.Error(t, err)
}

func TestAccountAuthenticator_FromKeyAndSignature(t *testing.T) {
	t.Parallel()

	t.Run("Ed25519", func(t *testing.T) {
		t.Parallel()
		key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)

		msg := []byte("test FromKeyAndSignature ed25519")
		auth, err := key.Sign(msg)
		require.NoError(t, err)

		ed25519Auth, ok := auth.Auth.(*Ed25519Authenticator)
		require.True(t, ok)
		pubKey := ed25519Auth.PubKey
		sig := ed25519Auth.Sig

		// Reconstruct from key and signature
		auth2 := &AccountAuthenticator{}
		err = auth2.FromKeyAndSignature(pubKey, sig)
		require.NoError(t, err)
		assert.Equal(t, AccountAuthenticatorEd25519, auth2.Variant)
		assert.True(t, auth2.Verify(msg))
	})

	t.Run("SingleKey_AnyPublicKey_AnySignature", func(t *testing.T) {
		t.Parallel()
		key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)
		signer := NewSingleSigner(key)

		msg := []byte("test FromKeyAndSignature single key")
		auth, err := signer.Sign(msg)
		require.NoError(t, err)

		singleAuth, ok := auth.Auth.(*SingleKeyAuthenticator)
		require.True(t, ok)
		anyPubKey := singleAuth.PubKey
		anySig := singleAuth.Sig

		// Reconstruct from AnyPublicKey and AnySignature
		auth2 := &AccountAuthenticator{}
		err = auth2.FromKeyAndSignature(anyPubKey, anySig)
		require.NoError(t, err)
		assert.Equal(t, AccountAuthenticatorSingleSender, auth2.Variant)
		assert.True(t, auth2.Verify(msg))
	})

	t.Run("InvalidSignatureTypeForEd25519", func(t *testing.T) {
		t.Parallel()
		key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)

		pubKey, ok := key.PubKey().(*Ed25519PublicKey)
		require.True(t, ok)

		// Use a Secp256k1 signature with an Ed25519 key - should fail
		secpKey, err := GenerateSecp256k1Key()
		require.NoError(t, err)
		secpSig, err := secpKey.SignMessage([]byte("wrong type"))
		require.NoError(t, err)

		auth := &AccountAuthenticator{}
		err = auth.FromKeyAndSignature(pubKey, secpSig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature type for Ed25519PublicKey")
	})

	t.Run("InvalidSignatureTypeForAnyPublicKey", func(t *testing.T) {
		t.Parallel()
		key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)
		signer := NewSingleSigner(key)

		msg := []byte("test invalid sig type for AnyPublicKey")
		auth, err := signer.Sign(msg)
		require.NoError(t, err)

		singleAuth, ok := auth.Auth.(*SingleKeyAuthenticator)
		require.True(t, ok)
		anyPubKey := singleAuth.PubKey

		// Pass an Ed25519Signature instead of AnySignature - should fail
		ed25519Key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)
		ed25519Auth, err := ed25519Key.Sign(msg)
		require.NoError(t, err)
		ed25519AuthImpl, ok := ed25519Auth.Auth.(*Ed25519Authenticator)
		require.True(t, ok)
		ed25519Sig := ed25519AuthImpl.Sig

		auth2 := &AccountAuthenticator{}
		err = auth2.FromKeyAndSignature(anyPubKey, ed25519Sig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature type for AnyPublicKey")
	})

	t.Run("InvalidSignatureTypeForMultiEd25519", func(t *testing.T) {
		t.Parallel()
		// Use a MultiEd25519PublicKey with a wrong signature type
		ed25519Key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)
		ed25519PubKey, ok := ed25519Key.PubKey().(*Ed25519PublicKey)
		require.True(t, ok)
		multiEd25519PubKey := &MultiEd25519PublicKey{
			PubKeys:            []*Ed25519PublicKey{ed25519PubKey},
			SignaturesRequired: 1,
		}

		// Pass an Ed25519Signature (not MultiEd25519Signature) - should fail
		ed25519Auth, err := ed25519Key.Sign([]byte("test"))
		require.NoError(t, err)
		ed25519AuthImpl, ok := ed25519Auth.Auth.(*Ed25519Authenticator)
		require.True(t, ok)
		ed25519Sig := ed25519AuthImpl.Sig

		auth := &AccountAuthenticator{}
		err = auth.FromKeyAndSignature(multiEd25519PubKey, ed25519Sig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature type for MultiEd25519PublicKey")
	})
}
