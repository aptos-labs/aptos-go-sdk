package api

import (
	"encoding/json"
)

//region Block

// Block describes a block properties and may have attached transactions
//
// Example:
//
//	{
//		"block_hash": "0x1234123412341234123412341234123412341234123412341234123412341234",
//		"block_height": "20",
//		"block_timestamp": "12345234343",
//		"first_version": "22",
//		"last_version": "25",
//		"transactions": []
//	}
type Block struct {
	BlockHash      Hash           // BlockHash of the block, a 32-byte hash in hexadecimal format
	BlockHeight    uint64         // BlockHeight of the block, starts at 0
	BlockTimestamp uint64         // BlockTimestamp is the Unix timestamp of the block, in milliseconds, may not be set for block 0
	FirstVersion   uint64         // FirstVersion of the block
	LastVersion    uint64         // LastVersion of the block inclusive, may be the same value as FirstVersion
	Transactions   []*Transaction // Transactions in the block if requested, otherwise it is empty
}

//region Block JSON

func (o *Block) MarshalJSON() ([]byte, error) {
	type inner struct {
		BlockHash      Hash              `json:"block_hash"`
		BlockHeight    U64               `json:"block_height"`
		BlockTimestamp U64               `json:"block_timestamp"`
		FirstVersion   U64               `json:"first_version"`
		LastVersion    U64               `json:"last_version"`
		Transactions   []json.RawMessage `json:"transactions"`
	}

	transactions := make([]json.RawMessage, len(o.Transactions))
	for i, tx := range o.Transactions {
		txn, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}
		transactions[i] = json.RawMessage(txn)
	}
	data := &inner{
		BlockHash:      o.BlockHash,
		BlockHeight:    U64(o.BlockHeight),
		BlockTimestamp: U64(o.BlockTimestamp),
		FirstVersion:   U64(o.FirstVersion),
		LastVersion:    U64(o.LastVersion),
		Transactions:   transactions,
	}
	return json.Marshal(data)
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

	o.BlockHash = data.BlockHash
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

//endregion
//endregion
