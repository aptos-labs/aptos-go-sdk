package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoinTransferPayload(t *testing.T) {
	t.Parallel()
	dest := AccountOne

	t.Run("nil coinType defaults to APT", func(t *testing.T) {
		t.Parallel()
		ef, err := CoinTransferPayload(nil, dest, 1000)
		require.NoError(t, err)
		assert.Equal(t, "aptos_account", ef.Module.Name)
		assert.Equal(t, "transfer", ef.Function)
		assert.Empty(t, ef.ArgTypes)
		assert.Len(t, ef.Args, 2)
	})

	t.Run("explicit APT coinType", func(t *testing.T) {
		t.Parallel()
		aptTag := AptosCoinTypeTag
		ef, err := CoinTransferPayload(&aptTag, dest, 500)
		require.NoError(t, err)
		assert.Equal(t, "transfer", ef.Function)
		assert.Empty(t, ef.ArgTypes)
	})

	t.Run("custom coin type", func(t *testing.T) {
		t.Parallel()
		customTag := TypeTag{Value: &StructTag{
			Address: AccountOne,
			Module:  "custom_coin",
			Name:    "CustomCoin",
		}}
		ef, err := CoinTransferPayload(&customTag, dest, 2000)
		require.NoError(t, err)
		assert.Equal(t, "transfer_coins", ef.Function)
		assert.Len(t, ef.ArgTypes, 1)
	})
}

func TestCoinBatchTransferPayload(t *testing.T) {
	t.Parallel()
	dests := []AccountAddress{AccountOne, AccountTwo}
	amounts := []uint64{100, 200}

	t.Run("nil coinType defaults to APT", func(t *testing.T) {
		t.Parallel()
		ef, err := CoinBatchTransferPayload(nil, dests, amounts)
		require.NoError(t, err)
		assert.Equal(t, "batch_transfer", ef.Function)
		assert.Empty(t, ef.ArgTypes)
		assert.Len(t, ef.Args, 2)
	})

	t.Run("custom coin type", func(t *testing.T) {
		t.Parallel()
		customTag := TypeTag{Value: &StructTag{
			Address: AccountOne,
			Module:  "custom_coin",
			Name:    "CustomCoin",
		}}
		ef, err := CoinBatchTransferPayload(&customTag, dests, amounts)
		require.NoError(t, err)
		assert.Equal(t, "batch_transfer_coins", ef.Function)
		assert.Len(t, ef.ArgTypes, 1)
	})
}
