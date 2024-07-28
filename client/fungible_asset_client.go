package client

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/aptos-labs/aptos-go-sdk/types"
	"github.com/aptos-labs/aptos-go-sdk/util"
)

// FungibleAssetClient This is an example client around a single fungible asset
type FungibleAssetClient struct {
	aptosClient     *Client               // Aptos client
	metadataAddress *types.AccountAddress // Metadata address of the fungible asset
}

// NewFungibleAssetClient verifies the [types.AccountAddress] of the metadata exists when creating the client
//
// TODO: Add lookup of other metadata information such as symbol, supply, etc
func NewFungibleAssetClient(client *Client, metadataAddress *types.AccountAddress) (faClient *FungibleAssetClient, err error) {
	// Retrieve the Metadata resource to ensure the fungible asset actually exists
	// TODO: all functions should take *types.AccountAddress
	_, err = client.AccountResource(*metadataAddress, "0x1::fungible_asset::Metadata")
	if err != nil {
		return
	}

	faClient = &FungibleAssetClient{
		client,
		metadataAddress,
	}
	return
}

// -- Entry functions -- //

// Transfer sends amount of the fungible asset from senderStore to receiverStore
func (client *FungibleAssetClient) Transfer(sender types.TransactionSigner, senderStore types.AccountAddress, receiverStore types.AccountAddress, amount uint64) (signedTxn *types.SignedTransaction, err error) {
	payload, err := types.FungibleAssetTransferPayload(client.metadataAddress, senderStore, receiverStore, amount)
	if err != nil {
		return nil, err
	}

	// Build transaction
	rawTxn, err := client.aptosClient.BuildTransaction(sender.AccountAddress(), types.TransactionPayload{Payload: payload})
	if err != nil {
		return
	}

	// Sign transaction

	return rawTxn.SignedTransaction(sender)
}

// TransferPrimaryStore sends amount of the fungible asset from the primary store of the sender to receiverAddress
func (client *FungibleAssetClient) TransferPrimaryStore(sender types.TransactionSigner, receiverAddress types.AccountAddress, amount uint64) (signedTxn *types.SignedTransaction, err error) {
	// Build transaction
	payload, err := types.FungibleAssetPrimaryStoreTransferPayload(client.metadataAddress, receiverAddress, amount)
	if err != nil {
		return nil, err
	}
	rawTxn, err := client.aptosClient.BuildTransaction(sender.AccountAddress(), types.TransactionPayload{Payload: payload})
	if err != nil {
		return
	}

	// Sign transaction
	return rawTxn.SignedTransaction(sender)
}

// -- View functions -- //

// PrimaryStoreAddress returns the [types.AccountAddress] of the primary store for the owner
//
// Note that the primary store may not exist at the address. Use [FungibleAssetClient.PrimaryStoreExists] to check.
func (client *FungibleAssetClient) PrimaryStoreAddress(owner *types.AccountAddress) (address *types.AccountAddress, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_address")
	if err != nil {
		return
	}
	address = &types.AccountAddress{}
	err = address.ParseStringRelaxed(val.(string))
	return
}

// PrimaryStoreExists returns true if the primary store for the owner exists
func (client *FungibleAssetClient) PrimaryStoreExists(owner *types.AccountAddress) (exists bool, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_exists")
	if err != nil {
		return
	}

	exists = val.(bool)
	return
}

// PrimaryBalance returns the balance of the primary store for the owner
func (client *FungibleAssetClient) PrimaryBalance(owner *types.AccountAddress) (balance uint64, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "balance")
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return util.StrToUint64(balanceStr)
}

// PrimaryIsFrozen returns true if the primary store for the owner is frozen
func (client *FungibleAssetClient) PrimaryIsFrozen(owner *types.AccountAddress) (isFrozen bool, err error) {
	val, err := client.viewPrimaryStore([][]byte{owner[:], client.metadataAddress[:]}, "is_frozen")
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

// Balance returns the balance of the store
func (client *FungibleAssetClient) Balance(storeAddress *types.AccountAddress) (balance uint64, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "balance")
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return strconv.ParseUint(balanceStr, 10, 64)
}

// IsFrozen returns true if the store is frozen
func (client *FungibleAssetClient) IsFrozen(storeAddress *types.AccountAddress) (isFrozen bool, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "is_frozen")
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

// StoreExists returns true if the store exists
func (client *FungibleAssetClient) StoreExists(storeAddress *types.AccountAddress) (exists bool, err error) {
	payload := &ViewPayload{
		Module: types.ModuleId{
			Address: types.AccountOne,
			Name:    "fungible_asset",
		},
		Function: "store_exists",
		ArgTypes: []types.TypeTag{},
		Args:     [][]byte{storeAddress[:]},
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	exists = vals[0].(bool)
	return
}

// StoreMetadata returns the [types.AccountAddress] of the metadata for the store
func (client *FungibleAssetClient) StoreMetadata(storeAddress *types.AccountAddress) (metadataAddress *types.AccountAddress, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "store_metadata")
	if err != nil {
		return
	}
	return unwrapObject(val)
}

// Supply returns the total supply of the fungible asset
func (client *FungibleAssetClient) Supply() (supply *big.Int, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "supply")
	if err != nil {
		return
	}
	return unwrapAggregator(val)
}

// Maximum returns the maximum possible supply of the fungible asset
func (client *FungibleAssetClient) Maximum() (maximum *big.Int, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "maximum")
	if err != nil {
		return
	}
	return unwrapAggregator(val)

}

// Name returns the name of the fungible asset
func (client *FungibleAssetClient) Name() (name string, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "name")
	if err != nil {
		return
	}
	name = val.(string)
	return
}

// Symbol returns the symbol of the fungible asset
func (client *FungibleAssetClient) Symbol() (symbol string, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "symbol")
	if err != nil {
		return
	}
	symbol = val.(string)
	return
}

// Decimals returns the number of decimal places for the fungible asset
func (client *FungibleAssetClient) Decimals() (decimals uint8, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "decimals")
	if err != nil {
		return
	}
	decimals = uint8(val.(float64))
	return
}

// viewMetadata calls a view function on the fungible asset metadata
func (client *FungibleAssetClient) viewMetadata(args [][]byte, functionName string) (result any, err error) {
	structTag := &types.StructTag{Address: types.AccountOne, Module: "fungible_asset", Name: "Metadata"}
	tt := types.TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: types.ModuleId{
			Address: types.AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []types.TypeTag{tt},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

// viewStore calls a view function on the fungible asset store
func (client *FungibleAssetClient) viewStore(args [][]byte, functionName string) (result any, err error) {
	structTag := &types.StructTag{Address: types.AccountOne, Module: "fungible_asset", Name: "FungibleStore"}
	tt := types.TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: types.ModuleId{
			Address: types.AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []types.TypeTag{tt},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

// viewPrimaryStore calls a view function on the primary fungible asset store
func (client *FungibleAssetClient) viewPrimaryStore(args [][]byte, functionName string) (result any, err error) {
	structTag := &types.StructTag{Address: types.AccountOne, Module: "fungible_asset", Name: "FungibleStore"}
	tt := types.TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: types.ModuleId{
			Address: types.AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []types.TypeTag{tt},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

// viewPrimaryStoreMetadata calls a view function on the primary fungible asset store metadata
func (client *FungibleAssetClient) viewPrimaryStoreMetadata(args [][]byte, functionName string) (result any, err error) {
	structTag := &types.StructTag{Address: types.AccountOne, Module: "fungible_asset", Name: "Metadata"}
	tt := types.TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: types.ModuleId{
			Address: types.AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []types.TypeTag{tt},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

// Helper function to pull out the object address
// TODO: Move to somewhere more useful
func unwrapObject(val any) (address *types.AccountAddress, err error) {
	inner, ok := val.(map[string]any)
	if !ok {
		err = errors.New("bad view return from node, could not unwrap object")
		return
	}
	addressString := inner["inner"].(string)
	address = &types.AccountAddress{}
	err = address.ParseStringRelaxed(addressString)
	return
}

// Helper function to pull out the object address
// TODO: Move to somewhere more useful
func unwrapAggregator(val any) (num *big.Int, err error) {
	inner, ok := val.(map[string]any)
	if !ok {
		err = errors.New("bad view return from node, could not unwrap aggregator")
		return
	}
	vals := inner["vec"].([]any)
	if len(vals) == 0 {
		return nil, nil
	}
	numStr, ok := vals[0].(string)
	if !ok {
		err = errors.New("bad view return from node, aggregator value is not a string")
		return
	}

	return util.StrToBigInt(numStr)
}
