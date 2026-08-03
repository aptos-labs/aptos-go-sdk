package ca

import (
	"math/big"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
)

func TestTwistedPublicKeyFromPrivateLE32_deterministic(t *testing.T) {
	t.Parallel()
	var dk [32]byte
	dk[0] = 9
	a, err := TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	b, err := TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatal("not deterministic")
	}
}

func TestScalarMultBytes(t *testing.T) {
	t.Parallel()
	g := BaseG.Encode(make([]byte, 0, 32))
	out, err := ScalarMultBytes(g, caed25519.ModN(big.NewInt(2)))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Encode(make([]byte, 0, 32))) != 32 {
		t.Fatal("bad encoding")
	}
}

func TestScalarMultElement(t *testing.T) {
	t.Parallel()
	w := caed25519.ModN(big.NewInt(3))
	out, err := ScalarMultElement(BaseG, w)
	if err != nil {
		t.Fatal(err)
	}
	if out.Equal(BaseG) == 1 && w.Cmp(big.NewInt(1)) != 0 {
		t.Fatal("unexpected equality")
	}
}
