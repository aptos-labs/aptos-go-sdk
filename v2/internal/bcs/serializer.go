package bcs

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"slices"
)

// Serializer writes BCS-encoded data to a buffer.
type Serializer struct {
	out bytes.Buffer
	err error
}

// NewSerializer creates a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Error returns any error that occurred during serialization.
func (ser *Serializer) Error() error {
	return ser.err
}

// SetError sets an error on the serializer.
// If an error is already set, this is a no-op.
func (ser *Serializer) SetError(err error) {
	if ser.err == nil {
		ser.err = err
	}
}

// ToBytes returns the serialized bytes.
func (ser *Serializer) ToBytes() []byte {
	return ser.out.Bytes()
}

// Reset clears the serializer for reuse.
// Zeroes the underlying buffer to prevent sensitive data leakage.
func (ser *Serializer) Reset() {
	// Clear sensitive data before resetting
	// This is important for pooled serializers to prevent data leakage
	if ser.out.Len() > 0 {
		buf := ser.out.Bytes()
		for i := range buf {
			buf[i] = 0
		}
	}
	ser.out.Reset()
	ser.err = nil
}

// Bool serializes a boolean as a single byte (0x00 or 0x01).
func (ser *Serializer) Bool(v bool) {
	if v {
		ser.U8(1)
	} else {
		ser.U8(0)
	}
}

// U8 serializes an unsigned 8-bit integer.
func (ser *Serializer) U8(v uint8) {
	ser.out.WriteByte(v)
}

// U16 serializes an unsigned 16-bit integer in little-endian format.
func (ser *Serializer) U16(v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	ser.out.Write(buf[:])
}

// U32 serializes an unsigned 32-bit integer in little-endian format.
func (ser *Serializer) U32(v uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	ser.out.Write(buf[:])
}

// U64 serializes an unsigned 64-bit integer in little-endian format.
func (ser *Serializer) U64(v uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	ser.out.Write(buf[:])
}

// U128 serializes an unsigned 128-bit integer in little-endian format.
func (ser *Serializer) U128(v *big.Int) {
	ser.serializeBigInt(16, v, false)
}

// U256 serializes an unsigned 256-bit integer in little-endian format.
func (ser *Serializer) U256(v *big.Int) {
	ser.serializeBigInt(32, v, false)
}

// I8 serializes a signed 8-bit integer.
func (ser *Serializer) I8(v int8) {
	ser.out.WriteByte(byte(v))
}

// I16 serializes a signed 16-bit integer in little-endian format.
func (ser *Serializer) I16(v int16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], uint16(v))
	ser.out.Write(buf[:])
}

// I32 serializes a signed 32-bit integer in little-endian format.
func (ser *Serializer) I32(v int32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(v))
	ser.out.Write(buf[:])
}

// I64 serializes a signed 64-bit integer in little-endian format.
func (ser *Serializer) I64(v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v))
	ser.out.Write(buf[:])
}

// I128 serializes a signed 128-bit integer in little-endian format.
func (ser *Serializer) I128(v *big.Int) {
	ser.serializeBigInt(16, v, true)
}

// I256 serializes a signed 256-bit integer in little-endian format.
func (ser *Serializer) I256(v *big.Int) {
	ser.serializeBigInt(32, v, true)
}

func (ser *Serializer) serializeBigInt(size int, v *big.Int, signed bool) {
	if v == nil {
		ser.SetError(ErrNilValue)
		return
	}

	toSerialize := new(big.Int).Set(v)

	// Handle negative numbers for signed types
	if signed && v.Sign() < 0 {
		// Two's complement: add 2^(size*8) to negative numbers
		modulus := new(big.Int).Lsh(big.NewInt(1), uint(size*8))
		toSerialize.Add(toSerialize, modulus)
	}

	buf := make([]byte, size)
	toSerialize.FillBytes(buf)
	// Reverse to little-endian
	slices.Reverse(buf)
	ser.out.Write(buf)
}

// Uleb128 serializes an unsigned 32-bit integer as ULEB128.
// This is used for sequence lengths and enum variants.
func (ser *Serializer) Uleb128(v uint32) {
	for v >= 0x80 {
		ser.out.WriteByte(byte(v&0x7f) | 0x80)
		v >>= 7
	}
	ser.out.WriteByte(byte(v))
}

// WriteBytes serializes a byte slice with a ULEB128 length prefix.
func (ser *Serializer) WriteBytes(v []byte) {
	if len(v) > 0xFFFFFFFF {
		ser.SetError(ErrOverflow)
		return
	}
	ser.Uleb128(uint32(len(v)))
	ser.out.Write(v)
}

// WriteString serializes a string as UTF-8 bytes with a ULEB128 length prefix.
func (ser *Serializer) WriteString(v string) {
	ser.WriteBytes([]byte(v))
}

// FixedBytes serializes a byte slice without a length prefix.
func (ser *Serializer) FixedBytes(v []byte) {
	ser.out.Write(v)
}

// Struct serializes a Marshaler.
func (ser *Serializer) Struct(v Marshaler) {
	if v == nil {
		ser.SetError(ErrNilValue)
		return
	}
	v.MarshalBCS(ser)
}

// SerializeSequence serializes a slice of Marshaler values with a length prefix.
func SerializeSequence[T Marshaler](ser *Serializer, items []T) {
	if len(items) > 0xFFFFFFFF {
		ser.SetError(ErrOverflow)
		return
	}
	ser.Uleb128(uint32(len(items)))
	for i, item := range items {
		item.MarshalBCS(ser)
		if ser.err != nil {
			return
		}
		_ = i // Avoid unused variable warning
	}
}

// SerializeSequenceFunc serializes a slice using a custom serialization function.
func SerializeSequenceFunc[T any](ser *Serializer, items []T, serialize func(*Serializer, T)) {
	if len(items) > 0xFFFFFFFF {
		ser.SetError(ErrOverflow)
		return
	}
	ser.Uleb128(uint32(len(items)))
	for _, item := range items {
		serialize(ser, item)
		if ser.err != nil {
			return
		}
	}
}

// SerializeOption serializes an optional value as a 0 or 1 length array.
func SerializeOption[T any](ser *Serializer, v *T, serialize func(*Serializer, T)) {
	if v == nil {
		ser.Uleb128(0)
	} else {
		ser.Uleb128(1)
		serialize(ser, *v)
	}
}

// Helper functions for common serializations

// SerializeBool serializes a single boolean.
func SerializeBool(v bool) ([]byte, error) {
	ser := &Serializer{}
	ser.Bool(v)
	return ser.ToBytes(), ser.err
}

// SerializeU8 serializes a single uint8.
func SerializeU8(v uint8) ([]byte, error) {
	ser := &Serializer{}
	ser.U8(v)
	return ser.ToBytes(), ser.err
}

// SerializeU16 serializes a single uint16.
func SerializeU16(v uint16) ([]byte, error) {
	ser := &Serializer{}
	ser.U16(v)
	return ser.ToBytes(), ser.err
}

// SerializeU32 serializes a single uint32.
func SerializeU32(v uint32) ([]byte, error) {
	ser := &Serializer{}
	ser.U32(v)
	return ser.ToBytes(), ser.err
}

// SerializeU64 serializes a single uint64.
func SerializeU64(v uint64) ([]byte, error) {
	ser := &Serializer{}
	ser.U64(v)
	return ser.ToBytes(), ser.err
}

// SerializeBytes serializes a byte slice with length prefix.
func SerializeBytes(v []byte) ([]byte, error) {
	ser := &Serializer{}
	ser.WriteBytes(v)
	return ser.ToBytes(), ser.err
}

// SerializeString serializes a string with length prefix.
func SerializeString(v string) ([]byte, error) {
	ser := &Serializer{}
	ser.WriteString(v)
	return ser.ToBytes(), ser.err
}
