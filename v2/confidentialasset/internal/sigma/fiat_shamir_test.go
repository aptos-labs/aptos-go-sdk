package sigma

import (
	"testing"
)

func Test_scalarFromUniform64_shortInput(t *testing.T) {
	// short input triggers the SHA512 padding path
	s := scalarFromUniform64([]byte("short"))
	if s == nil {
		t.Fatal("expected non-nil scalar")
	}
}
