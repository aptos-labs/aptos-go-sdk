package aptos

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testConfig = LocalnetConfig

// createTestClient to use for testing for only one place to configure the network
// TODO: Allow overrides with environment variable?
func createTestClient() (*Client, error) {
	return NewClient(testConfig)
}

func TestSHA3_256Hash(t *testing.T) {
	t.Parallel()
	input := [][]byte{{0x1}, {0x2}, {0x3}}
	expected, err := ParseHex("fd1780a6fc9ee0dab26ceb4b3941ab03e66ccd970d1db91612c66df4515b0a0a")
	require.NoError(t, err)
	assert.Equal(t, expected, Sha3256Hash(input))
}

func TestParseHex(t *testing.T) {
	t.Parallel()
	// Last case is needed from the JSON API, as an empty array comes out as just 0x
	inputs := []string{"0x012345", "012345", "0x"}
	expected := [][]byte{{0x01, 0x23, 0x45}, {0x01, 0x23, 0x45}, {}}

	for i, input := range inputs {
		val, err := ParseHex(input)
		require.NoError(t, err)
		assert.Equal(t, expected[i], val)
	}
}

func TestBytesToHex(t *testing.T) {
	t.Parallel()
	inputs := [][]byte{{0x01, 0x23, 0x45}, {0x01}, {}}
	expected := []string{"0x012345", "0x01", "0x"}

	for i, input := range inputs {
		val := BytesToHex(input)
		assert.Equal(t, expected[i], val)
	}
}

func TestStrToUint64(t *testing.T) {
	t.Parallel()
	inputs := []string{"0", "1", "100"}
	expected := []uint64{0, 1, 100}

	for i, input := range inputs {
		val, err := StrToUint64(input)
		require.NoError(t, err)
		assert.Equal(t, expected[i], val)
	}
}

func TestStrToBigInt(t *testing.T) {
	t.Parallel()
	inputs := []string{"0", "1", "100"}
	expected := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(100)}

	for i, input := range inputs {
		val, err := StrToBigInt(input)
		require.NoError(t, err)
		assert.Equal(t, expected[i], val)
	}
}

func TestStrToBigIntError(t *testing.T) {
	t.Parallel()
	inputs := []string{"hello", "1a", "", "0.01"}

	for _, input := range inputs {
		_, err := StrToBigInt(input)
		require.Error(t, err)
	}
}
