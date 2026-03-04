package aptos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockClient creates a test client backed by an httptest.Server.
// The provided handler is called for every request except "/", which returns
// valid NodeInfo JSON so that NewClient's initial Info() call succeeds.
// The caller is responsible for calling server.Close() via the returned cleanup.
func newMockClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
			return
		}
		handler(w, r)
	}))
	client, err := NewClient(NetworkConfig{
		Name:    "mocknet",
		NodeUrl: server.URL,
		ChainId: 4,
	})
	require.NoError(t, err)
	return client, server
}

func TestNodeClient_Account(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.URL.Path, "/accounts/"))
		_ = json.NewEncoder(w).Encode(AccountInfo{
			SequenceNumberStr:    "42",
			AuthenticationKeyHex: "0xabcdef",
		})
	})
	defer server.Close()

	info, err := client.Account(AccountOne)
	require.NoError(t, err)
	seq, err := info.SequenceNumber()
	require.NoError(t, err)
	assert.Equal(t, uint64(42), seq)
}

func TestNodeClient_AccountResource(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/resource/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
			"data": map[string]any{"coin": map[string]any{"value": "1000"}},
		})
	})
	defer server.Close()

	data, err := client.AccountResource(AccountOne, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	require.NoError(t, err)
	assert.NotNil(t, data["data"])
}

func TestNodeClient_AccountResources(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, "/resources"))
		_ = json.NewEncoder(w).Encode([]AccountResourceInfo{
			{Type: "0x1::account::Account", Data: map[string]any{"sequence_number": "5"}},
			{Type: "0x1::coin::CoinStore", Data: map[string]any{}},
		})
	})
	defer server.Close()

	resources, err := client.AccountResources(AccountOne)
	require.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, "0x1::account::Account", resources[0].Type)
}

func TestNodeClient_AccountModule(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/module/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"bytecode": "0xdead",
			"abi": map[string]any{
				"address":           "0x1",
				"name":              "coin",
				"friends":           []any{},
				"exposed_functions": []any{},
				"structs":           []any{},
			},
		})
	})
	defer server.Close()

	module, err := client.AccountModule(AccountOne, "coin")
	require.NoError(t, err)
	assert.NotNil(t, module)
}

func TestNodeClient_BlockByHeight(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/blocks/by_height/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"block_hash":      "0xabc",
			"block_height":    "100",
			"block_timestamp": "1665609760857472",
			"first_version":   "200",
			"last_version":    "210",
		})
	})
	defer server.Close()

	block, err := client.BlockByHeight(100, false)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), block.BlockHeight)
	assert.Equal(t, uint64(200), block.FirstVersion)
	assert.Equal(t, uint64(210), block.LastVersion)
}

func TestNodeClient_BlockByVersion(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/blocks/by_version/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"block_hash":      "0xdef",
			"block_height":    "50",
			"block_timestamp": "1665609760857472",
			"first_version":   "100",
			"last_version":    "100",
		})
	})
	defer server.Close()

	block, err := client.BlockByVersion(100, false)
	require.NoError(t, err)
	assert.Equal(t, uint64(50), block.BlockHeight)
}

func TestNodeClient_TransactionByHash(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/transactions/by_hash/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":                      "user_transaction",
			"hash":                      "0xdeadbeef",
			"version":                   "999",
			"success":                   true,
			"sender":                    "0x1",
			"sequence_number":           "0",
			"max_gas_amount":            "100000",
			"gas_unit_price":            "100",
			"expiration_timestamp_secs": "9999999999",
			"gas_used":                  "42",
			"vm_status":                 "Executed successfully",
			"timestamp":                 "1000000",
			"accumulator_root_hash":     "0x0",
			"state_change_hash":         "0x0",
			"event_root_hash":           "0x0",
			"changes":                   []any{},
			"events":                    []any{},
		})
	})
	defer server.Close()

	txn, err := client.TransactionByHash("0xdeadbeef")
	require.NoError(t, err)
	assert.Equal(t, "0xdeadbeef", txn.Hash())
}

func TestNodeClient_TransactionByVersion(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/transactions/by_version/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":                  "state_checkpoint_transaction",
			"hash":                  "0xaabbccdd",
			"version":               "500",
			"success":               true,
			"accumulator_root_hash": "0x0",
			"state_change_hash":     "0x0",
			"event_root_hash":       "0x0",
			"changes":               []any{},
			"events":                []any{},
			"timestamp":             "1000000",
		})
	})
	defer server.Close()

	txn, err := client.TransactionByVersion(500)
	require.NoError(t, err)
	assert.Equal(t, uint64(500), txn.Version())
}

func TestNodeClient_Transactions(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions", r.URL.Path)
		startStr := r.URL.Query().Get("start")
		limitStr := r.URL.Query().Get("limit")
		start := uint64(0)
		limit := uint64(5)
		if startStr != "" {
			start, _ = strconv.ParseUint(startStr, 10, 64)
		}
		if limitStr != "" {
			limit, _ = strconv.ParseUint(limitStr, 10, 64)
		}

		txns := make([]map[string]any, 0, limit)
		for i := range limit {
			txns = append(txns, map[string]any{
				"type":                  "state_checkpoint_transaction",
				"hash":                  "0x" + strconv.FormatUint(start+i, 16),
				"version":               strconv.FormatUint(start+i, 10),
				"success":               true,
				"accumulator_root_hash": "0x0",
				"state_change_hash":     "0x0",
				"event_root_hash":       "0x0",
				"changes":               []any{},
				"events":                []any{},
				"timestamp":             "1000000",
				"state_checkpoint_hash": "0x0",
			})
		}
		_ = json.NewEncoder(w).Encode(txns)
	})
	defer server.Close()

	start := uint64(10)
	limit := uint64(5)
	txns, err := client.Transactions(&start, &limit)
	require.NoError(t, err)
	assert.Len(t, txns, 5)
}

func TestNodeClient_AccountTransactions(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/transactions")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"type":                      "user_transaction",
				"hash":                      "0xaaa",
				"version":                   "100",
				"success":                   true,
				"sender":                    "0x1",
				"sequence_number":           "0",
				"max_gas_amount":            "100000",
				"gas_unit_price":            "100",
				"expiration_timestamp_secs": "9999999999",
				"gas_used":                  "42",
				"vm_status":                 "Executed successfully",
				"timestamp":                 "1000000",
				"accumulator_root_hash":     "0x0",
				"state_change_hash":         "0x0",
				"event_root_hash":           "0x0",
				"changes":                   []any{},
				"events":                    []any{},
			},
		})
	})
	defer server.Close()

	start := uint64(0)
	limit := uint64(1)
	txns, err := client.AccountTransactions(AccountOne, &start, &limit)
	require.NoError(t, err)
	assert.Len(t, txns, 1)
}

func TestNodeClient_SubmitTransaction(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/transactions" && r.Method == http.MethodPost {
			contentType := r.Header.Get("Content-Type")
			assert.Equal(t, ContentTypeAptosSignedTxnBcs, contentType)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"hash":                      "0xsubmitted",
				"sender":                    "0x1",
				"sequence_number":           "0",
				"max_gas_amount":            "100000",
				"gas_unit_price":            "100",
				"expiration_timestamp_secs": "9999999999",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := &RawTransaction{
		Sender:                     sender.Address,
		SequenceNumber:             0,
		Payload:                    TransactionPayload{Payload: &EntryFunction{Module: ModuleId{Address: AccountOne, Name: "aptos_account"}, Function: "transfer", ArgTypes: []TypeTag{}, Args: [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}

	signedTxn, err := rawTxn.SignedTransaction(sender)
	require.NoError(t, err)

	resp, err := client.SubmitTransaction(signedTxn)
	require.NoError(t, err)
	assert.Equal(t, "0xsubmitted", resp.Hash)
}

func TestNodeClient_SimulateTransaction(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/accounts/"+AccountOne.String():
			_ = json.NewEncoder(w).Encode(AccountInfo{SequenceNumberStr: "0", AuthenticationKeyHex: "0x00"})
		case r.URL.Path == "/estimate_gas_price":
			_ = json.NewEncoder(w).Encode(EstimateGasInfo{GasEstimate: 100})
		case strings.HasPrefix(r.URL.Path, "/transactions/simulate"):
			contentType := r.Header.Get("Content-Type")
			assert.Equal(t, ContentTypeAptosSignedTxnBcs, contentType)
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"version":                   "1",
					"hash":                      "0xsim",
					"success":                   true,
					"sender":                    "0x1",
					"sequence_number":           "0",
					"max_gas_amount":            "100000",
					"gas_unit_price":            "100",
					"expiration_timestamp_secs": "9999999999",
					"gas_used":                  "50",
					"vm_status":                 "Executed successfully",
					"timestamp":                 "1000000",
					"accumulator_root_hash":     "0x0",
					"state_change_hash":         "0x0",
					"event_root_hash":           "0x0",
					"changes":                   []any{},
					"events":                    []any{},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := &RawTransaction{
		Sender:                     sender.Address,
		SequenceNumber:             0,
		Payload:                    TransactionPayload{Payload: &EntryFunction{Module: ModuleId{Address: AccountOne, Name: "aptos_account"}, Function: "transfer", ArgTypes: []TypeTag{}, Args: [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}

	results, err := client.SimulateTransaction(rawTxn, sender)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.True(t, results[0].Success)
}

func TestNodeClient_BuildTransaction(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/accounts/"):
			_ = json.NewEncoder(w).Encode(AccountInfo{SequenceNumberStr: "10", AuthenticationKeyHex: "0x00"})
		case r.URL.Path == "/estimate_gas_price":
			_ = json.NewEncoder(w).Encode(EstimateGasInfo{GasEstimate: 150})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	payload := TransactionPayload{Payload: &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{AccountTwo[:], {0xe8, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}}

	rawTxn, err := client.BuildTransaction(AccountOne, payload)
	require.NoError(t, err)
	assert.Equal(t, AccountOne, rawTxn.Sender)
	assert.Equal(t, uint64(10), rawTxn.SequenceNumber)
	assert.Equal(t, uint64(150), rawTxn.GasUnitPrice)
	assert.Equal(t, uint8(4), rawTxn.ChainId)
}

func TestNodeClient_BuildSignAndSubmitTransaction(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/transactions" && r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"hash":                      "0xsubmitted",
				"sender":                    "0x1",
				"sequence_number":           "0",
				"max_gas_amount":            "100000",
				"gas_unit_price":            "100",
				"expiration_timestamp_secs": "9999999999",
			})
		case strings.HasPrefix(r.URL.Path, "/accounts/"):
			_ = json.NewEncoder(w).Encode(AccountInfo{SequenceNumberStr: "5", AuthenticationKeyHex: "0x00"})
		case r.URL.Path == "/estimate_gas_price":
			_ = json.NewEncoder(w).Encode(EstimateGasInfo{GasEstimate: 100})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	payload := TransactionPayload{Payload: &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}}

	resp, err := client.BuildSignAndSubmitTransaction(sender, payload)
	require.NoError(t, err)
	assert.Equal(t, "0xsubmitted", resp.Hash)
}

func TestNodeClient_View(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/view" && r.Method == http.MethodPost {
			contentType := r.Header.Get("Content-Type")
			assert.Equal(t, ContentTypeAptosViewFunctionBcs, contentType)
			_ = json.NewEncoder(w).Encode([]any{"12345"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	payload := &ViewPayload{
		Module:   ModuleId{Address: AccountOne, Name: "coin"},
		Function: "balance",
		ArgTypes: []TypeTag{AptosCoinTypeTag},
		Args:     [][]byte{AccountOne[:]},
	}
	result, err := client.View(payload)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "12345", result[0])
}

func TestNodeClient_EstimateGasPrice(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/estimate_gas_price", r.URL.Path)
		_ = json.NewEncoder(w).Encode(EstimateGasInfo{
			DeprioritizedGasEstimate: 50,
			GasEstimate:              100,
			PrioritizedGasEstimate:   200,
		})
	})
	defer server.Close()

	info, err := client.EstimateGasPrice()
	require.NoError(t, err)
	assert.Equal(t, uint64(100), info.GasEstimate)
	assert.Equal(t, uint64(50), info.DeprioritizedGasEstimate)
	assert.Equal(t, uint64(200), info.PrioritizedGasEstimate)
}

func TestNodeClient_AccountAPTBalance(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/view" && r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode([]any{"500000000"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	balance, err := client.AccountAPTBalance(AccountOne)
	require.NoError(t, err)
	assert.Equal(t, uint64(500000000), balance)
}

func TestNodeClient_NodeAPIHealthCheck(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/-/healthy", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "aptos-node:ok",
		})
	})
	defer server.Close()

	resp, err := client.NodeAPIHealthCheck()
	require.NoError(t, err)
	assert.Equal(t, "aptos-node:ok", resp.Message)
}

func TestNodeClient_SetTimeout(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client.SetTimeout(5 * time.Second)
	// Verify we can still make requests
	_, err := client.Info()
	require.NoError(t, err)
}

func TestNodeClient_SetHeader_RemoveHeader(t *testing.T) {
	t.Parallel()
	var receivedAuth string
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(EstimateGasInfo{GasEstimate: 100})
	})
	defer server.Close()

	client.SetHeader("Authorization", "Bearer test-token")
	_, err := client.EstimateGasPrice()
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token", receivedAuth)

	client.RemoveHeader("Authorization")
	_, err = client.EstimateGasPrice()
	require.NoError(t, err)
	assert.Empty(t, receivedAuth)
}

func TestNodeClient_GetChainId(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	chainId, err := client.GetChainId()
	require.NoError(t, err)
	assert.Equal(t, uint8(4), chainId)
}

func TestNodeClient_ErrorResponse(t *testing.T) {
	t.Parallel()

	t.Run("400 error", func(t *testing.T) {
		t.Parallel()
		client, server := newMockClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad request"}`))
		})
		defer server.Close()

		_, err := client.Account(AccountOne)
		require.Error(t, err)
	})

	t.Run("404 error", func(t *testing.T) {
		t.Parallel()
		client, server := newMockClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		})
		defer server.Close()

		_, err := client.TransactionByHash("0xbad")
		require.Error(t, err)
	})

	t.Run("500 error", func(t *testing.T) {
		t.Parallel()
		client, server := newMockClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"internal error"}`))
		})
		defer server.Close()

		_, err := client.EstimateGasPrice()
		require.Error(t, err)
	})
}

func TestNodeClient_Info(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(NodeInfo{
			ChainId:          4,
			EpochStr:         "100",
			LedgerVersionStr: "999",
			BlockHeightStr:   "500",
			NodeRole:         "full_node",
		})
	}))
	defer server.Close()

	client, err := NewClient(NetworkConfig{Name: "mocknet", NodeUrl: server.URL})
	require.NoError(t, err)

	info, err := client.Info()
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainId)
	assert.Equal(t, uint64(100), info.Epoch())
	assert.Equal(t, uint64(999), info.LedgerVersion())
	assert.Equal(t, uint64(500), info.BlockHeight())
	assert.Equal(t, "full_node", info.NodeRole)
}

func TestNodeClient_BatchSubmitTransaction(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/transactions/batch" && r.Method == http.MethodPost {
			contentType := r.Header.Get("Content-Type")
			assert.Equal(t, ContentTypeAptosSignedTxnBcs, contentType)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"transaction_failures": []any{},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := &RawTransaction{
		Sender:                     sender.Address,
		SequenceNumber:             0,
		Payload:                    TransactionPayload{Payload: &EntryFunction{Module: ModuleId{Address: AccountOne, Name: "aptos_account"}, Function: "transfer", ArgTypes: []TypeTag{}, Args: [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}

	signedTxn, err := rawTxn.SignedTransaction(sender)
	require.NoError(t, err)

	resp, err := client.BatchSubmitTransaction([]*SignedTransaction{signedTxn})
	require.NoError(t, err)
	assert.Empty(t, resp.TransactionFailures)
}

func TestPollForTransaction(t *testing.T) {
	t.Parallel()
	// this doesn't need to actually have an aptos-node!
	// API error on every GET is fine, poll for a few milliseconds then return error
	client, err := NewClient(LocalnetConfig)
	require.NoError(t, err)

	start := time.Now()
	err = client.PollForTransactions([]string{"alice", "bob"}, PollTimeout(10*time.Millisecond), PollPeriod(2*time.Millisecond))
	dt := time.Since(start)

	assert.GreaterOrEqual(t, dt, 9*time.Millisecond)
	// Use a generous upper bound to avoid flaky failures due to system load or CI variability
	assert.Less(t, dt, 500*time.Millisecond)
	require.Error(t, err)
}

func TestEventsByHandle(t *testing.T) {
	t.Parallel()
	createMockServer := func(t *testing.T) *httptest.Server {
		t.Helper()
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				// handle initial request from client
				w.WriteHeader(http.StatusOK)
				return
			}

			assert.Equal(t, "/accounts/0x0/events/0x2/transfer", r.URL.Path)

			start := r.URL.Query().Get("start")
			limit := r.URL.Query().Get("limit")

			var startInt uint64
			var limitInt uint64

			if start != "" {
				startInt, _ = strconv.ParseUint(start, 10, 64)
			}
			if limit != "" {
				limitInt, _ = strconv.ParseUint(limit, 10, 64)
			} else {
				limitInt = 100
			}

			events := make([]map[string]interface{}, 0, limitInt)
			for i := range limitInt {
				events = append(events, map[string]interface{}{
					"type": "0x1::coin::TransferEvent",
					"guid": map[string]interface{}{
						"creation_number": "1",
						"account_address": AccountZero.String(),
					},
					"sequence_number": strconv.FormatUint(startInt+i, 10),
					"data": map[string]interface{}{
						"amount": strconv.FormatUint((startInt+i)*100, 10),
					},
				})
			}

			err := json.NewEncoder(w).Encode(events)
			if err != nil {
				t.Error(err)
				return
			}
		}))
	}

	t.Run("pagination with concurrent fetching", func(t *testing.T) {
		t.Parallel()
		mockServer := createMockServer(t)
		defer mockServer.Close()

		client, err := NewClient(NetworkConfig{
			Name:    "mocknet",
			NodeUrl: mockServer.URL,
		})
		require.NoError(t, err)

		start := uint64(0)
		limit := uint64(150)
		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			&start,
			&limit,
		)

		require.NoError(t, err)
		assert.Len(t, events, 150)
	})

	t.Run("default page size when limit not provided", func(t *testing.T) {
		t.Parallel()
		mockServer := createMockServer(t)
		defer mockServer.Close()

		client, err := NewClient(NetworkConfig{
			Name:    "mocknet",
			NodeUrl: mockServer.URL,
		})
		require.NoError(t, err)

		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			nil,
			nil,
		)

		require.NoError(t, err)
		assert.Len(t, events, 100)
		assert.Equal(t, uint64(99), events[99].SequenceNumber)
	})

	t.Run("single page fetch", func(t *testing.T) {
		t.Parallel()
		mockServer := createMockServer(t)
		defer mockServer.Close()

		client, err := NewClient(NetworkConfig{
			Name:    "mocknet",
			NodeUrl: mockServer.URL,
		})
		require.NoError(t, err)

		start := uint64(50)
		limit := uint64(5)
		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			&start,
			&limit,
		)

		require.NoError(t, err)
		assert.Len(t, events, 5)
		assert.Equal(t, uint64(50), events[0].SequenceNumber)
		assert.Equal(t, uint64(54), events[4].SequenceNumber)
	})
}

func TestEventsByCreationNumber(t *testing.T) {
	t.Parallel()
	createMockServer := func(t *testing.T) *httptest.Server {
		t.Helper()
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				// handle initial request from client
				w.WriteHeader(http.StatusOK)
				return
			}

			assert.Equal(t, "/accounts/0x0/events/123", r.URL.Path)

			start := r.URL.Query().Get("start")
			limit := r.URL.Query().Get("limit")

			var startInt uint64
			var limitInt uint64

			if start != "" {
				startInt, _ = strconv.ParseUint(start, 10, 64)
			}
			if limit != "" {
				limitInt, _ = strconv.ParseUint(limit, 10, 64)
			} else {
				limitInt = 100
			}

			events := make([]map[string]interface{}, 0, limitInt)
			for i := range limitInt {
				events = append(events, map[string]interface{}{
					"type": "0x1::coin::TransferEvent",
					"guid": map[string]interface{}{
						"creation_number": "123",
						"account_address": AccountZero.String(),
					},
					"sequence_number": strconv.FormatUint(startInt+i, 10),
					"data": map[string]interface{}{
						"amount": strconv.FormatUint((startInt+i)*100, 10),
					},
				})
			}

			err := json.NewEncoder(w).Encode(events)
			if err != nil {
				t.Error(err)
				return
			}
		}))
	}

	t.Run("pagination with concurrent fetching", func(t *testing.T) {
		t.Parallel()
		mockServer := createMockServer(t)
		defer mockServer.Close()

		client, err := NewClient(NetworkConfig{
			Name:    "mocknet",
			NodeUrl: mockServer.URL,
		})
		require.NoError(t, err)

		start := uint64(0)
		limit := uint64(150)
		events, err := client.EventsByCreationNumber(
			AccountZero,
			"123",
			&start,
			&limit,
		)

		require.NoError(t, err)
		assert.Len(t, events, 150)
	})

	t.Run("default page size when limit not provided", func(t *testing.T) {
		t.Parallel()
		mockServer := createMockServer(t)
		defer mockServer.Close()

		client, err := NewClient(NetworkConfig{
			Name:    "mocknet",
			NodeUrl: mockServer.URL,
		})
		require.NoError(t, err)

		events, err := client.EventsByCreationNumber(
			AccountZero,
			"123",
			nil,
			nil,
		)

		require.NoError(t, err)
		assert.Len(t, events, 100)
		assert.Equal(t, uint64(99), events[99].SequenceNumber)
	})

	t.Run("single page fetch", func(t *testing.T) {
		t.Parallel()
		mockServer := createMockServer(t)
		defer mockServer.Close()

		client, err := NewClient(NetworkConfig{
			Name:    "mocknet",
			NodeUrl: mockServer.URL,
		})
		require.NoError(t, err)

		start := uint64(50)
		limit := uint64(5)
		events, err := client.EventsByCreationNumber(
			AccountZero,
			"123",
			&start,
			&limit,
		)

		require.NoError(t, err)
		assert.Len(t, events, 5)
		assert.Equal(t, uint64(50), events[0].SequenceNumber)
		assert.Equal(t, uint64(54), events[4].SequenceNumber)
	})
}

func TestNodeClient_WaitTransactionByHash(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/transactions/wait_by_hash/")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":                      "user_transaction",
			"hash":                      "0xdeadbeef",
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
	})
	defer server.Close()

	txn, err := client.WaitTransactionByHash("0xdeadbeef")
	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "user_transaction", string(txn.Type))
}

func TestNodeClient_NodeHealthCheck(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/-/healthy", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "aptos-node:ok",
		})
	})
	defer server.Close()

	resp, err := client.NodeAPIHealthCheck()
	require.NoError(t, err)
	assert.Equal(t, "aptos-node:ok", resp.Message)
}

func TestNodeClient_NodeHealthCheck_Deprecated(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
			return
		}
		assert.Equal(t, "/-/healthy", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "aptos-node:ok",
		})
	}))
	defer server.Close()

	nodeClient, err := NewNodeClient(server.URL, 4)
	require.NoError(t, err)

	resp, err := nodeClient.NodeHealthCheck()
	require.NoError(t, err)
	assert.Equal(t, "aptos-node:ok", resp.Message)
}

func TestNodeClient_GetBCS(t *testing.T) {
	t.Parallel()
	expectedBytes := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
			return
		}
		w.Header().Set("Content-Type", "application/x-bcs")
		_, _ = w.Write(expectedBytes)
	}))
	defer server.Close()

	nodeClient, err := NewNodeClient(server.URL, 4)
	require.NoError(t, err)

	result, err := nodeClient.GetBCS(server.URL + "/some-path")
	require.NoError(t, err)
	assert.Equal(t, expectedBytes, result)
}

func TestNodeClient_EventsByHandle_Simple(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "events/")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"guid": map[string]any{
				"creation_number": "1",
				"account_address": "0x1",
			},
			"sequence_number": "0",
			"type":            "0x1::coin::DepositEvent",
			"data":            map[string]any{"amount": "1000"},
		}})
	})
	defer server.Close()

	events, err := client.EventsByHandle(AccountOne, "0x1::coin::CoinStore", "deposit_events", nil, nil)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "0x1::coin::DepositEvent", events[0].Type)
	assert.Equal(t, uint64(0), events[0].SequenceNumber)
}

func TestNodeClient_EventsByCreationNumber_Simple(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "events/")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"guid": map[string]any{
				"creation_number": "1",
				"account_address": "0x1",
			},
			"sequence_number": "0",
			"type":            "0x1::coin::DepositEvent",
			"data":            map[string]any{"amount": "1000"},
		}})
	})
	defer server.Close()

	events, err := client.EventsByCreationNumber(AccountOne, "1", nil, nil)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "0x1::coin::DepositEvent", events[0].Type)
	assert.Equal(t, uint64(0), events[0].SequenceNumber)
}

func TestClient_PollForTransaction_Mock(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/transactions/wait_by_hash/") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":                      "user_transaction",
				"hash":                      "0xdeadbeef",
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
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	userTxn, err := client.PollForTransaction("0xdeadbeef")
	require.NoError(t, err)
	require.NotNil(t, userTxn)
	assert.True(t, userTxn.Success)
	assert.Equal(t, uint64(1), userTxn.Version)
}

func TestClient_WaitForTransaction_Mock(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/transactions/wait_by_hash/") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":                      "user_transaction",
				"hash":                      "0xdeadbeef",
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
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	userTxn, err := client.WaitForTransaction("0xdeadbeef")
	require.NoError(t, err)
	require.NotNil(t, userTxn)
	assert.True(t, userTxn.Success)
	assert.Equal(t, uint64(1), userTxn.Version)
}

func TestNodeClient_BuildTransactionMultiAgent(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/accounts/"):
			_ = json.NewEncoder(w).Encode(AccountInfo{SequenceNumberStr: "5", AuthenticationKeyHex: "0x00"})
		case r.URL.Path == "/estimate_gas_price":
			_ = json.NewEncoder(w).Encode(EstimateGasInfo{GasEstimate: 100})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	payload := TransactionPayload{Payload: &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}}

	feePayer := AccountTwo
	rawTxn, err := client.BuildTransactionMultiAgent(AccountOne, payload, FeePayer(&feePayer))
	require.NoError(t, err)
	assert.Equal(t, MultiAgentWithFeePayerRawTransactionWithDataVariant, rawTxn.Variant)
}

func TestNodeClient_BuildTransactionMultiAgent_NoFeePayer(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/accounts/"):
			_ = json.NewEncoder(w).Encode(AccountInfo{SequenceNumberStr: "5", AuthenticationKeyHex: "0x00"})
		case r.URL.Path == "/estimate_gas_price":
			_ = json.NewEncoder(w).Encode(EstimateGasInfo{GasEstimate: 100})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	payload := TransactionPayload{Payload: &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}}

	rawTxn, err := client.BuildTransactionMultiAgent(AccountOne, payload)
	require.NoError(t, err)
	assert.Equal(t, MultiAgentRawTransactionWithDataVariant, rawTxn.Variant)
}

func TestNodeClient_SimulateTransactionMultiAgent(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/transactions/simulate"):
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"version":                   "1",
				"hash":                      "0xsim",
				"success":                   true,
				"sender":                    "0x1",
				"sequence_number":           "0",
				"max_gas_amount":            "100000",
				"gas_unit_price":            "100",
				"expiration_timestamp_secs": "9999999999",
				"gas_used":                  "50",
				"vm_status":                 "Executed successfully",
				"timestamp":                 "1000000",
				"accumulator_root_hash":     "0x0",
				"state_change_hash":         "0x0",
				"event_root_hash":           "0x0",
				"changes":                   []any{},
				"events":                    []any{},
			}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn := &RawTransaction{
		Sender:                     sender.Address,
		SequenceNumber:             0,
		Payload:                    TransactionPayload{Payload: &EntryFunction{Module: ModuleId{Address: AccountOne, Name: "aptos_account"}, Function: "transfer", ArgTypes: []TypeTag{}, Args: [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}

	rawTxnWithData := &RawTransactionWithData{
		Variant: MultiAgentWithFeePayerRawTransactionWithDataVariant,
		Inner: &MultiAgentWithFeePayerRawTransactionWithData{
			RawTxn:           rawTxn,
			FeePayer:         &AccountAddress{},
			SecondarySigners: []AccountAddress{},
		},
	}

	feePayer := AccountTwo
	results, err := client.SimulateTransactionMultiAgent(rawTxnWithData, sender, FeePayer(&feePayer))
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestNodeClient_NodeAPIHealthCheck_WithDuration(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/-/healthy", r.URL.Path)
		assert.Equal(t, "30", r.URL.Query().Get("duration_secs"))
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "aptos-node:ok"})
	})
	defer server.Close()

	resp, err := client.NodeAPIHealthCheck(30)
	require.NoError(t, err)
	assert.Equal(t, "aptos-node:ok", resp.Message)
}

func TestNodeClient_AccountResource_WithLedgerVersion(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/resource/")
		assert.Equal(t, "42", r.URL.Query().Get("ledger_version"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
			"data": map[string]any{"coin": map[string]any{"value": "1000"}},
		})
	})
	defer server.Close()

	data, err := client.AccountResource(AccountOne, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", 42)
	require.NoError(t, err)
	assert.NotNil(t, data["data"])
}

func TestNodeClient_Account_WithLedgerVersion(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.URL.Path, "/accounts/"))
		assert.Equal(t, "99", r.URL.Query().Get("ledger_version"))
		_ = json.NewEncoder(w).Encode(AccountInfo{
			SequenceNumberStr:    "10",
			AuthenticationKeyHex: "0xabcdef",
		})
	})
	defer server.Close()

	info, err := client.Account(AccountOne, 99)
	require.NoError(t, err)
	seq, err := info.SequenceNumber()
	require.NoError(t, err)
	assert.Equal(t, uint64(10), seq)
}
