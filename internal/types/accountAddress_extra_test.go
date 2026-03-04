package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountAddress_NamedObjectAddress(t *testing.T) {
	t.Parallel()
	result := AccountOne.NamedObjectAddress([]byte("test"))
	assert.Len(t, result[:], 32)
	assert.NotEqual(t, AccountOne, result)
}

func TestAccountAddress_ResourceAccount(t *testing.T) {
	t.Parallel()
	result := AccountOne.ResourceAccount([]byte("test"))
	assert.Len(t, result[:], 32)
	assert.NotEqual(t, AccountOne, result)
}

func TestAccountAddress_ParseStringWithPrefixRelaxed(t *testing.T) {
	t.Parallel()

	t.Run("valid with 0x prefix", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := addr.ParseStringWithPrefixRelaxed("0x1")
		require.NoError(t, err)
		assert.Equal(t, AccountOne, addr)
	})

	t.Run("missing 0x prefix", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := addr.ParseStringWithPrefixRelaxed("1")
		require.Error(t, err)
		assert.Equal(t, ErrAddressMissing0x, err)
	})

	t.Run("full address with prefix", func(t *testing.T) {
		t.Parallel()
		var addr AccountAddress
		err := addr.ParseStringWithPrefixRelaxed("0x0000000000000000000000000000000000000000000000000000000000000001")
		require.NoError(t, err)
		assert.Equal(t, AccountOne, addr)
	})
}
