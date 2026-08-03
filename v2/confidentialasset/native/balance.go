//go:build cgo

package native

import (
	"context"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/confidential-asset-bindings/bindings/go/aptosconfidential"
	"github.com/gtank/ristretto255"
)

// GetBalance matches TS getBalance: views + local decrypt (requires CGO + FFI static lib).
func (c *Client) GetBalance(ctx context.Context, acct *account.Account, token aptos.AccountAddress, twistedHex string) (*confidentialasset.ConfidentialBalance, error) {
	tw, err := confidentialasset.TwistedDecryptionKey32(acct, twistedHex)
	if err != nil {
		return nil, err
	}
	modS, err := scalarModNFromTwistedKeyLE32(tw)
	if err != nil {
		return nil, err
	}

	solver := aptosconfidential.NewSolver()
	if solver == nil {
		return nil, fmt.Errorf("aptosconfidential.NewSolver returned nil (FFI missing?)")
	}
	defer func() { _ = solver.Close() }()

	availC, availD, err := c.FetchBalanceCipherChunks(ctx, acct.Address(), token, "get_available_balance")
	if err != nil {
		return nil, fmt.Errorf("get_available_balance: %w", err)
	}
	pendC, pendD, err := c.FetchBalanceCipherChunks(ctx, acct.Address(), token, "get_pending_balance")
	if err != nil {
		return nil, fmt.Errorf("get_pending_balance: %w", err)
	}

	avail, err := decryptChunkedOctas(solver, modS, availC, availD)
	if err != nil {
		return nil, fmt.Errorf("decrypt available: %w", err)
	}
	pend, err := decryptChunkedOctas(solver, modS, pendC, pendD)
	if err != nil {
		return nil, fmt.Errorf("decrypt pending: %w", err)
	}
	return &confidentialasset.ConfidentialBalance{AvailableOctas: avail, PendingOctas: pend}, nil
}

func scalarModNFromTwistedKeyLE32(twisted [32]byte) (*ristretto255.Scalar, error) {
	v := leBytesToBig(twisted[:])
	v.Mod(v, ed25519Order())
	le := bigToLE32(v)
	var s ristretto255.Scalar
	if err := s.Decode(le[:]); err != nil {
		return nil, err
	}
	return &s, nil
}

func ciphertextMGCompressed(c32, d32 []byte, modS *ristretto255.Scalar) ([]byte, error) {
	var C, D, sD, mG ristretto255.Element
	if err := C.Decode(c32); err != nil {
		return nil, err
	}
	if err := D.Decode(d32); err != nil {
		return nil, err
	}
	sD.ScalarMult(modS, &D)
	mG.Subtract(&C, &sD)
	return mG.Encode(make([]byte, 0, 32)), nil
}

func isAllZero32(b []byte) bool {
	if len(b) != 32 {
		return false
	}
	for _, x := range b {
		if x != 0 {
			return false
		}
	}
	return true
}

func decryptChunkWithSolver(solver *aptosconfidential.Solver, mg []byte) (uint64, error) {
	if isAllZero32(mg) {
		return 0, nil
	}
	v, err := solver.Solve(mg, 16)
	if err == nil {
		return v, nil
	}
	return solver.Solve(mg, 32)
}

const (
	confidentialChunkBits = 16
	maxCipherChunks       = 8
)

func decryptChunkedOctas(solver *aptosconfidential.Solver, modS *ristretto255.Scalar, c, d [][]byte) (uint64, error) {
	if len(c) != len(d) {
		return 0, fmt.Errorf("P len %d != R len %d", len(c), len(d))
	}
	if len(c) == 0 {
		return 0, nil
	}
	if len(c) > maxCipherChunks {
		return 0, fmt.Errorf("unexpected chunk count %d (max %d)", len(c), maxCipherChunks)
	}
	// A balance is not guaranteed to be normalized here (get_pending_balance accumulates deposits
	// homomorphically), so an individual chunk may exceed 16 bits. decryptChunkWithSolver recovers
	// such chunks via a 32-bit solve; bound them at MaxChunkValue and let ChunksToAmountChecked do
	// the overflow-safe reassembly so a >uint64 balance errors instead of silently wrapping.
	parts := make([]uint64, len(c))
	for i := 0; i < len(c); i++ {
		mg, err := ciphertextMGCompressed(c[i], d[i], modS)
		if err != nil {
			return 0, fmt.Errorf("chunk %d mg: %w", i, err)
		}
		part, err := decryptChunkWithSolver(solver, mg)
		if err != nil {
			return 0, fmt.Errorf("chunk %d solve: %w", i, err)
		}
		if part >= ca.MaxChunkValue {
			return 0, fmt.Errorf("chunk %d out of range: %d", i, part)
		}
		parts[i] = part
	}
	return ca.ChunksToAmountChecked(parts)
}

// decryptAvailableAmountChunks decrypts get_available_balance ciphertext into per-chunk values (CGO).
// Returns sender encryption public key (twisted ek), chunk values, and raw C/D rows for sigma statements.
func (c *Client) decryptAvailableAmountChunks(ctx context.Context, acct *account.Account, token aptos.AccountAddress, twistedHex string) (pub [32]byte, chunks []uint64, c32, d32 [][]byte, err error) {
	tw, err := confidentialasset.TwistedDecryptionKey32(acct, twistedHex)
	if err != nil {
		return pub, nil, nil, nil, err
	}
	modS, err := scalarModNFromTwistedKeyLE32(tw)
	if err != nil {
		return pub, nil, nil, nil, err
	}
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(tw)
	if err != nil {
		return pub, nil, nil, nil, err
	}
	copy(pub[:], ek)

	solver := aptosconfidential.NewSolver()
	if solver == nil {
		return pub, nil, nil, nil, fmt.Errorf("aptosconfidential.NewSolver returned nil (FFI missing?)")
	}
	defer func() { _ = solver.Close() }()

	availC, availD, err := c.FetchBalanceCipherChunks(ctx, acct.Address(), token, "get_available_balance")
	if err != nil {
		return pub, nil, nil, nil, err
	}
	if len(availC) > maxCipherChunks {
		return pub, nil, nil, nil, fmt.Errorf("unexpected chunk count %d (max %d)", len(availC), maxCipherChunks)
	}
	chunks = make([]uint64, len(availC))
	for i := range availC {
		mg, err := ciphertextMGCompressed(availC[i], availD[i], modS)
		if err != nil {
			return pub, nil, nil, nil, fmt.Errorf("chunk %d: %w", i, err)
		}
		v, err := decryptChunkWithSolver(solver, mg)
		if err != nil {
			return pub, nil, nil, nil, fmt.Errorf("chunk %d solve: %w", i, err)
		}
		if v >= (1 << confidentialChunkBits) {
			return pub, nil, nil, nil, fmt.Errorf("chunk %d out of range: %d", i, v)
		}
		chunks[i] = v
	}
	return pub, chunks, availC, availD, nil
}
