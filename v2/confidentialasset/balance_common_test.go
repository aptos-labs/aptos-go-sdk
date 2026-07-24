package confidentialasset

import (
	"context"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
)

func TestTwistedDecryptionKey32_fromHex(t *testing.T) {
	t.Parallel()
	var acct account.Account
	got, err := TwistedDecryptionKey32(&acct, testTwistedHex)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 0x01 {
		t.Fatalf("first byte=%#x", got[0])
	}
}

func TestTwistedDecryptionKey32_fromAccount(t *testing.T) {
	t.Parallel()
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	got, err := TwistedDecryptionKey32(acct, "")
	if err != nil {
		t.Fatal(err)
	}
	if got == ([32]byte{}) {
		t.Fatal("expected derived key")
	}
}

func TestFetchBalanceCipherChunks(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(ctx context.Context, payload *aptos.ViewPayload, opts ...aptos.ViewOption) ([]any, error) {
		if payload.Function == "get_available_balance" {
			return testCipherViewJSON(), nil
		}
		return testViewFunc(ctx, payload, opts...)
	})
	ctx := context.Background()
	token := aptos.AccountOne
	c, d, err := cc.FetchBalanceCipherChunks(ctx, aptos.AccountOne, token, "get_available_balance")
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 1 || len(d) != 1 || len(c[0]) != 32 {
		t.Fatalf("c=%d d=%d", len(c), len(d))
	}
}
