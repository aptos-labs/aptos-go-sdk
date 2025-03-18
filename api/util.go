package api

import (
	"encoding/json"

	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// GUID describes a GUID associated with things like V1 events
//
// Note that this can only be used to deserialize events in the `events` field, and not the `GUID` resource in `changes`.
type GUID struct {
	CreationNumber uint64                // CreationNumber is the number of the GUID
	AccountAddress *types.AccountAddress // AccountAddress is the account address of the creator of the GUID
}

// UnmarshalJSON deserializes a JSON data blob into a [GUIDId]
func (o *GUID) UnmarshalJSON(b []byte) error {
	type inner struct {
		CreationNumber U64                   `json:"creation_number"`
		AccountAddress *types.AccountAddress `json:"account_address"`
	}

	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.AccountAddress = data.AccountAddress
	o.CreationNumber = data.CreationNumber.ToUint64()
	return nil
}

// U64 is a type for handling JSON string representations of the uint64
type U64 uint64

// UnmarshalJSON deserializes a JSON data blob into a [U64]
func (u *U64) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	uv, err := util.StrToUint64(str)
	if err != nil {
		return err
	}
	*u = U64(uv)
	return nil
}

// ToUint64 converts a [U64] to an uint64
//
// We can guarantee that it's safe to convert a [U64] to an uint64 because we've already validated the input on JSON parsing.
func (u *U64) ToUint64() uint64 {
	return uint64(*u)
}

// HexBytes is a type for handling Bytes encoded as hex in JSON
type HexBytes []byte

// UnmarshalJSON deserializes a JSON data blob into a [HexBytes]
//
// Example:
//
//	"0x123456" -> []byte{0x12, 0x34, 0x56}
func (u *HexBytes) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	bytes, err := util.ParseHex(str)
	if err != nil {
		return err
	}
	*u = bytes
	return nil
}

// Hash is a representation of a hash as Hex in JSON
//
// # This is always represented as a 32-byte hash in hexadecimal format
//
// Example:
//
//	0xf4d07fdb8b5151971886a910e516d418a790dd5f6e068b0588066518a395a600
type Hash = string // TODO: do we make this a 32 byte array? or byte array?
