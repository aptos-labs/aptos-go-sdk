package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const MultiagentScript = "0xa11ceb0b0700000a0601000403040d04110405151b07302f085f2000000001010203040001000306020100010105010704060c060c03030205050001060c010501090003060c05030109010d6170746f735f6163636f756e74067369676e65720a616464726573735f6f660e7472616e736665725f636f696e73000000000000000000000000000000000000000000000000000000000000000102000000010f0a0011000c040a0111000c050b000b050b0238000b010b040b03380102"

const FundAmount = 100_000_000
const TransferAmount uint64 = 1

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
		panic("Failed to create sender:" + err.Error())
	}

	bob, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	fmt.Printf("\n=== Addresses ===\n")
	fmt.Printf("Alice: %s\n", alice.Address.String())
	fmt.Printf("Bob:%s\n", bob.Address.String())

	// Fund the sender with the faucet to create it on-chain
	err = client.Fund(alice.Address, FundAmount)
	if err != nil {
		panic("Failed to fund alice:" + err.Error())
	}
	err = client.Fund(bob.Address, FundAmount)
	if err != nil {
		panic("Failed to fund bob:" + err.Error())
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

	// 1. Build transaction
	// script, err := util.ParseHex(MultiagentScript)
	// if err != nil {
	// 	panic("Failed to deserialize script:" + err.Error())
	// }

	serializer := &bcs.Serializer{}

	bob.Address.MarshalBCS(serializer)
	accountBytes := serializer.ToBytes()

	if err != nil {
		panic("Failed to serialize alice's address:" + err.Error())
	}

	serializer = &bcs.Serializer{}

	serializer.U64(TransferAmount)
	amountBytes := serializer.ToBytes()

	if err != nil {
		panic("Failed to serialize transfer amount:" + err.Error())
	}

	rawTxn, err := client.BuildTransactionMultiAgent(alice.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: aptos.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{
				accountBytes,
				amountBytes,
			},
		}}, aptos.FeePayer(&bob.Address))
	if err != nil {
		panic("Failed to build multiagent raw transaction:" + err.Error())
	}

	// 2. Simulate transaction (optional)
	// This is useful for understanding how much the transaction will cost
	// and to ensure that the transaction is valid before sending it to the network
	// This is optional, but recommended
	// TODO: Support simulate transaction with multi-agent / fee payer
	simulationResult, err := client.SimulateMultiTransaction(rawTxn, alice)
	if err != nil {
		panic("Failed to simulate transaction:" + err.Error())
	}
	fmt.Printf("\n=== Simulation ===\n")
	fmt.Printf("Gas unit price: %d\n", simulationResult[0].GasUnitPrice)
	fmt.Printf("Gas used: %d\n", simulationResult[0].GasUsed)
	fmt.Printf("Total gas fee: %d\n", simulationResult[0].GasUsed*simulationResult[0].GasUnitPrice)
	fmt.Printf("Status: %s\n", simulationResult[0].VmStatus)

	// 3. Sign transaction with both parties separately, this would be on different machines or places
	aliceAuth, err := rawTxn.Sign(alice)
	if err != nil {
		panic("Failed to sign multiagent transaction with alice:" + err.Error())
	}
	bobAuth, err := rawTxn.Sign(bob)
	if err != nil {
		panic("Failed to sign multiagent transaction with bob:" + err.Error())
	}

	// 3.a. merge the signatures together into a single transaction
	signedTxn, ok := rawTxn.ToMultiAgentSignedTransaction(aliceAuth, []crypto.AccountAuthenticator{*bobAuth})
	if !ok {
		panic("Failed to build a signed multiagent transaction")
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
