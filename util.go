package aptos

import (
	"encoding/binary"
	"fmt"
	"time"
)

// MaxGasAmount is an option to APTTransferTransaction
type MaxGasAmount uint64

// GasUnitPrice is an option to APTTransferTransaction
type GasUnitPrice uint64

// ValidSeconds is an option to APTTransferTransaction
type ValidSeconds int64

// ValidUntil is an option to APTTransferTransaction
type ValidUntil int64

// SequenceNumber is an option to APTTransferTransaction
type SequenceNumber uint64

// ChainIdOption is an option to APTTransferTransaction
type ChainIdOption uint8

// Move some APT from sender to dest
// Amount in Octas (10^-8 APT)
//
// options may be: MaxGasAmount, GasUnitPrice, ValidSeconds, ValidUntil, SequenceNumber, ChainIdOption
func APTTransferTransaction(client *Client, sender *Account, dest AccountAddress, amount uint64, options ...any) (stxn *SignedTransaction, err error) {
	max_gas_amount := uint64(100_000)
	gas_unit_price := uint64(100)
	var validArg any
	validSet := false
	sequence_number := uint64(0)
	haveSequenceNumber := false
	chainId := uint8(0)
	haveChainId := false

	for opti, option := range options {
		switch ovalue := option.(type) {
		case MaxGasAmount:
			max_gas_amount = uint64(ovalue)
		case GasUnitPrice:
			gas_unit_price = uint64(ovalue)
		case ValidSeconds:
			if validSet {
				err = fmt.Errorf("APTTransferTransaction arg [%d] but valid time already set with %#v", opti+4, validArg)
				return
			}
			validArg = option
			validSet = true
		case ValidUntil:
			if validSet {
				err = fmt.Errorf("APTTransferTransaction arg [%d] but valid time already set with %#v", opti+4, validArg)
				return
			}
			validArg = option
			validSet = true
		case SequenceNumber:
			sequence_number = uint64(ovalue)
			haveSequenceNumber = true
		case ChainIdOption:
			chainId = uint8(ovalue)
			haveChainId = true
		default:
			err = fmt.Errorf("APTTransferTransaction arg [%d] unknown option type %T", opti+4, option)
			return
		}
	}

	if !haveChainId {
		chainId, err = client.GetChainId()
		if err != nil {
			return
		}
	}

	if !haveSequenceNumber {
		info, err := client.Account(sender.Address)
		if err != nil {
			return nil, err
		}
		sequence_number, err = info.SequenceNumber()
		if err != nil {
			return nil, err
		}
	}

	var amountbytes [8]byte
	binary.LittleEndian.PutUint64(amountbytes[:], amount)

	var expiration_timestamp uint64
	if validSet {
		switch ovalue := validArg.(type) {
		case ValidSeconds:
			expiration_timestamp = uint64(time.Now().Unix() + int64(ovalue))
		case ValidUntil:
			expiration_timestamp = uint64(ovalue)
		}
	} else {
		expiration_timestamp = uint64(time.Now().Unix() + int64(600))
	}
	txn := RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: sequence_number,
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
		MaxGasAmount:              max_gas_amount,
		GasUnitPrice:              gas_unit_price,
		ExpirationTimetampSeconds: expiration_timestamp,
		ChainId:                   chainId,
	}

	stxn, err = txn.Sign(sender.PrivateKey)
	return stxn, err
}
