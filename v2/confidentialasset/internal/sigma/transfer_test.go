package sigma

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
)

func TestProveTransfer_minimal(t *testing.T) {
	t.Parallel()
	var dk, sender, recipient, token [32]byte
	dk[0] = 13
	sender[0] = 1
	recipient[0] = 2
	token[0] = 3
	senderEK, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	recipientEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{7})
	if err != nil {
		t.Fatal(err)
	}
	var senderPub, recvPub [32]byte
	copy(senderPub[:], senderEK)
	copy(recvPub[:], recipientEK)

	xferRand, err := caed25519.GenListOfRandom(ca.TransferAmountChunkCount)
	if err != nil {
		t.Fatal(err)
	}
	newBalRand, err := caed25519.GenListOfRandom(ca.AvailableBalanceChunkCount)
	if err != nil {
		t.Fatal(err)
	}

	oldEnc, err := ca.NewEncryptedAmountFromAmount(1000, senderPub, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldC, oldD := oldEnc.RowsCD()
	newEnc, err := ca.NewEncryptedAmountFromAmount(900, senderPub, newBalRand)
	if err != nil {
		t.Fatal(err)
	}
	newC, newD := newEnc.RowsCD()

	amountSender, err := ca.NewEncryptedTransferAmount(100, senderPub, xferRand)
	if err != nil {
		t.Fatal(err)
	}
	amountRecv, err := ca.NewEncryptedTransferAmount(100, recvPub, xferRand)
	if err != nil {
		t.Fatal(err)
	}
	tsC, tsDsid := amountSender.RowsCD()
	_, tsDrid := amountRecv.RowsCD()

	proof, err := ProveTransfer(TransferProofArgs{
		DK32:                 dk,
		Sender32:             sender,
		Recipient32:          recipient,
		Token32:              token,
		ChainID:              4,
		SenderEK32:           senderEK,
		RecipientEK32:        recipientEK,
		OldC:                 oldC,
		OldD:                 oldD,
		NewC:                 newC,
		NewD:                 newD,
		NewAmountChunks:      newEnc.AmountChunks,
		NewRandomness:        newEnc.Randomness,
		TransferC:            tsC,
		TransferDSid:         tsDsid,
		TransferDRid:         tsDrid,
		TransferAmountChunks: amountSender.AmountChunks,
		TransferRandomness:   xferRand,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(proof.Commitment) == 0 {
		t.Fatal("empty proof")
	}
}

func TestProveTransfer_withEffectiveAuditor(t *testing.T) {
	t.Parallel()
	var dk, sender, recipient, token [32]byte
	dk[0] = 13
	sender[0] = 1
	recipient[0] = 2
	token[0] = 3
	senderEK, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	recipientEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{7})
	if err != nil {
		t.Fatal(err)
	}
	var senderPub, recvPub [32]byte
	copy(senderPub[:], senderEK)
	copy(recvPub[:], recipientEK)
	audEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{99})
	if err != nil {
		t.Fatal(err)
	}

	xferRand, err := caed25519.GenListOfRandom(ca.TransferAmountChunkCount)
	if err != nil {
		t.Fatal(err)
	}
	newBalRand, err := caed25519.GenListOfRandom(ca.AvailableBalanceChunkCount)
	if err != nil {
		t.Fatal(err)
	}
	oldEnc, _ := ca.NewEncryptedAmountFromAmount(1000, senderPub, nil)
	oldC, oldD := oldEnc.RowsCD()
	newEnc, _ := ca.NewEncryptedAmountFromAmount(900, senderPub, newBalRand)
	newC, newD := newEnc.RowsCD()
	amountSender, _ := ca.NewEncryptedTransferAmount(100, senderPub, xferRand)
	amountRecv, _ := ca.NewEncryptedTransferAmount(100, recvPub, xferRand)
	tsC, tsDsid := amountSender.RowsCD()
	_, tsDrid := amountRecv.RowsCD()

	var audPub [32]byte
	copy(audPub[:], audEK)
	neAud, _ := ca.NewEncryptedAmountFromAmount(900, audPub, newBalRand)
	_, ndAud := neAud.RowsCD()
	teAud, _ := ca.NewEncryptedTransferAmount(100, audPub, xferRand)
	_, tdAud := teAud.RowsCD()

	proof, err := ProveTransfer(TransferProofArgs{
		DK32:                 dk,
		Sender32:             sender,
		Recipient32:          recipient,
		Token32:              token,
		ChainID:              4,
		SenderEK32:           senderEK,
		RecipientEK32:        recipientEK,
		OldC:                 oldC,
		OldD:                 oldD,
		NewC:                 newC,
		NewD:                 newD,
		NewAmountChunks:      newEnc.AmountChunks,
		NewRandomness:        newEnc.Randomness,
		TransferC:            tsC,
		TransferDSid:         tsDsid,
		TransferDRid:         tsDrid,
		TransferAmountChunks: amountSender.AmountChunks,
		TransferRandomness:   xferRand,
		HasEffectiveAuditor:  true,
		AuditorEK32s:         [][]byte{audEK},
		NewBalanceDAud:       [][][]byte{ndAud},
		TransferAmountDAud:   [][][]byte{tdAud},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(proof.Commitment) == 0 {
		t.Fatal("empty proof")
	}
}

func TestProveTransfer_validationErrors(t *testing.T) {
	t.Parallel()
	var dk, sender, recipient, token [32]byte
	dk[0] = 13
	senderEK, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	recipientEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{7})
	if err != nil {
		t.Fatal(err)
	}
	var senderPub [32]byte
	copy(senderPub[:], senderEK)

	enc, err := ca.NewEncryptedAmountFromAmount(1000, senderPub, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldC, oldD := enc.RowsCD()
	xferEnc, err := ca.NewEncryptedAmountFromAmount(100, senderPub, nil)
	if err != nil {
		t.Fatal(err)
	}
	tsC, tsDsid := xferEnc.RowsCD()
	_, tsDrid := xferEnc.RowsCD()

	t.Run("ell==0", func(t *testing.T) {
		_, err := ProveTransfer(TransferProofArgs{
			DK32:          dk,
			Sender32:      sender,
			Recipient32:   recipient,
			Token32:       token,
			SenderEK32:    senderEK,
			RecipientEK32: recipientEK,
			TransferC:     tsC,
			TransferDSid:  tsDsid,
			TransferDRid:  tsDrid,
		})
		if err == nil {
			t.Fatal("expected ell==0 error")
		}
	})

	t.Run("n==0", func(t *testing.T) {
		_, err := ProveTransfer(TransferProofArgs{
			DK32:          dk,
			Sender32:      sender,
			Recipient32:   recipient,
			Token32:       token,
			SenderEK32:    senderEK,
			RecipientEK32: recipientEK,
			OldC:          oldC,
			OldD:          oldD,
		})
		if err == nil {
			t.Fatal("expected n==0 error")
		}
	})

	t.Run("ell_mismatch", func(t *testing.T) {
		_, err := ProveTransfer(TransferProofArgs{
			DK32:          dk,
			Sender32:      sender,
			Recipient32:   recipient,
			Token32:       token,
			SenderEK32:    senderEK,
			RecipientEK32: recipientEK,
			OldC:          oldC,
			OldD:          oldD[:len(oldD)-1], // mismatch
			NewC:          oldC,
			NewD:          oldD,
			TransferC:     tsC,
			TransferDSid:  tsDsid,
			TransferDRid:  tsDrid,
		})
		if err == nil {
			t.Fatal("expected ell mismatch error")
		}
	})

	t.Run("n_mismatch", func(t *testing.T) {
		_, err := ProveTransfer(TransferProofArgs{
			DK32:          dk,
			Sender32:      sender,
			Recipient32:   recipient,
			Token32:       token,
			SenderEK32:    senderEK,
			RecipientEK32: recipientEK,
			OldC:          oldC,
			OldD:          oldD,
			NewC:          oldC,
			NewD:          oldD,
			TransferC:     tsC,
			TransferDSid:  tsDsid[:len(tsDsid)-1], // mismatch
			TransferDRid:  tsDrid,
		})
		if err == nil {
			t.Fatal("expected n mismatch error")
		}
	})

	t.Run("effective_auditor_no_keys", func(t *testing.T) {
		_, err := ProveTransfer(TransferProofArgs{
			DK32:                dk,
			Sender32:            sender,
			Recipient32:         recipient,
			Token32:             token,
			SenderEK32:          senderEK,
			RecipientEK32:       recipientEK,
			OldC:                oldC,
			OldD:                oldD,
			NewC:                oldC,
			NewD:                oldD,
			TransferC:           tsC,
			TransferDSid:        tsDsid,
			TransferDRid:        tsDrid,
			HasEffectiveAuditor: true,
			AuditorEK32s:        nil, // no keys
		})
		if err == nil {
			t.Fatal("expected effective auditor but no keys error")
		}
	})

	t.Run("auditor_ciphertext_rows_mismatch", func(t *testing.T) {
		audEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{99})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ProveTransfer(TransferProofArgs{
			DK32:                dk,
			Sender32:            sender,
			Recipient32:         recipient,
			Token32:             token,
			SenderEK32:          senderEK,
			RecipientEK32:       recipientEK,
			OldC:                oldC,
			OldD:                oldD,
			NewC:                oldC,
			NewD:                oldD,
			TransferC:           tsC,
			TransferDSid:        tsDsid,
			TransferDRid:        tsDrid,
			HasEffectiveAuditor: true,
			AuditorEK32s:        [][]byte{audEK},
			NewBalanceDAud:      nil, // mismatch: should be len 1
			TransferAmountDAud:  nil,
		})
		if err == nil {
			t.Fatal("expected auditor ciphertext rows mismatch error")
		}
	})
}

func TestProveKeyRotation(t *testing.T) {
	t.Parallel()
	var oldDK, newDK, sender, token [32]byte
	oldDK[0] = 3
	newDK[0] = 5
	sender[0] = 1
	token[0] = 2
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(oldDK)
	if err != nil {
		t.Fatal(err)
	}
	var pub [32]byte
	copy(pub[:], ek)
	enc, err := ca.NewEncryptedAmountFromAmount(0, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, oldD := enc.RowsCD()
	kr, err := ProveKeyRotation(oldDK, newDK, oldD, sender, token, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(kr.NewEkBytes) != 32 || len(kr.Proof.Commitment) == 0 {
		t.Fatal("bad key rotation proof")
	}
}
