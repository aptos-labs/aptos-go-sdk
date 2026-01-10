package ans

import (
	"testing"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  *Name
		expectErr error
	}{
		{
			name:     "simple name",
			input:    "alice",
			expected: &Name{Domain: "alice"},
		},
		{
			name:     "with TLD",
			input:    "alice.apt",
			expected: &Name{Domain: "alice"},
		},
		{
			name:     "subdomain",
			input:    "wallet.alice",
			expected: &Name{Domain: "alice", Subdomain: "wallet"},
		},
		{
			name:     "subdomain with TLD",
			input:    "wallet.alice.apt",
			expected: &Name{Domain: "alice", Subdomain: "wallet"},
		},
		{
			name:     "uppercase normalized",
			input:    "ALICE.APT",
			expected: &Name{Domain: "alice"},
		},
		{
			name:     "with hyphens",
			input:    "my-name.apt",
			expected: &Name{Domain: "my-name"},
		},
		{
			name:     "with numbers",
			input:    "alice123.apt",
			expected: &Name{Domain: "alice123"},
		},
		{
			name:      "too short",
			input:     "ab",
			expectErr: ErrInvalidName,
		},
		{
			name:      "invalid characters",
			input:     "alice_name",
			expectErr: ErrInvalidName,
		},
		{
			name:      "too many parts",
			input:     "a.b.c.apt",
			expectErr: ErrInvalidName,
		},
		{
			name:      "empty",
			input:     "",
			expectErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseName(tt.input)
			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestName_String(t *testing.T) {
	tests := []struct {
		name     Name
		expected string
	}{
		{
			name:     Name{Domain: "alice"},
			expected: "alice.apt",
		},
		{
			name:     Name{Domain: "alice", Subdomain: "wallet"},
			expected: "wallet.alice.apt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.name.String())
		})
	}
}

func TestIsValidLabel(t *testing.T) {
	tests := []struct {
		label string
		valid bool
	}{
		{"alice", true},
		{"alice123", true},
		{"my-name", true},
		{"abc", true},
		{"ab", false},         // Too short
		{"ALICE", false},      // Uppercase
		{"alice_name", false}, // Underscore
		{"alice.name", false}, // Dot
		{"a", false},          // Too short
		{"", false},           // Empty
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidLabel(tt.label))
		})
	}
}

func TestNewClient(t *testing.T) {
	// Test that client is created with default router address
	client := NewClient(nil)
	assert.Equal(t, RouterAddress, client.routerAddress)
}

func TestNewTestnetClient(t *testing.T) {
	client := NewTestnetClient(nil)
	assert.Equal(t, TestnetRouterAddress, client.routerAddress)
}

func TestClient_WithRouterAddress(t *testing.T) {
	customAddr := aptos.MustParseAddress("0x123")
	client := NewClient(nil).WithRouterAddress(customAddr)
	assert.Equal(t, customAddr, client.routerAddress)
}

func TestClient_RegisterPayload(t *testing.T) {
	client := NewClient(nil)

	t.Run("valid registration", func(t *testing.T) {
		payload, err := client.RegisterPayload("alice.apt", RegisterOptions{Years: 2})
		require.NoError(t, err)

		assert.Equal(t, "router", payload.Module.Name)
		assert.Equal(t, "register_domain", payload.Function)
		assert.Contains(t, payload.Args, "alice")
		assert.Contains(t, payload.Args, 2)
	})

	t.Run("defaults to 1 year", func(t *testing.T) {
		payload, err := client.RegisterPayload("alice.apt", RegisterOptions{})
		require.NoError(t, err)
		assert.Contains(t, payload.Args, 1)
	})

	t.Run("rejects subdomain registration", func(t *testing.T) {
		_, err := client.RegisterPayload("wallet.alice.apt", RegisterOptions{})
		assert.Error(t, err)
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		_, err := client.RegisterPayload("ab", RegisterOptions{})
		assert.ErrorIs(t, err, ErrInvalidName)
	})
}

func TestClient_SetPrimaryNamePayload(t *testing.T) {
	client := NewClient(nil)

	t.Run("primary name", func(t *testing.T) {
		payload, err := client.SetPrimaryNamePayload("alice.apt")
		require.NoError(t, err)

		assert.Equal(t, "set_primary_name", payload.Function)
		assert.Contains(t, payload.Args, "alice")
		assert.Contains(t, payload.Args, "") // Empty subdomain
	})

	t.Run("subdomain as primary", func(t *testing.T) {
		payload, err := client.SetPrimaryNamePayload("wallet.alice.apt")
		require.NoError(t, err)

		assert.Contains(t, payload.Args, "alice")
		assert.Contains(t, payload.Args, "wallet")
	})
}

func TestClient_SetTargetAddressPayload(t *testing.T) {
	client := NewClient(nil)
	target := aptos.MustParseAddress("0x123")

	payload, err := client.SetTargetAddressPayload("alice.apt", target)
	require.NoError(t, err)

	assert.Equal(t, "set_target_addr", payload.Function)
	assert.Contains(t, payload.Args, "alice")
	assert.Contains(t, payload.Args, target.String())
}

func TestClient_RenewPayload(t *testing.T) {
	client := NewClient(nil)

	t.Run("valid renewal", func(t *testing.T) {
		payload, err := client.RenewPayload("alice.apt", 3)
		require.NoError(t, err)

		assert.Equal(t, "renew_domain", payload.Function)
		assert.Contains(t, payload.Args, "alice")
		assert.Contains(t, payload.Args, 3)
	})

	t.Run("defaults to 1 year", func(t *testing.T) {
		payload, err := client.RenewPayload("alice.apt", 0)
		require.NoError(t, err)
		assert.Contains(t, payload.Args, 1)
	})

	t.Run("rejects subdomain renewal", func(t *testing.T) {
		_, err := client.RenewPayload("wallet.alice.apt", 1)
		assert.Error(t, err)
	})
}

func TestClient_AddSubdomainPayload(t *testing.T) {
	client := NewClient(nil)
	target := aptos.MustParseAddress("0x456")

	t.Run("valid subdomain", func(t *testing.T) {
		payload, err := client.AddSubdomainPayload("alice.apt", "wallet", target)
		require.NoError(t, err)

		assert.Equal(t, "register_subdomain", payload.Function)
		assert.Contains(t, payload.Args, "alice")
		assert.Contains(t, payload.Args, "wallet")
	})

	t.Run("rejects invalid subdomain", func(t *testing.T) {
		_, err := client.AddSubdomainPayload("alice.apt", "ab", target)
		assert.ErrorIs(t, err, ErrInvalidName)
	})
}
