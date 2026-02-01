// Package bcs implements Binary Canonical Serialization (BCS) for Aptos.
//
// BCS is the serialization format used by the Aptos blockchain for all on-chain data.
// It is a deterministic, non-self-describing binary format that provides:
//   - Canonical encoding (same data always produces same bytes)
//   - Compact representation
//   - Fast serialization/deserialization
//
// # Interface-based Serialization
//
// For fine-grained control, implement the Marshaler and Unmarshaler interfaces:
//
//	type MyStruct struct {
//	    Value uint64
//	    Name  string
//	}
//
//	func (s *MyStruct) MarshalBCS(ser *Serializer) {
//	    ser.U64(s.Value)
//	    ser.WriteString(s.Name)
//	}
//
//	func (s *MyStruct) UnmarshalBCS(des *Deserializer) {
//	    s.Value = des.U64()
//	    s.Name = des.ReadString()
//	}
//
//	// Serialize
//	data, err := bcs.Serialize(&myStruct)
//
//	// Deserialize
//	var result MyStruct
//	err := bcs.Deserialize(&result, data)
//
// # Reflection-based Serialization
//
// For convenience, use struct tags with Marshal/Unmarshal:
//
//	type MyStruct struct {
//	    Value uint64 `bcs:"1"`
//	    Name  string `bcs:"2"`
//	}
//
//	data, err := bcs.Marshal(&myStruct)
//	err := bcs.Unmarshal(data, &result)
//
// # Supported Types
//
// Primitives:
//   - bool, uint8, uint16, uint32, uint64, int8, int16, int32, int64
//   - *big.Int for U128/U256/I128/I256
//   - string (UTF-8 with length prefix)
//   - []byte (with length prefix)
//
// Composite types:
//   - Structs (fields serialized in order)
//   - Slices (length prefix + elements)
//   - Maps (length prefix + key-value pairs, sorted by key)
//   - Pointers (Option type: 0 for nil, 1 + value for non-nil)
//
// # Thread Safety
//
// Serializer and Deserializer are NOT thread-safe. Each goroutine should
// use its own instance. The Serialize/Deserialize functions handle this
// automatically using sync.Pool.
//
// # Performance
//
// The package uses several optimizations:
//   - sync.Pool for Serializer reuse
//   - Stack-allocated buffers for small integers
//   - Zeroing of pooled buffers to prevent data leakage
package bcs
