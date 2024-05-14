package aptos_go_sdk_test

import (
	"encoding/binary"
	"github.com/aptos-labs/aptos-go-sdk/types"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"github.com/aptos-labs/aptos-go-sdk/crypto"

	"github.com/stretchr/testify/assert"
)

func TestRawTransactionSign(t *testing.T) {
	sender, err := core.NewEd25519Account()
	assert.NoError(t, err)

	var dest core.AccountAddress
	dest.Random()

	sn := uint64(1)
	amount := uint64(10_000)
	var amountbytes [8]byte
	binary.LittleEndian.PutUint64(amountbytes[:], amount)
	txn := types.RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: sn + 1,
		Payload: types.TransactionPayload{Payload: &types.EntryFunction{
			Module: types.ModuleId{
				Address: core.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []types.TypeTag{},
			Args: [][]byte{
				dest[:],
				amountbytes[:],
			},
		}},
		MaxGasAmount:              1000,
		GasUnitPrice:              2000,
		ExpirationTimetampSeconds: 1714158778,
		ChainId:                   4,
	}

	stxn, err := txn.Sign(sender)
	assert.NoError(t, err)

	_, ok := stxn.Authenticator.Auth.(*crypto.Ed25519Authenticator)
	assert.True(t, ok)

	assert.NoError(t, stxn.Verify())

	// Serialize, Deserialze, Serialize
	// out1 and out3 should be the same
	ser := bcs.Serializer{}
	txn.MarshalBCS(&ser)
	assert.NoError(t, ser.Error())
	txn2 := types.RawTransaction{}
	txn1Bytes := ser.ToBytes()
	bcs.Deserialize(&txn2, txn1Bytes)
	ser2 := bcs.Serializer{}
	txn2.MarshalBCS(&ser2)
	txn2Bytes := ser2.ToBytes()
	assert.Equal(t, txn1Bytes, txn2Bytes)
	assert.Equal(t, txn, txn2)
}

func TestTPMarshal(t *testing.T) {
	var wat types.TransactionPayload
	var ser bcs.Serializer
	wat.MarshalBCS(&ser)
	// without payload it should fail
	assert.Error(t, ser.Error())
}
