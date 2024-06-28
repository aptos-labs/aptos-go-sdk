package api

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	assert.Equal(t, "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80", data.BlockHash)
	assert.Equal(t, uint64(1665609760857472), data.BlockTimestamp)
	assert.Equal(t, uint64(1), data.BlockHeight)
	assert.Equal(t, uint64(1), data.FirstVersion)
	assert.Equal(t, uint64(1), data.LastVersion)
	assert.Empty(t, data.Transactions)
}
