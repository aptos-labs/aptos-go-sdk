package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to test round-trip: unmarshal from JSON, marshal back, unmarshal again, compare.
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

func TestRoundTrip_U64(t *testing.T) {
	t.Parallel()
	var u U64
	err := json.Unmarshal([]byte(`"12345"`), &u)
	require.NoError(t, err)
	assert.Equal(t, U64(12345), u)

	b, err := json.Marshal(u)
	require.NoError(t, err)
	assert.Equal(t, `"12345"`, string(b))

	var u2 U64
	err = json.Unmarshal(b, &u2)
	require.NoError(t, err)
	assert.Equal(t, u, u2)
}

func TestRoundTrip_U64_Zero(t *testing.T) {
	t.Parallel()
	var u U64
	err := json.Unmarshal([]byte(`"0"`), &u)
	require.NoError(t, err)
	assert.Equal(t, U64(0), u)

	b, err := json.Marshal(u)
	require.NoError(t, err)
	assert.Equal(t, `"0"`, string(b))
}

func TestRoundTrip_GUID(t *testing.T) {
	t.Parallel()
	testJSON := `{"creation_number":"3","account_address":"0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b"}`
	assertRoundTrip[GUID](t, testJSON)
}

func TestRoundTrip_UserTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "1010733903",
  "hash": "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69",
  "state_change_hash": "0x3e8340786d2085a2160fa368c380ed412d4a5a3c5ccad692092c4bc0074fde3e",
  "event_root_hash": "0xe6e2ae41a57d9ab1c7dc58851d7beb4d5be43797ba7225d3e2a3b69c35fe7c2d",
  "gas_used": "5",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0xf9fdaddf6051311cb54e3756a343faa346f1c9137370762f6eef8e375a7031bb",
  "changes": [],
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
  "events": [],
  "timestamp": "1719965096135309",
  "type": "user_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)
}

func TestRoundTrip_UserTransaction_WithReplayNonce(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "6781425728",
  "hash": "0x650da387b6e7de0e9a83ca5f7c12e41ef865d0d12c38e793b55091da7ec90b39",
  "state_change_hash": "0x5b30b4306a1cbebe71dd319a02b84a78305f1a228de00f965c0e3350220e9f95",
  "event_root_hash": "0x5d3d8778f6d08e1668d9c2827841e63402531f8696a4df5b9bd36954ff048799",
  "gas_used": "57",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x24585b5a178b55a7f899522e6101a6b2bb65baf3e33bdc4215034a1c463c13bb",
  "changes": [],
  "sender": "0xa5b2dd0b5fe37a06f151d424cc880f67458c6fe3d8dbf44a55c6021cd111c10d",
  "sequence_number": "18446744073709551615",
  "max_gas_amount": "85",
  "gas_unit_price": "100",
  "expiration_timestamp_secs": "1749521012",
  "payload": {
    "function": "0x1::aptos_account::transfer",
    "type_arguments": [],
    "arguments": [
      "0x49ffe7968750a5ffea80af6fd7657bb246ff8ce6657cbf2c8ed13d9276096b3f",
      "20000"
    ],
    "type": "entry_function_payload"
  },
  "signature": {
    "public_key": "0xd782a5d884d64b26d73670109669f309224c553768a517fa0ebcce08ffedaa09",
    "signature": "0xc76ec8b8024497d0a70932da7b9e3ba9effff96e24626cb8bc6133c4b1544b54df6ea9fec05c53910034d341ef8424a502761df1ed658abba22b15a084224c06",
    "type": "ed25519_signature"
  },
  "replay_protection_nonce": "13894781796064092640",
  "events": [],
  "timestamp": "1749520982956875",
  "type": "user_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)

	// Also verify the nonce value round-trips correctly
	var txn Transaction
	require.NoError(t, json.Unmarshal([]byte(testJSON), &txn))
	marshaled, err := json.Marshal(&txn)
	require.NoError(t, err)
	var txn2 Transaction
	require.NoError(t, json.Unmarshal(marshaled, &txn2))
	ut1, _ := txn.UserTransaction()
	ut2, _ := txn2.UserTransaction()
	require.NotNil(t, ut1.ReplayProtectionNonce)
	require.NotNil(t, ut2.ReplayProtectionNonce)
	assert.Equal(t, *ut1.ReplayProtectionNonce, *ut2.ReplayProtectionNonce)
}

func TestRoundTrip_PendingTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "hash": "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69",
  "sender": "0xa46c6c7a65d605685e23055a6a906fb7284ba87849cbeb579d5c07424938241e",
  "sequence_number": "242217",
  "max_gas_amount": "2018",
  "gas_unit_price": "100",
  "expiration_timestamp_secs": "1719968695",
  "payload": {
    "function": "0x1::object::transfer",
    "type_arguments": ["0x4::token::Token"],
    "arguments": ["arg1"],
    "type": "entry_function_payload"
  },
  "signature": {
    "public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
    "signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
    "type": "ed25519_signature"
  },
  "type": "pending_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
}

func TestRoundTrip_GenesisTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "0",
  "hash": "0xcf5b7e186572be74741f81e2015146e6df15263082c2660690eccbd66a194043",
  "state_change_hash": "0xf5b27d111c2e8ce1de621031f456c8c8539b3a02822533f421692f041e586da7",
  "event_root_hash": "0x87862d624eb74dbdaeed74d0f6b9dc9f6eddc6ee1d167f9cc02c895524ad5a90",
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0xd57ebc779d5f764459dc6e618224f465313433b5a8615d8aa4864106e098b395",
  "changes": [],
  "events": [],
  "payload": {
    "write_set": {
      "changes": [],
      "events": [],
      "type": "direct_write_set"
    },
    "type": "write_set_payload"
  },
  "type": "genesis_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)
}

func TestRoundTrip_BlockMetadataTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "1",
  "hash": "0x30f2fea17d9cbab6bb06b34dd9cfb1d47a1eb20538c31ebaa508ce56d00628de",
  "state_change_hash": "0x0f75bad28c6be6f416befa62b67da6aac64fda84b7c3587c8a5b6064a37fc170",
  "event_root_hash": "0x050810c4262ab16c6dfccbc217e2fa5460319eea8b8e39de321c6c3824d8547f",
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x26fe2b1d7291824708f3b2beef477d654225ce8afdfc2b114957073b49a67f3c",
  "changes": [],
  "id": "0x81f7099ac9f45238ed4a98275add46f4da0a35ff62be0537846ca3d7c52bfbfc",
  "epoch": "1",
  "round": "1",
  "events": [],
  "previous_block_votes_bitvec": [0],
  "proposer": "0x90693588b138a37dbb37cb96c42ffb02bf48611fc9e78adeb57c8708ee3ac03e",
  "failed_proposer_indices": [1, 2],
  "timestamp": "1719520421743738",
  "type": "block_metadata_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)
}

func TestRoundTrip_StateCheckpointTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "3",
  "hash": "0x77da2c7a41ba6d46dc015c58f489c8d6ee030f98d95cca5b096578ca9e144aa6",
  "state_change_hash": "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6",
  "event_root_hash": "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000",
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x5e8e44711fba04cd509484a14b6071e50b06071e36d4b6ccf8edd724af0d6393",
  "changes": [],
  "timestamp": "1662686657332551",
  "type": "state_checkpoint_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)
}

func TestRoundTrip_BlockEpilogueTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "2",
  "hash": "0x1f19608413baaa8f39b670fbf001d17443ba7b975e0c22733bf742cea99fbdaf",
  "state_change_hash": "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6",
  "event_root_hash": "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000",
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x957c214e74b1aded27be7fd78b50c96fc0bfc25a70ad1555a08968a8fdc05cb1",
  "changes": [],
  "timestamp": "1719520421743738",
  "block_end_info": {
    "block_gas_limit_reached": false,
    "block_output_limit_reached": false,
    "block_effective_block_gas_units": 0,
    "block_approx_output_size": 3590
  },
  "type": "block_epilogue_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)
}

func TestRoundTrip_UnknownTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "2",
  "hash": "0x957c214e74b1aded27be7fd78b50c96fc0bfc25a70ad1555a08968a8fdc05cb1",
  "success": true,
  "type": "block_imaginary_transaction"
}`
	assertRoundTrip[Transaction](t, testJSON)
	assertRoundTrip[CommittedTransaction](t, testJSON)
}

func TestRoundTrip_Event(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "guid": {
    "creation_number": "3",
    "account_address": "0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b"
  },
  "sequence_number": "0",
  "type": "0x1::coin::WithdrawEvent",
  "data": {
    "amount": "1000"
  }
}`
	assertRoundTrip[Event](t, testJSON)
}

func TestRoundTrip_Block(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
  "block_height": "1",
  "block_timestamp": "1665609760857472",
  "first_version": "1",
  "last_version": "1",
  "transactions": null
}`
	assertRoundTrip[Block](t, testJSON)
}

func TestRoundTrip_BlockWithTransactions(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
  "block_height": "1",
  "block_timestamp": "1665609760857472",
  "first_version": "1",
  "last_version": "3",
  "transactions": [
    {
      "version": "1",
      "hash": "0x30f2fea17d9cbab6bb06b34dd9cfb1d47a1eb20538c31ebaa508ce56d00628de",
      "state_change_hash": "0x0f75bad28c6be6f416befa62b67da6aac64fda84b7c3587c8a5b6064a37fc170",
      "event_root_hash": "0x050810c4262ab16c6dfccbc217e2fa5460319eea8b8e39de321c6c3824d8547f",
      "gas_used": "0",
      "success": true,
      "vm_status": "Executed successfully",
      "accumulator_root_hash": "0x26fe2b1d7291824708f3b2beef477d654225ce8afdfc2b114957073b49a67f3c",
      "changes": [],
      "id": "0x81f7099ac9f45238ed4a98275add46f4da0a35ff62be0537846ca3d7c52bfbfc",
      "epoch": "1",
      "round": "1",
      "events": [],
      "previous_block_votes_bitvec": [0],
      "proposer": "0x90693588b138a37dbb37cb96c42ffb02bf48611fc9e78adeb57c8708ee3ac03e",
      "failed_proposer_indices": [],
      "timestamp": "1719520421743738",
      "type": "block_metadata_transaction"
    },
    {
      "version": "3",
      "hash": "0x77da2c7a41ba6d46dc015c58f489c8d6ee030f98d95cca5b096578ca9e144aa6",
      "state_change_hash": "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6",
      "event_root_hash": "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000",
      "gas_used": "0",
      "success": true,
      "vm_status": "Executed successfully",
      "accumulator_root_hash": "0x5e8e44711fba04cd509484a14b6071e50b06071e36d4b6ccf8edd724af0d6393",
      "changes": [],
      "timestamp": "1662686657332551",
      "type": "state_checkpoint_transaction"
    }
  ]
}`
	assertRoundTrip[Block](t, testJSON)
}

func TestRoundTrip_WriteSetChange(t *testing.T) {
	t.Parallel()

	// write_resource
	testJSON := `{
  "address": "0x1",
  "state_key_hash": "0x5ddf404c60e96e9485beafcabb95609fed8e38e941a725cae4dcec8296fb32d7",
  "data": {
    "type": "0x1::block::BlockResource",
    "data": {
      "height": "1"
    }
  },
  "type": "write_resource"
}`
	assertRoundTrip[WriteSetChange](t, testJSON)

	// delete_resource
	testJSON2 := `{
  "address": "0x1",
  "state_key_hash": "0x5ddf404c60e96e9485beafcabb95609fed8e38e941a725cae4dcec8296fb32d7",
  "resource": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
  "type": "delete_resource"
}`
	assertRoundTrip[WriteSetChange](t, testJSON2)

	// write_table_item
	testJSON3 := `{
  "state_key_hash": "0x6e4b28d40f98a106a65163530924c0dcb40c1349d3aa915d108b4d6cfc1ddb19",
  "handle": "0x1b854694ae746cdbd8d44186ca4929b2b337df21d1c74633be19b2710552fdca",
  "key": "0x0619dc29a0aac8fa146714058e8dd6d2d0f3bdf5f6331907bf91f3acd81e6935",
  "value": "0xd6b01fdfed0ebb020000000000000000",
  "type": "write_table_item"
}`
	assertRoundTrip[WriteSetChange](t, testJSON3)
}

func TestRoundTrip_TransactionPayload(t *testing.T) {
	t.Parallel()

	// entry_function_payload
	testJSON := `{
  "function": "0x1::object::transfer",
  "type_arguments": ["0x4::token::Token"],
  "arguments": ["arg1", "arg2"],
  "type": "entry_function_payload"
}`
	assertRoundTrip[TransactionPayload](t, testJSON)

	// write_set_payload
	testJSON2 := `{
  "write_set": {
    "changes": [],
    "events": [],
    "type": "direct_write_set"
  },
  "type": "write_set_payload"
}`
	assertRoundTrip[TransactionPayload](t, testJSON2)
}

func TestRoundTrip_Signature(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
  "signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
  "type": "ed25519_signature"
}`
	assertRoundTrip[Signature](t, testJSON)
}

func TestRoundTrip_FeePayerSignature(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "sender": {
    "public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
    "signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
    "type": "ed25519_signature"
  },
  "secondary_signer_addresses": [],
  "secondary_signers": [],
  "fee_payer_address": "0x1",
  "fee_payer_signer": {
    "public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
    "signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
    "type": "ed25519_signature"
  },
  "type": "fee_payer_signature"
}`
	assertRoundTrip[Signature](t, testJSON)
}

func TestRoundTrip_WriteSet(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "changes": [],
  "events": [],
  "type": "direct_write_set"
}`
	assertRoundTrip[WriteSet](t, testJSON)
}

// TestMarshalUnmarshal_FullTransaction tests the exact scenario from the issue:
// marshal a Transaction, then unmarshal it back successfully.
func TestMarshalUnmarshal_FullTransaction(t *testing.T) {
	t.Parallel()
	testJSON := `{
  "version": "1010733903",
  "hash": "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69",
  "state_change_hash": "0x3e8340786d2085a2160fa368c380ed412d4a5a3c5ccad692092c4bc0074fde3e",
  "event_root_hash": "0xe6e2ae41a57d9ab1c7dc58851d7beb4d5be43797ba7225d3e2a3b69c35fe7c2d",
  "gas_used": "5",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0xf9fdaddf6051311cb54e3756a343faa346f1c9137370762f6eef8e375a7031bb",
  "changes": [
    {
      "address": "0x1",
      "state_key_hash": "0x5ddf404c60e96e9485beafcabb95609fed8e38e941a725cae4dcec8296fb32d7",
      "data": {
        "type": "0x1::block::BlockResource",
        "data": {
          "height": "1"
        }
      },
      "type": "write_resource"
    }
  ],
  "sender": "0xa46c6c7a65d605685e23055a6a906fb7284ba87849cbeb579d5c07424938241e",
  "sequence_number": "242217",
  "max_gas_amount": "2018",
  "gas_unit_price": "100",
  "expiration_timestamp_secs": "1719968695",
  "payload": {
    "function": "0x1::object::transfer",
    "type_arguments": ["0x4::token::Token"],
    "arguments": ["arg1"],
    "type": "entry_function_payload"
  },
  "signature": {
    "public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
    "signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
    "type": "ed25519_signature"
  },
  "events": [
    {
      "guid": {
        "creation_number": "0",
        "account_address": "0x0"
      },
      "sequence_number": "0",
      "type": "0x1::transaction_fee::FeeStatement",
      "data": {
        "total_charge_gas_units": "5"
      }
    }
  ],
  "timestamp": "1719965096135309",
  "type": "user_transaction"
}`

	// Step 1: Unmarshal original JSON
	var tx1 Transaction
	err := json.Unmarshal([]byte(testJSON), &tx1)
	require.NoError(t, err)

	// Step 2: Marshal it back (this is what the issue requests)
	jsonBytes, err := json.Marshal(&tx1)
	require.NoError(t, err)

	// Step 3: Unmarshal the marshaled JSON (this should not error - the core of the issue)
	var tx2 Transaction
	err = json.Unmarshal(jsonBytes, &tx2)
	require.NoError(t, err, "re-unmarshaling marshaled JSON failed: %s", string(jsonBytes))

	// Step 4: Verify the structs are equal
	assert.Equal(t, tx1.Type, tx2.Type)
	assert.Equal(t, tx1.Hash(), tx2.Hash())

	ut1, err := tx1.UserTransaction()
	require.NoError(t, err)
	ut2, err := tx2.UserTransaction()
	require.NoError(t, err)

	assert.Equal(t, ut1.Version, ut2.Version)
	assert.Equal(t, ut1.Hash, ut2.Hash)
	assert.Equal(t, ut1.Sender, ut2.Sender)
	assert.Equal(t, ut1.SequenceNumber, ut2.SequenceNumber)
	assert.Equal(t, ut1.GasUsed, ut2.GasUsed)
	assert.Equal(t, ut1.Success, ut2.Success)
	assert.Equal(t, ut1.Timestamp, ut2.Timestamp)
	assert.Len(t, ut2.Changes, len(ut1.Changes))
	assert.Len(t, ut2.Events, len(ut1.Events))
}
