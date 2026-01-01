package integration_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/qimeila/aptos-go-sdk"
	"github.com/qimeila/aptos-go-sdk/internal/testutil"
)

func TestBuildSignAndSubmitTransactionsWithSignFnAndWorkerPoolWithOneSender(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test expects network connection to devnet")
	}
	t.Parallel()
	const (
		numTransactions = uint32(5)
		transferAmount  = uint64(100)
		numWorkers      = uint32(3)
		initialFunding  = uint64(200_000_000)
	)

	clients := testutil.SetupTestClients(t)
	sender := testutil.SetupTestAccount(t, clients.Client, initialFunding)
	receiver := testutil.SetupTestAccount(t, clients.Client, 0)

	payloads := make(chan aptos.TransactionBuildPayload, numTransactions)
	responses := make(chan aptos.TransactionSubmissionResponse, numTransactions)
	workerPoolConfig := aptos.WorkerPoolConfig{
		NumWorkers:          numWorkers,
		BuildResponseBuffer: numTransactions,
		SubmissionBuffer:    numTransactions,
	}

	startTime := time.Now()

	go clients.NodeClient.BuildSignAndSubmitTransactionsWithSignFnAndWorkerPool(
		sender.Account.Address,
		payloads,
		responses,
		func(rawTxn aptos.RawTransactionImpl) (*aptos.SignedTransaction, error) {
			switch txn := rawTxn.(type) {
			case *aptos.RawTransaction:
				return txn.SignedTransaction(sender.Account)
			default:
				return nil, errors.New("unsupported transaction type")
			}
		},
		workerPoolConfig,
	)

	for txNum := range numTransactions {
		payload := testutil.CreateTransferPayload(t, receiver.Account.Address, transferAmount)
		payloads <- aptos.TransactionBuildPayload{
			Id:    uint64(txNum),
			Inner: payload,
			Type:  aptos.TransactionSubmissionTypeSingle,
		}
	}
	close(payloads)

	for range numTransactions {
		resp := <-responses
		if resp.Err != nil {
			t.Errorf("Transaction failed: %v", resp.Err)
			continue
		}
		fmt.Printf("[%s] hash: %s\n",
			time.Now().Format("15:04:05.000"),
			resp.Response.Hash)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nTotal Duration: %v\n", duration)
}
