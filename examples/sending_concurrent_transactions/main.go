// sending_concurrent_transactions shows how to submit transactions serially or concurrently on a single account
package main

import (
	"time"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/client"
	"github.com/aptos-labs/aptos-go-sdk/types"
)

func setup(networkConfig client.NetworkConfig) (*client.Client, types.TransactionSigner) {
	aptosClient, err := client.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	sender, err := types.NewEd25519Account()
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	err = aptosClient.Fund(sender.Address, 100_000_000)
	if err != nil {
		panic("Failed to fund sender:" + err.Error())
	}

	return aptosClient, sender
}

func payload() types.TransactionPayload {
	receiver := types.AccountAddress{}
	err := receiver.ParseStringRelaxed("0xBEEF")
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	amount := uint64(100)
	p, err := types.CoinTransferPayload(nil, receiver, amount)
	if err != nil {
		panic("Failed to serialize arguments:" + err.Error())
	}
	return types.TransactionPayload{Payload: p}
}

func sendManyTransactionsSerially(networkConfig client.NetworkConfig, numTransactions uint64) {
	aptosClient, sender := setup(networkConfig)

	responses := make([]*api.SubmitTransactionResponse, numTransactions)
	payload := payload()

	senderAddress := sender.AccountAddress()
	sequenceNumber := uint64(0)
	for i := uint64(0); i < numTransactions; i++ {
		rawTxn, err := aptosClient.BuildTransaction(senderAddress, payload, client.SequenceNumber(sequenceNumber))
		if err != nil {
			panic("Failed to build transaction:" + err.Error())
		}

		signedTxn, err := rawTxn.SignedTransaction(sender)
		if err != nil {
			panic("Failed to sign transaction:" + err.Error())
		}

		submitResult, err := aptosClient.SubmitTransaction(signedTxn)
		if err != nil {
			panic("Failed to submit transaction:" + err.Error())
		}
		responses[i] = submitResult
		sequenceNumber++
	}

	// Wait on last transaction
	response, err := aptosClient.WaitForTransaction(responses[numTransactions-1].Hash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
	if response.Success == false {
		panic("Transaction failed due to " + response.VmStatus)
	}
}

func sendManyTransactionsConcurrently(networkConfig client.NetworkConfig, numTransactions uint64) {
	aptosClient, sender := setup(networkConfig)
	payload := payload()

	// start submission goroutine
	payloads := make(chan client.TransactionBuildPayload, 50)
	results := make(chan client.TransactionSubmissionResponse, 50)
	go aptosClient.BuildSignAndSubmitTransactions(sender, payloads, results)

	// Submit transactions to goroutine
	go func() {
		for i := uint64(0); i < numTransactions; i++ {
			payloads <- client.TransactionBuildPayload{
				Id:    i,
				Type:  client.TransactionSubmissionTypeSingle,
				Inner: payload,
			}
		}
		close(payloads)
	}()

	// Wait for all transactions to be processed
	for result := range results {
		if result.Err != nil {
			panic("Failed to submit and wait for transaction:" + result.Err.Error())
		}
	}
}

// example This example shows you how to improve performance of the transaction submission
//
// Speed can be improved by locally handling the sequence number, gas price, and other factors
func example(networkConfig client.NetworkConfig, numTransactions uint64) {
	println("Sending", numTransactions, "transactions Serially")
	startSerial := time.Now()
	sendManyTransactionsSerially(networkConfig, numTransactions)
	endSerial := time.Now()
	println("Serial:", time.Duration.Milliseconds(endSerial.Sub(startSerial)), "ms")

	println("Sending", numTransactions, "transactions Concurrently")
	startConcurrent := time.Now()
	sendManyTransactionsConcurrently(networkConfig, numTransactions)
	endConcurrent := time.Now()
	println("Concurrent:", time.Duration.Milliseconds(endConcurrent.Sub(startConcurrent)), "ms")

	println("Concurrent is", time.Duration.Milliseconds(endSerial.Sub(startSerial)-endConcurrent.Sub(startConcurrent)), "ms faster than Serial")
}

func main() {
	example(client.DevnetConfig, 100)
}
