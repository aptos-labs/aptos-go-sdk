// transfer_coin is an example of how to make a coin transfer transaction in the simplest possible way
package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

const (
	FundAmount     = 100_000_000
	TransferAmount = 1_000
)

const contract = "0xbe4a375aaa9d2b43e57b482b7e7124859c7c84c4d55dfcb1b207a720fc789304"

// example This example shows you how to make an APT transfer transaction in the simplest possible way
func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create accounts locally for alice and bob
	// alice, err := aptos.NewEd25519Account()
	// if err != nil {
	// 	panic("Failed to create alice:" + err.Error())
	// }
	// fmt.Println(alice.PrivateKeyString())

	privateKey := &crypto.Ed25519PrivateKey{}
	err = privateKey.FromHex("0xef770601d8ec57e411bb18b053caa6bb76e20421cb55c3ea51ffe4aac31fd9d4")
	if err != nil {
		panic("Failed to parse private key:" + err.Error())
	}
	alice, err := aptos.NewAccountFromSigner(privateKey)
	if err != nil {
		panic("Failed to create myAlice:" + err.Error())
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

	// 1. Build transaction, with a single move function ABI
	// testU8FunctionAbi := &api.MoveFunction{
	// 	Name:              "test_u8",
	// 	Visibility:        "public",
	// 	IsEntry:           true,
	// 	IsView:            false,
	// 	GenericTypeParams: []*api.GenericTypeParam{},
	// 	Params:            []string{"u8"},
	// 	Return:            []string{},
	// }

	testVectorU8FunctionAbi := &api.MoveFunction{
		Name:              "test_vector_u8",
		Visibility:        "public",
		IsEntry:           true,
		IsView:            false,
		GenericTypeParams: []*api.GenericTypeParam{},
		Params:            []string{"vector<u8>"},
		Return:            []string{},
	}

	addressBytes, err := util.ParseHex(contract)
	if err != nil {
		panic(fmt.Sprintf("Failed to ParseHex: %s", err.Error()))
	}

	var address types.AccountAddress
	copy(address[:], addressBytes)

	payload, err := aptos.EntryFunctionFromAbi(testVectorU8FunctionAbi, address, "args_test", "test_vector_u8", []any{}, []any{[]any{1, 2, 3}})
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
}

func main() {
	example(aptos.DevnetConfig)
}
