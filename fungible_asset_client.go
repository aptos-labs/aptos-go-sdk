package aptos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

// FungibleAssetClient This is an example client around a single fungible asset
type FungibleAssetClient struct {
	aptosClient     *Client
	metadataAddress AccountAddress
}

// NewFungibleAssetClient verifies the address exists when creating the client
// TODO: Add lookup of other metadata information such as symbol, supply, etc
func NewFungibleAssetClient(client *Client, metadataAddress AccountAddress) (faClient *FungibleAssetClient, err error) {
	// Retrieve the Metadata resource to ensure the fungible asset actually exists
	_, err = client.AccountResource(metadataAddress, "0x1::fungible_asset::Metadata")
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

func (client *FungibleAssetClient) Transfer(sender Account, senderStore AccountAddress, receiverStore AccountAddress, amount uint64) (stxn *SignedTransaction, err error) {
	// Encode inputs
	var amountBytes [8]byte
	binary.LittleEndian.PutUint64(amountBytes[:], amount)

	structTag := &StructTag{Address: Account0x1, Module: "fungible_asset", Name: "store"}
	typeTag := TypeTag{structTag}

	// Build transaction
	rawTxn, err := client.aptosClient.nodeClient.BuildTransaction(sender.Address, TransactionPayload{Payload: &EntryFunction{
		Module: ModuleId{
			Address: Account0x1,
			Name:    "fungible_asset",
		},
		Function: "transfer",
		ArgTypes: []TypeTag{
			typeTag,
		},
		Args: [][]byte{
			senderStore[:],
			receiverStore[:],
			amountBytes[:],
		},
	}})
	if err != nil {
		return
	}

	// Sign transaction
	stxn, err = rawTxn.Sign(sender.PrivateKey)
	return stxn, err
}

func (client *FungibleAssetClient) TransferPrimaryStore(sender Account, receiverAddress AccountAddress, amount uint64) (stxn *SignedTransaction, err error) {
	// Encode inputs
	var amountBytes [8]byte
	binary.LittleEndian.PutUint64(amountBytes[:], amount)

	structTag := &StructTag{Address: Account0x1, Module: "fungible_asset", Name: "store"}
	typeTag := TypeTag{structTag}

	// Build transaction
	rawTxn, err := client.aptosClient.nodeClient.BuildTransaction(sender.Address, TransactionPayload{Payload: &EntryFunction{
		Module: ModuleId{
			Address: Account0x1,
			Name:    "primary_fungible_store",
		},
		Function: "transfer",
		ArgTypes: []TypeTag{
			typeTag,
		},
		Args: [][]byte{
			client.metadataAddress[:],
			receiverAddress[:],
			amountBytes[:],
		},
	}})
	if err != nil {
		return
	}

	// Sign transaction
	stxn, err = rawTxn.Sign(sender.PrivateKey)
	return stxn, err
}

// -- View functions -- //

func (client *FungibleAssetClient) PrimaryStoreAddress(owner AccountAddress) (address AccountAddress, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_address")
	if err != nil {
		return
	}

	err = address.ParseStringRelaxed(val.(string))
	return
}

func (client *FungibleAssetClient) PrimaryStoreExists(owner AccountAddress) (exists bool, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_exists")
	if err != nil {
		return
	}

	exists = val.(bool)
	return
}

func (client *FungibleAssetClient) PrimaryBalance(owner AccountAddress) (balance uint64, err error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "balance")
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return ToU64(balanceStr)
}

func (client *FungibleAssetClient) PrimaryIsFrozen(owner AccountAddress) (isFrozen bool, err error) {
	val, err := client.viewPrimaryStore([][]byte{owner[:], client.metadataAddress[:]}, "is_frozen")
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

func (client *FungibleAssetClient) Balance(storeAddress AccountAddress) (balance uint64, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "balance")
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return strconv.ParseUint(balanceStr, 10, 64)
}
func (client *FungibleAssetClient) IsFrozen(storeAddress AccountAddress) (isFrozen bool, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "is_frozen")
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

func (client *FungibleAssetClient) StoreExists(storeAddress AccountAddress) (exists bool, err error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: Account0x1,
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

func (client *FungibleAssetClient) StoreMetadata(storeAddress AccountAddress) (metadataAddress AccountAddress, err error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "store_metadata")
	if err != nil {
		return
	}
	return unwrapObject(val)
}

func (client *FungibleAssetClient) Supply() (supply *big.Int, err error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "maximum")
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
	structTag := &StructTag{Address: Account0x1, Module: "fungible_asset", Name: "Metadata"}
	typeTag := TypeTag{structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: Account0x1,
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
	structTag := &StructTag{Address: Account0x1, Module: "fungible_asset", Name: "FungibleStore"}
	typeTag := TypeTag{structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: Account0x1,
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
	structTag := &StructTag{Address: Account0x1, Module: "fungible_asset", Name: "FungibleStore"}
	typeTag := TypeTag{structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: Account0x1,
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
	structTag := &StructTag{Address: Account0x1, Module: "fungible_asset", Name: "Metadata"}
	typeTag := TypeTag{structTag}
	payload := &ViewPayload{
		Module: ModuleId{
			Address: Account0x1,
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
func unwrapObject(val any) (address AccountAddress, err error) {
	inner, ok := val.(map[string]any)
	if !ok {
		err = errors.New("bad view return from node, could not unwrap object")
		return
	}
	addressString := inner["inner"].(string)
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

	return ToU128OrU256(numStr)
}

// TODO: Move somewhere more useful
func ToU64(val string) (uint64, error) {
	return strconv.ParseUint(val, 10, 64)
}

// TODO: move somewhere more useful
func ToU128OrU256(val string) (num *big.Int, err error) {
	_, ok := num.SetString(val, 10)
	if !ok {
		return num, fmt.Errorf("num %s is not integer", val)
	}
	return
}