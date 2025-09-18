package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// -----
// Note that all of these are re-exported, and are only internal to prevent circular dependencies
// -----

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
var AccountOne = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}

// AccountTwo is [AccountAddress] 0x2
var AccountTwo = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02}

// AccountThree is [AccountAddress] 0x3
var AccountThree = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x03}

// AccountFour is [AccountAddress] 0x4
var AccountFour = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x04}

// AccountTen is [AccountAddress] 0xA
var AccountTen = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0A}

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
// These are AIP-40 compliant.
//
// Please use [AccountAddress.StringLong] for all indexer queries.
func (aa *AccountAddress) String() string {
	if aa.IsSpecial() {
		return fmt.Sprintf("0x%x", aa[31])
	}

	return util.BytesToHex(aa[:])
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

// StringShort Returns the short string representation of the AccountAddress
func (aa *AccountAddress) StringShort() string {
	msb := aa[0]
	msbIdx := 0
	for msb == 0 && msbIdx < 31 {
		msbIdx++
		msb = aa[msbIdx]
	}
	return fmt.Sprintf("0x%x%x", msb, aa[msbIdx+1:])
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
		return fmt.Errorf("failed to convert input to AccountAddress: %w", err)
	}
	err = aa.ParseStringRelaxed(str)
	if err != nil {
		return fmt.Errorf("failed to convert input to AccountAddress: %w", err)
	}
	return nil
}

// NamedObjectAddress derives a named object address based on the input address as the creator
func (aa *AccountAddress) NamedObjectAddress(seed []byte) AccountAddress {
	return aa.DerivedAddress(seed, crypto.NamedObjectScheme)
}

// ObjectAddressFromObject derives an object address based on the input address as the creator object
func (aa *AccountAddress) ObjectAddressFromObject(objectAddress *AccountAddress) AccountAddress {
	return aa.DerivedAddress(objectAddress[:], crypto.DeriveObjectScheme)
}

// ResourceAccount derives an object address based on the input address as the creator
func (aa *AccountAddress) ResourceAccount(seed []byte) AccountAddress {
	return aa.DerivedAddress(seed, crypto.ResourceAccountScheme)
}

// DerivedAddress addresses are derived by the address, the seed, then the type byte
func (aa *AccountAddress) DerivedAddress(seed []byte, typeByte uint8) AccountAddress {
	authKey := aa.AuthKey()
	authKey.FromBytesAndScheme(append(authKey[:], seed...), typeByte)
	return AccountAddress(authKey[:])
}

// ErrAddressMissing0x is returned when an AccountAddress is missing the leading 0x
var ErrAddressMissing0x = errors.New("AccountAddress missing 0x")

// ErrAddressTooShort is returned when an AccountAddress is too short
var ErrAddressTooShort = errors.New("AccountAddress too short")

// ErrAddressTooLong is returned when an AccountAddress is too long
var ErrAddressTooLong = errors.New("AccountAddress too long")

// ParseStringRelaxed parses a string into an AccountAddress
// TODO: add strict mode checking
func (aa *AccountAddress) ParseStringRelaxed(x string) error {
	x = strings.TrimPrefix(x, "0x")
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

// ParseStringWithPrefixRelaxed parses a string into an AccountAddress
func (aa *AccountAddress) ParseStringWithPrefixRelaxed(x string) error {
	if !strings.HasPrefix(x, "0x") {
		return ErrAddressMissing0x
	}
	x = x[2:]
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
