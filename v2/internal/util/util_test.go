package util

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSha3256Hash(t *testing.T) {
	t.Parallel()

	t.Run("known input", func(t *testing.T) {
		t.Parallel()
		result := Sha3256Hash([][]byte{[]byte("hello")})
		assert.Len(t, result, 32)
		// SHA3-256 of "hello" is deterministic
		result2 := Sha3256Hash([][]byte{[]byte("hello")})
		assert.Equal(t, result, result2)
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		result := Sha3256Hash([][]byte{})
		assert.Len(t, result, 32)
	})

	t.Run("nil slice", func(t *testing.T) {
		t.Parallel()
		result := Sha3256Hash(nil)
		assert.Len(t, result, 32)
	})

	t.Run("multiple inputs", func(t *testing.T) {
		t.Parallel()
		combined := Sha3256Hash([][]byte{[]byte("hello"), []byte("world")})
		single := Sha3256Hash([][]byte{[]byte("helloworld")})
		assert.Equal(t, combined, single, "concatenated inputs should produce same hash")
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		t.Parallel()
		h1 := Sha3256Hash([][]byte{[]byte("a")})
		h2 := Sha3256Hash([][]byte{[]byte("b")})
		assert.NotEqual(t, h1, h2)
	})
}

func TestParseHex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"with 0x prefix", "0xaabb", []byte{0xaa, 0xbb}, false},
		{"without prefix", "aabb", []byte{0xaa, 0xbb}, false},
		{"just 0x", "0x", []byte{}, false},
		{"empty string", "", []byte{}, false},
		{"single byte", "0xff", []byte{0xff}, false},
		{"uppercase", "0xAABB", []byte{0xaa, 0xbb}, false},
		{"odd length", "0xaab", nil, true},
		{"invalid hex chars", "0xzzzz", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseHex(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBytesToHex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"empty", []byte{}, "0x"},
		{"single byte", []byte{0xff}, "0xff"},
		{"multiple bytes", []byte{0xaa, 0xbb, 0xcc}, "0xaabbcc"},
		{"zeros", []byte{0x00, 0x00}, "0x0000"},
		{"nil", nil, "0x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, BytesToHex(tt.input))
		})
	}
}

func TestParseHex_BytesToHex_RoundTrip(t *testing.T) {
	t.Parallel()

	inputs := [][]byte{
		{},
		{0x01},
		{0xaa, 0xbb, 0xcc, 0xdd},
		{0x00, 0xff, 0x00, 0xff},
	}

	for _, input := range inputs {
		hexStr := BytesToHex(input)
		result, err := ParseHex(hexStr)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	}
}

func TestStrToUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    uint64
		wantErr bool
	}{
		{"zero", "0", 0, false},
		{"positive", "12345", 12345, false},
		{"max uint64", "18446744073709551615", math.MaxUint64, false},
		{"overflow", "18446744073709551616", 0, true},
		{"negative", "-1", 0, true},
		{"invalid", "abc", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := StrToUint64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestStrToBigInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    *big.Int
		wantErr bool
	}{
		{"zero", "0", big.NewInt(0), false},
		{"positive", "123", big.NewInt(123), false},
		{"negative", "-1", big.NewInt(-1), false},
		{"large", "999999999999999999999999999", nil, false},
		{"invalid", "abc", nil, true},
		{"empty", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := StrToBigInt(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, 0, tt.want.Cmp(got))
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

func TestIntToU8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   int
		want    uint8
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"max", 255, 255, false},
		{"mid", 128, 128, false},
		{"overflow", 256, 0, true},
		{"negative", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := IntToU8(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIntToU16(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   int
		want    uint16
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"max", 65535, 65535, false},
		{"overflow", 65536, 0, true},
		{"negative", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := IntToU16(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIntToU32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   int
		want    uint32
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"max", math.MaxUint32, math.MaxUint32, false},
		{"overflow", math.MaxUint32 + 1, 0, true},
		{"negative", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := IntToU32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUint32ToU8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   uint32
		want    uint8
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"max u8", 255, 255, false},
		{"overflow", 256, 0, true},
		{"large overflow", 1000, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Uint32ToU8(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetBuffer32(t *testing.T) {
	t.Parallel()

	buf := GetBuffer32()
	require.NotNil(t, buf)
	assert.Len(t, *buf, 32)
	assert.Equal(t, 32, cap(*buf))

	// Write some data
	for i := range *buf {
		(*buf)[i] = byte(i)
	}

	// Return to pool
	PutBuffer32(buf)
}

func TestPutBuffer32(t *testing.T) {
	t.Parallel()

	t.Run("valid buffer is cleared", func(t *testing.T) {
		t.Parallel()
		buf := GetBuffer32()
		for i := range *buf {
			(*buf)[i] = 0xff
		}
		PutBuffer32(buf)

		// Verify the buffer was zeroed in place by PutBuffer32
		for _, b := range *buf {
			assert.Equal(t, byte(0), b)
		}
	})

	t.Run("nil buffer rejected", func(t *testing.T) {
		t.Parallel()
		PutBuffer32(nil) // Should not panic
	})

	t.Run("wrong length buffer rejected", func(t *testing.T) {
		t.Parallel()
		wrong := make([]byte, 16)
		PutBuffer32(&wrong) // Should not panic, silently rejected
	})
}

func TestGetBuffer64(t *testing.T) {
	t.Parallel()

	buf := GetBuffer64()
	require.NotNil(t, buf)
	assert.Len(t, *buf, 64)
	assert.Equal(t, 64, cap(*buf))

	PutBuffer64(buf)
}

func TestPutBuffer64(t *testing.T) {
	t.Parallel()

	t.Run("valid buffer is cleared", func(t *testing.T) {
		t.Parallel()
		buf := GetBuffer64()
		for i := range *buf {
			(*buf)[i] = 0xff
		}
		PutBuffer64(buf)

		// Verify the buffer was zeroed in place by PutBuffer64
		for _, b := range *buf {
			assert.Equal(t, byte(0), b)
		}
	})

	t.Run("nil buffer rejected", func(t *testing.T) {
		t.Parallel()
		PutBuffer64(nil) // Should not panic
	})

	t.Run("wrong length buffer rejected", func(t *testing.T) {
		t.Parallel()
		wrong := make([]byte, 32)
		PutBuffer64(&wrong) // Should not panic, silently rejected
	})
}
