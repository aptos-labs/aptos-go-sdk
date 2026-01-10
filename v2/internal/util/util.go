// Package util provides internal helper functions for the Aptos Go SDK.
package util

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

// Sha3256Hash hashes the input byte slices using SHA3-256.
func Sha3256Hash(inputs [][]byte) []byte {
	hasher := sha3.New256()
	for _, b := range inputs {
		hasher.Write(b)
	}
	return hasher.Sum(nil)
}

// ParseHex parses a hex string, handling optional "0x" prefix.
func ParseHex(hexStr string) ([]byte, error) {
	if hexStr == "0x" {
		return []byte{}, nil
	}
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return hex.DecodeString(hexStr)
}

// BytesToHex converts bytes to a hex string with "0x" prefix.
func BytesToHex(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
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
