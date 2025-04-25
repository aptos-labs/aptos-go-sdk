package api

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
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
