package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// -- Note these are copied from internal/util/util.go to prevent package loops, but still allow devs to use it

// ParseHex Convenience function to deal with 0x at the beginning of hex strings
func ParseHex(hexStr string) ([]byte, error) {
	// This had to be redefined separately to get around a package loop
	return util.ParseHex(hexStr)
}

// Sha3256Hash takes a hash of the given sets of bytes
func Sha3256Hash(bytes [][]byte) (output []byte) {
	return util.Sha3256Hash(bytes)
}

var testConfig = TestnetConfig

// createTestClient to use for testing for only one place to configure the network
// TODO: Allow overrides with environment variable?
func createTestClient() (*Client, error) {
	return NewClient(testConfig)
}
