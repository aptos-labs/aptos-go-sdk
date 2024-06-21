package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const FundAmount = 100_000_000
const TransferAmount = 1_000

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
	sponsor, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create sponsor:" + err.Error())
	}

	fmt.Printf("\n=== Addresses ===\n")
	fmt.Printf("Alice: %s\n", alice.Address.String())
	fmt.Printf("Bob:%s\n", bob.Address.String())
	fmt.Printf("Sponsor:%s\n", sponsor.Address.String())

	// Fund the alice with the faucet to create it on-chain
	err = client.Fund(alice.Address, FundAmount)
	if err != nil {
		panic("Failed to fund alice:" + err.Error())
	}

	// And the sponsor
	err = client.Fund(sponsor.Address, FundAmount)
	if err != nil {
		panic("Failed to fund sponsor:" + err.Error())
	}

	aliceBalance, err := client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err := client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	sponsorBalance, err := client.AccountAPTBalance(sponsor.Address)
	if err != nil {
		panic("Failed to retrieve sponsor balance:" + err.Error())
	}
	fmt.Printf("\n=== Initial Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob: %d\n", bobBalance)
	fmt.Printf("Sponsor: %d\n", sponsorBalance)

	// Build transaction
	transferPayload, err := aptos.CoinTransferPayload(&aptos.AptosCoinTypeTag, bob.Address, TransferAmount)
	if err != nil {
		panic("Failed to build transfer payload:" + err.Error())
	}
	rawTxn, err := client.BuildTransactionMultiAgent(
		alice.Address,
		aptos.TransactionPayload{
			Payload: transferPayload,
		},
		aptos.FeePayer(&sponsor.Address),
	)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Sign transaction
	aliceAuth, err := rawTxn.Sign(alice)
	if err != nil {
		panic("Failed to sign transaction as sender:" + err.Error())
	}
	sponsorAuth, err := rawTxn.Sign(sponsor)
	if err != nil {
		panic("Failed to sign transaction as sponsor:" + err.Error())
	}

	signedFeePayerTxn, ok := rawTxn.ToFeePayerSignedTransaction(
		aliceAuth,
		sponsorAuth,
		[]crypto.AccountAuthenticator{},
	)
	if !ok {
		panic("Failed to build fee payer signed transaction")
	}

	// Submit and wait for it to complete
	submitResult, err := client.SubmitTransaction(signedFeePayerTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash
	println("Submitted transaction hash:", txnHash)

	// Wait for the transaction
	_, err = client.WaitForTransaction(txnHash)
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
	sponsorBalance, err = client.AccountAPTBalance(sponsor.Address)
	if err != nil {
		panic("Failed to retrieve sponsor balance:" + err.Error())
	}
	fmt.Printf("\n=== Intermediate Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob: %d\n", bobBalance)
	fmt.Printf("Sponsor: %d\n", sponsorBalance)

	fmt.Printf("\n=== Now do it without knowing the signer ahead of time ===\n")

	rawTxn, err = client.BuildTransactionMultiAgent(
		alice.Address,
		aptos.TransactionPayload{
			Payload: transferPayload,
		},
		aptos.FeePayer(&aptos.AccountZero), // Note that the Address is 0x0, because we don't know the signer
	)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Alice signs the transaction, without knowing the sponsor
	aliceAuth, err = rawTxn.Sign(alice)
	if err != nil {
		panic("Failed to sign transaction as sender:" + err.Error())
	}

	// The sponsor has to add themselves to the transaction to sign, note that this would likely be on a different
	// server
	ok = rawTxn.SetFeePayer(sponsor.Address)
	if !ok {
		panic("Failed to set fee payer")
	}

	sponsorAuth, err = rawTxn.Sign(sponsor)
	if err != nil {
		panic("Failed to sign transaction as sponsor:" + err.Error())
	}

	signedFeePayerTxn, ok = rawTxn.ToFeePayerSignedTransaction(
		aliceAuth,
		sponsorAuth,
		[]crypto.AccountAuthenticator{},
	)
	if !ok {
		panic("Failed to build fee payer signed transaction")
	}

	// Submit and wait for it to complete
	submitResult, err = client.SubmitTransaction(signedFeePayerTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash = submitResult.Hash
	println("Submitted transaction hash:", txnHash)

	// Wait for the transaction
	_, err = client.WaitForTransaction(txnHash)
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
	sponsorBalance, err = client.AccountAPTBalance(sponsor.Address)
	if err != nil {
		panic("Failed to retrieve sponsor balance:" + err.Error())
	}
	fmt.Printf("\n=== Final Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob: %d\n", bobBalance)
	fmt.Printf("Sponsor: %d\n", sponsorBalance)
}

func main() {
	example(aptos.DevnetConfig)
}
