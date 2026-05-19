//go:build cgo

package native

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/aptos-labs/confidential-asset-bindings/bindings/go/aptosconfidential"
)

// Golden Pedersen bases from confidential-asset-bindings/tests/fixtures/golden_batch_range_proof.json.
var (
	valBaseGoldenBatchSmoke, _  = hex.DecodeString("e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76")
	randBaseGoldenBatchSmoke, _ = hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134")
)

func skipIfBindingsDisabled(t *testing.T) {
	t.Helper()
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SKIP_CONFIDENTIAL_BINDINGS")))
	if v == "1" || v == "true" || v == "yes" {
		t.Skip("SKIP_CONFIDENTIAL_BINDINGS set")
	}
}

// Same as confidential-asset-bindings/examples/go/smoke_test.go — FFI linked and Solver constructs.
func TestFFILinkedAndSolverConstructs(t *testing.T) {
	skipIfBindingsDisabled(t)
	s := aptosconfidential.NewSolver()
	if s == nil {
		t.Fatal("expected non-nil solver")
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Errorf("Solver.Close: %v", err)
		}
	}()
	runtime.KeepAlive(s)
}

// Golden BatchRangeProof → BatchVerifyProof round-trip (bindings FFI linkage smoke).
func TestBatchRangeProofGoldenRoundTrip(t *testing.T) {
	skipIfBindingsDisabled(t)
	blinding := make([]byte, 32)
	if _, err := rand.Read(blinding); err != nil {
		t.Fatal(err)
	}
	blindingsFlat, err := aptosconfidential.FlattenBlindings([][]byte{blinding})
	if err != nil {
		t.Fatalf("flatten blindings: %v", err)
	}
	proof, commsFlat, err := aptosconfidential.BatchRangeProof(
		[]uint64{42},
		blindingsFlat,
		valBaseGoldenBatchSmoke,
		randBaseGoldenBatchSmoke,
		32,
	)
	if err != nil {
		t.Fatalf("batch range proof: %v", err)
	}
	ok, err := aptosconfidential.BatchVerifyProof(proof, commsFlat, valBaseGoldenBatchSmoke, randBaseGoldenBatchSmoke, 32)
	if err != nil {
		t.Fatalf("batch verify: %v", err)
	}
	if !ok {
		t.Fatal("batch verify unexpectedly false")
	}
}
