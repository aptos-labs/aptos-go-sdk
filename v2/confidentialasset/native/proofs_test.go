//go:build cgo

package native

import (
	"context"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
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
	nc := Wrap(confidentialasset.NewClient(testutil.NewFakeClient()))
	_, err := nc.NormalizeBalance(context.Background(), wrongSigner{}, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeBalance_submit(t *testing.T) {
	skipIfBindingsDisabled(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("100000000000"))
	}))
	defer srv.Close()

	var dk [32]byte
	raw, _ := hex.DecodeString(strings.TrimPrefix(testTwistedHex, "0x"))
	copy(dk[:], raw)
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	viewFn, err := cipherViewFunc8(ek, 10)
	if err != nil {
		t.Fatal(err)
	}
	fc := testutil.NewFakeClient().WithViewFunc(viewFn)
	cc := confidentialasset.NewClient(fc, confidentialasset.WithRESTBaseURL(srv.URL))
	nc := Wrap(cc)
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	fc.WithAccount(acct.Address(), &aptos.AccountInfo{SequenceNumber: 0})
	tx, err := nc.NormalizeBalance(context.Background(), acct, aptos.AccountOne, testTwistedHex, "0xa")
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil || !tx.Success {
		t.Fatalf("tx=%+v", tx)
	}
}

func TestWithdraw_insufficientViaViews(t *testing.T) {
	skipIfBindingsDisabled(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("100000000000"))
	}))
	defer srv.Close()

	var dk [32]byte
	raw, _ := hex.DecodeString(strings.TrimPrefix(testTwistedHex, "0x"))
	copy(dk[:], raw)
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	viewFn, err := cipherViewFunc8(ek, 10)
	if err != nil {
		t.Fatal(err)
	}
	fc := testutil.NewFakeClient().WithViewFunc(viewFn)
	cc := confidentialasset.NewClient(fc, confidentialasset.WithRESTBaseURL(srv.URL))
	nc := Wrap(cc)
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	fc.WithAccount(acct.Address(), &aptos.AccountInfo{SequenceNumber: 0})
	_, err = nc.Withdraw(context.Background(), acct, aptos.AccountOne, 1000, aptos.AccountOne, testTwistedHex, "0xa")
	if err == nil || !strings.Contains(err.Error(), "insufficient") {
		t.Fatalf("err=%v", err)
	}
}

func cipherViewFunc8(ek []byte, amount uint64) (func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error), error) {
	chunks := ca.AmountToChunks(amount, ca.AvailableBalanceChunkCount)
	cs := make([][]byte, len(chunks))
	ds := make([][]byte, len(chunks))
	for i, ch := range chunks {
		r := new(big.Int).SetInt64(int64(i + 3))
		c, d, err := ca.EncryptTwistedElGamal(ch, ek, r)
		if err != nil {
			return nil, err
		}
		cs[i], ds[i] = c, d
	}
	return func(_ context.Context, payload *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		point := func(b []byte) map[string]any {
			return map[string]any{"data": "0x" + hex.EncodeToString(b)}
		}
		points := make([]any, len(cs))
		rpoints := make([]any, len(ds))
		for i := range cs {
			points[i] = point(cs[i])
			rpoints[i] = point(ds[i])
		}
		row := []any{map[string]any{"P": points, "R": rpoints}}
		switch payload.Function {
		case "get_available_balance", "get_pending_balance":
			return row, nil
		case "is_normalized":
			return []any{true}, nil
		case "get_effective_auditor_config":
			return []any{map[string]any{
				"config": map[string]any{
					"ek": map[string]any{"vec": []any{}},
				},
			}}, nil
		default:
			return []any{}, nil
		}
	}, nil
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
