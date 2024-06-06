package aptos

import (
	"errors"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

/**
 * The purpose of this file is to contain entry function payloads for reuse across multiple different signers.  This is
 * because FeePayer, and Multi-sig transactions will use these payloads, in addition to SingleSigner transactions
 */

// CoinTransferEntryFunction builds an EntryFunction payload for transferring coins
//
// For coinType, if none is provided, it will transfer 0x1::aptos_coin:AptosCoin
func CoinTransferEntryFunction(coinType *TypeTag, dest AccountAddress, amount uint64) (payload *EntryFunction, err error) {
	amountBytes, err := bcs.SerializeU64(amount)
	if err != nil {
		return nil, err
	}

	if coinType == nil || *coinType == AptosCoinTypeTag {
		return &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args: [][]byte{
				dest[:],
				amountBytes,
			},
		}, nil
	} else {
		return &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer_coins",
			ArgTypes: []TypeTag{*coinType},
			Args: [][]byte{
				dest[:],
				amountBytes,
			},
		}, nil
	}
}

// FungibleAssetTransferPrimaryStoreEntryFunction builds an EntryFunction payload to transfer between two primary stores
// This is similar to CoinTransferEntryFunction
//
// For now, if metadata is nil, then it will fail to build, but in the future, APT will be the default
func FungibleAssetTransferPrimaryStoreEntryFunction(faMetadataAddress *AccountAddress, dest AccountAddress, amount uint64) (payload *EntryFunction, err error) {
	if faMetadataAddress == nil {
		return nil, errors.New("fa metadata address is nil")
	}
	amountBytes, err := bcs.SerializeU64(amount)
	if err != nil {
		return nil, err
	}

	// Build up the associated struct tag
	structTag := &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "FungibleStore"}
	typeTag := TypeTag{Value: structTag}

	return &EntryFunction{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "primary_fungible_store",
		},
		Function: "transfer",
		ArgTypes: []TypeTag{typeTag},
		Args: [][]byte{
			faMetadataAddress[:],
			dest[:],
			amountBytes,
		},
	}, nil
}

// FungibleAssetTransferEntryFunction builds an EntryFunction payload to transfer between two fungible asset stores
//
// For now, if metadata is nil, then it will fail to build, but in the future, APT will be the default
func FungibleAssetTransferEntryFunction(faMetadataAddress *AccountAddress, source AccountAddress, dest AccountAddress, amount uint64) (payload *EntryFunction, err error) {
	if faMetadataAddress == nil {
		return nil, errors.New("fa metadata address is nil")
	}
	amountBytes, err := bcs.SerializeU64(amount)
	if err != nil {
		return nil, err
	}

	// Build up the associated struct tag
	structTag := &StructTag{Address: AccountOne, Module: "fungible_asset", Name: "FungibleStore"}
	typeTag := TypeTag{Value: structTag}

	return &EntryFunction{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "fungible_asset",
		},
		Function: "transfer",
		ArgTypes: []TypeTag{typeTag},
		Args: [][]byte{
			faMetadataAddress[:],
			source[:],
			dest[:],
			amountBytes,
		},
	}, nil
}
