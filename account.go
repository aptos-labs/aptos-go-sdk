package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// Re-export types so that way the user experience doesn't change

type AccountAddress = types.AccountAddress
type Account = types.Account

var AccountZero = types.AccountZero
var AccountOne = types.AccountOne
var AccountTwo = types.AccountTwo
var AccountThree = types.AccountThree
var AccountFour = types.AccountFour

func NewAccountFromSigner(signer crypto.Signer, authKey ...crypto.AuthenticationKey) (*Account, error) {
	return types.NewAccountFromSigner(signer, authKey...)
}

func NewEd25519Account() (*Account, error) {
	return types.NewEd25519Account()
}
