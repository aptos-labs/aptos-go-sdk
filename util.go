package aptos

import (
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// -- Note these are copied from internal/util/util.go to prevent package loops, but still allow devs to use it

// ParseHex Convenience function to deal with 0x at the beginning of hex strings
func ParseHex(hexStr string) ([]byte, error) {
	// This had to be redefined separately to get around a package loop
	return util.ParseHex(hexStr)
}

// Sha3256Hash takes a hash of the given sets of bytes
func Sha3256Hash(bytes [][]byte) []byte {
	return util.Sha3256Hash(bytes)
}

// BytesToHex converts bytes to a 0x prefixed hex string
func BytesToHex(bytes []byte) string {
	return util.BytesToHex(bytes)
}

// StrToUint64 converts a string to a uint64
func StrToUint64(s string) (uint64, error) {
	return util.StrToUint64(s)
}

// StrToBigInt converts a string to a big.Int
func StrToBigInt(val string) (*big.Int, error) {
	return util.StrToBigInt(val)
}
