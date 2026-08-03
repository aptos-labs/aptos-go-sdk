package ca

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/gtank/ristretto255"
)

// HASH_BASE_POINT is TS twistedEd25519.HASH_BASE_POINT (SHA3-512 hash of basepoint, hex).
const HASH_BASE_POINT = "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"

// HRistretto is the encryption-key basepoint H (TS H_RISTRETTO).
var HRistretto = func() *ristretto255.Element {
	raw, err := hex.DecodeString(HASH_BASE_POINT)
	if err != nil || len(raw) != 32 {
		panic("ca: bad H_RISTRETTO hex")
	}
	var p ristretto255.Element
	if err := p.Decode(raw); err != nil {
		panic(err)
	}
	return &p
}()

// BaseG is the Ristretto255 basepoint (TS ristretto255.Point.BASE).
var BaseG = func() *ristretto255.Element {
	var g ristretto255.Element
	return g.Base()
}()

// TwistedPublicKeyFromPrivateLE32 returns ek = (LE32(dk)^{-1}) * H (TS TwistedEd25519PrivateKey.publicKey).
func TwistedPublicKeyFromPrivateLE32(dk32 [32]byte) ([]byte, error) {
	dk := caed25519.ModN(caed25519.BytesToNumberLE(dk32[:]))
	var dkSc ristretto255.Scalar
	if err := dkSc.Decode(caed25519.NumberToBytesLE32(dk)); err != nil {
		return nil, err
	}
	var inv ristretto255.Scalar
	inv.Invert(&dkSc)
	var out ristretto255.Element
	out.ScalarMult(&inv, HRistretto)
	return out.Encode(make([]byte, 0, 32)), nil
}

// ScalarMultBytes decodes 32-byte compressed point and multiplies by bigint witness (mod n).
func ScalarMultBytes(pt []byte, witness *big.Int) (*ristretto255.Element, error) {
	var P ristretto255.Element
	if err := P.Decode(pt); err != nil {
		return nil, err
	}
	var s ristretto255.Scalar
	if err := s.Decode(caed25519.NumberToBytesLE32(witness)); err != nil {
		return nil, fmt.Errorf("scalar decode: %w", err)
	}
	var out ristretto255.Element
	out.ScalarMult(&s, &P)
	return &out, nil
}

// ScalarMultElement is P * witness (witness as mod-n scalar).
func ScalarMultElement(P *ristretto255.Element, witness *big.Int) (*ristretto255.Element, error) {
	var s ristretto255.Scalar
	if err := s.Decode(caed25519.NumberToBytesLE32(witness)); err != nil {
		return nil, err
	}
	var out ristretto255.Element
	out.ScalarMult(&s, P)
	return &out, nil
}
