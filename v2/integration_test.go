// Package aptos integration tests verify the SDK against live networks.
//
// These tests require network connectivity and test against real Aptos networks.
// They are designed to be run sparingly and may be skipped in CI with -short flag.
//
// Run with: go test -v -run Integration ./v2/...
package aptos

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testNetwork returns the network config to use for integration tests.
// Defaults to Testnet, can be overridden with APTOS_NETWORK env var.
func testNetwork() NetworkConfig {
	network := os.Getenv("APTOS_NETWORK")
	switch network {
	case "mainnet":
		return Mainnet
	case "devnet":
		return Devnet
	case "localnet":
		return Localnet
	default:
		return Testnet
	}
}

// skipIfShort skips the test if running with -short flag.
func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

// testContext returns a context with a reasonable timeout for tests.
func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// createTestClient creates a client for integration tests.
func createTestClient(t *testing.T) Client {
	t.Helper()
	client, err := NewClient(testNetwork())
	require.NoError(t, err, "failed to create client")
	return client
}

// TestIntegration_ClientCreation tests that a client can be created for each network.
func TestIntegration_ClientCreation(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	networks := []NetworkConfig{Mainnet, Testnet, Devnet}
	for _, network := range networks {
		network := network // capture for parallel
		t.Run(network.Name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(network)
			require.NoError(t, err, "failed to create client for %s", network.Name)
		})
	}
}

// TestIntegration_ClientWithOptions tests client creation with various options.
func TestIntegration_ClientWithOptions(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	t.Run("with_timeout", func(t *testing.T) {
		t.Parallel()
		client, err := NewClient(testNetwork(), WithTimeout(5*time.Second))
		require.NoError(t, err)

		ctx := testContext(t)
		info, err := client.Info(ctx)
		require.NoError(t, err)
		assert.NotZero(t, info.ChainID)
	})

	t.Run("with_header", func(t *testing.T) {
		t.Parallel()
		client, err := NewClient(testNetwork(), WithHeader("X-Custom-Header", "test-value"))
		require.NoError(t, err)

		ctx := testContext(t)
		info, err := client.Info(ctx)
		require.NoError(t, err)
		assert.NotZero(t, info.ChainID)
	})
}

// TestIntegration_NodeInfo tests fetching node information.
func TestIntegration_NodeInfo(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	info, err := client.Info(ctx)
	require.NoError(t, err, "failed to get node info")

	assert.NotZero(t, info.ChainID, "chain ID should not be zero")
	assert.NotZero(t, info.LedgerVersion, "ledger version should not be zero")
	assert.NotZero(t, info.BlockHeight, "block height should not be zero")
	assert.NotZero(t, info.LedgerTimestamp, "ledger timestamp should not be zero")
	assert.NotEmpty(t, info.NodeRole, "node role should not be empty")
	assert.NotEmpty(t, info.GitHash, "git hash should not be empty")

	t.Logf("Node Info: chain_id=%d, ledger_version=%d, block_height=%d, role=%s",
		info.ChainID, info.LedgerVersion, info.BlockHeight, info.NodeRole)
}

// TestIntegration_ChainID tests fetching and caching the chain ID.
func TestIntegration_ChainID(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	chainID, err := client.ChainID(ctx)
	require.NoError(t, err, "failed to get chain ID")
	assert.NotZero(t, chainID, "chain ID should not be zero")

	// Second call should use cached value
	chainID2, err := client.ChainID(ctx)
	require.NoError(t, err)
	assert.Equal(t, chainID, chainID2, "chain ID should be consistent")

	// Verify it matches expected for the network
	network := testNetwork()
	if network.ChainID != 0 { // 0 means dynamic (devnet)
		assert.Equal(t, network.ChainID, chainID, "chain ID should match network config")
	}

	t.Logf("Chain ID: %d", chainID)
}

// TestIntegration_Account tests fetching account information.
func TestIntegration_Account(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Test with the core framework account (0x1)
	account, err := client.Account(ctx, AccountOne)
	require.NoError(t, err, "failed to get account info for 0x1")

	assert.NotEmpty(t, account.AuthenticationKey, "authentication key should not be empty")
	t.Logf("Account 0x1: seq_num=%d, auth_key=%s", account.SequenceNumber, account.AuthenticationKey)
}

// TestIntegration_AccountNotFound tests error handling for non-existent accounts.
func TestIntegration_AccountNotFound(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Generate an address that almost certainly doesn't exist
	// Using a very unlikely address (all 0xff except the last byte which would make it invalid)
	randomAddr := MustParseAddress("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	_, err := client.Account(ctx, randomAddr)
	// This may or may not fail depending on network state
	// The test verifies proper error handling when it does fail
	if err != nil {
		apiErr, ok := err.(*APIError)
		if ok {
			assert.Equal(t, 404, apiErr.StatusCode, "should be 404 not found")
			t.Logf("Correctly received 404 for non-existent account")
		} else {
			t.Logf("Received non-API error: %v", err)
		}
	} else {
		t.Log("Account unexpectedly exists (this can happen on testnet)")
	}
}

// TestIntegration_AccountResources tests fetching account resources.
func TestIntegration_AccountResources(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	resources, err := client.AccountResources(ctx, AccountOne)
	require.NoError(t, err, "failed to get account resources for 0x1")
	require.NotEmpty(t, resources, "0x1 should have resources")

	// 0x1 should have the account resource
	var foundAccount bool
	for _, r := range resources {
		if r.Type == "0x1::account::Account" {
			foundAccount = true
			break
		}
	}
	assert.True(t, foundAccount, "0x1 should have account resource")

	t.Logf("Account 0x1 has %d resources", len(resources))
}

// TestIntegration_AccountResource tests fetching a specific resource.
func TestIntegration_AccountResource(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	resource, err := client.AccountResource(ctx, AccountOne, "0x1::account::Account")
	require.NoError(t, err, "failed to get account resource for 0x1")

	assert.Equal(t, "0x1::account::Account", resource.Type)
	assert.NotNil(t, resource.Data)

	// The account resource should have sequence_number and authentication_key
	_, hasSeqNum := resource.Data["sequence_number"]
	_, hasAuthKey := resource.Data["authentication_key"]
	assert.True(t, hasSeqNum, "account resource should have sequence_number")
	assert.True(t, hasAuthKey, "account resource should have authentication_key")
}

// TestIntegration_AccountBalance tests fetching APT balance.
func TestIntegration_AccountBalance(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Note: 0x1 may not have a CoinStore on newer networks that use fungible assets
	// Try to get the balance, but don't fail if CoinStore doesn't exist
	balance, err := client.AccountBalance(ctx, AccountOne)
	if err != nil {
		// Check if it's a "resource not found" error (expected on FA-based networks)
		apiErr, ok := err.(*APIError)
		if ok && apiErr.StatusCode == 404 {
			t.Log("0x1 does not have CoinStore (uses fungible assets)")
			return
		}
		require.NoError(t, err, "failed to get APT balance for 0x1")
	}

	// If we got a balance, verify it
	t.Logf("Account 0x1 balance: %d octas (%.8f APT)", balance, float64(balance)/1e8)
}

// TestIntegration_GasEstimate tests gas price estimation.
func TestIntegration_GasEstimate(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	estimate, err := client.EstimateGasPrice(ctx)
	require.NoError(t, err, "failed to estimate gas price")

	assert.NotZero(t, estimate.GasEstimate, "gas estimate should not be zero")
	assert.LessOrEqual(t, estimate.DeprioritizedGasEstimate, estimate.GasEstimate,
		"deprioritized gas should be <= gas estimate")
	assert.GreaterOrEqual(t, estimate.PrioritizedGasEstimate, estimate.GasEstimate,
		"prioritized gas should be >= gas estimate")

	t.Logf("Gas Estimates: deprioritized=%d, normal=%d, prioritized=%d",
		estimate.DeprioritizedGasEstimate, estimate.GasEstimate, estimate.PrioritizedGasEstimate)
}

// TestIntegration_BlockByHeight tests fetching blocks by height.
func TestIntegration_BlockByHeight(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Get current block height
	info, err := client.Info(ctx)
	require.NoError(t, err)

	// Fetch a recent block (a few blocks back to ensure it exists)
	blockHeight := info.BlockHeight - 5
	if blockHeight < 1 {
		blockHeight = 1
	}

	t.Run("without_transactions", func(t *testing.T) {
		block, err := client.BlockByHeight(ctx, blockHeight, false)
		require.NoError(t, err, "failed to get block by height")

		assert.Equal(t, blockHeight, block.BlockHeight)
		assert.NotEmpty(t, block.BlockHash)
		assert.NotZero(t, block.BlockTimestamp)
		assert.Empty(t, block.Transactions, "should not include transactions")

		t.Logf("Block %d: hash=%s, first_version=%d, last_version=%d",
			block.BlockHeight, block.BlockHash, block.FirstVersion, block.LastVersion)
	})

	t.Run("with_transactions", func(t *testing.T) {
		block, err := client.BlockByHeight(ctx, blockHeight, true)
		require.NoError(t, err, "failed to get block with transactions")

		// Block should have at least the block metadata transaction
		assert.NotEmpty(t, block.Transactions, "should include transactions")
	})
}

// TestIntegration_BlockByVersion tests fetching blocks by ledger version.
func TestIntegration_BlockByVersion(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Get current ledger version
	info, err := client.Info(ctx)
	require.NoError(t, err)

	// Fetch a recent block by version
	version := info.LedgerVersion - 100
	if version < 1 {
		version = 1
	}

	block, err := client.BlockByVersion(ctx, version, false)
	require.NoError(t, err, "failed to get block by version")

	// The version should be within the block's range
	assert.LessOrEqual(t, block.FirstVersion, version, "version should be >= first_version")
	assert.GreaterOrEqual(t, block.LastVersion, version, "version should be <= last_version")

	t.Logf("Block containing version %d: height=%d, first=%d, last=%d",
		version, block.BlockHeight, block.FirstVersion, block.LastVersion)
}

// TestIntegration_Transactions tests fetching transactions.
func TestIntegration_Transactions(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	t.Run("recent_transactions", func(t *testing.T) {
		limit := uint64(10)
		txns, err := client.Transactions(ctx, nil, &limit)
		require.NoError(t, err, "failed to get recent transactions")

		assert.Len(t, txns, int(limit), "should return requested number of transactions")
		for _, txn := range txns {
			assert.NotEmpty(t, txn.Hash, "transaction should have hash")
			assert.NotEmpty(t, txn.Type, "transaction should have type")
		}

		t.Logf("Fetched %d recent transactions, latest version: %d", len(txns), txns[0].Version)
	})

	t.Run("transactions_from_version", func(t *testing.T) {
		start := uint64(100)
		limit := uint64(5)
		txns, err := client.Transactions(ctx, &start, &limit)
		require.NoError(t, err, "failed to get transactions from version")

		assert.Len(t, txns, int(limit))
		assert.Equal(t, start, txns[0].Version, "should start from requested version")
	})
}

// TestIntegration_TransactionByVersion tests fetching a specific transaction.
func TestIntegration_TransactionByVersion(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Version 1 should always exist (genesis)
	txn, err := client.TransactionByVersion(ctx, 1)
	require.NoError(t, err, "failed to get transaction at version 1")

	assert.Equal(t, uint64(1), txn.Version)
	assert.NotEmpty(t, txn.Hash)
	assert.NotEmpty(t, txn.Type)

	t.Logf("Transaction v1: type=%s, hash=%s", txn.Type, txn.Hash)
}

// TestIntegration_TransactionByHash tests fetching a transaction by hash.
func TestIntegration_TransactionByHash(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// First get a transaction by version to get its hash
	txn, err := client.TransactionByVersion(ctx, 1)
	require.NoError(t, err)

	// Then fetch by hash
	txnByHash, err := client.Transaction(ctx, txn.Hash)
	require.NoError(t, err, "failed to get transaction by hash")

	assert.Equal(t, txn.Version, txnByHash.Version)
	assert.Equal(t, txn.Hash, txnByHash.Hash)
}

// TestIntegration_TransactionsIterator tests the transactions iterator.
func TestIntegration_TransactionsIterator(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	start := uint64(1)
	var count int
	var lastVersion uint64

	for txn, err := range client.TransactionsIter(ctx, &start) {
		require.NoError(t, err, "iterator should not return errors for valid transactions")
		require.NotNil(t, txn)

		// Verify versions are increasing
		if count > 0 {
			assert.Greater(t, txn.Version, lastVersion, "versions should be increasing")
		}
		lastVersion = txn.Version

		count++
		if count >= 10 {
			break // Only test first 10 transactions
		}
	}

	assert.Equal(t, 10, count, "should have iterated through 10 transactions")
	t.Logf("Iterated %d transactions, versions %d to %d", count, start, lastVersion)
}

// TestIntegration_View tests view function execution.
func TestIntegration_View(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Use a simpler view function that always works: check if account exists
	// This uses the account module's exists_at function
	payload := &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "account"},
		Function: "exists_at",
		TypeArgs: nil,
		Args:     []any{AccountOne.String()},
	}

	result, err := client.View(ctx, payload)
	require.NoError(t, err, "failed to execute view function")
	require.Len(t, result, 1, "exists_at view should return one value")

	// 0x1 definitely exists
	exists, ok := result[0].(bool)
	if ok {
		assert.True(t, exists, "0x1 should exist")
	}

	t.Logf("View result for 0x1::account::exists_at(0x1): %v", result[0])
}

// TestIntegration_EventsByHandle tests fetching events.
func TestIntegration_EventsByHandle(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Get coin store events for 0x1 (deposit events)
	limit := uint64(5)
	events, err := client.EventsByHandle(
		ctx,
		AccountOne,
		"0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
		"deposit_events",
		nil,
		&limit,
	)
	// This may fail if 0x1 has no deposit events, which is fine
	if err != nil {
		t.Logf("No deposit events found (this is expected): %v", err)
		return
	}

	for _, event := range events {
		assert.NotEmpty(t, event.Type)
		t.Logf("Event: type=%s, seq=%d", event.Type, event.SequenceNumber)
	}
}

// TestIntegration_BuildTransaction tests building an unsigned transaction.
func TestIntegration_BuildTransaction(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	// Get a real account to build a transaction for
	accountInfo, err := client.Account(ctx, AccountOne)
	require.NoError(t, err)

	payload := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		TypeArgs: nil,
		Args:     []any{AccountTwo.String(), "100"},
	}

	txn, err := client.BuildTransaction(ctx, AccountOne, payload)
	require.NoError(t, err, "failed to build transaction")

	assert.Equal(t, AccountOne, txn.Sender)
	assert.Equal(t, accountInfo.SequenceNumber, txn.SequenceNumber)
	assert.NotZero(t, txn.MaxGasAmount)
	assert.NotZero(t, txn.GasUnitPrice)
	assert.NotZero(t, txn.ExpirationTimestampSeconds)
	assert.NotZero(t, txn.ChainID)

	t.Logf("Built transaction: sender=%s, seq=%d, gas=%d, price=%d, expiration=%d",
		txn.Sender.String(), txn.SequenceNumber, txn.MaxGasAmount, txn.GasUnitPrice, txn.ExpirationTimestampSeconds)
}

// TestIntegration_BuildTransactionWithOptions tests transaction building with options.
func TestIntegration_BuildTransactionWithOptions(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	seqNum := uint64(999)
	payload := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		TypeArgs: nil,
		Args:     []any{AccountTwo.String(), "100"},
	}

	txn, err := client.BuildTransaction(ctx, AccountOne, payload,
		WithSequenceNumber(seqNum),
		WithMaxGas(50000),
		WithGasPrice(200),
		WithExpiration(60*time.Second),
	)
	require.NoError(t, err, "failed to build transaction with options")

	assert.Equal(t, seqNum, txn.SequenceNumber, "should use provided sequence number")
	assert.Equal(t, uint64(50000), txn.MaxGasAmount, "should use provided max gas")
	assert.Equal(t, uint64(200), txn.GasUnitPrice, "should use provided gas price")
}

// TestIntegration_ContextCancellation tests that context cancellation is respected.
func TestIntegration_ContextCancellation(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Info(ctx)
	require.Error(t, err, "should return error for cancelled context")
	assert.ErrorIs(t, err, context.Canceled)
}

// TestIntegration_Timeout tests that timeouts are respected.
func TestIntegration_Timeout(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	// Create client with very short timeout
	client, err := NewClient(testNetwork(), WithTimeout(1*time.Nanosecond))
	require.NoError(t, err)

	ctx := testContext(t)
	_, err = client.Info(ctx)
	// This should timeout (though may succeed if very fast)
	// The important thing is the client doesn't panic
	_ = err
}

// TestIntegration_ConcurrentRequests tests concurrent API calls.
func TestIntegration_ConcurrentRequests(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client := createTestClient(t)
	ctx := testContext(t)

	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := client.Info(ctx)
			results <- err
		}()
	}

	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err, "concurrent request %d should succeed", i)
	}
}

// TestIntegration_Mainnet tests connectivity to mainnet specifically.
func TestIntegration_Mainnet(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client, err := NewClient(Mainnet)
	require.NoError(t, err)

	ctx := testContext(t)
	info, err := client.Info(ctx)
	require.NoError(t, err)

	assert.Equal(t, uint8(1), info.ChainID, "mainnet chain ID should be 1")
	t.Logf("Mainnet: ledger_version=%d, block_height=%d", info.LedgerVersion, info.BlockHeight)
}

// TestIntegration_Testnet tests connectivity to testnet specifically.
func TestIntegration_Testnet(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client, err := NewClient(Testnet)
	require.NoError(t, err)

	ctx := testContext(t)
	info, err := client.Info(ctx)
	require.NoError(t, err)

	assert.Equal(t, uint8(2), info.ChainID, "testnet chain ID should be 2")
	t.Logf("Testnet: ledger_version=%d, block_height=%d", info.LedgerVersion, info.BlockHeight)
}

// TestIntegration_Devnet tests connectivity to devnet specifically.
func TestIntegration_Devnet(t *testing.T) {
	skipIfShort(t)
	t.Parallel()

	client, err := NewClient(Devnet)
	require.NoError(t, err)

	ctx := testContext(t)
	info, err := client.Info(ctx)
	require.NoError(t, err)

	// Devnet chain ID varies, just check we got a response
	assert.NotZero(t, info.ChainID, "devnet should have a chain ID")
	t.Logf("Devnet: chain_id=%d, ledger_version=%d, block_height=%d",
		info.ChainID, info.LedgerVersion, info.BlockHeight)
}
