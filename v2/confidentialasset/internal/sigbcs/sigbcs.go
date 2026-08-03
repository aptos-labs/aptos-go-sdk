// Package sigbcs serializes Sigma Fiat–Shamir inputs the same way as TS Serializer / Move BCS.
package sigbcs

// AppendULEB128 appends a u32 in ULEB128 form (TS serializeU32AsUleb128).
func AppendULEB128(dst []byte, v uint32) []byte {
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
			dst = append(dst, b)
		} else {
			dst = append(dst, b)
			break
		}
	}
	return dst
}

// AppendBytes is BCS bytes: ULEB128(len) || data (TS serializeBytes).
func AppendBytes(dst []byte, b []byte) []byte {
	dst = AppendULEB128(dst, uint32(len(b)))
	return append(dst, b...)
}

func AppendU8(dst []byte, v byte) []byte {
	return append(dst, v)
}

func AppendU64LE(dst []byte, v uint64) []byte {
	for i := 0; i < 8; i++ {
		dst = append(dst, byte(v>>(8*i)))
	}
	return dst
}

func AppendBool(dst []byte, v bool) []byte {
	if v {
		return append(dst, 1)
	}
	return append(dst, 0)
}
