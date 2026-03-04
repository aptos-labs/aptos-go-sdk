package bcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSerialized(t *testing.T) {
	t.Parallel()
	input := []byte{0x01, 0x02, 0x03}
	s := NewSerialized(input)
	require.NotNil(t, s)
	assert.Equal(t, input, s.Value)
}

func TestSerialized_Serialized(t *testing.T) {
	t.Parallel()
	input := []byte{0xAA, 0xBB, 0xCC}
	s := NewSerialized(input)

	ser := &Serializer{}
	s.Serialized(ser)
	require.NoError(t, ser.Error())

	// WriteBytes prepends the length as ULEB128
	expected := append([]byte{0x03}, input...)
	assert.Equal(t, expected, ser.ToBytes())
}

func TestSerialized_SerializedForEntryFunction(t *testing.T) {
	t.Parallel()
	input := []byte{0x01, 0x02}
	s := NewSerialized(input)

	ser := &Serializer{}
	s.SerializedForEntryFunction(ser)
	require.NoError(t, ser.Error())

	// Entry function serialization is the same as regular serialization
	ser2 := &Serializer{}
	s.Serialized(ser2)
	require.NoError(t, ser2.Error())

	assert.Equal(t, ser2.ToBytes(), ser.ToBytes())
}

func TestSerialized_SerializedForScriptFunction(t *testing.T) {
	t.Parallel()
	input := []byte{0x01, 0x02}
	s := NewSerialized(input)

	ser := &Serializer{}
	s.SerializedForScriptFunction(ser)
	require.NoError(t, ser.Error())

	bytes := ser.ToBytes()
	// Script function prepends ULEB128(9) before the WriteBytes call
	// ULEB128(9) = 0x09
	assert.Equal(t, byte(0x09), bytes[0])
	// Then the rest is WriteBytes output (length-prefixed)
	assert.Greater(t, len(bytes), 1)
}

func TestSerialized_EmptyValue(t *testing.T) {
	t.Parallel()
	s := NewSerialized([]byte{})

	ser := &Serializer{}
	s.Serialized(ser)
	require.NoError(t, ser.Error())

	// Empty bytes: length 0 encoded as ULEB128
	assert.Equal(t, []byte{0x00}, ser.ToBytes())
}
