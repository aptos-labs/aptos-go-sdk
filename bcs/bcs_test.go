package bcs

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

type TestStruct struct {
	num uint8
	b   bool
}

func (st *TestStruct) MarshalBCS(bcs *Serializer) {
	bcs.U8(st.num)
	bcs.Bool(st.b)
}
func (st *TestStruct) UnmarshalBCS(bcs *Deserializer) {
	st.num = bcs.U8()
	st.b = bcs.Bool()
}

func Test_U8(t *testing.T) {
	serialized := []string{"00", "01", "ff"}
	deserialized := []uint8{0, 1, 0xff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint8) {
		serializer.U8(input)
	}, func(deserializer *Deserializer) uint8 {
		return deserializer.U8()
	})
}

func Test_U16(t *testing.T) {
	serialized := []string{"0000", "0100", "ff00", "ffff"}
	deserialized := []uint16{0, 1, 0xff, 0xffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint16) {
		serializer.U16(input)
	}, func(deserializer *Deserializer) uint16 {
		return deserializer.U16()
	})
}

func Test_U32(t *testing.T) {
	serialized := []string{"00000000", "01000000", "ff000000", "ffffffff"}
	deserialized := []uint32{0, 1, 0xff, 0xffffffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint32) {
		serializer.U32(input)
	}, func(deserializer *Deserializer) uint32 {
		return deserializer.U32()
	})
}

func Test_U64(t *testing.T) {
	serialized := []string{"0000000000000000", "0100000000000000", "ff00000000000000", "ffffffffffffffff"}
	deserialized := []uint64{0, 1, 0xff, 0xffffffffffffffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint64) {
		serializer.U64(input)
	}, func(deserializer *Deserializer) uint64 {
		return deserializer.U64()
	})
}

func Test_U128(t *testing.T) {
	serialized := []string{"00000000000000000000000000000000", "01000000000000000000000000000000", "ff000000000000000000000000000000"}
	deserialized := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0xff)}

	helperBigInt(t, serialized, deserialized, func(serializer *Serializer, input *big.Int) {
		serializer.U128(*input)
	}, func(deserializer *Deserializer) big.Int {
		return deserializer.U128()
	})
}

func Test_U256(t *testing.T) {
	serialized := []string{"0000000000000000000000000000000000000000000000000000000000000000", "0100000000000000000000000000000000000000000000000000000000000000", "ff00000000000000000000000000000000000000000000000000000000000000"}
	deserialized := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0xff)}

	helperBigInt(t, serialized, deserialized, func(serializer *Serializer, input *big.Int) {
		serializer.U256(*input)
	}, func(deserializer *Deserializer) big.Int {
		return deserializer.U256()
	})
}

func Test_Uleb128(t *testing.T) {
	serialized := []string{"00", "01", "7f", "ff7f", "ffff03", "ffffffff0f"}
	deserialized := []uint32{0, 1, 127, 16383, 65535, 0xffffffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint32) {
		serializer.Uleb128(input)
	}, func(deserializer *Deserializer) uint32 {
		return deserializer.Uleb128()
	})
}

func Test_Bool(t *testing.T) {
	serialized := []string{"00", "01"}
	deserialized := []bool{false, true}

	helper(t, serialized, deserialized, func(serializer *Serializer, input bool) {
		serializer.Bool(input)
	}, func(deserializer *Deserializer) bool {
		return deserializer.Bool()
	})
}

func Test_String(t *testing.T) {
	serialized := []string{"0461626364", "0568656c6c6f"}
	deserialized := []string{"abcd", "hello"}

	helper(t, serialized, deserialized, func(serializer *Serializer, input string) {
		serializer.WriteString(input)
	}, func(deserializer *Deserializer) string {
		return deserializer.ReadString()
	})
}

func Test_FixedBytes(t *testing.T) {
	serialized := []string{"123456", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}
	deserialized := []string{"123456", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"}

	// Serialize
	for i, input := range deserialized {
		bytes, _ := hex.DecodeString(input)
		serializer := Serializer{}
		expect, _ := hex.DecodeString(deserialized[i])
		serializer.FixedBytes(bytes)
		assert.Equal(t, expect, serializer.ToBytes())
		assert.NoError(t, serializer.Error())
	}

	// Deserialize
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		expect, _ := hex.DecodeString(deserialized[i])
		assert.Equal(t, expect, deserializer.ReadFixedBytes(len(bytes)))
		assert.NoError(t, deserializer.Error())
	}
}

func Test_Bytes(t *testing.T) {
	serialized := []string{"03123456", "2cffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}
	deserialized := []string{"123456", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}

	helper(t, serialized, deserialized, func(serializer *Serializer, input string) {
		bytes, _ := hex.DecodeString(input)
		serializer.WriteBytes(bytes)
	}, func(deserializer *Deserializer) string {
		bytes := deserializer.ReadBytes()
		return hex.EncodeToString(bytes)
	})
}

func Test_Struct(t *testing.T) {
	serialized := []string{"0000", "0001", "FF01"}
	deserialized := []TestStruct{{0, false}, {0, true}, {255, true}}

	// Serializer
	for i, input := range deserialized {
		serializer := &Serializer{}
		serializer.Struct(&input)
		expected, _ := hex.DecodeString(serialized[i])
		assert.Equal(t, expected, serializer.ToBytes())
		assert.NoError(t, serializer.Error())
	}

	// Deserializer
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := &Deserializer{source: bytes}
		st := TestStruct{}
		deserializer.Struct(&st)
		assert.Equal(t, deserialized[i], st)
		assert.NoError(t, deserializer.Error())
	}
}

func helper[TYPE uint8 | uint16 | uint32 | uint64 | bool | []byte | string](t *testing.T, serialized []string, deserialized []TYPE, serialize func(serializer *Serializer, val TYPE), deserialize func(deserializer *Deserializer) TYPE) {

	// Serializer
	for i, input := range deserialized {
		serializer := &Serializer{}
		serialize(serializer, input)
		expected, _ := hex.DecodeString(serialized[i])
		assert.Equal(t, expected, serializer.ToBytes())
		assert.NoError(t, serializer.Error())
	}

	// Deserializer
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := &Deserializer{source: bytes}
		assert.Equal(t, deserialized[i], deserialize(deserializer))
		assert.NoError(t, deserializer.Error())
	}
}

func helperBigInt(t *testing.T, serialized []string, deserialized []*big.Int, serialize func(serializer *Serializer, val *big.Int), deserialize func(deserializer *Deserializer) big.Int) {

	// Serializer
	for i, input := range deserialized {
		serializer := &Serializer{}
		serialize(serializer, input)
		expected, _ := hex.DecodeString(serialized[i])
		bytes := serializer.ToBytes()
		assert.Equal(t, expected, bytes)
		assert.NoError(t, serializer.Error())
	}

	// Deserializer
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := &Deserializer{source: bytes}
		actual := deserialize(deserializer)
		assert.NoError(t, deserializer.Error())
		assert.Equal(t, 0, deserialized[i].Cmp(&actual))
	}
}
