//go:build cgo

package native

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
)

func ristrettoValRandBases() (valBase, randBase []byte) {
	return ca.BaseG.Encode(make([]byte, 0, 32)), ca.HRistretto.Encode(make([]byte, 0, 32))
}

func flattenBlindingsLE32(rs []*big.Int) []byte {
	out := make([]byte, 0, len(rs)*32)
	for _, r := range rs {
		out = append(out, caed25519.NumberToBytesLE32(r)...)
	}
	return out
}

func decodeHex32(s string) ([]byte, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	return b, nil
}

const ed25519OrderHex = "1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3ed"

func ed25519Order() *big.Int {
	n, _ := new(big.Int).SetString(ed25519OrderHex, 16)
	return n
}

func leBytesToBig(b []byte) *big.Int {
	x := new(big.Int)
	for i := range b {
		xi := new(big.Int).SetUint64(uint64(b[i]))
		xi.Lsh(xi, uint(8*i))
		x.Add(x, xi)
	}
	return x
}

func bigToLE32(x *big.Int) [32]byte {
	n := ed25519Order()
	v := new(big.Int).Mod(new(big.Int).Set(x), n)
	var out [32]byte
	for i := 0; i < 32; i++ {
		sh := new(big.Int).Rsh(v, uint(8*i))
		out[i] = byte(new(big.Int).And(sh, big.NewInt(0xff)).Uint64())
	}
	return out
}
