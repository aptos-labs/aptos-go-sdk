package aptos

import (
	"errors"
	"math/big"
	"strconv"
)

// FungibleAssetClient This is an example client around a single fungible asset
type FungibleAssetClient struct {
	aptosClient     *Client
	metadataAddress *AccountAddress
}

// NewFungibleAssetClient verifies the address exists when creating the client
// TODO: Add lookup of other metadata information such as symbol, supply, etc
func NewFungibleAssetClient(client *Client, metadataAddress *AccountAddress) (faClient *FungibleAssetClient, err error) {
	// Retrieve the Metadata resource to ensure the fungible asset actually exists
	// TODO: all functions should take *AccountAddress
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

func (client *FungibleAssetClient) Transfer(sender TransactionSigner, senderStore AccountAddress, receiverStore AccountAddress, amount uint64) (signedTxn *SignedTransaction, err error) {
	payload, err := FungibleAssetTransferPayload(client.metadataAddress, senderStore, receiverStore, amount)
	if err != nil {
		return nil, err
	}

	// Build transaction
	rawTxn, err := client.aptosClient.BuildTransaction(sender.AccountAddress(), TransactionPayload{Payload: payload})
	if err != nil {
		return
	}

	// Sign transaction

	return rawTxn.SignedTransaction(sender)
}

func (client *FungibleAssetClient) TransferPrimaryStore(sender TransactionSigner, receiverAddress AccountAddress, amount uint64) (signedTxn *SignedTransaction, err error) {
	// Build transaction
	payload, err := FungibleAssetPrimaryStoreTransferPayload(client.metadataAddress, receiverAddress, amount)
	if err != nil {
		return nil, err
	}
	rawTxn, err := client.aptosClient.BuildTransaction(sender.AccountAddress(), TransactionPayload{Payload: payload})
	if err != nil {
		return
	}

	// Sign transaction
	return rawTxn.SignedTransaction(sender)
}

// -- View functions -- //

func (client *FungibleAssetClient) PrimaryStoreAddress(owner *AccountAddress) (address *AccountAddress, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_address")
	if err != nil {
		return
	}
	address = &AccountAddress{}
	err = address.ParseStringRelaxed(val.(string))
	return
}

func (client *FungibleAssetClient) PrimaryStoreExists(owner *AccountAddress) (exists bool, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_exists")
	if err != nil {
		return
	}

	exists = val.(bool)
	return
}

func (client *FungibleAssetClient) PrimaryBalance(owner *AccountAddress) (balance uint64, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "balance")
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return StrToUint64(balanceStr)
}

func (client *FungibleAssetClient) PrimaryIsFrozen(owner *AccountAddress) (isFrozen bool, err error) {
	val, err := client.viewPrimaryStore([][]byte{owner[:], client.metadataAddress[:]}, "is_frozen")
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

func (client *FungibleAssetClient) Balance(storeAddress *AccountAddress) (balance uint64, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "balance")
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return strconv.ParseUint(balanceStr, 10, 64)
}
func (client *FungibleAssetClient) IsFrozen(storeAddress *AccountAddress) (isFrozen bool, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "is_frozen")
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

func (client *FungibleAssetClient) StoreExists(storeAddress *AccountAddress) (exists bool, err error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: "store_exists",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{storeAddress[:]},
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	exists = vals[0].(bool)
	return
}

func (client *FungibleAssetClient) StoreMetadata(storeAddress *AccountAddress) (metadataAddress *AccountAddress, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "store_metadata")
	if err != nil {
		return
	}
	return unwrapObject(val)
}

func (client *FungibleAssetClient) Supply() (supply *big.Int, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "supply")
	if err != nil {
		return
	}
	return unwrapAggregator(val)
}

func (client *FungibleAssetClient) Maximum() (maximum *big.Int, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "maximum")
	if err != nil {
		return
	}
	return unwrapAggregator(val)

}

func (client *FungibleAssetClient) Name() (name string, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "name")
	if err != nil {
		return
	}
	name = val.(string)
	return
}

func (client *FungibleAssetClient) Symbol() (symbol string, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "symbol")
	if err != nil {
		return
	}
	symbol = val.(string)
	return
}

func (client *FungibleAssetClient) Decimals() (decimals uint8, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "decimals")
	if err != nil {
		return
	}
	decimals = uint8(val.(float64))
	return
}

func (client *FungibleAssetClient) viewMetadata(args [][]byte, functionName string) (result any, err error) {
	structTag := &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "Metadata"}
	typeTag := TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []TypeTag{typeTag},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

func (client *FungibleAssetClient) viewStore(args [][]byte, functionName string) (result any, err error) {
	structTag := &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "FungibleStore"}
	typeTag := TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []TypeTag{typeTag},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

func (client *FungibleAssetClient) viewPrimaryStore(args [][]byte, functionName string) (result any, err error) {
	structTag := &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "FungibleStore"}
	typeTag := TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []TypeTag{typeTag},
		Args:     args,
	}

	vals, err := client.aptosClient.View(payload)
	if err != nil {
		return
	}

	return vals[0], nil
}

func (client *FungibleAssetClient) viewPrimaryStoreMetadata(args [][]byte, functionName string) (result any, err error) {
	structTag := &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "Metadata"}
	typeTag := TypeTag{Value: structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []TypeTag{typeTag},
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
func unwrapObject(val any) (address *AccountAddress, err error) {
	inner, ok := val.(map[string]any)
	if !ok {
		err = errors.New("bad view return from node, could not unwrap object")
		return
	}
	addressString := inner["inner"].(string)
	address = &AccountAddress{}
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

	return StrToBigInt(numStr)
}
