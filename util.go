package aptos

import (
	"encoding/binary"
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"strings"
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
				Address: AccountOne,
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
	signedTxn, err = rawTxn.Sign(sender)
	return
}

func SHA3_256Hash(bytes [][]byte) (output []byte) {
	hasher := sha3.New256()
	for _, b := range bytes {
		hasher.Write(b)
	}
	return hasher.Sum([]byte{})
}

// ParseHex Convenience function to deal with 0x at the beginning of hex strings
func ParseHex(hexStr string) ([]byte, error) {
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}
	return hex.DecodeString(hexStr)
}
