package aptos

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperRawTransaction creates a RawTransaction with an EntryFunction payload for testing.
func helperRawTransaction(t *testing.T) *RawTransaction {
	t.Helper()

	destBytes, err := bcs.Serialize(&AccountTwo)
	require.NoError(t, err)

	amountBytes, err := bcs.SerializeU64(1000)
	require.NoError(t, err)

	payload := TransactionPayload{
		Payload: &EntryFunction{
			Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args:     [][]byte{destBytes, amountBytes},
		},
	}

	return &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		Payload:                    payload,
		MaxGasAmount:               2000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1714158778,
		ChainId:                    4,
	}
}

func TestRawTransaction_SigningMessage(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	message, err := rawTxn.SigningMessage()
	require.NoError(t, err)
	require.NotEmpty(t, message)

	// The signing message must start with the RawTransactionPrehash bytes.
	prehash := RawTransactionPrehash()
	require.Greater(t, len(message), len(prehash), "signing message must be longer than prehash")
	assert.Equal(t, prehash, message[:len(prehash)])
}

func TestRawTransaction_Sign(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	account, err := NewEd25519Account()
	require.NoError(t, err)

	auth, err := rawTxn.Sign(account.Signer)
	require.NoError(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, crypto.AccountAuthenticatorEd25519, auth.Variant)
}

func TestRawTransaction_SignedTransaction(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	account, err := NewEd25519Account()
	require.NoError(t, err)

	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)
	require.NotNil(t, signedTxn)
	assert.NotNil(t, signedTxn.Transaction)
	assert.NotNil(t, signedTxn.Authenticator)
}

func TestRawTransaction_String(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	str := rawTxn.String()
	require.NotEmpty(t, str)

	// The output should be valid JSON.
	assert.True(t, json.Valid([]byte(str)), "String() should return valid JSON")
}

func TestRawTransaction_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	serialized, err := bcs.Serialize(rawTxn)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &RawTransaction{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	// Verify all fields match.
	assert.Equal(t, rawTxn.Sender, deserialized.Sender)
	assert.Equal(t, rawTxn.SequenceNumber, deserialized.SequenceNumber)
	assert.Equal(t, rawTxn.MaxGasAmount, deserialized.MaxGasAmount)
	assert.Equal(t, rawTxn.GasUnitPrice, deserialized.GasUnitPrice)
	assert.Equal(t, rawTxn.ExpirationTimestampSeconds, deserialized.ExpirationTimestampSeconds)
	assert.Equal(t, rawTxn.ChainId, deserialized.ChainId)

	// Re-serialize and compare bytes for full round-trip verification.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)
}

func TestRawTransactionPrehash(t *testing.T) {
	t.Parallel()

	prehash := RawTransactionPrehash()
	require.Len(t, prehash, 32, "prehash must be 32 bytes (SHA3-256)")

	// Verify determinism: calling again should return the same value.
	prehash2 := RawTransactionPrehash()
	assert.Equal(t, prehash, prehash2)
}

func TestRawTransactionWithDataPrehash(t *testing.T) {
	t.Parallel()

	prehash := RawTransactionWithDataPrehash()
	require.Len(t, prehash, 32, "prehash must be 32 bytes (SHA3-256)")

	// Verify determinism: calling again should return the same value.
	prehash2 := RawTransactionWithDataPrehash()
	assert.Equal(t, prehash, prehash2)

	// RawTransactionWithDataPrehash and RawTransactionPrehash should differ.
	rawPrehash := RawTransactionPrehash()
	assert.NotEqual(t, prehash, rawPrehash, "the two prehashes should be different")
}

func TestRawTransactionWithData_SetFeePayer(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	t.Run("MultiAgentWithFeePayer variant succeeds", func(t *testing.T) {
		t.Parallel()

		txnWithData := &RawTransactionWithData{
			Variant: MultiAgentWithFeePayerRawTransactionWithDataVariant,
			Inner: &MultiAgentWithFeePayerRawTransactionWithData{
				RawTxn:           rawTxn,
				SecondarySigners: []AccountAddress{},
				FeePayer:         &AccountOne,
			},
		}

		newFeePayer := AccountTwo
		ok := txnWithData.SetFeePayer(newFeePayer)
		assert.True(t, ok, "SetFeePayer should succeed for MultiAgentWithFeePayer variant")

		inner, ok2 := txnWithData.Inner.(*MultiAgentWithFeePayerRawTransactionWithData)
		require.True(t, ok2)
		assert.Equal(t, newFeePayer, *inner.FeePayer)
	})

	t.Run("MultiAgent variant fails", func(t *testing.T) {
		t.Parallel()

		txnWithData := &RawTransactionWithData{
			Variant: MultiAgentRawTransactionWithDataVariant,
			Inner: &MultiAgentRawTransactionWithData{
				RawTxn:           rawTxn,
				SecondarySigners: []AccountAddress{},
			},
		}

		ok := txnWithData.SetFeePayer(AccountTwo)
		assert.False(t, ok, "SetFeePayer should fail for MultiAgent variant")
	})
}

func TestRawTransactionWithData_ToMultiAgentSignedTransaction(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	account, err := NewEd25519Account()
	require.NoError(t, err)

	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentRawTransactionWithDataVariant,
		Inner: &MultiAgentRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{},
		},
	}

	senderAuth, err := txnWithData.Sign(account.Signer)
	require.NoError(t, err)

	signedTxn, ok := txnWithData.ToMultiAgentSignedTransaction(senderAuth, []crypto.AccountAuthenticator{})
	require.True(t, ok, "ToMultiAgentSignedTransaction should succeed for MultiAgent variant")
	require.NotNil(t, signedTxn)
	assert.NotNil(t, signedTxn.Transaction)
	assert.NotNil(t, signedTxn.Authenticator)
	assert.Equal(t, TransactionAuthenticatorMultiAgent, signedTxn.Authenticator.Variant)
}

func TestRawTransactionWithData_ToFeePayerSignedTransaction(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	feePayer, err := NewEd25519Account()
	require.NoError(t, err)

	feePayerAddr := feePayer.Address
	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentWithFeePayerRawTransactionWithDataVariant,
		Inner: &MultiAgentWithFeePayerRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{},
			FeePayer:         &feePayerAddr,
		},
	}

	senderAuth, err := txnWithData.Sign(sender.Signer)
	require.NoError(t, err)

	feePayerAuth, err := txnWithData.Sign(feePayer.Signer)
	require.NoError(t, err)

	signedTxn, ok := txnWithData.ToFeePayerSignedTransaction(senderAuth, feePayerAuth, []crypto.AccountAuthenticator{})
	require.True(t, ok, "ToFeePayerSignedTransaction should succeed for MultiAgentWithFeePayer variant")
	require.NotNil(t, signedTxn)
	assert.NotNil(t, signedTxn.Transaction)
	assert.NotNil(t, signedTxn.Authenticator)
	assert.Equal(t, TransactionAuthenticatorFeePayer, signedTxn.Authenticator.Variant)
}

func TestRawTransactionWithData_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentRawTransactionWithDataVariant,
		Inner: &MultiAgentRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{AccountTwo, AccountThree},
		},
	}

	serialized, err := bcs.Serialize(txnWithData)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &RawTransactionWithData{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, txnWithData.Variant, deserialized.Variant)

	// Verify the inner data matches by comparing the raw transaction fields.
	originalInner, ok := txnWithData.Inner.(*MultiAgentRawTransactionWithData)
	require.True(t, ok)
	deserializedInner, ok := deserialized.Inner.(*MultiAgentRawTransactionWithData)
	require.True(t, ok)

	assert.Equal(t, originalInner.RawTxn.Sender, deserializedInner.RawTxn.Sender)
	assert.Equal(t, originalInner.RawTxn.SequenceNumber, deserializedInner.RawTxn.SequenceNumber)
	assert.Equal(t, originalInner.RawTxn.MaxGasAmount, deserializedInner.RawTxn.MaxGasAmount)
	assert.Equal(t, originalInner.RawTxn.ChainId, deserializedInner.RawTxn.ChainId)
	assert.Equal(t, originalInner.SecondarySigners, deserializedInner.SecondarySigners)

	// Full round-trip: re-serialize and compare bytes.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)
}

func TestRawTransactionWithData_String(t *testing.T) {
	t.Parallel()

	rawTxn := helperRawTransaction(t)

	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentRawTransactionWithDataVariant,
		Inner: &MultiAgentRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{},
		},
	}

	str := txnWithData.String()
	require.NotEmpty(t, str)

	// The output should be valid JSON.
	assert.True(t, json.Valid([]byte(str)), "String() should return valid JSON")

	// It should not be an error message.
	assert.False(t, strings.HasPrefix(str, "Error"), "String() should not return an error message")
}
