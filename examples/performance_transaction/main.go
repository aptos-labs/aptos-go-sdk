// performance_transaction shows how to improve performance of the transaction submission of a single transaction
package main

import (
	"encoding/json"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
)

// example This example shows you how to improve performance of the transaction submission
//
// Speed can be improved by locally handling the sequence number, gas price, and other factors
func example(networkConfig aptos.NetworkConfig) {
	start := time.Now()
	before := time.Now()
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}
	println("New client:    ", time.Since(before).Milliseconds(), "ms")

	// Create a sender locally
	sender, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	println("Create sender:", time.Since(before).Milliseconds(), "ms")

	before = time.Now()

	// Fund the sender with the faucet to create it on-chain
	err = client.Fund(sender.Address, 100_000_000)
	if err != nil {
		panic("Failed to fund:" + err.Error())
	}

	println("Fund sender:", time.Since(before).Milliseconds(), "ms")

	before = time.Now()

	// Prep arguments
	receiver := aptos.AccountAddress{}
	err = receiver.ParseStringRelaxed("0xBEEF")
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	amount := uint64(100)

	// Serialize arguments
	payload, err := aptos.CoinTransferPayload(nil, receiver, amount)
	if err != nil {
		panic("Failed to serialize arguments:" + err.Error())
	}

	rawTxn, err := client.BuildTransaction(sender.Address,
		aptos.TransactionPayload{Payload: payload}, aptos.SequenceNumber(0)) // Use the sequence number to skip fetching it
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	println("Build transaction:", time.Since(before).Milliseconds(), "ms")

	// Sign transaction
	before = time.Now()

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(sender)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}
	println("Sign transaction:", time.Since(before).Milliseconds(), "ms")

	before = time.Now()
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash
	println("Submit transaction:", time.Since(before).Milliseconds(), "ms")

	// Wait for the transaction
	before = time.Now()
	txn, err := client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
	println("Wait for transaction:", time.Since(before).Milliseconds(), "ms")

	println("Total time:    ", time.Since(start).Milliseconds(), "ms")
	txnStr, err := json.Marshal(txn)
	if err != nil {
		panic("Failed to marshal transaction:" + err.Error())
	}
	println(string(txnStr))
}

func main() {
	example(aptos.DevnetConfig)
}
