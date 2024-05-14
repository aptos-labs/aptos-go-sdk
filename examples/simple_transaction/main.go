package main

import (
	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"github.com/aptos-labs/aptos-go-sdk/examples"
)

// main This example shows you how to make an APT transfer transaction in the simplest possible way
func main() {
	// Create a client for Aptos
	client, err := aptos_go_sdk.NewClient(aptos_go_sdk.DevnetConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create a sender locally
	sender, err := core.NewEd25519Account()
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	// Fund the sender with the faucet to create it onchain
	err = client.Fund(sender.Address, 100_000_000)

	// Prep arguments
	receiver := core.AccountAddress{}
	err = receiver.ParseStringRelaxed("0xBEEF")
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	amount := uint64(100)

	// Sign transaction
	signedTxn, err := aptos_go_sdk.APTTransferTransaction(client, sender, receiver, amount)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// Submit and wait for it to complete
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult["hash"].(string)

	// Wait for the transaction
	txn, err := client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	println(examples.PrettyJson(txn))
}
