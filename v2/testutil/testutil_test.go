package testutil

import (
	"context"
	"testing"
	"time"

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

func TestSampleAccountInfo(t *testing.T) {
	info := SampleAccountInfo(42)
	assert.Equal(t, uint64(42), info.SequenceNumber)
	assert.NotEmpty(t, info.AuthenticationKey)
	assert.Contains(t, info.AuthenticationKey, "0x")
}

func TestSampleNodeInfo(t *testing.T) {
	info := SampleNodeInfo()
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, "full_node", info.NodeRole)
	assert.NotEmpty(t, info.GitHash)
	assert.Greater(t, info.LedgerVersion, uint64(0))
}

func TestSampleGasEstimate(t *testing.T) {
	estimate := SampleGasEstimate()
	assert.Equal(t, uint64(100), estimate.GasEstimate)
	assert.Equal(t, uint64(50), estimate.DeprioritizedGasEstimate)
	assert.Equal(t, uint64(150), estimate.PrioritizedGasEstimate)
}

func TestSampleResource(t *testing.T) {
	t.Run("with data", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		resource := SampleResource("0x1::test::Resource", data)
		assert.Equal(t, "0x1::test::Resource", resource.Type)
		assert.Equal(t, "value", resource.Data["key"])
	})

	t.Run("without data", func(t *testing.T) {
		resource := SampleResource("0x1::test::Resource", nil)
		assert.Equal(t, "0x1::test::Resource", resource.Type)
		assert.NotNil(t, resource.Data)
	})
}

func TestSampleEvent(t *testing.T) {
	event := SampleEvent("0x1::coin::DepositEvent", map[string]any{"amount": 100})
	assert.Equal(t, "0x1::coin::DepositEvent", event.Type)
	assert.NotEmpty(t, event.GUID.AccountAddress)
	assert.Equal(t, uint64(0), event.SequenceNumber)

	data, ok := event.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 100, data["amount"])
}

func TestAssertEventually(t *testing.T) {
	t.Run("returns true when condition passes", func(t *testing.T) {
		counter := 0
		result := AssertEventually(func() bool {
			counter++
			return counter >= 3
		}, 100*time.Millisecond, 10*time.Millisecond)
		assert.True(t, result)
	})

	t.Run("returns false on timeout", func(t *testing.T) {
		result := AssertEventually(func() bool {
			return false
		}, 50*time.Millisecond, 10*time.Millisecond)
		assert.False(t, result)
	})
}

func TestFakeClient_WithNodeInfo(t *testing.T) {
	info := SampleNodeInfo()
	info.ChainID = 42

	client := NewFakeClient().WithNodeInfo(info)
	ctx := context.Background()

	result, err := client.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint8(42), result.ChainID)
}

func TestFakeClient_ChainID(t *testing.T) {
	info := SampleNodeInfo()
	info.ChainID = 99

	client := NewFakeClient().WithNodeInfo(info)
	ctx := context.Background()

	chainID, err := client.ChainID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint8(99), chainID)
}

func TestFakeClient_WithGasEstimate(t *testing.T) {
	estimate := SampleGasEstimate()
	estimate.GasEstimate = 500

	client := NewFakeClient().WithGasEstimate(estimate)
	ctx := context.Background()

	result, err := client.EstimateGasPrice(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(500), result.GasEstimate)
}

func TestFakeClient_WithResources(t *testing.T) {
	addr := RandomAddress()
	resources := []aptos.Resource{
		SampleCoinStoreResource("0x1::aptos_coin::AptosCoin", 1000),
		SampleResource("0x1::test::TestResource", nil),
	}

	client := NewFakeClient().WithResources(addr, resources)
	ctx := context.Background()

	result, err := client.AccountResources(ctx, addr)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}
