package crypto

import (
	"crypto/ed25519"
	"testing"

	"github.com/qimeila/aptos-go-sdk/bcs"
	"github.com/qimeila/aptos-go-sdk/internal/util"
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
