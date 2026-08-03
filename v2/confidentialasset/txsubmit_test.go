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

func TestFetchPublicFABalanceOctas(t *testing.T) {
	t.Parallel()
	t.Run("ok", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("1000000"))
		}))
		defer srv.Close()
		cc := NewClient(testutil.NewFakeClient(), WithRESTBaseURL(srv.URL))
		n, err := cc.FetchPublicFABalanceOctas(context.Background(), aptos.AccountOne, "0xa")
		if err != nil || n != 1_000_000 {
			t.Fatalf("n=%d err=%v", n, err)
		}
	})
	t.Run("not_found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()
		cc := NewClient(testutil.NewFakeClient(), WithRESTBaseURL(srv.URL))
		n, err := cc.FetchPublicFABalanceOctas(context.Background(), aptos.AccountOne, "0xa")
		if err != nil || n != 0 {
			t.Fatalf("n=%d err=%v", n, err)
		}
	})
	t.Run("json_string", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`"5000"`))
		}))
		defer srv.Close()
		cc := NewClient(testutil.NewFakeClient(), WithRESTBaseURL(srv.URL))
		n, err := cc.FetchPublicFABalanceOctas(context.Background(), aptos.AccountOne, "0xa")
		if err != nil || n != 5000 {
			t.Fatalf("n=%d err=%v", n, err)
		}
	})
	t.Run("missing_rest", func(t *testing.T) {
		cc := NewClient(testutil.NewFakeClient())
		_, err := cc.FetchPublicFABalanceOctas(context.Background(), aptos.AccountOne, "0xa")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestDeposit_submit(t *testing.T) {
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
	tx, err := cc.Deposit(ctx, acct, aptos.AccountOne, 100, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestRolloverPendingBalance_submit(t *testing.T) {
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
	tx, err := cc.RolloverPendingBalance(ctx, acct, aptos.AccountOne, true, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestSubmitWithSimulatedGas_feePayer(t *testing.T) {
	t.Parallel()
	cc := NewClient(testutil.NewFakeClient(), WithFeePayer(true), WithRESTBaseURL("http://x"))
	acct, _ := account.NewEd25519()
	_, err := cc.SubmitWithSimulatedGas(context.Background(), acct, "x", &aptos.EntryFunctionPayload{
		Module:   cc.ViewModule(),
		Function: "deposit",
		Args:     []any{aptos.AccountOne, uint64(1)},
	}, "0xa")
	if err == nil {
		t.Fatal("expected fee payer error")
	}
}
