package bcs

import (
	"runtime"
	"testing"
)

// uleb encodes n as a BCS ULEB128 length prefix (LSB first, 7 bits per byte).
func uleb(n uint64) []byte {
	var out []byte
	for {
		b := byte(n & 0x7f)
		n >>= 7
		if n != 0 {
			out = append(out, b|0x80)
		} else {
			return append(out, b)
		}
	}
}

// A sequence length is read from an untrusted ULEB128 (up to ~4.29e9). Before the
// bound was added, make([]T, length) ran before any element was read, so a 5-byte
// input could drive a huge allocation or, for a large element type, a makeslice
// panic. The length must not exceed the bytes that remain (>=1 byte per element).
func TestDeserializeSequenceRejectsLengthBeyondRemaining(t *testing.T) {
	in := uleb(0xFFFFFFFF) // 5 bytes claiming ~4.29e9 elements, no payload
	des := NewDeserializer(in)
	out := DeserializeSequenceWithFunction(des, func(d *Deserializer, o *uint8) { *o = d.U8() })
	if des.Error() == nil {
		t.Fatalf("expected an error for a sequence length beyond the remaining bytes, got nil (len=%d)", len(out))
	}
}

// ReadBytes must reject a length that exceeds the remaining bytes before allocating.
// Before the fix the make([]byte, length) ran first, so a 4-byte input claiming
// 100 MiB allocated ~100 MiB and only then failed. Assert the allocation stays small.
func TestReadBytesDoesNotAllocateBeyondRemaining(t *testing.T) {
	in := uleb(100 * 1024 * 1024) // claims 100 MiB, no payload follows
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	des := NewDeserializer(in)
	got := des.ReadBytes()
	runtime.ReadMemStats(&m2)
	if des.Error() == nil {
		t.Fatalf("expected an error for a byte length beyond the remaining bytes, got nil (returned %d bytes)", len(got))
	}
	if grew := m2.TotalAlloc - m1.TotalAlloc; grew > 1<<20 {
		t.Fatalf("ReadBytes allocated %d bytes for a %d-byte input claiming 100 MiB; it must reject before allocating", grew, len(in))
	}
}

// The bound must not reject a well-formed length-prefixed value.
func TestReadBytesAndSequenceStillAcceptValidInput(t *testing.T) {
	in := append(uleb(3), 'a', 'b', 'c')
	des := NewDeserializer(in)
	if got := des.ReadBytes(); des.Error() != nil || string(got) != "abc" {
		t.Fatalf("valid ReadBytes regressed: got %q, err %v", got, des.Error())
	}

	// A 3-element U8 sequence: length prefix 3 followed by 3 bytes.
	seqIn := append(uleb(3), 10, 20, 30)
	sdes := NewDeserializer(seqIn)
	out := DeserializeSequenceWithFunction(sdes, func(d *Deserializer, o *uint8) { *o = d.U8() })
	if sdes.Error() != nil || len(out) != 3 || out[0] != 10 || out[2] != 30 {
		t.Fatalf("valid sequence regressed: got %v, err %v", out, sdes.Error())
	}
}
