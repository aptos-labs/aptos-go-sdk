package aptos

import (
	"context"
	"errors"
	"math/big"
	"strconv"
)

// FungibleAssetClient This is an example client around a single fungible asset
type FungibleAssetClient struct {
	aptosClient     *Client         // Aptos client
	metadataAddress *AccountAddress // Metadata address of the fungible asset
}

// NewFungibleAssetClient verifies the [AccountAddress] of the metadata exists when creating the client
func NewFungibleAssetClient(client *Client, metadataAddress *AccountAddress) (faClient *FungibleAssetClient, err error) {
	// Retrieve the Metadata resource to ensure the fungible asset actually exists
	_, err = client.AccountResource(context.Background(), *metadataAddress, "0x1::fungible_asset::Metadata")
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
func (client *FungibleAssetClient) Transfer(ctx context.Context, sender TransactionSigner, senderStore AccountAddress, receiverStore AccountAddress, amount uint64) (signedTxn *SignedTransaction, err error) {
	payload, err := FungibleAssetTransferPayload(client.metadataAddress, senderStore, receiverStore, amount)
	if err != nil {
		return nil, err
	}

	// Build transaction
	rawTxn, err := client.aptosClient.BuildTransaction(ctx, sender.AccountAddress(), TransactionPayload{Payload: payload})
	if err != nil {
		return
	}

	// Sign transaction

	return rawTxn.SignedTransaction(sender)
}

// TransferPrimaryStore sends amount of the fungible asset from the primary store of the sender to receiverAddress
func (client *FungibleAssetClient) TransferPrimaryStore(ctx context.Context, sender TransactionSigner, receiverAddress AccountAddress, amount uint64) (signedTxn *SignedTransaction, err error) {
	// Build transaction
	payload, err := FungibleAssetPrimaryStoreTransferPayload(client.metadataAddress, receiverAddress, amount)
	if err != nil {
		return nil, err
	}
	rawTxn, err := client.aptosClient.BuildTransaction(ctx, sender.AccountAddress(), TransactionPayload{Payload: payload})
	if err != nil {
		return
	}

	// Sign transaction
	return rawTxn.SignedTransaction(sender)
}

// -- View functions -- //

// PrimaryStoreAddress returns the [AccountAddress] of the primary store for the owner
//
// Note that the primary store may not exist at the address. Use [FungibleAssetClient.PrimaryStoreExists] to check.
func (client *FungibleAssetClient) PrimaryStoreAddress(ctx context.Context, owner *AccountAddress) (address *AccountAddress, err error) {
	val, err := client.viewPrimaryStoreMetadata(ctx, [][]byte{owner[:], client.metadataAddress[:]}, "primary_store_address")
	if err != nil {
		return
	}
	address = &AccountAddress{}
	err = address.ParseStringRelaxed(val.(string))
	return
}

// PrimaryStoreExists returns true if the primary store for the owner exists
func (client *FungibleAssetClient) PrimaryStoreExists(ctx context.Context, owner *AccountAddress) (exists bool, err error) {
	val, err := client.viewPrimaryStoreMetadata(ctx, [][]byte{owner[:], client.metadataAddress[:]}, "primary_store_exists")
	if err != nil {
		return
	}

	exists = val.(bool)
	return
}

// PrimaryBalance returns the balance of the primary store for the owner
func (client *FungibleAssetClient) PrimaryBalance(ctx context.Context, owner *AccountAddress, ledgerVersion ...uint64) (balance uint64, err error) {
	val, err := client.viewPrimaryStoreMetadata(ctx, [][]byte{owner[:], client.metadataAddress[:]}, "balance", ledgerVersion...)
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return StrToUint64(balanceStr)
}

// PrimaryIsFrozen returns true if the primary store for the owner is frozen
func (client *FungibleAssetClient) PrimaryIsFrozen(ctx context.Context, owner *AccountAddress, ledgerVersion ...uint64) (isFrozen bool, err error) {
	val, err := client.viewPrimaryStore(ctx, [][]byte{owner[:], client.metadataAddress[:]}, "is_frozen", ledgerVersion...)
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

// Balance returns the balance of the store
func (client *FungibleAssetClient) Balance(ctx context.Context, storeAddress *AccountAddress, ledgerVersion ...uint64) (balance uint64, err error) {
	val, err := client.viewStore(ctx, [][]byte{storeAddress[:]}, "balance", ledgerVersion...)
	if err != nil {
		return
	}
	balanceStr := val.(string)
	return strconv.ParseUint(balanceStr, 10, 64)
}

// IsFrozen returns true if the store is frozen
func (client *FungibleAssetClient) IsFrozen(ctx context.Context, storeAddress *AccountAddress, ledgerVersion ...uint64) (isFrozen bool, err error) {
	val, err := client.viewStore(ctx, [][]byte{storeAddress[:]}, "is_frozen", ledgerVersion...)
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

// IsUntransferable returns true if the store can't be transferred
func (client *FungibleAssetClient) IsUntransferable(ctx context.Context, storeAddress *AccountAddress, ledgerVersion ...uint64) (isFrozen bool, err error) {
	val, err := client.viewStore(ctx, [][]byte{storeAddress[:]}, "is_untransferable", ledgerVersion...)
	if err != nil {
		return
	}
	isFrozen = val.(bool)
	return
}

// StoreExists returns true if the store exists
func (client *FungibleAssetClient) StoreExists(ctx context.Context, storeAddress *AccountAddress, ledgerVersion ...uint64) (exists bool, err error) {
	val, err := client.viewStore(ctx, [][]byte{storeAddress[:]}, "store_exists", ledgerVersion...)
	if err != nil {
		return
	}

	exists = val.(bool)
	return
}

// StoreMetadata returns the [AccountAddress] of the metadata for the store
func (client *FungibleAssetClient) StoreMetadata(ctx context.Context, storeAddress *AccountAddress, ledgerVersion ...uint64) (metadataAddress *AccountAddress, err error) {
	val, err := client.viewStore(ctx, [][]byte{storeAddress[:]}, "store_metadata", ledgerVersion...)
	if err != nil {
		return
	}
	return unwrapObject(val)
}

// Supply returns the total supply of the fungible asset
func (client *FungibleAssetClient) Supply(ctx context.Context, ledgerVersion ...uint64) (supply *big.Int, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "supply", ledgerVersion...)
	if err != nil {
		return
	}
	return unwrapAggregator(val)
}

// Maximum returns the maximum possible supply of the fungible asset
func (client *FungibleAssetClient) Maximum(ctx context.Context, ledgerVersion ...uint64) (maximum *big.Int, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "maximum", ledgerVersion...)
	if err != nil {
		return
	}
	return unwrapAggregator(val)
}

// Name returns the name of the fungible asset
func (client *FungibleAssetClient) Name(ctx context.Context) (name string, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "name")
	if err != nil {
		return
	}
	name = val.(string)
	return
}

// Symbol returns the symbol of the fungible asset
func (client *FungibleAssetClient) Symbol(ctx context.Context) (symbol string, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "symbol")
	if err != nil {
		return
	}
	symbol = val.(string)
	return
}

// Decimals returns the number of decimal places for the fungible asset
func (client *FungibleAssetClient) Decimals(ctx context.Context) (decimals uint8, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "decimals")
	if err != nil {
		return
	}
	decimals = uint8(val.(float64))
	return
}

// IconUri returns the URI of the icon for the fungible asset
func (client *FungibleAssetClient) IconUri(ctx context.Context, uri string, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "icon_uri")
	if err != nil {
		return
	}
	uri = val.(string)
	return
}

// ProjectUri returns the URI of the project for the fungible asset
func (client *FungibleAssetClient) ProjectUri(ctx context.Context, uri string, err error) {
	val, err := client.viewMetadata(ctx, [][]byte{client.metadataAddress[:]}, "project_uri")
	if err != nil {
		return
	}
	uri = val.(string)
	return
}

// viewMetadata calls a view function on the fungible asset metadata
func (client *FungibleAssetClient) viewMetadata(ctx context.Context, args [][]byte, functionName string, ledgerVersion ...uint64) (result any, err error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []TypeTag{metadataStructTag()},
		Args:     args,
	}
	return client.view(ctx, payload, ledgerVersion...)
}

// viewStore calls a view function on the fungible asset store
func (client *FungibleAssetClient) viewStore(ctx context.Context, args [][]byte, functionName string, ledgerVersion ...uint64) (result any, err error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []TypeTag{storeStructTag()},
		Args:     args,
	}
	return client.view(ctx, payload, ledgerVersion...)
}

// viewPrimaryStore calls a view function on the primary fungible asset store
func (client *FungibleAssetClient) viewPrimaryStore(ctx context.Context, args [][]byte, functionName string, ledgerVersion ...uint64) (result any, err error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []TypeTag{storeStructTag()},
		Args:     args,
	}
	return client.view(ctx, payload, ledgerVersion...)
}

// viewPrimaryStoreMetadata calls a view function on the primary fungible asset store metadata
func (client *FungibleAssetClient) viewPrimaryStoreMetadata(ctx context.Context, args [][]byte, functionName string, ledgerVersion ...uint64) (result any, err error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []TypeTag{metadataStructTag()},
		Args:     args,
	}
	return client.view(ctx, payload, ledgerVersion...)
}

func metadataStructTag() TypeTag {
	return TypeTag{Value: &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "Metadata"}}
}

func storeStructTag() TypeTag {
	return TypeTag{Value: &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "FungibleStore"}}
}

func (client *FungibleAssetClient) view(ctx context.Context, payload *ViewPayload, ledgerVersion ...uint64) (result any, err error) {
	vals, err := client.aptosClient.View(ctx, payload, ledgerVersion...)
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
