package aptos

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleID_String(t *testing.T) {
	tests := []struct {
		name     string
		module   ModuleID
		expected string
	}{
		{
			name:     "account one coin",
			module:   ModuleID{Address: AccountOne, Name: "coin"},
			expected: "0x1::coin",
		},
		{
			name:     "zero address aptos_coin",
			module:   ModuleID{Address: AccountZero, Name: "aptos_coin"},
			expected: "0x0::aptos_coin",
		},
		{
			name:     "full address",
			module:   ModuleID{Address: MustParseAddress("0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"), Name: "my_module"},
			expected: "0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234::my_module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAccountInfo_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"sequence_number": "42",
		"authentication_key": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	}`

	var info AccountInfo
	err := json.Unmarshal([]byte(jsonData), &info)
	require.NoError(t, err)

	assert.Equal(t, uint64(42), info.SequenceNumber)
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", info.AuthenticationKey)
}

func TestNodeInfo_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"chain_id": 4,
		"epoch": "100",
		"ledger_version": "1000000",
		"oldest_ledger_version": "1",
		"ledger_timestamp": "1700000000000000",
		"node_role": "full_node",
		"oldest_block_height": "0",
		"block_height": "50000",
		"git_hash": "abc123def456"
	}`

	var info NodeInfo
	err := json.Unmarshal([]byte(jsonData), &info)
	require.NoError(t, err)

	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, uint64(100), info.Epoch)
	assert.Equal(t, uint64(1000000), info.LedgerVersion)
	assert.Equal(t, uint64(1), info.OldestLedgerVersion)
	assert.Equal(t, uint64(1700000000000000), info.LedgerTimestamp)
	assert.Equal(t, "full_node", info.NodeRole)
	assert.Equal(t, uint64(0), info.OldestBlockHeight)
	assert.Equal(t, uint64(50000), info.BlockHeight)
	assert.Equal(t, "abc123def456", info.GitHash)
}

func TestGasEstimate_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"gas_estimate": 100,
		"deprioritized_gas_estimate": 50,
		"prioritized_gas_estimate": 200
	}`

	var estimate GasEstimate
	err := json.Unmarshal([]byte(jsonData), &estimate)
	require.NoError(t, err)

	assert.Equal(t, uint64(100), estimate.GasEstimate)
	assert.Equal(t, uint64(50), estimate.DeprioritizedGasEstimate)
	assert.Equal(t, uint64(200), estimate.PrioritizedGasEstimate)
}

func TestSubmitResult_JSONUnmarshal(t *testing.T) {
	jsonData := `{"hash": "0xabcdef123456"}`

	var result SubmitResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	assert.Equal(t, "0xabcdef123456", result.Hash)
}

func TestSimulationResult_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"success": true,
		"vm_status": "Executed successfully",
		"gas_used": "100",
		"gas_unit_price": "1",
		"changes": [],
		"events": []
	}`

	var result SimulationResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "Executed successfully", result.VMStatus)
	assert.Equal(t, uint64(100), result.GasUsed)
	assert.Equal(t, uint64(1), result.GasUnitPrice)
}

func TestTransaction_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"type": "user_transaction",
		"version": "1000",
		"hash": "0xabc123",
		"state_change_hash": "0xdef456",
		"event_root_hash": "0x789",
		"gas_used": "100",
		"success": true,
		"vm_status": "Executed successfully",
		"accumulator_root_hash": "0xacc",
		"timestamp": "1700000000000000",
		"sender": "0x1",
		"sequence_number": "5",
		"max_gas_amount": "10000",
		"gas_unit_price": "1",
		"expiration_timestamp_secs": "1700000100"
	}`

	var txn Transaction
	err := json.Unmarshal([]byte(jsonData), &txn)
	require.NoError(t, err)

	assert.Equal(t, "user_transaction", txn.Type)
	assert.Equal(t, uint64(1000), txn.Version)
	assert.Equal(t, "0xabc123", txn.Hash)
	assert.True(t, txn.Success)
	assert.Equal(t, uint64(100), txn.GasUsed)
	assert.Equal(t, "0x1", txn.Sender)
	assert.Equal(t, uint64(5), txn.SequenceNumber)
}

func TestBlock_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"block_height": "100",
		"block_hash": "0xblockhash",
		"block_timestamp": "1700000000000000",
		"first_version": "1000",
		"last_version": "1100"
	}`

	var block Block
	err := json.Unmarshal([]byte(jsonData), &block)
	require.NoError(t, err)

	assert.Equal(t, uint64(100), block.BlockHeight)
	assert.Equal(t, "0xblockhash", block.BlockHash)
	assert.Equal(t, uint64(1000), block.FirstVersion)
	assert.Equal(t, uint64(1100), block.LastVersion)
}

func TestEvent_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"guid": {
			"creation_number": "1",
			"account_address": "0x1"
		},
		"sequence_number": "10",
		"type": "0x1::coin::WithdrawEvent",
		"data": {"amount": "1000"}
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	require.NoError(t, err)

	assert.Equal(t, uint64(1), event.GUID.CreationNumber)
	assert.Equal(t, "0x1", event.GUID.AccountAddress)
	assert.Equal(t, uint64(10), event.SequenceNumber)
	assert.Equal(t, "0x1::coin::WithdrawEvent", event.Type)
}

func TestBatchSubmitResult_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"transaction_failures": [
			{
				"error": {
					"message": "Transaction failed",
					"error_code": "invalid_transaction"
				},
				"transaction_index": 2
			}
		]
	}`

	var result BatchSubmitResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	require.Len(t, result.TransactionFailures, 1)
	assert.Equal(t, 2, result.TransactionFailures[0].TransactionIndex)
	assert.Equal(t, "Transaction failed", result.TransactionFailures[0].Error.Message)
}

func TestResource_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"type": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
		"data": {
			"coin": {
				"value": "1000000"
			}
		}
	}`

	var resource Resource
	err := json.Unmarshal([]byte(jsonData), &resource)
	require.NoError(t, err)

	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", resource.Type)
	assert.NotNil(t, resource.Data["coin"])
}

func TestModuleABI_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"address": "0x1",
		"name": "coin",
		"friends": ["0x1::account"],
		"exposed_functions": [
			{
				"name": "transfer",
				"visibility": "public",
				"is_entry": true,
				"is_view": false,
				"generic_type_params": [],
				"params": ["&signer", "address", "u64"],
				"return": []
			}
		],
		"structs": [
			{
				"name": "CoinStore",
				"is_native": false,
				"abilities": ["key"],
				"generic_type_params": [],
				"fields": [
					{"name": "coin", "type": "Coin<CoinType>"}
				]
			}
		]
	}`

	var abi ModuleABI
	err := json.Unmarshal([]byte(jsonData), &abi)
	require.NoError(t, err)

	assert.Equal(t, "0x1", abi.Address)
	assert.Equal(t, "coin", abi.Name)
	require.Len(t, abi.ExposedFunctions, 1)
	assert.Equal(t, "transfer", abi.ExposedFunctions[0].Name)
	assert.True(t, abi.ExposedFunctions[0].IsEntry)
	require.Len(t, abi.Structs, 1)
	assert.Equal(t, "CoinStore", abi.Structs[0].Name)
}

func TestHealthCheckResponse_JSONUnmarshal(t *testing.T) {
	jsonData := `{"message": "aptos-node:ok"}`

	var resp HealthCheckResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)

	assert.Equal(t, "aptos-node:ok", resp.Message)
}
