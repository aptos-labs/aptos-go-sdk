package aptos_test

import (
	"context"
	"fmt"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/iter"
	"github.com/aptos-labs/aptos-go-sdk/v2/transaction"
)

func Example_basicUsage() {
	ctx := context.Background()

	// Create a client for devnet
	client, err := aptos.NewClient(aptos.Devnet)
	if err != nil {
		panic(err)
	}

	// Get node info
	info, err := client.Info(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Chain ID: %d\n", info.ChainID)
	fmt.Printf("Block Height: %d\n", info.BlockHeight)
}

func Example_clientOptions() {
	// Create a client with custom options
	_, _ = aptos.NewClient(aptos.Mainnet,
		aptos.WithTimeout(60*time.Second),
		aptos.WithRetry(5, 200*time.Millisecond),
		aptos.WithAPIKey("your-api-key"),
	)
}

func Example_accountOperations() {
	ctx := context.Background()

	client, _ := aptos.NewClient(aptos.Devnet)

	// Parse an address
	address := aptos.MustParseAddress("0x1")

	// Get account info
	info, err := client.Account(ctx, address)
	if err != nil {
		// Handle error
		return
	}

	fmt.Printf("Sequence Number: %d\n", info.SequenceNumber)

	// Get account balance
	balance, err := client.AccountBalance(ctx, address)
	if err != nil {
		return
	}

	fmt.Printf("Balance: %d octas (%.4f APT)\n", balance, float64(balance)/1e8)
}

func Example_transactionBuilder() {
	ctx := context.Background()

	client, _ := aptos.NewClient(aptos.Devnet)

	sender := aptos.MustParseAddress("0x123")
	recipient := aptos.MustParseAddress("0x456")
	amount := uint64(1_000_000) // 0.01 APT

	// Build a transfer transaction using the fluent builder
	txn, err := transaction.New().
		Sender(sender).
		EntryFunction("0x1::aptos_account::transfer",
			nil, // no type args
			recipient.Bytes(),
			amount,
		).
		MaxGas(2000).
		GasPrice(100).
		Expiration(30*time.Second).
		Build(ctx, client)
	if err != nil {
		return
	}

	fmt.Printf("Built transaction for sender: %s\n", txn.Sender)
}

func Example_transferAPT() {
	ctx := context.Background()
	client, _ := aptos.NewClient(aptos.Devnet)

	sender := aptos.MustParseAddress("0x123")
	recipient := aptos.MustParseAddress("0x456")

	// Use the convenience builder
	txn, err := transaction.TransferAPT(sender, recipient, 1_000_000).
		MaxGas(2000).
		Build(ctx, client)
	if err != nil {
		return
	}

	_ = txn
}

func Example_viewFunction() {
	ctx := context.Background()
	client, _ := aptos.NewClient(aptos.Devnet)

	// Call a view function
	result, err := client.View(ctx, &aptos.ViewPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
		Function: "balance",
		TypeArgs: []aptos.TypeTag{aptos.AptosCoinTypeTag},
		Args:     []any{"0x1"},
	})
	if err != nil {
		return
	}

	fmt.Printf("Result: %v\n", result)
}

func Example_iterators() {
	ctx := context.Background()
	client, _ := aptos.NewClient(aptos.Devnet)

	// Stream transactions
	for txn, err := range client.TransactionsIter(ctx, nil) {
		if err != nil {
			break
		}
		fmt.Printf("Transaction: %s\n", txn.Hash)
		break // Stop after first for example
	}

	// Use iterator utilities
	txns := client.TransactionsIter(ctx, nil)

	// Filter successful transactions
	successful := iter.Filter(txns, func(t *aptos.Transaction) bool {
		return t.Success
	})

	// Take first 5
	first5, _ := iter.CollectN(successful, 5)
	fmt.Printf("Got %d transactions\n", len(first5))
}

func Example_errorHandling() {
	ctx := context.Background()
	client, _ := aptos.NewClient(aptos.Devnet)

	address := aptos.MustParseAddress("0xnonexistent")

	_, err := client.Account(ctx, address)
	if err != nil {
		// Check for specific error types
		if aptos.IsNotFound(err) {
			fmt.Println("Account not found")
		} else if aptos.IsRateLimited(err) {
			fmt.Println("Rate limited")
		} else if aptos.IsTimeout(err) {
			fmt.Println("Request timed out")
		}
	}
}

func Example_parseAddress() {
	// Parse addresses in various formats

	// Full address
	addr1, _ := aptos.ParseAddress("0x0000000000000000000000000000000000000000000000000000000000000001")

	// Short form
	addr2, _ := aptos.ParseAddress("0x1")

	// Without 0x prefix
	addr3, _ := aptos.ParseAddress("1")

	// All equal to AccountOne
	fmt.Printf("All equal: %v\n", addr1 == addr2 && addr2 == addr3)

	// Use MustParseAddress for compile-time constants
	coreAddress := aptos.MustParseAddress("0x1")
	fmt.Printf("Core framework: %s\n", coreAddress)
}

func Example_typeTag() {
	// Create type tags for Move types

	// Parse from string
	coinType, _ := aptos.ParseTypeTag("0x1::aptos_coin::AptosCoin")
	fmt.Printf("Coin type: %s\n", coinType.String())

	// Use predefined
	aptCoin := aptos.AptosCoinTypeTag
	fmt.Printf("APT coin: %s\n", aptCoin.String())

	// Create struct tag
	structTag := &aptos.StructTag{
		Address: aptos.AccountOne,
		Module:  "coin",
		Name:    "CoinStore",
		TypeParams: []aptos.TypeTag{
			aptos.AptosCoinTypeTag,
		},
	}
	fmt.Printf("Struct: %s\n", structTag.String())
}
