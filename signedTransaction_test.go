package aptos

import (
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignedTransaction_Verify(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	account, err := NewEd25519Account()
	require.NoError(t, err)

	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)
	require.NotNil(t, signedTxn)

	// A correctly signed transaction should verify without error.
	err = signedTxn.Verify()
	assert.NoError(t, err)
}

func TestSignedTransaction_Hash(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	account, err := NewEd25519Account()
	require.NoError(t, err)

	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)

	hash, err := signedTxn.Hash()
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// The hash should be a hex string starting with "0x".
	assert.True(t, strings.HasPrefix(hash, "0x"), "hash should start with 0x")

	// Calling Hash() again should return the same value (deterministic).
	hash2, err := signedTxn.Hash()
	require.NoError(t, err)
	assert.Equal(t, hash, hash2)
}

func TestSignedTransaction_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	account, err := NewEd25519Account()
	require.NoError(t, err)

	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)

	serialized, err := bcs.Serialize(signedTxn)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &SignedTransaction{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	// Verify the deserialized transaction fields match.
	assert.Equal(t, signedTxn.Transaction.Sender, deserialized.Transaction.Sender)
	assert.Equal(t, signedTxn.Transaction.SequenceNumber, deserialized.Transaction.SequenceNumber)
	assert.Equal(t, signedTxn.Transaction.MaxGasAmount, deserialized.Transaction.MaxGasAmount)
	assert.Equal(t, signedTxn.Transaction.GasUnitPrice, deserialized.Transaction.GasUnitPrice)
	assert.Equal(t, signedTxn.Transaction.ExpirationTimestampSeconds, deserialized.Transaction.ExpirationTimestampSeconds)
	assert.Equal(t, signedTxn.Transaction.ChainId, deserialized.Transaction.ChainId)
	assert.Equal(t, signedTxn.Authenticator.Variant, deserialized.Authenticator.Variant)

	// Full round-trip: re-serialize and compare bytes.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)

	// The deserialized transaction should also verify correctly.
	err = deserialized.Verify()
	assert.NoError(t, err)
}
