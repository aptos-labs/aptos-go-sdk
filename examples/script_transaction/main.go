// fungible_asset is an example of how to create and transfer fungible assets
package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const testEd25519PrivateKey = "ed25519-priv-0xc5338cd251c22daa8c9c9cc94f498cc8a5c7e1d2e75287a5dda91096fe64efa5"

const TransferScriptPayload = "0x008601a11ceb0b0700000a06010002030206050806070e1708252010451900000001010000010003060c05030d6170746f735f6163636f756e74087472616e736665720000000000000000000000000000000000000000000000000000000000000001166170746f733a3a7363726970745f636f6d706f7365720100000100050a000b010b0211000200020920000000000000000000000000000000000000000000000000000000000000000209086400000000000000"
const TransferAmount = uint64(1_000)
const FundAmount = uint64(100_000_000)

func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create a sender locally
	key := crypto.Ed25519PrivateKey{}
	err = key.FromHex(testEd25519PrivateKey)
	if err != nil {
		panic("Failed to decode Ed25519 private key:" + err.Error())
	}
	sender, err := aptos.NewAccountFromSigner(&key)
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	// Fund the sender with the faucet to create it on-chain
	println("SENDER: ", sender.Address.String())
	err = client.Fund(sender.Address, FundAmount)
	if err != nil {
		panic("Failed to fund sender:" + err.Error())
	}
	receiver := &aptos.AccountAddress{}

	// Now run a script version
	fmt.Printf("\n== Now running script version ==\n")
	runScript(client, sender, receiver)

	if err != nil {
		panic("Failed to get store balance:" + err.Error())
	}
	// fmt.Printf("After Script: Receiver Before transfer: %d, after transfer: %d\n", receiverAfterBalance, receiverAfterAfterBalance)
}

func runScript(client *aptos.Client, alice *aptos.Account, bob *aptos.AccountAddress) {
	scriptBytes, err := aptos.ParseHex(TransferScriptPayload)
	if err != nil {
		panic("Failed to parse script:" + err.Error())
	}

	var payload aptos.TransactionPayload
	payload.UnmarshalBCS(bcs.NewDeserializer(scriptBytes))

	// 1. Build transaction
	rawTxn, err := client.BuildTransaction(alice.AccountAddress(), payload)

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
}

// main This example shows how to send a transaction with a script
func main() {
	example(aptos.DevnetConfig)
}
