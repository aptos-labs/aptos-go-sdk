//go:build !cgo

package confidentialasset

import (
	"context"
	"errors"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
)

// Ensures !cgo stubs never return (nil, nil) for proof-heavy entrypoints (regression guard).
func TestStub_normalizeWithdrawTransfer_ErrCGODisabled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	c := NewClient(nil)
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	token := aptos.AccountOne
	meta := "0x1"

	_, err = c.NormalizeBalance(ctx, acct, token, "", meta)
	if !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("NormalizeBalance: want ErrCGODisabled, got %v", err)
	}
	_, err = c.Withdraw(ctx, acct, token, 1, aptos.AccountZero, "", meta)
	if !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("Withdraw: want ErrCGODisabled, got %v", err)
	}
	_, err = c.Transfer(ctx, acct, token, 1, aptos.AccountZero, "", meta)
	if !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("Transfer: want ErrCGODisabled, got %v", err)
	}
}
