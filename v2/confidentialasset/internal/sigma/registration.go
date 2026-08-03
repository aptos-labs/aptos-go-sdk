package sigma

import (
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/gtank/ristretto255"
)

const (
	registrationProtocolID = "AptosConfidentialAsset/RegistrationV1"
	registrationTypeName   = "0x1::sigma_protocol_registration::Registration"
)

// BCSRegistrationSession matches Move RegistrationSession (TS bcsSerializeRegistrationSession).
func BCSRegistrationSession(sender32, token32 [32]byte) []byte {
	b := make([]byte, 0, 64)
	b = append(b, sender32[:]...)
	b = append(b, token32[:]...)
	return b
}

// ProveRegistration is TS proveRegistration.
func ProveRegistration(dk32 [32]byte, sender32, token32 [32]byte, chainID byte) (*Proof, error) {
	dkBig := caed25519.ModN(caed25519.BytesToNumberLE(dk32[:]))
	ekBytes, err := ca.TwistedPublicKeyFromPrivateLE32(dk32)
	if err != nil {
		return nil, err
	}
	var ek ristretto255.Element
	if err := ek.Decode(ekBytes); err != nil {
		return nil, err
	}
	H := ca.HRistretto
	stmt := &Statement{
		Points:           []*ristretto255.Element{H, &ek},
		CompressedPoints: [][]byte{H.Encode(make([]byte, 0, 32)), ekBytes},
		Scalars:          nil,
	}
	witness := []*big.Int{dkBig}
	dst := DomainSeparator{
		ContractAddress: AptosFrameworkAddress,
		ChainID:         chainID,
		ProtocolID:      []byte(registrationProtocolID),
		SessionID:       BCSRegistrationSession(sender32, token32),
	}
	psi := func(s *Statement, w []*big.Int) []*ristretto255.Element {
		if len(w) < 1 {
			return nil
		}
		out, err := ca.ScalarMultElement(s.Points[1], w[0])
		if err != nil {
			return nil
		}
		return []*ristretto255.Element{out}
	}
	p, err := Prove(dst, registrationTypeName, psi, stmt, witness)
	if err != nil {
		return nil, fmt.Errorf("registration prove: %w", err)
	}
	return p, nil
}
