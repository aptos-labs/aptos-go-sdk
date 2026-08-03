package ca

import (
	"math/big"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
)

func TestEncryptTwistedElGamal(t *testing.T) {
	t.Parallel()
	pub, err := TwistedPublicKeyFromPrivateLE32([32]byte{1})
	if err != nil {
		t.Fatal(err)
	}
	r := caed25519.ModN(big.NewInt(7))
	c, d, err := EncryptTwistedElGamal(42, pub, r)
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 32 || len(d) != 32 {
		t.Fatalf("c=%d d=%d", len(c), len(d))
	}
}
