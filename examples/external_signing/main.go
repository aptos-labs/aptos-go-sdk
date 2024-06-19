package main

import (
	"encoding/binary"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"golang.org/x/crypto/ed25519"
)

type ExternalSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func (signer *ExternalSigner) PublicKey() *crypto.Ed25519PublicKey {
	pubKey := &crypto.Ed25519PublicKey{}
	err := pubKey.FromBytes(signer.publicKey)
	if err != nil {
		panic("Public key is not valid")
	}
	return pubKey
}

func (signer *ExternalSigner) PubKey() crypto.PublicKey {
	return signer.PublicKey()
}

func (signer *ExternalSigner) ToHex() string {
	return ""
}

func (signer *ExternalSigner) Sign(msg []byte) (authenticator *crypto.AccountAuthenticator, err error) {
	sig, err := signer.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	pubKey := signer.PublicKey()
	auth := &crypto.Ed25519Authenticator{
		PubKey: pubKey,
		Sig:    sig.(*crypto.Ed25519Signature),
	}
	// TODO: maybe make convenience functions for this
	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorEd25519,
		Auth:    auth,
	}, nil
}

func (signer *ExternalSigner) SignMessage(msg []byte) (signature crypto.Signature, err error) {
	sigBytes := ed25519.Sign(signer.privateKey, msg)
	sig := &crypto.Ed25519Signature{}
	copy(sig.Inner[:], sigBytes)
	return sig, nil
}

func (signer *ExternalSigner) AuthKey() *crypto.AuthenticationKey {
	authKey := &crypto.AuthenticationKey{}
	pubKey := signer.PublicKey()
	authKey.FromPublicKey(pubKey)
	return authKey
}

// main This example shows you how to make an alternative signer for the SDK, if you prefer a different library
func main() {
	// Create a client for Aptos
	client, err := aptos.NewClient(aptos.DevnetConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	println("We create a signer that we are calling 'externally' to the Go SDK, this could be on another server")
	publicKey, privateKey, _ := ed25519.GenerateKey(nil)
	signer := &ExternalSigner{
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
	var amountBytes [8]byte
	binary.LittleEndian.PutUint64(amountBytes[:], amount)

	// Sign transaction
	fmt.Printf("Submit a coin transfer to address %s\n", receiver.String())
	rawTxn := &aptos.RawTransaction{
		Sender:         sender.Address,
		SequenceNumber: 0,
		Payload: aptos.TransactionPayload{Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: aptos.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{
				receiver[:],
				amountBytes[:],
			},
		}},
		MaxGasAmount:               1000,
		GasUnitPrice:               2000,
		ExpirationTimestampSeconds: 1714158778,
		ChainId:                    4,
	}
	fmt.Printf("Sign the message %s\n", receiver.String())
	// Build a signing message
	signingMessage, err := rawTxn.SigningMessage()
	if err != nil {
		panic("Failed to build signing message:" + err.Error())
	}

	// Send it to our external signer
	auth, err := signer.Sign(signingMessage)
	if err != nil {
		panic("Failed to sign message:" + err.Error())
	}

	txnAuth := &aptos.TransactionAuthenticator{
		Variant: aptos.TransactionAuthenticatorEd25519,
		Auth:    auth,
	}

	// Build a signed transaction
	signedTxn := &aptos.SignedTransaction{
		Transaction:   rawTxn,
		Authenticator: txnAuth,
	}
	// TODO: Show how to send over a wire with an encoding

	// Submit and wait for it to complete
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// Wait for the transaction
	fmt.Printf("And we wait for the transaction %s to complete...\n", txnHash)
	userTxn, err := client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	fmt.Printf("The transaction completed with hash: %s and version %d\n", userTxn.Hash, userTxn.Version)
}
