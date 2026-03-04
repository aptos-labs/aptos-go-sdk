package api

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexBytesEncoding(t *testing.T) {
	t.Parallel()

	must := func(s string) HexBytes {
		res, err := util.ParseHex(s)
		require.NoError(t, err)
		return res
	}

	for _, testCase := range []HexBytes{
		must("0x123456"),
		must("2345"),
		must(""),
		{},
	} {
		encoded, err := testCase.MarshalJSON()
		require.NoError(t, err)
		b := new(HexBytes)
		require.NoError(t, b.UnmarshalJSON(encoded))
		require.Equal(t, testCase, *b)
	}
}

func TestU64_UnmarshalJSON_ValidString(t *testing.T) {
	t.Parallel()
	var u U64
	err := u.UnmarshalJSON([]byte(`"12345"`))
	require.NoError(t, err)
	assert.Equal(t, uint64(12345), u.ToUint64())
}

func TestU64_UnmarshalJSON_InvalidString(t *testing.T) {
	t.Parallel()
	var u U64
	err := u.UnmarshalJSON([]byte(`"not_a_number"`))
	require.Error(t, err)
}

func TestU64_UnmarshalJSON_InvalidJSON(t *testing.T) {
	t.Parallel()
	var u U64
	err := u.UnmarshalJSON([]byte(`invalid`))
	require.Error(t, err)
}

func TestHexBytes_UnmarshalJSON_InvalidHex(t *testing.T) {
	t.Parallel()
	var h HexBytes
	err := h.UnmarshalJSON([]byte(`"0xZZZZ"`))
	require.Error(t, err)
}

func TestHexBytes_MarshalJSON_Nil(t *testing.T) {
	t.Parallel()
	var h *HexBytes
	result, err := h.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte("null"), result)
}

func TestGUID_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	data := []byte(`{"creation_number": "42", "account_address": "0x1"}`)
	var g GUID
	err := json.Unmarshal(data, &g)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), g.CreationNumber)
	assert.NotNil(t, g.AccountAddress)
}
