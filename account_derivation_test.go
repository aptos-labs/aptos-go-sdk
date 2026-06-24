package aptos

import (
	"testing"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestNewEd25519AccountFromMnemonic(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519AccountFromMnemonic(testMnemonic)
	if err != nil {
		t.Fatalf("NewEd25519AccountFromMnemonic() error = %v", err)
	}
	wantAddress := "0xeb663b681209e7087d681c5d3eed12aaa8e1915e7c87794542c3f96e94b3d3bf"
	if acc.Address.String() != wantAddress {
		t.Errorf("Address = %s, want %s", acc.Address.String(), wantAddress)
	}
}
