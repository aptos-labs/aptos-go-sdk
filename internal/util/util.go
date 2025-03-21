package util

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

// Sha3256Hash hashes the input bytes using SHA3-256.
func Sha3256Hash(bytes [][]byte) []byte {
	hasher := sha3.New256()
	for _, b := range bytes {
		hasher.Write(b)
	}

	return hasher.Sum([]byte{})
}

// ParseHex Convenience function to deal with 0x at the beginning of hex strings.
func ParseHex(hexStr string) ([]byte, error) {
	// Allow hex encoding "empty hex"
	if hexStr == "0x" {
		return []byte{}, nil
	}

	hexStr = strings.TrimPrefix(hexStr, "0x")

	return hex.DecodeString(hexStr)
}

// BytesToHex converts a byte slice to a hex string with a leading 0x.
func BytesToHex(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
}

// StrToUint64 converts a string to a uint64.
func StrToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// StrToBigInt converts a string to a big.Int for u128 and u256 values.
func StrToBigInt(val string) (*big.Int, error) {
	num := &big.Int{}
	_, ok := num.SetString(val, 10)

	if !ok {
		return nil, fmt.Errorf("num %s is not an integer", val)
	}

	return num, nil
}

func UintToU8(u uint) (*uint8, error) {
	if u > 255 {
		return nil, fmt.Errorf("u %d is greater than 255", u)
	} else {
		val := uint8(u)
		return &val, nil
	}
}

func UintToU16(u uint) (*uint16, error) {
	if u > 65535 {
		return nil, fmt.Errorf("u %d is greater than 65535", u)
	}

	val := uint16(u)
	return &val, nil
}

func UintToU32(u uint) (*uint32, error) {
	if u > 4294967295 {
		return nil, fmt.Errorf("u %d is greater than 4294967295", u)
	}
	val := uint32(u)
	return &val, nil
}

func UintToU64(u uint) (*uint64, error) {
	// FIXME max U64
	if u > 4294967295 {
		return nil, fmt.Errorf("u %d is greater than 4294967295", u)
	}

	val := uint64(u)
	return &val, nil
}

func IntToU8(u int) (*uint8, error) {
	if u > 255 {
		return nil, fmt.Errorf("u %d is greater than 255", u)
	} else if u < 0 {
		return nil, fmt.Errorf("u %d is less than 0", u)
	}

	val := uint8(u)
	return &val, nil
}

func IntToU16(u int) (*uint16, error) {
	if u > 65535 {
		return nil, fmt.Errorf("u %d is greater than 65535", u)
	} else if u < 0 {
		return nil, fmt.Errorf("u %d is less than 0", u)
	}

	val := uint16(u)
	return &val, nil
}

func IntToU32(u int) (*uint32, error) {
	if u > 4294967295 {
		return nil, fmt.Errorf("u %d is greater than 4294967295", u)
	} else if u < 0 {
		return nil, fmt.Errorf("u %d is less than 0", u)
	}

	val := uint32(u)
	return &val, nil
}

func IntToU64(u int) (*uint64, error) {
	// FIXME: max u64
	if u > 4294967295 {
		return nil, fmt.Errorf("u %d is greater than 4294967295", u)
	} else if u < 0 {
		return nil, fmt.Errorf("u %d is less than 0", u)
	}

	val := uint64(u)
	return &val, nil
}
