package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/internal/testutil"
)

func TestBuildSignAndSubmitTransactionsWithSignFnAndWorkerPoolWithMultipleSenders(t *testing.T) {
	const (
		numSenders     = uint64(3)
		txPerSender    = uint64(5)
		initialFunding = uint64(100_000_000)
		transferAmount = uint64(100)
	)

	clients := testutil.SetupTestClients(t)

	// Create and fund senders
	senders := make([]testutil.TestAccount, numSenders)
	for i := uint64(0); i < numSenders; i++ {
		senders[i] = testutil.SetupTestAccount(t, clients.Client, initialFunding)
	}

	receiver := testutil.SetupTestAccount(t, clients.Client, 0)

	startTime := time.Now()

	// Process transactions for each sender
	doneCh := make(chan struct{})

	for senderIdx := uint64(0); senderIdx < numSenders; senderIdx++ {
		go func(senderIdx uint64) {
			defer func() {
				doneCh <- struct{}{}
			}()

			sender := senders[senderIdx]
			payloads := make(chan aptos.TransactionBuildPayload, txPerSender)
			responses := make(chan aptos.TransactionSubmissionResponse, txPerSender)

			go clients.NodeClient.BuildSignAndSubmitTransactionsWithSignFnAndWorkerPool(
				sender.Account.Address,
				payloads,
				responses,
				func(rawTxn aptos.RawTransactionImpl) (*aptos.SignedTransaction, error) {
					switch txn := rawTxn.(type) {
					case *aptos.RawTransaction:
						return txn.SignedTransaction(sender.Account)
					default:
						return nil, fmt.Errorf("unsupported transaction type")
					}
				},
				aptos.WorkerPoolConfig{NumWorkers: 3},
			)

			workerStartTime := time.Now()
			for txNum := uint64(0); txNum < txPerSender; txNum++ {
				payload := testutil.CreateTransferPayload(t, receiver.Account.Address, transferAmount)

				payloads <- aptos.TransactionBuildPayload{
					Id:    uint64(txNum),
					Inner: payload,
					Type:  aptos.TransactionSubmissionTypeSingle,
				}
			}
			close(payloads)

			for i := uint64(0); i < txPerSender; i++ {
				resp := <-responses
				if resp.Err != nil {
					t.Errorf("Transaction failed: %v", resp.Err)
					continue
				}
				fmt.Printf("[%s] Worker %d → hash: %s\n",
					time.Now().Format("15:04:05.000"),
					senderIdx,
					resp.Response.Hash)
			}

			fmt.Printf("[%s] Worker %d completed all transactions (t+%v)\n",
				time.Now().Format("15:04:05.000"),
				senderIdx,
				time.Since(workerStartTime).Round(time.Millisecond))
		}(senderIdx)
	}

	// Wait for all senders to complete
	for i := uint64(0); i < numSenders; i++ {
		<-doneCh
	}

	duration := time.Since(startTime)
	fmt.Printf("\nTotal Duration: %v\n", duration)
}
