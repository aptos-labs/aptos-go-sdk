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
	assert.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
}

func TestNewEd25519Account(t *testing.T) {
	t.Parallel()
	account, err := NewEd25519Account()
	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
}

func TestNewEd25519SingleSenderAccount(t *testing.T) {
	t.Parallel()
	account, err := NewEd25519SingleSenderAccount()
	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
}

func TestNewSecp256k1Account(t *testing.T) {
	t.Parallel()
	account, err := NewSecp256k1Account()
	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
}
