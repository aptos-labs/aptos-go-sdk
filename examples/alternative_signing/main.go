package main

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"golang.org/x/crypto/ed25519"
)

type AlternativeSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func (signer *AlternativeSigner) PublicKey() *crypto.Ed25519PublicKey {
	pubKey := &crypto.Ed25519PublicKey{}
	err := pubKey.FromBytes(signer.publicKey)
	if err != nil {
		panic("Public key is not valid")
	}
	return pubKey
}

func (signer *AlternativeSigner) PubKey() crypto.PublicKey {
	return signer.PublicKey()
}

func (signer *AlternativeSigner) ToHex() string {
	return ""
}

func (signer *AlternativeSigner) Sign(msg []byte) (authenticator *crypto.AccountAuthenticator, err error) {
	sigBytes := ed25519.Sign(signer.privateKey, msg)
	sig := &crypto.Ed25519Signature{}
	copy(sig.Inner[:], sigBytes)
	pubKey := signer.PublicKey()
	auth := &crypto.Ed25519Authenticator{
		PubKey: pubKey,
		Sig:    sig,
	}
	// TODO: maybe make convenience functions for this
	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorEd25519,
		Auth:    auth,
	}, nil
}

func (signer *AlternativeSigner) AuthKey() *crypto.AuthenticationKey {
	authKey := &crypto.AuthenticationKey{}
	pubKey := signer.PublicKey()
	authKey.FromPublicKey(pubKey)
	return authKey
}

// main This example shows you how to make an alternative signer for the SDK, if you prefer a different library
func main() {
	// Create a client for Aptos
	client, err := aptos.NewClient(aptos.LocalnetConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	println("We create a signer that we are calling 'externally' to the Go SDK, this could be on another server")
	publicKey, privateKey, _ := ed25519.GenerateKey(nil)
	signer := &AlternativeSigner{
		privateKey, publicKey,
	}

	// Create the sender from the key locally
	sender, err := aptos.NewAccountFromSigner(signer)
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	// Fund the sender with the faucet to create it on-chain
	err = client.Fund(sender.Address, 100_000_000)
	fmt.Printf("We fund the signer account %s with the faucet\n", sender.Address.String())

	// Prep arguments
	receiver := aptos.AccountAddress{}
	err = receiver.ParseStringRelaxed("0xBEEF")
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	amount := uint64(100)

	// Sign transaction
	signedTxn, err := aptos.APTTransferTransaction(client, sender, receiver, amount)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}
	fmt.Printf("Submit a coin transfer to address %s\n", receiver.String())

	// Submit and wait for it to complete
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// Wait for the transaction
	fmt.Printf("And we wait for the transaction %s to complete...\n", txnHash)
	waitResponse, err := client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
	userTxn, _ := waitResponse.UserTransaction()
	fmt.Printf("The transaction completed with hash: %s and version %d\n", userTxn.Hash, userTxn.Version)
}
