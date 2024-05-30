package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"strconv"
	"strings"
)

// PrettyJson a simple pretty print for JSON examples
func PrettyJson(x any) string {
	out := strings.Builder{}
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "  ")
	err := enc.Encode(x)
	if err != nil {
		return ""
	}
	return out.String()
}

// GUID describes a GUID associated with things like V1 events
type GUID struct {
	CreationNumber uint64
	AccountAddress *types.AccountAddress
}

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
	o.CreationNumber = data.CreationNumber.toUint64()
	o.AccountAddress = data.AccountAddress
	return nil
}

type U64 uint64

func (u *U64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(u.toUint64(), 10))
}

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

type HexBytes []byte

func (u *HexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(util.BytesToHex(*u))
}

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

type Hash = string // TODO: do we make this a 32 byte array? or byte array?
