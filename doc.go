// Package aptos is a Go interface into the Aptos blockchain.
//
// The Aptos Go SDK provides a way to read on-chain data, submit transactions, and generally interact with the blockchain.
//
// Quick links:
//
//   - [Aptos Dev Docs] for learning more about Aptos and how to use it.
//   - [Examples] are standalone runnable examples of how to use the SDK.
//
// You can create a client and send a transfer transaction with the below example:
//
//	// Create a Client
//	client := NewClient(DevnetConfig)
//
//	// Create an account, and fund it
//	account := NewEd25519Account()
//	err := client.Fund(account.AccountAddress())
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to fund account %s %w", account.AccountAddress().ToString(), err))
//	}
//
//	// Send funds to a different address
//	receiver := &AccountAddress{}
//	receiver.ParseStringRelaxed("0xcafe")
//
//	// Build a transaction to send 1 APT to the receiver
//	amount := 100_000_000 // 1 APT
//	transferTransaction, err := APTTransferTransaction(client, account, receiver, amount)
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to build transaction %w", err))
//	}
//
//	// Submit transaction to the blockchain
//	submitResponse, err := client.SubmitTransaction(transferTransaction)
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to submit transaction %w", err))
//	}
//
//	// Wait for transaction to complete
//	err := client.WaitForTransaction(submitResponse.Hash)
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to wait for transaction %w", err))
//	}
//
// [Aptos Dev Docs]: https://aptos.dev
// [Examples]: https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk/examples
package aptos
