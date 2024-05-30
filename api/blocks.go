package api

import (
	"encoding/json"
)

// Block describes a block properties and may have attached transactions
type Block struct {
	BlockHash      Hash
	BlockHeight    uint64
	BlockTimestamp uint64
	FirstVersion   uint64
	LastVersion    uint64
	Transactions   []*Transaction
}

func (o *Block) UnmarshalJSON(b []byte) error {
	type inner struct {
		BlockHash      Hash              `json:"block_hash"`
		BlockHeight    U64               `json:"block_height"`
		BlockTimestamp U64               `json:"block_timestamp"`
		FirstVersion   U64               `json:"first_version"`
		LastVersion    U64               `json:"last_version"`
		Transactions   []json.RawMessage `json:"transactions"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	o.BlockHash = Hash(data.BlockHash)
	o.BlockHeight = data.BlockHeight.toUint64()
	o.BlockTimestamp = data.BlockTimestamp.toUint64()
	o.FirstVersion = data.FirstVersion.toUint64()
	o.LastVersion = data.LastVersion.toUint64()
	o.Transactions = make([]*Transaction, len(data.Transactions))
	for i, tx := range data.Transactions {
		err = json.Unmarshal(tx, &o.Transactions[i])
		if err != nil {
			return err
		}
	}
	return nil
}
