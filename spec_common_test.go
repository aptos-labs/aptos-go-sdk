package aptos

import "github.com/aptos-labs/aptos-go-sdk/internal/util"

// TODO Add specific keys that must be deserialized and compared
const Test64ByteHex = "0x1234123412341234123412341234123412341234123412341234123412341234"
const Test65ByteHex = "0x1234123412341234123412341234123412341234123412341234123412341234A"
const Test128ByteHex = "0x12341234123412341234123412341234123412341234123412341234123412341234123412341234123412341234123412341234123412341234123412341234"
const TestInvalidHex = "0x0101"
const TestInvalidHexCharacters = "Not-Hex"

// Ed25519 test data
const (
	TestEd25519PrivateKeyHex = "0xc5338cd251c22daa8c9c9cc94f498cc8a5c7e1d2e75287a5dda91096fe64efa5"
	TestEd25519PublicKeyHex  = "0xde19e5d1880cac87d57484ce9ed2e84cf0f9599f12e7cc3a52e4e7657a763f2c"
	TestEd25519AddressHex    = "0x978c213990c4833df71548df7ce49d54c759d6b6d932de22b24d56060b7af2aa"
	TestEd25519Message       = "0x68656c6c6f20776f726c64" // ("Hello world")
	TestEd25519SignatureHex  = "0x9e653d56a09247570bb174a389e85b9226abd5c403ea6c504b386626a145158cd4efd66fc5e071c0e19538a96a05ddbda24d3c51e1e6a9dacc6bb1ce775cce07"
)

// SingleSignerEd25519 test data
const (
	TestSingleSignerEd25519PrivateKeyHex = "0xf508cbef4e0fe463204aab724a90791c9a9dbe60a53b4978bbddbc712b55f2fd"
	TestSingleSignerEd25519PublicKeyHex  = "0xe425451a5dc888ac871976c3c724dec6118910e7d11d344b4b07a22cd94e8c2e"
	TestSingleSignerEd25519AddressHex    = "0x5bdf77d5bf826c8c04273d4e7323f7bc4a85ee7ee34b37bd7458b7aed3639dd3"
	TestSingleSignerEd25519Message       = "0x68656c6c6f20776f726c64" // ("Hello world")
	TestSingleSignerEd25519SignatureHex  = "0xc6f50f4e0cb1961f6f7b28be1a1d80e3ece240dfbb7bd8a8b03cc26bfd144fc176295d7c322c5bf3d9669d2ad49d8bdbfe77254b4a6393d8c49da04b40cee600"
)

// Secp256k1 test data
const (
	TestSecp256k1PrivateKeyHex = "0xd107155adf816a0a94c6db3c9489c13ad8a1eda7ada2e558ba3bfa47c020347e"
	TestSecp256k1PublicKeyHex  = "0x04acdd16651b839c24665b7e2033b55225f384554949fef46c397b5275f37f6ee95554d70fb5d9f93c5831ebf695c7206e7477ce708f03ae9bb2862dc6c9e033ea"
	TestSecp256k1AddressHex    = "0x5792c985bc96f436270bd2a3c692210b09c7febb8889345ceefdbae4bacfe498"
	TestSecp256k1Message       = "0x68656c6c6f20776f726c64" // ("Hello world")
	TestSecp256k1SignatureHex  = "0xd0d634e843b61339473b028105930ace022980708b2855954b977da09df84a770c0b68c29c8ca1b5409a5085b0ec263be80e433c83fcf6debb82f3447e71edca"
)

const (
	OtherMessage = "0x1337deadbeef"
)

// ParseHex parses hex, but skips the error
func parseHex(hex string) []byte {
	bytes, err := util.ParseHex(hex)
	if err != nil {
		panic("Failed to parse hex: " + hex + " " + err.Error())
	}
	return bytes
}
