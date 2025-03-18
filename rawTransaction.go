package aptos

import (
	"fmt"
	"sync"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"golang.org/x/crypto/sha3"
)

// region RawTransaction
var (
	rawTransactionPrehash         []byte
	rawTransactionWithPrehashOnce sync.Once
)

const rawTransactionPrehashStr = "APTOS::RawTransaction"

// RawTransactionPrehash Return the sha3-256 prehash for RawTransaction
// Do not write to the []byte returned
func RawTransactionPrehash() []byte {
	// Cache the prehash
	rawTransactionWithPrehashOnce.Do(func() {
		b32 := sha3.Sum256([]byte(rawTransactionPrehashStr))
		rawTransactionPrehash = b32[:]
	})
	return rawTransactionPrehash
}

type RawTransactionImpl interface {
	bcs.Struct

	// SigningMessage creates a raw signing message for the transaction
	// Note that this should only be used externally if signing transactions outside the SDK.  Otherwise, use Sign.
	SigningMessage() (message []byte, err error)

	// Sign signs a transaction and returns the associated AccountAuthenticator, it will underneath sign the SigningMessage
	Sign(signer crypto.Signer) (*crypto.AccountAuthenticator, error)
}

// RawTransaction representation of a transaction's parts prior to signing
// Implements crypto.MessageSigner, crypto.Signer, bcs.Struct
type RawTransaction struct {
	Sender         AccountAddress
	SequenceNumber uint64
	Payload        TransactionPayload
	MaxGasAmount   uint64
	GasUnitPrice   uint64

	// ExpirationTimestampSeconds is seconds since Unix epoch
	ExpirationTimestampSeconds uint64

	ChainId uint8
}

func (txn *RawTransaction) SignedTransaction(sender crypto.Signer) (*SignedTransaction, error) {
	auth, err := txn.Sign(sender)
	if err != nil {
		return nil, err
	}
	return txn.SignedTransactionWithAuthenticator(auth)
}

// SignedTransactionWithAuthenticator signs the sender only signed transaction
func (txn *RawTransaction) SignedTransactionWithAuthenticator(auth *crypto.AccountAuthenticator) (*SignedTransaction, error) {
	txnAuth, err := NewTransactionAuthenticator(auth)
	if err != nil {
		return nil, err
	}
	return &SignedTransaction{
		Transaction:   txn,
		Authenticator: txnAuth,
	}, nil
}

// region RawTransaction bcs.Struct
func (txn *RawTransaction) MarshalBCS(ser *bcs.Serializer) {
	txn.Sender.MarshalBCS(ser)
	ser.U64(txn.SequenceNumber)
	txn.Payload.MarshalBCS(ser)
	ser.U64(txn.MaxGasAmount)
	ser.U64(txn.GasUnitPrice)
	ser.U64(txn.ExpirationTimestampSeconds)
	ser.U8(txn.ChainId)
}

func (txn *RawTransaction) UnmarshalBCS(des *bcs.Deserializer) {
	txn.Sender.UnmarshalBCS(des)
	txn.SequenceNumber = des.U64()
	txn.Payload.UnmarshalBCS(des)
	txn.MaxGasAmount = des.U64()
	txn.GasUnitPrice = des.U64()
	txn.ExpirationTimestampSeconds = des.U64()
	txn.ChainId = des.U8()
}

// endregion

// region RawTransaction MessageSigner

// SigningMessage generates the bytes needed to be signed by a signer
func (txn *RawTransaction) SigningMessage() (message []byte, err error) {
	txnBytes, err := bcs.Serialize(txn)
	if err != nil {
		return
	}
	prehash := RawTransactionPrehash()
	message = make([]byte, len(prehash)+len(txnBytes))
	copy(message, prehash)
	copy(message[len(prehash):], txnBytes)
	return message, nil
}

// endregion

// region RawTransaction Signer
func (txn *RawTransaction) Sign(signer crypto.Signer) (authenticator *crypto.AccountAuthenticator, err error) {
	message, err := txn.SigningMessage()
	if err != nil {
		return
	}
	return signer.Sign(message)
}

// endregion
// endregion

// region RawTransactionWithData
var (
	rawTransactionWithDataPrehash     []byte
	rawTransactionWithDataPrehashOnce sync.Once
)

const rawTransactionWithDataPrehashStr = "APTOS::RawTransactionWithData"

// RawTransactionWithDataPrehash Return the sha3-256 prehash for RawTransactionWithData
// Do not write to the []byte returned
func RawTransactionWithDataPrehash() []byte {
	// Cache the prehash
	rawTransactionWithDataPrehashOnce.Do(func() {
		b32 := sha3.Sum256([]byte(rawTransactionWithDataPrehashStr))
		rawTransactionWithDataPrehash = b32[:]
	})
	return rawTransactionWithDataPrehash
}

type RawTransactionWithDataVariant uint32

const (
	MultiAgentRawTransactionWithDataVariant             RawTransactionWithDataVariant = 0
	MultiAgentWithFeePayerRawTransactionWithDataVariant RawTransactionWithDataVariant = 1
)

type RawTransactionWithDataImpl interface {
	bcs.Struct
}

// TODO: make a function to make creating this easier

type RawTransactionWithData struct {
	Variant RawTransactionWithDataVariant
	Inner   RawTransactionWithDataImpl
}

func (txn *RawTransactionWithData) SetFeePayer(
	feePayer AccountAddress,
) bool {
	if txn.Variant == MultiAgentWithFeePayerRawTransactionWithDataVariant {
		inner := txn.Inner.(*MultiAgentWithFeePayerRawTransactionWithData)
		inner.FeePayer = &feePayer
		return true
	} else {
		return false
	}
}

func (txn *RawTransactionWithData) ToMultiAgentSignedTransaction(
	sender *crypto.AccountAuthenticator,
	additionalSigners []crypto.AccountAuthenticator,
) (*SignedTransaction, bool) {
	if txn.Variant != MultiAgentRawTransactionWithDataVariant {
		return nil, false
	}
	multiAgent := txn.Inner.(*MultiAgentRawTransactionWithData)

	return &SignedTransaction{
		Transaction: multiAgent.RawTxn,
		Authenticator: &TransactionAuthenticator{
			Variant: TransactionAuthenticatorMultiAgent,
			Auth: &MultiAgentTransactionAuthenticator{
				Sender:                   sender,
				SecondarySignerAddresses: multiAgent.SecondarySigners,
				SecondarySigners:         additionalSigners,
			},
		},
	}, true
}

func (txn *RawTransactionWithData) ToFeePayerSignedTransaction(
	sender *crypto.AccountAuthenticator,
	feePayerAuthenticator *crypto.AccountAuthenticator,
	additionalSigners []crypto.AccountAuthenticator,
) (*SignedTransaction, bool) {
	if txn.Variant != MultiAgentWithFeePayerRawTransactionWithDataVariant {
		return nil, false
	}
	feePayerTxn := txn.Inner.(*MultiAgentWithFeePayerRawTransactionWithData)
	return &SignedTransaction{
		Transaction: feePayerTxn.RawTxn,
		Authenticator: &TransactionAuthenticator{
			Variant: TransactionAuthenticatorFeePayer,
			Auth: &FeePayerTransactionAuthenticator{
				Sender:                   sender,
				SecondarySignerAddresses: feePayerTxn.SecondarySigners,
				SecondarySigners:         additionalSigners,
				FeePayer:                 feePayerTxn.FeePayer,
				FeePayerAuthenticator:    feePayerAuthenticator,
			},
		},
	}, true
}

// region RawTransactionWithData Signer
func (txn *RawTransactionWithData) Sign(signer crypto.Signer) (authenticator *crypto.AccountAuthenticator, err error) {
	message, err := txn.SigningMessage()
	if err != nil {
		return
	}
	return signer.Sign(message)
}

// endregion

// region RawTransactionWithData MessageSigner
func (txn *RawTransactionWithData) SigningMessage() (message []byte, err error) {
	txnBytes, err := bcs.Serialize(txn)
	if err != nil {
		return
	}
	prehash := RawTransactionWithDataPrehash()
	message = make([]byte, len(prehash)+len(txnBytes))
	copy(message, prehash)
	copy(message[len(prehash):], txnBytes)
	return message, nil
}

// endregion

// region RawTransactionWithData bcs.Struct
func (txn *RawTransactionWithData) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(txn.Variant))
	ser.Struct(txn.Inner)
}

func (txn *RawTransactionWithData) UnmarshalBCS(des *bcs.Deserializer) {
	txn.Variant = RawTransactionWithDataVariant(des.Uleb128())
	switch txn.Variant {
	case MultiAgentRawTransactionWithDataVariant:
		txn.Inner = &MultiAgentRawTransactionWithData{}
	case MultiAgentWithFeePayerRawTransactionWithDataVariant:
		txn.Inner = &MultiAgentWithFeePayerRawTransactionWithData{}
	default:
		des.SetError(fmt.Errorf("unknown RawTransactionWithData variant %d", txn.Variant))
		return
	}
	des.Struct(txn.Inner)
}

// endregion
// endregion

// region MultiAgentRawTransactionWithData
type MultiAgentRawTransactionWithData struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
}

// region MultiAgentRawTransactionWithData bcs.Struct
func (txn *MultiAgentRawTransactionWithData) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(txn.RawTxn)
	bcs.SerializeSequence(txn.SecondarySigners, ser)
}

func (txn *MultiAgentRawTransactionWithData) UnmarshalBCS(des *bcs.Deserializer) {
	txn.RawTxn = &RawTransaction{}
	des.Struct(txn.RawTxn)
	txn.SecondarySigners = bcs.DeserializeSequence[AccountAddress](des)
}

// endregion
// endregion

// region MultiAgentWithFeePayerRawTransactionWithData
type MultiAgentWithFeePayerRawTransactionWithData struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
	FeePayer         *AccountAddress
}

// region MultiAgentWithFeePayerRawTransactionWithData bcs.Struct
func (txn *MultiAgentWithFeePayerRawTransactionWithData) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(txn.RawTxn)
	bcs.SerializeSequence(txn.SecondarySigners, ser)
	ser.Struct(txn.FeePayer)
}

func (txn *MultiAgentWithFeePayerRawTransactionWithData) UnmarshalBCS(des *bcs.Deserializer) {
	txn.RawTxn = &RawTransaction{}
	des.Struct(txn.RawTxn)
	txn.SecondarySigners = bcs.DeserializeSequence[AccountAddress](des)
	txn.FeePayer = &AccountAddress{}
	des.Struct(txn.FeePayer)
}

// endregion
// endregion
