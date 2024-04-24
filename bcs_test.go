package aptos

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_U8(t *testing.T) {
	inputs := []string{"00", "01", "ff"}
	expected := []uint8{0, 1, 0xff, 0xff}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.U8())
	}
}

func Test_U16(t *testing.T) {
	inputs := []string{"0000", "0100", "ff00", "ffff"}
	expected := []uint16{0, 1, 0xff, 0xffff}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.U16())
	}
}

func Test_U32(t *testing.T) {
	inputs := []string{"00000000", "01000000", "ff000000", "ffffffff"}
	expected := []uint32{0, 1, 0xff, 0xffffffff}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.U32())
	}
}

func Test_U64(t *testing.T) {
	inputs := []string{"0000000000000000", "0100000000000000", "ff00000000000000", "ffffffffffffffff"}
	expected := []uint64{0, 1, 0xff, 0xffffffffffffffff}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.U64())
	}
}

func Test_U128(t *testing.T) {
	// TODO: I don't know Go well enough yet
	inputs := []string{"00000000000000000000000000000000", "01000000000000000000000000000000", "ff000000000000000000000000000000"}
	expected := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0xff)}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.U128())

	}
}

func Test_U256(t *testing.T) {
	// TODO: I don't know Go well enough yet
	inputs := []string{"0000000000000000000000000000000000000000000000000000000000000000", "0100000000000000000000000000000000000000000000000000000000000000", "ff00000000000000000000000000000000000000000000000000000000000000"}
	expected := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0xff)}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, *expected[i], deserializer.U256())
	}
}

func Test_Uleb128(t *testing.T) {
	inputs := []string{"00", "01", "7f", "ff7f", "ffff03"}
	expected := []uint64{0, 1, 127, 16383, 65535}
	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.Uleb128())
	}
}

func Test_Bool(t *testing.T) {
	inputs := []string{"00", "01"}
	expected := []bool{false, true}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.Bool())
	}
}

func Test_String(t *testing.T) {
	inputs := []string{"0461626364", "0568656c6c6f"}
	expected := []string{"abcd", "hello"}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		assert.Equal(t, expected[i], deserializer.ReadString())
	}
}

func Test_FixedBytes(t *testing.T) {
	inputs := []string{"123456", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}
	expected := []string{"123456", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:"}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		expect, _ := hex.DecodeString(expected[i])
		assert.Equal(t, expect, deserializer.ReadFixedBytes(len(bytes)))
	}
}

func Test_Bytes(t *testing.T) {
	inputs := []string{"03123456", "2cffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}
	expected := []string{"123456", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"}

	for i, input := range inputs {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		expect, _ := hex.DecodeString(expected[i])
		assert.Equal(t, expect, deserializer.ReadBytes())
	}
}
