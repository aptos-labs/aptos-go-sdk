package aptos

import (
	"strings"
	"sync"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

const (
	singleSignerScript = "a11ceb0b060000000701000402040a030e0c041a04051e20073e30086e2000000001010204010001000308000104030401000105050601000002010203060c0305010b0001080101080102060c03010b0001090002050b00010900000a6170746f735f636f696e04636f696e04436f696e094170746f73436f696e087769746864726177076465706f7369740000000000000000000000000000000000000000000000000000000000000001000001080b000b0138000c030b020b03380102"
	fundAmount         = 100_000_000
	vmStatusSuccess    = "Executed successfully"
)

var TestSigners map[string]func() (TransactionSigner, error)

func init() {
	TestSigners = make(map[string]func() (TransactionSigner, error))
	TestSigners["Legacy Ed25519"] = func() (TransactionSigner, error) {
		signer, err := NewEd25519Account()
		return any(signer).(TransactionSigner), err
	}
	TestSigners["Single Sender Ed25519"] = func() (TransactionSigner, error) {
		signer, err := NewEd25519SingleSenderAccount()
		return any(signer).(TransactionSigner), err
	}
	TestSigners["Single Sender Secp256k1"] = func() (TransactionSigner, error) {
		signer, err := NewSecp256k1Account()
		return any(signer).(TransactionSigner), err
	}
	TestSigners["MultiKey"] = func() (TransactionSigner, error) {
		signer, err := NewMultiKeyTestSigner(3, 2)
		return any(signer).(TransactionSigner), err
	}
	/* TODO: MultiEd25519 is not supported ATM
	TestSigners["MultiEd25519"] = func() (TransactionSigner, error) {
		signer, err := NewMultiEd25519Signer(3, 2)
		return any(signer).(TransactionSigner), err
	}
	*/
}

func TestNamedConfig(t *testing.T) {
	names := []string{"mainnet", "devnet", "testnet", "localnet"}
	for _, name := range names {
		assert.Equal(t, name, NamedNetworks[name].Name)
	}
}

func TestAptosClientHeaderValue(t *testing.T) {
	assert.Greater(t, len(ClientHeaderValue), 0)
	assert.NotEqual(t, "aptos-go-sdk/unk", ClientHeaderValue)
}

func Test_EntryFunctionFlow(t *testing.T) {
	testAllSigners(t, func(t *testing.T, name string, createSigner func() (TransactionSigner, error)) {
		println("Entry function: " + name)
		testTransaction(t, createSigner, submitEntryFunction)
	})
}

func Test_EntryFunctionSimulation(t *testing.T) {
	testAllSigners(t, func(t *testing.T, name string, createSigner func() (TransactionSigner, error)) {
		println("Entry Function simulation: " + name)
		testTransactionSimulation(t, createSigner, submitEntryFunction)
	})
}

func Test_ScriptFlow(t *testing.T) {
	testAllSigners(t, func(t *testing.T, name string, createSigner func() (TransactionSigner, error)) {
		println("Script: " + name)
		testTransaction(t, createSigner, submitScript)
	})
}

func Test_ScriptSimulation(t *testing.T) {
	testAllSigners(t, func(t *testing.T, name string, createSigner func() (TransactionSigner, error)) {
		println("Script simulation: " + name)
		testTransactionSimulation(t, createSigner, submitScript)
	})
}

func Test_FeePayerFlow_EntryFunction(t *testing.T) {
	testAllSigners(t, func(t *testing.T, name string, createSigner func() (TransactionSigner, error)) {
		println("Entry function fee payer: " + name)
		testFeePayerTransaction(t, createSigner, submitEntryFunctionFeePayer)
	})
}

func Test_FeePayerFlow_Script(t *testing.T) {
	testAllSigners(t, func(t *testing.T, name string, createSigner func() (TransactionSigner, error)) {
		println("Script fee payer: " + name)
		testFeePayerTransaction(t, createSigner, submitScriptFeePayer)
	})
}

func setupIntegrationTest(t *testing.T, createAccount func() (TransactionSigner, error)) (*Client, TransactionSigner) {
	// All of these run against localnet
	if testing.Short() {
		t.Skip("integration test expects network connection to localnet")
	}
	// Create a client
	client, err := createTestClient()
	assert.NoError(t, err)

	// Verify chain id retrieval works
	chainId, err := client.GetChainId()
	assert.NoError(t, err)
	if testConfig == DevnetConfig {
		assert.Greater(t, chainId, LocalnetConfig.ChainId)
	} else {
		assert.Equal(t, testConfig.ChainId, chainId)
	}

	// Verify gas estimation works
	_, err = client.EstimateGasPrice()
	assert.NoError(t, err)

	// Create an account
	account, err := createAccount()
	assert.NoError(t, err)

	// Fund the account with 1 APT
	err = client.Fund(account.AccountAddress(), fundAmount)
	assert.NoError(t, err)

	return client, account
}

func testTransaction(t *testing.T, createAccount func() (TransactionSigner, error), buildTransaction func(t *testing.T, client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error)) {
	client, account := setupIntegrationTest(t, createAccount)

	// Build transaction
	rawTxn, err := buildTransaction(t, client, account)
	assert.NoError(t, err)

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(account)
	assert.NoError(t, err)

	// Send transaction
	result, err := client.SubmitTransaction(signedTxn)
	assert.NoError(t, err)

	hash := result.Hash

	// Wait for the transaction
	_, err = client.WaitForTransaction(hash)
	assert.NoError(t, err)

	// Read transaction by hash
	txn, err := client.TransactionByHash(hash)
	assert.NoError(t, err)

	// Read transaction by version
	userTxn, _ := txn.Inner.(*api.UserTransaction)
	version := userTxn.Version

	// Load the transaction again
	txnByVersion, err := client.TransactionByVersion(version)
	assert.NoError(t, err)

	// Assert that both are the same
	assert.Equal(t, txn, txnByVersion)
}

func testFeePayerTransaction(t *testing.T, createAccount func() (TransactionSigner, error), buildTransaction func(t *testing.T, client *Client, sender TransactionSigner) (*RawTransactionWithData, error)) {
	client, account := setupIntegrationTest(t, createAccount)

	// Create sponsor
	sponsor, err := createAccount()
	assert.NoError(t, err)

	// Fund sponsor
	err = client.Fund(sponsor.AccountAddress(), fundAmount)
	assert.NoError(t, err)

	// Build transaction
	rawTxn, err := buildTransaction(t, client, account)
	assert.NoError(t, err)

	// Sign as user
	userAuth, err := rawTxn.Sign(account)
	assert.NoError(t, err)

	// Sign as sponsor
	rawTxn.SetFeePayer(sponsor.AccountAddress())
	sponsorAuth, err := rawTxn.Sign(sponsor)
	assert.NoError(t, err)

	// Sign transaction
	signedTxn, ok := rawTxn.ToFeePayerSignedTransaction(
		userAuth,
		sponsorAuth,
		[]crypto.AccountAuthenticator{},
	)
	assert.True(t, ok)

	// Send transaction
	result, err := client.SubmitTransaction(signedTxn)
	assert.NoError(t, err)

	hash := result.Hash

	// Wait for the transaction
	_, err = client.WaitForTransaction(hash)
	assert.NoError(t, err)

	// Read transaction by hash
	txn, err := client.TransactionByHash(hash)
	assert.NoError(t, err)

	// Read transaction by version
	userTxn, _ := txn.Inner.(*api.UserTransaction)
	version := userTxn.Version

	// Load the transaction again
	txnByVersion, err := client.TransactionByVersion(version)
	assert.NoError(t, err)

	// Assert that both are the same
	assert.Equal(t, txn, txnByVersion)
}

func testTransactionSimulation(t *testing.T, createAccount func() (TransactionSigner, error), buildTransaction func(t *testing.T, client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error)) {
	client, account := setupIntegrationTest(t, createAccount)

	// Simulate transaction (no options)
	rawTxn, err := buildTransaction(t, client, account)
	assert.NoError(t, err)
	simulatedTxn, err := client.SimulateTransaction(rawTxn, account)
	switch account.(type) {
	case *MultiKeyTestSigner:
		// multikey simulation currently not supported
		assert.Error(t, err)
		assert.ErrorContains(t, err, "currently unsupported sender derivation scheme")
		return // skip rest of the tests
	default:
		assert.NoError(t, err)
		assert.Equal(t, true, simulatedTxn[0].Success)
		assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
		assert.Greater(t, simulatedTxn[0].GasUsed, uint64(0))
	}

	// simulate transaction (estimate gas unit price)
	rawTxnZeroGasUnitPrice, err := buildTransaction(t, client, account, GasUnitPrice(0))
	assert.NoError(t, err)
	simulatedTxn, err = client.SimulateTransaction(rawTxnZeroGasUnitPrice, account, EstimateGasUnitPrice(true))
	assert.NoError(t, err)
	assert.Equal(t, true, simulatedTxn[0].Success)
	assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
	estimatedGasUnitPrice := simulatedTxn[0].GasUnitPrice
	assert.Greater(t, estimatedGasUnitPrice, uint64(0))

	// simulate transaction (estimate max gas amount)
	rawTxnZeroMaxGasAmount, err := buildTransaction(t, client, account, MaxGasAmount(0))
	assert.NoError(t, err)
	simulatedTxn, err = client.SimulateTransaction(rawTxnZeroMaxGasAmount, account, EstimateMaxGasAmount(true))
	assert.NoError(t, err)
	assert.Equal(t, true, simulatedTxn[0].Success)
	assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
	assert.Greater(t, simulatedTxn[0].MaxGasAmount, uint64(0))

	// simulate transaction (estimate prioritized gas unit price and max gas amount)
	rawTxnZeroGasConfig, err := buildTransaction(t, client, account, GasUnitPrice(0), MaxGasAmount(0))
	assert.NoError(t, err)
	simulatedTxn, err = client.SimulateTransaction(rawTxnZeroGasConfig, account, EstimatePrioritizedGasUnitPrice(true), EstimateMaxGasAmount(true))
	assert.NoError(t, err)
	assert.Equal(t, true, simulatedTxn[0].Success)
	assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
	estimatedGasUnitPrice = simulatedTxn[0].GasUnitPrice
	assert.Greater(t, estimatedGasUnitPrice, uint64(0))
	assert.Greater(t, simulatedTxn[0].MaxGasAmount, uint64(0))
}

func TestAPTTransferTransaction(t *testing.T) {
	sender, err := NewEd25519Account()
	assert.NoError(t, err)
	dest, err := NewEd25519Account()
	assert.NoError(t, err)

	client, err := createTestClient()
	assert.NoError(t, err)
	signedTxn, err := APTTransferTransaction(client, sender, dest.Address, 1337, MaxGasAmount(123123), GasUnitPrice(111), ExpirationSeconds(42), ChainIdOption(71), SequenceNumber(31337))
	assert.NoError(t, err)
	assert.NotNil(t, signedTxn)

	// use defaults for: max gas amount, gas unit price
	signedTxn, err = APTTransferTransaction(client, sender, dest.Address, 1337, ExpirationSeconds(42), ChainIdOption(71), SequenceNumber(31337))
	assert.NoError(t, err)
	assert.NotNil(t, signedTxn)
}

func Test_Indexer(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)

	// Fund account
	account, err := NewEd25519Account()
	assert.NoError(t, err)

	err = client.Fund(account.AccountAddress(), 10)
	assert.NoError(t, err)

	balance, err := client.AccountAPTBalance(account.AccountAddress())
	assert.NoError(t, err)

	// TODO: May need to wait on indexer for this one
	coins, err := client.GetCoinBalances(account.AccountAddress())
	assert.NoError(t, err)
	switch len(coins) {
	case 0:
	// TODO we need to wait on the indexer, we'll skip for now
	case 1:
		expectedBalance := CoinBalance{
			CoinType: "0x1::aptos_coin::AptosCoin",
			Amount:   balance,
		}

		assert.Equal(t, expectedBalance, coins[0])
	default:
		panic("Unexpected amount of coins")
	}

	// Get current version
	status, err := client.GetProcessorStatus("default_processor")
	assert.NoError(t, err)
	// TODO: When we have waiting on indexer, we can add this check to be more accurate
	assert.GreaterOrEqual(t, status, uint64(0))
}

func Test_Block(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)
	info, err := client.Info()
	assert.NoError(t, err)

	// TODO: I need to add hardcoded testing sets for these conversions
	numToCheck := uint64(10)
	blockHeight := info.BlockHeight()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(int(numToCheck))

	for i := uint64(0); i < numToCheck; i++ {
		go func() {
			blockNumber := blockHeight - i
			println("BLOCK:", blockNumber)
			blockByHeight, err := client.BlockByHeight(blockNumber, true)
			assert.NoError(t, err)

			assert.Equal(t, blockNumber, blockByHeight.BlockHeight)

			// Block should always be last - first + 1 (since they would be 1 if they're the same (inclusive)
			assert.Equal(t, 1+blockByHeight.LastVersion-blockByHeight.FirstVersion, uint64(len(blockByHeight.Transactions)))

			// Version should be the same
			blockByVersion, err := client.BlockByVersion(blockByHeight.FirstVersion, true)
			assert.NoError(t, err)

			assert.Equal(t, blockByHeight, blockByVersion)
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
}

func Test_Account(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)
	account, err := client.Account(AccountOne)
	assert.NoError(t, err)
	sequenceNumber, err := account.SequenceNumber()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), sequenceNumber)
	authKey, err := account.AuthenticationKey()
	assert.NoError(t, err)
	assert.Equal(t, AccountOne[:], authKey[:])
}

func Test_Transactions(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)

	start := uint64(1)
	count := uint64(2)
	// Specific 2 should only give 2
	transactions, err := client.Transactions(&start, &count)
	assert.NoError(t, err)
	assert.Len(t, transactions, 2)

	// This will give the latest 2
	transactions, err = client.Transactions(nil, &count)
	assert.NoError(t, err)
	assert.Len(t, transactions, 2)

	// This will give the 25 from 2
	transactions, err = client.Transactions(&start, nil)
	assert.NoError(t, err)
	assert.Len(t, transactions, 25)

	// This will give the latest 25
	transactions, err = client.Transactions(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, transactions, 25)
}

func Test_Info(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)

	info, err := client.Info()
	assert.NoError(t, err)
	assert.Greater(t, info.BlockHeight(), uint64(0))
}

func Test_AccountResources(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)

	resources, err := client.AccountResources(AccountOne)
	assert.NoError(t, err)
	assert.Greater(t, len(resources), 0)

	resourcesBcs, err := client.AccountResourcesBCS(AccountOne)
	assert.NoError(t, err)
	assert.Greater(t, len(resourcesBcs), 0)
}

func Test_Concurrent_Submission(t *testing.T) {
	const numTxns = uint64(10)

	client, err := NewClient(LocalnetConfig)
	assert.NoError(t, err)

	account1, err := NewEd25519Account()
	assert.NoError(t, err)

	err = client.Fund(account1.AccountAddress(), 100_000_000)
	assert.NoError(t, err)

	// start submission goroutine
	payloads := make(chan TransactionSubmissionPayload)
	results := make(chan TransactionSubmissionResponse)
	go client.nodeClient.BuildSignAndSubmitTransactions(account1, payloads, results)

	transferAmount, err := bcs.SerializeU64(100)
	assert.NoError(t, err)

	// Generate transactions
	for i := uint64(0); i < numTxns; i++ {
		payloads <- TransactionSubmissionPayload{
			Id:   i,
			Type: TransactionSubmissionTypeSingle, // TODO: not needed?
			Inner: TransactionPayload{Payload: &EntryFunction{
				Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
				Function: "transfer",
				ArgTypes: []TypeTag{},
				Args:     [][]byte{AccountOne[:], transferAmount},
			}},
		}
	}

	// Start waiting on txns
	// TODO: These final steps should be concurrent rather than serial like this
	waitResults := make(chan ConcResponse[*api.UserTransaction], numTxns)

	// It's interesting, this had to be wrapped in a goroutine to ensure blocking on results don't block
	go func() {
		for response := range results {
			assert.NoError(t, response.Err)

			go fetch[*api.UserTransaction](func() (*api.UserTransaction, error) {
				return client.WaitForTransaction(response.Response.Hash)
			}, waitResults)
		}
	}()

	// Wait on all the results, recording the succeeding ones
	txnMap := make(map[uint64]bool)

	// We could wait on a close, but I'm going to be a little pickier here
	for i := uint64(0); i < numTxns; i++ {
		response := <-waitResults
		assert.NoError(t, response.Err)
		assert.True(t, response.Result.Success)
		txnMap[response.Result.SequenceNumber] = true
	}

	// Check all transactions were successful from [0-numTxns)
	for i := uint64(0); i < numTxns; i++ {
		assert.True(t, txnMap[i])
	}
}

func TestClient_BlockByHeight(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)

	_, err = client.BlockByHeight(1, true)
	assert.NoError(t, err)
}

func TestClient_NodeAPIHealthCheck(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)
	response, err := client.NodeAPIHealthCheck()
	assert.NoError(t, err)
	assert.True(t, strings.Contains(response.Message, "ok"), "Node API health check failed"+response.Message)

	// Now, check node API health check with a time that should never fail
	response, err = client.NodeAPIHealthCheck(10000)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(response.Message, "ok"), "Node API health check failed"+response.Message)

	// Now, check node API health check with a time that should probably fail
	response, err = client.NodeAPIHealthCheck(0)
	assert.Error(t, err)
}

func submitEntryFunction(_ *testing.T, client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error) {
	return APTTransferTransaction(client, sender, AccountOne, 100, options...)
}

func submitScript(t *testing.T, client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error) {
	scriptBytes, err := ParseHex(singleSignerScript)
	assert.NoError(t, err)

	amount := uint64(1)
	dest := AccountOne

	rawTxn, err := client.BuildTransaction(sender.AccountAddress(),
		TransactionPayload{Payload: &Script{
			Code:     scriptBytes,
			ArgTypes: []TypeTag{},
			Args: []ScriptArgument{{
				Variant: ScriptArgumentU64,
				Value:   amount,
			}, {
				Variant: ScriptArgumentAddress,
				Value:   dest,
			}},
		}},
		options...,
	)
	if err != nil {
		return nil, err
	}

	return rawTxn, nil
}

func submitEntryFunctionFeePayer(_ *testing.T, client *Client, sender TransactionSigner) (*RawTransactionWithData, error) {
	payload, err := CoinTransferPayload(&AptosCoinTypeTag, AccountOne, 100)
	if err != nil {
		return nil, err
	}
	return client.BuildTransactionMultiAgent(sender.AccountAddress(),
		TransactionPayload{
			Payload: payload,
		},
		FeePayer(&AccountZero))
}

func submitScriptFeePayer(t *testing.T, client *Client, sender TransactionSigner) (*RawTransactionWithData, error) {
	scriptBytes, err := ParseHex(singleSignerScript)
	assert.NoError(t, err)

	amount := uint64(1)
	dest := AccountOne

	return client.BuildTransactionMultiAgent(sender.AccountAddress(),
		TransactionPayload{
			Payload: &Script{
				Code:     scriptBytes,
				ArgTypes: []TypeTag{},
				Args: []ScriptArgument{{
					Variant: ScriptArgumentU64,
					Value:   amount,
				}, {
					Variant: ScriptArgumentAddress,
					Value:   dest,
				}},
			},
		},
		FeePayer(&AccountZero))
}

func testAllSigners(t *testing.T, runOne func(t *testing.T, name string, createSigner func() (TransactionSigner, error))) {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(len(TestSigners))
	for name, createSigner := range TestSigners {
		go func() {
			runOne(t, name, createSigner)
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
}
