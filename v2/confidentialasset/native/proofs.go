//go:build cgo

package native

import (
	"context"
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/movearg"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/rangeproof"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigma"
)

// NormalizeBalance submits normalize_raw.
func (c *Client) NormalizeBalance(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, twistedHex, faMetadataHex string) (*aptos.Transaction, error) {
	acct, ok := signer.(*account.Account)
	if !ok {
		return nil, fmt.Errorf("normalize: signer must be *account.Account")
	}
	pub, chunks, oldC, oldD, err := c.decryptAvailableAmountChunks(ctx, acct, token, twistedHex)
	if err != nil {
		return nil, err
	}
	oldEnc, err := ca.FromCipherChunks(pub, chunks, oldC, oldD)
	if err != nil {
		return nil, err
	}
	newRand, err := caed25519.GenListOfRandom(ca.AvailableBalanceChunkCount)
	if err != nil {
		return nil, err
	}
	newEnc, err := ca.NewEncryptedAmountFromAmount(oldEnc.Amount, pub, newRand)
	if err != nil {
		return nil, err
	}
	var audPub []byte
	var audNewD [][]byte
	audHex, err := c.GetEffectiveAuditorEncryptionKeyHex(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("normalize: auditor key: %w", err)
	}
	if audHex != "" {
		raw, err := decodeHex32(audHex)
		if err != nil {
			return nil, fmt.Errorf("normalize: auditor key decode: %w", err)
		}
		if len(raw) != 32 {
			return nil, fmt.Errorf("normalize: auditor key: expected 32 bytes, got %d", len(raw))
		}
		var ap [32]byte
		copy(ap[:], raw)
		audEnc, err := ca.NewEncryptedAmountFromAmount(oldEnc.Amount, ap, newRand)
		if err != nil {
			return nil, fmt.Errorf("normalize: auditor encrypt: %w", err)
		}
		_, audNewD = audEnc.RowsCD()
		audPub = raw
	}
	ch, err := c.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	dk32, err := confidentialasset.TwistedDecryptionKey32(acct, twistedHex)
	if err != nil {
		return nil, err
	}
	newC, newD := newEnc.RowsCD()
	wargs := sigma.WithdrawProofArgs{
		DK32:            dk32,
		Sender32:        signer.Address(),
		Token32:         token,
		ChainID:         ch,
		Amount:          big.NewInt(0),
		OldC:            oldC,
		OldD:            oldD,
		NewC:            newC,
		NewD:            newD,
		NewAmountChunks: newEnc.AmountChunks,
		NewRandomness:   newEnc.Randomness,
		AuditorPub32:    audPub,
		NewDAud:         audNewD,
	}
	sigmaProof, err := sigma.ProveWithdrawal(wargs)
	if err != nil {
		return nil, err
	}
	valB, randB := ristrettoValRandBases()
	rp, _, err := rangeproof.BatchRangeProof(newEnc.AmountChunks, flattenBlindingsLE32(newEnc.Randomness), valB, randB, ca.ChunkBits)
	if err != nil {
		return nil, err
	}
	var newBalanceAArg any = movearg.VectorVectorU8(nil)
	if len(audNewD) > 0 {
		newBalanceAArg = movearg.VectorVectorU8(audNewD)
	}
	payload := &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: "normalize_raw",
		TypeArgs: nil,
		Args: []any{
			&token,
			movearg.VectorVectorU8(newC),
			movearg.VectorVectorU8(newD),
			newBalanceAArg,
			movearg.VectorU8(rp),
			movearg.VectorVectorU8(sigmaProof.Commitment),
			movearg.VectorVectorU8(sigmaProof.Response),
		},
	}
	return c.SubmitWithSimulatedGas(ctx, signer, "normalize_raw", payload, faMetadataHex)
}

// Withdraw submits withdraw_to_raw.
func (c *Client) Withdraw(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, amountOctas uint64, recipient aptos.AccountAddress, twistedHex, faMetadataHex string) (*aptos.Transaction, error) {
	if recipient == (aptos.AccountAddress{}) {
		recipient = signer.Address()
	}
	acct, ok := signer.(*account.Account)
	if !ok {
		return nil, fmt.Errorf("withdraw: signer must be *account.Account")
	}
	norm, err := c.IsBalanceNormalized(ctx, acct.Address(), token)
	if err != nil {
		return nil, fmt.Errorf("withdraw: check normalized: %w", err)
	}
	if !norm {
		return nil, fmt.Errorf("withdraw: balance not normalized; call NormalizeBalance first")
	}
	pub, chunks, oldC, oldD, err := c.decryptAvailableAmountChunks(ctx, acct, token, twistedHex)
	if err != nil {
		return nil, err
	}
	oldEnc, err := ca.FromCipherChunks(pub, chunks, oldC, oldD)
	if err != nil {
		return nil, err
	}
	if oldEnc.Amount < amountOctas {
		return nil, fmt.Errorf("withdraw: insufficient balance")
	}
	rem := oldEnc.Amount - amountOctas
	newRand, err := caed25519.GenListOfRandom(ca.AvailableBalanceChunkCount)
	if err != nil {
		return nil, err
	}
	newEnc, err := ca.NewEncryptedAmountFromAmount(rem, pub, newRand)
	if err != nil {
		return nil, err
	}
	var audPub []byte
	var audNewD [][]byte
	audHexW, err := c.GetEffectiveAuditorEncryptionKeyHex(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("withdraw: auditor key: %w", err)
	}
	if audHexW != "" {
		raw, err := decodeHex32(audHexW)
		if err != nil {
			return nil, fmt.Errorf("withdraw: auditor key decode: %w", err)
		}
		if len(raw) != 32 {
			return nil, fmt.Errorf("withdraw: auditor key: expected 32 bytes, got %d", len(raw))
		}
		var ap [32]byte
		copy(ap[:], raw)
		audEnc, err := ca.NewEncryptedAmountFromAmount(rem, ap, newRand)
		if err != nil {
			return nil, fmt.Errorf("withdraw: auditor encrypt: %w", err)
		}
		_, audNewD = audEnc.RowsCD()
		audPub = raw
	}
	ch, err := c.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	dk32, err := confidentialasset.TwistedDecryptionKey32(acct, twistedHex)
	if err != nil {
		return nil, err
	}
	newC, newD := newEnc.RowsCD()
	wargs := sigma.WithdrawProofArgs{
		DK32:            dk32,
		Sender32:        signer.Address(),
		Token32:         token,
		ChainID:         ch,
		Amount:          new(big.Int).SetUint64(amountOctas),
		OldC:            oldC,
		OldD:            oldD,
		NewC:            newC,
		NewD:            newD,
		NewAmountChunks: newEnc.AmountChunks,
		NewRandomness:   newEnc.Randomness,
		AuditorPub32:    audPub,
		NewDAud:         audNewD,
	}
	sigmaProof, err := sigma.ProveWithdrawal(wargs)
	if err != nil {
		return nil, err
	}
	valB, randB := ristrettoValRandBases()
	rp, _, err := rangeproof.BatchRangeProof(newEnc.AmountChunks, flattenBlindingsLE32(newEnc.Randomness), valB, randB, ca.ChunkBits)
	if err != nil {
		return nil, err
	}
	var newBalanceAArg any = movearg.VectorVectorU8(nil)
	if len(audNewD) > 0 {
		newBalanceAArg = movearg.VectorVectorU8(audNewD)
	}
	payload := &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: "withdraw_to_raw",
		TypeArgs: nil,
		Args: []any{
			&token,
			&recipient,
			amountOctas,
			movearg.VectorVectorU8(newC),
			movearg.VectorVectorU8(newD),
			newBalanceAArg,
			movearg.VectorU8(rp),
			movearg.VectorVectorU8(sigmaProof.Commitment),
			movearg.VectorVectorU8(sigmaProof.Response),
		},
	}
	return c.SubmitWithSimulatedGas(ctx, signer, "withdraw_to_raw", payload, faMetadataHex)
}

// Transfer submits confidential_transfer_raw (memo empty; extend caller if needed).
func (c *Client) Transfer(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, amountOctas uint64, recipient aptos.AccountAddress, twistedHex, faMetadataHex string) (*aptos.Transaction, error) {
	return c.transferWithMemo(ctx, signer, token, amountOctas, recipient, twistedHex, faMetadataHex, "")
}

func (c *Client) transferWithMemo(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, amountOctas uint64, recipient aptos.AccountAddress, twistedHex, faMetadataHex, memo string) (*aptos.Transaction, error) {
	if recipient == (aptos.AccountAddress{}) {
		return nil, fmt.Errorf("transfer: recipient cannot be zero address")
	}
	acct, ok := signer.(*account.Account)
	if !ok {
		return nil, fmt.Errorf("transfer: signer must be *account.Account")
	}
	norm, err := c.IsBalanceNormalized(ctx, acct.Address(), token)
	if err != nil {
		return nil, fmt.Errorf("transfer: check normalized: %w", err)
	}
	if !norm {
		return nil, fmt.Errorf("transfer: balance not normalized; call NormalizeBalance first")
	}
	pub, chunks, oldC, oldD, err := c.decryptAvailableAmountChunks(ctx, acct, token, twistedHex)
	if err != nil {
		return nil, err
	}
	oldEnc, err := ca.FromCipherChunks(pub, chunks, oldC, oldD)
	if err != nil {
		return nil, err
	}
	if oldEnc.Amount < amountOctas {
		return nil, fmt.Errorf("transfer: insufficient balance")
	}
	rem := oldEnc.Amount - amountOctas
	recvHex, err := c.GetEncryptionKeyHex(ctx, recipient, token)
	if err != nil {
		return nil, fmt.Errorf("transfer: recipient key view: %w", err)
	}
	if recvHex == "" {
		return nil, fmt.Errorf("transfer: recipient has no encryption key")
	}
	recvEK, err := decodeHex32(recvHex)
	if err != nil {
		return nil, err
	}
	var recvPub [32]byte
	copy(recvPub[:], recvEK)

	newBalRand, err := caed25519.GenListOfRandom(ca.AvailableBalanceChunkCount)
	if err != nil {
		return nil, err
	}
	xferRand, err := caed25519.GenListOfRandom(ca.TransferAmountChunkCount)
	if err != nil {
		return nil, err
	}
	newEnc, err := ca.NewEncryptedAmountFromAmount(rem, pub, newBalRand)
	if err != nil {
		return nil, err
	}
	amountSender, err := ca.NewEncryptedTransferAmount(amountOctas, pub, xferRand)
	if err != nil {
		return nil, err
	}
	amountRecv, err := ca.NewEncryptedTransferAmount(amountOctas, recvPub, xferRand)
	if err != nil {
		return nil, err
	}
	var audKeys [][]byte
	var newBalDAud [][][]byte
	var xferDAud [][][]byte
	hasEff := false
	audHexT, err := c.GetEffectiveAuditorEncryptionKeyHex(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("transfer: auditor key: %w", err)
	}
	if audHexT != "" {
		raw, err := decodeHex32(audHexT)
		if err != nil {
			return nil, fmt.Errorf("transfer: auditor key decode: %w", err)
		}
		if len(raw) != 32 {
			return nil, fmt.Errorf("transfer: auditor key: expected 32 bytes, got %d", len(raw))
		}
		var ap [32]byte
		copy(ap[:], raw)
		audKeys = append(audKeys, raw)
		ne, err := ca.NewEncryptedAmountFromAmount(rem, ap, newBalRand)
		if err != nil {
			return nil, fmt.Errorf("auditor new-balance encryption: %w", err)
		}
		_, nd := ne.RowsCD()
		newBalDAud = append(newBalDAud, nd)
		te, err := ca.NewEncryptedTransferAmount(amountOctas, ap, xferRand)
		if err != nil {
			return nil, fmt.Errorf("auditor transfer encryption: %w", err)
		}
		_, td := te.RowsCD()
		xferDAud = append(xferDAud, td)
		hasEff = true
	}
	ch, err := c.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	dk32, err := confidentialasset.TwistedDecryptionKey32(acct, twistedHex)
	if err != nil {
		return nil, err
	}
	senderEK, err := ca.TwistedPublicKeyFromPrivateLE32(dk32)
	if err != nil {
		return nil, fmt.Errorf("sender twisted public key: %w", err)
	}
	newC, newD := newEnc.RowsCD()
	tsC, tsDsid := amountSender.RowsCD()
	_, tsDrid := amountRecv.RowsCD()
	targs := sigma.TransferProofArgs{
		DK32:                 dk32,
		Sender32:             signer.Address(),
		Recipient32:          recipient,
		Token32:              token,
		ChainID:              ch,
		SenderEK32:           senderEK,
		RecipientEK32:        recvEK,
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
		TransferRandomness:   xferRand[:ca.TransferAmountChunkCount],
		HasEffectiveAuditor:  hasEff,
		AuditorEK32s:         audKeys,
		NewBalanceDAud:       newBalDAud,
		TransferAmountDAud:   xferDAud,
	}
	sigmaProof, err := sigma.ProveTransfer(targs)
	if err != nil {
		return nil, err
	}
	valB, randB := ristrettoValRandBases()
	rpAmt, _, err := rangeproof.BatchRangeProof(amountSender.AmountChunks, flattenBlindingsLE32(xferRand[:ca.TransferAmountChunkCount]), valB, randB, ca.ChunkBits)
	if err != nil {
		return nil, err
	}
	rpNew, _, err := rangeproof.BatchRangeProof(newEnc.AmountChunks, flattenBlindingsLE32(newEnc.Randomness), valB, randB, ca.ChunkBits)
	if err != nil {
		return nil, err
	}
	var newBalanceAArg any = movearg.VectorVectorU8(nil)
	if hasEff && len(newBalDAud) > 0 {
		newBalanceAArg = movearg.VectorVectorU8(newBalDAud[len(newBalDAud)-1])
	}
	volKeySlice := [][]byte{}
	volTrip := [][][]byte{}
	if len(xferDAud) > 1 {
		for i := 0; i < len(xferDAud)-1; i++ {
			if i < len(audKeys)-1 {
				volKeySlice = append(volKeySlice, audKeys[i])
			}
			volTrip = append(volTrip, xferDAud[i])
		}
	}
	effD := [][]byte(nil)
	if hasEff && len(xferDAud) > 0 {
		effD = xferDAud[len(xferDAud)-1]
	}
	payload := &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: "confidential_transfer_raw",
		TypeArgs: nil,
		Args: []any{
			&token,
			&recipient,
			movearg.VectorVectorU8(newC),
			movearg.VectorVectorU8(newD),
			newBalanceAArg,
			movearg.VectorVectorU8(tsC),
			movearg.VectorVectorU8(tsDsid),
			movearg.VectorVectorU8(tsDrid),
			movearg.VectorVectorU8(effD),
			movearg.VectorVectorU8(volKeySlice),
			movearg.VectorTripleVecU8(volTrip),
			movearg.VectorU8(rpNew),
			movearg.VectorU8(rpAmt),
			movearg.VectorVectorU8(sigmaProof.Commitment),
			movearg.VectorVectorU8(sigmaProof.Response),
			memoArg(memo),
		},
	}
	return c.SubmitWithSimulatedGas(ctx, signer, "confidential_transfer_raw", payload, faMetadataHex)
}

func memoArg(memo string) any {
	if memo == "" {
		return movearg.VectorU8(nil)
	}
	return movearg.VectorU8([]byte(memo))
}
