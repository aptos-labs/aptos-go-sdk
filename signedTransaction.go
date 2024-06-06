package aptos

import (
	"errors"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

// TransactionSigner is a generic interface for a way to sign transactions.  The default implementation is Account
//
// Note that AccountAddress is needed to be the correct on-chain value for proper signing.  This may differ from the
// AuthKey provided by the crypto.Signer
type TransactionSigner interface {
	crypto.Signer

	// AccountAddress returns the address of the signer, this may differ from the AuthKey derived from the inner signer
	AccountAddress() AccountAddress
}

// SignedTransaction a raw transaction plus its authenticator for a fully verifiable message
type SignedTransaction struct {
	Transaction   RawTransactionImpl
	Authenticator *TransactionAuthenticator
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

func (txn *SignedTransaction) Hash() ([]byte, error) {
	if TransactionPrefix == nil {
		hash := Sha3256Hash([][]byte{[]byte("APTOS::Transaction")})
		TransactionPrefix = &hash
	}

	txnBytes, err := bcs.Serialize(txn)
	if err != nil {
		return nil, err
	}

	// Transaction signature is defined as, the domain separated prefix based on struct (Transaction)
	// Then followed by the type of the transaction for the enum, UserTransaction is 0
	// Then followed by BCS encoded bytes of the signed transaction
	return Sha3256Hash([][]byte{*TransactionPrefix, {0}, txnBytes}), nil
}

var TransactionPrefix *[]byte
