//go:build cgo

package native

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
)

func Test_memoArg(t *testing.T) {
	if memoArg("") == nil {
		t.Fatal("expected non-nil empty memo arg")
	}
	if memoArg("hi") == nil {
		t.Fatal("expected memo arg")
	}
}

func TestNormalizeBalance_wrongSigner(t *testing.T) {
	nc, _, _ := newSubmitReadyNativeClient(t, func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error) {
		return nil, nil
	})
	_, err := nc.NormalizeBalance(context.Background(), wrongSigner{}, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeBalance_submit(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 10, isNormalized: true})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	tx, err := nc.NormalizeBalance(context.Background(), acct, aptos.AccountOne, testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestWithdraw_wrongSigner(t *testing.T) {
	nc, _, _ := newSubmitReadyNativeClient(t, func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error) {
		return nil, nil
	})
	_, err := nc.Withdraw(context.Background(), wrongSigner{}, aptos.AccountOne, 10, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error for non-Account signer")
	}
}

func TestWithdraw_notNormalized(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 100, isNormalized: false})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	_, err = nc.Withdraw(context.Background(), acct, aptos.AccountOne, 10, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error: balance not normalized")
	}
}

func TestTransfer_wrongSigner(t *testing.T) {
	nc, _, _ := newSubmitReadyNativeClient(t, func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error) {
		return nil, nil
	})
	recipientAcct, _ := account.NewEd25519()
	_, err := nc.Transfer(context.Background(), wrongSigner{}, aptos.AccountOne, 10, recipientAcct.Address(), testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error for non-Account signer")
	}
}

func TestTransfer_zeroRecipient(t *testing.T) {
	nc, acct, _ := newSubmitReadyNativeClient(t, func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error) {
		return nil, nil
	})
	var zero aptos.AccountAddress
	_, err := nc.Transfer(context.Background(), acct, aptos.AccountOne, 10, zero, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error for zero recipient")
	}
}

func TestWithdraw_submit(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 100, isNormalized: true})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	tx, err := nc.Withdraw(context.Background(), acct, aptos.AccountOne, 50, aptos.AccountOne, testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestWithdraw_defaultRecipient(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 100, isNormalized: true})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	var zeroRecipient aptos.AccountAddress
	tx, err := nc.Withdraw(context.Background(), acct, aptos.AccountOne, 10, zeroRecipient, testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestWithdraw_withAuditor(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{
		amount:       100,
		isNormalized: true,
		auditorEKHex: testAuditorPointHex,
	})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	tx, err := nc.Withdraw(context.Background(), acct, aptos.AccountOne, 10, aptos.AccountOne, testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestWithdraw_insufficientViaViews(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 10, isNormalized: true})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	_, err = nc.Withdraw(context.Background(), acct, aptos.AccountOne, 1000, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestTransfer_submit(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	recipientEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{7})
	if err != nil {
		t.Fatal(err)
	}
	recipientAcct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{
		amount:         100,
		isNormalized:   true,
		recipientEKHex: "0x" + hex.EncodeToString(recipientEK),
	})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	tx, err := nc.Transfer(context.Background(), acct, aptos.AccountOne, 20, recipientAcct.Address(), testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestTransfer_withMemo(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	recipientEK, err := ca.TwistedPublicKeyFromPrivateLE32([32]byte{8})
	if err != nil {
		t.Fatal(err)
	}
	recipientAcct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{
		amount:         100,
		isNormalized:   true,
		recipientEKHex: "0x" + hex.EncodeToString(recipientEK),
		auditorEKHex:   testAuditorPointHex,
	})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	// transferWithMemo is unexported; exercise via same code path with memo through package test in native
	tx, err := nc.transferWithMemo(context.Background(), acct, aptos.AccountOne, 15, recipientAcct.Address(), testTwistedHex, "0xa", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestTransfer_insufficient(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	recipientEK, _ := ca.TwistedPublicKeyFromPrivateLE32([32]byte{9})
	recipientAcct, _ := account.NewEd25519()
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{
		amount:         10,
		recipientEKHex: "0x" + hex.EncodeToString(recipientEK),
	})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	_, err = nc.Transfer(context.Background(), acct, aptos.AccountOne, 1000, recipientAcct.Address(), testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected insufficient balance")
	}
}

func TestTransfer_noRecipientKey(t *testing.T) {
	senderEK := senderEKFromTwistedHex(t)
	recipientAcct, _ := account.NewEd25519()
	viewFn, err := cipherViewFunc8WithViews(senderEK, aptos.AccountOne, cipherViewOpts{amount: 100})
	if err != nil {
		t.Fatal(err)
	}
	nc, acct, _ := newSubmitReadyNativeClient(t, viewFn)
	_, err = nc.Transfer(context.Background(), acct, aptos.AccountOne, 10, recipientAcct.Address(), testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected no encryption key error")
	}
}

type wrongSigner struct{ aptos.AccountAddress }

func (w wrongSigner) Address() aptos.AccountAddress                    { return w.AccountAddress }
func (w wrongSigner) Sign([]byte) (*aptos.AccountAuthenticator, error) { return nil, nil }
func (w wrongSigner) SignMessage([]byte) (aptos.Signature, error)      { return nil, nil }
func (w wrongSigner) SimulationAuthenticator() *aptos.AccountAuthenticator {
	return nil
}
func (w wrongSigner) AuthKey() *aptos.AuthenticationKey { return nil }
func (w wrongSigner) PubKey() aptos.PublicKey           { return nil }
