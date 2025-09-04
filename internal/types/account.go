package types

import (
	"errors"

	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

// -----
// Note that all of these are re-exported, and are only internal to prevent circular dependencies
// -----

// Account represents an on-chain account, with an associated signer, which must be a [crypto.Signer]
//
// Implements:
//   - [crypto.Signer]
type Account struct {
	Address AccountAddress
	Signer  crypto.Signer
}

// NewAccountFromSigner creates an account from a [crypto.Signer] with an optional [crypto.AuthenticationKey]
func NewAccountFromSigner(signer crypto.Signer, address ...AccountAddress) (*Account, error) {
	out := &Account{}
	switch len(address) {
	case 0:
		copy(out.Address[:], signer.AuthKey()[:])
	case 1:
		copy(out.Address[:], address[0][:])
	default:
		return nil, errors.New("must only provide one auth key")
	}
	out.Signer = signer
	return out, nil
}

// NewEd25519Account creates an account with a new random Ed25519 private key
func NewEd25519Account() (*Account, error) {
	privateKey, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		return nil, err
	}
	return NewAccountFromSigner(privateKey)
}

// NewEd25519SingleSignerAccount creates a new random Ed25519 account
func NewEd25519SingleSignerAccount() (*Account, error) {
	privateKey, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		return nil, err
	}
	signer := &crypto.SingleSigner{Signer: privateKey}
	return NewAccountFromSigner(signer)
}

// NewSecp256k1Account creates an account with a new random Secp256k1 private key
func NewSecp256k1Account() (*Account, error) {
	privateKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		return nil, err
	}
	signer := crypto.NewSingleSigner(privateKey)
	return NewAccountFromSigner(signer)
}

// MessageSigner extracts the message signer from the account for
func (account *Account) MessageSigner() (crypto.MessageSigner, bool) {
	ed25519PrivateKey, ok := account.Signer.(*crypto.Ed25519PrivateKey)
	if ok {
		return ed25519PrivateKey, ok
	}
	singleSigner, ok := account.Signer.(*crypto.SingleSigner)
	if ok {
		return singleSigner.Signer, ok
	}
	return nil, false
}

// PrivateKeyString extracts the private key as an AIP-80 compliant string
func (account *Account) PrivateKeyString() (string, error) {
	// Handle the key by itself
	ed25519PrivateKey, ok := account.Signer.(*crypto.Ed25519PrivateKey)
	if ok {
		return ed25519PrivateKey.ToAIP80()
	}

	// Handle key in single signer
	singleSigner, ok := account.Signer.(*crypto.SingleSigner)
	if ok {
		switch innerSigner := singleSigner.Signer.(type) {
		case *crypto.Ed25519PrivateKey:
			return innerSigner.ToAIP80()
		case *crypto.Secp256k1PrivateKey:
			return innerSigner.ToAIP80()
		}
	}

	return "", errors.New("signer is not a private key")
}

// Sign signs a message, returning an appropriate authenticator for the signer
func (account *Account) Sign(message []byte) (*crypto.AccountAuthenticator, error) {
	return account.Signer.Sign(message)
}

// SignMessage signs a message and returns the raw signature without a public key for verification
func (account *Account) SignMessage(message []byte) (crypto.Signature, error) {
	return account.Signer.SignMessage(message)
}

// SimulationAuthenticator creates a new authenticator for simulation purposes
func (account *Account) SimulationAuthenticator() *crypto.AccountAuthenticator {
	return account.Signer.SimulationAuthenticator()
}

// PubKey retrieves the public key for signature verification
func (account *Account) PubKey() crypto.PublicKey {
	return account.Signer.PubKey()
}

// AuthKey retrieves the authentication key associated with the signer
func (account *Account) AuthKey() *crypto.AuthenticationKey {
	return account.Signer.AuthKey()
}

// AccountAddress retrieves the account address
func (account *Account) AccountAddress() AccountAddress {
	return account.Address
}
