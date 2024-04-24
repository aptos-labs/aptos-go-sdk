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

	pubkey, privkey, err := ed25519.GenerateKey(nil)

	assert.NoError(t, err)

	aa := txn.SignEd25519(privkey)
	//t.Log(aa)

	eaa, ok := aa.Auth.(*Ed25519Authenticator)
	assert.True(t, ok)
	epk := ed25519.PublicKey(eaa.PublicKey[:])
	assert.Equal(t, epk, pubkey)
}
