// sponsored_transaction demonstrates fee payer (sponsored) transactions.
//
// This example shows the complete flow of sponsored transactions where
// a separate account pays gas fees on behalf of the sender.
//
// In a sponsored transaction:
//   - The sender initiates a transaction
//   - The fee payer (sponsor) agrees to pay the gas fees
//   - Both parties sign the transaction
//   - The transaction is submitted to the network
//
// This enables gasless transactions for users who don't have APT for gas.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := aptos.NewClient(aptos.Testnet)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Aptos Go SDK v2 - Sponsored Transaction Example ===")
	fmt.Println()

	// ============================================================
	// Part 1: Understanding Sponsored Transactions
	// ============================================================
	fmt.Println("Part 1: Understanding Sponsored Transactions")
	fmt.Println("---------------------------------------------")
	fmt.Println()

	fmt.Println("In a sponsored (fee payer) transaction:")
	fmt.Println("  - Sender: Initiates the action (e.g., transfers tokens)")
	fmt.Println("  - Fee Payer: Pays the gas fees for the transaction")
	fmt.Println("  - Receiver: (Optional) Receives the action's result")
	fmt.Println()

	fmt.Println("Benefits:")
	fmt.Println("  - Users don't need APT for gas fees")
	fmt.Println("  - Better UX for new users onboarding")
	fmt.Println("  - dApps can subsidize user transactions")
	fmt.Println("  - Enable true gasless meta-transactions")
	fmt.Println()

	// ============================================================
	// Part 2: The Signing Flow
	// ============================================================
	fmt.Println("Part 2: The Signing Flow")
	fmt.Println("------------------------")
	fmt.Println()

	fmt.Println("1. Build transaction with fee payer address")
	fmt.Println("   - Sender specifies the intent")
	fmt.Println("   - Fee payer address is included")
	fmt.Println()

	fmt.Println("2. Create the signing message")
	fmt.Println("   - Different from regular transactions")
	fmt.Println("   - Uses RawTransactionWithData salt")
	fmt.Println("   - Includes fee payer in the message")
	fmt.Println()

	fmt.Println("3. Both parties sign")
	fmt.Println("   - Sender signs: confirms the intent")
	fmt.Println("   - Fee payer signs: agrees to pay gas")
	fmt.Println()

	fmt.Println("4. Combine signatures and submit")
	fmt.Println("   - FeePayerAuthenticator contains both signatures")
	fmt.Println("   - Transaction is submitted to the network")
	fmt.Println()

	// ============================================================
	// Part 3: Code Example
	// ============================================================
	fmt.Println("Part 3: Code Example")
	fmt.Println("--------------------")
	fmt.Println()

	// Demo addresses (in a real app, these come from wallets)
	senderAddr := aptos.MustParseAddress("0x1234")
	receiverAddr := aptos.MustParseAddress("0x5678")
	feePayerAddr := aptos.MustParseAddress("0xabcd")

	fmt.Printf("Sender (Alice):    %s\n", senderAddr.String())
	fmt.Printf("Receiver (Bob):    %s\n", receiverAddr.String())
	fmt.Printf("Fee Payer (Sponsor): %s\n", feePayerAddr.String())
	fmt.Println()

	// Create the transfer payload
	transferAmount := uint64(1000)
	payload := &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "aptos_account"},
		Function: "transfer",
		TypeArgs: nil,
		Args:     []any{receiverAddr, transferAmount},
	}

	fmt.Printf("Transfer: %d octas from Alice to Bob\n", transferAmount)
	fmt.Printf("Gas paid by: Sponsor\n")
	fmt.Println()

	// Build the transaction with fee payer option
	rawTxn, err := client.BuildTransaction(ctx, senderAddr, payload,
		aptos.WithFeePayer(feePayerAddr),
		aptos.WithGasEstimation(),
	)
	if err != nil {
		log.Fatalf("Failed to build transaction: %v", err)
	}

	fmt.Println("Transaction built successfully:")
	fmt.Printf("  Sender: %s\n", rawTxn.Sender.String())
	fmt.Printf("  Sequence Number: %d\n", rawTxn.SequenceNumber)
	fmt.Printf("  Max Gas: %d\n", rawTxn.MaxGasAmount)
	fmt.Printf("  Gas Price: %d\n", rawTxn.GasUnitPrice)
	fmt.Println()

	// ============================================================
	// Part 4: Complete Signing Example (Conceptual)
	// ============================================================
	fmt.Println("Part 4: Complete Signing Example")
	fmt.Println("---------------------------------")
	fmt.Println()

	fmt.Println("// Step 1: Create the fee payer transaction for signing")
	fmt.Println("feePayerTxn := &aptos.FeePayerTransaction{")
	fmt.Println("    RawTxn:           rawTxn,")
	fmt.Println("    SecondarySigners: []aptos.AccountAddress{}, // empty for simple transfer")
	fmt.Println("    FeePayer:         feePayerAddr,")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("// Step 2: Get the signing message (includes fee payer)")
	fmt.Println("signingMessage, _ := feePayerTxn.SigningMessage()")
	fmt.Println("")
	fmt.Println("// Step 3: Sign with sender (Alice)")
	fmt.Println("senderAuth, _ := aliceSigner.Sign(signingMessage)")
	fmt.Println("")
	fmt.Println("// Step 4: Sign with fee payer (Sponsor)")
	fmt.Println("feePayerAuth, _ := sponsorSigner.Sign(signingMessage)")
	fmt.Println("")
	fmt.Println("// Step 5: Create the signed transaction")
	fmt.Println("signedTxn, _ := aptos.NewFeePayerSignedTransaction(")
	fmt.Println("    rawTxn,")
	fmt.Println("    senderAuth,")
	fmt.Println("    []aptos.AccountAddress{}, // no secondary signers")
	fmt.Println("    []*aptos.AccountAuthenticator{}, // no secondary auths")
	fmt.Println("    feePayerAddr,")
	fmt.Println("    feePayerAuth,")
	fmt.Println(")")
	fmt.Println("")
	fmt.Println("// Step 6: Submit to the network")
	fmt.Println("result, _ := client.SubmitTransaction(ctx, signedTxn)")
	fmt.Println("fmt.Printf(\"Transaction hash: \"+result.Hash)")
	fmt.Println()

	// ============================================================
	// Part 5: Using the Helper Function
	// ============================================================
	fmt.Println("Part 5: Using the Helper Function")
	fmt.Println("----------------------------------")
	fmt.Println()

	fmt.Println("// The SDK provides a helper function for simpler signing:")
	fmt.Println("signedTxn, err := aptos.SignFeePayerTransaction(")
	fmt.Println("    aliceSigner,     // sender signer")
	fmt.Println("    sponsorSigner,   // fee payer signer")
	fmt.Println("    rawTxn,          // the raw transaction")
	fmt.Println("    sponsorAddr,     // fee payer address")
	fmt.Println("    nil,             // no secondary signers")
	fmt.Println("    nil,             // no secondary addresses")
	fmt.Println(")")
	fmt.Println("if err != nil {")
	fmt.Println("    log.Fatalf(\"Failed to sign\", err)")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("result, err := client.SubmitTransaction(ctx, signedTxn)")
	fmt.Println()

	fmt.Println("=== End of Sponsored Transaction Example ===")
}
