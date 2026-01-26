// Package types contains core types used throughout the Aptos Go SDK.
package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// AccountAddress is a 32-byte address on the Aptos blockchain.
// It can represent an account, an object, or other on-chain entities.
type AccountAddress [32]byte

// Common errors for address parsing.
var (
	ErrAddressTooShort   = errors.New("address too short")
	ErrAddressTooLong    = errors.New("address too long")
	ErrAddressInvalidHex = errors.New("address contains invalid hex characters")
)

// ParseAddress parses a hex string into an AccountAddress.
// It accepts addresses with or without the "0x" prefix, and handles
// both short (special) addresses like "0x1" and full 64-character addresses.
func ParseAddress(s string) (AccountAddress, error) {
	var addr AccountAddress

	// Remove 0x prefix if present
	s = strings.TrimPrefix(s, "0x")

	if len(s) == 0 {
		return addr, ErrAddressTooShort
	}
	if len(s) > 64 {
		return addr, ErrAddressTooLong
	}

	// Pad with leading zeros if necessary
	if len(s)%2 != 0 {
		s = "0" + s
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return addr, fmt.Errorf("%w: %w", ErrAddressInvalidHex, err)
	}

	// Copy right-aligned (addresses are big-endian)
	copy(addr[32-len(bytes):], bytes)
	return addr, nil
}

// MustParseAddress parses a hex string into an AccountAddress, panicking on error.
// This is useful for compile-time constants.
func MustParseAddress(s string) AccountAddress {
	addr, err := ParseAddress(s)
	if err != nil {
		panic(fmt.Sprintf("invalid address %q: %v", s, err))
	}
	return addr
}

// IsSpecial returns true if this is a "special" address (0x0 through 0xf).
// Special addresses use short-form string representation per AIP-40.
func (a AccountAddress) IsSpecial() bool {
	for i := 0; i < 31; i++ {
		if a[i] != 0 {
			return false
		}
	}
	return a[31] < 0x10
}

// String returns the canonical string representation of the address.
// Special addresses (0x0-0xf) use short form, others use full 64-character hex.
func (a AccountAddress) String() string {
	if a.IsSpecial() {
		return fmt.Sprintf("0x%x", a[31])
	}
	return "0x" + hex.EncodeToString(a[:])
}

// StringLong returns the full 64-character hex representation.
// This is required for indexer queries.
func (a AccountAddress) StringLong() string {
	return "0x" + hex.EncodeToString(a[:])
}

// StringShort returns the shortest possible hex representation.
// Leading zeros are stripped (except the last byte for non-zero addresses).
func (a AccountAddress) StringShort() string {
	// Find first non-zero byte
	firstNonZero := 0
	for i := 0; i < 32; i++ {
		if a[i] != 0 {
			firstNonZero = i
			break
		}
		if i == 31 {
			return "0x0"
		}
	}
	return "0x" + hex.EncodeToString(a[firstNonZero:])
}

// Bytes returns a copy of the address bytes.
func (a AccountAddress) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, a[:])
	return b
}

// IsZero returns true if this is the zero address.
func (a AccountAddress) IsZero() bool {
	return a == AccountAddress{}
}

// MarshalJSON implements json.Marshaler.
func (a AccountAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AccountAddress) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("address must be a string: %w", err)
	}
	addr, err := ParseAddress(s)
	if err != nil {
		return err
	}
	*a = addr
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (a AccountAddress) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (a *AccountAddress) UnmarshalText(text []byte) error {
	addr, err := ParseAddress(string(text))
	if err != nil {
		return err
	}
	*a = addr
	return nil
}

// MarshalBCS serializes the address to BCS (32 fixed bytes).
//
// Implements [bcs.Marshaler].
func (a *AccountAddress) MarshalBCS(ser *bcs.Serializer) {
	ser.FixedBytes(a[:])
}

// UnmarshalBCS deserializes the address from BCS.
//
// Implements [bcs.Unmarshaler].
func (a *AccountAddress) UnmarshalBCS(des *bcs.Deserializer) {
	des.ReadFixedBytesInto(a[:])
}

// Common special addresses.
var (
	AccountZero  = MustParseAddress("0x0")
	AccountOne   = MustParseAddress("0x1")
	AccountTwo   = MustParseAddress("0x2")
	AccountThree = MustParseAddress("0x3")
	AccountFour  = MustParseAddress("0x4")
	AccountTen   = MustParseAddress("0xa")
)
