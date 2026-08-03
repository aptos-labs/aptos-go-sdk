package confidentialasset

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

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

func twistedKeyFromSigBig(sig64 []byte) ([32]byte, error) {
	var out [32]byte
	if len(sig64) != 64 {
		return out, fmt.Errorf("expected 64-byte signature")
	}
	scalarLE := leBytesToBig(sig64)
	scalarLE.Mod(scalarLE, ed25519Order())
	le := bigToLE32(scalarLE)
	copy(out[:], le[:])
	return out, nil
}
