//go:build cgo

package native

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

const testTwistedHex = "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

func TestGetBalance_decryptFixture(t *testing.T) {
	skipIfBindingsDisabled(t)
	var dk [32]byte
	raw, _ := hex.DecodeString(strings.TrimPrefix(testTwistedHex, "0x"))
	copy(dk[:], raw)
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	r := new(big.Int).SetInt64(9)
	c, d, err := ca.EncryptTwistedElGamal(42, ek, r)
	if err != nil {
		t.Fatal(err)
	}
	zero := make([]byte, 32)
	viewFn := func(_ context.Context, payload *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		point := func(b []byte) map[string]any {
			return map[string]any{"data": "0x" + hex.EncodeToString(b)}
		}
		row := func(c32, d32 []byte) []any {
			return []any{map[string]any{
				"P": []any{point(c32)},
				"R": []any{point(d32)},
			}}
		}
		switch payload.Function {
		case "get_available_balance":
			return row(c, d), nil
		case "get_pending_balance":
			return row(zero, zero), nil
		default:
			return []any{}, nil
		}
	}
	fc := testutil.NewFakeClient().WithViewFunc(viewFn)
	nc := Wrap(confidentialasset.NewClient(fc))
	acct, err := account.NewEd25519()
	if err != nil {
		t.Fatal(err)
	}
	bal, err := nc.GetBalance(context.Background(), acct, aptos.AccountOne, testTwistedHex)
	if err != nil {
		t.Fatal(err)
	}
	if bal.AvailableOctas != 42 || bal.PendingOctas != 0 {
		t.Fatalf("avail=%d pend=%d", bal.AvailableOctas, bal.PendingOctas)
	}
}

func Test_scalarModNFromTwistedKeyLE32(t *testing.T) {
	var tw [32]byte
	tw[0] = 1
	s, err := scalarModNFromTwistedKeyLE32(tw)
	if err != nil || s == nil {
		t.Fatalf("err=%v", err)
	}
}
