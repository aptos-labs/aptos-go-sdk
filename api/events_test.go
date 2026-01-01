package api

import (
	"encoding/json"
	"testing"

	"github.com/qimeila/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent_V1(t *testing.T) {
	t.Parallel()
	testJson := `{
		"type": "0x1::coin::WithdrawEvent",
		"guid": {
            "account_address": "0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b",
		    "creation_number": "3"
		},
		"sequence_number": "0",
		"data": {
			"amount": "1000"
		}
	}`
	data := &Event{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, "0x1::coin::WithdrawEvent", data.Type)
	assert.Equal(t, uint64(0), data.SequenceNumber)
	val, ok := data.Data["amount"].(string)
	require.True(t, ok)
	assert.Equal(t, "1000", val)
	assert.Equal(t, uint64(3), data.Guid.CreationNumber)

	addr := &types.AccountAddress{}
	err = addr.ParseStringRelaxed("0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b")
	require.NoError(t, err)
	assert.Equal(t, addr, data.Guid.AccountAddress)
}

func TestEvent_V2(t *testing.T) {
	t.Parallel()
	testJson := `	{
		"type": "0x1::fungible_asset::Withdraw",
		"guid": {
            "account_address": "0x0",
			"creation_number": "0"
		},
		"sequence_number": "0",
		"data": {
			"store": "0x1234123412341234123412341234123412341234123412341234123412341234",
			"amount": "1000"
		}
	}`
	data := &Event{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)

	assert.Equal(t, "0x1::fungible_asset::Withdraw", data.Type)
	assert.Equal(t, uint64(0), data.SequenceNumber)
	val, ok := data.Data["amount"].(string)
	require.True(t, ok)
	assert.Equal(t, "1000", val)
	val, ok = data.Data["store"].(string)
	require.True(t, ok)
	assert.Equal(t, "0x1234123412341234123412341234123412341234123412341234123412341234", val)
	assert.Equal(t, uint64(0), data.Guid.CreationNumber)
	assert.Equal(t, &types.AccountZero, data.Guid.AccountAddress)
}

func TestEvent_V2_Other(t *testing.T) {
	t.Parallel()
	testJson := `	{
		"type": "vector<u64>",
		"guid": {
            "account_address": "0x0",
			"creation_number": "0"
		},
		"sequence_number": "0",
		"data": ["0","1","2"]
	}`
	data := &Event{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)

	assert.Equal(t, "vector<u64>", data.Type)
	assert.Equal(t, uint64(0), data.SequenceNumber)
	val, ok := data.Data[AnyDataName].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"0", "1", "2"}, val)
	assert.Equal(t, uint64(0), data.Guid.CreationNumber)
	assert.Equal(t, &types.AccountZero, data.Guid.AccountAddress)
}
