package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
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

func TestNewSlhDsaSha2128sAccount(t *testing.T) {
	t.Parallel()
	account, err := NewSlhDsaSha2128sAccount()
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.NotEqual(t, AccountAddress{}, account.Address)
	// Address must be the SingleKey auth key of the SLH-DSA public key.
	assert.Equal(t, account.AuthKey().Bytes(), account.Address[:])

	msg := []byte("slh-dsa test")
	auth, err := account.Signer.Sign(msg)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong")))

	privStr, err := account.PrivateKeyString()
	require.NoError(t, err)
	assert.Greater(t, len(privStr), len("slh-dsa-sha2-128s-priv-0x"))
	assert.Equal(t, "slh-dsa-sha2-128s-priv-", privStr[:len("slh-dsa-sha2-128s-priv-")])
}

func TestNewSlhDsaSha2128sAccountFromMnemonic(t *testing.T) {
	t.Parallel()
	const mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	a1, err := NewSlhDsaSha2128sAccountFromMnemonic(mnemonic)
	require.NoError(t, err)
	a2, err := NewSlhDsaSha2128sAccountFromDerivationPath(mnemonic, crypto.DefaultDerivationPath)
	require.NoError(t, err)
	assert.Equal(t, a1.Address, a2.Address)

	_, err = NewSlhDsaSha2128sAccountFromDerivationPath(mnemonic, "m/44'/637'/0'/0'/0")
	require.Error(t, err, "non-hardened final segment must be rejected for SLH-DSA")
}

func TestSlhDsaSha2128sSignedTransaction(t *testing.T) {
	t.Parallel()
	sender, err := NewSlhDsaSha2128sAccount()
	require.NoError(t, err)
	receiver, err := NewEd25519Account()
	require.NoError(t, err)

	payload, err := CoinTransferPayload(nil, receiver.Address, 10_000)
	require.NoError(t, err)
	txn := RawTransaction{
		Sender:                     sender.Address,
		SequenceNumber:             1,
		Payload:                    TransactionPayload{Payload: payload},
		MaxGasAmount:               1000,
		GasUnitPrice:               2000,
		ExpirationTimestampSeconds: 1714158778,
		ChainId:                    4,
	}
	signedTxn, err := txn.SignedTransaction(sender)
	require.NoError(t, err)
	assert.Equal(t, TransactionAuthenticatorSingleSender, signedTxn.Authenticator.Variant)
	require.NoError(t, signedTxn.Verify())

	serialized, err := bcs.Serialize(signedTxn)
	require.NoError(t, err)
	deserialized := &SignedTransaction{}
	require.NoError(t, bcs.Deserialize(deserialized, serialized))
	require.NoError(t, deserialized.Verify())
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)
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
