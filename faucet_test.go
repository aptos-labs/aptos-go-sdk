package aptos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaucetClient_Fund(t *testing.T) {
	t.Parallel()

	// Create a mock server that handles both faucet mint and node endpoints
	txnHash := "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			// Node info for NewClient
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chain_id": 4,
			})
		case "/mint":
			// Faucet mint endpoint
			assert.Equal(t, r.URL.Query().Get("address"), AccountOne.String())
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]string{txnHash})
		case "/transactions/wait_by_hash/" + txnHash:
			// Wait for transaction endpoint
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":                      "user_transaction",
				"hash":                      txnHash,
				"version":                   "1",
				"success":                   true,
				"sender":                    "0x1",
				"sequence_number":           "0",
				"max_gas_amount":            "100000",
				"gas_unit_price":            "100",
				"expiration_timestamp_secs": "9999999999",
				"gas_used":                  "100",
				"vm_status":                 "Executed successfully",
				"timestamp":                 "1000000",
				"accumulator_root_hash":     "0x0",
				"state_change_hash":         "0x0",
				"event_root_hash":           "0x0",
				"changes":                   []any{},
				"events":                    []any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	nodeClient, err := NewNodeClient(server.URL, 4)
	require.NoError(t, err)

	faucetClient, err := NewFaucetClient(nodeClient, server.URL)
	require.NoError(t, err)

	err = faucetClient.Fund(AccountOne, 100_000_000)
	require.NoError(t, err)
}

func TestFaucetClient_Fund_NilNodeClient(t *testing.T) {
	t.Parallel()
	faucetClient := &FaucetClient{nodeClient: nil}
	err := faucetClient.Fund(AccountOne, 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestFaucetClient_Fund_HttpError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer server.Close()

	nodeClient, err := NewNodeClient(server.URL, 4)
	require.NoError(t, err)

	faucetClient, err := NewFaucetClient(nodeClient, server.URL)
	require.NoError(t, err)

	err = faucetClient.Fund(AccountOne, 100)
	require.Error(t, err)
}
