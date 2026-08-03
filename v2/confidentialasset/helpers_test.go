package confidentialasset

import (
	"bytes"
	"math/big"
	"testing"
)

func Test_decodeHex32(t *testing.T) {
	t.Parallel()
	b, err := decodeHex32("0x" + testPointP)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 32 {
		t.Fatalf("len=%d", len(b))
	}
	if _, err := decodeHex32("0x0102"); err == nil {
		t.Fatal("expected error for short hex")
	}
}

func Test_twistedKeyFromSigBig(t *testing.T) {
	t.Parallel()
	sig := make([]byte, 64)
	sig[0] = 1
	got, err := twistedKeyFromSigBig(sig)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(got[:], make([]byte, 32)) {
		t.Fatal("expected non-zero key")
	}
	if _, err := twistedKeyFromSigBig([]byte{1, 2}); err == nil {
		t.Fatal("expected error for short sig")
	}
}

func Test_bigToLE32_leBytesToBig_roundTrip(t *testing.T) {
	t.Parallel()
	x := big.NewInt(123456789)
	le := bigToLE32(x)
	back := leBytesToBig(le[:])
	if back.Cmp(x) != 0 {
		t.Fatalf("got %v want %v", back, x)
	}
}
