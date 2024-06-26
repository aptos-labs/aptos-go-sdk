package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvent_V1(t *testing.T) {
	testJson := `{
		"type": "0x1::coin::WithdrawEvent",
		"guid": {
			"id": {
				"account_address": "0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b",
				"creation_number": "3"
			}
		},
		"sequence_number": "0",
		"data": {
			"amount": "1000"
		}
	}`
	data := &Event{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	assert.Equal(t, "0x1::coin::WithdrawEvent", data.Type)
	assert.Equal(t, uint64(0), data.SequenceNumber)
	assert.Equal(t, "1000", data.Data["amount"].(string))
	assert.Equal(t, uint64(3), data.Guid.Id.CreationNumber)

	addr := &types.AccountAddress{}
	err = addr.ParseStringRelaxed("0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b")
	assert.NoError(t, err)
	assert.Equal(t, addr, data.Guid.Id.AccountAddress)
}

func TestEvent_V2(t *testing.T) {
	testJson := `	{
		"type": "0x1::fungible_asset::Withdraw",
		"guid": {
			"id": {
				"account_address": "0x0",
				"creation_number": "0"
			}
		},
		"sequence_number": "0",
		"data": {
			"store": "0x1234123412341234123412341234123412341234123412341234123412341234",
			"amount": "1000"
		}
	}`
	data := &Event{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)

	assert.Equal(t, "0x1::fungible_asset::Withdraw", data.Type)
	assert.Equal(t, uint64(0), data.SequenceNumber)
	assert.Equal(t, "1000", data.Data["amount"].(string))
	assert.Equal(t, "0x1234123412341234123412341234123412341234123412341234123412341234", data.Data["store"].(string))
	assert.Equal(t, uint64(0), data.Guid.Id.CreationNumber)
	assert.Equal(t, &types.AccountZero, data.Guid.Id.AccountAddress)
}
