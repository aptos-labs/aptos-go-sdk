package sigbcs

import "testing"

func TestAppendU8_and_U64LE(t *testing.T) {
	t.Parallel()
	b := AppendU8(nil, 0xab)
	if len(b) != 1 || b[0] != 0xab {
		t.Fatalf("u8=%v", b)
	}
	b = AppendU64LE(b, 0x0102030405060708)
	if len(b) != 9 {
		t.Fatalf("len=%d", len(b))
	}
}
