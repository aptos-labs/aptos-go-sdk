package aptos

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawTransactionSign(t *testing.T) {
	txn := RawTransaction{
		Sender:                    [32]byte{0, 0, 0, 1},
		SequenceNumber:            137,
		MaxGasAmount:              1337,
		GasUnitPrice:              42,
		ExpirationTimetampSeconds: 1713982684,
		ChainId:                   1,
	}

	txn.Payload.Payload = &Script{
		Code:     []byte("fake code lol"),
		ArgTypes: nil,
		Args:     nil,
	}

	pubkey, privkey, err := ed25519.GenerateKey(nil)

	assert.NoError(t, err)

	aa, err := txn.SignEd25519(privkey)
	assert.NoError(t, err)
	//t.Log(aa)

	eaa, ok := aa.Auth.(*Ed25519Authenticator)
	assert.True(t, ok)
	epk := ed25519.PublicKey(eaa.PublicKey[:])
	assert.Equal(t, epk, pubkey)
}

func TestTPMarshal(t *testing.T) {
	var wat TransactionPayload
	var ser Serializer
	wat.MarshalBCS(&ser)
	// without payload it should fail
	assert.Error(t, ser.Error())
}
