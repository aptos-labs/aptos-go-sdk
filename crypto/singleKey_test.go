package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingleSigner_EmptySignature_Ed25519(t *testing.T) {
	t.Parallel()
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	emptySig := signer.EmptySignature()
	require.NotNil(t, emptySig)
	assert.Equal(t, AnySignatureVariantEd25519, emptySig.Variant)
	_, ok := emptySig.Signature.(*Ed25519Signature)
	assert.True(t, ok)
}

func TestSingleSigner_EmptySignature_Secp256k1(t *testing.T) {
	t.Parallel()
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	emptySig := signer.EmptySignature()
	require.NotNil(t, emptySig)
	assert.Equal(t, AnySignatureVariantSecp256k1, emptySig.Variant)
	_, ok := emptySig.Signature.(*Secp256k1Signature)
	assert.True(t, ok)
}

func TestSingleSigner_SimulationAuthenticator_Ed25519(t *testing.T) {
	t.Parallel()
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	auth := signer.SimulationAuthenticator()
	require.NotNil(t, auth)
	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)

	singleKeyAuth, ok := auth.Auth.(*SingleKeyAuthenticator)
	require.True(t, ok)
	assert.NotNil(t, singleKeyAuth.PubKey)
	assert.NotNil(t, singleKeyAuth.Sig)
}

func TestSingleSigner_SimulationAuthenticator_Secp256k1(t *testing.T) {
	t.Parallel()
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	auth := signer.SimulationAuthenticator()
	require.NotNil(t, auth)
	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)

	singleKeyAuth, ok := auth.Auth.(*SingleKeyAuthenticator)
	require.True(t, ok)
	assert.Equal(t, AnySignatureVariantSecp256k1, singleKeyAuth.Sig.Variant)
}

func TestSingleSigner_SignatureVariant(t *testing.T) {
	t.Parallel()

	t.Run("Ed25519", func(t *testing.T) {
		t.Parallel()
		key, err := GenerateEd25519PrivateKey()
		require.NoError(t, err)
		signer := NewSingleSigner(key)
		assert.Equal(t, AnySignatureVariantEd25519, signer.SignatureVariant())
	})

	t.Run("Secp256k1", func(t *testing.T) {
		t.Parallel()
		key, err := GenerateSecp256k1Key()
		require.NoError(t, err)
		signer := NewSingleSigner(key)
		assert.Equal(t, AnySignatureVariantSecp256k1, signer.SignatureVariant())
	})
}
