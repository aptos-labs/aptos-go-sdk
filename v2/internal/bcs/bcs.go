// Package bcs implements Binary Canonical Serialization (BCS) for Aptos.
//
// BCS is the serialization format used by the Aptos blockchain for all on-chain data.
// It is a deterministic, non-self-describing binary format.
//
// This package provides two ways to serialize/deserialize data:
//
// 1. Interface-based: Implement Marshaler/Unmarshaler for fine-grained control
// 2. Reflection-based: Use Marshal/Unmarshal with struct tags for convenience
//
// # Interface-based Serialization
//
//	type MyStruct struct {
//	    Num     uint64
//	    Enabled bool
//	}
//
//	func (s *MyStruct) MarshalBCS(ser *Serializer) {
//	    ser.U64(s.Num)
//	    ser.Bool(s.Enabled)
//	}
//
//	func (s *MyStruct) UnmarshalBCS(des *Deserializer) {
//	    s.Num = des.U64()
//	    s.Enabled = des.Bool()
//	}
//
// # Reflection-based Serialization
//
//	type MyStruct struct {
//	    Num     uint64 `bcs:"1"`
//	    Enabled bool   `bcs:"2"`
//	}
//
//	data, err := bcs.Marshal(&MyStruct{Num: 42, Enabled: true})
//	var result MyStruct
//	err = bcs.Unmarshal(data, &result)
//
// # Struct Tags
//
// The following struct tags are supported:
//   - `bcs:"N"` - Field order for serialization (required for reflection)
//   - `bcs:"-"` - Skip this field
//   - `bcs:"optional"` - Treat as Option<T> (serialize as 0/1 length array)
//   - `bcs:"bytes"` - Treat string as raw bytes
package bcs

import (
	"errors"
)

// Common errors for BCS operations.
var (
	ErrNilValue         = errors.New("bcs: cannot marshal/unmarshal nil value")
	ErrInvalidType      = errors.New("bcs: invalid type for BCS serialization")
	ErrNotEnoughBytes   = errors.New("bcs: not enough bytes to deserialize")
	ErrRemainingBytes   = errors.New("bcs: unexpected remaining bytes after deserialization")
	ErrInvalidBool      = errors.New("bcs: invalid boolean value")
	ErrInvalidUleb128   = errors.New("bcs: invalid ULEB128 encoding")
	ErrOverflow         = errors.New("bcs: integer overflow")
	ErrInvalidOptionLen = errors.New("bcs: option must have 0 or 1 elements")
)

// Marshaler is an interface for types that can serialize themselves to BCS.
//
// Implementations should write to the serializer and call SetError on failure.
type Marshaler interface {
	MarshalBCS(ser *Serializer)
}

// Unmarshaler is an interface for types that can deserialize themselves from BCS.
//
// Implementations should read from the deserializer and call SetError on failure.
type Unmarshaler interface {
	UnmarshalBCS(des *Deserializer)
}

// Struct combines both Marshaler and Unmarshaler interfaces.
// Types implementing Struct can be both serialized and deserialized.
type Struct interface {
	Marshaler
	Unmarshaler
}

// Serialize serializes a Marshaler to bytes.
func Serialize(v Marshaler) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}
	ser := &Serializer{}
	v.MarshalBCS(ser)
	if ser.err != nil {
		return nil, ser.err
	}
	return ser.ToBytes(), nil
}

// Deserialize deserializes bytes into an Unmarshaler.
// Returns an error if there are remaining bytes after deserialization.
func Deserialize(v Unmarshaler, data []byte) error {
	if v == nil {
		return ErrNilValue
	}
	des := NewDeserializer(data)
	v.UnmarshalBCS(des)
	if des.err != nil {
		return des.err
	}
	if des.Remaining() > 0 {
		return ErrRemainingBytes
	}
	return nil
}

// Marshal serializes any value to BCS bytes using reflection.
// The value must be a pointer to a struct with bcs tags, or implement Marshaler.
func Marshal(v any) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	// If it implements Marshaler, use that
	if m, ok := v.(Marshaler); ok {
		return Serialize(m)
	}

	// Use reflection-based serialization
	ser := &Serializer{}
	if err := marshalReflect(ser, v); err != nil {
		return nil, err
	}
	if ser.err != nil {
		return nil, ser.err
	}
	return ser.ToBytes(), nil
}

// Unmarshal deserializes BCS bytes into a value using reflection.
// The value must be a pointer to a struct with bcs tags, or implement Unmarshaler.
func Unmarshal(data []byte, v any) error {
	if v == nil {
		return ErrNilValue
	}

	// If it implements Unmarshaler, use that
	if u, ok := v.(Unmarshaler); ok {
		return Deserialize(u, data)
	}

	// Use reflection-based deserialization
	des := NewDeserializer(data)
	if err := unmarshalReflect(des, v); err != nil {
		return err
	}
	if des.err != nil {
		return des.err
	}
	if des.Remaining() > 0 {
		return ErrRemainingBytes
	}
	return nil
}
