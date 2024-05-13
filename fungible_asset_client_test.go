package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	defaultMetadata = "0x2ebb2ccac5e027a87fa0e2e5f656a3a4238d6a48d93ec9b610d570fc0aa0df12"
	defaultStore    = "0x8a9d57692a9d4deb1680eaf107b83c152436e10f7bb521143fa403fa95ef76a"
	defaultOwner    = "0xc67545d6f3d36ed01efc9b28cbfd0c1ae326d5d262dd077a29539bcee0edce9e"
)

func TestClient(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test expects network connection to mainnet")
	}

	// Create a new Aptos client
	aptosClient, err := NewClient(MainnetConfig)
	assert.NoError(t, err)

	// Owner address
	ownerAddress := core.AccountAddress{}
	err = ownerAddress.ParseStringRelaxed(defaultOwner)
	assert.NoError(t, err)

	// TODO: This flow seems awkward and I made mistakes by running Parse on the same address multiple times
	metadataAddress := core.AccountAddress{}
	err = metadataAddress.ParseStringRelaxed(defaultMetadata)
	assert.NoError(t, err)

	primaryStoreAddress := core.AccountAddress{}
	err = primaryStoreAddress.ParseStringRelaxed(defaultStore)
	assert.NoError(t, err)

	// Create a fungible asset client
	faClient, err := NewFungibleAssetClient(aptosClient, metadataAddress)
	assert.NoError(t, err)

	// Primary store by direct access
	balance, err := faClient.Balance(primaryStoreAddress)
	assert.NoError(t, err)
	println("BALANCE: ", balance)

	name, err := faClient.Name()
	assert.NoError(t, err)
	println("NAME: ", name)
	symbol, err := faClient.Symbol()
	assert.NoError(t, err)
	println("Symbol: ", symbol)

	supply, err := faClient.Supply()
	assert.NoError(t, err)
	println("Supply: ", supply.String())

	maximum, err := faClient.Maximum()
	assert.NoError(t, err)
	println("Maximum: ", maximum.String())

	storeExists, err := faClient.StoreExists(primaryStoreAddress)
	assert.NoError(t, err)
	assert.True(t, storeExists)

	// This should hold
	storeNotExist, err := faClient.StoreExists(core.AccountOne)
	assert.NoError(t, err)
	assert.False(t, storeNotExist)

	storeMetadataAddress, err := faClient.StoreMetadata(primaryStoreAddress)
	assert.NoError(t, err)
	assert.Equal(t, metadataAddress, storeMetadataAddress)

	decimals, err := faClient.Decimals()
	assert.NoError(t, err)
	println("DECIMALS: ", decimals)

	storePrimaryStoreAddress, err := faClient.PrimaryStoreAddress(ownerAddress)
	assert.NoError(t, err)
	assert.Equal(t, primaryStoreAddress, storePrimaryStoreAddress)

	primaryStoreExists, err := faClient.PrimaryStoreExists(ownerAddress)
	assert.NoError(t, err)
	assert.True(t, primaryStoreExists)

	// Primary store by default
	primaryBalance, err := faClient.PrimaryBalance(ownerAddress)
	assert.NoError(t, err)
	println("PRIMARY BALANCE: ", primaryBalance)

	isFrozen, err := faClient.IsFrozen(primaryStoreAddress)
	assert.NoError(t, err)
	assert.False(t, isFrozen)

}
