// Package types is an internal package that contains the types used in the Aptos Go SDK.
//
// It was purposely built because of circular dependencies between packages, and it is re-exported directly in the aptos-go-sdk package.
package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"strings"
)

// AccountAddress a 32-byte representation of an on-chain address
//
// Implements:
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [json.Marshaler]
//   - [json.Unmarshaler]
type AccountAddress [32]byte

// AccountZero is [AccountAddress] 0x0
var AccountZero = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// AccountOne is [AccountAddress] 0x1
var AccountOne = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

// AccountTwo is [AccountAddress] 0x2
var AccountTwo = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}

// AccountThree is [AccountAddress] 0x3
var AccountThree = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}

// AccountFour is [AccountAddress] 0x4
var AccountFour = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}

// IsSpecial Returns whether the address is a "special" address. Addresses are considered
// special if the first 63 characters of the hex string are zero. In other words,
// an address is special if the first 31 bytes are zero and the last byte is
// smaller than `0b10000` (16). In other words, special is defined as an address
// that matches the following regex: `^0x0{63}[0-9a-f]$`. In short form this means
// the addresses in the range from `0x0` to `0xf` (inclusive) are special.
// For more details see the v1 address standard defined as part of AIP-40:
// https://github.com/aptos-foundation/AIPs/blob/main/aips/aip-40.md
func (aa *AccountAddress) IsSpecial() bool {
	for _, b := range aa[:31] {
		if b != 0 {
			return false
		}
	}
	return aa[31] < 0x10
}

// String Returns the canonical string representation of the [AccountAddress]
//
// Please use [AccountAddress.StringLong] for all indexer queries.
func (aa *AccountAddress) String() string {
	if aa.IsSpecial() {
		return fmt.Sprintf("0x%x", aa[31])
	} else {
		return util.BytesToHex(aa[:])
	}
}

// FromAuthKey converts [crypto.AuthenticationKey] to [AccountAddress]
func (aa *AccountAddress) FromAuthKey(authKey *crypto.AuthenticationKey) {
	copy(aa[:], authKey[:])
}

// AuthKey converts [AccountAddress] to [crypto.AuthenticationKey]
func (aa *AccountAddress) AuthKey() *crypto.AuthenticationKey {
	authKey := &crypto.AuthenticationKey{}
	copy(authKey[:], aa[:])
	return authKey
}

// StringLong Returns the long string representation of the AccountAddress
//
// This is most commonly used for all indexer queries.
func (aa *AccountAddress) StringLong() string {
	return util.BytesToHex(aa[:])
}

// MarshalBCS Converts the AccountAddress to BCS encoded bytes
func (aa *AccountAddress) MarshalBCS(ser *bcs.Serializer) {
	ser.FixedBytes(aa[:])
}

// UnmarshalBCS Converts the AccountAddress from BCS encoded bytes
func (aa *AccountAddress) UnmarshalBCS(des *bcs.Deserializer) {
	des.ReadFixedBytesInto((*aa)[:])
}

// MarshalJSON converts the AccountAddress to JSON
func (aa *AccountAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(aa.String())
}

// UnmarshalJSON converts the AccountAddress from JSON
func (aa *AccountAddress) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return fmt.Errorf("failed to convert input to AccountAdddress: %w", err)
	}
	err = aa.ParseStringRelaxed(str)
	if err != nil {
		return fmt.Errorf("failed to convert input to AccountAdddress: %w", err)
	}
	return nil
}

// NamedObjectAddress derives a named object address based on the input address as the creator
func (aa *AccountAddress) NamedObjectAddress(seed []byte) (accountAddress AccountAddress) {
	return aa.DerivedAddress(seed, crypto.NamedObjectScheme)
}

// ObjectAddressFromObject derives an object address based on the input address as the creator object
func (aa *AccountAddress) ObjectAddressFromObject(objectAddress *AccountAddress) (accountAddress AccountAddress) {
	return aa.DerivedAddress(objectAddress[:], crypto.DeriveObjectScheme)
}

// ResourceAccount derives an object address based on the input address as the creator
func (aa *AccountAddress) ResourceAccount(seed []byte) (accountAddress AccountAddress) {
	return aa.DerivedAddress(seed, crypto.ResourceAccountScheme)
}

// DerivedAddress addresses are derived by the address, the seed, then the type byte
func (aa *AccountAddress) DerivedAddress(seed []byte, typeByte uint8) (accountAddress AccountAddress) {
	authKey := aa.AuthKey()
	authKey.FromBytesAndScheme(append(authKey[:], seed[:]...), typeByte)
	copy(accountAddress[:], authKey[:])
	return
}

// Account represents an on-chain account, with an associated signer, which must be a [crypto.Signer]
//
// Implements:
//   - [crypto.Signer]
type Account struct {
	Address AccountAddress
	Signer  crypto.Signer
}

// NewAccountFromSigner creates an account from a [crypto.Signer] with an optional [crypto.AuthenticationKey]
func NewAccountFromSigner(signer crypto.Signer, authKey ...crypto.AuthenticationKey) (*Account, error) {
	out := &Account{}
	if len(authKey) == 1 {
		copy(out.Address[:], authKey[0][:])
	} else if len(authKey) > 1 {
		// Throw error
		return nil, errors.New("must only provide one auth key")
	} else {
		copy(out.Address[:], signer.AuthKey()[:])
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

// Sign signs a message, returning an appropriate authenticator for the signer
func (account *Account) Sign(message []byte) (authenticator *crypto.AccountAuthenticator, err error) {
	return account.Signer.Sign(message)
}

// SignMessage signs a message and returns the raw signature without a public key for verification
func (account *Account) SignMessage(message []byte) (signature crypto.Signature, err error) {
	return account.Signer.SignMessage(message)
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

// ErrAddressTooShort is returned when an AccountAddress is too short
var ErrAddressTooShort = errors.New("AccountAddress too short")

// ErrAddressTooLong is returned when an AccountAddress is too long
var ErrAddressTooLong = errors.New("AccountAddress too long")

// ParseStringRelaxed parses a string into an AccountAddress
// TODO: add strict mode checking
func (aa *AccountAddress) ParseStringRelaxed(x string) error {
	if strings.HasPrefix(x, "0x") {
		x = x[2:]
	}
	if len(x) < 1 {
		return ErrAddressTooShort
	}
	if len(x) > 64 {
		return ErrAddressTooLong
	}
	if len(x)%2 != 0 {
		x = "0" + x
	}
	bytes, err := hex.DecodeString(x)
	if err != nil {
		return err
	}
	// zero-prefix/right-align what bytes we got
	copy((*aa)[32-len(bytes):], bytes)

	return nil
}
