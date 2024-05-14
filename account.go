package aptos

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"strings"
)

// AccountAddress a 32-byte representation of an onchain address
type AccountAddress [32]byte

var AccountZero = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var AccountOne = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
var AccountTwo = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
var AccountThree = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
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

// String Returns the canonical string representation of the AccountAddress
func (aa *AccountAddress) String() string {
	if aa.IsSpecial() {
		return fmt.Sprintf("0x%x", aa[31])
	} else {
		return "0x" + hex.EncodeToString(aa[:])
	}
}

// MarshalBCS Converts the AccountAddress to BCS encoded bytes
func (aa AccountAddress) MarshalBCS(bcs *bcs.Serializer) {
	bcs.FixedBytes(aa[:])
}

// UnmarshalBCS Converts the AccountAddress from BCS encoded bytes
func (aa *AccountAddress) UnmarshalBCS(bcs *bcs.Deserializer) {
	bcs.ReadFixedBytesInto((*aa)[:])
}

// Random generates a random account address, mainly for testing
func (aa *AccountAddress) Random() {
	rand.Read((*aa)[:])
}

// FromPublicKey Generates an account address from a public key using the appropriate
// account address scheme
func (aa *AccountAddress) FromPublicKey(pubkey crypto.PublicKey) {
	bytes := util.SHA3_256Hash([][]byte{
		pubkey.Bytes(),
		{pubkey.Scheme()},
	})
	copy((*aa)[:], bytes)
}

// NamedObjectAddress derives a named object address based on the input address as the creator
func (aa *AccountAddress) NamedObjectAddress(seed []byte) (accountAddress AccountAddress) {
	return aa.DerivedAddress(seed, crypto.NamedObjectScheme)
}

// ObjectAddressFromObject derives a object address based on the input address as the creator object
func (aa *AccountAddress) ObjectAddressFromObject(objectAddress *AccountAddress) (accountAddress AccountAddress) {
	return aa.DerivedAddress(objectAddress[:], crypto.DeriveObjectScheme)
}

// ResourceAccount derives a object address based on the input address as the creator
func (aa *AccountAddress) ResourceAccount(seed []byte) (accountAddress AccountAddress) {
	return aa.DerivedAddress(seed, crypto.ResourceAccountScheme)
}

// DerivedAddress addresses are derived by the address, the seed, then the type byte
func (aa *AccountAddress) DerivedAddress(seed []byte, typeByte uint8) (accountAddress AccountAddress) {
	bytes := util.SHA3_256Hash([][]byte{
		aa[:],
		seed[:],
		{typeByte},
	})
	copy(accountAddress[:], bytes)
	return
}

// Account represents an onchain account, with an associated signer, which may be a PrivateKey
type Account struct {
	Address AccountAddress
	Signer  crypto.Signer
}

// NewEd25519Account creates an account with a new random Ed25519 private key
func NewEd25519Account() (*Account, error) {
	privkey, pubkey, err := crypto.GenerateEd5519Keys()
	if err != nil {
		return nil, err
	}
	out := &Account{}
	out.Address.FromPublicKey(&pubkey)
	out.Signer = &privkey
	return out, nil
}

// Sign signs a message, returning an appropriate authenticator for the signer
func (account *Account) Sign(message []byte) (authenticator crypto.Authenticator, err error) {
	return account.Signer.Sign(message)
}

var ErrAddressTooShort = errors.New("AccountAddress too short")
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
	abytes, err := hex.DecodeString(x)
	if err != nil {
		return err
	}
	// zero-prefix/right-align what bytes we got
	copy((*aa)[32-len(abytes):], abytes)

	return nil
}