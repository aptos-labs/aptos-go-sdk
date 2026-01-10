package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected AccountAddress
		wantErr  bool
	}{
		{
			name:     "zero address short",
			input:    "0x0",
			expected: AccountAddress{},
			wantErr:  false,
		},
		{
			name:     "one address short",
			input:    "0x1",
			expected: AccountAddress{31: 0x01},
			wantErr:  false,
		},
		{
			name:     "special address f",
			input:    "0xf",
			expected: AccountAddress{31: 0x0f},
			wantErr:  false,
		},
		{
			name:     "without 0x prefix",
			input:    "1",
			expected: AccountAddress{31: 0x01},
			wantErr:  false,
		},
		{
			name:     "full address",
			input:    "0x0000000000000000000000000000000000000000000000000000000000000001",
			expected: AccountAddress{31: 0x01},
			wantErr:  false,
		},
		{
			name:     "mixed case",
			input:    "0xAbCdEf",
			expected: AccountAddress{29: 0xab, 30: 0xcd, 31: 0xef},
			wantErr:  false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "just 0x",
			input:   "0x",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "0x00000000000000000000000000000000000000000000000000000000000000001",
			wantErr: true,
		},
		{
			name:    "invalid hex",
			input:   "0xgg",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addr, err := ParseAddress(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, addr)
			}
		})
	}
}

func TestMustParseAddress(t *testing.T) {
	t.Parallel()

	t.Run("valid address", func(t *testing.T) {
		t.Parallel()
		addr := MustParseAddress("0x1")
		assert.Equal(t, AccountAddress{31: 0x01}, addr)
	})

	t.Run("panics on invalid", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			MustParseAddress("invalid")
		})
	})
}

func TestAccountAddress_IsSpecial(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addr     AccountAddress
		expected bool
	}{
		{"zero", AccountAddress{}, true},
		{"one", AccountAddress{31: 0x01}, true},
		{"fifteen", AccountAddress{31: 0x0f}, true},
		{"sixteen", AccountAddress{31: 0x10}, false},
		{"non-zero prefix", AccountAddress{0: 0x01, 31: 0x01}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.addr.IsSpecial())
		})
	}
}

func TestAccountAddress_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addr     AccountAddress
		expected string
	}{
		{"zero", AccountAddress{}, "0x0"},
		{"one", AccountAddress{31: 0x01}, "0x1"},
		{"fifteen", AccountAddress{31: 0x0f}, "0xf"},
		{"sixteen", AccountAddress{31: 0x10}, "0x0000000000000000000000000000000000000000000000000000000000000010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.addr.String())
		})
	}
}

func TestAccountAddress_StringLong(t *testing.T) {
	t.Parallel()

	addr := AccountAddress{31: 0x01}
	expected := "0x0000000000000000000000000000000000000000000000000000000000000001"
	assert.Equal(t, expected, addr.StringLong())
}

func TestAccountAddress_StringShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addr     AccountAddress
		expected string
	}{
		{"zero", AccountAddress{}, "0x0"},
		{"one", AccountAddress{31: 0x01}, "0x01"},
		{"large", AccountAddress{28: 0xab, 29: 0xcd, 30: 0xef, 31: 0x12}, "0xabcdef12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.addr.StringShort())
		})
	}
}

func TestAccountAddress_Bytes(t *testing.T) {
	t.Parallel()

	addr := AccountAddress{31: 0x01}
	bytes := addr.Bytes()
	assert.Len(t, bytes, 32)
	assert.Equal(t, byte(0x01), bytes[31])

	// Ensure it's a copy
	bytes[31] = 0xff
	assert.Equal(t, byte(0x01), addr[31])
}

func TestAccountAddress_IsZero(t *testing.T) {
	t.Parallel()

	assert.True(t, AccountAddress{}.IsZero())
	assert.False(t, AccountAddress{31: 0x01}.IsZero())
}

func TestAccountAddress_JSON(t *testing.T) {
	t.Parallel()

	t.Run("marshal", func(t *testing.T) {
		t.Parallel()
		addr := AccountAddress{31: 0x01}
		data, err := json.Marshal(addr)
		require.NoError(t, err)
		assert.Equal(t, `"0x1"`, string(data))
	})

	t.Run("unmarshal", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := json.Unmarshal([]byte(`"0x1"`), &addr)
		require.NoError(t, err)
		assert.Equal(t, AccountAddress{31: 0x01}, addr)
	})

	t.Run("unmarshal full", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := json.Unmarshal([]byte(`"0x0000000000000000000000000000000000000000000000000000000000000001"`), &addr)
		require.NoError(t, err)
		assert.Equal(t, AccountAddress{31: 0x01}, addr)
	})

	t.Run("unmarshal invalid", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := json.Unmarshal([]byte(`"invalid"`), &addr)
		assert.Error(t, err)
	})

	t.Run("unmarshal non-string", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := json.Unmarshal([]byte(`123`), &addr)
		assert.Error(t, err)
	})
}

func TestAccountAddress_Text(t *testing.T) {
	t.Parallel()

	t.Run("marshal", func(t *testing.T) {
		t.Parallel()
		addr := AccountAddress{31: 0x01}
		data, err := addr.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "0x1", string(data))
	})

	t.Run("unmarshal", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := addr.UnmarshalText([]byte("0x1"))
		require.NoError(t, err)
		assert.Equal(t, AccountAddress{31: 0x01}, addr)
	})
}

func TestAccountAddress_RoundTrip(t *testing.T) {
	t.Parallel()

	addresses := []string{
		"0x0",
		"0x1",
		"0xf",
		"0x10",
		"0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	}

	for _, input := range addresses {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			addr, err := ParseAddress(input)
			require.NoError(t, err)

			// Parse the string representation back
			addr2, err := ParseAddress(addr.String())
			require.NoError(t, err)
			assert.Equal(t, addr, addr2)
		})
	}
}
