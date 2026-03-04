package aptos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a nodeClient pointing at a test server.
func newTestClient(t *testing.T, handler http.Handler) (*nodeClient, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := newNodeClient(&ClientConfig{
		network: NetworkConfig{
			NodeURL:   server.URL,
			FaucetURL: server.URL,
			ChainID:   4,
		},
		timeout: 0, // no timeout for tests
	})
	require.NoError(t, err)
	return client, server
}

func jsonHandler(statusCode int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(body)
	}
}

func TestNodeClient_HealthCheck(t *testing.T) {
	t.Parallel()
	client, _ := newTestClient(t, jsonHandler(http.StatusOK, HealthCheckResponse{Message: "aptos-node:ok"}))
	resp, err := client.HealthCheck(context.Background())
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "ok")
}

func TestNodeClient_HealthCheck_WithDuration(t *testing.T) {
	t.Parallel()
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "duration_secs=30")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthCheckResponse{Message: "aptos-node:ok"})
	}))
	resp, err := client.HealthCheck(context.Background(), 30)
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "ok")
}

func TestNodeClient_AccountModule(t *testing.T) {
	t.Parallel()
	module := ModuleBytecode{
		Bytecode: "0xdeadbeef",
		ABI:      &ModuleABI{Address: AccountOne.String(), Name: "coin"},
	}
	client, _ := newTestClient(t, jsonHandler(http.StatusOK, module))

	result, err := client.AccountModule(context.Background(), AccountOne, "coin")
	require.NoError(t, err)
	assert.Equal(t, "0xdeadbeef", result.Bytecode)
	assert.Equal(t, "coin", result.ABI.Name)
}

func TestNodeClient_AccountTransactions(t *testing.T) {
	t.Parallel()
	txns := []*Transaction{{Hash: "0xabc", Type: "user_transaction"}}
	client, _ := newTestClient(t, jsonHandler(http.StatusOK, txns))

	result, err := client.AccountTransactions(context.Background(), AccountOne, nil, nil)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "0xabc", result[0].Hash)
}

func TestNodeClient_AccountTransactions_WithParams(t *testing.T) {
	t.Parallel()
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "start=5")
		assert.Contains(t, r.URL.RawQuery, "limit=10")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Transaction{})
	}))

	start := uint64(5)
	limit := uint64(10)
	_, err := client.AccountTransactions(context.Background(), AccountOne, &start, &limit)
	require.NoError(t, err)
}

func TestNodeClient_EventsByCreationNumber(t *testing.T) {
	t.Parallel()
	events := []Event{{Type: "0x1::coin::DepositEvent", SequenceNumber: 0}}
	client, _ := newTestClient(t, jsonHandler(http.StatusOK, events))

	result, err := client.EventsByCreationNumber(context.Background(), AccountOne, 0, nil, nil)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestNodeClient_SimulateTransaction(t *testing.T) {
	t.Parallel()
	simResult := []*SimulationResult{{Success: true, VMStatus: "Executed successfully", GasUsed: 500}}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "transactions/simulate")
		assert.Equal(t, "application/x.aptos.signed_transaction+bcs", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(simResult)
	}))

	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
		},
	}

	result, err := client.SimulateTransaction(context.Background(), txn, key)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, uint64(500), result.GasUsed)
}

func TestNodeClient_SubmitTransaction(t *testing.T) {
	t.Parallel()
	submitResult := SubmitResult{Hash: "0xsubmitted"}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions", r.URL.Path)
		assert.Equal(t, "application/x.aptos.signed_transaction+bcs", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(submitResult)
	}))

	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
		},
	}

	signed, err := SignTransaction(key, txn)
	require.NoError(t, err)

	result, err := client.SubmitTransaction(context.Background(), signed)
	require.NoError(t, err)
	assert.Equal(t, "0xsubmitted", result.Hash)
}

func TestNodeClient_WaitForTransaction_Immediate(t *testing.T) {
	t.Parallel()
	txn := Transaction{Hash: "0xhash", Type: "user_transaction", Success: true}
	client, _ := newTestClient(t, jsonHandler(http.StatusOK, txn))

	result, err := client.WaitForTransaction(context.Background(), "0xhash")
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestNodeClient_Fund(t *testing.T) {
	t.Parallel()
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/mint")
		assert.Contains(t, r.URL.RawQuery, "address=")
		assert.Contains(t, r.URL.RawQuery, "amount=100000000")
		w.WriteHeader(http.StatusOK)
	}))

	err := client.Fund(context.Background(), AccountOne, 100000000)
	require.NoError(t, err)
}

func TestNodeClient_Fund_Error(t *testing.T) {
	t.Parallel()
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))

	err := client.Fund(context.Background(), AccountOne, 100)
	assert.Error(t, err)
}

func TestNodeClient_Fund_NoFaucet(t *testing.T) {
	t.Parallel()
	// Create client with no faucet URL
	client, err := newNodeClient(&ClientConfig{
		network: NetworkConfig{
			NodeURL: "http://localhost:8080",
			ChainID: 4,
		},
	})
	require.NoError(t, err)

	err = client.Fund(context.Background(), AccountOne, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "faucet not available")
}

func TestNodeClient_BatchSubmitTransaction(t *testing.T) {
	t.Parallel()
	batchResult := BatchSubmitResult{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "transactions/batch")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(batchResult)
	}))

	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
		},
	}

	signed, err := SignTransaction(key, txn)
	require.NoError(t, err)

	result, err := client.BatchSubmitTransaction(context.Background(), []*SignedTransaction{signed})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestNodeClient_PostBCS(t *testing.T) {
	t.Parallel()
	// Test that postBCS correctly sends BCS-encoded data
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x.aptos.signed_transaction+bcs", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmitResult{Hash: "0xtest"})
	}))

	var result SubmitResult
	err := client.postBCS(context.Background(), "transactions", []byte{0x01, 0x02}, &result)
	require.NoError(t, err)
	assert.Equal(t, "0xtest", result.Hash)
}

func TestNodeClient_APIError(t *testing.T) {
	t.Parallel()
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message":    "Account not found",
			"error_code": "account_not_found",
		})
	}))

	_, err := client.Account(context.Background(), AccountOne)
	require.Error(t, err)

	var apiErr *APIError
	assert.True(t, isAPIError(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestIsAPIError_NonAPIError(t *testing.T) {
	t.Parallel()
	var apiErr *APIError
	assert.False(t, isAPIError(assert.AnError, &apiErr))
}
