package sigma

import (
	"crypto/sha512"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigbcs"
	"github.com/gtank/ristretto255"
)

// AptosFrameworkAddress is 0x1 as 32 raw bytes (TS APTOS_FRAMEWORK_ADDRESS).
var AptosFrameworkAddress = func() [32]byte {
	var a [32]byte
	a[31] = 1
	return a
}()

// DomainSeparator matches Move / TS sigma_protocol.DomainSeparator V1.
type DomainSeparator struct {
	ContractAddress [32]byte
	ChainID         byte
	ProtocolID      []byte
	SessionID       []byte
}

func appendDomainSeparator(dst []byte, d DomainSeparator) []byte {
	dst = sigbcs.AppendULEB128(dst, 0) // V1 variant
	dst = append(dst, d.ContractAddress[:]...)
	dst = sigbcs.AppendU8(dst, d.ChainID)
	dst = sigbcs.AppendBytes(dst, d.ProtocolID)
	return sigbcs.AppendBytes(dst, d.SessionID)
}

// AppendFiatShamirInputs BCS-serializes FiatShamirInputs (TS BcsFiatShamirInputs).
func AppendFiatShamirInputs(dst []byte, d DomainSeparator, typeName string, k uint64, stmtX, stmtx, proofA [][]byte) []byte {
	dst = appendDomainSeparator(dst, d)
	dst = sigbcs.AppendBytes(dst, []byte(typeName))
	dst = sigbcs.AppendU64LE(dst, k)
	dst = sigbcs.AppendULEB128(dst, uint32(len(stmtX)))
	for _, p := range stmtX {
		dst = sigbcs.AppendBytes(dst, p)
	}
	dst = sigbcs.AppendULEB128(dst, uint32(len(stmtx)))
	for _, s := range stmtx {
		dst = sigbcs.AppendBytes(dst, s)
	}
	dst = sigbcs.AppendULEB128(dst, uint32(len(proofA)))
	for _, a := range proofA {
		dst = sigbcs.AppendBytes(dst, a)
	}
	return dst
}

func scalarFromUniform64(hash []byte) *big.Int {
	if len(hash) < 64 {
		h := sha512.Sum512(hash)
		hash = h[:]
	}
	return caed25519.ModN(caed25519.BytesToNumberLE(hash[:64]))
}

// FiatShamirChallenge returns e (TS sigmaProtocolFiatShamir, only e is used by the prover).
func FiatShamirChallenge(dst DomainSeparator, typeName string, stmtCompressed [][]byte, stmtScalars [][]byte, compressedA [][]byte, k int) (*ristretto255.Scalar, error) {
	buf := AppendFiatShamirInputs(nil, dst, typeName, uint64(k), stmtCompressed, stmtScalars, compressedA)
	seed := sha512.Sum512(buf)
	eInput := make([]byte, len(seed)+1)
	copy(eInput, seed[:])
	eInput[len(seed)] = 0x00
	eHash := sha512.Sum512(eInput)
	eBig := scalarFromUniform64(eHash[:])
	var e ristretto255.Scalar
	if err := e.Decode(caed25519.NumberToBytesLE32(eBig)); err != nil {
		return nil, err
	}
	return &e, nil
}
