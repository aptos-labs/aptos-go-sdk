// Package util provides internal helper functions for the Aptos Go SDK.
//
// This internal package contains utility functions used throughout the SDK:
//   - Hash functions
//   - Hex encoding/decoding
//   - Type conversions
//   - Buffer pooling
//
// # Hash Functions
//
// SHA3-256 hashing with pooled hashers for performance:
//
//	hash := util.Sha3256Hash([][]byte{data1, data2})
//
// The hasher pool reduces allocations in hot paths like signature verification
// and authentication key derivation.
//
// # Hex Encoding
//
// Convert between bytes and hex strings:
//
//	// Encode with 0x prefix
//	hex := util.BytesToHex(bytes)  // "0x..."
//
//	// Decode with optional 0x prefix
//	bytes, err := util.ParseHex("0x1234")
//	bytes, err := util.ParseHex("1234")
//
// BytesToHex is optimized to avoid intermediate allocations.
//
// # Type Conversions
//
// Safe integer conversions with bounds checking:
//
//	u8, err := util.IntToU8(255)     // OK
//	u8, err := util.IntToU8(256)     // Error: out of range
//
//	u16, err := util.IntToU16(value)
//	u32, err := util.IntToU32(value)
//	u8, err := util.Uint32ToU8(value)
//
// String to number conversions:
//
//	n, err := util.StrToUint64("123")
//	bigInt, err := util.StrToBigInt("123456789012345678901234567890")
//
// # Buffer Pools
//
// Reusable byte buffers for common sizes:
//
//	buf := util.GetBuffer32()
//	defer util.PutBuffer32(buf)
//	// Use (*buf)[:] for operations
//
//	buf := util.GetBuffer64()
//	defer util.PutBuffer64(buf)
//
// Buffer pools:
//   - Clear sensitive data before returning to pool
//   - Validate both length AND capacity on return
//   - Reject modified buffers to prevent data leakage
//
// # Thread Safety
//
// All functions in this package are thread-safe. Pooled resources use
// sync.Pool which handles concurrent access.
package util
