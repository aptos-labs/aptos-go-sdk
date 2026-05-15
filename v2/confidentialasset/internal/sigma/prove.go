package sigma

import (
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/gtank/ristretto255"
)

// Statement holds decompressed points + compressed + optional scalars (TS SigmaProtocolStatement).
type Statement struct {
	Points           []*ristretto255.Element
	CompressedPoints [][]byte
	Scalars          [][]byte
}

// Proof is commitment compressed points + response scalars as 32-byte LE (TS SigmaProtocolProof).
type Proof struct {
	Commitment [][]byte
	Response   [][]byte
}

// Psi applies homomorphism to witness scalars (bigints mod n, same as TS).
type Psi func(stmt *Statement, witness []*big.Int) []*ristretto255.Element

// Prove runs the generic sigma prover (TS sigmaProtocolProve).
func Prove(dst DomainSeparator, typeName string, psi Psi, stmt *Statement, witness []*big.Int) (*Proof, error) {
	k := len(witness)
	alpha, err := caed25519.GenListOfRandom(k)
	if err != nil {
		return nil, err
	}
	A := psi(stmt, alpha)
	if len(A) == 0 {
		return nil, fmt.Errorf("sigma: empty commitment")
	}
	commitment := make([][]byte, len(A))
	for i, p := range A {
		commitment[i] = p.Encode(make([]byte, 0, 32))
	}
	e, err := FiatShamirChallenge(dst, typeName, stmt.CompressedPoints, stmt.Scalars, commitment, k)
	if err != nil {
		return nil, err
	}
	var wS, aS, eW, sum ristretto255.Scalar
	response := make([][]byte, k)
	for i := 0; i < k; i++ {
		if err := wS.Decode(caed25519.NumberToBytesLE32(witness[i])); err != nil {
			return nil, fmt.Errorf("witness %d: %w", i, err)
		}
		if err := aS.Decode(caed25519.NumberToBytesLE32(alpha[i])); err != nil {
			return nil, fmt.Errorf("alpha %d: %w", i, err)
		}
		eW.Multiply(e, &wS)
		sum.Add(&aS, &eW)
		response[i] = sum.Encode(make([]byte, 0, 32))
	}
	return &Proof{Commitment: commitment, Response: response}, nil
}
