// Package util provides internal helper functions for the Aptos Go SDK.
package util

import (
	"encoding/hex"
	"fmt"
	"hash"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/sha3"
)

// sha3Pool provides reusable SHA3-256 hashers to reduce allocations.
var sha3Pool = sync.Pool{
	New: func() interface{} {
		return sha3.New256()
	},
}

// Byte buffer pools for common sizes to reduce allocations.
var (
	buf32Pool = sync.Pool{New: func() interface{} { b := make([]byte, 32); return &b }}
	buf64Pool = sync.Pool{New: func() interface{} { b := make([]byte, 64); return &b }}
)

// GetBuffer32 gets a 32-byte buffer from the pool.
// Call PutBuffer32 when done to return it.
func GetBuffer32() *[]byte {
	return buf32Pool.Get().(*[]byte)
}

// PutBuffer32 returns a 32-byte buffer to the pool.
// Only accepts buffers with exactly 32 bytes length AND capacity to prevent
// sensitive data leakage in capacity beyond length.
func PutBuffer32(buf *[]byte) {
	if buf != nil && len(*buf) == 32 && cap(*buf) == 32 {
		// Clear sensitive data
		for i := range *buf {
			(*buf)[i] = 0
		}
		buf32Pool.Put(buf)
	}
	// If validation fails, don't return to pool - let GC handle it
}

// GetBuffer64 gets a 64-byte buffer from the pool.
// Call PutBuffer64 when done to return it.
func GetBuffer64() *[]byte {
	return buf64Pool.Get().(*[]byte)
}

// PutBuffer64 returns a 64-byte buffer to the pool.
// Only accepts buffers with exactly 64 bytes length AND capacity to prevent
// sensitive data leakage in capacity beyond length.
func PutBuffer64(buf *[]byte) {
	if buf != nil && len(*buf) == 64 && cap(*buf) == 64 {
		// Clear sensitive data
		for i := range *buf {
			(*buf)[i] = 0
		}
		buf64Pool.Put(buf)
	}
	// If validation fails, don't return to pool - let GC handle it
}

// Sha3256Hash hashes the input byte slices using SHA3-256.
// Uses a pool of hashers to minimize allocations in hot paths.
//
// Security note: hash.Hash.Reset() is guaranteed by the Go crypto interface
// to return the hasher to its initial state, clearing all internal buffers.
// SHA3 (Keccak) specifically zeros its state on Reset().
func Sha3256Hash(inputs [][]byte) []byte {
	hasher := sha3Pool.Get().(hash.Hash)
	hasher.Reset() // Clears all internal state per hash.Hash contract
	for _, b := range inputs {
		hasher.Write(b)
	}
	result := hasher.Sum(nil)
	sha3Pool.Put(hasher)
	return result
}

// ParseHex parses a hex string, handling optional "0x" prefix.
func ParseHex(hexStr string) ([]byte, error) {
	if hexStr == "0x" {
		return []byte{}, nil
	}
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return hex.DecodeString(hexStr)
}

// hexTable is the hex encoding alphabet.
const hexTable = "0123456789abcdef"

// BytesToHex converts bytes to a hex string with "0x" prefix.
// Optimized to avoid intermediate allocations.
func BytesToHex(bytes []byte) string {
	// Pre-allocate exact size: 2 (prefix) + 2*len (hex chars)
	buf := make([]byte, 2+len(bytes)*2)
	buf[0] = '0'
	buf[1] = 'x'
	for i, b := range bytes {
		buf[2+i*2] = hexTable[b>>4]
		buf[2+i*2+1] = hexTable[b&0x0f]
	}
	return string(buf)
}

// StrToUint64 converts a string to uint64.
func StrToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// StrToBigInt converts a decimal string to *big.Int.
func StrToBigInt(val string) (*big.Int, error) {
	num := &big.Int{}
	num, ok := num.SetString(val, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer: %s", val)
	}
	return num, nil
}

// IntToU8 converts int to uint8 with bounds checking.
func IntToU8(u int) (uint8, error) {
	if u > math.MaxUint8 || u < 0 {
		return 0, fmt.Errorf("value %d out of uint8 range", u)
	}
	return uint8(u), nil
}

// IntToU16 converts int to uint16 with bounds checking.
func IntToU16(u int) (uint16, error) {
	if u > math.MaxUint16 || u < 0 {
		return 0, fmt.Errorf("value %d out of uint16 range", u)
	}
	return uint16(u), nil
}

// IntToU32 converts int to uint32 with bounds checking.
func IntToU32(u int) (uint32, error) {
	if u > math.MaxUint32 || u < 0 {
		return 0, fmt.Errorf("value %d out of uint32 range", u)
	}
	return uint32(u), nil
}

// Uint32ToU8 converts uint32 to uint8 with bounds checking.
func Uint32ToU8(u uint32) (uint8, error) {
	if u > math.MaxUint8 {
		return 0, fmt.Errorf("value %d out of uint8 range", u)
	}
	return uint8(u), nil
}
