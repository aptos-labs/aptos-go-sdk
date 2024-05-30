package api

import (
	"encoding/json"
)

// Block describes a block properties and may have attached transactions
type Block struct {
	BlockHash      string
	BlockHeight    uint64
	BlockTimestamp uint64
	FirstVersion   uint64
	LastVersion    uint64
	Transactions   []*Transaction
}

func (o *Block) UnmarshalJSON(b []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	o.BlockHash, err = toHash(data, "block_hash")
	if err != nil {
		return err
	}
	o.BlockHeight, err = toUint64(data, "block_height")
	if err != nil {
		return err
	}
	o.BlockTimestamp, err = toUint64(data, "block_timestamp")
	if err != nil {
		return err
	}
	o.FirstVersion, err = toUint64(data, "first_version")
	if err != nil {
		return err
	}
	o.LastVersion, err = toUint64(data, "last_version")
	if err != nil {
		return err
	}
	o.Transactions, err = toTransactions(data, "transactions")
	return err
}
