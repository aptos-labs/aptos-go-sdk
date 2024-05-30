package api

import "github.com/aptos-labs/aptos-go-sdk/internal/types"

// GUID describes a GUID associated with things like V1 events
type GUID struct {
	CreationNumber uint64
	AccountAddress *types.AccountAddress
}

func (o *GUID) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.CreationNumber, err = toUint64(data, "creation_number")
	if err != nil {
		return err
	}
	o.AccountAddress, err = toAccountAddress(data, "account_address")
	return err
}
