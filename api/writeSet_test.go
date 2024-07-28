package api

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestWriteSet_WriteModule(t *testing.T) {
	testJson := `{
"address": "0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5",
"state_key_hash": "0xa9fd877ba16b362e10efda9410f3e718ae114e567858cbe120732935aceb1f0e",
"data": {
"bytecode": "0xa11ceb0b060000000a01000602060c031235044708054f350784017c088002400ac002090cc9025a0da303020000010101020003080002090402030100010004000100000502010000060201000007030100010a020400020b0601020302020c0107020304020d070102030002070809020300050506050705080503060c03030001060c02060c03010502030303070b01020900090109000901010b01020900090102070b01020900090109000109010a7461626c656d616e6961067369676e6572117461626c655f776974685f6c656e67746804426c616803616464066372656174650664656c6574650672656d6f7665057461626c650f5461626c65576974684c656e6774680a616464726573735f6f6606757073657274036e65770d64657374726f795f656d707479e42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc50000000000000000000000000000000000000000000000000000000000000001000201080b01020303000004010001080b0011042a000f000b010b0238000201000400010d0a001104290020040a0b00380112002d00050c0b000102020004010001060b0011042c001300380202030004010001080b0011042a000f000b0138030102000000",
"abi": {
"address": "0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5",
"name": "tablemania",
"friends": [],
"exposed_functions": [
{
"name": "add",
"visibility": "private",
"is_entry": true,
"is_view": false,
"generic_type_params": [],
"params": [
"&signer",
"u64",
"u64"
],
"return": []
},
{
"name": "create",
"visibility": "private",
"is_entry": true,
"is_view": false,
"generic_type_params": [],
"params": [
"&signer"
],
"return": []
},
{
"name": "delete",
"visibility": "private",
"is_entry": true,
"is_view": false,
"generic_type_params": [],
"params": [
"&signer"
],
"return": []
},
{
"name": "remove",
"visibility": "private",
"is_entry": true,
"is_view": false,
"generic_type_params": [],
"params": [
"&signer",
"u64"
],
"return": []
}
],
"structs": [
{
"name": "Blah",
"is_native": false,
"abilities": [
"key"
],
"generic_type_params": [],
"fields": [
{
"name": "table",
"type": "0x1::table_with_length::TableWithLength<u64, u64>"
}
]
}
]
}
},
"type": "write_module"
}`
	data := &WriteSetChange{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)

	assert.Equal(t, WriteSetChangeVariantWriteModule, data.Type)
	inner := data.Inner.(*WriteSetChangeWriteModule) // TODO: probably make this a bit cleaner with a function
	expectedAddress := &types.AccountAddress{}
	err = expectedAddress.ParseStringRelaxed("0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5")
	assert.NoError(t, err)
	assert.Equal(t, expectedAddress, inner.Address)
	assert.Equal(t, "0xa9fd877ba16b362e10efda9410f3e718ae114e567858cbe120732935aceb1f0e", inner.StateKeyHash)
	assert.NotNil(t, inner.Data) // TODO: more verification
}

func TestWriteSet_WriteResource(t *testing.T) {
	testJson := `{
  "address": "0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5",
  "state_key_hash": "0xa396667bfbfc6af66d8969edfbda02ef9c2f4e4468bf4c71f165a5427afdf6dc",
  "data": {
    "type": "0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5::tablemania::Blah",
    "data": {
      "table": {
        "inner": {
          "handle": "0x18cca5d121ebb854e2f16bd2892d0aad9ae0460e21250bc25daa2cdd6f93a070"
        },
        "length": "0"
      }
    }
  },
  "type": "write_resource"
}`
	data := &WriteSetChange{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)

	assert.Equal(t, WriteSetChangeVariantWriteResource, data.Type)
	inner := data.Inner.(*WriteSetChangeWriteResource) // TODO: probably make this a bit cleaner with a function
	expectedAddress := &types.AccountAddress{}
	err = expectedAddress.ParseStringRelaxed("0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5")
	assert.NoError(t, err)
	assert.Equal(t, expectedAddress, inner.Address)
	assert.Equal(t, "0xa396667bfbfc6af66d8969edfbda02ef9c2f4e4468bf4c71f165a5427afdf6dc", inner.StateKeyHash)
	assert.Equal(t, "0xe42895bdea9ffef448368a95f51b4c883a8e025be3f8e7d08df39f46861a0dc5::tablemania::Blah", inner.Data.Type)
}

func TestWriteSet_DeleteResource(t *testing.T) {
	testJson := `{
      "address": "0x307401f7dd9ca5371ed820070dabaff6cf2196b500c0e359c0e388897987ca6a",
      "state_key_hash": "0x3775f4dbd6900b26cf6c833b112fdfda084f84ef4e562678ca6b54a4791063fd",
      "resource": "0x1::object::ObjectGroup",
      "type": "delete_resource"
    }`
	data := &WriteSetChange{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)

	assert.Equal(t, WriteSetChangeVariantDeleteResource, data.Type)
	inner := data.Inner.(*WriteSetChangeDeleteResource) // TODO: probably make this a bit cleaner with a function
	expectedAddress := &types.AccountAddress{}
	err = expectedAddress.ParseStringRelaxed("0x307401f7dd9ca5371ed820070dabaff6cf2196b500c0e359c0e388897987ca6a")
	assert.NoError(t, err)
	assert.Equal(t, expectedAddress, inner.Address)
	assert.Equal(t, "0x3775f4dbd6900b26cf6c833b112fdfda084f84ef4e562678ca6b54a4791063fd", inner.StateKeyHash)
	assert.Equal(t, "0x1::object::ObjectGroup", inner.Resource)
}

func TestWriteSet_WriteTableItem(t *testing.T) {
	testJson := `{
  "state_key_hash": "0x6e4b28d40f98a106a65163530924c0dcb40c1349d3aa915d108b4d6cfc1ddb19",
  "handle": "0x1b854694ae746cdbd8d44186ca4929b2b337df21d1c74633be19b2710552fdca",
  "key": "0x0619dc29a0aac8fa146714058e8dd6d2d0f3bdf5f6331907bf91f3acd81e6935",
  "value": "0x465192b7fc2a88010000000000000000",
  "data": null,
  "type": "write_table_item"
}`
	data := &WriteSetChange{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)

	assert.Equal(t, WriteSetChangeVariantWriteTableItem, data.Type)
	inner := data.Inner.(*WriteSetChangeWriteTableItem) // TODO: probably make this a bit cleaner with a function
	expectedAddress := &types.AccountAddress{}
	err = expectedAddress.ParseStringRelaxed("0x307401f7dd9ca5371ed820070dabaff6cf2196b500c0e359c0e388897987ca6a")
	assert.NoError(t, err)
	assert.Nil(t, inner.Data)
	assert.Equal(t, "0x6e4b28d40f98a106a65163530924c0dcb40c1349d3aa915d108b4d6cfc1ddb19", inner.StateKeyHash)
	assert.Equal(t, "0x1b854694ae746cdbd8d44186ca4929b2b337df21d1c74633be19b2710552fdca", inner.Handle)
	assert.Equal(t, "0x0619dc29a0aac8fa146714058e8dd6d2d0f3bdf5f6331907bf91f3acd81e6935", inner.Key)
	assert.Equal(t, "0x465192b7fc2a88010000000000000000", inner.Value)
}

func TestWriteSet_DeleteTableItem(t *testing.T) {
	testJson := `{
  "state_key_hash": "0x6b89622e7799dc7c46060ba5941b0d1655c1fc96311f7c6f70f0099f99d467cf",
  "handle": "0x18cca5d121ebb854e2f16bd2892d0aad9ae0460e21250bc25daa2cdd6f93a070",
  "key": "0x0000000000000000",
  "data": null,
  "type": "delete_table_item"
}`
	data := &WriteSetChange{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)

	assert.Equal(t, WriteSetChangeVariantDeleteTableItem, data.Type)
	inner := data.Inner.(*WriteSetChangeDeleteTableItem) // TODO: probably make this a bit cleaner with a function
	expectedAddress := &types.AccountAddress{}
	err = expectedAddress.ParseStringRelaxed("0x307401f7dd9ca5371ed820070dabaff6cf2196b500c0e359c0e388897987ca6a")
	assert.NoError(t, err)
	assert.Nil(t, inner.Data)
	assert.Equal(t, "0x6b89622e7799dc7c46060ba5941b0d1655c1fc96311f7c6f70f0099f99d467cf", inner.StateKeyHash)
	assert.Equal(t, "0x18cca5d121ebb854e2f16bd2892d0aad9ae0460e21250bc25daa2cdd6f93a070", inner.Handle)
	assert.Equal(t, "0x0000000000000000", inner.Key)
}
