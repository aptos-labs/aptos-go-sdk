package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountFromSigner(t *testing.T) {
	t.Parallel()
	key, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	account, err := NewAccountFromSigner(key)
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)

	// Verify the account can sign and the signature verifies
	msg := []byte("test message")
	auth, err := account.Signer.Sign(msg)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
}

func TestNewEd25519Account(t *testing.T) {
	t.Parallel()
	account, err := NewEd25519Account()
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
	assert.NotNil(t, account.Signer)

	// Verify sign + verify round-trip
	msg := []byte("ed25519 test")
	auth, err := account.Signer.Sign(msg)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
}

func TestNewEd25519SingleSenderAccount(t *testing.T) {
	t.Parallel()
	account, err := NewEd25519SingleSenderAccount()
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
	assert.NotNil(t, account.Signer)

	// Verify sign + verify round-trip
	msg := []byte("single sender test")
	auth, err := account.Signer.Sign(msg)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
}

func TestNewSecp256k1Account(t *testing.T) {
	t.Parallel()
	account, err := NewSecp256k1Account()
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
	assert.NotNil(t, account.Signer)

	// Verify sign + verify round-trip
	msg := []byte("secp256k1 test")
	auth, err := account.Signer.Sign(msg)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
}
