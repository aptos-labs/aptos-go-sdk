package aptos

import (
	"encoding/json"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// MultiAgentTransaction represents a transaction that requires multiple signers
type MultiAgentTransaction struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
	FeePayer         *AccountAddress
}

// NewMultiAgentTransaction creates a new MultiAgentTransaction
func NewMultiAgentTransaction(rawTxn *RawTransaction, secondarySigners []AccountAddress, feePayer *AccountAddress) *MultiAgentTransaction {
	return &MultiAgentTransaction{
		RawTxn:           rawTxn,
		SecondarySigners: secondarySigners,
		FeePayer:         feePayer,
	}
}

// String returns a JSON formatted string representation of the MultiAgentTransaction
func (txn *MultiAgentTransaction) String() string {
	jsonBytes, err := json.MarshalIndent(txn, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling MultiAgentTransaction: %v", err)
	}
	return string(jsonBytes)
}

// MarshalBCS implements bcs.Struct
func (txn *MultiAgentTransaction) MarshalBCS(ser *bcs.Serializer) {
	txn.RawTxn.MarshalBCS(ser)
	bcs.SerializeSequence(txn.SecondarySigners, ser)
	if txn.FeePayer != nil {
		ser.Bool(true)
		txn.FeePayer.MarshalBCS(ser)
	} else {
		ser.Bool(false)
	}
}

// UnmarshalBCS implements bcs.Struct
func (txn *MultiAgentTransaction) UnmarshalBCS(des *bcs.Deserializer) {
	txn.RawTxn = &RawTransaction{}
	txn.RawTxn.UnmarshalBCS(des)
	txn.SecondarySigners = bcs.DeserializeSequence[AccountAddress](des)
	hasFeePayer := des.Bool()
	if hasFeePayer {
		txn.FeePayer = &AccountAddress{}
		txn.FeePayer.UnmarshalBCS(des)
	}
}
