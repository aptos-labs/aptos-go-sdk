//go:build cgo

package native

import (
	"context"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

const testTwistedHex = "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

const testAuditorPointHex = "0xe2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76"

func skipIfBindingsDisabled(t *testing.T) {
	t.Helper()
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SKIP_CONFIDENTIAL_BINDINGS")))
	if v == "1" || v == "true" || v == "yes" {
		t.Skip("SKIP_CONFIDENTIAL_BINDINGS set")
	}
}

func newRESTMock() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("100000000000"))
	}))
}

type cipherViewOpts struct {
	amount         uint64
	recipientEKHex string // 0x + 64 hex for get_encryption_key on recipient address
	auditorEKHex   string // non-empty enables auditor config branch
	isNormalized   bool
}

func senderEKFromTwistedHex(t *testing.T) []byte {
	t.Helper()
	raw, err := hex.DecodeString(strings.TrimPrefix(testTwistedHex, "0x"))
	if err != nil {
		t.Fatal(err)
	}
	var dk [32]byte
	copy(dk[:], raw)
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	return ek
}

func cipherViewFunc8WithViews(senderEK []byte, signerAddr aptos.AccountAddress, opts cipherViewOpts) (func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error), error) {
	chunks := ca.AmountToChunks(opts.amount, ca.AvailableBalanceChunkCount)
	cs := make([][]byte, len(chunks))
	ds := make([][]byte, len(chunks))
	for i, ch := range chunks {
		r := new(big.Int).SetInt64(int64(i + 3))
		c, d, err := ca.EncryptTwistedElGamal(ch, senderEK, r)
		if err != nil {
			return nil, err
		}
		cs[i], ds[i] = c, d
	}
	senderEKHex := "0x" + hex.EncodeToString(senderEK)
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
			return []any{opts.isNormalized}, nil
		case "get_encryption_key":
			if len(payload.Args) >= 1 {
				if s, ok := payload.Args[0].(string); ok && s != signerAddr.String() {
					if opts.recipientEKHex == "" {
						return []any{map[string]any{"data": ""}}, nil
					}
					return []any{map[string]any{"data": opts.recipientEKHex}}, nil
				}
			}
			return []any{map[string]any{"data": senderEKHex}}, nil
		case "get_effective_auditor_config":
			if opts.auditorEKHex != "" {
				return []any{map[string]any{
					"config": map[string]any{
						"ek": map[string]any{
							"vec": []any{map[string]any{"data": opts.auditorEKHex}},
						},
					},
				}}, nil
			}
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

func newSubmitReadyNativeClient(
	t *testing.T,
	viewFn func(context.Context, *aptos.ViewPayload, ...aptos.ViewOption) ([]any, error),
) (*Client, *account.Account, *testutil.FakeClient) {
	t.Helper()
	skipIfBindingsDisabled(t)
	srv := newRESTMock()
	t.Cleanup(srv.Close)
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	fc := testutil.NewFakeClient().WithViewFunc(viewFn)
	fc.WithAccount(acct.Address(), &aptos.AccountInfo{SequenceNumber: 0})
	cc := confidentialasset.NewClient(fc, confidentialasset.WithRESTBaseURL(srv.URL))
	return Wrap(cc), acct, fc
}
