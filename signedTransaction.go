package aptos

import (
	"errors"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

type SignedTransactionVariant uint8

// TransactionSigner is a generic interface for a way to sign transactions.  The default implementation is Account
//
// Note that AccountAddress is needed to be the correct on-chain value for proper signing.  This may differ from the
// AuthKey provided by the crypto.Signer
type TransactionSigner interface {
	crypto.Signer

	// AccountAddress returns the address of the signer, this may differ from the AuthKey derived from the inner signer
	AccountAddress() AccountAddress
}

//region SignedTransaction

const UserTransactionVariant SignedTransactionVariant = 0

// SignedTransaction a raw transaction plus its authenticator for a fully verifiable message
type SignedTransaction struct {
	Transaction   RawTransactionImpl
	Authenticator *TransactionAuthenticator
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

// TransactionPrefix is a cached hash prefix for taking transaction hashes
var TransactionPrefix *[]byte

// Hash takes the hash of the SignedTransaction
//
// Note: At the moment, this assumes that the transaction is a UserTransaction
func (txn *SignedTransaction) Hash() (string, error) {
	if TransactionPrefix == nil {
		hash := Sha3256Hash([][]byte{[]byte("APTOS::Transaction")})
		TransactionPrefix = &hash
	}

	txnBytes, err := bcs.Serialize(txn)
	if err != nil {
		return "", err
	}

	// Transaction signature is defined as, the domain separated prefix based on struct (Transaction)
	// Then followed by the type of the transaction for the enum, UserTransaction is 0
	// Then followed by BCS encoded bytes of the signed transaction
	hashBytes := Sha3256Hash([][]byte{*TransactionPrefix, {byte(UserTransactionVariant)}, txnBytes})
	return BytesToHex(hashBytes), nil
}

//region SignedTransaction bcs.Struct

func (txn *SignedTransaction) MarshalBCS(ser *bcs.Serializer) {
	txn.Transaction.MarshalBCS(ser)
	txn.Authenticator.MarshalBCS(ser)
}
func (txn *SignedTransaction) UnmarshalBCS(des *bcs.Deserializer) {
	txn.Transaction.UnmarshalBCS(des)
	txn.Authenticator.UnmarshalBCS(des)
}

//endregion
//endregion
