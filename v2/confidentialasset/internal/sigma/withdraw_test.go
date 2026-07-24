package sigma

import (
	"math/big"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
)

func TestProveWithdrawal_normalize(t *testing.T) {
	t.Parallel()
	var dk, sender, token [32]byte
	dk[0] = 11
	sender[0] = 1
	token[0] = 2
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	var pub [32]byte
	copy(pub[:], ek)
	oldEnc, err := ca.NewEncryptedAmountFromAmount(0, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldC, oldD := oldEnc.RowsCD()
	newEnc, err := ca.NewEncryptedAmountFromAmount(0, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	newC, newD := newEnc.RowsCD()
	proof, err := ProveWithdrawal(WithdrawProofArgs{
		DK32:            dk,
		Sender32:        sender,
		Token32:         token,
		ChainID:         4,
		Amount:          big.NewInt(0),
		OldC:            oldC,
		OldD:            oldD,
		NewC:            newC,
		NewD:            newD,
		NewAmountChunks: newEnc.AmountChunks,
		NewRandomness:   newEnc.Randomness,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(proof.Commitment) == 0 {
		t.Fatal("empty proof")
	}
}

func TestProveWithdrawal_withAuditor(t *testing.T) {
	t.Parallel()
	var dk, sender, token [32]byte
	dk[0] = 11
	sender[0] = 1
	token[0] = 2
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	var pub [32]byte
	copy(pub[:], ek)
	oldEnc, err := ca.NewEncryptedAmountFromAmount(100, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldC, oldD := oldEnc.RowsCD()
	newEnc, err := ca.NewEncryptedAmountFromAmount(50, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	newC, newD := newEnc.RowsCD()
	audEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{99})
	if err != nil {
		t.Fatal(err)
	}
	var audPub [32]byte
	copy(audPub[:], audEK)
	audEnc, err := ca.NewEncryptedAmountFromAmount(50, audPub, newEnc.Randomness)
	if err != nil {
		t.Fatal(err)
	}
	_, audNewD := audEnc.RowsCD()
	proof, err := ProveWithdrawal(WithdrawProofArgs{
		DK32:            dk,
		Sender32:        sender,
		Token32:         token,
		ChainID:         4,
		Amount:          big.NewInt(50),
		OldC:            oldC,
		OldD:            oldD,
		NewC:            newC,
		NewD:            newD,
		NewAmountChunks: newEnc.AmountChunks,
		NewRandomness:   newEnc.Randomness,
		AuditorPub32:    audPub[:],
		NewDAud:         audNewD,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(proof.Commitment) == 0 {
		t.Fatal("empty proof")
	}
}
