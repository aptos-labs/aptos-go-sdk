package aptos

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"golang.org/x/crypto/sha3"
)

//region RawTransaction

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

type RawTransactionImpl interface {
	bcs.Struct

	SigningMessage() (message []byte, err error)
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
	txnAuth := &TransactionAuthenticator{}
	switch auth.Variant {
	case crypto.AccountAuthenticatorEd25519:
		txnAuth.Variant = TransactionAuthenticatorEd25519
		txnAuth.Auth = &Ed25519TransactionAuthenticator{
			Sender: auth,
		}
	case crypto.AccountAuthenticatorMultiEd25519:
		txnAuth.Variant = TransactionAuthenticatorMultiEd25519
		txnAuth.Auth = &MultiEd25519TransactionAuthenticator{
			Sender: auth,
		}
	case crypto.AccountAuthenticatorSingleSender:
		txnAuth.Variant = TransactionAuthenticatorSingleSender
		txnAuth.Auth = &SingleSenderTransactionAuthenticator{
			Sender: auth,
		}
	case crypto.AccountAuthenticatorMultiKey:
		txnAuth.Variant = TransactionAuthenticatorSingleSender
		txnAuth.Auth = &SingleSenderTransactionAuthenticator{
			Sender: auth,
		}
	default:
		return nil, fmt.Errorf("unknown authenticator type %d", auth.Variant)
	}

	return &SignedTransaction{
		Transaction:   txn,
		Authenticator: txnAuth,
	}, nil
}

//region RawTransaction bcs.Struct

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

//endregion

//region RawTransaction MessageSigner

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

//endregion

//region RawTransaction Signer

func (txn *RawTransaction) Sign(signer crypto.Signer) (authenticator *crypto.AccountAuthenticator, err error) {
	message, err := txn.SigningMessage()
	if err != nil {
		return
	}
	return signer.Sign(message)
}

//endregion
//endregion

//region RawTransactionWithData

var rawTransactionWithDataPrehash []byte

const rawTransactionWithDataPrehashStr = "APTOS::RawTransactionWithData"

// RawTransactionWithDataPrehash Return the sha3-256 prehash for RawTransactionWithData
// Do not write to the []byte returned
func RawTransactionWithDataPrehash() []byte {
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

func (txn *RawTransactionWithData) ToMultiAgentSignedTransaction(
	sender *crypto.AccountAuthenticator,
	additionalSigners []crypto.AccountAuthenticator,
) (*SignedTransaction, bool) {
	if txn.Variant != MultiAgentRawTransactionWithDataVariant {
		return nil, false
	}
	multiAgent := txn.Inner.(*MultiAgentRawTransactionWithData)

	return &SignedTransaction{
		Transaction: txn,
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
	feePayer *AccountAddress,
	feePayerAuthenticator *crypto.AccountAuthenticator,
	additionalSigners []crypto.AccountAuthenticator,
	additionalAddresses []AccountAddress,
) (*SignedTransaction, bool) {
	if txn.Variant != MultiAgentWithFeePayerRawTransactionWithDataVariant {
		return nil, false
	}
	return &SignedTransaction{
		Transaction: txn,
		Authenticator: &TransactionAuthenticator{
			Variant: TransactionAuthenticatorFeePayer,
			Auth: &FeePayerTransactionAuthenticator{
				Sender:                   sender,
				SecondarySignerAddresses: additionalAddresses,
				SecondarySigners:         additionalSigners,
				FeePayer:                 feePayer,
				FeePayerAuthenticator:    feePayerAuthenticator,
			},
		},
	}, true
}

//region RawTransactionWithData Signer

func (txn *RawTransactionWithData) Sign(signer crypto.Signer) (authenticator *crypto.AccountAuthenticator, err error) {
	message, err := txn.SigningMessage()
	if err != nil {
		return
	}
	return signer.Sign(message)
}

//endregion

//region RawTransactionWithData MessageSigner

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

//endregion

//region RawTransactionWithData bcs.Struct

func (txn *RawTransactionWithData) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(txn.Variant))
	bcs.Struct(txn.Inner)
}

func (txn *RawTransactionWithData) UnmarshalBCS(bcs *bcs.Deserializer) {
	txn.Variant = RawTransactionWithDataVariant(bcs.Uleb128())
	switch txn.Variant {
	case MultiAgentRawTransactionWithDataVariant:
		txn.Inner = &MultiAgentRawTransactionWithData{}
	case MultiAgentWithFeePayerRawTransactionWithDataVariant:
		txn.Inner = &MultiAgentWithFeePayerRawTransactionWithData{}
	default:
		bcs.SetError(fmt.Errorf("unknown RawTransactionWithData variant %d", txn.Variant))
	}
	bcs.Struct(txn.Inner)
}

//endregion
//endregion

//region MultiAgentRawTransactionWithData

type MultiAgentRawTransactionWithData struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
}

//region MultiAgentRawTransactionWithData bcs.Struct

func (txn *MultiAgentRawTransactionWithData) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(txn.RawTxn)
	bcs.SerializeSequence(txn.SecondarySigners, ser)
}

func (txn *MultiAgentRawTransactionWithData) UnmarshalBCS(des *bcs.Deserializer) {
	txn.RawTxn = &RawTransaction{}
	des.Struct(txn.RawTxn)
	txn.SecondarySigners = bcs.DeserializeSequence[AccountAddress](des)
}

//endregion
//endregion

//region MultiAgentWithFeePayerRawTransactionWithData

type MultiAgentWithFeePayerRawTransactionWithData struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
	FeePayer         *AccountAddress
}

//region MultiAgentWithFeePayerRawTransactionWithData bcs.Struct

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

//endregion
//endregion
