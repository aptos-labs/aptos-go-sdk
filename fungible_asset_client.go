package aptos

import (
	"encoding/binary"
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

func (client *FungibleAssetClient) PrimaryBalance(owner *AccountAddress) (balance uint64, err error) {
	primaryStoreAddress := owner.ObjectAddressFromObject(&client.metadataAddress)
	return client.Balance(primaryStoreAddress)
}

func (client *FungibleAssetClient) Balance(address AccountAddress) (balance uint64, err error) {
	fungibleStore, err := client.aptosClient.AccountResource(address, "0x1::fungible_asset::FungibleStore")
	if err != nil {
		return
	}

	val := fungibleStore["data"].(map[string]any)["balance"].(string)
	balance, err = strconv.ParseUint(val, 10, 64)
	return
}

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

func (client *FungibleAssetClient) TransferPrimaryStore(sender Account, metadataAddress AccountAddress, receiverAddress AccountAddress, amount uint64) (stxn *SignedTransaction, err error) {
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
			metadataAddress[:],
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
