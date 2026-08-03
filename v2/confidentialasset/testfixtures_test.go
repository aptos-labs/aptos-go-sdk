package confidentialasset

import (
	"context"
	"encoding/json"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

const (
	testPointP = "e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76"
	testPointR = "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"
)

// testTwistedHex is a fixed 32-byte twisted decryption key for tests.
const testTwistedHex = "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

func testCipherViewJSON() []any {
	raw := []byte(`[{"P":[{"data":"0x` + testPointP + `"}],"R":[{"data":"0x` + testPointR + `"}]}]`)
	var top []any
	if err := json.Unmarshal(raw, &top); err != nil {
		panic(err)
	}
	return top
}

// testCipherViewJSON8 returns one row with eight P/R chunks (key rotation needs 8 D rows).
func testCipherViewJSON8() []any {
	points := make([]any, 8)
	rpoints := make([]any, 8)
	for i := 0; i < 8; i++ {
		points[i] = map[string]any{"data": "0x" + testPointP}
		rpoints[i] = map[string]any{"data": "0x" + testPointR}
	}
	return []any{map[string]any{"P": points, "R": rpoints}}
}

func newTestConfidentialClient() (*Client, *testutil.FakeClient) {
	fc := testutil.NewFakeClient()
	cc := NewClient(fc, WithRESTBaseURL("http://unused"))
	fc.WithViewFunc(testViewFunc)
	return cc, fc
}

func testViewFunc(_ context.Context, payload *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
	switch payload.Function {
	case "has_confidential_store", "is_normalized", "incoming_transfers_paused", "is_emergency_paused":
		return []any{true}, nil
	case "get_encryption_key":
		return []any{map[string]any{"data": "0x" + testPointP}}, nil
	case "get_effective_auditor_hint":
		return []any{map[string]any{
			"vec": []any{map[string]any{
				"is_global": true,
				"epoch":     "42",
			}},
		}}, nil
	case "get_effective_auditor_config":
		return []any{map[string]any{
			"config": map[string]any{
				"ek": map[string]any{"vec": []any{}},
			},
		}}, nil
	case "get_max_memo_bytes":
		return []any{"256"}, nil
	case "get_available_balance", "get_pending_balance":
		return testCipherViewJSON8(), nil
	default:
		return []any{}, nil
	}
}
