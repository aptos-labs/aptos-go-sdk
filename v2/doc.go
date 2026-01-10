// Package aptos provides a fully idiomatic Go SDK for interacting with the Aptos blockchain.
//
// This is version 2 of the Aptos Go SDK, featuring:
//   - context.Context on all client methods for cancellation and timeouts
//   - Functional options pattern for configuration
//   - Comprehensive error types with sentinel errors and rich context
//   - Go 1.23 iterators for streaming data
//   - Builder patterns for complex transaction construction
//   - Full support for Keyless authentication and Aptos Names Service (ANS)
//
// # Quick Start
//
// Create a client and interact with the Aptos blockchain:
//
//	client, err := aptos.NewClient(aptos.Testnet)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get account info
//	ctx := context.Background()
//	info, err := client.Account(ctx, myAddress)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Transaction Building
//
// Use the builder pattern for constructing transactions:
//
//	result, err := aptos.NewTransaction(client, sender.Address()).
//	    EntryFunction(aptos.ModuleID{Address: aptos.AccountOne, Name: "aptos_account"}, "transfer", nil, []any{recipient, amount}).
//	    MaxGas(10000).
//	    SignAndSubmit(ctx, sender)
//
// # Keyless Authentication
//
// Authenticate users with OpenID Connect providers:
//
//	account, err := keyless.DeriveAccount(ctx, keyless.DefaultConfig(), jwt, ephemeralKey)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Aptos Names Service
//
// Resolve human-readable names to addresses:
//
//	ansClient := ans.New(client)
//	addr, err := ansClient.ResolveAddress(ctx, "alice.apt")
//
// # Testing
//
// Use the testutil package for unit testing:
//
//	fake := testutil.NewFakeClient()
//	fake.AccountFunc = func(ctx context.Context, addr AccountAddress) (*AccountInfo, error) {
//	    return &AccountInfo{SequenceNumber: 42}, nil
//	}
//
// For more information, see the package documentation for each sub-package.
package aptos
