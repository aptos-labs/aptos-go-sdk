package aptos

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertRoundTrip verifies that a type can be unmarshaled, marshaled back, and re-unmarshaled to an equal value.
func assertRoundTrip[T any](t *testing.T, testJSON string) {
	t.Helper()

	var v1 T
	err := json.Unmarshal([]byte(testJSON), &v1)
	require.NoError(t, err, "initial unmarshal failed")

	marshaled, err := json.Marshal(&v1)
	require.NoError(t, err, "marshal failed")

	var v2 T
	err = json.Unmarshal(marshaled, &v2)
	require.NoError(t, err, "re-unmarshal failed: %s", string(marshaled))

	assert.Equal(t, v1, v2, "round-trip produced different values")
}

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

func TestRoundTrip_AccountInfo(t *testing.T) {
	t.Parallel()
	assertRoundTrip[AccountInfo](t, `{"sequence_number":"42","authentication_key":"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"}`)
}

func TestRoundTrip_NodeInfo(t *testing.T) {
	t.Parallel()
	assertRoundTrip[NodeInfo](t, `{
		"chain_id": 4,
		"epoch": "100",
		"ledger_version": "1000000",
		"oldest_ledger_version": "1",
		"ledger_timestamp": "1700000000000000",
		"node_role": "full_node",
		"oldest_block_height": "0",
		"block_height": "50000",
		"git_hash": "abc123def456"
	}`)
}

func TestRoundTrip_GasEstimate(t *testing.T) {
	t.Parallel()
	assertRoundTrip[GasEstimate](t, `{
		"gas_estimate": 100,
		"deprioritized_gas_estimate": 50,
		"prioritized_gas_estimate": 200
	}`)
}

func TestRoundTrip_SubmitResult(t *testing.T) {
	t.Parallel()
	assertRoundTrip[SubmitResult](t, `{"hash":"0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69"}`)
}

func TestRoundTrip_SimulationResult(t *testing.T) {
	t.Parallel()
	assertRoundTrip[SimulationResult](t, `{
		"success": true,
		"vm_status": "Executed successfully",
		"gas_used": "100",
		"gas_unit_price": "1",
		"changes": [],
		"events": []
	}`)
}

func TestRoundTrip_Transaction(t *testing.T) {
	t.Parallel()
	assertRoundTrip[Transaction](t, `{
		"type": "user_transaction",
		"version": "1010733903",
		"hash": "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69",
		"state_change_hash": "0x3e8340786d2085a2160fa368c380ed412d4a5a3c5ccad692092c4bc0074fde3e",
		"event_root_hash": "0xe6e2ae41a57d9ab1c7dc58851d7beb4d5be43797ba7225d3e2a3b69c35fe7c2d",
		"gas_used": "5",
		"success": true,
		"vm_status": "Executed successfully",
		"accumulator_root_hash": "0xf9fdaddf6051311cb54e3756a343faa346f1c9137370762f6eef8e375a7031bb",
		"timestamp": "1719965096135309",
		"sender": "0xa46c6c7a65d605685e23055a6a906fb7284ba87849cbeb579d5c07424938241e",
		"sequence_number": "242217",
		"max_gas_amount": "2018",
		"gas_unit_price": "100",
		"expiration_timestamp_secs": "1719968695",
		"payload": {
			"function": "0x1::object::transfer",
			"type_arguments": ["0x4::token::Token"],
			"arguments": ["0x8038df5e61a19a5f86ad01f4389736b08250dad1b4aa864afc4fc639a2581ca8"],
			"type": "entry_function_payload"
		},
		"signature": {
			"public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
			"signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
			"type": "ed25519_signature"
		},
		"events": [
			{
				"guid": {"creation_number": "0", "account_address": "0x0"},
				"sequence_number": "0",
				"type": "0x1::transaction_fee::FeeStatement",
				"data": {"total_charge_gas_units": "5"}
			}
		],
		"changes": [
			{
				"type": "write_resource",
				"address": "0x1",
				"state_key_hash": "0x5ddf404c60e96e9485beafcabb95609fed8e38e941a725cae4dcec8296fb32d7",
				"data": {"type": "0x1::block::BlockResource", "data": {"height": "1"}}
			}
		]
	}`)
}

func TestRoundTrip_Block(t *testing.T) {
	t.Parallel()
	assertRoundTrip[Block](t, `{
		"block_height": "100",
		"block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
		"block_timestamp": "1665609760857472",
		"first_version": "1",
		"last_version": "3",
		"transactions": [
			{
				"type": "block_metadata_transaction",
				"version": "1",
				"hash": "0x30f2fea17d9cbab6bb06b34dd9cfb1d47a1eb20538c31ebaa508ce56d00628de",
				"state_change_hash": "0x0f75bad28c6be6f416befa62b67da6aac64fda84b7c3587c8a5b6064a37fc170",
				"event_root_hash": "0x050810c4262ab16c6dfccbc217e2fa5460319eea8b8e39de321c6c3824d8547f",
				"gas_used": "0",
				"success": true,
				"vm_status": "Executed successfully",
				"accumulator_root_hash": "0x26fe2b1d7291824708f3b2beef477d654225ce8afdfc2b114957073b49a67f3c",
				"timestamp": "1719520421743738"
			}
		]
	}`)
}

func TestRoundTrip_Event(t *testing.T) {
	t.Parallel()
	assertRoundTrip[Event](t, `{
		"guid": {"creation_number": "3", "account_address": "0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b"},
		"sequence_number": "10",
		"type": "0x1::coin::WithdrawEvent",
		"data": {"amount": "1000"}
	}`)
}

func TestRoundTrip_Change(t *testing.T) {
	t.Parallel()
	assertRoundTrip[Change](t, `{
		"type": "write_resource",
		"address": "0x1",
		"state_key_hash": "0x5ddf404c60e96e9485beafcabb95609fed8e38e941a725cae4dcec8296fb32d7",
		"data": {"type": "0x1::block::BlockResource", "data": {"height": "1"}}
	}`)
}

func TestRoundTrip_Resource(t *testing.T) {
	t.Parallel()
	assertRoundTrip[Resource](t, `{
		"type": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
		"data": {"coin": {"value": "1000000"}}
	}`)
}

func TestRoundTrip_BatchSubmitResult(t *testing.T) {
	t.Parallel()
	assertRoundTrip[BatchSubmitResult](t, `{
		"transaction_failures": [
			{
				"error": {"message": "Transaction failed", "error_code": "invalid_transaction"},
				"transaction_index": 2
			}
		]
	}`)
}

func TestRoundTrip_HealthCheckResponse(t *testing.T) {
	t.Parallel()
	assertRoundTrip[HealthCheckResponse](t, `{"message":"aptos-node:ok"}`)
}

func TestRoundTrip_ModuleABI(t *testing.T) {
	t.Parallel()
	assertRoundTrip[ModuleABI](t, `{
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
				"fields": [{"name": "coin", "type": "Coin<CoinType>"}]
			}
		]
	}`)
}
