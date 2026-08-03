//go:build cgo

package native

import (
	"context"
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

func TestNewConfidentialAsset(t *testing.T) {
	cc := confidentialasset.NewClient(testutil.NewFakeClient())
	a := NewConfidentialAsset(cc)
	if a == nil || a.Client == nil {
		t.Fatal("expected wrapped native client")
	}
}

func TestRolloverPendingBalance_alreadyNormalized(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 0, isNormalized: true})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	a := NewConfidentialAsset(nc.Client)
	txs, err := a.RolloverPendingBalance(context.Background(), acct, aptos.AccountOne, RolloverOpts{FAMetadataHex: "0xa"})
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 || txs[0] == nil || !txs[0].Success {
		t.Fatalf("txs=%v", txs)
	}
}

func TestRolloverPendingBalance_normalizeFirst(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 10, isNormalized: false})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	a := NewConfidentialAsset(nc.Client)
	txs, err := a.RolloverPendingBalance(context.Background(), acct, aptos.AccountOne, RolloverOpts{
		TwistedHex:    testTwistedHex,
		FAMetadataHex: "0xa",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 2 {
		t.Fatalf("expected two txs (normalize + rollover) after normalize+rollover path, got %d", len(txs))
	}
}

func TestRolloverPendingBalance_missingTwistedKey(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{isNormalized: false})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	a := NewConfidentialAsset(nc.Client)
	_, err = a.RolloverPendingBalance(context.Background(), acct, aptos.AccountOne, RolloverOpts{FAMetadataHex: "0xa"})
	if err == nil || !strings.Contains(err.Error(), "twisted") {
		t.Fatalf("err=%v", err)
	}
}
