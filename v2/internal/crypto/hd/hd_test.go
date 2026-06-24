package hd

import (
	"strings"
	"testing"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestNormalizeMnemonic(t *testing.T) {
	t.Parallel()
	got := NormalizeMnemonic("  Abandon  abandon\nabandon abandon abandon abandon abandon abandon abandon abandon abandon about  ")
	want := testMnemonic
	if got != want {
		t.Errorf("NormalizeMnemonic() = %q, want %q", got, want)
	}
}

func TestValidateMnemonic(t *testing.T) {
	t.Parallel()
	if !ValidateMnemonic(testMnemonic) {
		t.Fatal("expected valid mnemonic")
	}
	if ValidateMnemonic("not a valid mnemonic phrase") {
		t.Fatal("expected invalid mnemonic")
	}
}

func TestIsValidHardenedPath(t *testing.T) {
	t.Parallel()
	valid := []string{
		DefaultDerivationPath,
		"m/44'/637'/0'/0'/0",
		"m/44'/637'/1'/0'/2'",
	}
	for _, path := range valid {
		if !IsValidHardenedPath(path) {
			t.Errorf("expected valid path %q", path)
		}
	}

	invalid := []string{
		"",
		"m/44'/60'/0'/0'/0'",
		"m/44'/637'/0'/0/0",
		"m/44/637/0/0/0",
		"m/44'/637'/0'/0'/0'/1'",
	}
	for _, path := range invalid {
		if IsValidHardenedPath(path) {
			t.Errorf("expected invalid path %q", path)
		}
	}
}

func TestSplitPath(t *testing.T) {
	t.Parallel()
	segments, err := SplitPath(DefaultDerivationPath)
	if err != nil {
		t.Fatalf("SplitPath() error = %v", err)
	}
	want := []uint32{44, AptosCoinType, 0, 0, 0}
	for i, seg := range segments {
		if seg != want[i] {
			t.Fatalf("segment[%d] = %d, want %d", i, seg, want[i])
		}
	}
}

func TestDeriveEd25519PrivateKeyDeterministic(t *testing.T) {
	t.Parallel()
	seed, err := MnemonicToSeed(testMnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}

	key1, err := DeriveEd25519PrivateKey(DefaultDerivationPath, seed)
	if err != nil {
		t.Fatalf("DeriveEd25519PrivateKey() error = %v", err)
	}
	key2, err := DeriveEd25519PrivateKey(DefaultDerivationPath, seed)
	if err != nil {
		t.Fatalf("DeriveEd25519PrivateKey() error = %v", err)
	}
	if len(key1) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key1))
	}
	for i := range key1 {
		if key1[i] != key2[i] {
			t.Fatal("derivation is not deterministic")
		}
	}
}

func TestDeriveEd25519PrivateKeyDifferentPaths(t *testing.T) {
	t.Parallel()
	seed, err := MnemonicToSeed(testMnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}

	key0, err := DeriveEd25519PrivateKey(DefaultDerivationPath, seed)
	if err != nil {
		t.Fatalf("DeriveEd25519PrivateKey() error = %v", err)
	}
	key1, err := DeriveEd25519PrivateKey("m/44'/637'/1'/0'/0'", seed)
	if err != nil {
		t.Fatalf("DeriveEd25519PrivateKey() error = %v", err)
	}
	for i := range key0 {
		if key0[i] == key1[i] {
			continue
		}
		return
	}
	t.Fatal("different paths should produce different keys")
}

func TestMnemonicToSeedPassphrase(t *testing.T) {
	t.Parallel()
	seedNoPass, err := MnemonicToSeed(testMnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}
	seedWithPass, err := MnemonicToSeed(testMnemonic, "secret")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}
	for i := range seedNoPass {
		if seedNoPass[i] != seedWithPass[i] {
			return
		}
	}
	t.Fatal("passphrase should change the derived seed")
}

func TestMnemonicToSeedInvalid(t *testing.T) {
	t.Parallel()
	_, err := MnemonicToSeed("not a valid mnemonic phrase", "")
	if err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}

func TestSplitPathErrors(t *testing.T) {
	t.Parallel()
	_, err := SplitPath("bad/path")
	if err == nil {
		t.Fatal("expected error for path without m/ prefix")
	}

	_, err = SplitPath("m/")
	if err == nil {
		t.Fatal("expected error for empty segments")
	}

	_, err = SplitPath("m/44'/bad'/0'/0'/0'")
	if err == nil {
		t.Fatal("expected error for non-numeric segment")
	}
}

func TestDeriveEd25519PrivateKeyInvalidPath(t *testing.T) {
	t.Parallel()
	seed, err := MnemonicToSeed(testMnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}

	_, err = DeriveEd25519PrivateKey("m/44'/637'/0'/0'/0'/1'", seed)
	if err == nil {
		t.Fatal("expected error for path with extra trailing segment")
	}
}

func TestDeriveEd25519PrivateKeyHardenedSegmentOverflow(t *testing.T) {
	t.Parallel()
	seed, err := MnemonicToSeed(testMnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}

	// 0x80000000 is the hardened offset; segments must stay below it.
	_, err = DeriveEd25519PrivateKey("m/44'/637'/2147483648'/0'/0'", seed)
	if err == nil {
		t.Fatal("expected error for segment >= hardened offset")
	}
	if !strings.Contains(err.Error(), "index 2") || !strings.Contains(err.Error(), "2147483648") {
		t.Fatalf("expected actionable segment error, got: %v", err)
	}
}
