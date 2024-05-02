package aptos

import (
	"encoding/binary"
	"time"
)

// Move some APT from sender to dest
// Amount in Octas (10^-8 APT)
func APTTransferTransaction(client *Client, sender *Account, dest AccountAddress, amount uint64) (stxn *SignedTransaction, err error) {
	// TODO: options for MaxGasAmount, GasUnitPrice, validSeconds, sequenceNumber
	validSeconds := int64(600_000)
	var chainId uint8
	chainId, err = client.GetChainId()
	if err != nil {
		return
	}

	info, err := client.Account(sender.Address)
	if err != nil {
		return nil, err
	}
	sn, err := info.SequenceNumber()
	if err != nil {
		return nil, err
	}

	var amountbytes [8]byte
	binary.LittleEndian.PutUint64(amountbytes[:], amount)

	now := time.Now().Unix()
	txn := RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: sn,
		Payload: TransactionPayload{Payload: &EntryFunction{
			Module: ModuleId{
				Address: Account0x1,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args: [][]byte{
				dest[:],
				amountbytes[:],
			},
		}},
		MaxGasAmount:              100_000,
		GasUnitPrice:              100,
		ExpirationTimetampSeconds: uint64(now + validSeconds),
		ChainId:                   chainId,
	}

	stxn, err = txn.Sign(sender.PrivateKey)
	return stxn, err
}
