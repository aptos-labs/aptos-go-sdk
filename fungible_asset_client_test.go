package aptos

import (
	"math/big"
	"testing"

	"github.com/qimeila/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("integration test expects network connection to mainnet")
	}

	// Create a new Aptos client
	aptosClient, err := NewClient(LocalnetConfig)
	require.NoError(t, err)

	// Fund sender
	sender, err := NewEd25519Account()
	require.NoError(t, err)

	err = aptosClient.Fund(sender.AccountAddress(), 100000000)
	require.NoError(t, err)

	// Convert to FA APT
	module, err := aptosClient.AccountModule(AccountOne, "coin")
	require.NoError(t, err)
	convertPayload, err := EntryFunctionFromAbi(module.Abi, AccountOne, "coin", "migrate_to_fungible_store", []any{AptosCoinTypeTag}, []any{})
	require.NoError(t, err)

	// TODO: verify that it worked
	_, err = aptosClient.BuildSignAndSubmitTransaction(sender, TransactionPayload{Payload: convertPayload})
	require.NoError(t, err)

	// Owner address
	receiver, err := NewEd25519Account()
	require.NoError(t, err)

	ownerAddress := &receiver.Address
	require.NoError(t, err)

	metadataAddress := &types.AccountTen

	// Create a fungible asset client
	faClient, err := NewFungibleAssetClient(aptosClient, metadataAddress)
	require.NoError(t, err)

	// Retrieve primary store address
	primaryStoreAddress, err := faClient.PrimaryStoreAddress(ownerAddress)
	require.NoError(t, err)

	// Check store exists (it won't)
	storeExists, err := faClient.StoreExists(primaryStoreAddress)
	require.NoError(t, err)
	assert.False(t, storeExists)

	// Send to that address

	transferTxn, err := faClient.TransferPrimaryStore(sender, receiver.Address, 100, SequenceNumber(1))
	require.NoError(t, err)
	submittedTxn, err := aptosClient.SubmitTransaction(transferTxn)
	require.NoError(t, err)
	transaction, err := aptosClient.WaitForTransaction(submittedTxn.Hash)
	require.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.True(t, transaction.Success)

	// Primary store by direct access
	balance, err := faClient.Balance(primaryStoreAddress)
	require.NoError(t, err)
	println("BALANCE: ", balance)

	name, err := faClient.Name()
	require.NoError(t, err)
	println("NAME: ", name)
	symbol, err := faClient.Symbol()
	require.NoError(t, err)
	println("Symbol: ", symbol)

	supply, err := faClient.Supply()
	require.NoError(t, err)
	assert.NotNil(t, supply)

	// TODO: need a custom coin contract to check more
	maximum, err := faClient.Maximum()
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(-1), maximum)

	// This should hold
	storeNotExist, err := faClient.StoreExists(&types.AccountOne)
	require.NoError(t, err)
	assert.False(t, storeNotExist)

	storeMetadataAddress, err := faClient.StoreMetadata(primaryStoreAddress)
	require.NoError(t, err)
	assert.Equal(t, metadataAddress, storeMetadataAddress)

	decimals, err := faClient.Decimals()
	require.NoError(t, err)
	assert.Equal(t, uint8(8), decimals)

	icon, err := faClient.IconUri()
	require.NoError(t, err)
	assert.Empty(t, icon)

	project, err := faClient.ProjectUri()
	require.NoError(t, err)
	assert.Empty(t, project)

	storePrimaryStoreAddress, err := faClient.PrimaryStoreAddress(ownerAddress)
	require.NoError(t, err)
	assert.Equal(t, primaryStoreAddress, storePrimaryStoreAddress)

	primaryStoreExists, err := faClient.PrimaryStoreExists(ownerAddress)
	require.NoError(t, err)
	assert.True(t, primaryStoreExists)

	primaryFrozen, err := faClient.PrimaryIsFrozen(ownerAddress)
	require.NoError(t, err)
	assert.False(t, primaryFrozen)

	primaryUntransferrable, err := faClient.IsUntransferable(primaryStoreAddress)
	require.NoError(t, err)
	assert.False(t, primaryUntransferrable)

	// Primary store by default
	primaryBalance, err := faClient.PrimaryBalance(ownerAddress)
	require.NoError(t, err)
	println("PRIMARY BALANCE: ", primaryBalance)

	isFrozen, err := faClient.IsFrozen(primaryStoreAddress)
	require.NoError(t, err)
	assert.False(t, isFrozen)
}
