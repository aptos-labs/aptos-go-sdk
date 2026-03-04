package testutil

import (
	"context"
	"testing"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
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

func TestFakeClient_SimulateTransaction(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	txn := &aptos.RawTransaction{
		Sender:   aptos.AccountOne,
		ChainID:  4,
		Payload:  &aptos.EntryFunctionPayload{Module: aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"}, Function: "transfer"},
	}

	signer := RandomSigner()
	result, err := client.SimulateTransaction(ctx, txn, signer)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.VMStatus)
}

func TestFakeClient_SimulateTransaction_Error(t *testing.T) {
	client := NewFakeClient().WithError("SimulateTransaction", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.SimulateTransaction(ctx, &aptos.RawTransaction{}, RandomSigner())
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_HealthCheck(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	resp, err := client.HealthCheck(ctx)
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "ok")
}

func TestFakeClient_HealthCheck_Error(t *testing.T) {
	client := NewFakeClient().WithError("HealthCheck", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.HealthCheck(ctx)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_AccountModule(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	mod, err := client.AccountModule(ctx, aptos.AccountOne, "coin")
	require.NoError(t, err)
	assert.Equal(t, "0x", mod.Bytecode)
	assert.Equal(t, "coin", mod.ABI.Name)
}

func TestFakeClient_AccountModule_Error(t *testing.T) {
	client := NewFakeClient().WithError("AccountModule", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.AccountModule(ctx, aptos.AccountOne, "coin")
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_AccountTransactions(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	txns, err := client.AccountTransactions(ctx, aptos.AccountOne, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, txns)
}

func TestFakeClient_AccountTransactions_Error(t *testing.T) {
	client := NewFakeClient().WithError("AccountTransactions", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.AccountTransactions(ctx, aptos.AccountOne, nil, nil)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_BatchSubmitTransaction(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	result, err := client.BatchSubmitTransaction(ctx, nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFakeClient_BatchSubmitTransaction_Error(t *testing.T) {
	client := NewFakeClient().WithError("BatchSubmitTransaction", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.BatchSubmitTransaction(ctx, nil)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_EventsByHandle(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	events, err := client.EventsByHandle(ctx, aptos.AccountOne, "0x1::coin::CoinStore", "deposit_events", nil, nil)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestFakeClient_EventsByHandle_Error(t *testing.T) {
	client := NewFakeClient().WithError("EventsByHandle", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.EventsByHandle(ctx, aptos.AccountOne, "handle", "field", nil, nil)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_EventsByCreationNumber(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	events, err := client.EventsByCreationNumber(ctx, aptos.AccountOne, 0, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestFakeClient_EventsByCreationNumber_Error(t *testing.T) {
	client := NewFakeClient().WithError("EventsByCreationNumber", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.EventsByCreationNumber(ctx, aptos.AccountOne, 0, nil, nil)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_AccountResource(t *testing.T) {
	addr := RandomAddress()
	resources := []aptos.Resource{
		SampleCoinStoreResource("0x1::aptos_coin::AptosCoin", 1000),
		SampleResource("0x1::test::TestResource", nil),
	}
	client := NewFakeClient().WithResources(addr, resources)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		r, err := client.AccountResource(ctx, addr, "0x1::test::TestResource")
		require.NoError(t, err)
		assert.Equal(t, "0x1::test::TestResource", r.Type)
	})

	t.Run("not found type", func(t *testing.T) {
		_, err := client.AccountResource(ctx, addr, "0x1::missing::Resource")
		assert.ErrorIs(t, err, aptos.ErrNotFound)
	})

	t.Run("not found address", func(t *testing.T) {
		_, err := client.AccountResource(ctx, RandomAddress(), "0x1::test::TestResource")
		assert.ErrorIs(t, err, aptos.ErrNotFound)
	})
}

func TestFakeClient_AccountResource_Error(t *testing.T) {
	client := NewFakeClient().WithError("AccountResource", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.AccountResource(ctx, aptos.AccountOne, "type")
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_SubmitTransaction(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	result, err := client.SubmitTransaction(ctx, &aptos.SignedTransaction{})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Hash)
}

func TestFakeClient_SubmitTransaction_Error(t *testing.T) {
	client := NewFakeClient().WithError("SubmitTransaction", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.SubmitTransaction(ctx, &aptos.SignedTransaction{})
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_SignAndSubmitTransaction(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	acct, err := account.NewEd25519()
	require.NoError(t, err)
	payload := &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
		Function: "transfer",
	}

	result, err := client.SignAndSubmitTransaction(ctx, acct, payload)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Hash)
}

func TestFakeClient_SignAndSubmitTransaction_Error(t *testing.T) {
	client := NewFakeClient().WithError("SignAndSubmitTransaction", aptos.ErrNotFound)
	ctx := context.Background()

	acct, err := account.NewEd25519()
	require.NoError(t, err)

	_, err = client.SignAndSubmitTransaction(ctx, acct, &aptos.EntryFunctionPayload{})
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_WaitForTransaction(t *testing.T) {
	txn := SampleTransaction()
	client := NewFakeClient().WithTransaction(txn)
	ctx := context.Background()

	t.Run("known hash", func(t *testing.T) {
		result, err := client.WaitForTransaction(ctx, txn.Hash)
		require.NoError(t, err)
		assert.Equal(t, txn.Hash, result.Hash)
	})

	t.Run("unknown hash returns default success", func(t *testing.T) {
		result, err := client.WaitForTransaction(ctx, "0xunknown")
		require.NoError(t, err)
		assert.True(t, result.Success)
	})
}

func TestFakeClient_WaitForTransaction_Error(t *testing.T) {
	client := NewFakeClient().WithError("WaitForTransaction", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.WaitForTransaction(ctx, "0xhash")
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_TransactionsList(t *testing.T) {
	txn1 := SampleTransaction()
	txn2 := SampleTransaction()
	txn2.Hash = "0xdifferenthash"
	txn2.Version = 2000
	client := NewFakeClient().WithTransaction(txn1).WithTransaction(txn2)
	ctx := context.Background()

	result, err := client.Transactions(ctx, nil, nil)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestFakeClient_TransactionsList_Error(t *testing.T) {
	client := NewFakeClient().WithError("Transactions", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.Transactions(ctx, nil, nil)
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}

func TestFakeClient_TransactionsIter(t *testing.T) {
	txn := SampleTransaction()
	client := NewFakeClient().WithTransaction(txn)
	ctx := context.Background()

	count := 0
	for _, err := range client.TransactionsIter(ctx, nil) {
		require.NoError(t, err)
		count++
	}
	assert.Equal(t, 1, count)
}

func TestFakeClient_TransactionsIter_Error(t *testing.T) {
	client := NewFakeClient().WithError("TransactionsIter", aptos.ErrNotFound)
	ctx := context.Background()

	for _, err := range client.TransactionsIter(ctx, nil) {
		assert.ErrorIs(t, err, aptos.ErrNotFound)
	}
}

func TestFakeClient_View(t *testing.T) {
	client := NewFakeClient()
	ctx := context.Background()

	result, err := client.View(ctx, &aptos.ViewPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
		Function: "balance",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFakeClient_View_Error(t *testing.T) {
	client := NewFakeClient().WithError("View", aptos.ErrNotFound)
	ctx := context.Background()

	_, err := client.View(ctx, &aptos.ViewPayload{})
	assert.ErrorIs(t, err, aptos.ErrNotFound)
}
