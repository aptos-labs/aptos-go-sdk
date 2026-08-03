package sigbcs

import (
	"bytes"
	"testing"
)

func TestAppendULEB128_golden(t *testing.T) {
	t.Parallel()
	cases := []struct {
		v    uint32
		want []byte
	}{
		{0, []byte{0x00}},
		{127, []byte{0x7f}},
		{128, []byte{0x80, 0x01}},
		{16383, []byte{0xff, 0x7f}},
		{16384, []byte{0x80, 0x80, 0x01}},
	}
	for _, tc := range cases {
		got := AppendULEB128(nil, tc.v)
		if !bytes.Equal(got, tc.want) {
			t.Fatalf("ULEB128(%d): got %x want %x", tc.v, got, tc.want)
		}
	}
}

func TestAppendBytes_roundTripLength(t *testing.T) {
	t.Parallel()
	payload := []byte{0xde, 0xad}
	got := AppendBytes(nil, payload)
	if len(got) < len(payload) {
		t.Fatal("short")
	}
	// First byte(s) are ULEB128(len); 2 encodes as single 0x02 for len<128
	if got[0] != 2 || !bytes.Equal(got[1:], payload) {
		t.Fatalf("got %x", got)
	}
}

func TestAppendBool(t *testing.T) {
	t.Parallel()
	if b := AppendBool(nil, true); len(b) != 1 || b[0] != 1 {
		t.Fatalf("true: %x", b)
	}
	if b := AppendBool(nil, false); len(b) != 1 || b[0] != 0 {
		t.Fatalf("false: %x", b)
	}
}
