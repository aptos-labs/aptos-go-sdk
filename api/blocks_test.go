package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlock(t *testing.T) {
	testJson := `{
		"block_height": "1",
		"block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
		"block_timestamp": "1665609760857472",
		"first_version": "1",
		"last_version": "1",
		"transactions": null
	}`
	data := &Block{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)

	assert.Equal(t, "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80", data.BlockHash)
	assert.Equal(t, uint64(1665609760857472), data.BlockTimestamp)
	assert.Equal(t, uint64(1), data.BlockHeight)
	assert.Equal(t, uint64(1), data.FirstVersion)
	assert.Equal(t, uint64(1), data.LastVersion)
	assert.Empty(t, data.Transactions)
}

func TestBlockWithNoTransactions(t *testing.T) {
	testJson := `{
		"block_height": "1",
		"block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
		"block_timestamp": "1665609760857472",
		"first_version": "1",
		"last_version": "1",
		"transactions": []
	}`
	data := &Block{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)

	assert.Equal(t, "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80", data.BlockHash)
	assert.Equal(t, uint64(1665609760857472), data.BlockTimestamp)
	assert.Equal(t, uint64(1), data.BlockHeight)
	assert.Equal(t, uint64(1), data.FirstVersion)
	assert.Equal(t, uint64(1), data.LastVersion)
	assert.Empty(t, data.Transactions)
}

func TestBlockWithTransactions(t *testing.T) {
	testJson := `{
		"block_height": "1",
		"block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
		"block_timestamp": "1665609760857472",
		"first_version": "1",
		"last_version": "2",
		"transactions": [
{
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
},
{
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
}
      ]
	}`
	data := &Block{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)

	assert.Equal(t, "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80", data.BlockHash)
	assert.Equal(t, uint64(1665609760857472), data.BlockTimestamp)
	assert.Equal(t, uint64(1), data.BlockHeight)
	assert.Equal(t, uint64(1), data.FirstVersion)
	assert.Equal(t, uint64(2), data.LastVersion)
	assert.NotEmpty(t, data.Transactions)
	assert.Equal(t, uint64(2), data.Transactions[1].Version())
}
