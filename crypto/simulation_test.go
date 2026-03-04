package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoAuthenticator_PublicKey(t *testing.T) {
	t.Parallel()
	auth := &NoAuthenticator{}
	assert.Nil(t, auth.PublicKey())
}

func TestNoAuthenticator_Signature(t *testing.T) {
	t.Parallel()
	auth := &NoAuthenticator{}
	assert.Nil(t, auth.Signature())
}

func TestNoAuthenticator_Verify(t *testing.T) {
	t.Parallel()
	auth := &NoAuthenticator{}
	assert.False(t, auth.Verify([]byte("hello")))
	assert.False(t, auth.Verify(nil))
}

func TestNoAuthenticator_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	auth := &NoAuthenticator{}

	// Serialize
	serializer := &bcs.Serializer{}
	auth.MarshalBCS(serializer)
	require.NoError(t, serializer.Error())

	// Deserialize
	data := serializer.ToBytes()
	deserializer := bcs.NewDeserializer(data)
	auth2 := &NoAuthenticator{}
	auth2.UnmarshalBCS(deserializer)
	require.NoError(t, deserializer.Error())
}

func TestNoAccountAuthenticator_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	auth := NoAccountAuthenticator()

	// Serialize the full AccountAuthenticator (goes through MarshalBCS on both levels)
	serialized, err := bcs.Serialize(auth)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	// Deserialize
	auth2 := &AccountAuthenticator{}
	err = bcs.Deserialize(auth2, serialized)
	require.NoError(t, err)
	assert.Equal(t, AccountAuthenticatorNone, auth2.Variant)
}

func TestNoAccountAuthenticator(t *testing.T) {
	t.Parallel()
	auth := NoAccountAuthenticator()
	assert.NotNil(t, auth)
	assert.Equal(t, AccountAuthenticatorNone, auth.Variant)
	_, ok := auth.Auth.(*NoAuthenticator)
	assert.True(t, ok)

	// Verify the authenticator works as expected
	assert.Nil(t, auth.Auth.PublicKey())
	assert.Nil(t, auth.Auth.Signature())
	assert.False(t, auth.Verify([]byte("test")))
}
