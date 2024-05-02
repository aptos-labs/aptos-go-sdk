package aptos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAPTTransferTransaction(t *testing.T) {
	sender, err := NewAccount()
	assert.NoError(t, err)
	dest, err := NewAccount()
	assert.NoError(t, err)
	// yes, client is nil, with everything specified we don't need to ask for client state or system state
	stxn, err := APTTransferTransaction(nil, sender, dest.Address, 1337, MaxGasAmount(123123), GasUnitPrice(111), ValidSeconds(42), ChainIdOption(71), SequenceNumebr(31337))
	assert.NoError(t, err)
	assert.NotNil(t, stxn)

	// s/ValidSeconds/ValidUntil/
	stxn, err = APTTransferTransaction(nil, sender, dest.Address, 1337, MaxGasAmount(123123), GasUnitPrice(111), ValidUntil(time.Now().Unix()+42), ChainIdOption(71), SequenceNumebr(31337))
	assert.NoError(t, err)
	assert.NotNil(t, stxn)

	// use defaults for: max gas amount, gas unit price
	stxn, err = APTTransferTransaction(nil, sender, dest.Address, 1337, ValidSeconds(42), ChainIdOption(71), SequenceNumebr(31337))
	assert.NoError(t, err)
	assert.NotNil(t, stxn)

	// can't set valid time twice
	stxn, err = APTTransferTransaction(nil, sender, dest.Address, 1337, ValidSeconds(42), ValidUntil(31337), ChainIdOption(71), SequenceNumebr(31337))
	assert.ErrorContains(t, err, "valid time already set")
	assert.Nil(t, stxn)
}
