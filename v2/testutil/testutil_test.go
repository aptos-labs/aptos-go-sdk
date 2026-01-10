package testutil

import (
	"context"
	"testing"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeClient_Account(t *testing.T) {
	addr := RandomAddress()
	client := NewFakeClient().
		WithAccount(addr, &aptos.AccountInfo{SequenceNumber: 42})

	ctx := context.Background()

	t.Run("returns configured account", func(t *testing.T) {
		info, err := client.Account(ctx, addr)
		require.NoError(t, err)
		assert.Equal(t, uint64(42), info.SequenceNumber)
	})

	t.Run("returns not found for unknown address", func(t *testing.T) {
		unknownAddr := RandomAddress()
		_, err := client.Account(ctx, unknownAddr)
		assert.ErrorIs(t, err, aptos.ErrNotFound)
	})
}

func TestFakeClient_Balance(t *testing.T) {
	addr := RandomAddress()
	client := NewFakeClient().
		WithBalance(addr, 100_000_000)

	ctx := context.Background()

	balance, err := client.AccountBalance(ctx, addr)
	require.NoError(t, err)
	assert.Equal(t, uint64(100_000_000), balance)
}

func TestFakeClient_Fund(t *testing.T) {
	addr := RandomAddress()
	client := NewFakeClient()

	ctx := context.Background()

	// Fund the account
	err := client.Fund(ctx, addr, 50_000_000)
	require.NoError(t, err)

	// Check balance increased
	balance, err := client.AccountBalance(ctx, addr)
	require.NoError(t, err)
	assert.Equal(t, uint64(50_000_000), balance)

	// Account should now exist
	_, err = client.Account(ctx, addr)
	require.NoError(t, err)
}

func TestFakeClient_ErrorSimulation(t *testing.T) {
	client := NewFakeClient().
		WithError("Account", aptos.ErrNotFound)

	ctx := context.Background()
	addr := RandomAddress()

	_, err := client.Account(ctx, addr)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_Recording(t *testing.T) {
	addr := RandomAddress()
	client := NewFakeClient().
		WithRecording().
		WithAccount(addr, &aptos.AccountInfo{SequenceNumber: 1})

	ctx := context.Background()

	// Make some calls
	_, _ = client.Account(ctx, addr)
	_, _ = client.AccountBalance(ctx, addr)
	_, _ = client.Info(ctx)

	// Check recorded calls
	calls := client.RecordedCalls()
	require.Len(t, calls, 3)
	assert.Equal(t, "Account", calls[0].Method)
	assert.Equal(t, "AccountBalance", calls[1].Method)
	assert.Equal(t, "Info", calls[2].Method)

	// Clear and verify
	client.ClearRecordedCalls()
	assert.Empty(t, client.RecordedCalls())
}

func TestFakeClient_BuildTransaction(t *testing.T) {
	sender := RandomAddress()
	client := NewFakeClient().
		WithAccount(sender, &aptos.AccountInfo{SequenceNumber: 10})

	ctx := context.Background()

	payload := &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
		Function: "transfer",
	}

	txn, err := client.BuildTransaction(ctx, sender, payload)
	require.NoError(t, err)
	assert.Equal(t, sender, txn.Sender)
	assert.Equal(t, uint64(10), txn.SequenceNumber)
	assert.Equal(t, uint8(4), txn.ChainID)
}

func TestFakeClient_Transactions(t *testing.T) {
	txn1 := SampleTransaction()
	txn2 := SampleTransaction()
	txn2.Version = 1001

	client := NewFakeClient().
		WithTransaction(txn1).
		WithTransaction(txn2)

	ctx := context.Background()

	t.Run("get by hash", func(t *testing.T) {
		result, err := client.Transaction(ctx, txn1.Hash)
		require.NoError(t, err)
		assert.Equal(t, txn1.Hash, result.Hash)
	})

	t.Run("get by version", func(t *testing.T) {
		result, err := client.TransactionByVersion(ctx, txn2.Version)
		require.NoError(t, err)
		assert.Equal(t, txn2.Version, result.Version)
	})
}

func TestFakeClient_Blocks(t *testing.T) {
	block := SampleBlock(100, 5)
	client := NewFakeClient().WithBlock(block)

	ctx := context.Background()

	t.Run("get by height", func(t *testing.T) {
		result, err := client.BlockByHeight(ctx, 100, true)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), result.BlockHeight)
		assert.Len(t, result.Transactions, 5)
	})

	t.Run("get by version", func(t *testing.T) {
		result, err := client.BlockByVersion(ctx, block.FirstVersion+2, true)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), result.BlockHeight)
	})
}

func TestRandomAddress(t *testing.T) {
	addr1 := RandomAddress()
	addr2 := RandomAddress()

	// Should generate different addresses
	assert.NotEqual(t, addr1, addr2)

	// Should be 32 bytes
	assert.Len(t, addr1[:], 32)
}

func TestRandomSigner(t *testing.T) {
	signer1 := RandomSigner()
	signer2 := RandomSigner()

	// Should generate different signers
	assert.NotEqual(t, signer1.Bytes(), signer2.Bytes())
}

func TestSampleTransaction(t *testing.T) {
	txn := SampleTransaction()
	assert.True(t, txn.Success)
	assert.NotEmpty(t, txn.Hash)
	assert.Equal(t, "user_transaction", txn.Type)
}

func TestSampleFailedTransaction(t *testing.T) {
	txn := SampleFailedTransaction()
	assert.False(t, txn.Success)
	assert.Contains(t, txn.VMStatus, "abort")
}

func TestSampleBlock(t *testing.T) {
	block := SampleBlock(50, 3)
	assert.Equal(t, uint64(50), block.BlockHeight)
	assert.Len(t, block.Transactions, 3)
	assert.Equal(t, block.FirstVersion, block.Transactions[0].Version)
}

func TestSampleCoinStoreResource(t *testing.T) {
	resource := SampleCoinStoreResource("0x1::aptos_coin::AptosCoin", 1_000_000)
	assert.Contains(t, resource.Type, "CoinStore")

	coin, ok := resource.Data["coin"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "1000000", coin["value"])
}

func TestPredefinedSigners(t *testing.T) {
	// Should have 5 predefined signers
	assert.Len(t, PredefinedSigners, 5)

	// All should be different
	for i := 0; i < len(PredefinedSigners); i++ {
		for j := i + 1; j < len(PredefinedSigners); j++ {
			assert.NotEqual(t, PredefinedSigners[i].Bytes(), PredefinedSigners[j].Bytes())
		}
	}
}

func TestWellKnownAddresses(t *testing.T) {
	assert.Equal(t, aptos.AccountZero, WellKnownAddresses.Zero)
	assert.Equal(t, aptos.AccountOne, WellKnownAddresses.One)
	assert.Equal(t, aptos.AccountTwo, WellKnownAddresses.Two)
	assert.Equal(t, aptos.AccountThree, WellKnownAddresses.Three)
}
