package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteSet_DeleteResource(t *testing.T) {
	testJson := `{
      "address": "0x307401f7dd9ca5371ed820070dabaff6cf2196b500c0e359c0e388897987ca6a",
      "state_key_hash": "0x3775f4dbd6900b26cf6c833b112fdfda084f84ef4e562678ca6b54a4791063fd",
      "resource": "0x1::object::ObjectGroup",
      "type": "delete_resource"
    }`
	data := &WriteSetChangeDeleteResource{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	expectedAddress := &types.AccountAddress{}
	err = expectedAddress.ParseStringRelaxed("0x307401f7dd9ca5371ed820070dabaff6cf2196b500c0e359c0e388897987ca6a")
	assert.NoError(t, err)
	assert.Equal(t, expectedAddress, data.Address)
	assert.Equal(t, "0x3775f4dbd6900b26cf6c833b112fdfda084f84ef4e562678ca6b54a4791063fd", data.StateKeyHash)
	assert.Equal(t, "0x1::object::ObjectGroup", data.Resource)
}
