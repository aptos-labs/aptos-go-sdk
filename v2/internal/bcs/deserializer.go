package bcs

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"slices"
)

// Deserializer reads BCS-encoded data from a buffer.
type Deserializer struct {
	source []byte
	pos    int
	err    error
}

// NewDeserializer creates a new Deserializer from a byte slice.
func NewDeserializer(data []byte) *Deserializer {
	return &Deserializer{
		source: data,
		pos:    0,
	}
}

// Error returns any error that occurred during deserialization.
func (des *Deserializer) Error() error {
	return des.err
}

// SetError sets an error on the deserializer.
// If an error is already set, this is a no-op.
func (des *Deserializer) SetError(err error) {
	if des.err == nil {
		des.err = err
	}
}

// Remaining returns the number of bytes remaining to be read.
func (des *Deserializer) Remaining() int {
	return len(des.source) - des.pos
}

// Position returns the current read position.
func (des *Deserializer) Position() int {
	return des.pos
}

// Bool deserializes a boolean (single byte: 0x00 or 0x01).
func (des *Deserializer) Bool() bool {
	if des.err != nil {
		return false
	}
	if des.pos >= len(des.source) {
		des.SetError(fmt.Errorf("%w: bool", ErrNotEnoughBytes))
		return false
	}

	b := des.source[des.pos]
	des.pos++

	switch b {
	case 0:
		return false
	case 1:
		return true
	default:
		des.SetError(fmt.Errorf("%w: got 0x%02x", ErrInvalidBool, b))
		return false
	}
}

// U8 deserializes an unsigned 8-bit integer.
func (des *Deserializer) U8() uint8 {
	if des.err != nil {
		return 0
	}
	if des.pos >= len(des.source) {
		des.SetError(fmt.Errorf("%w: u8", ErrNotEnoughBytes))
		return 0
	}

	v := des.source[des.pos]
	des.pos++
	return v
}

// U16 deserializes an unsigned 16-bit integer in little-endian format.
func (des *Deserializer) U16() uint16 {
	if des.err != nil {
		return 0
	}
	if des.pos+2 > len(des.source) {
		des.SetError(fmt.Errorf("%w: u16", ErrNotEnoughBytes))
		return 0
	}

	v := binary.LittleEndian.Uint16(des.source[des.pos:])
	des.pos += 2
	return v
}

// U32 deserializes an unsigned 32-bit integer in little-endian format.
func (des *Deserializer) U32() uint32 {
	if des.err != nil {
		return 0
	}
	if des.pos+4 > len(des.source) {
		des.SetError(fmt.Errorf("%w: u32", ErrNotEnoughBytes))
		return 0
	}

	v := binary.LittleEndian.Uint32(des.source[des.pos:])
	des.pos += 4
	return v
}

// U64 deserializes an unsigned 64-bit integer in little-endian format.
func (des *Deserializer) U64() uint64 {
	if des.err != nil {
		return 0
	}
	if des.pos+8 > len(des.source) {
		des.SetError(fmt.Errorf("%w: u64", ErrNotEnoughBytes))
		return 0
	}

	v := binary.LittleEndian.Uint64(des.source[des.pos:])
	des.pos += 8
	return v
}

// U128 deserializes an unsigned 128-bit integer in little-endian format.
func (des *Deserializer) U128() *big.Int {
	return des.deserializeBigInt(16, false)
}

// U256 deserializes an unsigned 256-bit integer in little-endian format.
func (des *Deserializer) U256() *big.Int {
	return des.deserializeBigInt(32, false)
}

// I8 deserializes a signed 8-bit integer.
func (des *Deserializer) I8() int8 {
	return int8(des.U8())
}

// I16 deserializes a signed 16-bit integer in little-endian format.
func (des *Deserializer) I16() int16 {
	return int16(des.U16())
}

// I32 deserializes a signed 32-bit integer in little-endian format.
func (des *Deserializer) I32() int32 {
	return int32(des.U32())
}

// I64 deserializes a signed 64-bit integer in little-endian format.
func (des *Deserializer) I64() int64 {
	return int64(des.U64())
}

// I128 deserializes a signed 128-bit integer in little-endian format.
func (des *Deserializer) I128() *big.Int {
	return des.deserializeBigInt(16, true)
}

// I256 deserializes a signed 256-bit integer in little-endian format.
func (des *Deserializer) I256() *big.Int {
	return des.deserializeBigInt(32, true)
}

func (des *Deserializer) deserializeBigInt(size int, signed bool) *big.Int {
	if des.err != nil {
		return big.NewInt(0)
	}
	if des.pos+size > len(des.source) {
		des.SetError(fmt.Errorf("%w: %d-byte integer", ErrNotEnoughBytes, size))
		return big.NewInt(0)
	}

	// Copy and reverse to big-endian
	buf := make([]byte, size)
	copy(buf, des.source[des.pos:des.pos+size])
	des.pos += size
	slices.Reverse(buf)

	result := new(big.Int).SetBytes(buf)

	// Handle signed integers (two's complement)
	if signed && len(buf) > 0 && buf[0]&0x80 != 0 {
		modulus := new(big.Int).Lsh(big.NewInt(1), uint(size*8))
		result.Sub(result, modulus)
	}

	return result
}

// Uleb128 deserializes an unsigned 32-bit integer from ULEB128 encoding.
func (des *Deserializer) Uleb128() uint32 {
	if des.err != nil {
		return 0
	}

	var result uint64
	var shift uint

	for {
		if des.pos >= len(des.source) {
			des.SetError(fmt.Errorf("%w: uleb128", ErrNotEnoughBytes))
			return 0
		}

		b := des.source[des.pos]
		des.pos++

		result |= uint64(b&0x7f) << shift

		if b&0x80 == 0 {
			break
		}

		shift += 7
		if shift >= 35 {
			des.SetError(ErrInvalidUleb128)
			return 0
		}
	}

	if result > 0xFFFFFFFF {
		des.SetError(ErrOverflow)
		return 0
	}

	return uint32(result)
}

// ReadBytes deserializes a byte slice with a ULEB128 length prefix.
func (des *Deserializer) ReadBytes() []byte {
	if des.err != nil {
		return nil
	}

	length := des.Uleb128()
	if des.err != nil {
		return nil
	}

	return des.ReadFixedBytes(int(length))
}

// ReadBoundedBytes deserializes a byte slice with bounds checking BEFORE allocation.
// This provides DoS protection by rejecting oversized payloads before allocating memory.
// Returns nil and sets error if length is outside [minLen, maxLen].
func (des *Deserializer) ReadBoundedBytes(minLen, maxLen int) []byte {
	if des.err != nil {
		return nil
	}

	length := des.Uleb128()
	if des.err != nil {
		return nil
	}

	// Validate bounds BEFORE allocation to prevent DoS
	if int(length) < minLen {
		des.SetError(fmt.Errorf("byte slice too short: %d < %d", length, minLen))
		return nil
	}
	if int(length) > maxLen {
		des.SetError(fmt.Errorf("byte slice too large: %d > %d", length, maxLen))
		return nil
	}

	return des.ReadFixedBytes(int(length))
}

// ReadBoundedString deserializes a string with bounds checking BEFORE allocation.
// This provides DoS protection by rejecting oversized payloads before allocating memory.
func (des *Deserializer) ReadBoundedString(maxLen int) string {
	if des.err != nil {
		return ""
	}

	length := des.Uleb128()
	if des.err != nil {
		return ""
	}

	// Validate bounds BEFORE allocation
	if int(length) > maxLen {
		des.SetError(fmt.Errorf("string too large: %d > %d", length, maxLen))
		return ""
	}

	return string(des.ReadFixedBytes(int(length)))
}

// ReadString deserializes a UTF-8 string with a ULEB128 length prefix.
func (des *Deserializer) ReadString() string {
	return string(des.ReadBytes())
}

// ReadFixedBytes deserializes a fixed number of bytes without a length prefix.
func (des *Deserializer) ReadFixedBytes(length int) []byte {
	if des.err != nil {
		return nil
	}
	if des.pos+length > len(des.source) {
		des.SetError(fmt.Errorf("%w: %d bytes", ErrNotEnoughBytes, length))
		return nil
	}

	result := make([]byte, length)
	copy(result, des.source[des.pos:des.pos+length])
	des.pos += length
	return result
}

// ReadFixedBytesInto deserializes bytes into an existing slice.
func (des *Deserializer) ReadFixedBytesInto(dest []byte) {
	if des.err != nil {
		return
	}
	length := len(dest)
	if des.pos+length > len(des.source) {
		des.SetError(fmt.Errorf("%w: %d bytes", ErrNotEnoughBytes, length))
		return
	}

	copy(dest, des.source[des.pos:des.pos+length])
	des.pos += length
}

// Struct deserializes an Unmarshaler.
func (des *Deserializer) Struct(v Unmarshaler) {
	if des.err != nil {
		return
	}
	if v == nil {
		des.SetError(ErrNilValue)
		return
	}
	v.UnmarshalBCS(des)
}

// DeserializeSequence deserializes a slice of Unmarshaler values with a length prefix.
func DeserializeSequence[T Unmarshaler](des *Deserializer, newItem func() T) []T {
	if des.err != nil {
		return nil
	}

	length := des.Uleb128()
	if des.err != nil {
		return nil
	}

	result := make([]T, length)
	for i := range length {
		result[i] = newItem()
		result[i].UnmarshalBCS(des)
		if des.err != nil {
			return nil
		}
	}
	return result
}

// DeserializeSequenceFunc deserializes a slice using a custom deserialization function.
func DeserializeSequenceFunc[T any](des *Deserializer, deserialize func(*Deserializer) T) []T {
	if des.err != nil {
		return nil
	}

	length := des.Uleb128()
	if des.err != nil {
		return nil
	}

	result := make([]T, length)
	for i := range length {
		result[i] = deserialize(des)
		if des.err != nil {
			return nil
		}
	}
	return result
}

// DeserializeOption deserializes an optional value (0 or 1 length array).
func DeserializeOption[T any](des *Deserializer, deserialize func(*Deserializer) T) *T {
	if des.err != nil {
		return nil
	}

	length := des.Uleb128()
	if des.err != nil {
		return nil
	}

	switch length {
	case 0:
		return nil
	case 1:
		v := deserialize(des)
		return &v
	default:
		des.SetError(fmt.Errorf("%w: got %d elements", ErrInvalidOptionLen, length))
		return nil
	}
}

// Helper functions for common deserializations

// DeserializeBool deserializes a single boolean.
func DeserializeBool(data []byte) (bool, error) {
	des := NewDeserializer(data)
	v := des.Bool()
	if des.err != nil {
		return false, des.err
	}
	if des.Remaining() > 0 {
		return false, ErrRemainingBytes
	}
	return v, nil
}

// DeserializeU8 deserializes a single uint8.
func DeserializeU8(data []byte) (uint8, error) {
	des := NewDeserializer(data)
	v := des.U8()
	if des.err != nil {
		return 0, des.err
	}
	if des.Remaining() > 0 {
		return 0, ErrRemainingBytes
	}
	return v, nil
}

// DeserializeU16 deserializes a single uint16.
func DeserializeU16(data []byte) (uint16, error) {
	des := NewDeserializer(data)
	v := des.U16()
	if des.err != nil {
		return 0, des.err
	}
	if des.Remaining() > 0 {
		return 0, ErrRemainingBytes
	}
	return v, nil
}

// DeserializeU32 deserializes a single uint32.
func DeserializeU32(data []byte) (uint32, error) {
	des := NewDeserializer(data)
	v := des.U32()
	if des.err != nil {
		return 0, des.err
	}
	if des.Remaining() > 0 {
		return 0, ErrRemainingBytes
	}
	return v, nil
}

// DeserializeU64 deserializes a single uint64.
func DeserializeU64(data []byte) (uint64, error) {
	des := NewDeserializer(data)
	v := des.U64()
	if des.err != nil {
		return 0, des.err
	}
	if des.Remaining() > 0 {
		return 0, ErrRemainingBytes
	}
	return v, nil
}

// DeserializeBytes deserializes a byte slice with length prefix.
func DeserializeBytes(data []byte) ([]byte, error) {
	des := NewDeserializer(data)
	v := des.ReadBytes()
	if des.err != nil {
		return nil, des.err
	}
	if des.Remaining() > 0 {
		return nil, ErrRemainingBytes
	}
	return v, nil
}

// DeserializeString deserializes a string with length prefix.
func DeserializeString(data []byte) (string, error) {
	b, err := DeserializeBytes(data)
	return string(b), err
}
