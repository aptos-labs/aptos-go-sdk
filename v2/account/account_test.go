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

func TestAccountSignMessage(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	msg := []byte("test message")
	sig, err := acc.SignMessage(msg)
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}
	if sig == nil {
		t.Error("SignMessage() returned nil signature")
	}
}

func TestAccountPubKey(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	pubKey := acc.PubKey()
	if pubKey == nil {
		t.Error("PubKey() returned nil")
	}
}

func TestAccountSigner(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	signer := acc.Signer()
	if signer == nil {
		t.Error("Signer() returned nil")
	}

	// Verify that the signer can sign
	msg := []byte("test")
	auth, err := signer.Sign(msg)
	if err != nil {
		t.Fatalf("Signer.Sign() error = %v", err)
	}
	if auth == nil {
		t.Error("Signer.Sign() returned nil authenticator")
	}
}

func TestNewEd25519SingleKey(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519SingleKey()
	if err != nil {
		t.Fatalf("NewEd25519SingleKey() error = %v", err)
	}

	// Verify the account works
	msg := []byte("test message")
	auth, err := acc.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if auth == nil {
		t.Error("Sign() returned nil authenticator")
	}
	if !auth.Verify(msg) {
		t.Error("Signature verification failed")
	}
}

func TestAccountToAIP80(t *testing.T) {
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	aip80, err := acc.ToAIP80()
	if err != nil {
		t.Fatalf("ToAIP80() error = %v", err)
	}
	if aip80 == "" {
		t.Error("ToAIP80() returned empty string")
	}

	// Should be able to recreate the account from AIP-80
	acc2, err := FromAIP80(aip80)
	if err != nil {
		t.Fatalf("FromAIP80() error = %v", err)
	}

	if !bytes.Equal(acc.AuthKey()[:], acc2.AuthKey()[:]) {
		t.Error("Round-trip failed: AuthKey mismatch")
	}
}

func TestSecp256k1ToAIP80_NotSupported(t *testing.T) {
	t.Parallel()
	acc, err := NewSecp256k1()
	if err != nil {
		t.Fatalf("NewSecp256k1() error = %v", err)
	}

	// Secp256k1 wrapped in SingleSigner doesn't support ToAIP80
	_, err = acc.ToAIP80()
	if err == nil {
		t.Error("Expected error for SingleSigner ToAIP80")
	}
}

func TestFromPrivateKeyHex_Ed25519(t *testing.T) {
	t.Parallel()

	// Create an Ed25519 account first
	original, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	// Get the private key hex
	privateKeyHex := original.signer.(*crypto.Ed25519PrivateKey).ToHex()

	// Recreate from hex
	acc, err := FromPrivateKeyHex(privateKeyHex)
	if err != nil {
		t.Fatalf("FromPrivateKeyHex() error = %v", err)
	}

	if !bytes.Equal(original.AuthKey()[:], acc.AuthKey()[:]) {
		t.Error("AuthKey mismatch after FromPrivateKeyHex")
	}
}

func TestFromPrivateKeyHex_Secp256k1(t *testing.T) {
	t.Parallel()

	// Create a Secp256k1 account
	secpKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatalf("GenerateSecp256k1Key() error = %v", err)
	}

	privateKeyHex := secpKey.ToHex()

	// Recreate from hex
	acc, err := FromPrivateKeyHex(privateKeyHex)
	if err != nil {
		t.Fatalf("FromPrivateKeyHex() error = %v", err)
	}

	if acc == nil {
		t.Error("FromPrivateKeyHex returned nil account")
	}
}

func TestFromPrivateKeyHex_Invalid(t *testing.T) {
	t.Parallel()

	// Test with invalid hex
	_, err := FromPrivateKeyHex("not-valid-hex")
	if err == nil {
		t.Error("Expected error for invalid hex")
	}

	// Test with too short hex
	_, err = FromPrivateKeyHex("0x1234")
	if err == nil {
		t.Error("Expected error for too short hex")
	}
}

func TestFromAIP80_Ed25519(t *testing.T) {
	t.Parallel()

	// Create an Ed25519 account
	original, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	aip80Key, err := original.ToAIP80()
	if err != nil {
		t.Fatalf("ToAIP80() error = %v", err)
	}

	// Parse it back
	acc, err := FromAIP80(aip80Key)
	if err != nil {
		t.Fatalf("FromAIP80() error = %v", err)
	}

	if !bytes.Equal(original.AuthKey()[:], acc.AuthKey()[:]) {
		t.Error("AuthKey mismatch after FromAIP80")
	}
}

func TestFromAIP80_Secp256k1(t *testing.T) {
	t.Parallel()

	// Create a Secp256k1 key and format it as AIP-80 manually
	secpKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatalf("GenerateSecp256k1Key() error = %v", err)
	}

	aip80Key, err := crypto.FormatPrivateKey(secpKey.Bytes(), crypto.PrivateKeyVariantSecp256k1)
	if err != nil {
		t.Fatalf("FormatPrivateKey() error = %v", err)
	}

	// Parse it
	acc, err := FromAIP80(aip80Key)
	if err != nil {
		t.Fatalf("FromAIP80() error = %v", err)
	}

	if acc == nil {
		t.Error("FromAIP80 returned nil account")
	}
}

func TestFromAIP80_InvalidPrefix(t *testing.T) {
	t.Parallel()

	_, err := FromAIP80("invalid-prefix-0x1234")
	if err == nil {
		t.Error("Expected error for invalid AIP-80 prefix")
	}
}

func TestFromAIP80_InvalidFormat(t *testing.T) {
	t.Parallel()

	// Valid prefix but invalid key data
	_, err := FromAIP80("ed25519-priv-invalid")
	if err == nil {
		t.Error("Expected error for invalid key data")
	}
}

func TestFromAIP80_TooShort(t *testing.T) {
	t.Parallel()

	// Too short for any key type
	_, err := FromAIP80("short")
	if err == nil {
		t.Error("Expected error for too short input")
	}
}

func TestNewEd25519_Error(t *testing.T) {
	// This tests the normal path which should succeed
	t.Parallel()
	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() unexpected error = %v", err)
	}
	if acc == nil {
		t.Error("NewEd25519() returned nil account")
	}
}

func TestNewSecp256k1_Valid(t *testing.T) {
	t.Parallel()
	acc, err := NewSecp256k1()
	if err != nil {
		t.Fatalf("NewSecp256k1() error = %v", err)
	}
	if acc == nil {
		t.Error("NewSecp256k1() returned nil account")
	}

	// Verify address is populated
	addr := acc.Address()
	var zeroAddr [32]byte
	if bytes.Equal(addr[:], zeroAddr[:]) {
		t.Error("Account address is empty")
	}
}

func TestAccount_String(t *testing.T) {
	t.Parallel()

	acc, err := NewEd25519()
	if err != nil {
		t.Fatalf("NewEd25519() error = %v", err)
	}

	str := acc.String()
	if str == "" {
		t.Error("String() returned empty")
	}
	if !bytes.Contains([]byte(str), []byte("Account{")) {
		t.Error("String() should contain 'Account{'")
	}
}
