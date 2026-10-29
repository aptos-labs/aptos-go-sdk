package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/crypto/hd"
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

func TestSlhDsaSha2128sPrivateKeyFromDerivationPath(t *testing.T) {
	t.Parallel()
	const mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	key, err := SlhDsaSha2128sPrivateKeyFromDerivationPath(mnemonic, hd.DefaultDerivationPath, "")
	if err != nil {
		t.Fatalf("SlhDsaSha2128sPrivateKeyFromDerivationPath() error = %v", err)
	}
	if len(key.Bytes()) != SlhDsaSha2128sPrivateKeyLength {
		t.Fatalf("expected %d-byte key, got %d", SlhDsaSha2128sPrivateKeyLength, len(key.Bytes()))
	}
	key2, err := SlhDsaSha2128sPrivateKeyFromDerivationPath(mnemonic, hd.DefaultDerivationPath, "")
	if err != nil {
		t.Fatalf("second derivation error = %v", err)
	}
	if key.ToHex() != key2.ToHex() {
		t.Fatal("derivation is not deterministic")
	}
	if _, err := SlhDsaSha2128sPrivateKeyFromDerivationPath(mnemonic, "m/44'/637'/0'/0'/0", ""); err == nil {
		t.Fatal("expected error for non-hardened path")
	}
	if _, err := SlhDsaSha2128sPrivateKeyFromDerivationPath("invalid mnemonic", hd.DefaultDerivationPath, ""); err == nil {
		t.Fatal("expected error for invalid mnemonic")
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
