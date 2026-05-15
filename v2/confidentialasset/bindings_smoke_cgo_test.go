//go:build cgo

package confidentialasset

import (
	"runtime"
	"testing"

	"github.com/aptos-labs/confidential-asset-bindings/bindings/go/aptosconfidential"
)

// Same as confidential-asset-bindings/examples/go/smoke_test.go — FFI linked and Solver constructs.
func TestFFILinkedAndSolverConstructs(t *testing.T) {
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

func TestRunBindingsBatchSmoke(t *testing.T) {
	t.Setenv("SKIP_CONFIDENTIAL_BINDINGS", "")
	if err := RunBindingsBatchSmoke(); err != nil {
		t.Fatal(err)
	}
}
