package movearg

import (
	"bytes"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigbcs"
)

func TestVectorU8_golden(t *testing.T) {
	t.Parallel()
	got := VectorU8([]byte{0xab, 0xcd})
	want := sigbcs.AppendULEB128(nil, 2)
	want = append(want, 0xab, 0xcd)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}

func TestVectorVectorU8_golden(t *testing.T) {
	t.Parallel()
	got := VectorVectorU8([][]byte{{0x01}, {0x02, 0x03}})
	want := []byte{0x02}
	want = append(want, sigbcs.AppendULEB128(nil, 1)...)
	want = append(want, 0x01)
	want = append(want, sigbcs.AppendULEB128(nil, 2)...)
	want = append(want, 0x02, 0x03)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}

func TestVectorTripleVecU8_empty(t *testing.T) {
	t.Parallel()
	got := VectorTripleVecU8(nil)
	want := sigbcs.AppendULEB128(nil, 0)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}
