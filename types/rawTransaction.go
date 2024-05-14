package types

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"golang.org/x/crypto/sha3"
)

type RawTransaction struct {
	Sender         core.AccountAddress
	SequenceNumber uint64
	Payload        TransactionPayload
	MaxGasAmount   uint64
	GasUnitPrice   uint64

	// ExpirationTimetampSeconds is seconds since Unix epoch
	ExpirationTimetampSeconds uint64

	ChainId uint8
}

func (txn *RawTransaction) MarshalBCS(bcs *bcs.Serializer) {
	txn.Sender.MarshalBCS(bcs)
	bcs.U64(txn.SequenceNumber)
	txn.Payload.MarshalBCS(bcs)
	bcs.U64(txn.MaxGasAmount)
	bcs.U64(txn.GasUnitPrice)
	bcs.U64(txn.ExpirationTimetampSeconds)
	bcs.U8(txn.ChainId)
}

func (txn *RawTransaction) UnmarshalBCS(bcs *bcs.Deserializer) {
	txn.Sender.UnmarshalBCS(bcs)
	txn.SequenceNumber = bcs.U64()
	txn.Payload.UnmarshalBCS(bcs)
	txn.MaxGasAmount = bcs.U64()
	txn.GasUnitPrice = bcs.U64()
	txn.ExpirationTimetampSeconds = bcs.U64()
	txn.ChainId = bcs.U8()
}

func (txn *RawTransaction) SignableBytes() (signableBytes []byte, err error) {
	ser := bcs.Serializer{}
	txn.MarshalBCS(&ser)
	err = ser.Error()
	if err != nil {
		return
	}
	prehash := RawTransactionPrehash()
	txnbytes := ser.ToBytes()
	signableBytes = make([]byte, len(prehash)+len(txnbytes))
	copy(signableBytes, prehash)
	copy(signableBytes[len(prehash):], txnbytes)
	return signableBytes, nil
}

func (txn *RawTransaction) Sign(sender *core.Account) (stxn *SignedTransaction, err error) {
	signableBytes, err := txn.SignableBytes()
	if err != nil {
		return
	}
	authenticator, err := sender.Sign(signableBytes)
	if err != nil {
		return
	}

	stxn = &SignedTransaction{
		Transaction:   *txn,
		Authenticator: authenticator,
	}
	return
}

var rawTransactionPrehash []byte

const rawTransactionPrehashStr = "APTOS::RawTransaction"

// Return the sha3-256 prehash for RawTransaction
// Do not write to the []byte returned
func RawTransactionPrehash() []byte {
	if rawTransactionPrehash == nil {
		b32 := sha3.Sum256([]byte(rawTransactionPrehashStr))
		out := make([]byte, len(b32))
		copy(out, b32[:])
		rawTransactionPrehash = out
		return out
	}
	return rawTransactionPrehash
}
