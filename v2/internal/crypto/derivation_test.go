package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto/hd"
)

func TestEd25519PrivateKeyFromDerivationPath(t *testing.T) {
	t.Parallel()
	const mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	key, err := Ed25519PrivateKeyFromDerivationPath(mnemonic, hd.DefaultDerivationPath, "")
	if err != nil {
		t.Fatalf("Ed25519PrivateKeyFromDerivationPath() error = %v", err)
	}
	wantHex := "0xcc92c0eaf80206d817f150e21917f797e49cf644a33ac514de3c316baa2f1bf5"
	if key.ToHex() != wantHex {
		t.Errorf("ToHex() = %s, want %s", key.ToHex(), wantHex)
	}
}

func TestEd25519PrivateKeyFromDerivationPathInvalidMnemonic(t *testing.T) {
	t.Parallel()
	_, err := Ed25519PrivateKeyFromDerivationPath("invalid mnemonic", hd.DefaultDerivationPath, "")
	if err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}

func TestEd25519PrivateKeyFromDerivationPathInvalidPath(t *testing.T) {
	t.Parallel()
	const mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	_, err := Ed25519PrivateKeyFromDerivationPath(mnemonic, "m/44'/637'/0'/0'/0'/1'", "")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestValidateMnemonic(t *testing.T) {
	t.Parallel()
	if !ValidateMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about") {
		t.Fatal("expected valid mnemonic")
	}
	if ValidateMnemonic("not a valid mnemonic phrase") {
		t.Fatal("expected invalid mnemonic")
	}
}
