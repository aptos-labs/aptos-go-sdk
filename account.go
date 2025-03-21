package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// Re-export types so that way the user experience doesn't change

// AccountAddress is a 32 byte address on the Aptos blockchain
// It can represent an Object, an Account, and much more.
type AccountAddress = types.AccountAddress

// Account is a wrapper for a signer, handling the AccountAddress and signing.
type Account = types.Account

// AccountZero represents the 0x0 address
var AccountZero = types.AccountZero

// AccountOne represents the 0x1 address
var AccountOne = types.AccountOne

// AccountTwo represents the 0x2 address
var AccountTwo = types.AccountTwo

// AccountThree represents the 0x3 address
var AccountThree = types.AccountThree

// AccountFour represents the 0x4 address
var AccountFour = types.AccountFour

// NewAccountFromSigner creates an account from a Signer, which is most commonly a private key
func NewAccountFromSigner(signer crypto.Signer, accountAddress ...AccountAddress) (*Account, error) {
	return types.NewAccountFromSigner(signer, accountAddress...)
}

// NewEd25519Account creates a legacy Ed25519 account, this is most commonly used in wallets
func NewEd25519Account() (*Account, error) {
	return types.NewEd25519Account()
}

// NewEd25519SingleSenderAccount creates a single signer Ed25519 account
func NewEd25519SingleSenderAccount() (*Account, error) {
	return types.NewEd25519SingleSignerAccount()
}

// NewSecp256k1Account creates a Secp256k1 account
func NewSecp256k1Account() (*Account, error) {
	return types.NewSecp256k1Account()
}
