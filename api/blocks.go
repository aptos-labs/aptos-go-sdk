package api

import (
	"encoding/json"
)

// region Block

// Block describes a block properties and may have attached transactions
//
// Example:
//
//	{
//		"block_height": "1",
//		"block_hash": "0x014e30aafd9f715ab6262322bf919abebd66d948f6822ffb8a2699a57722fb80",
//		"block_timestamp": "1665609760857472",
//		"first_version": "1",
//		"last_version": "1",
//		"transactions": null
//	}
type Block struct {
	BlockHash      Hash                    // BlockHash of the block, a 32-byte hash in hexadecimal format
	BlockHeight    uint64                  // BlockHeight of the block, starts at 0
	BlockTimestamp uint64                  // BlockTimestamp is the Unix timestamp of the block, in microseconds, may not be set for block 0
	FirstVersion   uint64                  // FirstVersion of the block
	LastVersion    uint64                  // LastVersion of the block inclusive, may be the same value as FirstVersion
	Transactions   []*CommittedTransaction // Transactions in the block if requested, otherwise it is empty
}

// region Block JSON

// UnmarshalJSON deserializes a JSON data blob into a [Block]
//
// It will fail if not all fields are present, or a transaction is unparsable.
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

	o.BlockHash = data.BlockHash
	o.BlockHeight = data.BlockHeight.ToUint64()
	o.BlockTimestamp = data.BlockTimestamp.ToUint64()
	o.FirstVersion = data.FirstVersion.ToUint64()
	o.LastVersion = data.LastVersion.ToUint64()
	o.Transactions = make([]*CommittedTransaction, len(data.Transactions))
	for i, tx := range data.Transactions {
		// TODO: Do I just save transactions as "unknown" if I can't parse them?
		err = json.Unmarshal(tx, &o.Transactions[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// endregion
// endregion
