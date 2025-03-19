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

// PrettyJSON a simple pretty print for JSON examples.
func PrettyJSON(x any) string {
	out := strings.Builder{}
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "  ")
	err := enc.Encode(x)
	if err != nil {
		return ""
	}

	return out.String()
}
