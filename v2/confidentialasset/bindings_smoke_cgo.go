//go:build cgo

package confidentialasset

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aptos-labs/confidential-asset-bindings/bindings/go/aptosconfidential"
)

// ErrBindingsSmokeSkipped is returned when env SKIP_CONFIDENTIAL_BINDINGS prevents the smoke run.
var ErrBindingsSmokeSkipped = errors.New("confidentialasset: SKIP_CONFIDENTIAL_BINDINGS")

// Golden Pedersen bases from confidential-asset-bindings/tests/fixtures/golden_batch_range_proof.json.
var (
	valBaseGoldenBatchSmoke, _  = hex.DecodeString("e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76")
	randBaseGoldenBatchSmoke, _ = hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134")
)

// RunBindingsBatchSmoke runs one aptosconfidential BatchRangeProof → BatchVerifyProof round-trip
// against the golden bases (same check as confidential-asset-bindings/examples/go).
//
// If env SKIP_CONFIDENTIAL_BINDINGS is 1/true/yes, returns ErrBindingsSmokeSkipped without running FFI.
func RunBindingsBatchSmoke() error {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("SKIP_CONFIDENTIAL_BINDINGS"))); v == "1" || v == "true" || v == "yes" {
		return fmt.Errorf("%w", ErrBindingsSmokeSkipped)
	}
	blinding := make([]byte, 32)
	if _, err := rand.Read(blinding); err != nil {
		return err
	}
	blindingsFlat, err := aptosconfidential.FlattenBlindings([][]byte{blinding})
	if err != nil {
		return fmt.Errorf("flatten blindings: %w", err)
	}
	proof, commsFlat, err := aptosconfidential.BatchRangeProof(
		[]uint64{42},
		blindingsFlat,
		valBaseGoldenBatchSmoke,
		randBaseGoldenBatchSmoke,
		32,
	)
	if err != nil {
		return fmt.Errorf("batch range proof: %w", err)
	}
	ok, err := aptosconfidential.BatchVerifyProof(proof, commsFlat, valBaseGoldenBatchSmoke, randBaseGoldenBatchSmoke, 32)
	if err != nil {
		return fmt.Errorf("batch verify: %w", err)
	}
	if !ok {
		return errors.New("batch verify unexpectedly false")
	}
	return nil
}
