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
	withdrawProtocolID = "AptosConfidentialAsset/WithdrawalV1"
	withdrawTypeName   = "0x1::sigma_protocol_withdraw::Withdrawal"
)

// BCSWithdrawSession matches Move WithdrawSession.
func BCSWithdrawSession(sender32, token32 [32]byte, numChunks uint64, hasAuditor bool) []byte {
	b := make([]byte, 0, 80)
	b = append(b, sender32[:]...)
	b = append(b, token32[:]...)
	b = sigbcs.AppendU64LE(b, numChunks)
	b = sigbcs.AppendBool(b, hasAuditor)
	return b
}

func computeBPowers(count int) []*big.Int {
	if count == 0 {
		return nil
	}
	const B = int64(1 << 16)
	powers := make([]*big.Int, count)
	powers[0] = big.NewInt(1)
	for i := 1; i < count; i++ {
		powers[i] = caed25519.ModN(new(big.Int).Mul(powers[i-1], big.NewInt(B)))
	}
	return powers
}

const (
	idxWG     = 0
	idxWH     = 1
	idxWEk    = 2
	startOldP = 3
)

func startOldR(ell int) int    { return startOldP + ell }
func startNewP(ell int) int    { return startOldP + 2*ell }
func startNewR(ell int) int    { return startOldP + 3*ell }
func idxEkAud(ell int) int     { return startOldP + 4*ell }
func startNewRAud(ell int) int { return startOldP + 4*ell + 1 }

// WithdrawProofArgs matches TS WithdrawProofArgs (points from EncryptedAmount).
type WithdrawProofArgs struct {
	DK32            [32]byte
	Sender32        [32]byte
	Token32         [32]byte
	ChainID         byte
	Amount          *big.Int // 0 for normalize
	OldC, OldD      [][]byte
	NewC, NewD      [][]byte
	NewAmountChunks []uint64
	NewRandomness   []*big.Int
	AuditorPub32    []byte   // optional, nil if absent
	NewDAud         [][]byte // optional D under auditor, same ell as balance
}

// ProveWithdrawal is TS proveWithdrawal / proveNormalization (amount=0).
func ProveWithdrawal(a WithdrawProofArgs) (*Proof, error) {
	ell := len(a.OldC)
	if ell == 0 {
		return nil, fmt.Errorf("sigma withdraw: no ciphertext chunks")
	}
	if ell != len(a.OldD) || ell != len(a.NewC) || ell != len(a.NewD) || ell != len(a.NewAmountChunks) || ell != len(a.NewRandomness) {
		return nil, fmt.Errorf("sigma withdraw: mismatched ell")
	}
	hasAuditor := len(a.AuditorPub32) == 32 && a.NewDAud != nil
	if hasAuditor && len(a.NewDAud) != ell {
		return nil, fmt.Errorf("sigma withdraw: auditor D len")
	}
	dec := func(p []byte) (*ristretto255.Element, error) {
		var e ristretto255.Element
		if err := e.Decode(p); err != nil {
			return nil, err
		}
		return &e, nil
	}
	dkBig := caed25519.ModN(caed25519.BytesToNumberLE(a.DK32[:]))
	ekBytes, err := ca.TwistedPublicKeyFromPrivateLE32(a.DK32)
	if err != nil {
		return nil, err
	}
	var ek ristretto255.Element
	if err := ek.Decode(ekBytes); err != nil {
		return nil, err
	}
	G := ca.BaseG
	H := ca.HRistretto
	points := []*ristretto255.Element{G, H, &ek}
	compressed := [][]byte{
		G.Encode(make([]byte, 0, 32)),
		H.Encode(make([]byte, 0, 32)),
		ekBytes,
	}
	for i := 0; i < ell; i++ {
		op, err := dec(a.OldC[i])
		if err != nil {
			return nil, fmt.Errorf("oldC[%d]: %w", i, err)
		}
		points = append(points, op)
		compressed = append(compressed, append([]byte(nil), a.OldC[i]...))
	}
	for i := 0; i < ell; i++ {
		op, err := dec(a.OldD[i])
		if err != nil {
			return nil, fmt.Errorf("oldD[%d]: %w", i, err)
		}
		points = append(points, op)
		compressed = append(compressed, append([]byte(nil), a.OldD[i]...))
	}
	for i := 0; i < ell; i++ {
		op, err := dec(a.NewC[i])
		if err != nil {
			return nil, fmt.Errorf("newC[%d]: %w", i, err)
		}
		points = append(points, op)
		compressed = append(compressed, append([]byte(nil), a.NewC[i]...))
	}
	for i := 0; i < ell; i++ {
		op, err := dec(a.NewD[i])
		if err != nil {
			return nil, fmt.Errorf("newD[%d]: %w", i, err)
		}
		points = append(points, op)
		compressed = append(compressed, append([]byte(nil), a.NewD[i]...))
	}
	if hasAuditor {
		var ekAud ristretto255.Element
		if err := ekAud.Decode(a.AuditorPub32); err != nil {
			return nil, err
		}
		points = append(points, &ekAud)
		compressed = append(compressed, append([]byte(nil), a.AuditorPub32...))
		for i := 0; i < ell; i++ {
			op, err := dec(a.NewDAud[i])
			if err != nil {
				return nil, fmt.Errorf("newDAud[%d]: %w", i, err)
			}
			points = append(points, op)
			compressed = append(compressed, append([]byte(nil), a.NewDAud[i]...))
		}
	}
	vScalar := caed25519.NumberToBytesLE32(caed25519.ModN(a.Amount))
	stmt := &Statement{
		Points:           points,
		CompressedPoints: compressed,
		Scalars:          [][]byte{vScalar},
	}
	witness := make([]*big.Int, 0, 1+2*ell)
	witness = append(witness, dkBig)
	for _, c := range a.NewAmountChunks {
		witness = append(witness, new(big.Int).SetUint64(c))
	}
	witness = append(witness, a.NewRandomness...)

	dst := DomainSeparator{
		ContractAddress: AptosFrameworkAddress,
		ChainID:         a.ChainID,
		ProtocolID:      []byte(withdrawProtocolID),
		SessionID:       BCSWithdrawSession(a.Sender32, a.Token32, uint64(ell), hasAuditor),
	}
	psi := makeWithdrawPsi(ell, hasAuditor)
	return Prove(dst, withdrawTypeName, psi, stmt, witness)
}

func makeWithdrawPsi(ell int, hasAuditor bool) Psi {
	return func(s *Statement, w []*big.Int) []*ristretto255.Element {
		dk := w[0]
		newA := w[1 : 1+ell]
		newR := w[1+ell : 1+2*ell]
		H := s.Points[idxWH]
		ek := s.Points[idxWEk]
		out := make([]*ristretto255.Element, 0, 2+3*ell)
		t0, _ := ca.ScalarMultElement(ek, dk)
		out = append(out, t0)
		for i := 0; i < ell; i++ {
			var ai, ri ristretto255.Scalar
			_ = ai.Decode(caed25519.NumberToBytesLE32(newA[i]))
			_ = ri.Decode(caed25519.NumberToBytesLE32(newR[i]))
			var term1, term2 ristretto255.Element
			term1.ScalarBaseMult(&ai)
			term2.ScalarMult(&ri, H)
			sum := new(ristretto255.Element)
			sum.Add(&term1, &term2)
			out = append(out, sum)
		}
		for i := 0; i < ell; i++ {
			t, _ := ca.ScalarMultElement(ek, newR[i])
			out = append(out, t)
		}
		if hasAuditor {
			ekAud := s.Points[idxEkAud(ell)]
			for i := 0; i < ell; i++ {
				t, _ := ca.ScalarMultElement(ekAud, newR[i])
				out = append(out, t)
			}
		}
		bPowers := computeBPowers(ell)
		var bal ristretto255.Element
		bal.Zero()
		sOldR := startOldR(ell)
		for i := 0; i < ell; i++ {
			coef := caed25519.ModN(new(big.Int).Mul(dk, bPowers[i]))
			t, _ := ca.ScalarMultElement(s.Points[sOldR+i], coef)
			bal.Add(&bal, t)
		}
		for i := 0; i < ell; i++ {
			coef := caed25519.ModN(new(big.Int).Mul(newA[i], bPowers[i]))
			var sc ristretto255.Scalar
			_ = sc.Decode(caed25519.NumberToBytesLE32(coef))
			var term ristretto255.Element
			term.ScalarBaseMult(&sc)
			bal.Add(&bal, &term)
		}
		balCopy := new(ristretto255.Element)
		b := bal.Encode(make([]byte, 0, 32))
		_ = balCopy.Decode(b)
		out = append(out, balCopy)
		return out
	}
}
