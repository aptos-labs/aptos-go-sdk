// Package types provides core type definitions for the Aptos Go SDK.
//
// This internal package defines fundamental types used throughout the SDK:
//   - AccountAddress: 32-byte blockchain addresses
//   - TypeTag: Move type system representation
//
// # AccountAddress
//
// AccountAddress represents a 32-byte address on the Aptos blockchain:
//
//	// Parse from string
//	addr, err := types.ParseAddress("0x1")
//	addr, err := types.ParseAddress("0x0000000000000000000000000000000000000000000000000000000000000001")
//
//	// Must-parse (panics on error, useful for constants)
//	addr := types.MustParseAddress("0x1")
//
//	// String representation
//	str := addr.String()       // "0x1" for special addresses, full hex otherwise
//	str := addr.StringLong()   // Always full 64-char hex
//	str := addr.StringShort()  // Minimal hex without leading zeros
//
// Special addresses (0x0 through 0xf) use short-form representation per AIP-40.
//
// # TypeTag
//
// TypeTag represents Move types for generic functions:
//
//	// Parse from string
//	tag, err := types.ParseTypeTag("0x1::aptos_coin::AptosCoin")
//	tag, err := types.ParseTypeTag("vector<u8>")
//	tag, err := types.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
//
//	// Primitive types
//	tag := types.TypeTagBool
//	tag := types.TypeTagU64
//	tag := types.TypeTagAddress
//
// # BCS Serialization
//
// All types implement BCS serialization:
//
//	// AccountAddress serializes as 32 fixed bytes
//	data, err := bcs.Serialize(&addr)
//
//	// TypeTag serializes with variant prefix
//	data, err := bcs.Serialize(&tag)
//
// # JSON Serialization
//
// Types support JSON for API interactions:
//
//	// AccountAddress marshals to quoted hex string
//	data, _ := json.Marshal(addr)  // "\"0x1\""
//
//	// TypeTag marshals to string representation
//	data, _ := json.Marshal(tag)   // "\"0x1::aptos_coin::AptosCoin\""
//
// # Thread Safety
//
// AccountAddress and TypeTag are immutable value types and safe to use
// concurrently after creation.
//
// # Performance
//
// Address operations are optimized:
//   - ParseAddress avoids allocations for odd-length hex
//   - String() uses pre-computed table for special addresses (0x0-0xf)
package types
