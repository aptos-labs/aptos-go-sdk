# Aptos Go SDK v2 Integration Tests

This directory contains integration tests that verify the v2 SDK functionality against live Aptos networks.

## Overview

The integration tests are designed to:
- Verify connectivity to all Aptos networks (mainnet, testnet, devnet)
- Test all client API methods against real blockchain state
- Validate error handling and edge cases
- Test concurrent request handling
- Verify context cancellation and timeout behavior

## Running the Tests

### Run All Integration Tests

```bash
go test -v -run "TestIntegration_" -timeout 120s ./v2/...
```

### Run Only Unit Tests (Skip Integration)

```bash
go test -short ./v2/...
```

### Run Faucet Tests (Testnet/Devnet Only)

Faucet tests are disabled by default to avoid unnecessary testnet token requests:

```bash
APTOS_TEST_FAUCET=1 go test -v -run "TestIntegrationFaucet_" ./v2/...
```

### Target a Specific Network

By default, tests run against testnet. Override with `APTOS_NETWORK`:

```bash
APTOS_NETWORK=devnet go test -v -run "TestIntegration_" ./v2/...
APTOS_NETWORK=mainnet go test -v -run "TestIntegration_" ./v2/...
```

## Test Categories

### Client Creation (`TestIntegration_ClientCreation`)
Tests that clients can be created for all networks.

### Node Information (`TestIntegration_NodeInfo`, `TestIntegration_ChainID`)
Tests fetching node metadata and chain ID caching.

### Account Operations (`TestIntegration_Account*`)
- `TestIntegration_Account` - Get account info
- `TestIntegration_AccountNotFound` - Error handling for missing accounts
- `TestIntegration_AccountResources` - Get all resources
- `TestIntegration_AccountResource` - Get specific resource
- `TestIntegration_AccountBalance` - Get APT balance

### Transaction Operations (`TestIntegration_Transaction*`)
- `TestIntegration_Transactions` - List transactions
- `TestIntegration_TransactionByVersion` - Get by version
- `TestIntegration_TransactionByHash` - Get by hash
- `TestIntegration_TransactionsIterator` - Go 1.23 iterator

### Block Operations (`TestIntegration_Block*`)
- `TestIntegration_BlockByHeight` - Get block by height
- `TestIntegration_BlockByVersion` - Get block containing version

### View Functions (`TestIntegration_View`)
Tests executing read-only view functions.

### Events (`TestIntegration_EventsByHandle`)
Tests fetching events from event handles.

### Transaction Building (`TestIntegration_BuildTransaction*`)
Tests building unsigned transactions with various options.

### Network Connectivity
- `TestIntegration_Mainnet` - Verify mainnet connectivity
- `TestIntegration_Testnet` - Verify testnet connectivity
- `TestIntegration_Devnet` - Verify devnet connectivity

### Error Handling
- `TestIntegration_ContextCancellation` - Verify cancelled contexts are respected
- `TestIntegration_Timeout` - Verify timeouts work correctly

### Concurrency (`TestIntegration_ConcurrentRequests`)
Tests that concurrent requests are handled correctly.

## Rate Limiting

The public Aptos nodes have rate limits. If running many tests in quick succession, you may see 429 errors. This is expected behavior. To avoid this:

1. Run tests with a delay between runs
2. Use an API key via `WithAPIKey()` option
3. Run tests against a local node

## Writing New Integration Tests

When adding new integration tests:

1. Use `skipIfShort(t)` at the start to skip during `-short` mode
2. Use `t.Parallel()` for tests that can run concurrently
3. Use `testContext(t)` to get a context with reasonable timeout
4. Use `createTestClient(t)` to get a configured client
5. Handle network variability (accounts may or may not exist)
6. Log useful information for debugging

Example:

```go
func TestIntegration_NewFeature(t *testing.T) {
    skipIfShort(t)
    t.Parallel()

    client := createTestClient(t)
    ctx := testContext(t)

    result, err := client.NewFeature(ctx, ...)
    require.NoError(t, err)
    assert.NotNil(t, result)

    t.Logf("NewFeature result: %+v", result)
}
```

