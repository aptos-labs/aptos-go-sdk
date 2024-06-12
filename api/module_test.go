package api

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestModule_MoveBytecode(t *testing.T) {
	testJson := `{
		"bytecode": "0xa11ceb0b060000000901000202020403060f0515"
	}`
	data := &MoveBytecode{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	expectedRes, _ := util.ParseHex("0xa11ceb0b060000000901000202020403060f0515")
	assert.Equal(t, HexBytes(expectedRes), data.Bytecode)
}

func TestModule_MoveScript(t *testing.T) {
	testJson := `{
		"bytecode": "0xa11ceb0b060000000901000202020403060f0515"
	}`
	data := &MoveScript{}
	err := json.Unmarshal([]byte(testJson), &data)
	assert.NoError(t, err)
	expectedRes, _ := util.ParseHex("0xa11ceb0b060000000901000202020403060f0515")
	assert.Equal(t, HexBytes(expectedRes), data.Bytecode)
}
