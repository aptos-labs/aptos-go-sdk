// transfer_coin is an example of how to make a coin transfer transaction in the simplest possible way
package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
)

const (
	FundAmount     = 100_000_000
	TransferAmount = 1_000
)

// example This example shows you how to make an APT transfer transaction in the simplest possible way
func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create accounts locally for alice and bob
	alice, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create alice:" + err.Error())
	}
	bob, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create bob:" + err.Error())
	}

	fmt.Printf("\n=== Addresses ===\n")
	fmt.Printf("Alice: %s\n", alice.Address.String())
	fmt.Printf("Bob:%s\n", bob.Address.String())

	// Fund the sender with the faucet to create it on-chain
	err = client.Fund(alice.Address, FundAmount)
	if err != nil {
		panic("Failed to fund alice:" + err.Error())
	}

	aliceBalance, err := client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err := client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Initial Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)

	// 1. Build transaction, with easy mode, no BCS serialization necessary
	payload, err := client.EntryFunctionWithArgs(aptos.AccountOne, "aptos_account", "transfer_coins", []any{"0x1::aptos_coin::AptosCoin"}, []any{bob.Address, TransferAmount})
	if err != nil {
		panic("Failed to call EntryFunctionWithArgs:" + err.Error())
	}
	rawTxn, err := client.BuildTransaction(alice.AccountAddress(), aptos.TransactionPayload{
		Payload: payload,
	})
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// 2. Simulate transaction (optional)
	// This is useful for understanding how much the transaction will cost
	// and to ensure that the transaction is valid before sending it to the network
	// This is optional, but recommended
	simulationResult, err := client.SimulateTransaction(rawTxn, alice)
	if err != nil {
		panic("Failed to simulate transaction:" + err.Error())
	}
	fmt.Printf("\n=== Simulation ===\n")
	fmt.Printf("Gas unit price: %d\n", simulationResult[0].GasUnitPrice)
	fmt.Printf("Gas used: %d\n", simulationResult[0].GasUsed)
	fmt.Printf("Total gas fee: %d\n", simulationResult[0].GasUsed*simulationResult[0].GasUnitPrice)
	fmt.Printf("Status: %s\n", simulationResult[0].VmStatus)

	// 3. Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(alice)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// 4. Submit transaction
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// 5. Wait for the transaction to complete
	_, err = client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	// Check balances
	aliceBalance, err = client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err = client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Intermediate Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)

	// Now do it again, but with a different function, with types
	payload2, err := client.EntryFunctionWithArgs(aptos.AccountOne, "aptos_account", "transfer_coins", []any{aptos.AptosCoinTypeTag}, []any{bob.Address.String(), "1000"})
	if err != nil {
		panic("Failed to call EntryFunctionWithArgs:" + err.Error())
	}
	resp, err := client.BuildSignAndSubmitTransaction(alice, aptos.TransactionPayload{
		Payload: payload2,
	})
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	_, err = client.WaitForTransaction(resp.Hash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	aliceBalance, err = client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err = client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Final Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)
}

func main() {
	example(aptos.DevnetConfig)
}
