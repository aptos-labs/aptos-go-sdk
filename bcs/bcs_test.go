package bcs

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	num uint8
	b   bool
}

func (st *TestStruct) MarshalBCS(ser *Serializer) {
	ser.U8(st.num)
	ser.Bool(st.b)
}

func (st *TestStruct) UnmarshalBCS(des *Deserializer) {
	st.num = des.U8()
	st.b = des.Bool()
}

type TestStruct2 struct {
	num uint8
	b   bool
}

func (st *TestStruct2) MarshalBCS(ser *Serializer) {
	ser.U8(st.num)
	ser.Bool(st.b)
}

func (st *TestStruct2) UnmarshalBCS(des *Deserializer) {
	st.num = des.U8()
	st.b = des.Bool()
}

type TestStruct3 struct {
	num uint16
}

func (st *TestStruct3) MarshalBCS(ser *Serializer) {
	if st.num > 255 {
		ser.SetError(errors.New("value is greater than 255"))
		return
	}
	ser.U8(uint8(st.num))
}

func (st *TestStruct3) UnmarshalBCS(des *Deserializer) {
	st.num = uint16(des.U8())
}

func Test_U8(t *testing.T) {
	t.Parallel()
	serialized := []string{"00", "01", "ff"}
	deserialized := []uint8{0, 1, 0xff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint8) {
		serializer.U8(input)
	}, func(deserializer *Deserializer) uint8 {
		return deserializer.U8()
	})
}

func Test_U16(t *testing.T) {
	t.Parallel()
	serialized := []string{"0000", "0100", "ff00", "ffff"}
	deserialized := []uint16{0, 1, 0xff, 0xffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint16) {
		serializer.U16(input)
	}, func(deserializer *Deserializer) uint16 {
		return deserializer.U16()
	})
}

func Test_U32(t *testing.T) {
	t.Parallel()
	serialized := []string{"00000000", "01000000", "ff000000", "ffffffff"}
	deserialized := []uint32{0, 1, 0xff, 0xffffffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint32) {
		serializer.U32(input)
	}, func(deserializer *Deserializer) uint32 {
		return deserializer.U32()
	})
}

func Test_U64(t *testing.T) {
	t.Parallel()
	serialized := []string{"0000000000000000", "0100000000000000", "ff00000000000000", "ffffffffffffffff"}
	deserialized := []uint64{0, 1, 0xff, 0xffffffffffffffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint64) {
		serializer.U64(input)
	}, func(deserializer *Deserializer) uint64 {
		return deserializer.U64()
	})
}

func Test_U128(t *testing.T) {
	t.Parallel()
	serialized := []string{"00000000000000000000000000000000", "01000000000000000000000000000000", "ff000000000000000000000000000000"}
	deserialized := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0xff)}

	helperBigInt(t, serialized, deserialized, func(serializer *Serializer, input *big.Int) {
		serializer.U128(*input)
	}, func(deserializer *Deserializer) big.Int {
		return deserializer.U128()
	})
}

func Test_U256(t *testing.T) {
	t.Parallel()
	serialized := []string{"0000000000000000000000000000000000000000000000000000000000000000", "0100000000000000000000000000000000000000000000000000000000000000", "ff00000000000000000000000000000000000000000000000000000000000000"}
	deserialized := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0xff)}

	helperBigInt(t, serialized, deserialized, func(serializer *Serializer, input *big.Int) {
		serializer.U256(*input)
	}, func(deserializer *Deserializer) big.Int {
		return deserializer.U256()
	})
}

func Test_Uleb128(t *testing.T) {
	t.Parallel()
	serialized := []string{"00", "01", "7f", "ff7f", "ffff03", "ffffffff0f"}
	deserialized := []uint32{0, 1, 127, 16383, 65535, 0xffffffff}

	helper(t, serialized, deserialized, func(serializer *Serializer, input uint32) {
		serializer.Uleb128(input)
	}, func(deserializer *Deserializer) uint32 {
		return deserializer.Uleb128()
	})
}

func Test_Bool(t *testing.T) {
	t.Parallel()
	serialized := []string{"00", "01"}
	deserialized := []bool{false, true}

	helper(t, serialized, deserialized, func(serializer *Serializer, input bool) {
		serializer.Bool(input)
	}, func(deserializer *Deserializer) bool {
		return deserializer.Bool()
	})
}

func Test_String(t *testing.T) {
	t.Parallel()
	serialized := []string{"0461626364", "0568656c6c6f"}
	deserialized := []string{"abcd", "hello"}

	helper(t, serialized, deserialized, func(serializer *Serializer, input string) {
		serializer.WriteString(input)
	}, func(deserializer *Deserializer) string {
		return deserializer.ReadString()
	})
}

func Test_FixedBytes(t *testing.T) {
	t.Parallel()
	serialized := []string{"123456", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}
	deserialized := []string{"123456", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"}

	// Serialize
	for i, input := range deserialized {
		bytes, _ := hex.DecodeString(input)
		serializer := Serializer{}
		expect, _ := hex.DecodeString(deserialized[i])
		serializer.FixedBytes(bytes)
		assert.Equal(t, expect, serializer.ToBytes())
		require.NoError(t, serializer.Error())
	}

	// Deserialize
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := Deserializer{source: bytes}
		expect, _ := hex.DecodeString(deserialized[i])
		assert.Equal(t, expect, deserializer.ReadFixedBytes(len(bytes)))
		require.NoError(t, deserializer.Error())
	}
}

func Test_Bytes(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	serialized := []string{"0000", "0001", "FF01"}
	deserialized := []TestStruct{{0, false}, {0, true}, {255, true}}

	// Serializer
	for i, input := range deserialized {
		serializer := &Serializer{}
		serializer.Struct(&input)
		require.NoError(t, serializer.Error())
		expected, err := hex.DecodeString(serialized[i])
		require.NoError(t, err)
		assert.Equal(t, expected, serializer.ToBytes())
	}

	// Deserializer
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := NewDeserializer(bytes)
		st := TestStruct{}
		deserializer.Struct(&st)
		assert.Equal(t, deserialized[i], st)
		require.NoError(t, deserializer.Error())
	}
}

func Test_DeserializeSequence(t *testing.T) {
	t.Parallel()
	deserialized := []TestStruct{{0, false}, {5, true}, {255, true}}
	serialized := []byte{0x03, 0x00, 0x00, 0x05, 0x01, 0xFF, 0x01}

	actualSerialized, err := SerializeSingle(func(ser *Serializer) {
		SerializeSequence(deserialized, ser)
	})
	require.NoError(t, err)
	assert.Equal(t, serialized, actualSerialized)

	des := NewDeserializer(actualSerialized)
	actualDeserialized := DeserializeSequence[TestStruct](des)
	require.NoError(t, des.Error())
	assert.Equal(t, deserialized, actualDeserialized)
}

func Test_InvalidBool(t *testing.T) {
	t.Parallel()
	des := NewDeserializer([]byte{0x02})
	des.Bool()
	require.Error(t, des.Error())
}

func Test_InvalidBytes(t *testing.T) {
	t.Parallel()
	des := NewDeserializer([]byte{0x02})
	des.ReadBytes()
	require.Error(t, des.Error())
}

func Test_InvalidFixedBytesInto(t *testing.T) {
	t.Parallel()
	des := NewDeserializer([]byte{0x02})
	bytes := make([]byte, 2)
	des.ReadFixedBytesInto(bytes)
	require.Error(t, des.Error())
}

func Test_FailedStructSerialize(t *testing.T) {
	t.Parallel()
	str := TestStruct3{
		num: uint16(5),
	}
	_, err := Serialize(&str)
	require.NoError(t, err)
	str.num = uint16(256)
	_, err = Serialize(&str)
	require.Error(t, err)
}

func Test_FailedStructDeserialize(t *testing.T) {
	t.Parallel()
	str := TestStruct{}
	err := Deserialize(&str, []byte{})
	require.Error(t, err)
}

func Test_SerializeSequence(t *testing.T) {
	t.Parallel()
	// Test not implementing Marshal
	ser := Serializer{}
	SerializeSequence([]byte{0x00}, &ser)
	require.Error(t, ser.Error())

	// Test by reference
	testStruct := TestStruct{
		num: 22,
		b:   true,
	}
	data := []TestStruct{testStruct}
	ser = Serializer{}
	SerializeSequence(data, &ser)
	require.NoError(t, ser.Error())
	assert.NotEmpty(t, ser.ToBytes())

	// Test reset
	ser.Reset()
	assert.Empty(t, ser.ToBytes())

	// Test by value
	testStruct2 := TestStruct2{
		num: 52,
		b:   false,
	}
	data2 := []TestStruct2{testStruct2}
	SerializeSequence(data2, &ser)
	require.NoError(t, ser.Error())

	bytes := ser.ToBytes()

	// Test only by self
	onlyBytes, err := SerializeSequenceOnly(data2)
	require.NoError(t, err)
	assert.Equal(t, bytes, onlyBytes)
}

func Test_DeserializeSequenceError(t *testing.T) {
	t.Parallel()
	// Test no leading size byte
	des := NewDeserializer([]byte{})
	DeserializeSequence[TestStruct](des)
	require.Error(t, des.Error())

	// Test no bytes for struct
	des = NewDeserializer([]byte{0x01})
	DeserializeSequence[TestStruct](des)
	require.Error(t, des.Error())

	// Test not a struct type to deserialize
	des = NewDeserializer([]byte{0x01})
	DeserializeSequence[uint8](des)
	require.Error(t, des.Error())
}

func Test_DeserializerErrors(t *testing.T) {
	t.Parallel()
	serialized, _ := hex.DecodeString("000100FF")
	des := NewDeserializer(serialized)
	assert.Equal(t, 4, des.Remaining())
	assert.Equal(t, uint8(0), des.U8())
	assert.Equal(t, 3, des.Remaining())
	assert.Equal(t, uint16(1), des.U16())
	assert.Equal(t, 1, des.Remaining())
	des.U16()
	require.Error(t, des.Error())
	des.SetError(nil)
	assert.Equal(t, uint8(0xff), des.U8())
	require.NoError(t, des.Error())

	des.Bool()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.ReadFixedBytes(2)
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U16()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U32()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U64()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U128()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U256()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U256()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.Uleb128()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.ReadBytes()
	require.Error(t, des.Error())
	des.SetError(nil)
	des.U8()
	require.Error(t, des.Error())
}

func Test_ConvenienceFunctions(t *testing.T) {
	t.Parallel()
	str := TestStruct{
		num: 10,
		b:   true,
	}

	bytes, err := Serialize(&str)
	require.NoError(t, err)

	str2 := TestStruct{}
	err = Deserialize(&str2, bytes)
	require.NoError(t, err)

	assert.Equal(t, str, str2)

	serializedBool, err := SerializeBool(true)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01}, serializedBool)

	serializedU8, err := SerializeU8(1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01}, serializedU8)

	serializedU16, err := SerializeU16(2)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x02, 0x00}, serializedU16)

	serializedU32, err := SerializeU32(3)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x03, 0x00, 0x00, 0x00}, serializedU32)

	serializedU64, err := SerializeU64(4)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, serializedU64)

	serializedU128, err := SerializeU128(*big.NewInt(5))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, serializedU128)

	serializedU256, err := SerializeU256(*big.NewInt(6))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, serializedU256)

	serializedBytes, err := SerializeBytes([]byte{0x05})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x05}, serializedBytes)
}

func Test_SerializeOptional(t *testing.T) {
	t.Parallel()
	ser := Serializer{}
	someValue := uint8(0xFF)
	SerializeOption(&ser, &someValue, func(ser *Serializer, val uint8) {
		ser.U8(val)
	})
	assert.Equal(t, []byte{0x01, 0xFF}, ser.ToBytes())
	des := NewDeserializer(ser.ToBytes())
	desValue := DeserializeOption(des, func(des *Deserializer, out *uint8) {
		*out = des.U8()
	})
	assert.Equal(t, &someValue, desValue)

	ser2 := Serializer{}
	SerializeOption(&ser2, nil, func(ser *Serializer, val uint8) {
		ser.U8(val)
	})
	assert.Equal(t, []byte{0x00}, ser2.ToBytes())

	des2 := NewDeserializer(ser2.ToBytes())
	desValue2 := DeserializeOption(des2, func(des *Deserializer, out *uint8) {
		*out = des.U8()
	})
	assert.Nil(t, desValue2)
}

func Test_NilStructs(t *testing.T) {
	t.Parallel()
	ser := Serializer{}
	ser.Struct(nil)
	require.Error(t, ser.Error())

	des := NewDeserializer([]byte{})
	des.Struct(nil)
	require.Error(t, des.Error())
}

func Test_DeserializeNotEnoughBytes(t *testing.T) {
	t.Parallel()
	data := []byte{0x01, 0x00, 0x00}
	testStruct := &TestStruct{}
	err := Deserialize(testStruct, data)
	require.Error(t, err)
}

func helper[TYPE uint8 | uint16 | uint32 | uint64 | bool | []byte | string](t *testing.T, serialized []string, deserialized []TYPE, serialize func(serializer *Serializer, val TYPE), deserialize func(deserializer *Deserializer) TYPE) {
	t.Helper()
	// Serializer
	for i, input := range deserialized {
		serializer := &Serializer{}
		serialize(serializer, input)
		expected, _ := hex.DecodeString(serialized[i])
		assert.Equal(t, expected, serializer.ToBytes())
		require.NoError(t, serializer.Error())
	}

	// Deserializer
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := NewDeserializer(bytes)
		assert.Equal(t, deserialized[i], deserialize(deserializer))
		require.NoError(t, deserializer.Error())
	}
}

func helperBigInt(t *testing.T, serialized []string, deserialized []*big.Int, serialize func(serializer *Serializer, val *big.Int), deserialize func(deserializer *Deserializer) big.Int) {
	t.Helper()
	// Serializer
	for i, input := range deserialized {
		serializer := &Serializer{}
		serialize(serializer, input)
		expected, _ := hex.DecodeString(serialized[i])
		bytes := serializer.ToBytes()
		assert.Equal(t, expected, bytes)
		require.NoError(t, serializer.Error())
	}

	// Deserializer
	for i, input := range serialized {
		bytes, _ := hex.DecodeString(input)
		deserializer := NewDeserializer(bytes)
		actual := deserialize(deserializer)
		require.NoError(t, deserializer.Error())
		assert.Equal(t, 0, deserialized[i].Cmp(&actual))
	}
}
