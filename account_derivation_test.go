package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/crypto"
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

func TestNewEd25519AccountFromDerivationPath(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519AccountFromDerivationPath(testMnemonic, crypto.DefaultDerivationPath)
	if err != nil {
		t.Fatalf("NewEd25519AccountFromDerivationPath() error = %v", err)
	}
	wantAddress := "0xeb663b681209e7087d681c5d3eed12aaa8e1915e7c87794542c3f96e94b3d3bf"
	if acc.Address.String() != wantAddress {
		t.Errorf("Address = %s, want %s", acc.Address.String(), wantAddress)
	}
}
