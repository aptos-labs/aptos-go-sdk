package aptos

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var testConfig = LocalnetConfig

// createTestClient to use for testing for only one place to configure the network
// TODO: Allow overrides with environment variable?
func createTestClient() (*Client, error) {
	return NewClient(testConfig)
}

func TestSHA3_256Hash(t *testing.T) {
	input := [][]byte{{0x1}, {0x2}, {0x3}}
	expected, err := ParseHex("fd1780a6fc9ee0dab26ceb4b3941ab03e66ccd970d1db91612c66df4515b0a0a")
	assert.NoError(t, err)
	assert.Equal(t, expected, Sha3256Hash(input))
}

func TestParseHex(t *testing.T) {
	// Last case is needed from the JSON API, as an empty array comes out as just 0x
	inputs := []string{"0x012345", "012345", "0x"}
	expected := [][]byte{{0x01, 0x23, 0x45}, {0x01, 0x23, 0x45}, {}}

	for i, input := range inputs {
		val, err := ParseHex(input)
		assert.NoError(t, err)
		assert.Equal(t, expected[i], val)
	}
}

func TestBytesToHex(t *testing.T) {
	inputs := [][]byte{{0x01, 0x23, 0x45}, {0x01}, {}}
	expected := []string{"0x012345", "0x01", "0x"}

	for i, input := range inputs {
		val := BytesToHex(input)
		assert.Equal(t, expected[i], val)
	}
}

func TestStrToUint64(t *testing.T) {
	inputs := []string{"0", "1", "100"}
	expected := []uint64{0, 1, 100}

	for i, input := range inputs {
		val, err := StrToUint64(input)
		assert.NoError(t, err)
		assert.Equal(t, expected[i], val)
	}
}

func TestStrToBigInt(t *testing.T) {
	inputs := []string{"0", "1", "100"}
	expected := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(100)}

	for i, input := range inputs {
		val, err := StrToBigInt(input)
		assert.NoError(t, err)
		assert.Equal(t, expected[i], val)
	}
}

func TestStrToBigIntError(t *testing.T) {
	inputs := []string{"hello", "1a", "", "0.01"}

	for _, input := range inputs {
		_, err := StrToBigInt(input)
		assert.Error(t, err)
	}
}
