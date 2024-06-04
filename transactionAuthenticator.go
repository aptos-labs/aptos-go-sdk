package aptos

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

//region TransactionAuthenticator

type TransactionAuthenticatorVariant uint8

const (
	TransactionAuthenticatorEd25519      TransactionAuthenticatorVariant = 0
	TransactionAuthenticatorMultiEd25519 TransactionAuthenticatorVariant = 1
	TransactionAuthenticatorMultiAgent   TransactionAuthenticatorVariant = 2
	TransactionAuthenticatorFeePayer     TransactionAuthenticatorVariant = 3
	TransactionAuthenticatorSingleSender TransactionAuthenticatorVariant = 4
)

type TransactionAuthenticatorImpl interface {
	bcs.Struct
	// Verify Return true if this AccountAuthenticator approves
	Verify(data []byte) bool
}

// TransactionAuthenticator is used for authorizing a transaction.  This differs from crypto.AccountAuthenticator because it handles
// constructs like FeePayer and MultiAgent.  Some keys can't stand on their own as TransactionAuthenticators.
// Implements TransactionAuthenticatorImpl, bcs.Struct
type TransactionAuthenticator struct {
	Variant TransactionAuthenticatorVariant
	Auth    TransactionAuthenticatorImpl
}

//region TransactionAuthenticator TransactionAuthenticatorImpl

func (ea *TransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Auth.Verify(msg)
}

//endregion

//region TransactionAuthenticator bcs.Struct

func (ea *TransactionAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(ea.Variant))
	ea.Auth.MarshalBCS(bcs)
}

func (ea *TransactionAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	kindNum := bcs.Uleb128()
	if bcs.Error() != nil {
		return
	}
	ea.Variant = TransactionAuthenticatorVariant(kindNum)
	switch ea.Variant {
	case TransactionAuthenticatorEd25519:
		ea.Auth = &Ed25519TransactionAuthenticator{}
	case TransactionAuthenticatorMultiEd25519:
		ea.Auth = &MultiEd25519TransactionAuthenticator{}
	case TransactionAuthenticatorMultiAgent:
		ea.Auth = &MultiAgentTransactionAuthenticator{}
	case TransactionAuthenticatorFeePayer:
		ea.Auth = &FeePayerTransactionAuthenticator{}
	case TransactionAuthenticatorSingleSender:
		ea.Auth = &SingleSenderTransactionAuthenticator{}
	default:
		bcs.SetError(fmt.Errorf("unknown TransactionAuthenticator kind: %d", kindNum))
	}
	ea.Auth.UnmarshalBCS(bcs)
}

//endregion
//endregion

//region Ed25519TransactionAuthenticator

// Ed25519TransactionAuthenticator for legacy ED25519 accounts
// Implements TransactionAuthenticatorImpl, bcs.Struct
type Ed25519TransactionAuthenticator struct {
	Sender *crypto.AccountAuthenticator
}

//region Ed25519TransactionAuthenticator TransactionAuthenticatorImpl

func (ea *Ed25519TransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Sender.Verify(msg)
}

//endregion

//region Ed25519TransactionAuthenticator bcs.Struct

func (ea *Ed25519TransactionAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	ea.Sender.Auth.MarshalBCS(bcs)
}

func (ea *Ed25519TransactionAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.Sender = &crypto.AccountAuthenticator{}
	ea.Sender.Variant = crypto.AccountAuthenticatorEd25519
	ea.Sender.Auth = &crypto.Ed25519Authenticator{}
	bcs.Struct(ea.Sender.Auth)
}

//endregion
//endregion

//region MultiEd25519TransactionAuthenticator

type MultiEd25519TransactionAuthenticator struct {
	Sender *crypto.AccountAuthenticator
}

//region Ed25519TransactionAuthenticator TransactionAuthenticatorImpl

func (ea *MultiEd25519TransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Sender.Verify(msg)
}

//endregion

//region MultiEd25519TransactionAuthenticator bcs.Struct

func (ea *MultiEd25519TransactionAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	ea.Sender.MarshalBCS(bcs)
}

func (ea *MultiEd25519TransactionAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.Sender = &crypto.AccountAuthenticator{}
	ea.Sender.Variant = crypto.AccountAuthenticatorMultiEd25519
	ea.Sender.Auth = &crypto.MultiEd25519Authenticator{}
	bcs.Struct(ea.Sender.Auth)
}

//endregion
//endregion

//region MultiAgentTransactionAuthenticator

type MultiAgentTransactionAuthenticator struct {
	Sender                   *crypto.AccountAuthenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []crypto.AccountAuthenticator
}

//region MultiAgentTransactionAuthenticator TransactionAuthenticatorImpl

func (ea *MultiAgentTransactionAuthenticator) Verify(msg []byte) bool {
	sender := ea.Sender.Verify(msg)
	if !sender {
		return false
	}
	for _, sa := range ea.SecondarySigners {
		verified := sa.Verify(msg)
		if !verified {
			return false
		}
	}
	return true
}

//endregion

//region MultiAgentTransactionAuthenticator bcs.Struct

func (ea *MultiAgentTransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ea.Sender.MarshalBCS(ser)
	bcs.SerializeSequence(ea.SecondarySignerAddresses, ser)
	bcs.SerializeSequence(ea.SecondarySigners, ser)
}

func (ea *MultiAgentTransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.Sender = &crypto.AccountAuthenticator{}
	des.Struct(ea.Sender)
	ea.SecondarySignerAddresses = bcs.DeserializeSequence[AccountAddress](des)
	ea.SecondarySigners = bcs.DeserializeSequence[crypto.AccountAuthenticator](des)
}

//endregion
//endregion

//region FeePayerTransactionAuthenticator

type FeePayerTransactionAuthenticator struct {
	Sender                   *crypto.AccountAuthenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []crypto.AccountAuthenticator
	FeePayer                 *AccountAddress
	FeePayerAuthenticator    *crypto.AccountAuthenticator
}

//region FeePayerTransactionAuthenticator bcs.Struct

func (ea *FeePayerTransactionAuthenticator) Verify(msg []byte) bool {
	sender := ea.Sender.Verify(msg)
	if !sender {
		return false
	}
	for _, sa := range ea.SecondarySigners {
		verified := sa.Verify(msg)
		if !verified {
			return false
		}
	}
	return ea.FeePayerAuthenticator.Verify(msg)
}

//endregion

//region FeePayerTransactionAuthenticator bcs.Struct

func (ea *FeePayerTransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ea.Sender.MarshalBCS(ser)
	bcs.SerializeSequence(ea.SecondarySignerAddresses, ser)
	bcs.SerializeSequence(ea.SecondarySigners, ser)
	ser.Struct(ea.FeePayer)
	ser.Struct(ea.FeePayerAuthenticator)
}

func (ea *FeePayerTransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.Sender = &crypto.AccountAuthenticator{}
	des.Struct(ea.Sender)
	ea.SecondarySignerAddresses = bcs.DeserializeSequence[AccountAddress](des)
	ea.SecondarySigners = bcs.DeserializeSequence[crypto.AccountAuthenticator](des)

	ea.FeePayer = &AccountAddress{}
	des.Struct(ea.FeePayer)
	ea.FeePayerAuthenticator = &crypto.AccountAuthenticator{}
	des.Struct(ea.FeePayerAuthenticator)
}

//endregion
//endregion

//region SingleSenderTransactionAuthenticator

type SingleSenderTransactionAuthenticator struct {
	Sender *crypto.AccountAuthenticator
}

//region SingleSenderTransactionAuthenticator TransactionAuthenticatorImpl

func (ea *SingleSenderTransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Sender.Verify(msg)
}

//endregion

//region SingleSenderTransactionAuthenticator bcs.Struct

func (ea *SingleSenderTransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.Sender)
}

func (ea *SingleSenderTransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.Sender = &crypto.AccountAuthenticator{}
	des.Struct(ea.Sender)
}

//endregion
//endregion
