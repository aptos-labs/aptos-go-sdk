package types

import (
	"errors"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

// SignedTransaction a raw transaction plus its authenticator for a fully verifiable message
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

// Verify checks a signed transaction's signature
func (txn *SignedTransaction) Verify() error {
	bytes, err := txn.Transaction.SigningMessage()
	if err != nil {
		return err
	}
	if txn.Authenticator.Verify(bytes) {
		return nil
	}
	return errors.New("signature is invalid")
}
