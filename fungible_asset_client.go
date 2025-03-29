package aptos

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// FungibleAssetClient This is an example client around a single fungible asset
type FungibleAssetClient struct {
	aptosClient     *Client         // Aptos client
	metadataAddress *AccountAddress // Metadata address of the fungible asset
}

// NewFungibleAssetClient verifies the [AccountAddress] of the metadata exists when creating the client
func NewFungibleAssetClient(client *Client, metadataAddress *AccountAddress) (*FungibleAssetClient, error) {
	// Retrieve the Metadata resource to ensure the fungible asset actually exists
	_, err := client.AccountResource(*metadataAddress, "0x1::fungible_asset::Metadata")
	if err != nil {
		return nil, err
	}

	return &FungibleAssetClient{
		client,
		metadataAddress,
	}, nil
}

// -- Entry functions -- //

// Transfer sends amount of the fungible asset from senderStore to receiverStore
func (client *FungibleAssetClient) Transfer(sender TransactionSigner, senderStore AccountAddress, receiverStore AccountAddress, amount uint64, options ...any) (*SignedTransaction, error) {
	payload, err := FungibleAssetTransferPayload(client.metadataAddress, senderStore, receiverStore, amount)
	if err != nil {
		return nil, err
	}

	// Build transaction
	rawTxn, err := client.aptosClient.BuildTransaction(sender.AccountAddress(), TransactionPayload{Payload: payload}, options...)
	if err != nil {
		return nil, err
	}

	// Sign transaction

	return rawTxn.SignedTransaction(sender)
}

// TransferPrimaryStore sends amount of the fungible asset from the primary store of the sender to receiverAddress
func (client *FungibleAssetClient) TransferPrimaryStore(sender TransactionSigner, receiverAddress AccountAddress, amount uint64, options ...any) (*SignedTransaction, error) {
	// Build transaction
	payload, err := FungibleAssetPrimaryStoreTransferPayload(client.metadataAddress, receiverAddress, amount)
	if err != nil {
		return nil, err
	}
	rawTxn, err := client.aptosClient.BuildTransaction(sender.AccountAddress(), TransactionPayload{Payload: payload}, options...)
	if err != nil {
		return nil, err
	}

	// Sign transaction
	return rawTxn.SignedTransaction(sender)
}

// -- View functions -- //

// PrimaryStoreAddress returns the [AccountAddress] of the primary store for the owner
//
// Note that the primary store may not exist at the address. Use [FungibleAssetClient.PrimaryStoreExists] to check.
func (client *FungibleAssetClient) PrimaryStoreAddress(owner *AccountAddress) (*AccountAddress, error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_address")
	if err != nil {
		return nil, err
	}

	str, ok := val.(string)
	if !ok {
		return nil, errors.New("primary_store_address is not a string")
	}

	address := &AccountAddress{}
	err = address.ParseStringRelaxed(str)
	if err != nil {
		return nil, err
	}
	return address, nil
}

// PrimaryStoreExists returns true if the primary store for the owner exists
func (client *FungibleAssetClient) PrimaryStoreExists(owner *AccountAddress) (bool, error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "primary_store_exists")
	if err != nil {
		return false, err
	}

	exists, ok := val.(bool)
	if !ok {
		return false, errors.New("primary_store_exists is not a bool")
	}
	return exists, nil
}

// PrimaryBalance returns the balance of the primary store for the owner
func (client *FungibleAssetClient) PrimaryBalance(owner *AccountAddress, ledgerVersion ...uint64) (uint64, error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "balance", ledgerVersion...)
	if err != nil {
		return 0, err
	}
	balanceStr, ok := val.(string)
	if !ok {
		return 0, errors.New("primary_store_address is not a string")
	}
	return StrToUint64(balanceStr)
}

// PrimaryIsFrozen returns true if the primary store for the owner is frozen
func (client *FungibleAssetClient) PrimaryIsFrozen(owner *AccountAddress, ledgerVersion ...uint64) (bool, error) {
	val, err := client.viewPrimaryStoreMetadata([][]byte{owner[:], client.metadataAddress[:]}, "is_frozen", ledgerVersion...)
	if err != nil {
		return false, err
	}
	isFrozen, ok := val.(bool)
	if !ok {
		return false, errors.New("isFrozen is not a bool")
	}
	return isFrozen, nil
}

// Balance returns the balance of the store
func (client *FungibleAssetClient) Balance(storeAddress *AccountAddress, ledgerVersion ...uint64) (uint64, error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "balance", ledgerVersion...)
	if err != nil {
		return 0, err
	}
	balanceStr, ok := val.(string)
	if !ok {
		return 0, errors.New("balance is not a string")
	}
	return strconv.ParseUint(balanceStr, 10, 64)
}

// IsFrozen returns true if the store is frozen
func (client *FungibleAssetClient) IsFrozen(storeAddress *AccountAddress, ledgerVersion ...uint64) (bool, error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "is_frozen", ledgerVersion...)
	if err != nil {
		return false, err
	}

	valBool, ok := val.(bool)
	if !ok {
		return false, errors.New("is_frozen is not a bool")
	}
	return valBool, nil
}

// IsUntransferable returns true if the store can't be transferred
func (client *FungibleAssetClient) IsUntransferable(storeAddress *AccountAddress, ledgerVersion ...uint64) (bool, error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "is_untransferable", ledgerVersion...)
	if err != nil {
		return false, err
	}
	valBool, ok := val.(bool)
	if !ok {
		return false, errors.New("is_untransferable is not a bool")
	}
	return valBool, nil
}

// StoreExists returns true if the store exists
func (client *FungibleAssetClient) StoreExists(storeAddress *AccountAddress, ledgerVersion ...uint64) (bool, error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: "store_exists",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{storeAddress[:]},
	}
	val, err := client.view(payload, ledgerVersion...)
	if err != nil {
		return false, err
	}
	if val, ok := val.(bool); ok {
		return val, nil
	}
	return false, errors.New("store_exists is not a bool")
}

// StoreMetadata returns the [AccountAddress] of the metadata for the store
func (client *FungibleAssetClient) StoreMetadata(storeAddress *AccountAddress, ledgerVersion ...uint64) (*AccountAddress, error) {
	val, err := client.viewStore([][]byte{storeAddress[:]}, "store_metadata", ledgerVersion...)
	if err != nil {
		return nil, err
	}
	return unwrapObject(val)
}

// Supply returns the total supply of the fungible asset
func (client *FungibleAssetClient) Supply(ledgerVersion ...uint64) (*big.Int, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "supply", ledgerVersion...)
	if err != nil {
		return nil, err
	}
	return unwrapAggregator(val)
}

// Maximum returns the maximum possible supply of the fungible asset
func (client *FungibleAssetClient) Maximum(ledgerVersion ...uint64) (*big.Int, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "maximum", ledgerVersion...)
	if err != nil {
		return nil, err
	}
	return unwrapAggregator(val)
}

// Name returns the name of the fungible asset
func (client *FungibleAssetClient) Name() (string, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "name")
	if err != nil {
		return "", err
	}
	str, ok := val.(string)
	if !ok {
		return "", errors.New("name is not a string")
	}
	return str, nil
}

// Symbol returns the symbol of the fungible asset
func (client *FungibleAssetClient) Symbol() (string, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "symbol")
	if err != nil {
		return "", err
	}
	str, ok := val.(string)
	if !ok {
		return "", errors.New("symbol is not a string")
	}
	return str, nil
}

// Decimals returns the number of decimal places for the fungible asset
func (client *FungibleAssetClient) Decimals() (uint8, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "decimals")
	if err != nil {
		return 0, err
	}
	fl, ok := val.(float64)
	if !ok {
		return 0, errors.New("decimals is not a float64")
	}
	u8, err := util.Float64ToU8(fl)
	if err != nil {
		return 0, err
	}

	return u8, nil
}

// IconUri returns the URI of the icon for the fungible asset
func (client *FungibleAssetClient) IconUri() (string, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "icon_uri")
	if err != nil {
		return "", err
	}
	str, ok := val.(string)
	if !ok {
		return "", errors.New("icon_uri is not a string")
	}
	return str, nil
}

// ProjectUri returns the URI of the project for the fungible asset
func (client *FungibleAssetClient) ProjectUri() (string, error) {
	val, err := client.viewMetadata([][]byte{client.metadataAddress[:]}, "project_uri")
	if err != nil {
		return "", err
	}
	str, ok := val.(string)
	if !ok {
		return "", errors.New("project_uri is not a string")
	}
	return str, nil
}

// viewMetadata calls a view function on the fungible asset metadata
func (client *FungibleAssetClient) viewMetadata(args [][]byte, functionName string, ledgerVersion ...uint64) (any, error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []TypeTag{metadataStructTag()},
		Args:     args,
	}
	return client.view(payload, ledgerVersion...)
}

// viewStore calls a view function on the fungible asset store
func (client *FungibleAssetClient) viewStore(args [][]byte, functionName string, ledgerVersion ...uint64) (any, error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: functionName,
		ArgTypes: []TypeTag{storeStructTag()},
		Args:     args,
	}
	return client.view(payload, ledgerVersion...)
}

// viewPrimaryStoreMetadata calls a view function on the primary fungible asset store metadata
func (client *FungibleAssetClient) viewPrimaryStoreMetadata(args [][]byte, functionName string, ledgerVersion ...uint64) (any, error) {
	payload := &ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: functionName,
		ArgTypes: []TypeTag{metadataStructTag()},
		Args:     args,
	}
	return client.view(payload, ledgerVersion...)
}

func metadataStructTag() TypeTag {
	return TypeTag{Value: &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "Metadata"}}
}

func storeStructTag() TypeTag {
	return TypeTag{Value: &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "FungibleStore"}}
}

func (client *FungibleAssetClient) view(payload *ViewPayload, ledgerVersion ...uint64) (any, error) {
	vals, err := client.aptosClient.View(payload, ledgerVersion...)
	if err != nil {
		return nil, err
	}

	return vals[0], nil
}

// Helper function to pull out the object address
// TODO: Move to somewhere more useful
func unwrapObject(val any) (*AccountAddress, error) {
	inner, ok := val.(map[string]any)
	if !ok {
		return nil, errors.New("bad view return from node, could not unwrap object")
	}
	addressString, ok := inner["inner"].(string)
	if !ok {
		return nil, errors.New("bad view return from node, could not unwrap object")
	}
	address := &AccountAddress{}
	err := address.ParseStringRelaxed(addressString)
	if err != nil {
		return nil, err
	}
	return address, nil
}

// Helper function to pull out the object address
// TODO: Move to somewhere more useful
func unwrapAggregator(val any) (*big.Int, error) {
	inner, ok := val.(map[string]any)
	if !ok {
		return nil, errors.New("bad view return from node, could not unwrap aggregator")
	}
	vals, ok := inner["vec"].([]any)
	if !ok {
		return nil, errors.New("bad view return from node, could not unwrap aggregator")
	}
	if len(vals) == 0 {
		// Specifically in this case, there is no maximum, so we return -1
		return big.NewInt(-1), nil
	}
	numStr, ok := vals[0].(string)
	if !ok {
		return nil, errors.New("bad view return from node, aggregator value is not a string")
	}

	return StrToBigInt(numStr)
}
