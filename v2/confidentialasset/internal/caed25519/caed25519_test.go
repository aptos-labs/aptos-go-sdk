package caed25519

import (
	"math/big"
	"testing"
)

func TestModN_and_bytesRoundTrip(t *testing.T) {
	t.Parallel()
	x := ModN(new(big.Int).SetUint64(999999))
	le := NumberToBytesLE32(x)
	back := BytesToNumberLE(le)
	if ModN(back).Cmp(x) != 0 {
		t.Fatalf("back=%v x=%v", back, x)
	}
}

func TestBytesToNumberBE(t *testing.T) {
	t.Parallel()
	n := BytesToNumberBE([]byte{0, 1, 2})
	if n.Uint64() != 0x102 {
		t.Fatalf("got %v", n)
	}
}

func TestGenListOfRandom(t *testing.T) {
	t.Parallel()
	rs, err := GenListOfRandom(4)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 4 {
		t.Fatalf("len=%d", len(rs))
	}
}

func TestGenRandom(t *testing.T) {
	t.Parallel()
	r, err := GenRandom()
	if err != nil || r == nil {
		t.Fatalf("r=%v err=%v", r, err)
	}
}
