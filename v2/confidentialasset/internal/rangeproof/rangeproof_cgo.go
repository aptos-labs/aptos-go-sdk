//go:build cgo

package rangeproof

import (
	"github.com/aptos-labs/confidential-asset-bindings/bindings/go/aptosconfidential"
)

// BatchRangeProof delegates to aptosconfidential (same Rust as TS WASM bindings).
func BatchRangeProof(values []uint64, blindingsFlat, valBase, randBase []byte, numBits int) (proof []byte, commsFlat []byte, err error) {
	return aptosconfidential.BatchRangeProof(values, blindingsFlat, valBase, randBase, numBits)
}
