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

	// Handle odd-length hex strings without string concatenation
	// by decoding into a temporary buffer with proper offset
	hexLen := len(s)
	byteLen := (hexLen + 1) / 2
	offset := 32 - byteLen

	if hexLen%2 != 0 {
		// Odd length: decode first nibble separately, then rest
		firstNibble, ok := hexDigitValue(s[0])
		if !ok {
			return addr, fmt.Errorf("%w: invalid character at position 0", ErrAddressInvalidHex)
		}
		addr[offset] = firstNibble
		if hexLen > 1 {
			_, err := hex.Decode(addr[offset+1:], []byte(s[1:]))
			if err != nil {
				return addr, fmt.Errorf("%w: %w", ErrAddressInvalidHex, err)
			}
		}
	} else {
		// Even length: decode directly
		_, err := hex.Decode(addr[offset:], []byte(s))
		if err != nil {
			return addr, fmt.Errorf("%w: %w", ErrAddressInvalidHex, err)
		}
	}

	return addr, nil
}

// hexDigitValue returns the numeric value of a hex digit.
func hexDigitValue(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
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

// specialAddressStrings contains pre-computed strings for special addresses 0x0-0xf.
var specialAddressStrings = [16]string{
	"0x0", "0x1", "0x2", "0x3", "0x4", "0x5", "0x6", "0x7",
	"0x8", "0x9", "0xa", "0xb", "0xc", "0xd", "0xe", "0xf",
}

// String returns the canonical string representation of the address.
// Special addresses (0x0-0xf) use short form, others use full 64-character hex.
func (a AccountAddress) String() string {
	if a.IsSpecial() {
		return specialAddressStrings[a[31]]
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
