package account

import (
	"bytes"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
)

func TestNewEd25519(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	// Check that address is derived from auth key
	authKey := acc.AuthKey()
	addr := acc.Address()
	if !bytes.Equal(addr[:], authKey[:]) {
		t.Error("Address should equal AuthKey for new account")
	}

	// Test signing
	msg := []byte("test message")
	auth, err := acc.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if auth == nil {
		t.Error("Sign() returned nil authenticator")
	}

	// Verify signature
	if !auth.Verify(msg) {
		t.Error("Signature verification failed")
	}
}

func TestNewEd25519Deterministic(t *testing.T) {
	t.Parallel()
	// Create deterministic reader with known seed
	seed := bytes.NewReader(make([]byte, 32))

	acc1, err := NewEd25519(seed)
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	// Reset seed and create another account
	seed = bytes.NewReader(make([]byte, 32))
	acc2, err := NewEd25519(seed)
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	// Both should have the same address
	if acc1.Address() != acc2.Address() {
		t.Error("Deterministic accounts should have same address")
	}
}

func TestNewSecp256k1(t *testing.T) {
	t.Parallel()
	acc, err := NewSecp256k1()
	if err != nil {
		t.Fatalf("NewSecp256k1() error = %v", err)
	}

	// Check that address is derived from auth key
	authKey := acc.AuthKey()
	addr := acc.Address()
	if !bytes.Equal(addr[:], authKey[:]) {
		t.Error("Address should equal AuthKey for new account")
	}

	// Test signing
	msg := []byte("test message")
	auth, err := acc.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if auth == nil {
		t.Error("Sign() returned nil authenticator")
	}
}

func TestFromEd25519PrivateKey(t *testing.T) {
	t.Parallel()
	// Generate a key first
	key, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		t.Fatalf("GenerateEd25519PrivateKey() error = %v", err)
	}

	// Create account from key bytes
	acc, err := FromEd25519PrivateKey(key.Bytes())
	if err != nil {
		t.Fatalf("FromEd25519PrivateKey() error = %v", err)
	}

	// Should have same auth key
	if !bytes.Equal(acc.AuthKey()[:], key.AuthKey()[:]) {
		t.Error("AuthKey mismatch")
	}
}

func TestFromSecp256k1PrivateKey(t *testing.T) {
	t.Parallel()
	// Generate a key first
	key, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatalf("GenerateSecp256k1Key() error = %v", err)
	}

	// Create account from key bytes
	acc, err := FromSecp256k1PrivateKey(key.Bytes())
	if err != nil {
		t.Fatalf("FromSecp256k1PrivateKey() error = %v", err)
	}

	// Should be able to sign
	msg := []byte("test")
	auth, err := acc.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if auth == nil {
		t.Error("Sign() returned nil authenticator")
	}
}

func TestFromPrivateKeyHex(t *testing.T) {
	t.Parallel()
	// Generate Ed25519 key
	ed25519Key, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		t.Fatalf("GenerateEd25519PrivateKey() error = %v", err)
	}

	// Import from hex
	acc, err := FromPrivateKeyHex(ed25519Key.ToHex())
	if err != nil {
		t.Fatalf("FromPrivateKeyHex() error = %v", err)
	}

	if !bytes.Equal(acc.AuthKey()[:], ed25519Key.AuthKey()[:]) {
		t.Error("AuthKey mismatch")
	}
}

func TestFromAIP80(t *testing.T) {
	t.Parallel()
	// Generate Ed25519 key and format as AIP-80
	ed25519Key, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		t.Fatalf("GenerateEd25519PrivateKey() error = %v", err)
	}

	aip80Str, err := ed25519Key.ToAIP80()
	if err != nil {
		t.Fatalf("ToAIP80() error = %v", err)
	}

	// Import from AIP-80
	acc, err := FromAIP80(aip80Str)
	if err != nil {
		t.Fatalf("FromAIP80() error = %v", err)
	}

	if !bytes.Equal(acc.AuthKey()[:], ed25519Key.AuthKey()[:]) {
		t.Error("AuthKey mismatch")
	}
}

func TestFromAIP80Invalid(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no prefix", "0x1234"},
		{"invalid prefix", "unknown-priv-0x1234"},
		{"missing key", "ed25519-priv-"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := FromAIP80(tc.input)
			if err == nil {
				t.Error("Expected error for invalid AIP-80 input")
			}
		})
	}
}

func TestFromSignerWithAddress(t *testing.T) {
	t.Parallel()
	key, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		t.Fatalf("GenerateEd25519PrivateKey() error = %v", err)
	}

	// Custom address (e.g., after rotation)
	customAddr := AccountAddress{0x01}

	acc := FromSignerWithAddress(key, customAddr)

	if acc.Address() != customAddr {
		t.Error("Address should match custom address")
	}

	// But signing should still work
	msg := []byte("test")
	auth, err := acc.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if !auth.Verify(msg) {
		t.Error("Signature verification failed")
	}
}

func TestAccountString(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	str := acc.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

func TestSimulationAuthenticator(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	auth := acc.SimulationAuthenticator()
	if auth == nil {
		t.Error("SimulationAuthenticator() returned nil")
	}
}
