package aptos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient(t *testing.T) {
	// Create a new Aptos client
	aptosClient, err := NewClient(MainnetConfig)
	assert.NoError(t, err)

	// Owner address
	ownerAddress := AccountAddress{}
	err = ownerAddress.ParseStringRelaxed(defaultOwner)
	assert.NoError(t, err)

	// TODO: This flow seems awkward and I made mistakes by running Parse on the same address multiple times
	metadataAddress := AccountAddress{}
	err = metadataAddress.ParseStringRelaxed(defaultMetadata)
	assert.NoError(t, err)

	primaryStoreAddress := AccountAddress{}
	err = primaryStoreAddress.ParseStringRelaxed(defaultStore)
	assert.NoError(t, err)

	// Create a fungible asset client
	faClient, err := NewFungibleAssetClient(aptosClient, metadataAddress)
	assert.NoError(t, err)

	// Primary store by direct access
	balance, err := faClient.Balance(primaryStoreAddress)
	assert.NoError(t, err)
	println("BALANCE: ", balance)

	// Primary store by default
	primaryBalance, err := faClient.PrimaryBalance(&ownerAddress)
	assert.NoError(t, err)
	println("PRIMARY BALANCE: ", primaryBalance)
}
