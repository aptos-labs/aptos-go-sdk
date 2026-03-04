package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFungibleAssetPrimaryStoreTransferPayload(t *testing.T) {
	t.Parallel()
	dest := AccountTwo

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		metadata := AccountOne
		ef, err := FungibleAssetPrimaryStoreTransferPayload(&metadata, dest, 1000)
		require.NoError(t, err)
		assert.Equal(t, "primary_fungible_store", ef.Module.Name)
		assert.Equal(t, "transfer", ef.Function)
		assert.Len(t, ef.ArgTypes, 1)
		assert.Len(t, ef.Args, 3)
	})

	t.Run("nil metadata returns error", func(t *testing.T) {
		t.Parallel()
		_, err := FungibleAssetPrimaryStoreTransferPayload(nil, dest, 1000)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}

func TestFungibleAssetTransferPayload(t *testing.T) {
	t.Parallel()
	source := AccountOne
	dest := AccountTwo

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		metadata := AccountThree
		ef, err := FungibleAssetTransferPayload(&metadata, source, dest, 500)
		require.NoError(t, err)
		assert.Equal(t, "fungible_asset", ef.Module.Name)
		assert.Equal(t, "transfer", ef.Function)
		assert.Len(t, ef.ArgTypes, 1)
		assert.Len(t, ef.Args, 4)
	})

	t.Run("nil metadata returns error", func(t *testing.T) {
		t.Parallel()
		_, err := FungibleAssetTransferPayload(nil, source, dest, 500)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}
