package account

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto/hd"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestFromMnemonic(t *testing.T) {
	t.Parallel()
	acc, err := FromMnemonic(testMnemonic)
	if err != nil {
		t.Fatalf("FromMnemonic() error = %v", err)
	}

	// Cross-SDK test vector verified against @aptos-labs/ts-sdk Account.fromDerivationPath.
	wantAddress := "0xeb663b681209e7087d681c5d3eed12aaa8e1915e7c87794542c3f96e94b3d3bf"
	if acc.Address().String() != wantAddress {
		t.Errorf("Address() = %s, want %s", acc.Address().String(), wantAddress)
	}

	msg := []byte("aptos mnemonic derivation")
	auth, err := acc.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if !auth.Verify(msg) {
		t.Fatal("derived account signature verification failed")
	}
}

func TestFromDerivationPathCustomIndex(t *testing.T) {
	t.Parallel()
	acc0, err := FromDerivationPath(testMnemonic, DefaultDerivationPath)
	if err != nil {
		t.Fatalf("FromDerivationPath() error = %v", err)
	}
	acc1, err := FromDerivationPath(testMnemonic, "m/44'/637'/1'/0'/0'")
	if err != nil {
		t.Fatalf("FromDerivationPath() error = %v", err)
	}
	if acc0.Address() == acc1.Address() {
		t.Fatal("different account indices should produce different addresses")
	}
	wantAddress1 := "0xf867372dfec13fb6c0740d4b574363685e10e6f243e9554ffa8f6e698e940efa"
	if acc1.Address().String() != wantAddress1 {
		t.Errorf("Address() = %s, want %s", acc1.Address().String(), wantAddress1)
	}
}

func TestFromDerivationPathSingleKey(t *testing.T) {
	t.Parallel()
	legacy, err := FromDerivationPath(testMnemonic, DefaultDerivationPath)
	if err != nil {
		t.Fatalf("FromDerivationPath() error = %v", err)
	}
	singleKey, err := FromDerivationPath(testMnemonic, DefaultDerivationPath, &DerivationConfig{SingleKey: true})
	if err != nil {
		t.Fatalf("FromDerivationPath() error = %v", err)
	}
	if legacy.Address() == singleKey.Address() {
		t.Fatal("legacy and SingleKey accounts should have different addresses")
	}
	// Cross-SDK test vector verified against @aptos-labs/ts-sdk Account.fromDerivationPath({ legacy: false }).
	wantSingleKeyAddress := "0x80ce9a268bbff003590bbe815d508f9619aa29ffe300fdacad9740a68773fb75"
	if singleKey.Address().String() != wantSingleKeyAddress {
		t.Errorf("SingleKey Address() = %s, want %s", singleKey.Address().String(), wantSingleKeyAddress)
	}
}

func TestFromDerivationPathInvalidMnemonic(t *testing.T) {
	t.Parallel()
	_, err := FromMnemonic("invalid mnemonic phrase")
	if err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}

func TestFromDerivationPathInvalidPath(t *testing.T) {
	t.Parallel()
	_, err := FromDerivationPath(testMnemonic, "m/44'/60'/0'/0'/0'")
	if err == nil {
		t.Fatal("expected error for non-Aptos coin type path")
	}
}

func TestFromDerivationPathPassphrase(t *testing.T) {
	t.Parallel()
	accNoPass, err := FromMnemonic(testMnemonic)
	if err != nil {
		t.Fatalf("FromMnemonic() error = %v", err)
	}
	accWithPass, err := FromMnemonic(testMnemonic, &DerivationConfig{Passphrase: "secret"})
	if err != nil {
		t.Fatalf("FromMnemonic() error = %v", err)
	}
	if accNoPass.Address() == accWithPass.Address() {
		t.Fatal("passphrase should change the derived address")
	}
}

func TestDefaultDerivationPathConstant(t *testing.T) {
	t.Parallel()
	if DefaultDerivationPath != hd.DefaultDerivationPath {
		t.Fatalf("DefaultDerivationPath = %q, want %q", DefaultDerivationPath, hd.DefaultDerivationPath)
	}
}

func TestFromDerivationPathMultipleConfigs(t *testing.T) {
	t.Parallel()
	_, err := FromDerivationPath(testMnemonic, DefaultDerivationPath, &DerivationConfig{}, &DerivationConfig{})
	if err == nil {
		t.Fatal("expected error when multiple DerivationConfig values are provided")
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
