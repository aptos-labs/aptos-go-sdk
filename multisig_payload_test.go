package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultisigCreateAccountPayload(t *testing.T) {
	t.Parallel()
	additionalAddresses := []AccountAddress{AccountTwo, AccountThree}
	payload, err := MultisigCreateAccountPayload(2, additionalAddresses, []string{}, []byte{0})
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "create_with_owners", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
	assert.Len(t, payload.Args, 4)
}

func TestMultisigAddOwnerPayload(t *testing.T) {
	t.Parallel()
	payload := MultisigAddOwnerPayload(AccountTwo)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "add_owner", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
	assert.Len(t, payload.Args, 1)
	assert.Equal(t, AccountTwo[:], payload.Args[0])
}

func TestMultisigRemoveOwnerPayload(t *testing.T) {
	t.Parallel()
	payload := MultisigRemoveOwnerPayload(AccountTwo)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "remove_owner", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
	assert.Len(t, payload.Args, 1)
	assert.Equal(t, AccountTwo[:], payload.Args[0])
}

func TestMultisigChangeThresholdPayload(t *testing.T) {
	t.Parallel()
	payload, err := MultisigChangeThresholdPayload(3)
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "update_signatures_required", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
	assert.Len(t, payload.Args, 1)
}

func TestMultisigCreateTransactionPayload(t *testing.T) {
	t.Parallel()
	entryFunction, err := CoinTransferPayload(nil, AccountTwo, 1_000_000)
	require.NoError(t, err)

	multisigPayload := &MultisigTransactionPayload{
		Variant: MultisigTransactionPayloadVariantEntryFunction,
		Payload: entryFunction,
	}

	payload, err := MultisigCreateTransactionPayload(AccountTwo, multisigPayload)
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "create_transaction", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
}

func TestMultisigCreateTransactionPayloadWithHash(t *testing.T) {
	t.Parallel()
	entryFunction, err := CoinTransferPayload(nil, AccountTwo, 1_000_000)
	require.NoError(t, err)

	multisigPayload := &MultisigTransactionPayload{
		Variant: MultisigTransactionPayloadVariantEntryFunction,
		Payload: entryFunction,
	}

	payload, err := MultisigCreateTransactionPayloadWithHash(AccountTwo, multisigPayload)
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "create_transaction_with_hash", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
}

func TestMultisigApprovePayload(t *testing.T) {
	t.Parallel()
	payload, err := MultisigApprovePayload(AccountTwo, 1)
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "approve_transaction", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
	// First arg is the multisig address, second is the transaction ID
	assert.Len(t, payload.Args, 2)
	assert.Equal(t, AccountTwo[:], payload.Args[0])
}

func TestMultisigRejectPayload(t *testing.T) {
	t.Parallel()
	payload, err := MultisigRejectPayload(AccountTwo, 1)
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "multisig_account", payload.Module.Name)
	assert.Equal(t, "reject_transaction", payload.Function)
	assert.Empty(t, payload.ArgTypes)
	assert.NotNil(t, payload.Args)
	assert.Len(t, payload.Args, 2)
	assert.Equal(t, AccountTwo[:], payload.Args[0])
}
