package confidentialasset

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

func TestRegisterBalance_submit(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("100000000000"))
	}))
	defer srv.Close()
	fc := testutil.NewFakeClient()
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	fc.WithAccount(acct.Address(), &aptos.AccountInfo{SequenceNumber: 0})
	cc := NewClient(fc, WithRESTBaseURL(srv.URL))
	ctx := context.Background()
	tx, err := cc.RegisterBalance(ctx, acct, aptos.AccountOne, testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestRegisterBalance_wrongSigner(t *testing.T) {
	t.Parallel()
	cc, _ := newTestConfidentialClient()
	_, err := cc.RegisterBalance(context.Background(), wrongSigner{aptos.AccountOne}, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRotateEncryptionKey_submit(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("100000000000"))
	}))
	defer srv.Close()
	fc := testutil.NewFakeClient()
	fc.WithViewFunc(testViewFunc)
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	fc.WithAccount(acct.Address(), &aptos.AccountInfo{SequenceNumber: 0})
	newKey := "0xabababababababababababababababababababababababababababababababab"
	cc := NewClient(fc, WithRESTBaseURL(srv.URL))
	ctx := context.Background()
	tx, err := cc.RotateEncryptionKey(ctx, acct, aptos.AccountOne, testTwistedHex, newKey, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

type wrongSigner struct{ aptos.AccountAddress }

func (w wrongSigner) Address() aptos.AccountAddress { return w.AccountAddress }
func (w wrongSigner) Sign([]byte) (*aptos.AccountAuthenticator, error) { return nil, nil }
func (w wrongSigner) SignMessage([]byte) (aptos.Signature, error)       { return nil, nil }
func (w wrongSigner) SimulationAuthenticator() *aptos.AccountAuthenticator {
	return nil
}
func (w wrongSigner) AuthKey() *aptos.AuthenticationKey { return nil }
func (w wrongSigner) PubKey() aptos.PublicKey             { return nil }
