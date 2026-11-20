package hd

import (
	"encoding/hex"
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

	_, err = SplitPath("m/44''/0'/0'/0'")
	if err == nil {
		t.Fatal("expected error for empty segment")
	}

	_, err = SplitPath("m/44'/4294967296'/0'/0'/0'")
	if err == nil {
		t.Fatal("expected error for out-of-range segment")
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

func TestIsValidSlhDsaHardenedPath(t *testing.T) {
	t.Parallel()
	valid := []string{"m", "m/0'", DefaultDerivationPath, "m/0'/1'/2'/2'/1000000000'", "m/44'/637'/1'/0'/0'"}
	for _, path := range valid {
		if !IsValidSlhDsaHardenedPath(path) {
			t.Errorf("expected valid SLH-DSA hardened path %q", path)
		}
	}
	invalid := []string{"", "x", "m/", "m/0", "m/0'/1", "m/44'/637'/0'/0'/0", "m/'", "m/abc'", "m/2147483648'", "m/-1'", "n/0'", "m//0'"}
	for _, path := range invalid {
		if IsValidSlhDsaHardenedPath(path) {
			t.Errorf("expected invalid SLH-DSA hardened path %q", path)
		}
	}
}

func TestDeriveSlhDsaSha2128sPrivateKeyTSVectors(t *testing.T) {
	t.Parallel()
	seed, err := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	if err != nil {
		t.Fatalf("hex.DecodeString() error = %v", err)
	}
	vectors := []struct{ path, want string }{
		{"m", "53c14b2681fcca8600d3ac7ce3459d6ceece7abc0a3b4c0376fc845c1b6503d66c5107d9484171c02f9eb0409849fceb"},
		{"m/0'", "473049c70d5414379363b94f52de6b194c9f1651cb6a11b80cfee19f6ca67975201cdbec137ff57e2e9cd434fca0680d"},
		{"m/0'/1'/2'/2'/1000000000'", "c72d4e3fb352a785aa375be20e0767ae20edbce1e18de90e0ea6e9d1677d17e84a410bb833bf091655dfabde17cc2011"},
	}
	for _, v := range vectors {
		got, err := DeriveSlhDsaSha2128sPrivateKey(v.path, seed)
		if err != nil {
			t.Fatalf("DeriveSlhDsaSha2128sPrivateKey(%q) error = %v", v.path, err)
		}
		if len(got) != SlhDsaSha2128sPrivateKeyLength {
			t.Fatalf("DeriveSlhDsaSha2128sPrivateKey(%q) length = %d, want %d", v.path, len(got), SlhDsaSha2128sPrivateKeyLength)
		}
		if hex.EncodeToString(got) != v.want {
			t.Errorf("DeriveSlhDsaSha2128sPrivateKey(%q) = %x, want %s", v.path, got, v.want)
		}
	}
}

func TestDeriveSlhDsaSha2128sPrivateKeyErrors(t *testing.T) {
	t.Parallel()
	seed, err := MnemonicToSeed(testMnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed() error = %v", err)
	}
	for _, path := range []string{"", "m/0/1'", "not-a-path", "m/2147483648'", "m/abc'"} {
		if _, err := DeriveSlhDsaSha2128sPrivateKey(path, seed); err == nil {
			t.Errorf("expected error for path %q", path)
		}
	}
	// Different paths and the master path must all derive distinct keys.
	seen := map[string]string{}
	for _, path := range []string{"m", DefaultDerivationPath, "m/44'/637'/1'/0'/0'"} {
		got, err := DeriveSlhDsaSha2128sPrivateKey(path, seed)
		if err != nil {
			t.Fatalf("DeriveSlhDsaSha2128sPrivateKey(%q) error = %v", path, err)
		}
		if prev, dup := seen[hex.EncodeToString(got)]; dup {
			t.Errorf("path %q derived the same key as %q", path, prev)
		}
		seen[hex.EncodeToString(got)] = path
	}
	if Slip10SlhDsaSha2128sSeed != "SLH-DSA-SHA2-128s seed" {
		t.Errorf("unexpected SLIP-0010 seed string %q", Slip10SlhDsaSha2128sSeed)
	}
}

func TestParseSegment(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in      string
		want    uint32
		wantErr bool
	}{
		{"0", 0, false},
		{"44", 44, false},
		{"2147483647", 2147483647, false},
		{"", 0, true},
		{"abc", 0, true},
		{"12a", 0, true},
		{"2147483648", 0, true},
	} {
		got, err := parseSegment(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseSegment(%q): expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSegment(%q) error = %v", tc.in, err)
		} else if got != tc.want {
			t.Errorf("parseSegment(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
