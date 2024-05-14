package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// ParseHex Convenience function to deal with 0x at the beginning of hex strings
func ParseHex(hexStr string) ([]byte, error) {
	// This had to be redefined separately to get around a package loop
	return util.ParseHex(hexStr)
}
