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
