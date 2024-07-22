package bcs

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"slices"
)

// Deserializer is a type to deserialize a known set of bytes.
// The reader must know the types, as the format is not self-describing.
//
// Use [NewDeserializer] to initialize the Deserializer
//
//	bytes := []byte{0x01}
//	deserializer := NewDeserializer(bytes)
//	num := deserializer.U8()
//	if deserializer.Error() != nil {
//		return deserializer.Error()
//	}
type Deserializer struct {
	source []byte // Underlying data to parse
	pos    int    // Current position in the buffer
	err    error  // Any error that has happened so far
}

// NewDeserializer creates a new Deserializer from a byte array.
func NewDeserializer(bytes []byte) *Deserializer {
	return &Deserializer{
		source: bytes,
		pos:    0,
		err:    nil,
	}
}

// Deserialize deserializes a single item from bytes.
//
// This function will error if there are remaining bytes.
func Deserialize(dest Unmarshaler, bytes []byte) error {
	if dest == nil {
		return fmt.Errorf("cannot deserialize into nil")
	}
	des := Deserializer{
		source: bytes,
		pos:    0,
		err:    nil,
	}
	dest.UnmarshalBCS(&des)
	if des.err != nil {
		return des.err
	}
	if des.Remaining() > 0 {
		return fmt.Errorf("deserialize failed: remaining %d byte(s)", des.Remaining())
	}
	return nil
}

// Error If there has been any error, return it
func (des *Deserializer) Error() error {
	return des.err
}

// SetError If the data is well-formed but nonsense, UnmarshalBCS() code can set error
func (des *Deserializer) SetError(err error) {
	des.err = err
}

// Remaining tells the remaining bytes, which can be useful if there were more bytes than expected
//
//	bytes := []byte{0x01, 0x02}
//	deserializer := NewDeserializer(bytes)
//	num := deserializer.U8()
//	deserializer.Remaining == 1
func (des *Deserializer) Remaining() int {
	return len(des.source) - des.pos
}

// Bool deserializes a single byte as a bool
func (des *Deserializer) Bool() bool {
	if des.pos >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize bool")
		return false
	}

	out := false
	switch des.U8() {
	case 0:
		out = false
	case 1:
		out = true
	default:
		des.setError("bad bool at [%des]: %x", des.pos-1, des.source[des.pos-1])
	}
	return out
}

// U8 deserializes a single unsigned 8-bit integer
func (des *Deserializer) U8() uint8 {
	if des.pos >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize u8")
		return 0
	}
	out := des.source[des.pos]
	des.pos++
	return out
}

// U16 deserializes a single unsigned 16-bit integer
func (des *Deserializer) U16() uint16 {
	if des.pos+1 >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize u16")
		return 0
	}
	out := binary.LittleEndian.Uint16(des.source[des.pos : des.pos+2])
	des.pos += 2
	return out
}

// U32 deserializes a single unsigned 32-bit integer
func (des *Deserializer) U32() uint32 {
	if des.pos+3 >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize u32")
		return 0
	}
	out := binary.LittleEndian.Uint32(des.source[des.pos : des.pos+4])
	des.pos += 4
	return out
}

// U64 deserializes a single unsigned 64-bit integer
func (des *Deserializer) U64() uint64 {
	if des.pos+7 >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize u64")
		return 0
	}
	out := binary.LittleEndian.Uint64(des.source[des.pos : des.pos+8])
	des.pos += 8
	return out
}

// U128 deserializes a single unsigned 128-bit integer
func (des *Deserializer) U128() big.Int {
	if des.pos+15 >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize u128")
		return *big.NewInt(-1)
	}
	var bytesBigEndian [16]byte
	copy(bytesBigEndian[:], des.source[des.pos:des.pos+16])
	des.pos += 16
	slices.Reverse(bytesBigEndian[:])
	var out big.Int
	out.SetBytes(bytesBigEndian[:])
	return out
}

// U256 deserializes a single unsigned 256-bit integer
func (des *Deserializer) U256() big.Int {
	if des.pos+31 >= len(des.source) {
		des.setError("not enough bytes remaining to deserialize u256")
		return *big.NewInt(-1)
	}
	var bytesBigEndian [32]byte
	copy(bytesBigEndian[:], des.source[des.pos:des.pos+32])
	des.pos += 32
	slices.Reverse(bytesBigEndian[:])
	var out big.Int
	out.SetBytes(bytesBigEndian[:])
	return out
}

// Uleb128 deserializes a 32-bit integer from a variable length [Unsigned LEB128]
//
// [Unsigned LEB128]: https://en.wikipedia.org/wiki/LEB128#Unsigned_LEB128
func (des *Deserializer) Uleb128() uint32 {
	var out uint32 = 0
	shift := 0

	for {
		if des.pos >= len(des.source) {
			des.setError("not enough bytes remaining to deserialize uleb128")
			return 0
		}

		val := des.source[des.pos]
		out = out | (uint32(val&0x7f) << shift)
		des.pos++
		if (val & 0x80) == 0 {
			break
		}
		shift += 7
		// TODO: if shift is too much, error
	}

	return out
}

// ReadBytes reads bytes prefixed with a length
func (des *Deserializer) ReadBytes() []byte {
	length := des.Uleb128()
	if des.err != nil {
		return nil
	}
	if des.pos+int(length) > len(des.source) {
		des.setError("not enough bytes remaining to deserialize bytes")
		return nil
	}
	out := make([]byte, length)
	copy(out, des.source[des.pos:des.pos+int(length)])
	des.pos += int(length)
	return out
}

// ReadString reads UTF-8 bytes prefixed with a length
func (des *Deserializer) ReadString() string {
	return string(des.ReadBytes())
}

// ReadFixedBytes reads bytes not-prefixed with a length
func (des *Deserializer) ReadFixedBytes(length int) []byte {
	out := make([]byte, length)
	des.ReadFixedBytesInto(out)
	return out
}

// ReadFixedBytesInto reads bytes not-prefixed with a length into a byte array
func (des *Deserializer) ReadFixedBytesInto(dest []byte) {
	length := len(dest)
	if des.pos+length > len(des.source) {
		des.setError("not enough bytes remaining to deserialize fixedBytes")
		return
	}
	copy(dest, des.source[des.pos:des.pos+length])
	des.pos += length
}

// Struct reads an Unmarshaler implementation from bcs bytes
//
// This is used for handling types outside the provided primitives
func (des *Deserializer) Struct(v Unmarshaler) {
	if v == nil {
		des.setError("cannot deserialize into nil")
		return
	}
	v.UnmarshalBCS(des)
}

// DeserializeSequence deserializes an Unmarshaler implementation array
//
// This lets you deserialize a whole sequence of [Unmarshaler], and will fail if any member fails.
// All sequences are prefixed with an Uleb128 length.
func DeserializeSequence[T any](des *Deserializer) []T {
	return DeserializeSequenceWithFunction(des, func(des *Deserializer, out *T) {
		mv, ok := any(out).(Unmarshaler)
		if ok {
			mv.UnmarshalBCS(des)
		} else {
			// If it isn't of type Unmarshaler, we pass up an error
			des.setError("type is not Unmarshaler")
		}
	})
}

// DeserializeSequenceWithFunction deserializes any array with the given function
//
// This lets you deserialize a whole sequence of any type, and will fail if any member fails.
// All sequences are prefixed with an Uleb128 length.
func DeserializeSequenceWithFunction[T any](des *Deserializer, deserialize func(des *Deserializer, out *T)) []T {
	length := des.Uleb128()
	if des.Error() != nil {
		return nil
	}
	out := make([]T, length)
	for i := 0; i < int(length); i++ {
		deserialize(des, &out[i])

		if des.Error() != nil {
			des.setError("could not deserialize sequence[%d] member of %w", i, des.Error())
			return nil
		}
	}
	return out
}

// setError overrides the previous error, this can only be called from within the bcs package
func (des *Deserializer) setError(msg string, args ...any) {
	if des.err != nil {
		return
	}
	des.err = fmt.Errorf(msg, args...)
}
