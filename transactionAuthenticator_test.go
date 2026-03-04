package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperSignRawTxn creates an Ed25519 account, signs a raw transaction, and returns the account and authenticator.
func helperSignRawTxn(t *testing.T) (*Account, *crypto.AccountAuthenticator) {
	t.Helper()

	account, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = account.Address

	auth, err := rawTxn.Sign(account.Signer)
	require.NoError(t, err)

	return account, auth
}

func TestNewTransactionAuthenticator_Ed25519(t *testing.T) {
	t.Parallel()

	_, auth := helperSignRawTxn(t)
	assert.Equal(t, crypto.AccountAuthenticatorEd25519, auth.Variant)

	txnAuth, err := NewTransactionAuthenticator(auth)
	require.NoError(t, err)
	require.NotNil(t, txnAuth)
	assert.Equal(t, TransactionAuthenticatorEd25519, txnAuth.Variant)

	_, ok := txnAuth.Auth.(*Ed25519TransactionAuthenticator)
	assert.True(t, ok, "Auth should be *Ed25519TransactionAuthenticator")
}

func TestNewTransactionAuthenticator_SingleSender(t *testing.T) {
	t.Parallel()

	account, err := NewEd25519SingleSenderAccount()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = account.Address

	auth, err := rawTxn.Sign(account.Signer)
	require.NoError(t, err)
	assert.Equal(t, crypto.AccountAuthenticatorSingleSender, auth.Variant)

	txnAuth, err := NewTransactionAuthenticator(auth)
	require.NoError(t, err)
	require.NotNil(t, txnAuth)
	assert.Equal(t, TransactionAuthenticatorSingleSender, txnAuth.Variant)

	_, ok := txnAuth.Auth.(*SingleSenderTransactionAuthenticator)
	assert.True(t, ok, "Auth should be *SingleSenderTransactionAuthenticator")
}

func TestNewTransactionAuthenticator_MultiKey(t *testing.T) {
	t.Parallel()

	// A MultiKey variant should map to SingleSender in the TransactionAuthenticator.
	auth := &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiKey,
		Auth:    &crypto.MultiKeyAuthenticator{},
	}

	txnAuth, err := NewTransactionAuthenticator(auth)
	require.NoError(t, err)
	require.NotNil(t, txnAuth)
	assert.Equal(t, TransactionAuthenticatorSingleSender, txnAuth.Variant)

	_, ok := txnAuth.Auth.(*SingleSenderTransactionAuthenticator)
	assert.True(t, ok, "Auth should be *SingleSenderTransactionAuthenticator")
}

func TestNewTransactionAuthenticator_InvalidVariant(t *testing.T) {
	t.Parallel()

	auth := &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorType(255),
	}

	txnAuth, err := NewTransactionAuthenticator(auth)
	require.Error(t, err)
	assert.Nil(t, txnAuth)
	assert.Contains(t, err.Error(), "unknown authenticator type")
}

func TestTransactionAuthenticator_BCS_RoundTrip_Ed25519(t *testing.T) {
	t.Parallel()

	_, auth := helperSignRawTxn(t)

	txnAuth, err := NewTransactionAuthenticator(auth)
	require.NoError(t, err)

	serialized, err := bcs.Serialize(txnAuth)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &TransactionAuthenticator{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, txnAuth.Variant, deserialized.Variant)

	// For Ed25519TransactionAuthenticator, MarshalBCS serializes the inner Ed25519Authenticator
	// (not the full AccountAuthenticator), and UnmarshalBCS reconstructs it. Verify via re-serialization.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)

	// Verify the deserialized inner type.
	deserializedEd25519, ok := deserialized.Auth.(*Ed25519TransactionAuthenticator)
	require.True(t, ok)
	assert.Equal(t, crypto.AccountAuthenticatorEd25519, deserializedEd25519.Sender.Variant)
}

func TestTransactionAuthenticator_BCS_RoundTrip_SingleSender(t *testing.T) {
	t.Parallel()

	account, err := NewEd25519SingleSenderAccount()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = account.Address

	auth, err := rawTxn.Sign(account.Signer)
	require.NoError(t, err)

	txnAuth, err := NewTransactionAuthenticator(auth)
	require.NoError(t, err)

	serialized, err := bcs.Serialize(txnAuth)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &TransactionAuthenticator{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, txnAuth.Variant, deserialized.Variant)
	assert.Equal(t, TransactionAuthenticatorSingleSender, deserialized.Variant)

	// SingleSender round-trips cleanly through full AccountAuthenticator serialization.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)

	deserializedSS, ok := deserialized.Auth.(*SingleSenderTransactionAuthenticator)
	require.True(t, ok)
	assert.Equal(t, crypto.AccountAuthenticatorSingleSender, deserializedSS.Sender.Variant)
}

func TestMultiAgentTransactionAuthenticator_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	secondary, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = sender.Address

	// Build a multi-agent raw transaction with data.
	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentRawTransactionWithDataVariant,
		Inner: &MultiAgentRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{secondary.Address},
		},
	}

	senderAuth, err := txnWithData.Sign(sender.Signer)
	require.NoError(t, err)

	secondaryAuth, err := txnWithData.Sign(secondary.Signer)
	require.NoError(t, err)

	multiAgentAuth := &MultiAgentTransactionAuthenticator{
		Sender:                   senderAuth,
		SecondarySignerAddresses: []AccountAddress{secondary.Address},
		SecondarySigners:         []crypto.AccountAuthenticator{*secondaryAuth},
	}

	txnAuth := &TransactionAuthenticator{
		Variant: TransactionAuthenticatorMultiAgent,
		Auth:    multiAgentAuth,
	}

	serialized, err := bcs.Serialize(txnAuth)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &TransactionAuthenticator{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, TransactionAuthenticatorMultiAgent, deserialized.Variant)

	deserializedMA, ok := deserialized.Auth.(*MultiAgentTransactionAuthenticator)
	require.True(t, ok)
	assert.Len(t, deserializedMA.SecondarySignerAddresses, 1)
	assert.Equal(t, secondary.Address, deserializedMA.SecondarySignerAddresses[0])
	assert.Len(t, deserializedMA.SecondarySigners, 1)

	// Full round-trip byte comparison.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)
}

func TestMultiAgentTransactionAuthenticator_Verify(t *testing.T) {
	t.Parallel()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	secondary, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = sender.Address

	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentRawTransactionWithDataVariant,
		Inner: &MultiAgentRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{secondary.Address},
		},
	}

	msg, err := txnWithData.SigningMessage()
	require.NoError(t, err)

	senderAuth, err := sender.Signer.Sign(msg)
	require.NoError(t, err)

	secondaryAuth, err := secondary.Signer.Sign(msg)
	require.NoError(t, err)

	multiAgentAuth := &MultiAgentTransactionAuthenticator{
		Sender:                   senderAuth,
		SecondarySignerAddresses: []AccountAddress{secondary.Address},
		SecondarySigners:         []crypto.AccountAuthenticator{*secondaryAuth},
	}

	assert.True(t, multiAgentAuth.Verify(msg), "MultiAgent Verify should return true for valid signatures")

	// Wrap in TransactionAuthenticator and verify via the wrapper too.
	txnAuth := &TransactionAuthenticator{
		Variant: TransactionAuthenticatorMultiAgent,
		Auth:    multiAgentAuth,
	}
	assert.True(t, txnAuth.Verify(msg), "TransactionAuthenticator.Verify should delegate correctly")
}

func TestFeePayerTransactionAuthenticator_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	feePayer, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = sender.Address

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

	feePayerTxnAuth := &FeePayerTransactionAuthenticator{
		Sender:                   senderAuth,
		SecondarySignerAddresses: []AccountAddress{},
		SecondarySigners:         []crypto.AccountAuthenticator{},
		FeePayer:                 &feePayerAddr,
		FeePayerAuthenticator:    feePayerAuth,
	}

	txnAuth := &TransactionAuthenticator{
		Variant: TransactionAuthenticatorFeePayer,
		Auth:    feePayerTxnAuth,
	}

	serialized, err := bcs.Serialize(txnAuth)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &TransactionAuthenticator{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, TransactionAuthenticatorFeePayer, deserialized.Variant)

	deserializedFP, ok := deserialized.Auth.(*FeePayerTransactionAuthenticator)
	require.True(t, ok)
	assert.Equal(t, feePayerAddr, *deserializedFP.FeePayer)
	assert.Empty(t, deserializedFP.SecondarySignerAddresses)
	assert.Empty(t, deserializedFP.SecondarySigners)

	// Full round-trip byte comparison.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)
}

func TestFeePayerTransactionAuthenticator_Verify(t *testing.T) {
	t.Parallel()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	feePayer, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := helperRawTransaction(t)
	rawTxn.Sender = sender.Address

	feePayerAddr := feePayer.Address
	txnWithData := &RawTransactionWithData{
		Variant: MultiAgentWithFeePayerRawTransactionWithDataVariant,
		Inner: &MultiAgentWithFeePayerRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: []AccountAddress{},
			FeePayer:         &feePayerAddr,
		},
	}

	msg, err := txnWithData.SigningMessage()
	require.NoError(t, err)

	senderAuth, err := sender.Signer.Sign(msg)
	require.NoError(t, err)

	feePayerAuth, err := feePayer.Signer.Sign(msg)
	require.NoError(t, err)

	feePayerTxnAuth := &FeePayerTransactionAuthenticator{
		Sender:                   senderAuth,
		SecondarySignerAddresses: []AccountAddress{},
		SecondarySigners:         []crypto.AccountAuthenticator{},
		FeePayer:                 &feePayerAddr,
		FeePayerAuthenticator:    feePayerAuth,
	}

	assert.True(t, feePayerTxnAuth.Verify(msg), "FeePayer Verify should return true for valid signatures")

	// Wrap in TransactionAuthenticator and verify via the wrapper too.
	txnAuth := &TransactionAuthenticator{
		Variant: TransactionAuthenticatorFeePayer,
		Auth:    feePayerTxnAuth,
	}
	assert.True(t, txnAuth.Verify(msg), "TransactionAuthenticator.Verify should delegate correctly")
}
