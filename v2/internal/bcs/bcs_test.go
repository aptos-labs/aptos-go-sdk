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
