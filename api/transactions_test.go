package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransaction_GenesisTransaction(t *testing.T) {
	testJson := `{
  "version": "0",
  "hash": "0xcf5b7e186572be74741f81e2015146e6df15263082c2660690eccbd66a194043",
  "state_change_hash": "0xf5b27d111c2e8ce1de621031f456c8c8539b3a02822533f421692f041e586da7",
  "event_root_hash": "0x87862d624eb74dbdaeed74d0f6b9dc9f6eddc6ee1d167f9cc02c895524ad5a90",
  "state_checkpoint_hash": "0x638a52558edcfef9bd0e327c90f8d1c079a6305d64130128ed09b08b45bd63c7",
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0xd57ebc779d5f764459dc6e618224f465313433b5a8615d8aa4864106e098b395",
  "changes": [
    {
      "address": "0x1",
      "state_key_hash": "0xd0f08e53567bf9ab0f087675e1ee8f82c7f8d087b36363f2cf9e6387f2e44b3f",
      "data": {
        "bytecode": "0xa11ceb0b",
        "abi": {
          "address": "0x1",
          "name": "managed_coin",
          "friends": [],
          "exposed_functions": [
            {
              "name": "burn",
              "visibility": "public",
              "is_entry": true,
              "is_view": false,
              "generic_type_params": [
                {
                  "constraints": []
                }
              ],
              "params": [
                "&signer",
                "u64"
              ],
              "return": []
            },
            {
              "name": "initialize",
              "visibility": "public",
              "is_entry": true,
              "is_view": false,
              "generic_type_params": [
                {
                  "constraints": []
                }
              ],
              "params": [
                "&signer",
                "vector\u003Cu8\u003E",
                "vector\u003Cu8\u003E",
                "u8",
                "bool"
              ],
              "return": []
            },
            {
              "name": "mint",
              "visibility": "public",
              "is_entry": true,
              "is_view": false,
              "generic_type_params": [
                {
                  "constraints": []
                }
              ],
              "params": [
                "&signer",
                "address",
                "u64"
              ],
              "return": []
            },
            {
              "name": "register",
              "visibility": "public",
              "is_entry": true,
              "is_view": false,
              "generic_type_params": [
                {
                  "constraints": []
                }
              ],
              "params": [
                "&signer"
              ],
              "return": []
            }
          ],
          "structs": [
            {
              "name": "Capabilities",
              "is_native": false,
              "abilities": [
                "key"
              ],
              "generic_type_params": [
                {
                  "constraints": []
                }
              ],
              "fields": [
                {
                  "name": "burn_cap",
                  "type": "0x1::coin::BurnCapability\u003CT0\u003E"
                },
                {
                  "name": "freeze_cap",
                  "type": "0x1::coin::FreezeCapability\u003CT0\u003E"
                },
                {
                  "name": "mint_cap",
                  "type": "0x1::coin::MintCapability\u003CT0\u003E"
                }
              ]
            }
          ]
        }
      },
      "type": "write_module"
    }
  ],
  "events": [
    {
      "guid": {
        "creation_number": "0",
        "account_address": "0x0"
      },
      "sequence_number": "0",
      "type": "0x1::coin::PairCreation",
      "data": {
        "coin_type": {
          "account_address": "0x1",
          "module_name": "0x6170746f735f636f696e",
          "struct_name": "0x4170746f73436f696e"
        },
        "fungible_asset_metadata_address": "0xa"
      }
    }
  ],
  "payload": {
    "write_set": {
      "type": "direct_write_set"
    },
    "type": "write_set_payload"
  },
  "type": "genesis_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantGenesis, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantGenesis, data2.Type)

	txn, err := data.GenesisTransaction()
	assert.NoError(t, err)
	txn2, err := data2.GenesisTransaction()
	assert.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(0), txn.Version)

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}

func TestTransaction_BlockMetadataTransaction(t *testing.T) {
	testJson := `{
  "version": "1",
  "hash": "0x30f2fea17d9cbab6bb06b34dd9cfb1d47a1eb20538c31ebaa508ce56d00628de",
  "state_change_hash": "0x0f75bad28c6be6f416befa62b67da6aac64fda84b7c3587c8a5b6064a37fc170",
  "event_root_hash": "0x050810c4262ab16c6dfccbc217e2fa5460319eea8b8e39de321c6c3824d8547f",
  "state_checkpoint_hash": null,
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x26fe2b1d7291824708f3b2beef477d654225ce8afdfc2b114957073b49a67f3c",
  "changes": [
    {
      "address": "0x1",
      "state_key_hash": "0x5ddf404c60e96e9485beafcabb95609fed8e38e941a725cae4dcec8296fb32d7",
      "data": {
        "type": "0x1::block::BlockResource",
        "data": {
          "epoch_interval": "7200000000",
          "height": "1",
          "new_block_events": {
            "counter": "2",
            "guid": {
              "id": {
                "addr": "0x1",
                "creation_num": "3"
              }
            }
          },
          "update_epoch_interval_events": {
            "counter": "0",
            "guid": {
              "id": {
                "addr": "0x1",
                "creation_num": "4"
              }
            }
          }
        }
      },
      "type": "write_resource"
    }
  ],
  "id": "0x81f7099ac9f45238ed4a98275add46f4da0a35ff62be0537846ca3d7c52bfbfc",
  "epoch": "1",
  "round": "1",
  "events": [
    {
      "guid": {
        "creation_number": "0",
        "account_address": "0x0"
      },
      "sequence_number": "0",
      "type": "0x1::block::NewBlock",
      "data": {
        "epoch": "1",
        "failed_proposer_indices": [],
        "hash": "0x81f7099ac9f45238ed4a98275add46f4da0a35ff62be0537846ca3d7c52bfbfc",
        "height": "1",
        "previous_block_votes_bitvec": "0x00",
        "proposer": "0x90693588b138a37dbb37cb96c42ffb02bf48611fc9e78adeb57c8708ee3ac03e",
        "round": "1",
        "time_microseconds": "1719520421743738"
      }
    }
  ],
  "previous_block_votes_bitvec": [
    0
  ],
  "proposer": "0x90693588b138a37dbb37cb96c42ffb02bf48611fc9e78adeb57c8708ee3ac03e",
  "failed_proposer_indices": [1, 2],
  "timestamp": "1719520421743738",
  "type": "block_metadata_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockMetadata, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockMetadata, data2.Type)

	txn, err := data.BlockMetadataTransaction()
	assert.NoError(t, err)
	txn2, err := data2.BlockMetadataTransaction()
	assert.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(1), txn.Version)
	assert.Equal(t, "0x81f7099ac9f45238ed4a98275add46f4da0a35ff62be0537846ca3d7c52bfbfc", txn.Id)
	assert.Equal(t, uint64(1), txn.Epoch)
	assert.Equal(t, uint64(1), txn.Round)
	assert.Equal(t, uint64(1719520421743738), txn.Timestamp)

	address := &types.AccountAddress{}
	err = address.ParseStringRelaxed("0x90693588b138a37dbb37cb96c42ffb02bf48611fc9e78adeb57c8708ee3ac03e")
	assert.NoError(t, err)
	assert.Equal(t, address, txn.Proposer)
	assert.Equal(t, []uint32{1, 2}, txn.FailedProposerIndices)
	assert.Equal(t, []uint8{0}, txn.PreviousBlockVotesBitvec)

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}

func TestTransaction_StateCheckpointTransaction(t *testing.T) {
	testJson := `{
  "version": "3",
  "hash": "0x77da2c7a41ba6d46dc015c58f489c8d6ee030f98d95cca5b096578ca9e144aa6",
  "state_change_hash": "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6",
  "event_root_hash": "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000",
  "state_checkpoint_hash": "0x56bf9bb8d9049d2f56541c19f48da847dd5c12419529f8db97255b08c2cf42b7",
  "gas_used": "0",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x5e8e44711fba04cd509484a14b6071e50b06071e36d4b6ccf8edd724af0d6393",
  "changes": [],
  "timestamp": "1662686657332551",
  "type": "state_checkpoint_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantStateCheckpoint, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantStateCheckpoint, data2.Type)

	txn, err := data.StateCheckpointTransaction()
	assert.NoError(t, err)
	txn2, err := data2.StateCheckpointTransaction()
	assert.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(3), txn.Version)
	assert.Equal(t, "0x77da2c7a41ba6d46dc015c58f489c8d6ee030f98d95cca5b096578ca9e144aa6", txn.Hash)
	assert.Equal(t, "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6", txn.StateChangeHash)
	assert.Equal(t, "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000", txn.EventRootHash)
	assert.Equal(t, "0x56bf9bb8d9049d2f56541c19f48da847dd5c12419529f8db97255b08c2cf42b7", txn.StateCheckpointHash)
	assert.Equal(t, uint64(1662686657332551), txn.Timestamp)
	assert.Equal(t, uint64(0), txn.GasUsed)
	assert.True(t, txn.Success)
	assert.Equal(t, "Executed successfully", txn.VmStatus)
	assert.Equal(t, "0x5e8e44711fba04cd509484a14b6071e50b06071e36d4b6ccf8edd724af0d6393", txn.AccumulatorRootHash)
	assert.Empty(t, txn.Changes)

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}

func TestTransaction_BlockEpilogueTransaction(t *testing.T) {
	testJson := `{
  "version": "2",
  "hash": "0x1f19608413baaa8f39b670fbf001d17443ba7b975e0c22733bf742cea99fbdaf",
  "state_change_hash": "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6",
  "event_root_hash": "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000",
  "state_checkpoint_hash": "0x986343cd66e79d3f8b52fcd65df05da9801f0894ac4b5c27d079a8bdadbaa432",
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
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockEpilogue, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockEpilogue, data2.Type)

	txn, err := data.BlockEpilogueTransaction()
	assert.NoError(t, err)
	txn2, err := data2.BlockEpilogueTransaction()
	assert.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(2), txn.Version)
	assert.Equal(t, "0x1f19608413baaa8f39b670fbf001d17443ba7b975e0c22733bf742cea99fbdaf", txn.Hash)
	assert.Equal(t, "0xafb6e14fe47d850fd0a7395bcfb997ffacf4715e0f895cc162c218e4a7564bc6", txn.StateChangeHash)
	assert.Equal(t, "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000", txn.EventRootHash)
	assert.Equal(t, "0x986343cd66e79d3f8b52fcd65df05da9801f0894ac4b5c27d079a8bdadbaa432", txn.StateCheckpointHash)
	assert.Equal(t, uint64(1719520421743738), txn.Timestamp)
	assert.Equal(t, uint64(0), txn.GasUsed)
	assert.True(t, txn.Success)
	assert.Equal(t, "Executed successfully", txn.VmStatus)
	assert.Equal(t, "0x957c214e74b1aded27be7fd78b50c96fc0bfc25a70ad1555a08968a8fdc05cb1", txn.AccumulatorRootHash)
	assert.Empty(t, txn.Changes)
	assert.False(t, txn.BlockEndInfo.BlockGasLimitReached)
	assert.False(t, txn.BlockEndInfo.BlockOutputLimitReached)
	assert.Equal(t, uint64(0), txn.BlockEndInfo.BlockEffectiveBlockGasUnits)
	assert.Equal(t, uint64(3590), txn.BlockEndInfo.BlockApproxOutputSize)

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())

	// Check invalid cases
	_, err = data.UnknownTransaction()
	assert.Error(t, err)
	_, err = data2.UnknownTransaction()
	assert.Error(t, err)
}

func TestTransaction_UnknownTransaction(t *testing.T) {
	testJson := `{
  "version": "2",
  "hash": "0x957c214e74b1aded27be7fd78b50c96fc0bfc25a70ad1555a08968a8fdc05cb1",
  "success": true,
  "type": "block_imaginary_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantUnknown, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	assert.NoError(t, err)
	assert.Equal(t, TransactionVariantUnknown, data2.Type)

	txn, err := data.UnknownTransaction()
	assert.NoError(t, err)
	txn2, err := data2.UnknownTransaction()
	assert.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, "block_imaginary_transaction", txn.Type)
	assert.Equal(t, uint64(2), *txn.TxnVersion())
	assert.Equal(t, "0x957c214e74b1aded27be7fd78b50c96fc0bfc25a70ad1555a08968a8fdc05cb1", txn.TxnHash())
	assert.True(t, *txn.TxnSuccess())

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())

	// Check invalid cases
	_, err = data.GenesisTransaction()
	assert.Error(t, err)
	_, err = data.BlockMetadataTransaction()
	assert.Error(t, err)
	_, err = data.StateCheckpointTransaction()
	assert.Error(t, err)
	_, err = data.BlockEpilogueTransaction()
	assert.Error(t, err)
	_, err = data.UserTransaction()
	assert.Error(t, err)
	_, err = data.ValidatorTransaction()
	assert.Error(t, err)
	_, err = data.PendingTransaction()
	assert.Error(t, err)
	_, err = data2.GenesisTransaction()
	assert.Error(t, err)
	_, err = data2.BlockMetadataTransaction()
	assert.Error(t, err)
	_, err = data2.StateCheckpointTransaction()
	assert.Error(t, err)
	_, err = data2.BlockEpilogueTransaction()
	assert.Error(t, err)
	_, err = data2.UserTransaction()
	assert.Error(t, err)
	_, err = data2.ValidatorTransaction()
	assert.Error(t, err)
}
