package sigma

import (
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigbcs"
	"github.com/gtank/ristretto255"
)

const (
	transferProtocolID = "AptosConfidentialAsset/TransferV1"
	transferTypeName   = "0x1::sigma_protocol_transfer::Transfer"
)

// BCSTransferSession matches Move TransferSession (TS bcsSerializeTransferSession).
func BCSTransferSession(sender32, recipient32, token32 [32]byte, numAvail, numTransfer uint64, hasEffAud bool, numVolun uint64) []byte {
	b := make([]byte, 0, 128)
	b = append(b, sender32[:]...)
	b = append(b, recipient32[:]...)
	b = append(b, token32[:]...)
	b = sigbcs.AppendU64LE(b, numAvail)
	b = sigbcs.AppendU64LE(b, numTransfer)
	b = sigbcs.AppendBool(b, hasEffAud)
	b = sigbcs.AppendU64LE(b, numVolun)
	return b
}

const (
	trIdxG      = 0
	trIdxH      = 1
	trIdxEkSid  = 2
	trIdxEkRid  = 3
	trStartOldP = 4
)

func trStartOldR(ell int) int      { return trStartOldP + ell }
func trIdxEkAudEff(ell, n int) int { return trStartOldP + 4*ell + 3*n }
func trStartVolun(ell, n int, hasEff bool, numVolun int) int {
	base := trStartOldP + 4*ell + 3*n
	if hasEff {
		base += 1 + ell + n
	}
	return base
}

// TransferProofArgs matches TS TransferProofArgs.
type TransferProofArgs struct {
	DK32                       [32]byte
	Sender32, Recipient32      [32]byte
	Token32                    [32]byte
	ChainID                    byte
	SenderEK32, RecipientEK32  []byte
	OldC, OldD                 [][]byte
	NewC, NewD                 [][]byte
	NewAmountChunks            []uint64
	NewRandomness              []*big.Int
	TransferC                  [][]byte
	TransferDSid, TransferDRid [][]byte
	TransferAmountChunks       []uint64
	TransferRandomness         []*big.Int
	HasEffectiveAuditor        bool
	AuditorEK32s               [][]byte   // voluntary first, then effective if HasEffectiveAuditor
	NewBalanceDAud             [][][]byte // per auditor index, ell rows of D
	TransferAmountDAud         [][][]byte // per auditor, n rows
}

// ProveTransfer is TS proveTransfer.
func ProveTransfer(a TransferProofArgs) (*Proof, error) {
	ell := len(a.OldC)
	n := len(a.TransferC)
	if ell == 0 {
		return nil, fmt.Errorf("transfer: no balance ciphertext chunks")
	}
	if n == 0 {
		return nil, fmt.Errorf("transfer: no transfer ciphertext chunks")
	}
	if ell != len(a.OldD) || ell != len(a.NewC) || ell != len(a.NewD) {
		return nil, fmt.Errorf("transfer: ell mismatch")
	}
	if n != len(a.TransferDSid) || n != len(a.TransferDRid) || n != len(a.TransferAmountChunks) || n != len(a.TransferRandomness) {
		return nil, fmt.Errorf("transfer: n mismatch")
	}
	numAud := len(a.AuditorEK32s)
	numVolun := numAud
	if a.HasEffectiveAuditor {
		if numAud < 1 {
			return nil, fmt.Errorf("transfer: effective auditor but no keys")
		}
		numVolun = numAud - 1
	}
	if a.HasEffectiveAuditor {
		if len(a.NewBalanceDAud) != numAud || len(a.TransferAmountDAud) != numAud {
			return nil, fmt.Errorf("transfer: auditor ciphertext rows")
		}
		for _, rows := range a.NewBalanceDAud {
			if len(rows) != ell {
				return nil, fmt.Errorf("transfer: newBalanceDAud ell")
			}
		}
		for _, rows := range a.TransferAmountDAud {
			if len(rows) != n {
				return nil, fmt.Errorf("transfer: transferAmountDAud n")
			}
		}
	} else if numAud > 0 {
		if len(a.TransferAmountDAud) != numAud {
			return nil, fmt.Errorf("transfer: voluntary auditor D")
		}
		for _, rows := range a.TransferAmountDAud {
			if len(rows) != n {
				return nil, fmt.Errorf("transfer: voluntary auditor D n")
			}
		}
	}
	dec := func(p []byte) (*ristretto255.Element, error) {
		var e ristretto255.Element
		if err := e.Decode(p); err != nil {
			return nil, err
		}
		return &e, nil
	}
	dkBig := caed25519.ModN(caed25519.BytesToNumberLE(a.DK32[:]))
	G := ca.BaseG
	H := ca.HRistretto
	var ekSid, ekRid ristretto255.Element
	if err := ekSid.Decode(a.SenderEK32); err != nil {
		return nil, err
	}
	if err := ekRid.Decode(a.RecipientEK32); err != nil {
		return nil, err
	}
	points := []*ristretto255.Element{G, H, &ekSid, &ekRid}
	compressed := [][]byte{
		G.Encode(make([]byte, 0, 32)),
		H.Encode(make([]byte, 0, 32)),
		append([]byte(nil), a.SenderEK32...),
		append([]byte(nil), a.RecipientEK32...),
	}
	push := func(p []byte) error {
		ep, err := dec(p)
		if err != nil {
			return err
		}
		points = append(points, ep)
		compressed = append(compressed, append([]byte(nil), p...))
		return nil
	}
	for i := 0; i < ell; i++ {
		if err := push(a.OldC[i]); err != nil {
			return nil, fmt.Errorf("oldC: %w", err)
		}
	}
	for i := 0; i < ell; i++ {
		if err := push(a.OldD[i]); err != nil {
			return nil, fmt.Errorf("oldD: %w", err)
		}
	}
	for i := 0; i < ell; i++ {
		if err := push(a.NewC[i]); err != nil {
			return nil, fmt.Errorf("newC: %w", err)
		}
	}
	for i := 0; i < ell; i++ {
		if err := push(a.NewD[i]); err != nil {
			return nil, fmt.Errorf("newD: %w", err)
		}
	}
	for j := 0; j < n; j++ {
		if err := push(a.TransferC[j]); err != nil {
			return nil, err
		}
	}
	for j := 0; j < n; j++ {
		if err := push(a.TransferDSid[j]); err != nil {
			return nil, err
		}
	}
	for j := 0; j < n; j++ {
		if err := push(a.TransferDRid[j]); err != nil {
			return nil, err
		}
	}
	if a.HasEffectiveAuditor {
		eff := numAud - 1
		if err := push(a.AuditorEK32s[eff]); err != nil {
			return nil, err
		}
		for i := 0; i < ell; i++ {
			if err := push(a.NewBalanceDAud[eff][i]); err != nil {
				return nil, err
			}
		}
		for j := 0; j < n; j++ {
			if err := push(a.TransferAmountDAud[eff][j]); err != nil {
				return nil, err
			}
		}
	}
	for t := 0; t < numVolun; t++ {
		if err := push(a.AuditorEK32s[t]); err != nil {
			return nil, err
		}
		for j := 0; j < n; j++ {
			if err := push(a.TransferAmountDAud[t][j]); err != nil {
				return nil, err
			}
		}
	}
	stmt := &Statement{Points: points, CompressedPoints: compressed, Scalars: nil}
	witness := make([]*big.Int, 0, 1+2*ell+2*n)
	witness = append(witness, dkBig)
	for _, c := range a.NewAmountChunks {
		witness = append(witness, new(big.Int).SetUint64(c))
	}
	witness = append(witness, a.NewRandomness...)
	for _, c := range a.TransferAmountChunks {
		witness = append(witness, new(big.Int).SetUint64(c))
	}
	witness = append(witness, a.TransferRandomness...)
	dst := DomainSeparator{
		ContractAddress: AptosFrameworkAddress,
		ChainID:         a.ChainID,
		ProtocolID:      []byte(transferProtocolID),
		SessionID:       BCSTransferSession(a.Sender32, a.Recipient32, a.Token32, uint64(ell), uint64(n), a.HasEffectiveAuditor, uint64(numVolun)),
	}
	psi := makeTransferPsi(ell, n, a.HasEffectiveAuditor, numVolun)
	return Prove(dst, transferTypeName, psi, stmt, witness)
}

func makeTransferPsi(ell, n int, hasEff bool, numVolun int) Psi {
	return func(s *Statement, w []*big.Int) []*ristretto255.Element {
		dk := w[0]
		newA := w[1 : 1+ell]
		newR := w[1+ell : 1+2*ell]
		vChunks := w[1+2*ell : 1+2*ell+n]
		rTr := w[1+2*ell+n : 1+2*ell+2*n]
		H := s.Points[trIdxH]
		ekSid := s.Points[trIdxEkSid]
		ekRid := s.Points[trIdxEkRid]
		out := make([]*ristretto255.Element, 0, 32)
		t0, _ := ca.ScalarMultElement(ekSid, dk)
		out = append(out, t0)
		for i := 0; i < ell; i++ {
			var ai, ri ristretto255.Scalar
			_ = ai.Decode(caed25519.NumberToBytesLE32(newA[i]))
			_ = ri.Decode(caed25519.NumberToBytesLE32(newR[i]))
			var t1, t2 ristretto255.Element
			t1.ScalarBaseMult(&ai)
			t2.ScalarMult(&ri, H)
			sum := new(ristretto255.Element)
			sum.Add(&t1, &t2)
			out = append(out, sum)
		}
		for i := 0; i < ell; i++ {
			t, _ := ca.ScalarMultElement(ekSid, newR[i])
			out = append(out, t)
		}
		if hasEff {
			ekAud := s.Points[trIdxEkAudEff(ell, n)]
			for i := 0; i < ell; i++ {
				t, _ := ca.ScalarMultElement(ekAud, newR[i])
				out = append(out, t)
			}
		}
		bEll := computeBPowers(ell)
		bN := computeBPowers(n)
		var bal ristretto255.Element
		bal.Zero()
		sOldR := trStartOldR(ell)
		for i := 0; i < ell; i++ {
			coef := caed25519.ModN(new(big.Int).Mul(dk, bEll[i]))
			t, _ := ca.ScalarMultElement(s.Points[sOldR+i], coef)
			bal.Add(&bal, t)
		}
		for i := 0; i < ell; i++ {
			coef := caed25519.ModN(new(big.Int).Mul(newA[i], bEll[i]))
			var sc ristretto255.Scalar
			_ = sc.Decode(caed25519.NumberToBytesLE32(coef))
			var term ristretto255.Element
			term.ScalarBaseMult(&sc)
			bal.Add(&bal, &term)
		}
		for j := 0; j < n; j++ {
			coef := caed25519.ModN(new(big.Int).Mul(vChunks[j], bN[j]))
			var sc ristretto255.Scalar
			_ = sc.Decode(caed25519.NumberToBytesLE32(coef))
			var term ristretto255.Element
			term.ScalarBaseMult(&sc)
			bal.Add(&bal, &term)
		}
		bc := new(ristretto255.Element)
		_ = bc.Decode(bal.Encode(make([]byte, 0, 32)))
		out = append(out, bc)
		for j := 0; j < n; j++ {
			var vj, rj ristretto255.Scalar
			_ = vj.Decode(caed25519.NumberToBytesLE32(vChunks[j]))
			_ = rj.Decode(caed25519.NumberToBytesLE32(rTr[j]))
			var t1, t2 ristretto255.Element
			t1.ScalarBaseMult(&vj)
			t2.ScalarMult(&rj, H)
			sum := new(ristretto255.Element)
			sum.Add(&t1, &t2)
			out = append(out, sum)
		}
		for j := 0; j < n; j++ {
			t, _ := ca.ScalarMultElement(ekSid, rTr[j])
			out = append(out, t)
		}
		for j := 0; j < n; j++ {
			t, _ := ca.ScalarMultElement(ekRid, rTr[j])
			out = append(out, t)
		}
		if hasEff {
			ekAud := s.Points[trIdxEkAudEff(ell, n)]
			for j := 0; j < n; j++ {
				t, _ := ca.ScalarMultElement(ekAud, rTr[j])
				out = append(out, t)
			}
		}
		volunStart := trStartVolun(ell, n, hasEff, numVolun)
		for t := 0; t < numVolun; t++ {
			ekVolunIdx := volunStart + t*(1+n)
			ekVolun := s.Points[ekVolunIdx]
			for j := 0; j < n; j++ {
				t, _ := ca.ScalarMultElement(ekVolun, rTr[j])
				out = append(out, t)
			}
		}
		return out
	}
}
