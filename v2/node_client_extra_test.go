package aptos

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin behaviour of the v2 client paths that were either added or
// re-shaped in the SDK audit PR — specifically:
//
//   - View() must forward TypeArgs (previously the body was hardcoded to
//     `"type_arguments": []string{}`).
//   - AccountBalance() must call the 0x1::coin::balance view function rather
//     than reading 0x1::coin::CoinStore directly (the old impl 404'd on
//     fungible-asset accounts).
//   - Fund() must wait for the faucet's funding transactions to commit
//     (previously it returned as soon as POST /mint succeeded).
//   - parseFaucetHashes() must handle the localnet's bare-array body and
//     hosted faucets' {txn_hashes: [...]} body shape.
//
// They are pure HTTP-mock tests; no localnet required.

// --- View ----------------------------------------------------------------

// TestView_ForwardsTypeArguments asserts that ViewPayload.TypeArgs are
// serialized into the request body as the canonical Move type strings.
// Before the fix the body always contained an empty type_arguments array.
func TestView_ForwardsTypeArguments(t *testing.T) {
	t.Parallel()

	var captured struct {
		Function      string   `json:"function"`
		TypeArguments []string `json:"type_arguments"`
		Arguments     []any    `json:"arguments"`
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/view", r.URL.Path)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&captured))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{"100000"})
	}))

	addr := MustParseAddress("0x123")
	_, err := client.View(context.Background(), &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "coin"},
		Function: "balance",
		TypeArgs: []TypeTag{AptosCoinTypeTag},
		Args:     []any{addr.String()},
	})
	require.NoError(t, err)

	assert.Equal(t, "0x1::coin::balance", captured.Function)
	assert.Equal(t, []string{"0x1::aptos_coin::AptosCoin"}, captured.TypeArguments,
		"View must forward TypeTag.String() values, not drop them")
	require.Len(t, captured.Arguments, 1)
}

// TestView_NoTypeArgs sends an empty TypeArgs slice and asserts the body
// contains an empty (but present) type_arguments array.
func TestView_NoTypeArgs(t *testing.T) {
	t.Parallel()

	var captured struct {
		TypeArguments []string `json:"type_arguments"`
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode([]any{true})
	}))

	_, err := client.View(context.Background(), &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "account"},
		Function: "exists_at",
		Args:     []any{"0x1"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{}, captured.TypeArguments)
}

// TestView_LedgerVersion asserts AtLedgerVersion appends the query parameter.
func TestView_LedgerVersion(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ledger_version=12345", r.URL.RawQuery)
		_ = json.NewEncoder(w).Encode([]any{"42"})
	}))

	_, err := client.View(context.Background(), &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "account"},
		Function: "exists_at",
		Args:     []any{"0x1"},
	}, AtLedgerVersion(12345))
	require.NoError(t, err)
}

// --- AccountBalance ------------------------------------------------------

// TestAccountBalance_UsesViewFunction is the regression test for the
// fungible-asset bug: AccountBalance must hit /view, not /resource/.
// Returning a 404 to the resource path would surface the bug.
func TestAccountBalance_UsesViewFunction(t *testing.T) {
	t.Parallel()

	var calls struct {
		viewPath     atomic.Int32
		resourcePath atomic.Int32
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/view"):
			calls.viewPath.Add(1)
			_ = json.NewEncoder(w).Encode([]any{"123456"})
		case strings.Contains(r.URL.Path, "/resource/"):
			calls.resourcePath.Add(1)
			// Simulate the FA-account behaviour: no CoinStore.
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "resource_not_found",
				"message":    "Resource not found",
			})
		default:
			http.NotFound(w, r)
		}
	}))

	bal, err := client.AccountBalance(context.Background(), MustParseAddress("0x123"))
	require.NoError(t, err)
	assert.Equal(t, uint64(123456), bal)
	assert.Equal(t, int32(1), calls.viewPath.Load(), "should call /view")
	assert.Equal(t, int32(0), calls.resourcePath.Load(), "must NOT fall back to /resource/")
}

// TestAccountBalance_LedgerVersionForwarded asserts that a ResourceOption's
// ledger version is passed through to the underlying view call as the
// ledger_version query parameter. There is no public `WithVersion` helper
// (v2 API gap), so the option is constructed inline.
func TestAccountBalance_LedgerVersionForwarded(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "ledger_version=777")
		_ = json.NewEncoder(w).Encode([]any{"1"})
	}))

	withVersion := func(v uint64) ResourceOption {
		return func(c *ResourceConfig) { c.LedgerVersion = &v }
	}
	_, err := client.AccountBalance(context.Background(), MustParseAddress("0x1"), withVersion(777))
	require.NoError(t, err)
}

// TestAccountBalance_NumericFromNode covers the defensive float64 branch in
// case a non-conforming node returns the balance as a JSON number.
func TestAccountBalance_NumericFromNode(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, jsonHandler([]any{float64(999)}))
	bal, err := client.AccountBalance(context.Background(), AccountOne)
	require.NoError(t, err)
	assert.Equal(t, uint64(999), bal)
}

// TestAccountBalance_NumericOverflow rejects balances that exceed float64's
// exact-integer range (2^53). Without the guard, casting `uint64(v)` would
// silently truncate to a different, lower value — exactly the kind of bug
// that's invisible until someone hits a ~90M-APT account.
func TestAccountBalance_NumericOverflow(t *testing.T) {
	t.Parallel()

	// 2^53 + 2 — not exactly representable in float64; rounds to 2^53.
	bad := float64(1<<53) + 2
	client := newTestClient(t, jsonHandler([]any{bad}))
	_, err := client.AccountBalance(context.Background(), AccountOne)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "float64 exact-integer range")
}

// TestAccountBalance_NumericNonInteger rejects non-integer JSON numbers.
func TestAccountBalance_NumericNonInteger(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, jsonHandler([]any{1.5}))
	_, err := client.AccountBalance(context.Background(), AccountOne)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative integer")
}

// TestAccountBalance_NumericNegative rejects negative JSON numbers.
func TestAccountBalance_NumericNegative(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, jsonHandler([]any{float64(-1)}))
	_, err := client.AccountBalance(context.Background(), AccountOne)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative integer")
}

// TestAccountBalance_UnexpectedShape exercises the type-assertion fallback.
func TestAccountBalance_UnexpectedShape(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, jsonHandler([]any{true}))
	_, err := client.AccountBalance(context.Background(), AccountOne)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected balance value type")
}

// TestAccountBalance_EmptyView covers the "view function returned no values"
// error path.
func TestAccountBalance_EmptyView(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, jsonHandler([]any{}))
	_, err := client.AccountBalance(context.Background(), AccountOne)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no values")
}

// --- Fund / parseFaucetHashes ------------------------------------------

// TestParseFaucetHashes covers each response-shape branch of parseFaucetHashes.
func TestParseFaucetHashes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
		want []string
	}{
		{"empty body", "", nil},
		{"whitespace", "   \n\t ", nil},
		{"bare array (localnet)", `["0xabc","0xdef"]`, []string{"0xabc", "0xdef"}},
		{"empty array", `[]`, []string{}},
		{"hosted faucet object", `{"txn_hashes":["0xa","0xb"]}`, []string{"0xa", "0xb"}},
		{"unknown shape — number", `42`, nil},
		{"unknown shape — string", `"oops"`, nil},
		{"unknown shape — object", `{"foo":"bar"}`, nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseFaucetHashes([]byte(tt.body))
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestFund_WaitsForTransactions is the regression for the bug where Fund
// returned before the faucet's transactions had committed: it must call
// /transactions/by_hash for every hash returned, polling until each is no
// longer "pending_transaction".
func TestFund_WaitsForTransactions(t *testing.T) {
	t.Parallel()

	var (
		mintCalls atomic.Int32
		txByHash  atomic.Int32
	)
	hashes := []string{"0xabc", "0xdef"}

	mux := http.NewServeMux()
	mux.HandleFunc("/mint", func(w http.ResponseWriter, r *http.Request) {
		mintCalls.Add(1)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.RawQuery, "address=0x")
		assert.Contains(t, r.URL.RawQuery, "amount=1000000")
		_ = json.NewEncoder(w).Encode(hashes)
	})
	mux.HandleFunc("/transactions/by_hash/", func(w http.ResponseWriter, _ *http.Request) {
		txByHash.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":      "user_transaction",
			"hash":      "0xabc",
			"version":   "1",
			"success":   true,
			"vm_status": "Executed successfully",
		})
	})

	client := newTestClient(t, mux)
	err := client.Fund(context.Background(), MustParseAddress("0x123"), 1000000)
	require.NoError(t, err)

	assert.Equal(t, int32(1), mintCalls.Load(), "should POST /mint exactly once")
	assert.Equal(t, int64(len(hashes)), int64(txByHash.Load()),
		"should call /transactions/by_hash/<hash> for every hash the faucet emitted")
}

// TestFund_NoHashesNoWait verifies that when the faucet returns an empty
// body (or unknown shape), Fund still succeeds without trying to wait.
func TestFund_NoHashesNoWait(t *testing.T) {
	t.Parallel()

	var txCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/mint", func(w http.ResponseWriter, _ *http.Request) {
		// No body.
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/transactions/by_hash/", func(_ http.ResponseWriter, _ *http.Request) {
		txCalls.Add(1)
	})

	client := newTestClient(t, mux)
	err := client.Fund(context.Background(), MustParseAddress("0x1"), 100)
	require.NoError(t, err)
	assert.Equal(t, int32(0), txCalls.Load(), "no hashes ⇒ no wait")
}

// TestFund_BodyReadError ensures that a faucet response whose body fails
// mid-stream surfaces an error rather than silently falling through to
// "no hashes ⇒ no wait" — which would re-introduce the race the wait was
// added to fix. We use an http.Hijacker to send a Content-Length larger
// than the body and then close the connection so io.ReadAll observes EOF
// before reading the promised bytes.
func TestFund_BodyReadError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/mint", func(w http.ResponseWriter, _ *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijack")
			return
		}
		conn, bufrw, err := hijacker.Hijack()
		if !assert.NoError(t, err) {
			return
		}
		defer conn.Close()
		// Promise 1000 bytes, send 0, then drop the connection.
		// io.ReadAll on the client will see an unexpected EOF.
		_, _ = bufrw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n")
		_ = bufrw.Flush()
	})

	client := newTestClient(t, mux)
	err := client.Fund(context.Background(), MustParseAddress("0x1"), 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read faucet response")
}

// TestFund_FaucetError propagates non-2xx as APIError.
func TestFund_FaucetError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/mint", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	})

	client := newTestClient(t, mux)
	err := client.Fund(context.Background(), MustParseAddress("0x1"), 100)
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "boom")
}

// TestFund_NoFaucetURL covers the early-return error when the network
// has no faucet (e.g. mainnet).
func TestFund_NoFaucetURL(t *testing.T) {
	t.Parallel()

	c, err := newNodeClient(&ClientConfig{
		network: NetworkConfig{
			NodeURL: "http://127.0.0.1:1",
			ChainID: 1,
		},
	})
	require.NoError(t, err)

	err = c.Fund(context.Background(), AccountOne, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "faucet not available")
}

// TestFund_ContextCancelled verifies context cancellation is respected
// during the wait phase.
func TestFund_ContextCancelled(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/mint", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]string{"0xpending"})
	})
	mux.HandleFunc("/transactions/by_hash/", func(w http.ResponseWriter, _ *http.Request) {
		// Always pending.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "pending_transaction",
			"hash": "0xpending",
		})
	})

	client := newTestClient(t, mux)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := client.Fund(ctx, MustParseAddress("0x1"), 1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context"),
		"expected context cancellation error, got: %v", err)
}

// --- ChainID caching ----------------------------------------------------

// TestChainID_Cached asserts the chain ID returned by Info() is cached and
// subsequent calls do not hit the network. This protects against the easy
// regression of dropping the chainIDSet flag.
func TestChainID_Cached(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/v1"), "path: %s", r.URL.Path)
		_ = json.NewEncoder(w).Encode(NodeInfo{ChainID: 99, LedgerVersion: 1, BlockHeight: 1})
	}))
	t.Cleanup(server.Close)

	// ChainID == 0 ⇒ dynamic ⇒ first ChainID() call fetches Info.
	c, err := newNodeClient(&ClientConfig{
		network: NetworkConfig{NodeURL: server.URL + "/v1", ChainID: 0},
	})
	require.NoError(t, err)

	id, err := c.ChainID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(99), id)

	id, err = c.ChainID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(99), id)

	assert.Equal(t, int32(1), calls.Load(), "ChainID should be cached after the first fetch")
}

// TestChainID_PreconfiguredNoFetch asserts that a non-zero NetworkConfig
// ChainID is used directly without ever hitting the node.
func TestChainID_PreconfiguredNoFetch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatalf("ChainID() should not call the node when chain id is pre-configured")
	}))
	t.Cleanup(server.Close)

	c, err := newNodeClient(&ClientConfig{
		network: NetworkConfig{NodeURL: server.URL, ChainID: 4},
	})
	require.NoError(t, err)

	id, err := c.ChainID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), id)
}

// --- normalizeViewArg ---------------------------------------------------

// TestNormalizeViewArg covers the reflect-based normalization that converts
// Go-native types into the JSON shape Aptos's /view endpoint expects.
//
// In particular u64/int64 must be stringified (since JSON numbers can't hold
// the full u64 range), and []byte must pass through unchanged.
func TestNormalizeViewArg(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   any
		want any
	}{
		{"nil", nil, nil},
		{"uint64 stringified", uint64(123), "123"},
		{"int64 stringified", int64(-456), "-456"},
		{"uint64 max", uint64(1<<63 + 1), "9223372036854775809"},
		{"uint stringified", uint(789), "789"},
		{"int stringified", int(-789), "-789"},
		{"big.Int u128 stringified", new(big.Int).SetUint64(1<<63 + 1), "9223372036854775809"},
		{"big.Int nil", (*big.Int)(nil), nil},
		{"bool passes through", true, true},
		{"string passes through", "hello", "hello"},
		{"byte slice passes through", []byte{1, 2, 3}, []byte{1, 2, 3}},
		{"slice of u64 stringified element-wise", []uint64{1, 2}, []any{"1", "2"}},
		{"map with string keys recurses", map[string]any{"k": uint64(5)}, map[string]any{"k": "5"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeViewArg(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNormalizeViewArgs_NilAndEmpty pins behaviour on the slice helper.
func TestNormalizeViewArgs_NilAndEmpty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, normalizeViewArgs(nil))
	assert.Equal(t, []any{}, normalizeViewArgs([]any{}))
}

// --- Simulate variants --------------------------------------------------

// makeRawTxn builds a minimal valid RawTransaction for simulation tests.
func makeRawTxn(sender AccountAddress) *RawTransaction {
	return &RawTransaction{
		Sender:                     sender,
		SequenceNumber:             0,
		MaxGasAmount:               1000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1_700_000_000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "test"},
			Function: "f",
		},
	}
}

// simulateFakeHandler captures the simulate request and returns a fixed
// successful response. The captured query and body bytes are surfaced
// through pointers so tests can assert on them.
func simulateFakeHandler(t *testing.T, gotQuery *url.Values, gotBody *[]byte) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/transactions/simulate") {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			if gotBody != nil {
				*gotBody = body
			}
			if gotQuery != nil {
				q := r.URL.Query()
				*gotQuery = q
			}
			w.Header().Set("Content-Type", "application/json")
			// Simulate endpoint returns an array.
			_, _ = w.Write([]byte(`[{"success":true,"vm_status":"Executed","gas_used":"123","gas_unit_price":"100","changes":[],"events":[]}]`))
			return
		}
		http.NotFound(w, r)
	})
}

func TestSimulateTransaction_DefaultsEstimateGas(t *testing.T) {
	t.Parallel()

	var gotQuery url.Values
	client := newTestClient(t, simulateFakeHandler(t, &gotQuery, nil))

	signer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	res, err := client.SimulateTransaction(context.Background(), makeRawTxn(AccountOne), signer)
	require.NoError(t, err)
	assert.True(t, res.Success)
	// No opts → default estimation on, prioritized off.
	assert.Equal(t, "true", gotQuery.Get("estimate_gas_unit_price"))
	assert.Equal(t, "true", gotQuery.Get("estimate_max_gas_amount"))
	assert.Equal(t, "false", gotQuery.Get("estimate_prioritized_gas_unit_price"))
}

func TestSimulateTransaction_PrioritizedGasFlag(t *testing.T) {
	t.Parallel()

	var gotQuery url.Values
	client := newTestClient(t, simulateFakeHandler(t, &gotQuery, nil))

	signer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateTransaction(context.Background(), makeRawTxn(AccountOne), signer, WithPrioritizedGas())
	require.NoError(t, err)
	assert.Equal(t, "true", gotQuery.Get("estimate_prioritized_gas_unit_price"))
}

func TestSimulateMultiAgentTransaction_LengthMismatch(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, simulateFakeHandler(t, nil, nil))
	primary, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateMultiAgentTransaction(
		context.Background(),
		makeRawTxn(AccountOne),
		primary,
		[]Signer{},                   // 0 signers
		[]AccountAddress{AccountTwo}, // 1 address
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "length mismatch")
}

func TestSimulateFeePayerTransaction_LengthMismatch(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, simulateFakeHandler(t, nil, nil))
	primary, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	feePayer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateFeePayerTransaction(
		context.Background(),
		makeRawTxn(AccountOne),
		primary,
		[]Signer{primary, primary},
		[]AccountAddress{AccountTwo}, // length mismatch
		AccountThree,
		feePayer,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "length mismatch")
}

// TestSimulateTransaction_AuthVariantSingleSender pins the on-the-wire
// authenticator variant for a single-sender simulation. The Aptos node
// rejects a transaction whose authenticator variant doesn't match the
// outer transaction kind, so getting this wrong was the kind of bug we
// want a test to catch immediately.
//
// The body of the BCS request after the RawTransaction is the
// TransactionAuthenticator enum, ULEB128-encoded. For a SingleSender
// authenticator that's a single byte: 4.
func TestSimulateTransaction_AuthVariantSingleSender(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	client := newTestClient(t, simulateFakeHandler(t, nil, &gotBody))

	signer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateTransaction(context.Background(), makeRawTxn(AccountOne), signer)
	require.NoError(t, err)
	assertAuthenticatorVariant(t, gotBody, makeRawTxn(AccountOne), uint8(TransactionAuthenticatorVariantSingleSender))
}

func TestSimulateMultiAgentTransaction_AuthVariantAndSecondaries(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	client := newTestClient(t, simulateFakeHandler(t, nil, &gotBody))

	primary, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	sec1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	sec2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	res, err := client.SimulateMultiAgentTransaction(
		context.Background(),
		makeRawTxn(AccountOne),
		primary,
		[]Signer{sec1, sec2},
		[]AccountAddress{AccountTwo, AccountThree},
	)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assertAuthenticatorVariant(t, gotBody, makeRawTxn(AccountOne), uint8(TransactionAuthenticatorVariantMultiAgent))

	// Round-trip-deserialize the body to confirm the secondary
	// addresses arrived in the right slots. assertAuthenticatorVariant
	// only checks one byte; a regression that silently dropped the
	// secondaries would otherwise slip past.
	got := &SignedTransaction{}
	got.UnmarshalBCS(bcs.NewDeserializer(gotBody))
	ma, ok := got.Authenticator.(*MultiAgentAuthenticator)
	require.True(t, ok, "expected *MultiAgentAuthenticator, got %T", got.Authenticator)
	assert.Equal(
		t,
		[]AccountAddress{AccountTwo, AccountThree},
		ma.SecondarySignerAddresses,
		"secondary signer addresses must appear in the deserialized body",
	)
	assert.Len(t, ma.SecondarySigners, 2, "two secondary authenticators expected")
}

func TestSimulateFeePayerTransaction_AuthVariant(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	client := newTestClient(t, simulateFakeHandler(t, nil, &gotBody))

	primary, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	feePayer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateFeePayerTransaction(
		context.Background(),
		makeRawTxn(AccountOne),
		primary,
		nil,
		nil,
		AccountThree,
		feePayer,
	)
	require.NoError(t, err)
	assertAuthenticatorVariant(t, gotBody, makeRawTxn(AccountOne), uint8(TransactionAuthenticatorVariantFeePayer))
}

func TestSimulateFeePayerTransaction_WithSecondaries(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	client := newTestClient(t, simulateFakeHandler(t, nil, &gotBody))

	primary, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	sec, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	feePayer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateFeePayerTransaction(
		context.Background(),
		makeRawTxn(AccountOne),
		primary,
		[]Signer{sec},
		[]AccountAddress{AccountTwo},
		AccountThree,
		feePayer,
	)
	require.NoError(t, err)
	assertAuthenticatorVariant(t, gotBody, makeRawTxn(AccountOne), uint8(TransactionAuthenticatorVariantFeePayer))

	// Round-trip the body and assert the fee-payer address survived
	// the wire encoding — a regression that silently dropped it would
	// produce valid-looking bytes that the node would later reject.
	got := &SignedTransaction{}
	got.UnmarshalBCS(bcs.NewDeserializer(gotBody))
	fp, ok := got.Authenticator.(*FeePayerAuthenticator)
	require.True(t, ok, "expected *FeePayerAuthenticator, got %T", got.Authenticator)
	assert.Equal(t, []AccountAddress{AccountTwo}, fp.SecondarySignerAddresses)
	assert.Equal(t, AccountThree, fp.FeePayerAddress)
}

func TestSimulateSigned_EmptyResponseIsError(t *testing.T) {
	t.Parallel()

	// Simulate endpoint returns an empty array — the node never does
	// this in practice, but the code defensively errors rather than
	// dereferencing results[0]. Pin that behaviour.
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	signer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	_, err = client.SimulateTransaction(context.Background(), makeRawTxn(AccountOne), signer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no results")
}

// assertAuthenticatorVariant skips over the BCS-encoded RawTransaction at
// the head of the simulate request body and asserts the next byte (which
// the BCS encoding places at the start of the SignedTransaction's
// authenticator) is the expected variant.
func assertAuthenticatorVariant(t *testing.T, body []byte, txn *RawTransaction, wantVariant uint8) {
	t.Helper()
	txnBytes, err := bcs.Serialize(txn)
	require.NoError(t, err)
	require.Greater(t, len(body), len(txnBytes), "body shorter than serialized txn")
	require.Equal(t, txnBytes, body[:len(txnBytes)], "request body must start with the BCS-encoded raw txn")
	got := body[len(txnBytes)]
	assert.Equal(t, wantVariant, got, "expected authenticator variant %d, got %d", wantVariant, got)
}
