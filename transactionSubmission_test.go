package aptos

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPoolConfig_getBufferSizes_Default(t *testing.T) {
	t.Parallel()
	cfg := WorkerPoolConfig{}
	build, submission := cfg.getBufferSizes()
	assert.Equal(t, uint32(20), build)
	assert.Equal(t, uint32(20), submission)
}

func TestWorkerPoolConfig_getBufferSizes_CustomWorkers(t *testing.T) {
	t.Parallel()
	cfg := WorkerPoolConfig{NumWorkers: 10}
	build, submission := cfg.getBufferSizes()
	assert.Equal(t, uint32(10), build)
	assert.Equal(t, uint32(10), submission)
}

func TestWorkerPoolConfig_getBufferSizes_CustomBuffers(t *testing.T) {
	t.Parallel()
	cfg := WorkerPoolConfig{
		NumWorkers:          10,
		BuildResponseBuffer: 5,
		SubmissionBuffer:    15,
	}
	build, submission := cfg.getBufferSizes()
	assert.Equal(t, uint32(5), build)
	assert.Equal(t, uint32(15), submission)
}

func TestWorkerPoolConfig_getBufferSizes_OnlyBuildBuffer(t *testing.T) {
	t.Parallel()
	cfg := WorkerPoolConfig{BuildResponseBuffer: 30}
	build, submission := cfg.getBufferSizes()
	assert.Equal(t, uint32(30), build)
	assert.Equal(t, uint32(20), submission)
}

func TestWorkerPoolConfig_getBufferSizes_OnlySubmissionBuffer(t *testing.T) {
	t.Parallel()
	cfg := WorkerPoolConfig{SubmissionBuffer: 25}
	build, submission := cfg.getBufferSizes()
	assert.Equal(t, uint32(20), build)
	assert.Equal(t, uint32(25), submission)
}

func TestSubmitTransactions(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
			return
		}
		if r.URL.Path == "/transactions" && r.Method == http.MethodPost {
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
	}))
	defer server.Close()

	nodeClient, err := NewNodeClient(server.URL, 4)
	require.NoError(t, err)

	// Create a signed transaction
	sender, err := NewEd25519Account()
	require.NoError(t, err)
	rawTxn := &RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: 0,
		Payload: TransactionPayload{Payload: &EntryFunction{
			Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args:     [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}
	signedTxn, err := rawTxn.SignedTransaction(sender)
	require.NoError(t, err)

	requests := make(chan TransactionSubmissionRequest, 1)
	responses := make(chan TransactionSubmissionResponse, 1)

	go nodeClient.SubmitTransactions(requests, responses)
	requests <- TransactionSubmissionRequest{Id: 1, SignedTxn: signedTxn}
	close(requests)

	resp := <-responses
	require.NoError(t, resp.Err)
	assert.NotNil(t, resp.Response)
}

func TestBatchSubmitTransactions(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
			return
		}
		if r.URL.Path == "/transactions/batch" && r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"transaction_failures": []any{},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	nodeClient, err := NewNodeClient(server.URL, 4)
	require.NoError(t, err)

	sender, err := NewEd25519Account()
	require.NoError(t, err)
	rawTxn := &RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: 0,
		Payload: TransactionPayload{Payload: &EntryFunction{
			Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args:     [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}
	signedTxn, err := rawTxn.SignedTransaction(sender)
	require.NoError(t, err)

	requests := make(chan TransactionSubmissionRequest, 1)
	responses := make(chan TransactionSubmissionResponse, 1)

	go nodeClient.BatchSubmitTransactions(requests, responses)
	requests <- TransactionSubmissionRequest{Id: 1, SignedTxn: signedTxn}
	close(requests)

	resp := <-responses
	require.NoError(t, resp.Err)
	// BatchSubmitTransactions returns nil response on success (no individual responses)
	assert.Nil(t, resp.Response)
}

func TestStartSigningWorkers(t *testing.T) {
	t.Parallel()

	submissionRequests := make(chan TransactionSubmissionRequest, 10)
	responses := make(chan TransactionSubmissionResponse, 10)
	var signingWg sync.WaitGroup
	var transactionWg sync.WaitGroup

	// Create a sign function
	sender, err := NewEd25519Account()
	require.NoError(t, err)

	sign := func(rawTxn RawTransactionImpl) (*SignedTransaction, error) {
		rt, ok := rawTxn.(*RawTransaction)
		if !ok {
			return nil, errors.New("unexpected type")
		}
		return rt.SignedTransaction(sender)
	}

	transactionsToSign := startSigningWorkers(2, sign, submissionRequests, responses, &signingWg, &transactionWg)

	rawTxn := &RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: 0,
		Payload: TransactionPayload{Payload: &EntryFunction{
			Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args:     [][]byte{AccountTwo[:], {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		}},
		MaxGasAmount:               100_000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainId:                    4,
	}

	transactionWg.Add(1)
	transactionsToSign <- TransactionBuildResponse{Id: 1, Response: rawTxn}

	transactionWg.Wait()
	close(transactionsToSign)
	signingWg.Wait()
	close(submissionRequests)

	req := <-submissionRequests
	assert.NotNil(t, req.SignedTxn)
	assert.Equal(t, uint64(1), req.Id)
}
