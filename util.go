package aptos

import (
	"encoding/binary"
)

// Move some APT from sender to dest
// Amount in Octas (10^-8 APT)
//
// options may be: MaxGasAmount, GasUnitPrice, ExpirationSeconds, ValidUntil, SequenceNumber, ChainIdOption
func APTTransferTransaction(client *Client, sender *Account, dest AccountAddress, amount uint64, options ...any) (signedTxn *SignedTransaction, err error) {
	var amountBytes [8]byte
	binary.LittleEndian.PutUint64(amountBytes[:], amount)

	rawTxn, err := client.BuildTransaction(sender.Address,
		TransactionPayload{Payload: &EntryFunction{
			Module: ModuleId{
				Address: Account0x1,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args: [][]byte{
				dest[:],
				amountBytes[:],
			},
		}}, options...)
	if err != nil {
		return
	}
	signedTxn, err = rawTxn.Sign(sender.PrivateKey)
	return
}
