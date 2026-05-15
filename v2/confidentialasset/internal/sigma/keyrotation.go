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
	keyRotationProtocolID = "AptosConfidentialAsset/KeyRotationV1"
	keyRotationTypeName   = "0x1::sigma_protocol_key_rotation::KeyRotation"
)

// BCSKeyRotationSession matches Move KeyRotationSession.
func BCSKeyRotationSession(sender32, token32 [32]byte, numChunks uint64) []byte {
	b := make([]byte, 0, 72)
	b = append(b, sender32[:]...)
	b = append(b, token32[:]...)
	b = sigbcs.AppendU64LE(b, numChunks)
	return b
}

const (
	krIdxH = 0
	krIdxEk = 1
	krIdxEkNew = 2
	krStartOldD = 3
)

// KeyRotationProof bundles rotate_encryption_key_raw proof data.
type KeyRotationProof struct {
	NewEkBytes []byte
	NewDBytes  [][]byte
	Proof      *Proof
}

// ProveKeyRotation is TS ConfidentialKeyRotation.authorizeKeyRotation (sigma part + new D).
func ProveKeyRotation(oldDk32, newDk32 [32]byte, oldD [][]byte, sender32, token32 [32]byte, chainID byte) (*KeyRotationProof, error) {
	num := ca.AvailableBalanceChunkCount
	if len(oldD) != num {
		return nil, fmt.Errorf("key rotation: want %d D chunks", num)
	}
	oldDk := caed25519.ModN(caed25519.BytesToNumberLE(oldDk32[:]))
	newDk := caed25519.ModN(caed25519.BytesToNumberLE(newDk32[:]))
	newDkInv := new(big.Int).ModInverse(newDk, caed25519.Order)
	if newDkInv == nil {
		return nil, fmt.Errorf("key rotation: non-invertible new dk")
	}
	delta := caed25519.ModN(new(big.Int).Mul(oldDk, newDkInv))
	deltaInv := new(big.Int).ModInverse(delta, caed25519.Order)
	if deltaInv == nil {
		return nil, fmt.Errorf("key rotation: non-invertible delta")
	}
	oldEkBytes, err := ca.TwistedPublicKeyFromPrivateLE32(oldDk32)
	if err != nil {
		return nil, err
	}
	var oldEk ristretto255.Element
	if err := oldEk.Decode(oldEkBytes); err != nil {
		return nil, err
	}
	H := ca.HRistretto
	var deltaSc, deltaInvSc ristretto255.Scalar
	if err := deltaSc.Decode(caed25519.NumberToBytesLE32(delta)); err != nil {
		return nil, err
	}
	if err := deltaInvSc.Decode(caed25519.NumberToBytesLE32(deltaInv)); err != nil {
		return nil, err
	}
	var newEk ristretto255.Element
	newEk.ScalarMult(&deltaSc, &oldEk)
	newEkBytes := newEk.Encode(make([]byte, 0, 32))

	oldDpts := make([]*ristretto255.Element, num)
	oldDcompressed := make([][]byte, num)
	for i := range oldD {
		np := new(ristretto255.Element)
		if err := np.Decode(oldD[i]); err != nil {
			return nil, fmt.Errorf("oldD[%d]: %w", i, err)
		}
		oldDpts[i] = np
		oldDcompressed[i] = append([]byte(nil), oldD[i]...)
	}
	newDpts := make([]*ristretto255.Element, num)
	newDcompressed := make([][]byte, num)
	newDBytes := make([][]byte, num)
	for i := range oldDpts {
		var nd ristretto255.Element
		nd.ScalarMult(&deltaSc, oldDpts[i])
		b := nd.Encode(make([]byte, 0, 32))
		np := new(ristretto255.Element)
		if err := np.Decode(b); err != nil {
			return nil, err
		}
		newDpts[i] = np
		newDcompressed[i] = append([]byte(nil), b...)
		newDBytes[i] = append([]byte(nil), b...)
	}

	points := []*ristretto255.Element{H, &oldEk, &newEk}
	compressed := [][]byte{
		H.Encode(make([]byte, 0, 32)),
		append([]byte(nil), oldEkBytes...),
		append([]byte(nil), newEkBytes...),
	}
	for i := 0; i < num; i++ {
		points = append(points, oldDpts[i])
		compressed = append(compressed, oldDcompressed[i])
	}
	for i := 0; i < num; i++ {
		points = append(points, newDpts[i])
		compressed = append(compressed, newDcompressed[i])
	}
	stmt := &Statement{Points: points, CompressedPoints: compressed, Scalars: nil}
	witness := []*big.Int{oldDk, delta, deltaInv}
	dst := DomainSeparator{
		ContractAddress: AptosFrameworkAddress,
		ChainID:         chainID,
		ProtocolID:      []byte(keyRotationProtocolID),
		SessionID:       BCSKeyRotationSession(sender32, token32, uint64(num)),
	}
	psi := makeKeyRotationPsi(num)
	p, err := Prove(dst, keyRotationTypeName, psi, stmt, witness)
	if err != nil {
		return nil, err
	}
	return &KeyRotationProof{NewEkBytes: newEkBytes, NewDBytes: newDBytes, Proof: p}, nil
}

func makeKeyRotationPsi(numChunks int) Psi {
	return func(s *Statement, w []*big.Int) []*ristretto255.Element {
		dk := w[0]
		delta := w[1]
		deltaInv := w[2]
		ek := s.Points[krIdxEk]
		newEk := s.Points[krIdxEkNew]
		var out []*ristretto255.Element
		t0, _ := ca.ScalarMultElement(ek, dk)
		out = append(out, t0)
		t1, _ := ca.ScalarMultElement(ek, delta)
		out = append(out, t1)
		t2, _ := ca.ScalarMultElement(newEk, deltaInv)
		out = append(out, t2)
		for i := 0; i < numChunks; i++ {
			t, _ := ca.ScalarMultElement(s.Points[krStartOldD+i], delta)
			out = append(out, t)
		}
		return out
	}
}
