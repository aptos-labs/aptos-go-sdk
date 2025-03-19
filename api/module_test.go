package api

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
)

// TestModule_MoveBytecode tests the MoveBytecode struct
func TestModule_MoveBytecode(t *testing.T) {
	testJSON := `{
		"bytecode": "0xa11ceb0b060000000901000202020403060f0515"
	}`
	data := &MoveBytecode{}
	err := json.Unmarshal([]byte(testJSON), &data)
	assert.NoError(t, err)
	expectedRes, _ := util.ParseHex("0xa11ceb0b060000000901000202020403060f0515")
	assert.Equal(t, HexBytes(expectedRes), data.Bytecode)
}

// TestModule_MoveScript tests the MoveScript struct
func TestModule_MoveScript(t *testing.T) {
	testJSON := `{
		"bytecode": "0xa11ceb0b060000000901000202020403060f0515"
	}`
	data := &MoveScript{}
	err := json.Unmarshal([]byte(testJSON), &data)
	assert.NoError(t, err)
	expectedRes, _ := util.ParseHex("0xa11ceb0b060000000901000202020403060f0515")
	assert.Equal(t, HexBytes(expectedRes), data.Bytecode)
}
