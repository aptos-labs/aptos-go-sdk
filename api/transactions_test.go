package api

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransaction_GenesisTransaction(t *testing.T) {
	t.Parallel()
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
      "type": "direct_write_set"
    },
    "type": "write_set_payload"
  },
  "type": "genesis_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantGenesis, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantGenesis, data2.Type)

	txn, err := data.GenesisTransaction()
	require.NoError(t, err)
	txn2, err := data2.GenesisTransaction()
	require.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(0), txn.Version)

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}

func TestTransaction_PendingTransaction(t *testing.T) {
	t.Parallel()
	testJson := `{
  "hash": "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69",
  "state_checkpoint_hash": null,
  "sender": "0xa46c6c7a65d605685e23055a6a906fb7284ba87849cbeb579d5c07424938241e",
  "sequence_number": "242217",
  "max_gas_amount": "2018",
  "gas_unit_price": "100",
  "expiration_timestamp_secs": "1719968695",
  "payload": {
    "function": "0x1::object::transfer",
    "type_arguments": [
      "0x4::token::Token"
    ],
    "arguments": [
    {
      "inner": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
    },
    "0x8038df5e61a19a5f86ad01f4389736b08250dad1b4aa864afc4fc639a2581ca8"
    ],
    "type": "entry_function_payload"
  },
  "signature": {
    "public_key": "0x5e10e3db4e3c700142b9a3e18c40038db5903f2dedfe41d09aca74a8c68565d6",
    "signature": "0xa95686dab2c93cf1720e300b929e3656cc6cdc3a8389dc12bb9bd5a17ae3af975bee9d618f080266e3a60f1e2968220a83d773e2b3902edfe54127ed0a7b290b",
    "type": "ed25519_signature"
  },
  "type": "pending_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantPending, data.Type)

	// Pending isn't committed
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.Error(t, err)

	txn, err := data.PendingTransaction()
	require.NoError(t, err)

	assert.Equal(t, "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69", txn.Hash)
	assert.Equal(t, uint64(242217), txn.SequenceNumber)
	assert.Equal(t, uint64(100), txn.GasUnitPrice)
	assert.Equal(t, uint64(2018), txn.MaxGasAmount)
	assert.Equal(t, uint64(1719968695), txn.ExpirationTimestampSecs)

	// TODO: test some more

	// Check functions
	assert.Nil(t, data.Version())
	assert.Equal(t, "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69", data.Hash())
	assert.Nil(t, data.Success())
}

func TestTransaction_UserTransaction(t *testing.T) {
	t.Parallel()
	testJson := `{
  "version": "1010733903",
  "hash": "0xae3f1f751c6cacd61f46054a5e9e39ca9f094802875befbc54ceecbcdf6eff69",
  "state_change_hash": "0x3e8340786d2085a2160fa368c380ed412d4a5a3c5ccad692092c4bc0074fde3e",
  "event_root_hash": "0xe6e2ae41a57d9ab1c7dc58851d7beb4d5be43797ba7225d3e2a3b69c35fe7c2d",
  "state_checkpoint_hash": null,
  "gas_used": "5",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0xf9fdaddf6051311cb54e3756a343faa346f1c9137370762f6eef8e375a7031bb",
  "changes": [
  {
    "address": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
    "state_key_hash": "0xb59b40ac86b159eee2c76ff2eb121b91aa8638ef806d08ed6e061bd60c9b134d",
    "data": {
      "type": "0x1::object::ObjectCore",
      "data": {
        "allow_ungated_transfer": true,
        "guid_creation_num": "1125899906842626",
        "owner": "0x8038df5e61a19a5f86ad01f4389736b08250dad1b4aa864afc4fc639a2581ca8",
        "transfer_events": {
          "counter": "1",
          "guid": {
            "id": {
              "addr": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
              "creation_num": "1125899906842624"
            }
          }
        }
      }
    },
    "type": "write_resource"
  },
  {
    "address": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
    "state_key_hash": "0xb59b40ac86b159eee2c76ff2eb121b91aa8638ef806d08ed6e061bd60c9b134d",
    "data": {
      "type": "0x4::aptos_token::AptosToken",
      "data": {
        "burn_ref": {
          "vec": [
          {
            "inner": {
              "vec": [
              {
                "self": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
              }
              ]
            },
            "self": {
              "vec": []
            }
          }
          ]
        },
        "mutator_ref": {
          "vec": [
          {
            "self": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
          }
          ]
        },
        "property_mutator_ref": {
          "self": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
        },
        "transfer_ref": {
          "vec": [
          {
            "self": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
          }
          ]
        }
      }
    },
    "type": "write_resource"
  },
  {
    "address": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
    "state_key_hash": "0xb59b40ac86b159eee2c76ff2eb121b91aa8638ef806d08ed6e061bd60c9b134d",
    "data": {
      "type": "0x4::property_map::PropertyMap",
      "data": {
        "inner": {
	      "data": []
        }
      }
    },
    "type": "write_resource"
  },
  {
    "address": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
    "state_key_hash": "0xb59b40ac86b159eee2c76ff2eb121b91aa8638ef806d08ed6e061bd60c9b134d",
    "data": {
      "type": "0x4::token::Token",
      "data": {
        "collection": {
          "inner": "0x778adb39026a14009cf5aa93eb53d81299e40c7a8dbcdbf7b490cbc29749d259"
        },
        "description": "This is BLACK FLAG ARMY NFT",
        "index": "0",
        "mutation_events": {
          "counter": "0",
          "guid": {
            "id": {
              "addr": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
              "creation_num": "1125899906842625"
            }
          }
        },
        "name": "",
        "uri": "https://bafybeierhssqdg7fv64xkkjuvsq4bikj2yfmuxm4dvb6jxb2un4yw37ohi.ipfs.w3s.link/68.webp"
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
    "type_arguments": [
      "0x4::token::Token"
    ],
    "arguments": [
    {
      "inner": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
    },
    "0x8038df5e61a19a5f86ad01f4389736b08250dad1b4aa864afc4fc639a2581ca8"
    ],
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
      "creation_number": "1125899906842624",
      "account_address": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572"
    },
    "sequence_number": "0",
    "type": "0x1::object::TransferEvent",
    "data": {
      "from": "0xa46c6c7a65d605685e23055a6a906fb7284ba87849cbeb579d5c07424938241e",
      "object": "0x2932a152328163661f0ae591911270d0edfe0a765beb48a270b9b8a70e766572",
      "to": "0x8038df5e61a19a5f86ad01f4389736b08250dad1b4aa864afc4fc639a2581ca8"
    }
  },
  {
    "guid": {
      "creation_number": "0",
      "account_address": "0x0"
    },
    "sequence_number": "0",
    "type": "0x1::transaction_fee::FeeStatement",
    "data": {
      "execution_gas_units": "3",
      "io_gas_units": "2",
      "storage_fee_octas": "0",
      "storage_fee_refund_octas": "0",
      "total_charge_gas_units": "5"
    }
  }
  ],
  "timestamp": "1719965096135309",
  "type": "user_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantUser, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantUser, data2.Type)

	txn, err := data.UserTransaction()
	require.NoError(t, err)
	txn2, err := data2.UserTransaction()
	require.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(1010733903), txn.Version)
	assert.Equal(t, uint64(1719965096135309), txn.Timestamp)
	assert.Equal(t, uint64(242217), txn.SequenceNumber)
	assert.Equal(t, uint64(100), txn.GasUnitPrice)
	assert.Equal(t, uint64(2018), txn.MaxGasAmount)
	assert.Equal(t, uint64(1719968695), txn.ExpirationTimestampSecs)
	assert.Nil(t, txn.ReplayProtectionNonce)

	// TODO: test some more

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}

func TestTransaction_BlockMetadataTransaction(t *testing.T) {
	t.Parallel()
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
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockMetadata, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockMetadata, data2.Type)

	txn, err := data.BlockMetadataTransaction()
	require.NoError(t, err)
	txn2, err := data2.BlockMetadataTransaction()
	require.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(1), txn.Version)
	assert.Equal(t, "0x81f7099ac9f45238ed4a98275add46f4da0a35ff62be0537846ca3d7c52bfbfc", txn.Id)
	assert.Equal(t, uint64(1), txn.Epoch)
	assert.Equal(t, uint64(1), txn.Round)
	assert.Equal(t, uint64(1719520421743738), txn.Timestamp)

	address := &types.AccountAddress{}
	err = address.ParseStringRelaxed("0x90693588b138a37dbb37cb96c42ffb02bf48611fc9e78adeb57c8708ee3ac03e")
	require.NoError(t, err)
	assert.Equal(t, address, txn.Proposer)
	assert.Equal(t, []uint32{1, 2}, txn.FailedProposerIndices)
	assert.Equal(t, []uint8{0}, txn.PreviousBlockVotesBitvec)

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}

func TestTransaction_StateCheckpointTransaction(t *testing.T) {
	t.Parallel()
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
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantStateCheckpoint, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantStateCheckpoint, data2.Type)

	txn, err := data.StateCheckpointTransaction()
	require.NoError(t, err)
	txn2, err := data2.StateCheckpointTransaction()
	require.NoError(t, err)
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
	t.Parallel()
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
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockEpilogue, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantBlockEpilogue, data2.Type)

	txn, err := data.BlockEpilogueTransaction()
	require.NoError(t, err)
	txn2, err := data2.BlockEpilogueTransaction()
	require.NoError(t, err)
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
	require.Error(t, err)
	_, err = data2.UnknownTransaction()
	require.Error(t, err)
}

func TestTransaction_UnknownTransaction(t *testing.T) {
	t.Parallel()
	testJson := `{
  "version": "2",
  "hash": "0x957c214e74b1aded27be7fd78b50c96fc0bfc25a70ad1555a08968a8fdc05cb1",
  "success": true,
  "type": "block_imaginary_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantUnknown, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantUnknown, data2.Type)

	txn, err := data.UnknownTransaction()
	require.NoError(t, err)
	txn2, err := data2.UnknownTransaction()
	require.NoError(t, err)
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
	require.Error(t, err)
	_, err = data.BlockMetadataTransaction()
	require.Error(t, err)
	_, err = data.StateCheckpointTransaction()
	require.Error(t, err)
	_, err = data.BlockEpilogueTransaction()
	require.Error(t, err)
	_, err = data.UserTransaction()
	require.Error(t, err)
	_, err = data.ValidatorTransaction()
	require.Error(t, err)
	_, err = data.PendingTransaction()
	require.Error(t, err)
	_, err = data2.GenesisTransaction()
	require.Error(t, err)
	_, err = data2.BlockMetadataTransaction()
	require.Error(t, err)
	_, err = data2.StateCheckpointTransaction()
	require.Error(t, err)
	_, err = data2.BlockEpilogueTransaction()
	require.Error(t, err)
	_, err = data2.UserTransaction()
	require.Error(t, err)
	_, err = data2.ValidatorTransaction()
	require.Error(t, err)
}

func TestTransaction_WithReplayNonce(t *testing.T) {
	t.Parallel()
	testJson := `{
  "version": "6781425728",
  "hash": "0x650da387b6e7de0e9a83ca5f7c12e41ef865d0d12c38e793b55091da7ec90b39",
  "state_change_hash": "0x5b30b4306a1cbebe71dd319a02b84a78305f1a228de00f965c0e3350220e9f95",
  "event_root_hash": "0x5d3d8778f6d08e1668d9c2827841e63402531f8696a4df5b9bd36954ff048799",
  "state_checkpoint_hash": null,
  "gas_used": "57",
  "success": true,
  "vm_status": "Executed successfully",
  "accumulator_root_hash": "0x24585b5a178b55a7f899522e6101a6b2bb65baf3e33bdc4215034a1c463c13bb",
  "changes": [
    {
      "address": "0xa",
      "state_key_hash": "0x1db5441d8fa4229c5844f73fd66da4ad8176cb8793d8b3a7f6ca858722030043",
      "data": {
        "type": "0x1::coin::PairedCoinType",
        "data": {
          "type": {
            "account_address": "0x1",
            "module_name": "0x6170746f735f636f696e",
            "struct_name": "0x4170746f73436f696e"
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa",
      "state_key_hash": "0x1db5441d8fa4229c5844f73fd66da4ad8176cb8793d8b3a7f6ca858722030043",
      "data": {
        "type": "0x1::coin::PairedFungibleAssetRefs",
        "data": {
          "burn_ref_opt": {
            "vec": [
              {
                "metadata": {
                  "inner": "0xa"
                }
              }
            ]
          },
          "mint_ref_opt": {
            "vec": [
              {
                "metadata": {
                  "inner": "0xa"
                }
              }
            ]
          },
          "transfer_ref_opt": {
            "vec": [
              {
                "metadata": {
                  "inner": "0xa"
                }
              }
            ]
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa",
      "state_key_hash": "0x1db5441d8fa4229c5844f73fd66da4ad8176cb8793d8b3a7f6ca858722030043",
      "data": {
        "type": "0x1::fungible_asset::ConcurrentSupply",
        "data": {
          "current": {
            "max_value": "340282366920938463463374607431768211455",
            "value": "29015120094410588863"
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa",
      "state_key_hash": "0x1db5441d8fa4229c5844f73fd66da4ad8176cb8793d8b3a7f6ca858722030043",
      "data": {
        "type": "0x1::fungible_asset::Metadata",
        "data": {
          "decimals": 8,
          "icon_uri": "",
          "name": "Aptos Coin",
          "project_uri": "",
          "symbol": "APT"
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa",
      "state_key_hash": "0x1db5441d8fa4229c5844f73fd66da4ad8176cb8793d8b3a7f6ca858722030043",
      "data": {
        "type": "0x1::object::ObjectCore",
        "data": {
          "allow_ungated_transfer": true,
          "guid_creation_num": "1125899906842625",
          "owner": "0x1",
          "transfer_events": {
            "counter": "0",
            "guid": {
              "id": {
                "addr": "0xa",
                "creation_num": "1125899906842624"
              }
            }
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa",
      "state_key_hash": "0x1db5441d8fa4229c5844f73fd66da4ad8176cb8793d8b3a7f6ca858722030043",
      "data": {
        "type": "0x1::primary_fungible_store::DeriveRefPod",
        "data": {
          "metadata_derive_ref": {
            "self": "0xa"
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0x4f8733c98d484ea506fcfd7ecdec17c4b27f7489fb0c8962975bf0cf213b6242",
      "state_key_hash": "0x0002c250a4d5c3d717c489aec95f5989b6f09af282473fb4ad604ee02bbc0af8",
      "data": {
        "type": "0x1::fungible_asset::FungibleStore",
        "data": {
          "balance": "1994000",
          "frozen": false,
          "metadata": {
            "inner": "0xa"
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0x4f8733c98d484ea506fcfd7ecdec17c4b27f7489fb0c8962975bf0cf213b6242",
      "state_key_hash": "0x0002c250a4d5c3d717c489aec95f5989b6f09af282473fb4ad604ee02bbc0af8",
      "data": {
        "type": "0x1::object::ObjectCore",
        "data": {
          "allow_ungated_transfer": false,
          "guid_creation_num": "1125899906842625",
          "owner": "0x49ffe7968750a5ffea80af6fd7657bb246ff8ce6657cbf2c8ed13d9276096b3f",
          "transfer_events": {
            "counter": "0",
            "guid": {
              "id": {
                "addr": "0x4f8733c98d484ea506fcfd7ecdec17c4b27f7489fb0c8962975bf0cf213b6242",
                "creation_num": "1125899906842624"
              }
            }
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa66d5588e5e71987999dea776e5798be8926470f012e4a4f320e0737903ecba1",
      "state_key_hash": "0xc1c56630c82dbd34129a5a5347df7ecfb713e9a665407234e3bbfe94117c4e92",
      "data": {
        "type": "0x1::fungible_asset::FungibleStore",
        "data": {
          "balance": "29999948600",
          "frozen": false,
          "metadata": {
            "inner": "0xa"
          }
        }
      },
      "type": "write_resource"
    },
    {
      "address": "0xa66d5588e5e71987999dea776e5798be8926470f012e4a4f320e0737903ecba1",
      "state_key_hash": "0xc1c56630c82dbd34129a5a5347df7ecfb713e9a665407234e3bbfe94117c4e92",
      "data": {
        "type": "0x1::object::ObjectCore",
        "data": {
          "allow_ungated_transfer": false,
          "guid_creation_num": "1125899906842625",
          "owner": "0xa5b2dd0b5fe37a06f151d424cc880f67458c6fe3d8dbf44a55c6021cd111c10d",
          "transfer_events": {
            "counter": "0",
            "guid": {
              "id": {
                "addr": "0xa66d5588e5e71987999dea776e5798be8926470f012e4a4f320e0737903ecba1",
                "creation_num": "1125899906842624"
              }
            }
          }
        }
      },
      "type": "write_resource"
    },
    {
      "state_key_hash": "0x6e4b28d40f98a106a65163530924c0dcb40c1349d3aa915d108b4d6cfc1ddb19",
      "handle": "0x1b854694ae746cdbd8d44186ca4929b2b337df21d1c74633be19b2710552fdca",
      "key": "0x0619dc29a0aac8fa146714058e8dd6d2d0f3bdf5f6331907bf91f3acd81e6935",
      "value": "0xd6b01fdfed0ebb020000000000000000",
      "data": null,
      "type": "write_table_item"
    },
    {
      "state_key_hash": "0xae1988c58918023c1d5e3c83860a8fd21958557059c62a9fb6348815aa358b95",
      "handle": "0x4f68492eeffc337d93969722200c8ebf28d80eb30f5dee5c54981012da31f0a4",
      "key": "0x2627000000000000",
      "value": "0x00000100017492476800000000a5b2dd0b5fe37a06f151d424cc880f67458c6fe3d8dbf44a55c6021cd111c10de0e1f6016d2ed4c001010000000000000000000000000000000000012b52265feda0da555eb84a3e43894602937237ce11a4a323aa8c73b41d04204005000000000000000f00000000000000010e00000000000000050000000100000000000000010000000000000001550053000000010001a5b2dd0b5fe37a06f151d424cc880f67458c6fe3d8dbf44a55c6021cd111c10de0e1f6016d2ed4c0017492476800000000000000000000000000000000000000000001e532644d51f5b1be0bf1b1d41405b0199307f80c463d1fbc9263c2b804a9b13d05000000000000000f00000000000000010e0000000000000005000000010000000000000001000000000000000166005500",
      "data": null,
      "type": "write_table_item"
    }
  ],
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
  "events": [
    {
      "guid": {
        "creation_number": "0",
        "account_address": "0x0"
      },
      "sequence_number": "0",
      "type": "0x1::fungible_asset::Withdraw",
      "data": {
        "amount": "20000",
        "store": "0xa66d5588e5e71987999dea776e5798be8926470f012e4a4f320e0737903ecba1"
      }
    },
    {
      "guid": {
        "creation_number": "0",
        "account_address": "0x0"
      },
      "sequence_number": "0",
      "type": "0x1::fungible_asset::Deposit",
      "data": {
        "amount": "20000",
        "store": "0x4f8733c98d484ea506fcfd7ecdec17c4b27f7489fb0c8962975bf0cf213b6242"
      }
    },
    {
      "guid": {
        "creation_number": "0",
        "account_address": "0x0"
      },
      "sequence_number": "0",
      "type": "0x1::transaction_fee::FeeStatement",
      "data": {
        "execution_gas_units": "7",
        "io_gas_units": "11",
        "storage_fee_octas": "3960",
        "storage_fee_refund_octas": "0",
        "total_charge_gas_units": "57"
      }
    }
  ],
  "timestamp": "1749520982956875",
  "type": "user_transaction"
}`
	data := &Transaction{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantUser, data.Type)
	data2 := &CommittedTransaction{}
	err = json.Unmarshal([]byte(testJson), &data2)
	require.NoError(t, err)
	assert.Equal(t, TransactionVariantUser, data2.Type)

	txn, err := data.UserTransaction()
	require.NoError(t, err)
	txn2, err := data2.UserTransaction()
	require.NoError(t, err)
	assert.Equal(t, txn, txn2)

	assert.Equal(t, uint64(6781425728), txn.Version)
	assert.Equal(t, uint64(1749520982956875), txn.Timestamp)
	assert.Equal(t, uint64(0xffffffffffffffff), txn.SequenceNumber)
	assert.Equal(t, uint64(100), txn.GasUnitPrice)
	assert.Equal(t, uint64(85), txn.MaxGasAmount)
	assert.Equal(t, uint64(1749521012), txn.ExpirationTimestampSecs)
	assert.NotNil(t, txn.ReplayProtectionNonce)
	assert.Equal(t, uint64(13894781796064092640), *txn.ReplayProtectionNonce)

	// TODO: test some more

	// Check functions
	assert.Equal(t, *data.Version(), data2.Version())
	assert.Equal(t, data.Hash(), data2.Hash())
	assert.Equal(t, *data.Success(), data2.Success())
}
