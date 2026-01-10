// Package account provides account creation and management for Aptos.
//
// This package provides utilities for creating and managing Aptos accounts,
// which are the primary way to interact with the Aptos blockchain.
//
// # Creating Accounts
//
// The simplest way to create an account is to generate a new random key:
//
//	// Generate a new Ed25519 account (legacy scheme, most compatible)
//	account, err := account.NewEd25519()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Generate a Secp256k1 account (for Ethereum compatibility)
//	account, err := account.NewSecp256k1()
//
// # Importing Accounts
//
// You can import accounts from existing private keys:
//
//	// From hex-encoded private key
//	account, err := account.FromPrivateKeyHex("0x...")
//
//	// From AIP-80 formatted string
//	account, err := account.FromAIP80("ed25519-priv-...")
//
// # Using Accounts
//
// Accounts implement TransactionSigner and can be used directly with the client:
//
//	client, _ := aptos.NewClient(aptos.Testnet)
//	account, _ := account.NewEd25519()
//
//	// Fund the account (testnet only)
//	client.Fund(ctx, account.Address(), 100_000_000)
//
//	// Build and submit a transaction
//	payload := &aptos.EntryFunctionPayload{...}
//	result, err := client.SignAndSubmitTransaction(ctx, account, payload)
package account

import (
	"errors"
	"fmt"
	"io"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
)

// AccountAddress is re-exported for convenience.
type AccountAddress = types.AccountAddress

// Account represents an on-chain account with signing capability.
// It combines an AccountAddress with a Signer to allow both
// address-based lookups and transaction signing.
//
// Implements:
//   - [aptos.TransactionSigner]
//   - [aptos.Signer] (via embedded Signer)
type Account struct {
	address types.AccountAddress
	signer  crypto.Signer
}

// Address returns the account's on-chain address.
func (a *Account) Address() types.AccountAddress {
	return a.address
}

// Sign signs a message and returns an AccountAuthenticator.
func (a *Account) Sign(msg []byte) (*crypto.AccountAuthenticator, error) {
	return a.signer.Sign(msg)
}

// SignMessage signs a message and returns the raw signature.
func (a *Account) SignMessage(msg []byte) (crypto.Signature, error) {
	return a.signer.SignMessage(msg)
}

// SimulationAuthenticator returns an authenticator for simulation.
func (a *Account) SimulationAuthenticator() *crypto.AccountAuthenticator {
	return a.signer.SimulationAuthenticator()
}

// AuthKey returns the authentication key for this account.
func (a *Account) AuthKey() *crypto.AuthenticationKey {
	return a.signer.AuthKey()
}

// PubKey returns the public key for this account.
func (a *Account) PubKey() crypto.PublicKey {
	return a.signer.PubKey()
}

// Signer returns the underlying signer.
func (a *Account) Signer() crypto.Signer {
	return a.signer
}

// NewEd25519 creates a new account with a randomly generated Ed25519 key.
// This uses the legacy Ed25519Scheme and is the most widely compatible.
//
// An optional io.Reader can be provided for deterministic key generation
// (useful for testing). The reader must provide exactly 32 bytes.
func NewEd25519(rand ...io.Reader) (*Account, error) {
	key, err := crypto.GenerateEd25519PrivateKey(rand...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}
	return FromSigner(key)
}

// NewSecp256k1 creates a new account with a randomly generated Secp256k1 key.
// This uses the SingleKeyScheme and is compatible with Ethereum keys.
func NewSecp256k1() (*Account, error) {
	key, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Secp256k1 key: %w", err)
	}
	signer := crypto.NewSingleSigner(key)
	return FromSigner(signer)
}

// NewEd25519SingleKey creates a new account with an Ed25519 key using
// the SingleKeyScheme instead of the legacy Ed25519Scheme.
// This is useful for unified key handling.
//
// An optional io.Reader can be provided for deterministic key generation.
func NewEd25519SingleKey(rand ...io.Reader) (*Account, error) {
	key, err := crypto.GenerateEd25519PrivateKey(rand...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}
	signer := crypto.NewSingleSigner(key)
	return FromSigner(signer)
}

// FromSigner creates an account from any Signer implementation.
// The account address is derived from the signer's authentication key.
func FromSigner(signer crypto.Signer) (*Account, error) {
	authKey := signer.AuthKey()
	var address types.AccountAddress
	copy(address[:], authKey[:])
	return &Account{
		address: address,
		signer:  signer,
	}, nil
}

// FromSignerWithAddress creates an account from a Signer with a specific address.
// Use this when the account address differs from the derived auth key
// (e.g., after auth key rotation).
func FromSignerWithAddress(signer crypto.Signer, address types.AccountAddress) *Account {
	return &Account{
		address: address,
		signer:  signer,
	}
}

// FromEd25519PrivateKey creates an account from Ed25519 private key bytes.
// The bytes should be 32 bytes (seed) or 64 bytes (full key).
func FromEd25519PrivateKey(privateKeyBytes []byte) (*Account, error) {
	key := &crypto.Ed25519PrivateKey{}
	if err := key.FromBytes(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("invalid Ed25519 private key: %w", err)
	}
	return FromSigner(key)
}

// FromSecp256k1PrivateKey creates an account from Secp256k1 private key bytes.
// The bytes should be 32 bytes.
func FromSecp256k1PrivateKey(privateKeyBytes []byte) (*Account, error) {
	key := &crypto.Secp256k1PrivateKey{}
	if err := key.FromBytes(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("invalid Secp256k1 private key: %w", err)
	}
	signer := crypto.NewSingleSigner(key)
	return FromSigner(signer)
}

// FromPrivateKeyHex creates an account from a hex-encoded private key.
// The key type is automatically detected based on length.
func FromPrivateKeyHex(hexKey string) (*Account, error) {
	// Try Ed25519 first (32 bytes seed)
	ed25519Key := &crypto.Ed25519PrivateKey{}
	if err := ed25519Key.FromHex(hexKey); err == nil {
		return FromSigner(ed25519Key)
	}

	// Try Secp256k1 (32 bytes)
	secp256k1Key := &crypto.Secp256k1PrivateKey{}
	if err := secp256k1Key.FromHex(hexKey); err == nil {
		signer := crypto.NewSingleSigner(secp256k1Key)
		return FromSigner(signer)
	}

	return nil, errors.New("failed to parse private key: unsupported format")
}

// FromAIP80 creates an account from an AIP-80 formatted private key string.
// AIP-80 format: "ed25519-priv-..." or "secp256k1-priv-..."
func FromAIP80(aip80Key string) (*Account, error) {
	// Detect key type from prefix
	var keyType crypto.PrivateKeyVariant
	switch {
	case len(aip80Key) > 13 && aip80Key[:13] == "ed25519-priv-":
		keyType = crypto.PrivateKeyVariantEd25519
	case len(aip80Key) > 15 && aip80Key[:15] == "secp256k1-priv-":
		keyType = crypto.PrivateKeyVariantSecp256k1
	default:
		return nil, errors.New("invalid AIP-80 key: unrecognized prefix")
	}

	keyBytes, err := crypto.ParsePrivateKey(aip80Key, keyType, true)
	if err != nil {
		return nil, fmt.Errorf("invalid AIP-80 key: %w", err)
	}

	switch keyType {
	case crypto.PrivateKeyVariantEd25519:
		ed25519Key := &crypto.Ed25519PrivateKey{}
		if err := ed25519Key.FromBytes(keyBytes); err != nil {
			return nil, fmt.Errorf("invalid Ed25519 key bytes: %w", err)
		}
		return FromSigner(ed25519Key)

	case crypto.PrivateKeyVariantSecp256k1:
		secp256k1Key := &crypto.Secp256k1PrivateKey{}
		if err := secp256k1Key.FromBytes(keyBytes); err != nil {
			return nil, fmt.Errorf("invalid Secp256k1 key bytes: %w", err)
		}
		signer := crypto.NewSingleSigner(secp256k1Key)
		return FromSigner(signer)

	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// ToAIP80 returns the account's private key as an AIP-80 formatted string.
// Returns an error if the underlying key type doesn't support AIP-80 format.
func (a *Account) ToAIP80() (string, error) {
	// Handle Ed25519 directly
	if ed25519Key, ok := a.signer.(*crypto.Ed25519PrivateKey); ok {
		return ed25519Key.ToAIP80()
	}

	// Handle Secp256k1 directly (wrapped in SingleSigner)
	// SingleSigner's inner field is private, so we need to check if the
	// signer produces a Secp256k1 public key

	// For now, we need to type assert or use a different approach
	// The cleanest solution is to add a method to extract the private key
	// but that requires changes to internal/crypto

	return "", errors.New("signer does not support AIP-80 format export")
}

// String returns a human-readable representation of the account.
func (a *Account) String() string {
	return fmt.Sprintf("Account{address: %s}", a.address.String())
}
