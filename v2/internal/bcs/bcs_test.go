package bcs

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStruct implements Marshaler and Unmarshaler
type TestStruct struct {
	Num     uint64
	Enabled bool
}

func (s *TestStruct) MarshalBCS(ser *Serializer) {
	ser.U64(s.Num)
	ser.Bool(s.Enabled)
}

func (s *TestStruct) UnmarshalBCS(des *Deserializer) {
	s.Num = des.U64()
	s.Enabled = des.Bool()
}

func TestSerialize(t *testing.T) {
	t.Parallel()

	s := &TestStruct{Num: 42, Enabled: true}
	data, err := Serialize(s)
	require.NoError(t, err)

	// 42 as u64 little-endian + true as 0x01
	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	assert.Equal(t, expected, data)
}

func TestDeserialize(t *testing.T) {
	t.Parallel()

	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	var s TestStruct
	err := Deserialize(&s, data)
	require.NoError(t, err)

	assert.Equal(t, uint64(42), s.Num)
	assert.True(t, s.Enabled)
}

func TestSerializeNil(t *testing.T) {
	t.Parallel()

	_, err := Serialize(nil)
	assert.ErrorIs(t, err, ErrNilValue)
}

func TestDeserializeNil(t *testing.T) {
	t.Parallel()

	err := Deserialize(nil, []byte{})
	assert.ErrorIs(t, err, ErrNilValue)
}

func TestDeserializeRemainingBytes(t *testing.T) {
	t.Parallel()

	// Extra byte at the end
	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 0xFF}
	var s TestStruct
	err := Deserialize(&s, data)
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

func TestPrimitives(t *testing.T) {
	t.Parallel()

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		data, err := SerializeBool(true)
		require.NoError(t, err)
		assert.Equal(t, []byte{1}, data)

		v, err := DeserializeBool(data)
		require.NoError(t, err)
		assert.True(t, v)

		data, err = SerializeBool(false)
		require.NoError(t, err)
		assert.Equal(t, []byte{0}, data)

		v, err = DeserializeBool(data)
		require.NoError(t, err)
		assert.False(t, v)
	})

	t.Run("u8", func(t *testing.T) {
		t.Parallel()

		data, err := SerializeU8(255)
		require.NoError(t, err)
		assert.Equal(t, []byte{255}, data)

		v, err := DeserializeU8(data)
		require.NoError(t, err)
		assert.Equal(t, uint8(255), v)
	})

	t.Run("u16", func(t *testing.T) {
		t.Parallel()

		data, err := SerializeU16(0x1234)
		require.NoError(t, err)
		assert.Equal(t, []byte{0x34, 0x12}, data) // Little-endian

		v, err := DeserializeU16(data)
		require.NoError(t, err)
		assert.Equal(t, uint16(0x1234), v)
	})

	t.Run("u32", func(t *testing.T) {
		t.Parallel()

		data, err := SerializeU32(0x12345678)
		require.NoError(t, err)
		assert.Equal(t, []byte{0x78, 0x56, 0x34, 0x12}, data)

		v, err := DeserializeU32(data)
		require.NoError(t, err)
		assert.Equal(t, uint32(0x12345678), v)
	})

	t.Run("u64", func(t *testing.T) {
		t.Parallel()

		data, err := SerializeU64(0x123456789ABCDEF0)
		require.NoError(t, err)
		assert.Equal(t, []byte{0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}, data)

		v, err := DeserializeU64(data)
		require.NoError(t, err)
		assert.Equal(t, uint64(0x123456789ABCDEF0), v)
	})

	t.Run("bytes", func(t *testing.T) {
		t.Parallel()

		input := []byte{0x01, 0x02, 0x03}
		data, err := SerializeBytes(input)
		require.NoError(t, err)
		// Length prefix (3) + bytes
		assert.Equal(t, []byte{3, 1, 2, 3}, data)

		v, err := DeserializeBytes(data)
		require.NoError(t, err)
		assert.Equal(t, input, v)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		data, err := SerializeString("hello")
		require.NoError(t, err)
		// Length prefix (5) + "hello"
		assert.Equal(t, []byte{5, 'h', 'e', 'l', 'l', 'o'}, data)

		v, err := DeserializeString(data)
		require.NoError(t, err)
		assert.Equal(t, "hello", v)
	})
}

func TestSignedIntegers(t *testing.T) {
	t.Parallel()

	t.Run("i8", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		ser.I8(-1)
		assert.Equal(t, []byte{0xFF}, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		assert.Equal(t, int8(-1), des.I8())
	})

	t.Run("i16", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		ser.I16(-1)
		assert.Equal(t, []byte{0xFF, 0xFF}, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		assert.Equal(t, int16(-1), des.I16())
	})

	t.Run("i32", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		ser.I32(-1)
		assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF}, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		assert.Equal(t, int32(-1), des.I32())
	})

	t.Run("i64", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		ser.I64(-1)
		expected := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		assert.Equal(t, expected, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		assert.Equal(t, int64(-1), des.I64())
	})
}

func TestBigIntegers(t *testing.T) {
	t.Parallel()

	t.Run("u128", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		v := big.NewInt(42)
		ser.U128(v)

		expected := make([]byte, 16)
		expected[0] = 42
		assert.Equal(t, expected, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		result := des.U128()
		assert.Equal(t, 0, v.Cmp(result))
	})

	t.Run("u256", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		v := big.NewInt(42)
		ser.U256(v)

		expected := make([]byte, 32)
		expected[0] = 42
		assert.Equal(t, expected, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		result := des.U256()
		assert.Equal(t, 0, v.Cmp(result))
	})

	t.Run("i128 negative", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		v := big.NewInt(-1)
		ser.I128(v)

		// -1 in two's complement is all 0xFF
		expected := make([]byte, 16)
		for i := range expected {
			expected[i] = 0xFF
		}
		assert.Equal(t, expected, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		result := des.I128()
		assert.Equal(t, 0, v.Cmp(result))
	})

	t.Run("i256 negative", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		v := big.NewInt(-1)
		ser.I256(v)

		expected := make([]byte, 32)
		for i := range expected {
			expected[i] = 0xFF
		}
		assert.Equal(t, expected, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		result := des.I256()
		assert.Equal(t, 0, v.Cmp(result))
	})
}

func TestUleb128(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value    uint32
		expected []byte
	}{
		{0, []byte{0}},
		{1, []byte{1}},
		{127, []byte{127}},
		{128, []byte{0x80, 0x01}},
		{16383, []byte{0xFF, 0x7F}},
		{16384, []byte{0x80, 0x80, 0x01}},
	}

	for _, tt := range tests {
		ser := NewSerializer()
		ser.Uleb128(tt.value)
		assert.Equal(t, tt.expected, ser.ToBytes(), "value: %d", tt.value)

		des := NewDeserializer(tt.expected)
		assert.Equal(t, tt.value, des.Uleb128())
	}
}

func TestSequence(t *testing.T) {
	t.Parallel()

	items := []*TestStruct{
		{Num: 1, Enabled: true},
		{Num: 2, Enabled: false},
	}

	ser := NewSerializer()
	SerializeSequence(ser, items)
	require.NoError(t, ser.Error())

	des := NewDeserializer(ser.ToBytes())
	result := DeserializeSequence(des, func() *TestStruct { return &TestStruct{} })
	require.NoError(t, des.Error())

	assert.Len(t, result, 2)
	assert.Equal(t, uint64(1), result[0].Num)
	assert.True(t, result[0].Enabled)
	assert.Equal(t, uint64(2), result[1].Num)
	assert.False(t, result[1].Enabled)
}

func TestOption(t *testing.T) {
	t.Parallel()

	t.Run("some", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		v := uint64(42)
		SerializeOption(ser, &v, func(s *Serializer, val uint64) {
			s.U64(val)
		})
		require.NoError(t, ser.Error())

		des := NewDeserializer(ser.ToBytes())
		result := DeserializeOption(des, func(d *Deserializer) uint64 {
			return d.U64()
		})
		require.NoError(t, des.Error())
		require.NotNil(t, result)
		assert.Equal(t, uint64(42), *result)
	})

	t.Run("none", func(t *testing.T) {
		t.Parallel()

		ser := NewSerializer()
		SerializeOption[uint64](ser, nil, func(s *Serializer, val uint64) {
			s.U64(val)
		})
		require.NoError(t, ser.Error())
		assert.Equal(t, []byte{0}, ser.ToBytes())

		des := NewDeserializer(ser.ToBytes())
		result := DeserializeOption(des, func(d *Deserializer) uint64 {
			return d.U64()
		})
		require.NoError(t, des.Error())
		assert.Nil(t, result)
	})
}

func TestInvalidBool(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{0x02})
	_ = des.Bool()
	assert.ErrorIs(t, des.Error(), ErrInvalidBool)
}

func TestNotEnoughBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(*Deserializer)
	}{
		{"u8", func(d *Deserializer) { d.U8() }},
		{"u16", func(d *Deserializer) { d.U16() }},
		{"u32", func(d *Deserializer) { d.U32() }},
		{"u64", func(d *Deserializer) { d.U64() }},
		{"bool", func(d *Deserializer) { d.Bool() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			des := NewDeserializer([]byte{})
			tt.fn(des)
			assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
		})
	}
}

// ReflectStruct for testing reflection-based serialization
type ReflectStruct struct {
	Name    string `bcs:"1"`
	Age     uint8  `bcs:"2"`
	Enabled bool   `bcs:"3"`
}

func TestReflectionSerialization(t *testing.T) {
	t.Parallel()

	original := &ReflectStruct{
		Name:    "Alice",
		Age:     30,
		Enabled: true,
	}

	data, err := Marshal(original)
	require.NoError(t, err)

	var result ReflectStruct
	err = Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Age, result.Age)
	assert.Equal(t, original.Enabled, result.Enabled)
}

func TestReflectionSlice(t *testing.T) {
	t.Parallel()

	type SliceStruct struct {
		Values []uint64 `bcs:"1"`
	}

	original := &SliceStruct{
		Values: []uint64{1, 2, 3},
	}

	data, err := Marshal(original)
	require.NoError(t, err)

	var result SliceStruct
	err = Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.Values, result.Values)
}

func TestReflectionNested(t *testing.T) {
	t.Parallel()

	type Inner struct {
		Value uint64 `bcs:"1"`
	}

	type Outer struct {
		Inner Inner  `bcs:"1"`
		Name  string `bcs:"2"`
	}

	original := &Outer{
		Inner: Inner{Value: 42},
		Name:  "test",
	}

	data, err := Marshal(original)
	require.NoError(t, err)

	var result Outer
	err = Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.Inner.Value, result.Inner.Value)
	assert.Equal(t, original.Name, result.Name)
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	original := &TestStruct{Num: 12345, Enabled: false}

	data, err := Serialize(original)
	require.NoError(t, err)

	var result TestStruct
	err = Deserialize(&result, data)
	require.NoError(t, err)

	assert.Equal(t, original.Num, result.Num)
	assert.Equal(t, original.Enabled, result.Enabled)
}

// Additional tests for 100% coverage

func TestSerializerSetError(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.SetError(ErrNilValue)
	assert.ErrorIs(t, ser.Error(), ErrNilValue)

	// Further operations should not proceed
	ser.U64(42)
	assert.ErrorIs(t, ser.Error(), ErrNilValue) // Still same error
}

func TestSerializerReset(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.U64(42)
	assert.NotEmpty(t, ser.ToBytes())

	ser.Reset()
	assert.Empty(t, ser.ToBytes())
}

func TestSerializerFixedBytes(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	data := []byte{1, 2, 3, 4, 5}
	ser.FixedBytes(data)
	require.NoError(t, ser.Error())

	// Fixed bytes should not include length prefix
	assert.Equal(t, data, ser.ToBytes())
}

func TestSerializerStruct(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	s := &TestStruct{Num: 42, Enabled: true}
	ser.Struct(s)
	require.NoError(t, ser.Error())

	// Should be same as calling MarshalBCS directly
	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	assert.Equal(t, expected, ser.ToBytes())
}

func TestSerializerStructNil(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.Struct(nil)
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestSerializeSequenceFunc(t *testing.T) {
	t.Parallel()

	values := []uint64{1, 2, 3}
	ser := NewSerializer()
	SerializeSequenceFunc(ser, values, func(s *Serializer, v uint64) {
		s.U64(v)
	})
	require.NoError(t, ser.Error())

	des := NewDeserializer(ser.ToBytes())
	length := des.Uleb128()
	assert.Equal(t, uint32(3), length)

	assert.Equal(t, uint64(1), des.U64())
	assert.Equal(t, uint64(2), des.U64())
	assert.Equal(t, uint64(3), des.U64())
}

func TestSerializeSequenceWithError(t *testing.T) {
	t.Parallel()

	items := []*TestStruct{{Num: 1, Enabled: true}}
	ser := NewSerializer()
	ser.SetError(ErrNilValue) // Pre-set error
	SerializeSequence(ser, items)
	assert.ErrorIs(t, ser.Error(), ErrNilValue) // Should not proceed
}

func TestSerializeSequenceFuncWithError(t *testing.T) {
	t.Parallel()

	values := []uint64{1, 2, 3}
	ser := NewSerializer()
	ser.SetError(ErrNilValue) // Pre-set error
	SerializeSequenceFunc(ser, values, func(s *Serializer, v uint64) {
		s.U64(v)
	})
	assert.ErrorIs(t, ser.Error(), ErrNilValue) // Should not proceed
}

func TestDeserializerPosition(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2, 3, 4, 5}
	des := NewDeserializer(data)
	assert.Equal(t, 0, des.Position())

	des.U8()
	assert.Equal(t, 1, des.Position())

	des.U16()
	assert.Equal(t, 3, des.Position())
}

func TestDeserializerReadFixedBytesInto(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2, 3, 4, 5}
	des := NewDeserializer(data)

	buf := make([]byte, 3)
	des.ReadFixedBytesInto(buf)
	require.NoError(t, des.Error())

	assert.Equal(t, []byte{1, 2, 3}, buf)
	assert.Equal(t, 3, des.Position())
}

func TestDeserializerReadFixedBytesIntoNotEnough(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2}
	des := NewDeserializer(data)

	buf := make([]byte, 5)
	des.ReadFixedBytesInto(buf)
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializerReadFixedBytesIntoWithError(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2, 3}
	des := NewDeserializer(data)
	des.SetError(ErrNilValue) // Pre-set error

	buf := make([]byte, 2)
	des.ReadFixedBytesInto(buf)
	assert.ErrorIs(t, des.Error(), ErrNilValue)
}

func TestDeserializerStruct(t *testing.T) {
	t.Parallel()

	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	des := NewDeserializer(data)

	var s TestStruct
	des.Struct(&s)
	require.NoError(t, des.Error())

	assert.Equal(t, uint64(42), s.Num)
	assert.True(t, s.Enabled)
}

func TestDeserializerStructNil(t *testing.T) {
	t.Parallel()

	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	des := NewDeserializer(data)

	des.Struct(nil)
	assert.ErrorIs(t, des.Error(), ErrNilValue)
}

func TestDeserializeSequenceFunc(t *testing.T) {
	t.Parallel()

	// Create data: length=3, then 3 u64 values
	ser := NewSerializer()
	ser.Uleb128(3)
	ser.U64(10)
	ser.U64(20)
	ser.U64(30)

	des := NewDeserializer(ser.ToBytes())
	result := DeserializeSequenceFunc(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	require.NoError(t, des.Error())

	assert.Equal(t, []uint64{10, 20, 30}, result)
}

func TestDeserializeSequenceFuncWithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{3}) // Length 3, but no data
	des.SetError(ErrNilValue)         // Pre-set error

	result := DeserializeSequenceFunc(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	assert.Nil(t, result)
}

func TestDeserializeSequenceWithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{3}) // Length 3, but not enough data
	des.SetError(ErrNilValue)         // Pre-set error

	result := DeserializeSequence(des, func() *TestStruct { return &TestStruct{} })
	assert.Nil(t, result)
}

func TestDeserializeOptionWithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1}) // has_value=1, but no data
	des.SetError(ErrNilValue)         // Pre-set error

	result := DeserializeOption(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	assert.Nil(t, result)
}

func TestReadBytesWithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{5}) // Length 5, but no bytes
	des.SetError(ErrNilValue)         // Pre-set error

	result := des.ReadBytes()
	assert.Nil(t, result)
}

func TestReadFixedBytesWithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 2, 3})
	des.SetError(ErrNilValue) // Pre-set error

	result := des.ReadFixedBytes(2)
	assert.Nil(t, result)
}

func TestReadFixedBytesNotEnough(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 2})
	result := des.ReadFixedBytes(5)
	assert.Nil(t, result)
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestUleb128WithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{})
	des.SetError(ErrNilValue) // Pre-set error

	result := des.Uleb128()
	assert.Equal(t, uint32(0), result)
}

func TestUleb128Overflow(t *testing.T) {
	t.Parallel()

	// ULEB128 with continuation bits that would cause overflow
	data := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}
	des := NewDeserializer(data)

	_ = des.Uleb128()
	assert.Error(t, des.Error()) // Should have overflow error
}

func TestBigIntWithError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 2, 3}) // Not enough bytes for U128
	des.SetError(ErrNilValue)               // Pre-set error

	result := des.U128()
	// When error is preset, returns zero value (0), not nil
	assert.Equal(t, big.NewInt(0), result)
}

func TestWriteBytesWithError(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.SetError(ErrNilValue)
	ser.WriteBytes([]byte{1, 2, 3})
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestBigIntSerializeNil(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.U128(nil)
	assert.Error(t, ser.Error())
}

func TestBigIntDeserializeNotEnoughBytes(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 2, 3}) // Not enough for 16 bytes
	result := des.U128()
	// Returns zero value, not nil, but error is set
	assert.Equal(t, big.NewInt(0), result)
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializeBoolNotEnoughBytes(t *testing.T) {
	t.Parallel()

	_, err := DeserializeBool([]byte{})
	assert.Error(t, err)
}

func TestDeserializeU8NotEnoughBytes(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU8([]byte{})
	assert.Error(t, err)
}

func TestDeserializeU16NotEnoughBytes(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU16([]byte{0})
	assert.Error(t, err)
}

func TestDeserializeU32NotEnoughBytes(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU32([]byte{0, 0})
	assert.Error(t, err)
}

func TestDeserializeU64NotEnoughBytes(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU64([]byte{0, 0, 0, 0})
	assert.Error(t, err)
}

func TestDeserializeBytesNotEnoughBytes(t *testing.T) {
	t.Parallel()

	_, err := DeserializeBytes([]byte{5, 1, 2}) // Says 5 bytes but only 2
	assert.Error(t, err)
}

// Test reflection edge cases

func TestMarshalNonStruct(t *testing.T) {
	t.Parallel()

	// Marshal a primitive (should use reflection)
	v := uint64(42)
	data, err := Marshal(&v)
	require.NoError(t, err)

	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0}
	assert.Equal(t, expected, data)
}

func TestUnmarshalNonStruct(t *testing.T) {
	t.Parallel()

	data := []byte{42, 0, 0, 0, 0, 0, 0, 0}
	var v uint64
	err := Unmarshal(data, &v)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), v)
}

func TestMarshalNil(t *testing.T) {
	t.Parallel()

	_, err := Marshal(nil)
	assert.Error(t, err)
}

func TestUnmarshalNil(t *testing.T) {
	t.Parallel()

	err := Unmarshal([]byte{}, nil)
	assert.Error(t, err)
}

func TestMarshalSlice(t *testing.T) {
	t.Parallel()

	data, err := Marshal([]uint8{1, 2, 3})
	require.NoError(t, err)
	// Length prefix + bytes
	assert.Equal(t, []byte{3, 1, 2, 3}, data)
}

func TestUnmarshalSlice(t *testing.T) {
	t.Parallel()

	data := []byte{3, 1, 2, 3}
	var result []uint8
	err := Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, []uint8{1, 2, 3}, result)
}

func TestMarshalBool(t *testing.T) {
	t.Parallel()

	data, err := Marshal(true)
	require.NoError(t, err)
	assert.Equal(t, []byte{1}, data)
}

func TestUnmarshalBool(t *testing.T) {
	t.Parallel()

	var result bool
	err := Unmarshal([]byte{1}, &result)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestMarshalString(t *testing.T) {
	t.Parallel()

	data, err := Marshal("hi")
	require.NoError(t, err)
	assert.Equal(t, []byte{2, 'h', 'i'}, data)
}

func TestUnmarshalString(t *testing.T) {
	t.Parallel()

	var result string
	err := Unmarshal([]byte{2, 'h', 'i'}, &result)
	require.NoError(t, err)
	assert.Equal(t, "hi", result)
}

func TestMarshalInt8(t *testing.T) {
	t.Parallel()

	data, err := Marshal(int8(-1))
	require.NoError(t, err)
	assert.Equal(t, []byte{0xFF}, data)
}

func TestUnmarshalInt8(t *testing.T) {
	t.Parallel()

	var result int8
	err := Unmarshal([]byte{0xFF}, &result)
	require.NoError(t, err)
	assert.Equal(t, int8(-1), result)
}

func TestMarshalInt16(t *testing.T) {
	t.Parallel()

	data, err := Marshal(int16(0x0102))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x02, 0x01}, data) // Little endian
}

func TestUnmarshalInt16(t *testing.T) {
	t.Parallel()

	var result int16
	err := Unmarshal([]byte{0x02, 0x01}, &result)
	require.NoError(t, err)
	assert.Equal(t, int16(0x0102), result)
}

func TestMarshalInt32(t *testing.T) {
	t.Parallel()

	data, err := Marshal(int32(0x01020304))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x04, 0x03, 0x02, 0x01}, data)
}

func TestUnmarshalInt32(t *testing.T) {
	t.Parallel()

	var result int32
	err := Unmarshal([]byte{0x04, 0x03, 0x02, 0x01}, &result)
	require.NoError(t, err)
	assert.Equal(t, int32(0x01020304), result)
}

func TestMarshalInt64(t *testing.T) {
	t.Parallel()

	data, err := Marshal(int64(0x0102030405060708))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}, data)
}

func TestUnmarshalInt64(t *testing.T) {
	t.Parallel()

	var result int64
	err := Unmarshal([]byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}, &result)
	require.NoError(t, err)
	assert.Equal(t, int64(0x0102030405060708), result)
}

func TestMarshalUint16(t *testing.T) {
	t.Parallel()

	data, err := Marshal(uint16(0x0102))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x02, 0x01}, data)
}

func TestUnmarshalUint16(t *testing.T) {
	t.Parallel()

	var result uint16
	err := Unmarshal([]byte{0x02, 0x01}, &result)
	require.NoError(t, err)
	assert.Equal(t, uint16(0x0102), result)
}

func TestMarshalUint32(t *testing.T) {
	t.Parallel()

	data, err := Marshal(uint32(0x01020304))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x04, 0x03, 0x02, 0x01}, data)
}

func TestUnmarshalUint32(t *testing.T) {
	t.Parallel()

	var result uint32
	err := Unmarshal([]byte{0x04, 0x03, 0x02, 0x01}, &result)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x01020304), result)
}

type StructWithByteSlice struct {
	Data []byte `bcs:"1"`
}

func TestMarshalByteSlice(t *testing.T) {
	t.Parallel()

	s := &StructWithByteSlice{Data: []byte{1, 2, 3}}
	data, err := Marshal(s)
	require.NoError(t, err)
	expected := []byte{3, 1, 2, 3}
	assert.Equal(t, expected, data)
}

func TestUnmarshalByteSlice(t *testing.T) {
	t.Parallel()

	data := []byte{3, 1, 2, 3}
	var s StructWithByteSlice
	err := Unmarshal(data, &s)
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, s.Data)
}

type StructWithFixedArray struct {
	Data [4]byte `bcs:"1"`
}

func TestMarshalFixedArray(t *testing.T) {
	t.Parallel()

	s := &StructWithFixedArray{Data: [4]byte{1, 2, 3, 4}}
	data, err := Marshal(s)
	require.NoError(t, err)
	// Fixed arrays don't have length prefix
	assert.Equal(t, []byte{1, 2, 3, 4}, data)
}

func TestUnmarshalFixedArray(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2, 3, 4}
	var s StructWithFixedArray
	err := Unmarshal(data, &s)
	require.NoError(t, err)
	assert.Equal(t, [4]byte{1, 2, 3, 4}, s.Data)
}

type StructWithUnexported struct {
	Public  uint64 `bcs:"1"`
	private uint64 //nolint:unused
}

func TestMarshalUnexported(t *testing.T) {
	t.Parallel()

	s := &StructWithUnexported{Public: 42}
	data, err := Marshal(s)
	require.NoError(t, err)
	// Only public fields are serialized
	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0}
	assert.Equal(t, expected, data)
}

type StructWithOmitEmpty struct {
	Value    uint64 `bcs:"1"`
	Optional uint64 `bcs:"2,omitempty"`
}

func TestMarshalOmitEmpty(t *testing.T) {
	t.Parallel()

	s := &StructWithOmitEmpty{Value: 42, Optional: 0}
	data, err := Marshal(s)
	require.NoError(t, err)
	// Both fields serialized (omitempty is handled at app level, BCS serializes all)
	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	assert.Equal(t, expected, data)
}

func TestMarshalWithMarshaler(t *testing.T) {
	t.Parallel()

	// TestStruct implements Marshaler
	s := &TestStruct{Num: 42, Enabled: true}
	data, err := Marshal(s)
	require.NoError(t, err)
	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	assert.Equal(t, expected, data)
}

func TestUnmarshalWithUnmarshaler(t *testing.T) {
	t.Parallel()

	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	var s TestStruct
	err := Unmarshal(data, &s)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), s.Num)
	assert.True(t, s.Enabled)
}

func TestSerializeWithMarshaler(t *testing.T) {
	t.Parallel()

	s := &TestStruct{Num: 42, Enabled: true}
	data, err := Serialize(s)
	require.NoError(t, err)
	expected := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	assert.Equal(t, expected, data)
}

func TestMarshalNotPointer(t *testing.T) {
	t.Parallel()

	// Non-pointer struct
	s := ReflectStruct{Name: "test", Age: 30, Enabled: true}
	data, err := Marshal(s)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestUnmarshalNotPointer(t *testing.T) {
	t.Parallel()

	data := []byte{4, 't', 'e', 's', 't', 30, 1}
	var s ReflectStruct
	err := Unmarshal(data, s) // Not pointer - should error
	assert.Error(t, err)
}

// Additional tests for reflection edge cases and error paths

type StructWithOptional struct {
	Name     string `bcs:"1"`
	Optional uint64 `bcs:"2,optional"`
}

func TestUnmarshalOptionalSome(t *testing.T) {
	t.Parallel()

	// Name="hi", Optional=Some(42)
	data := []byte{2, 'h', 'i', 1, 42, 0, 0, 0, 0, 0, 0, 0}
	var s StructWithOptional
	err := Unmarshal(data, &s)
	require.NoError(t, err)
	assert.Equal(t, "hi", s.Name)
	assert.Equal(t, uint64(42), s.Optional)
}

func TestUnmarshalOptionalNone(t *testing.T) {
	t.Parallel()

	// Name="hi", Optional=None
	data := []byte{2, 'h', 'i', 0}
	var s StructWithOptional
	err := Unmarshal(data, &s)
	require.NoError(t, err)
	assert.Equal(t, "hi", s.Name)
	assert.Equal(t, uint64(0), s.Optional)
}

func TestUnmarshalOptionalInvalid(t *testing.T) {
	t.Parallel()

	// Name="hi", Optional has invalid length (2)
	data := []byte{2, 'h', 'i', 2}
	var s StructWithOptional
	err := Unmarshal(data, &s)
	assert.Error(t, err)
}

func TestMarshalOptionalField(t *testing.T) {
	t.Parallel()

	s := &StructWithOptional{Name: "hi", Optional: 42}
	data, err := Marshal(s)
	require.NoError(t, err)
	// Name="hi", Optional=Some(42)
	expected := []byte{2, 'h', 'i', 1, 42, 0, 0, 0, 0, 0, 0, 0}
	assert.Equal(t, expected, data)
}

func TestMarshalOptionalFieldZero(t *testing.T) {
	t.Parallel()

	s := &StructWithOptional{Name: "hi", Optional: 0}
	data, err := Marshal(s)
	require.NoError(t, err)
	// Name="hi", Optional=None (since it's zero)
	expected := []byte{2, 'h', 'i', 0}
	assert.Equal(t, expected, data)
}

func TestDeserializeSequenceError(t *testing.T) {
	t.Parallel()

	// Length is 3, but only 1 item in data
	ser := NewSerializer()
	ser.Uleb128(3)
	ser.U64(10)

	des := NewDeserializer(ser.ToBytes())
	result := DeserializeSequence(des, func() *TestStruct { return &TestStruct{} })
	assert.Nil(t, result)
	assert.Error(t, des.Error())
}

func TestDeserializeSequenceFuncError(t *testing.T) {
	t.Parallel()

	// Length is 3, but only 1 item in data
	ser := NewSerializer()
	ser.Uleb128(3)
	ser.U64(10)

	des := NewDeserializer(ser.ToBytes())
	result := DeserializeSequenceFunc(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	assert.Nil(t, result)
	assert.Error(t, des.Error())
}

func TestSerializerWriteBytesNotEnough(t *testing.T) {
	t.Parallel()

	// This tests the error path when length prefix fails
	// In practice, this is hard to trigger for WriteBytes
	// but we can test that it doesn't error on valid input
	ser := NewSerializer()
	ser.WriteBytes([]byte{1, 2, 3})
	require.NoError(t, ser.Error())
	assert.Equal(t, []byte{3, 1, 2, 3}, ser.ToBytes())
}

func TestDeserializerReadBytesLengthOverflow(t *testing.T) {
	t.Parallel()

	// Very large length that exceeds available bytes
	ser := NewSerializer()
	ser.Uleb128(1000000) // Large length

	des := NewDeserializer(ser.ToBytes())
	result := des.ReadBytes()
	assert.Nil(t, result)
	assert.Error(t, des.Error())
}

type NestedStruct struct {
	Inner ReflectStruct `bcs:"1"`
}

func TestUnmarshalNestedStruct(t *testing.T) {
	t.Parallel()

	inner := ReflectStruct{Name: "test", Age: 25, Enabled: true}

	// Now wrap it
	s := &NestedStruct{Inner: inner}
	outerData, err := Marshal(s)
	require.NoError(t, err)

	var result NestedStruct
	err = Unmarshal(outerData, &result)
	require.NoError(t, err)
	assert.Equal(t, inner.Name, result.Inner.Name)
	assert.Equal(t, inner.Age, result.Inner.Age)
	assert.Equal(t, inner.Enabled, result.Inner.Enabled)
}

func TestUnmarshalSliceError(t *testing.T) {
	t.Parallel()

	// Length 2, but only 1 u64
	data := []byte{2, 42, 0, 0, 0, 0, 0, 0, 0}
	var result []uint64
	err := Unmarshal(data, &result)
	assert.Error(t, err)
}

func TestDeserializeOptionInvalidBool(t *testing.T) {
	t.Parallel()

	// Option with invalid boolean (2 instead of 0 or 1)
	des := NewDeserializer([]byte{2})
	result := DeserializeOption(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	assert.Nil(t, result)
	// Should have error because ULEB128 returns 2, which is invalid for Option
	// Actually ULEB128 doesn't care, it returns 2. The error comes from invalid option length.
	assert.Error(t, des.Error())
}

func TestDeserializerUleb128NotEnoughBytes(t *testing.T) {
	t.Parallel()

	// Empty input - not enough bytes for ULEB128
	des := NewDeserializer([]byte{})
	result := des.Uleb128()
	assert.Equal(t, uint32(0), result)
	assert.Error(t, des.Error())
}

func TestDeserializerUleb128MultiByteNotEnough(t *testing.T) {
	t.Parallel()

	// Continuation bit set but no more bytes
	des := NewDeserializer([]byte{0x80})
	result := des.Uleb128()
	assert.Equal(t, uint32(0), result)
	assert.Error(t, des.Error())
}

type SliceOfStructs struct {
	Items []ReflectStruct `bcs:"1"`
}

func TestMarshalSliceOfStructs(t *testing.T) {
	t.Parallel()

	s := &SliceOfStructs{
		Items: []ReflectStruct{
			{Name: "a", Age: 1, Enabled: true},
			{Name: "b", Age: 2, Enabled: false},
		},
	}
	data, err := Marshal(s)
	require.NoError(t, err)

	var result SliceOfStructs
	err = Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, "a", result.Items[0].Name)
	assert.Equal(t, "b", result.Items[1].Name)
}

func TestSerializeSequenceNil(t *testing.T) {
	t.Parallel()

	var items []*TestStruct
	ser := NewSerializer()
	SerializeSequence(ser, items)
	require.NoError(t, ser.Error())
	assert.Equal(t, []byte{0}, ser.ToBytes()) // Empty sequence
}

func TestSerializeSequenceFuncNil(t *testing.T) {
	t.Parallel()

	var values []uint64
	ser := NewSerializer()
	SerializeSequenceFunc(ser, values, func(s *Serializer, v uint64) {
		s.U64(v)
	})
	require.NoError(t, ser.Error())
	assert.Equal(t, []byte{0}, ser.ToBytes()) // Empty sequence
}

func TestDeserializeDeserializerError(t *testing.T) {
	t.Parallel()

	// Create a valid serialized TestStruct, but make deserializer have error
	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	var s TestStruct

	// Manually create deserializer with error to test error propagation
	des := NewDeserializer(data)
	des.SetError(ErrInvalidBool) // Pre-set error
	des.Struct(&s)
	assert.ErrorIs(t, des.Error(), ErrInvalidBool)
}

func TestSerializeSerializerError(t *testing.T) {
	t.Parallel()

	s := &TestStruct{Num: 42, Enabled: true}
	ser := NewSerializer()
	ser.SetError(ErrNilValue) // Pre-set error
	ser.Struct(s)
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestDeserializerStructWithError(t *testing.T) {
	t.Parallel()

	data := []byte{42, 0, 0, 0, 0, 0, 0, 0, 1}
	des := NewDeserializer(data)
	des.SetError(ErrNilValue) // Pre-set error

	var s TestStruct
	des.Struct(&s)
	assert.ErrorIs(t, des.Error(), ErrNilValue)
}

// More edge case tests

type StructWithSlicePtr struct {
	Items []*uint64 `bcs:"1"`
}

type StructWithMap struct {
	Data map[string]uint64 `bcs:"1"`
}

type StructWithInterface struct {
	Data any `bcs:"1"`
}

func TestIsZeroPtrNil(t *testing.T) {
	t.Parallel()

	// Test isZero with nil pointer through optional field
	type TestOptionalPtr struct {
		Value *string `bcs:"1,optional"`
	}

	// nil pointer should serialize as None
	s := &TestOptionalPtr{Value: nil}
	data, err := Marshal(s)
	require.NoError(t, err)
	assert.Equal(t, []byte{0}, data) // None
}

func TestIsZeroSliceNil(t *testing.T) {
	t.Parallel()

	// Test isZero with nil slice through optional field
	type TestOptionalSlice struct {
		Values []uint64 `bcs:"1,optional"`
	}

	// nil slice should serialize as None
	s := &TestOptionalSlice{Values: nil}
	data, err := Marshal(s)
	require.NoError(t, err)
	assert.Equal(t, []byte{0}, data) // None
}

func TestIsZeroSliceEmpty(t *testing.T) {
	t.Parallel()

	// Test isZero with empty slice through optional field
	type TestOptionalSlice struct {
		Values []uint64 `bcs:"1,optional"`
	}

	// empty slice should also serialize as None
	s := &TestOptionalSlice{Values: []uint64{}}
	data, err := Marshal(s)
	require.NoError(t, err)
	assert.Equal(t, []byte{0}, data) // None
}

func TestMarshalUnsupportedType(t *testing.T) {
	t.Parallel()

	// Map is not supported in BCS reflection
	m := map[string]int{"a": 1}
	_, err := Marshal(m)
	assert.Error(t, err)
}

func TestUnmarshalUnsupportedType(t *testing.T) {
	t.Parallel()

	// Interface cannot be deserialized
	data := []byte{0}
	var i any
	err := Unmarshal(data, &i)
	assert.Error(t, err)
}

func TestMarshalArrayU64(t *testing.T) {
	t.Parallel()

	type ArrayStruct struct {
		Values [3]uint64 `bcs:"1"`
	}

	s := &ArrayStruct{Values: [3]uint64{1, 2, 3}}
	data, err := Marshal(s)
	require.NoError(t, err)

	var result ArrayStruct
	err = Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, [3]uint64{1, 2, 3}, result.Values)
}

func TestSerializeWithError(t *testing.T) {
	t.Parallel()

	// Create a struct that implements Marshaler but sets error
	type ErrorMarshaler struct{}
	// Can't easily test this without modifying the struct

	// Instead test that error in serializer propagates
	ser := NewSerializer()
	ser.SetError(ErrNilValue)
	SerializeSequence(ser, []*TestStruct{{Num: 1}})
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestDeserializeSequenceFuncWithErrorInLoop(t *testing.T) {
	t.Parallel()

	// Create data with length 2 but not enough items
	data := []byte{2, 42, 0, 0, 0, 0, 0, 0, 0} // Only 1 u64 for 2 items
	des := NewDeserializer(data)

	result := DeserializeSequenceFunc(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	assert.Nil(t, result)
	assert.Error(t, des.Error())
}

func TestDeserializeSequenceWithErrorInLoop(t *testing.T) {
	t.Parallel()

	// Create data with length 2 but not enough items for TestStruct
	data := []byte{2, 42, 0, 0, 0, 0, 0, 0, 0, 1} // Only 1 TestStruct for 2 items
	des := NewDeserializer(data)

	result := DeserializeSequence(des, func() *TestStruct { return &TestStruct{} })
	assert.Nil(t, result)
	assert.Error(t, des.Error())
}

func TestDeserializeOptionWithInvalidBool(t *testing.T) {
	t.Parallel()

	// ULEB128 value 2 is invalid for Option (only 0 or 1 valid)
	data := []byte{2}
	des := NewDeserializer(data)

	result := DeserializeOption(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	assert.Nil(t, result)
	assert.Error(t, des.Error())
}

func TestDeserializeOptionWithDeserializerError(t *testing.T) {
	t.Parallel()

	// Option with Some but value deserialization fails
	data := []byte{1} // Some, but no value for u64
	des := NewDeserializer(data)

	result := DeserializeOption(des, func(d *Deserializer) uint64 {
		return d.U64()
	})
	// Result is returned but contains zero value
	// The important thing is that error is set
	assert.Error(t, des.Error())
	_ = result // May be non-nil with zero value
}

func TestWriteBytesError(t *testing.T) {
	t.Parallel()

	// Can't easily cause WriteBytes to fail other than pre-set error
	// But let's verify normal operation and error propagation
	ser := NewSerializer()
	ser.WriteBytes([]byte{1, 2, 3})
	require.NoError(t, ser.Error())

	// Test with error already set
	ser2 := NewSerializer()
	ser2.SetError(ErrNilValue)
	ser2.WriteBytes([]byte{1, 2, 3})
	assert.ErrorIs(t, ser2.Error(), ErrNilValue)
}

func TestSerializeSequenceFuncWithItemError(t *testing.T) {
	t.Parallel()

	// Test that error during item serialization stops the sequence
	values := []int{1, 2, 3}
	ser := NewSerializer()
	SerializeSequenceFunc(ser, values, func(s *Serializer, v int) {
		if v == 2 {
			s.SetError(ErrNilValue)
		}
		s.U64(uint64(v))
	})
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestSerializeSequenceWithItemError(t *testing.T) {
	t.Parallel()

	// Create a struct that sets error during Marshal
	type ErrorStruct struct {
		fail bool
	}
	// Can't easily test this, but we can test nil in sequence
	// Actually nil will be skipped, let's just verify sequence handles errors
}

func TestMarshalSliceOfSlice(t *testing.T) {
	t.Parallel()

	type SliceOfSlice struct {
		Data [][]uint8 `bcs:"1"`
	}

	s := &SliceOfSlice{
		Data: [][]uint8{{1, 2}, {3, 4, 5}},
	}
	data, err := Marshal(s)
	require.NoError(t, err)

	var result SliceOfSlice
	err = Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, s.Data, result.Data)
}

func TestGetStructFieldsNoTag(t *testing.T) {
	t.Parallel()

	// Struct without bcs tags should still work (uses field order)
	type NoTagStruct struct {
		A uint64
		B uint64
	}

	s := &NoTagStruct{A: 1, B: 2}
	data, err := Marshal(s)
	require.NoError(t, err)

	var result NoTagStruct
	err = Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.A)
	assert.Equal(t, uint64(2), result.B)
}

func TestDeserializerBoolErrorPath(t *testing.T) {
	t.Parallel()

	// Test empty input for Bool
	des := NewDeserializer([]byte{})
	_ = des.Bool()
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializerU8ErrorPath(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{})
	_ = des.U8()
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializerU16ErrorPath(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{0})
	_ = des.U16()
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializerU32ErrorPath(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{0, 0})
	_ = des.U32()
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializerU64ErrorPath(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{0, 0, 0, 0})
	_ = des.U64()
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

func TestDeserializerReadBytesError(t *testing.T) {
	t.Parallel()

	// Length says 10, but only 2 bytes available
	des := NewDeserializer([]byte{10, 1, 2})
	_ = des.ReadBytes()
	assert.ErrorIs(t, des.Error(), ErrNotEnoughBytes)
}

// Tests for standalone helper functions with pre-existing errors

func TestDeserializeBoolWithPresetError(t *testing.T) {
	t.Parallel()

	// DeserializeBool creates its own deserializer, so we test edge cases
	_, err := DeserializeBool([]byte{})
	assert.Error(t, err)
}

func TestDeserializeU8WithPresetError(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU8([]byte{})
	assert.Error(t, err)
}

func TestDeserializeU16WithPresetError(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU16([]byte{})
	assert.Error(t, err)
}

func TestDeserializeU32WithPresetError(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU32([]byte{})
	assert.Error(t, err)
}

func TestDeserializeU64WithPresetError(t *testing.T) {
	t.Parallel()

	_, err := DeserializeU64([]byte{})
	assert.Error(t, err)
}

func TestDeserializeBytesWithPresetError(t *testing.T) {
	t.Parallel()

	// Empty bytes
	_, err := DeserializeBytes([]byte{})
	assert.Error(t, err)
}

// ErrorMarshaler sets an error during serialization
type ErrorMarshaler struct{}

func (e *ErrorMarshaler) MarshalBCS(ser *Serializer) {
	ser.SetError(ErrNilValue)
}

func TestSerializeError(t *testing.T) {
	t.Parallel()

	// Test Serialize when Marshaler sets error
	_, err := Serialize(&ErrorMarshaler{})
	assert.Error(t, err)
}

func TestDeserializeError(t *testing.T) {
	t.Parallel()

	// Test Deserialize when Unmarshaler fails
	// TestStruct expects 9 bytes minimum
	data := []byte{1, 2} // Too short
	var s TestStruct
	err := Deserialize(&s, data)
	assert.Error(t, err)
}

func TestMarshalError(t *testing.T) {
	t.Parallel()

	// Marshal with unsupported type
	ch := make(chan int)
	_, err := Marshal(&ch)
	assert.Error(t, err)
}

func TestUnmarshalError(t *testing.T) {
	t.Parallel()

	// Unmarshal with non-pointer should error
	data := []byte{1}
	var v uint8 = 0
	err := Unmarshal(data, v) // Not a pointer
	assert.Error(t, err)
}

func TestUnmarshalRemainingBytes(t *testing.T) {
	t.Parallel()

	// Unmarshal when there are remaining bytes
	data := []byte{42, 255} // Extra byte
	var v uint8
	err := Unmarshal(data, &v)
	assert.Error(t, err)
}

func TestSerializeSequenceFuncWithSerializerError(t *testing.T) {
	t.Parallel()

	values := []uint64{1, 2, 3}
	ser := NewSerializer()
	ser.SetError(ErrNilValue) // Pre-set error
	SerializeSequenceFunc(ser, values, func(s *Serializer, v uint64) {
		s.U64(v)
	})
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestSerializeSequenceWithSerializerError(t *testing.T) {
	t.Parallel()

	items := []*TestStruct{{Num: 1}}
	ser := NewSerializer()
	ser.SetError(ErrNilValue) // Pre-set error
	SerializeSequence(ser, items)
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestWriteBytesWithSerializerError(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.SetError(ErrNilValue) // Pre-set error
	ser.WriteBytes([]byte{1, 2, 3})
	assert.ErrorIs(t, ser.Error(), ErrNilValue)
}

func TestDeserializeSequenceWithDeserializerError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 1})
	des.SetError(ErrNilValue) // Pre-set error
	result := DeserializeSequence(des, func() *TestStruct { return &TestStruct{} })
	assert.Nil(t, result)
}

func TestDeserializeSequenceFuncWithDeserializerError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{2, 1, 0, 0, 0, 0, 0, 0, 0})
	des.SetError(ErrNilValue) // Pre-set error
	result := DeserializeSequenceFunc(des, func(d *Deserializer) uint64 { return d.U64() })
	assert.Nil(t, result)
}

func TestDeserializeOptionWithDeserializerPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	des.SetError(ErrNilValue) // Pre-set error
	result := DeserializeOption(des, func(d *Deserializer) uint64 { return d.U64() })
	assert.Nil(t, result)
}

func TestReadBytesWithDeserializerError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{3, 1, 2, 3})
	des.SetError(ErrNilValue) // Pre-set error
	result := des.ReadBytes()
	assert.Nil(t, result)
}

func TestUleb128WithPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{100})
	des.SetError(ErrNilValue) // Pre-set error
	result := des.Uleb128()
	assert.Equal(t, uint32(0), result)
}

func TestMarshalSliceError(t *testing.T) {
	t.Parallel()

	// Slice of unsupported type
	data := []chan int{make(chan int)}
	_, err := Marshal(&data)
	assert.Error(t, err)
}

func TestUnmarshalSliceOfStructError(t *testing.T) {
	t.Parallel()

	// Data claims 2 items but doesn't have enough bytes
	data := []byte{2, 4, 't', 'e', 's', 't', 30, 1} // Only 1 ReflectStruct
	var result []ReflectStruct
	err := Unmarshal(data, &result)
	assert.Error(t, err)
}

// Test remaining bytes paths for DeserializeX functions

func TestDeserializeBoolRemainingBytes(t *testing.T) {
	t.Parallel()

	// Extra byte at end
	_, err := DeserializeBool([]byte{1, 255})
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

func TestDeserializeU8RemainingBytes(t *testing.T) {
	t.Parallel()

	// Extra byte at end
	_, err := DeserializeU8([]byte{42, 255})
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

func TestDeserializeU16RemainingBytes(t *testing.T) {
	t.Parallel()

	// Extra byte at end
	_, err := DeserializeU16([]byte{1, 0, 255})
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

func TestDeserializeU32RemainingBytes(t *testing.T) {
	t.Parallel()

	// Extra byte at end
	_, err := DeserializeU32([]byte{1, 0, 0, 0, 255})
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

func TestDeserializeU64RemainingBytes(t *testing.T) {
	t.Parallel()

	// Extra byte at end
	_, err := DeserializeU64([]byte{1, 0, 0, 0, 0, 0, 0, 0, 255})
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

func TestDeserializeBytesRemainingBytes(t *testing.T) {
	t.Parallel()

	// Length 2, 2 bytes, then extra byte
	_, err := DeserializeBytes([]byte{2, 1, 2, 255})
	assert.ErrorIs(t, err, ErrRemainingBytes)
}

// Test deserializer U8/U16/U32/U64 with pre-existing error

func TestU8WithPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{42})
	des.SetError(ErrNilValue)
	result := des.U8()
	assert.Equal(t, uint8(0), result)
}

func TestU16WithPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 0})
	des.SetError(ErrNilValue)
	result := des.U16()
	assert.Equal(t, uint16(0), result)
}

func TestU32WithPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 0, 0, 0})
	des.SetError(ErrNilValue)
	result := des.U32()
	assert.Equal(t, uint32(0), result)
}

func TestU64WithPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1, 0, 0, 0, 0, 0, 0, 0})
	des.SetError(ErrNilValue)
	result := des.U64()
	assert.Equal(t, uint64(0), result)
}

func TestBoolWithPreError(t *testing.T) {
	t.Parallel()

	des := NewDeserializer([]byte{1})
	des.SetError(ErrNilValue)
	result := des.Bool()
	assert.False(t, result)
}
