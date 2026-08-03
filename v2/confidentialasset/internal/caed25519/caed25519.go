// Package caed25519 matches @aptos-labs/confidential-asset TS utils (ed25519modN, random scalars, LE32).
package caed25519

import (
	"crypto/rand"
	"math/big"
)

// Order is the prime-order subgroup order of Ed25519 / Ristretto255 (same as TS ed25519.Point.CURVE().n).
var Order, _ = new(big.Int).SetString("1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3ed", 16)

// ModN reduces z modulo the curve order.
func ModN(z *big.Int) *big.Int {
	return new(big.Int).Mod(new(big.Int).Set(z), Order)
}

// BytesToNumberLE interprets 32 bytes as little-endian integer (TS bytesToNumberLE).
func BytesToNumberLE(b []byte) *big.Int {
	x := new(big.Int)
	for i := len(b) - 1; i >= 0; i-- {
		x.Lsh(x, 8)
		x.Or(x, big.NewInt(int64(b[i])))
	}
	return x
}

// NumberToBytesLE32 encodes z mod n as 32-byte little-endian (TS numberToBytesLE(_, 32)).
func NumberToBytesLE32(z *big.Int) []byte {
	v := ModN(z)
	out := make([]byte, 32)
	for i := 0; i < 32; i++ {
		out[i] = byte(new(big.Int).And(v, big.NewInt(0xff)).Uint64())
		v.Rsh(v, 8)
	}
	return out
}

// BytesToNumberBE interprets 32 bytes as big-endian (TS bytesToNumberBE for RNG scalars).
func BytesToNumberBE(b []byte) *big.Int {
	x := new(big.Int)
	for _, bb := range b {
		x.Lsh(x, 8)
		x.Or(x, big.NewInt(int64(bb)))
	}
	return x
}

// GenRandom returns a uniform scalar in [0, n) using rejection sampling (TS ed25519GenRandom).
func GenRandom() (*big.Int, error) {
	buf := make([]byte, 32)
	for {
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		r := BytesToNumberBE(buf)
		if r.Sign() >= 0 && r.Cmp(Order) < 0 {
			return r, nil
		}
	}
}

// GenListOfRandom returns k independent scalars (TS ed25519GenListOfRandom).
func GenListOfRandom(k int) ([]*big.Int, error) {
	out := make([]*big.Int, k)
	for i := 0; i < k; i++ {
		r, err := GenRandom()
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return out, nil
}
