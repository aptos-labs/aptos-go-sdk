package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"golang.org/x/crypto/sha3"
)

// RawTransaction representation of a transaction's parts prior to signing
type RawTransaction struct {
	Sender         core.AccountAddress
	SequenceNumber uint64
	Payload        TransactionPayload
	MaxGasAmount   uint64
	GasUnitPrice   uint64

	// ExpirationTimestampSeconds is seconds since Unix epoch
	ExpirationTimestampSeconds uint64

	ChainId uint8
}

func (txn *RawTransaction) MarshalBCS(bcs *bcs.Serializer) {
	txn.Sender.MarshalBCS(bcs)
	bcs.U64(txn.SequenceNumber)
	txn.Payload.MarshalBCS(bcs)
	bcs.U64(txn.MaxGasAmount)
	bcs.U64(txn.GasUnitPrice)
	bcs.U64(txn.ExpirationTimestampSeconds)
	bcs.U8(txn.ChainId)
}

func (txn *RawTransaction) UnmarshalBCS(bcs *bcs.Deserializer) {
	txn.Sender.UnmarshalBCS(bcs)
	txn.SequenceNumber = bcs.U64()
	txn.Payload.UnmarshalBCS(bcs)
	txn.MaxGasAmount = bcs.U64()
	txn.GasUnitPrice = bcs.U64()
	txn.ExpirationTimestampSeconds = bcs.U64()
	txn.ChainId = bcs.U8()
}

// SigningMessage generates the bytes needed to be signed by a signer
func (txn *RawTransaction) SigningMessage() (message []byte, err error) {
	ser := bcs.Serializer{}
	txn.MarshalBCS(&ser)
	err = ser.Error()
	if err != nil {
		return
	}
	prehash := RawTransactionPrehash()
	txnBytes := ser.ToBytes()
	message = make([]byte, len(prehash)+len(txnBytes))
	copy(message, prehash)
	copy(message[len(prehash):], txnBytes)
	return message, nil
}

func (txn *RawTransaction) Sign(sender *core.Account) (signedTxn *SignedTransaction, err error) {
	message, err := txn.SigningMessage()
	if err != nil {
		return
	}
	authenticator, err := sender.Sign(message)
	if err != nil {
		return
	}

	signedTxn = &SignedTransaction{
		Transaction:   *txn,
		Authenticator: authenticator,
	}
	return
}

var rawTransactionPrehash []byte

const rawTransactionPrehashStr = "APTOS::RawTransaction"

// RawTransactionPrehash Return the sha3-256 prehash for RawTransaction
// Do not write to the []byte returned
func RawTransactionPrehash() []byte {
	// Cache the prehash
	if rawTransactionPrehash == nil {
		b32 := sha3.Sum256([]byte(rawTransactionPrehashStr))
		out := make([]byte, len(b32))
		copy(out, b32[:])
		rawTransactionPrehash = out
		return out
	}
	return rawTransactionPrehash
}
