package aptos

import "github.com/aptos-labs/aptos-go-sdk/bcs"

// CoinTransferPayload builds an EntryFunction payload for transferring coins
//
// For coinType, if none is provided, it will transfer 0x1::aptos_coin:AptosCoin
func CoinTransferPayload(coinType *TypeTag, dest AccountAddress, amount uint64) (payload *EntryFunction, err error) {
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

// CoinBatchTransferPayload builds an EntryFunction payload for transferring coins to multiple receivers
//
// For coinType, if none is provided, it will transfer 0x1::aptos_coin:AptosCoin
func CoinBatchTransferPayload(coinType *TypeTag, dests []AccountAddress, amounts []uint64) (payload *EntryFunction, err error) {
	destBytes, err := bcs.SerializeSequenceOnly(dests)
	if err != nil {
		return nil, err
	}
	amountsBytes, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		bcs.SerializeSequenceWithFunction(amounts, ser, func(ser *bcs.Serializer, amount uint64) {
			ser.U64(amount)
		})
	})
	if err != nil {
		return nil, err
	}

	if coinType == nil || *coinType == AptosCoinTypeTag {
		return &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "batch_transfer",
			ArgTypes: []TypeTag{},
			Args: [][]byte{
				destBytes,
				amountsBytes,
			},
		}, nil
	} else {
		return &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "batch_transfer_coins",
			ArgTypes: []TypeTag{*coinType},
			Args: [][]byte{
				destBytes,
				amountsBytes,
			},
		}, nil
	}
}
