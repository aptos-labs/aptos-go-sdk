//go:build cgo

package native

import (
	"math/big"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
)

func Test_decodeHex32(t *testing.T) {
	b, err := decodeHex32("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	if err != nil || len(b) != 32 {
		t.Fatalf("err=%v len=%d", err, len(b))
	}
}

func Test_flattenBlindingsLE32(t *testing.T) {
	out := flattenBlindingsLE32([]*big.Int{caed25519.ModN(big.NewInt(1))})
	if len(out) != 32 {
		t.Fatalf("len=%d", len(out))
	}
}

func Test_ristrettoValRandBases(t *testing.T) {
	v, r := ristrettoValRandBases()
	if len(v) != 32 || len(r) != 32 {
		t.Fatalf("v=%d r=%d", len(v), len(r))
	}
}

func Test_isAllZero32(t *testing.T) {
	if !isAllZero32(make([]byte, 32)) {
		t.Fatal("zeros")
	}
	if isAllZero32([]byte{1}) {
		t.Fatal("non-zero")
	}
}
