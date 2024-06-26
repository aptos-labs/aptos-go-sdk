package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// GUID describes a GUID associated with things like V1 events
type GUID struct {
	Id GUIDId `json:"id"`
}

type GUIDId struct {
	CreationNumber uint64                // CreationNumber is the number of the GUID
	AccountAddress *types.AccountAddress // AccountAddress is the account address of the creator of the GUID
}

func (o *GUIDId) UnmarshalJSON(b []byte) error {
	type inner struct {
		AccountAddress string `json:"account_address"`
		CreationNumber U64    `json:"creation_number"`
	}

	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.AccountAddress = &types.AccountAddress{}
	err = o.AccountAddress.ParseStringRelaxed(data.AccountAddress)
	if err != nil {
		return err
	}
	o.CreationNumber = data.CreationNumber.toUint64()
	return nil
}

// U64 is a type for handling JSON string representations of the uint64
type U64 uint64

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

func (u *U64) toUint64() uint64 {
	return uint64(*u)
}

// HexBytes is a type for handling Bytes encoded as hex in JSON
type HexBytes []byte

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
type Hash = string // TODO: do we make this a 32 byte array? or byte array?
