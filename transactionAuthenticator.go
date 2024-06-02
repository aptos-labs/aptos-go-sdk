package aptos

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

type TransactionAuthenticatorType uint8

const (
	TransactionAuthenticatorEd25519      TransactionAuthenticatorType = 0
	TransactionAuthenticatorMultiEd25519 TransactionAuthenticatorType = 1
	TransactionAuthenticatorMultiAgent   TransactionAuthenticatorType = 2
	TransactionAuthenticatorFeePayer     TransactionAuthenticatorType = 3
	TransactionAuthenticatorSingleSender TransactionAuthenticatorType = 4
)

type TransactionAuthenticatorImpl interface {
	bcs.Struct
	// Verify Return true if this Authenticator approves
	Verify(data []byte) bool
}

type TransactionAuthenticator struct {
	Kind TransactionAuthenticatorType
	Auth TransactionAuthenticatorImpl
}

func (ea *TransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Auth.Verify(msg)
}

func (ea *TransactionAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(ea.Kind))
	ea.Auth.MarshalBCS(bcs)
}

func (ea *TransactionAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	kindNum := bcs.Uleb128()
	if bcs.Error() != nil {
		return
	}
	ea.Kind = TransactionAuthenticatorType(kindNum)
	switch ea.Kind {
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

type Ed25519TransactionAuthenticator struct {
	Sender *crypto.Authenticator
}

func (ea *Ed25519TransactionAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	ea.Sender.MarshalBCS(bcs)
}

func (ea *Ed25519TransactionAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.Sender = &crypto.Authenticator{}
	bcs.Struct(ea.Sender)
}

func (ea *Ed25519TransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Sender.Verify(msg)
}

type MultiEd25519TransactionAuthenticator struct {
	Sender *crypto.Authenticator
}

func (ea *MultiEd25519TransactionAuthenticator) MarshalBCS(bcs *bcs.Serializer) {
	ea.Sender.MarshalBCS(bcs)
}

func (ea *MultiEd25519TransactionAuthenticator) UnmarshalBCS(bcs *bcs.Deserializer) {
	ea.Sender = &crypto.Authenticator{}
	bcs.Struct(ea.Sender)
}

func (ea *MultiEd25519TransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Sender.Verify(msg)
}

type MultiAgentTransactionAuthenticator struct {
	Sender                   *crypto.Authenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []crypto.Authenticator
}

func (ea *MultiAgentTransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ea.Sender.MarshalBCS(ser)
	bcs.SerializeSequence(ea.SecondarySignerAddresses, ser)
	bcs.SerializeSequence(ea.SecondarySigners, ser)
}

func (ea *MultiAgentTransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.Sender = &crypto.Authenticator{}
	des.Struct(ea.Sender)
	ea.SecondarySignerAddresses = bcs.DeserializeSequence[AccountAddress](des)
	ea.SecondarySigners = bcs.DeserializeSequence[crypto.Authenticator](des)
}

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

type FeePayerTransactionAuthenticator struct {
	Sender                   *crypto.Authenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []crypto.Authenticator
	FeePayer                 *AccountAddress
	FeePayerAuthenticator    *crypto.Authenticator
}

func (ea *FeePayerTransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ea.Sender.MarshalBCS(ser)
	bcs.SerializeSequence(ea.SecondarySignerAddresses, ser)
	bcs.SerializeSequence(ea.SecondarySigners, ser)
	ser.Struct(ea.FeePayer)
	ser.Struct(ea.FeePayerAuthenticator)
}

func (ea *FeePayerTransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.Sender = &crypto.Authenticator{}
	des.Struct(ea.Sender)
	ea.SecondarySignerAddresses = bcs.DeserializeSequence[AccountAddress](des)
	ea.SecondarySigners = bcs.DeserializeSequence[crypto.Authenticator](des)

	ea.FeePayer = &AccountAddress{}
	des.Struct(ea.FeePayer)
	ea.FeePayerAuthenticator = &crypto.Authenticator{}
	des.Struct(ea.FeePayerAuthenticator)
}

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

type SingleSenderTransactionAuthenticator struct {
	Sender *crypto.Authenticator
}

func (ea *SingleSenderTransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.Sender)
}

func (ea *SingleSenderTransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.Sender = &crypto.Authenticator{}
	des.Struct(ea.Sender)
}

func (ea *SingleSenderTransactionAuthenticator) Verify(msg []byte) bool {
	return ea.Sender.Verify(msg)
}
