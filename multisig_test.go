package aptos

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// test_Multisig is an integration test helper that requires a running local network.
// It is invoked by integration test suites, not directly.
var _ = test_Multisig //nolint:unused

func test_Multisig(t *testing.T, _ CreateSingleSignerPayload) {
	t.Helper()
	client, _ := createTestClient()

	recipient, err := NewEd25519Account()
	if err != nil {
		panic("Failed to create recipient " + err.Error())
	}
	accounts := generateOwnerAccounts()

	// Fund owners
	fundAccounts(client, []*AccountAddress{
		&accounts[0].Address,
		&accounts[1].Address,
		&accounts[2].Address,
	})
	println(accounts[0].Address.String())
	println(accounts[1].Address.String())
	println(accounts[2].Address.String())

	multisigAddress := setUpMultisig(client, accounts)

	// Fund multisig
	println("Funding the multisig account...")
	fundAccounts(client, []*AccountAddress{multisigAddress})

	// Run through flow with the full payload
	println("Creating a multisig transaction to transfer coins...")
	payload := createMultisigTransferTransaction(client, accounts[1], *multisigAddress, recipient.Address)
	println("Owner 1 rejects but owner 3 approves.")
	rejectAndApprove(client, *multisigAddress, accounts[0], accounts[2], 1)
	println("Owner 2 can now execute the transfer transaction as it already has 2 approvals (from owners 2 and 3).")
	executeTransaction(client, *multisigAddress, accounts[1], payload)

	// Check balance of recipient, should be 1_000_000
	assertBalance(client, recipient.Address, 1_000_000)
	println("Recipient's balance after transfer 1000000")

	// Run through flow with the full just a transaction hash
	println("Creating another multisig transaction using payload hash...")
	payload = createMultisigTransferTransactionWithHash(client, accounts[1], *multisigAddress, recipient.Address)
	println("Owner 3 rejects but owner 1 approves.")
	rejectAndApprove(client, *multisigAddress, accounts[2], accounts[0], 2)
	println("Owner 1 can now execute the transfer with hash transaction as it already has 2 approvals (from owners 1 and 2).")
	executeTransaction(client, *multisigAddress, accounts[0], payload)

	// Check balance of recipient, should be 2_000_000
	assertBalance(client, recipient.Address, 2_000_000)
	println("Recipient's balance after transfer 2000000")

	// Add an owner
	println("Adding an owner to the multisig account...")
	payload = addOwnerTransaction(client, accounts[1], *multisigAddress, recipient.Address)
	println("Owner 1 rejects but owner 3 approves.")
	rejectAndApprove(client, *multisigAddress, accounts[0], accounts[2], 3)
	println("Owner 2 can now execute the adding an owner transaction as it already has 2 approvals (from owners 1 and 3).")
	userTxn := executeTransaction(client, *multisigAddress, accounts[1], payload)

	time.Sleep(time.Second)

	_, owners := multisigResource(client, multisigAddress)
	println("Number of Owners:", len(owners))

	if len(owners) != 4 {
		panic(fmt.Sprintf("Expected 4 owners got %d txn %s", len(owners), userTxn.Hash))
	}

	// Remove an owner
	println("Removing an owner from the multisig account...")
	payload = removeOwnerTransaction(client, accounts[1], *multisigAddress, recipient.Address)
	println("Owner 1 rejects but owner 3 approves.")
	rejectAndApprove(client, *multisigAddress, accounts[0], accounts[2], 4)
	println("Owner 2 can now execute the removing an owner transaction as it already has 2 approvals (from owners 2 and 3).")
	userTxn = executeTransaction(client, *multisigAddress, accounts[1], payload)

	_, owners = multisigResource(client, multisigAddress)
	println("Number of Owners:", len(owners))
	if len(owners) != 3 {
		panic(fmt.Sprintf("Expected 3 owners got %d txn %s", len(owners), userTxn.Hash))
	}

	// Change threshold
	println("Changing the signature threshold to 3-of-3...")
	payload = changeThresholdTransaction(client, accounts[1], *multisigAddress, 3)
	println("Owner 1 rejects but owner 3 approves.")
	rejectAndApprove(client, *multisigAddress, accounts[0], accounts[2], 5)
	println("Owner 2 can now execute the change signature threshold transaction as it already has 2 approvals (from owners 2 and 3).")
	userTxn = executeTransaction(client, *multisigAddress, accounts[1], payload)

	threshold, _ := multisigResource(client, multisigAddress)
	println("Signature Threshold: ", threshold)
	if threshold != 3 {
		panic(fmt.Sprintf("Expected 3-of-3 owners got %d-of-3 txn %s", threshold, userTxn.Hash))
	}
	println("Multisig setup and transactions complete.")
}

func assertBalance(client *Client, address AccountAddress, expectedBalance uint64) {
	amount, err := client.AccountAPTBalance(address)
	if err != nil {
		panic("failed to get balance: " + err.Error())
	}
	if amount != expectedBalance {
		panic(fmt.Sprintf("balance mismatch, got %d instead of %d", amount, expectedBalance))
	}
}

func generateOwnerAccounts() []*Account {
	accounts := make([]*Account, 3)
	for i := range 3 {
		account, err := NewEd25519Account()
		if err != nil {
			panic("Failed to create account " + err.Error())
		}
		accounts[i] = account
	}
	return accounts
}

func fundAccounts(client *Client, accounts []*AccountAddress) {
	for _, account := range accounts {
		err := client.Fund(*account, 1_000_000_000)
		if err != nil {
			panic("Failed to fund account " + err.Error())
		}
	}
}

func setUpMultisig(client *Client, accounts []*Account) *AccountAddress {
	println("Setting up a 2-of-3 multisig account...")

	// Step 1: Set up a 2-of-3 multisig account
	// ===========================================================================================
	// Get the next multisig account address. This will be the same as the account address of the multisig account we'll
	// be creating.
	multisigAddress, err := client.FetchNextMultisigAddress(accounts[0].Address)
	if err != nil {
		panic("Failed to fetch next multisig address: " + err.Error())
	}

	// Create the multisig account with 3 owners and a signature threshold of 2.
	createMultisig(client, accounts[0], []AccountAddress{accounts[1].Address, accounts[2].Address})
	println("Multisig Account Address:", multisigAddress.String())

	// should be 2
	threshold, owners := multisigResource(client, multisigAddress)
	println("Signature Threshold:", threshold)

	// should be 3
	println("Number of Owners:", len(owners))

	return multisigAddress
}

func createMultisig(client *Client, account *Account, additionalAddresses []AccountAddress) {
	// TODO: Ideally, this would not be done, and the payload function would take an array of items to serialize
	metadataValue, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		bcs.SerializeSequenceWithFunction([]string{"example"}, ser, func(ser *bcs.Serializer, item string) {
			ser.WriteString(item)
		})
	})
	if err != nil {
		panic("Failed to serialize metadata value" + err.Error())
	}
	payload, err := MultisigCreateAccountPayload(
		2,                   // Required signers
		additionalAddresses, // Other owners
		[]string{"example"}, // Metadata keys
		metadataValue,       // Metadata values
	)
	if err != nil {
		panic("Failed to create multisig account payload " + err.Error())
	}

	submitAndWait(client, account, payload)
}

// TODO: This should be a view function
func multisigResource(client *Client, multisigAddress *AccountAddress) (uint64, []any) {
	resource, err := client.AccountResource(*multisigAddress, "0x1::multisig_account::MultisigAccount")
	if err != nil {
		panic("Failed to get resource for multisig account: " + err.Error())
	}
	// TODO: Add JSON types
	resourceData, ok := resource["data"].(map[string]any)
	if !ok {
		panic("Failed to cast resource data")
	}

	numSigsRequiredStr, ok := resourceData["num_signatures_required"].(string)
	if !ok {
		panic("Failed to cast num_signatures_required")
	}

	numSigsRequired, err := StrToUint64(numSigsRequiredStr)
	if err != nil {
		panic("Failed to convert string to u64: " + err.Error())
	}
	ownersArray, ok := resourceData["owners"].([]any)
	if !ok {
		panic("Failed to cast owners array")
	}

	return numSigsRequired, ownersArray
}

func createMultisigTransferTransaction(client *Client, sender *Account, multisigAddress AccountAddress, recipient AccountAddress) *MultisigTransactionPayload {
	entryFunctionPayload, err := CoinTransferPayload(nil, recipient, 1_000_000)
	if err != nil {
		panic("Failed to create payload for multisig transfer: " + err.Error())
	}

	multisigPayload := &MultisigTransactionPayload{
		Variant: MultisigTransactionPayloadVariantEntryFunction,
		Payload: entryFunctionPayload,
	}

	createTransactionPayload, err := MultisigCreateTransactionPayload(multisigAddress, multisigPayload)
	if err != nil {
		panic("Failed to create payload to create transaction for multisig transfer: " + err.Error())
	}

	// TODO: add simulation of multisig payload
	submitAndWait(client, sender, createTransactionPayload)
	return multisigPayload
}

func createMultisigTransferTransactionWithHash(client *Client, sender *Account, multisigAddress AccountAddress, recipient AccountAddress) *MultisigTransactionPayload {
	entryFunctionPayload, err := CoinTransferPayload(nil, recipient, 1_000_000)
	if err != nil {
		panic("Failed to create payload for multisig transfer: " + err.Error())
	}

	return createTransactionPayloadCommon(client, sender, multisigAddress, entryFunctionPayload)
}

func addOwnerTransaction(client *Client, sender *Account, multisigAddress AccountAddress, newOwner AccountAddress) *MultisigTransactionPayload {
	entryFunctionPayload := MultisigAddOwnerPayload(newOwner)
	return createTransactionPayloadCommon(client, sender, multisigAddress, entryFunctionPayload)
}

func removeOwnerTransaction(client *Client, sender *Account, multisigAddress AccountAddress, removedOwner AccountAddress) *MultisigTransactionPayload {
	entryFunctionPayload := MultisigRemoveOwnerPayload(removedOwner)
	return createTransactionPayloadCommon(client, sender, multisigAddress, entryFunctionPayload)
}

func changeThresholdTransaction(client *Client, sender *Account, multisigAddress AccountAddress, numSignaturesRequired uint64) *MultisigTransactionPayload {
	entryFunctionPayload, err := MultisigChangeThresholdPayload(numSignaturesRequired)
	if err != nil {
		panic("Failed to create payload for multisig remove owner: " + err.Error())
	}

	return createTransactionPayloadCommon(client, sender, multisigAddress, entryFunctionPayload)
}

func createTransactionPayloadCommon(client *Client, sender *Account, multisigAddress AccountAddress, entryFunctionPayload *EntryFunction) *MultisigTransactionPayload {
	multisigPayload := &MultisigTransactionPayload{
		Variant: MultisigTransactionPayloadVariantEntryFunction,
		Payload: entryFunctionPayload,
	}

	createTransactionPayload, err := MultisigCreateTransactionPayloadWithHash(multisigAddress, multisigPayload)
	if err != nil {
		panic("Failed to create payload to create transaction for multisig: " + err.Error())
	}

	// TODO: add simulation of multisig payload
	submitAndWait(client, sender, createTransactionPayload)
	return multisigPayload
}

func rejectAndApprove(client *Client, multisigAddress AccountAddress, rejector *Account, approver *Account, transactionId uint64) {
	rejectPayload, err := MultisigRejectPayload(multisigAddress, transactionId)
	if err != nil {
		panic("Failed to build reject transaction payload: " + err.Error())
	}
	submitAndWait(client, rejector, rejectPayload)

	approvePayload, err := MultisigApprovePayload(multisigAddress, transactionId)
	if err != nil {
		panic("Failed to build approve transaction payload: " + err.Error())
	}

	submitAndWait(client, approver, approvePayload)
}

func executeTransaction(client *Client, multisigAddress AccountAddress, sender *Account, payload *MultisigTransactionPayload) *api.UserTransaction {
	executionPayload := &Multisig{
		MultisigAddress: multisigAddress,
		Payload:         payload,
	}
	return submitAndWait(client, sender, executionPayload)
}

func submitAndWait(client *Client, sender *Account, payload TransactionPayloadImpl) *api.UserTransaction {
	submitResponse, err := client.BuildSignAndSubmitTransaction(sender, TransactionPayload{Payload: payload})
	if err != nil {
		panic("Failed to submit transaction: " + err.Error())
	}

	txn, err := client.WaitForTransaction(submitResponse.Hash)
	if err != nil {
		panic("Failed to wait for transaction: " + err.Error())
	}

	if !txn.Success {
		panic("Transaction failed: " + submitResponse.Hash)
	}

	// Now check that there's no event for failed multisig
	// TODO: make this a function on the user transaction
	for _, event := range txn.Events {
		if event.Type == "0x1::multisig_account::TransactionExecutionFailed" {
			eventStr, marshalErr := json.Marshal(event)
			if marshalErr != nil {
				panic("Failed to marshal event: " + marshalErr.Error())
			}
			panic(fmt.Sprintf("Multisig transaction failed. details: %s", eventStr))
		}
	}

	return txn
}
