package aptos

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type AccountAddress [32]byte

var Account0x1 AccountAddress = AccountAddress{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

// Returns whether the address is a "special" address. Addresses are considered
// special if the first 63 characters of the hex string are zero. In other words,
// an address is special if the first 31 bytes are zero and the last byte is
// smaller than `0b10000` (16). In other words, special is defined as an address
// that matches the following regex: `^0x0{63}[0-9a-f]$`. In short form this means
// the addresses in the range from `0x0` to `0xf` (inclusive) are special.

// For more details see the v1 address standard defined as part of AIP-40:
// https://github.com/aptos-foundation/AIPs/blob/main/aips/aip-40.md
func (aa AccountAddress) IsSpecial() bool {
	for _, b := range aa[:31] {
		if b != 0 {
			return false
		}
	}
	return aa[31] < 0x10
}

func (aa AccountAddress) String() string {
	if aa.IsSpecial() {
		return fmt.Sprintf("0x%x", aa[31])
	} else {
		return "0x" + hex.EncodeToString(aa[:])
	}
}

func (aa AccountAddress) MarshalBCS(bcs *Serializer) {
	bcs.FixedBytes(aa[:])
}
func (aa *AccountAddress) UnmarshalBCS(bcs *Deserializer) {
	bcs.ReadFixedBytesInto((*aa)[:])
}
func (aa *AccountAddress) Random() {
	rand.Read((*aa)[:])
}

type Account struct {
	Address    AccountAddress
	PrivateKey crypto.PrivateKey
}

func NewAccount() (*Account, error) {
	pubkey, privkey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}
	out := &Account{}
	copy(out.Address[:], pubkey)
	out.PrivateKey = privkey
	return out, nil
}

var ErrAddressTooShort = errors.New("AccountAddress too short")
var ErrAddressTooLong = errors.New("AccountAddress too long")

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
