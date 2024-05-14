package types

import (
	"errors"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

type SignedTransaction struct {
	Transaction   RawTransaction
	Authenticator crypto.Authenticator
}

func (txn *SignedTransaction) MarshalBCS(bcs *bcs.Serializer) {
	txn.Transaction.MarshalBCS(bcs)
	txn.Authenticator.MarshalBCS(bcs)
}
func (txn *SignedTransaction) UnmarshalBCS(bcs *bcs.Deserializer) {
	txn.Transaction.UnmarshalBCS(bcs)
	txn.Authenticator.UnmarshalBCS(bcs)
}

func (txn *SignedTransaction) Verify() error {
	tbytes, err := txn.Transaction.SignableBytes()
	if err != nil {
		return err
	}
	if txn.Authenticator.Verify(tbytes) {
		return nil
	}
	return errors.New("Bad Signature")
}
