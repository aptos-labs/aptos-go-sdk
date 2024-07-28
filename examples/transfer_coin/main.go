// transfer_coin is an example of how to make a coin transfer transaction in the simplest possible way
package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/client"
	"github.com/aptos-labs/aptos-go-sdk/types"
)

const FundAmount = 100_000_000
const TransferAmount = 1_000

// example This example shows you how to make an APT transfer transaction in the simplest possible way
func example(networkConfig client.NetworkConfig) {
	// Create a client for Aptos
	aptosClient, err := client.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create accounts locally for alice and bob
	alice, err := types.NewEd25519Account()
	if err != nil {
		panic("Failed to create alice:" + err.Error())
	}
	bob, err := types.NewEd25519Account()
	if err != nil {
		panic("Failed to create bob:" + err.Error())
	}

	fmt.Printf("\n=== Addresses ===\n")
	fmt.Printf("Alice: %s\n", alice.Address.String())
	fmt.Printf("Bob:%s\n", bob.Address.String())

	// Fund the sender with the faucet to create it on-chain
	err = aptosClient.Fund(alice.Address, FundAmount)
	if err != nil {
		panic("Failed to fund alice:" + err.Error())
	}

	aliceBalance, err := aptosClient.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err := aptosClient.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Initial Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)

	// Build transaction
	rawTxn, err := client.APTTransferTransaction(aptosClient, alice, bob.Address, TransferAmount)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(alice)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// Submit and wait for it to complete
	submitResult, err := aptosClient.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// Wait for the transaction
	_, err = aptosClient.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
	aliceBalance, err = aptosClient.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err = aptosClient.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Intermediate Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)

	// Now submit as a single call, with a custom payload
	accountBytes, err := bcs.Serialize(&bob.Address)
	if err != nil {
		panic("Failed to serialize bob's address:" + err.Error())
	}

	amountBytes, err := bcs.SerializeU64(TransferAmount)
	if err != nil {
		panic("Failed to serialize transfer amount:" + err.Error())
	}

	resp, err := aptosClient.BuildSignAndSubmitTransaction(alice, types.TransactionPayload{
		Payload: &types.EntryFunction{
			Module: types.ModuleId{
				Address: types.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []types.TypeTag{},
			Args: [][]byte{
				accountBytes,
				amountBytes,
			},
		}},
	)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	_, err = aptosClient.WaitForTransaction(resp.Hash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	aliceBalance, err = aptosClient.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err = aptosClient.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Final Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)
}

func main() {
	example(client.DevnetConfig)
}
