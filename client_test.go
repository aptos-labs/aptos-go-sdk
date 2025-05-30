package aptos

import (
	"sync"
	"testing"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	singleSignerScript = "a11ceb0b060000000701000402040a030e0c041a04051e20073e30086e2000000001010204010001000308000104030401000105050601000002010203060c0305010b0001080101080102060c03010b0001090002050b00010900000a6170746f735f636f696e04636f696e04436f696e094170746f73436f696e087769746864726177076465706f7369740000000000000000000000000000000000000000000000000000000000000001000001080b000b0138000c030b020b03380102"
	fundAmount         = 100_000_000
	vmStatusSuccess    = "Executed successfully"
)

type CreateSigner func() (TransactionSigner, error)

var TestSigners map[string]CreateSigner

type CreateSingleSignerPayload func(client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error)

var TestSingleSignerPayloads map[string]CreateSingleSignerPayload

func init() {
	initSigners()
	initSingleSignerPayloads()
}

func initSigners() {
	TestSigners = make(map[string]CreateSigner)
	TestSigners["Standard Ed25519"] = func() (TransactionSigner, error) {
		signer, err := NewEd25519Account()
		return signer, err
	}
	TestSigners["Single Sender Ed25519"] = func() (TransactionSigner, error) {
		signer, err := NewEd25519SingleSenderAccount()
		return signer, err
	}
	TestSigners["Single Sender Secp256k1"] = func() (TransactionSigner, error) {
		signer, err := NewSecp256k1Account()
		return signer, err
	}
	TestSigners["2-of-3 MultiKey"] = func() (TransactionSigner, error) {
		signer, err := NewMultiKeyTestSigner(3, 2)
		return signer, err
	}
	TestSigners["5-of-32 MultiKey"] = func() (TransactionSigner, error) {
		signer, err := NewMultiKeyTestSigner(32, 5)
		return signer, err
	}
	/* TODO: MultiEd25519 is not supported ATM
	TestSigners["MultiEd25519"] = func() (TransactionSigner, error) {
		signer, err := NewMultiEd25519Signer(3, 2)
		return any(signer).(TransactionSigner), err
	}
	*/
}

func initSingleSignerPayloads() {
	TestSingleSignerPayloads = make(map[string]CreateSingleSignerPayload)
	TestSingleSignerPayloads["Entry Function"] = buildSingleSignerEntryFunction
	TestSingleSignerPayloads["Script"] = buildSingleSignerScript
}

func TestNamedConfig(t *testing.T) {
	t.Parallel()
	names := []string{"mainnet", "devnet", "testnet", "localnet"}
	for _, name := range names {
		assert.Equal(t, name, NamedNetworks[name].Name)
	}
}

func TestAptosClientHeaderValue(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, ClientHeaderValue)
	assert.NotEqual(t, "aptos-go-sdk/unk", ClientHeaderValue)
}

func Test_SingleSignerFlows(t *testing.T) {
	t.Parallel()
	for name, signer := range TestSigners {
		for payloadName, buildSingleSignerPayload := range TestSingleSignerPayloads {
			t.Run(name+" "+payloadName, func(t *testing.T) {
				t.Parallel()
				testTransaction(t, signer, buildSingleSignerPayload)
			})
			t.Run(name+" "+payloadName+" simulation", func(t *testing.T) {
				t.Parallel()
				testTransactionSimulation(t, signer, buildSingleSignerPayload)
			})
		}
	}
}

func setupIntegrationTest(t *testing.T, createAccount CreateSigner) (*Client, TransactionSigner) {
	t.Helper()
	// All of these run against localnet
	if testing.Short() {
		t.Skip("integration test expects network connection to localnet")
	}
	// Create a client
	client, err := createTestClient()
	require.NoError(t, err)

	// Verify chain id retrieval works
	chainId, err := client.GetChainId()
	require.NoError(t, err)
	if testConfig == DevnetConfig {
		assert.Greater(t, chainId, LocalnetConfig.ChainId)
	} else {
		assert.Equal(t, testConfig.ChainId, chainId)
	}

	// Verify gas estimation works
	_, err = client.EstimateGasPrice()
	require.NoError(t, err)

	// Create an account
	account, err := createAccount()
	require.NoError(t, err)

	// Fund the account with 1 APT
	err = client.Fund(account.AccountAddress(), fundAmount)
	require.NoError(t, err)

	// This is messy... but there seems to be some race condition here, I don't have a ton of time to figure it out, so I'm just going to sleep
	time.Sleep(1 * time.Second)
	return client, account
}

func testTransaction(t *testing.T, createAccount CreateSigner, buildTransaction CreateSingleSignerPayload) {
	t.Helper()
	client, account := setupIntegrationTest(t, createAccount)

	// Build transaction
	rawTxn, err := buildTransaction(client, account)
	require.NoError(t, err)

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)

	// Send transaction
	result, err := client.SubmitTransaction(signedTxn)
	require.NoError(t, err)

	hash := result.Hash

	// Wait for the transaction
	_, err = client.WaitForTransaction(hash)
	require.NoError(t, err)

	// Read transaction by hash
	txn, err := client.TransactionByHash(hash)
	require.NoError(t, err)

	// Read transaction by version
	userTxn, _ := txn.Inner.(*api.UserTransaction)
	version := userTxn.Version

	// Load the transaction again
	txnByVersion, err := client.TransactionByVersion(version)
	require.NoError(t, err)

	// Assert that both are the same
	expectedTxn, err := txn.UserTransaction()
	require.NoError(t, err)
	actualTxn, err := txnByVersion.UserTransaction()
	require.NoError(t, err)
	assert.Equal(t, expectedTxn, actualTxn)
}

func testTransactionSimulation(t *testing.T, createAccount CreateSigner, buildTransaction CreateSingleSignerPayload) {
	t.Helper()
	client, account := setupIntegrationTest(t, createAccount)

	// Simulate transaction (no options)
	rawTxn, err := buildTransaction(client, account)
	require.NoError(t, err)
	simulatedTxn, err := client.SimulateTransaction(rawTxn, account)
	switch account.(type) {
	case *MultiKeyTestSigner:
		// multikey simulation currently not supported, skip it for now
		return // skip rest of the tests
	default:
		require.NoError(t, err)
		assert.True(t, simulatedTxn[0].Success)
		assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
		assert.Greater(t, simulatedTxn[0].GasUsed, uint64(1))
	}

	// simulate transaction (estimate gas unit price)
	rawTxnZeroGasUnitPrice, err := buildTransaction(client, account, GasUnitPrice(0))
	require.NoError(t, err)
	simulatedTxn, err = client.SimulateTransaction(rawTxnZeroGasUnitPrice, account, EstimateGasUnitPrice(true))
	require.NoError(t, err)
	assert.True(t, simulatedTxn[0].Success)
	assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
	estimatedGasUnitPrice := simulatedTxn[0].GasUnitPrice
	assert.Greater(t, estimatedGasUnitPrice, uint64(1))

	// simulate transaction (estimate max gas amount)
	rawTxnZeroMaxGasAmount, err := buildTransaction(client, account, MaxGasAmount(0))
	require.NoError(t, err)
	simulatedTxn, err = client.SimulateTransaction(rawTxnZeroMaxGasAmount, account, EstimateMaxGasAmount(true))
	require.NoError(t, err)
	assert.True(t, simulatedTxn[0].Success)
	assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
	assert.Greater(t, simulatedTxn[0].MaxGasAmount, uint64(1))

	// simulate transaction (estimate prioritized gas unit price and max gas amount)
	rawTxnZeroGasConfig, err := buildTransaction(client, account, GasUnitPrice(0), MaxGasAmount(0))
	require.NoError(t, err)
	simulatedTxn, err = client.SimulateTransaction(rawTxnZeroGasConfig, account, EstimatePrioritizedGasUnitPrice(true), EstimateMaxGasAmount(true))
	require.NoError(t, err)
	assert.True(t, simulatedTxn[0].Success)
	assert.Equal(t, vmStatusSuccess, simulatedTxn[0].VmStatus)
	estimatedGasUnitPrice = simulatedTxn[0].GasUnitPrice
	assert.Greater(t, estimatedGasUnitPrice, uint64(1))
	assert.Greater(t, simulatedTxn[0].MaxGasAmount, uint64(1))
}

func TestAPTTransferTransaction(t *testing.T) {
	t.Parallel()
	sender, err := NewEd25519Account()
	require.NoError(t, err)
	dest, err := NewEd25519Account()
	require.NoError(t, err)

	client, err := createTestClient()
	require.NoError(t, err)
	signedTxn, err := APTTransferTransaction(client, sender, dest.Address, 1337, MaxGasAmount(123123), GasUnitPrice(111), ExpirationSeconds(42), ChainIdOption(71), SequenceNumber(31337))
	require.NoError(t, err)
	assert.NotNil(t, signedTxn)

	// use defaults for: max gas amount, gas unit price
	signedTxn, err = APTTransferTransaction(client, sender, dest.Address, 1337, ExpirationSeconds(42), ChainIdOption(71), SequenceNumber(31337))
	require.NoError(t, err)
	assert.NotNil(t, signedTxn)
}

func Test_Indexer(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	// Fund account
	account, err := NewEd25519Account()
	require.NoError(t, err)

	err = client.Fund(account.AccountAddress(), 10)
	require.NoError(t, err)

	balance, err := client.AccountAPTBalance(account.AccountAddress())
	require.NoError(t, err)

	// TODO: May need to wait on indexer for this one
	coins, err := client.GetCoinBalances(account.AccountAddress())
	require.NoError(t, err)
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
	require.NoError(t, err)
	// TODO: When we have waiting on indexer, we can add this check to be more accurate
	assert.GreaterOrEqual(t, status, uint64(1))
}

func Test_Genesis(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	genesis, err := client.BlockByHeight(0, true)
	require.NoError(t, err)

	txn, err := genesis.Transactions[0].GenesisTransaction()
	require.NoError(t, err)

	assert.Equal(t, uint64(0), *txn.TxnVersion())
}

func Test_Block(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)
	info, err := client.Info()
	require.NoError(t, err)

	// TODO: I need to add hardcoded testing sets for these conversions
	numToCheck := uint64(10)
	blockHeight := info.BlockHeight()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(int(numToCheck))

	for i := range numToCheck {
		go func() {
			blockNumber := blockHeight - i
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
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)
	account, err := client.Account(AccountOne)
	require.NoError(t, err)
	sequenceNumber, err := account.SequenceNumber()
	require.NoError(t, err)
	assert.Equal(t, uint64(0), sequenceNumber)
	authKey, err := account.AuthenticationKey()
	require.NoError(t, err)
	assert.Equal(t, AccountOne[:], authKey)
}

func Test_Transactions(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	start := uint64(1)
	count := uint64(2)
	// Specific 2 should only give 2
	transactions, err := client.Transactions(&start, &count)
	require.NoError(t, err)
	assert.Len(t, transactions, 2)

	// This will give the latest 2
	transactions, err = client.Transactions(nil, &count)
	require.NoError(t, err)
	assert.Len(t, transactions, 2)

	// This will give the 25 from 2
	transactions, err = client.Transactions(&start, nil)
	require.NoError(t, err)
	assert.Len(t, transactions, 25)

	// This will give the latest 25
	transactions, err = client.Transactions(nil, nil)
	require.NoError(t, err)
	assert.Len(t, transactions, 25)
}

func submitAccountTransaction(t *testing.T, client *Client, account *Account, seqNo uint64) {
	t.Helper()
	payload, err := CoinTransferPayload(nil, AccountOne, 1)
	require.NoError(t, err)
	rawTxn, err := client.BuildTransaction(account.AccountAddress(), TransactionPayload{Payload: payload}, SequenceNumber(seqNo))
	require.NoError(t, err)
	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)
	txn, err := client.SubmitTransaction(signedTxn)
	require.NoError(t, err)
	_, err = client.WaitForTransaction(txn.Hash)
	require.NoError(t, err)
}

func Test_AccountTransactions(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	// Create a bunch of transactions so we can test the pagination
	account, err := NewEd25519Account()
	require.NoError(t, err)
	err = client.Fund(account.AccountAddress(), 100_000_000)
	require.NoError(t, err)

	// Build and submit 100 transactions
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(100)

	for i := range uint64(100) {
		go func() {
			//nolint:testifylint
			submitAccountTransaction(t, client, account, i)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()

	// Submit one more transaction
	submitAccountTransaction(t, client, account, 100)

	// Fetch default
	transactions, err := client.AccountTransactions(account.AccountAddress(), nil, nil)
	require.NoError(t, err)
	assert.Len(t, transactions, 25)

	// Fetch 101 with no start
	zero := uint64(0)
	one := uint64(1)
	ten := uint64(10)
	hundredOne := uint64(101)

	transactions, err = client.AccountTransactions(account.AccountAddress(), nil, &hundredOne)
	require.NoError(t, err)
	assert.Len(t, transactions, 101)

	// Fetch 101 with start
	transactions, err = client.AccountTransactions(account.AccountAddress(), &zero, &hundredOne)
	require.NoError(t, err)
	assert.Len(t, transactions, 101)

	// Fetch 100 from 1
	transactions, err = client.AccountTransactions(account.AccountAddress(), &one, &hundredOne)
	require.NoError(t, err)
	assert.Len(t, transactions, 100)

	// Fetch default from 0
	transactions, err = client.AccountTransactions(account.AccountAddress(), &zero, nil)
	require.NoError(t, err)
	assert.Len(t, transactions, 25)

	// Check global transactions API

	t.Run("Default transaction size, no start", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(nil, nil)
		require.NoError(t, err)
		assert.Len(t, transactions, 25)
	})
	t.Run("Default transaction size, start from zero", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(&zero, nil)
		require.NoError(t, err)
		assert.Len(t, transactions, 25)
	})
	t.Run("Default transaction size, start from one", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(&one, nil)
		require.NoError(t, err)
		assert.Len(t, transactions, 25)
	})

	t.Run("101 transactions, no start", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(nil, &hundredOne)
		require.NoError(t, err)
		assert.Len(t, transactions, 101)
	})

	t.Run("101 transactions, start zero", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(&zero, &hundredOne)
		require.NoError(t, err)
		assert.Len(t, transactions, 101)
	})

	t.Run("101 transactions, start one", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(&one, &hundredOne)
		require.NoError(t, err)
		assert.Len(t, transactions, 101)
	})

	t.Run("10 transactions, no start", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(nil, &ten)
		require.NoError(t, err)
		assert.Len(t, transactions, 10)
	})

	t.Run("10 transactions, start one", func(t *testing.T) {
		t.Parallel()
		transactions, err = client.Transactions(&one, &ten)
		require.NoError(t, err)
		assert.Len(t, transactions, 10)
	})
}

func Test_Info(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	info, err := client.Info()
	require.NoError(t, err)
	assert.Greater(t, info.BlockHeight(), uint64(1))
}

func Test_AccountResources(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	resources, err := client.AccountResources(AccountOne)
	require.NoError(t, err)
	assert.NotEmpty(t, resources)

	resourcesBcs, err := client.AccountResourcesBCS(AccountOne)
	require.NoError(t, err)
	assert.NotEmpty(t, resourcesBcs)
}

// A worker thread that reads from a chan of transactions that have been submitted and waits on their completion status
func concurrentTxnWaiter(
	t *testing.T,
	results chan TransactionSubmissionResponse,
	waitResults chan ConcResponse[*api.UserTransaction],
	client *Client,
	wg *sync.WaitGroup,
) {
	t.Helper()
	if wg != nil {
		defer wg.Done()
	}
	responseCount := 0
	for response := range results {
		responseCount++
		require.NoError(t, response.Err)

		waitResponse, err := client.WaitForTransaction(response.Response.Hash, PollTimeout(21*time.Second))
		switch err {
		case nil:
			t.Logf("%s err %s", response.Response.Hash, err)
		default:
			if waitResponse == nil {
				t.Logf("%s nil response", response.Response.Hash)
			} else if !waitResponse.Success {
				t.Logf("%s !Success", response.Response.Hash)
			}
		}
		waitResults <- ConcResponse[*api.UserTransaction]{Result: waitResponse, Err: err}
	}
	t.Logf("concurrentTxnWaiter done, %d responses", responseCount)
	// signal completion
	// (do not close the output as there may be other workers writing to it)
	waitResults <- ConcResponse[*api.UserTransaction]{Result: nil, Err: nil}
}

func Test_Concurrent_Submission(t *testing.T) {
	t.Parallel()
	const numTxns = uint64(100)
	const numWaiters = 4

	client, err := createTestClient()
	require.NoError(t, err)

	account1, err := NewEd25519Account()
	require.NoError(t, err)

	err = client.Fund(account1.AccountAddress(), 100_000_000)
	require.NoError(t, err)

	// start submission goroutine
	payloads := make(chan TransactionBuildPayload, 50)
	results := make(chan TransactionSubmissionResponse, 50)
	go client.BuildSignAndSubmitTransactions(account1, payloads, results, ExpirationSeconds(20))

	transferPayload, err := CoinTransferPayload(nil, AccountOne, 100)
	require.NoError(t, err)

	// Generate transactions
	for i := range numTxns {
		payloads <- TransactionBuildPayload{
			Id:   i,
			Type: TransactionSubmissionTypeSingle, // TODO: not needed?
			Inner: TransactionPayload{
				Payload: transferPayload,
			},
		}
	}
	close(payloads)
	t.Log("done submitting txns")

	// Start waiting on txns
	waitResults := make(chan ConcResponse[*api.UserTransaction], numWaiters*10)

	var wg sync.WaitGroup
	wg.Add(numWaiters)
	for range numWaiters {
		//nolint:testifylint
		go concurrentTxnWaiter(t, results, waitResults, client, &wg)
	}

	// Wait on all the results, recording the succeeding ones
	txnMap := make(map[uint64]bool)

	waitersRunning := numWaiters

	// We could wait on a close, but I'm going to be a little pickier here
	i := uint64(0)
	txnGoodEvents := 0
	for {
		response := <-waitResults
		if response.Err == nil && response.Result == nil {
			t.Log("txn waiter signaled done")
			waitersRunning--
			if waitersRunning == 0 {
				close(results)
				t.Log("last txn waiter done")
				break
			}
			continue
		}
		require.NoError(t, response.Err)
		assert.True(t, (response.Result != nil) && response.Result.Success)
		if response.Result != nil {
			txnMap[response.Result.SequenceNumber] = true
			txnGoodEvents++
		}
		i++
		if i >= numTxns {
			t.Logf("waited on %d txns, done", i)
			break
		}
	}
	t.Log("done waiting for txns, waiting for txn waiter threads")

	wg.Wait()

	// Check all transactions were successful from [0-numTxns)
	t.Logf("got %d(%d) successful txns of %d attempted, error submission indexes:", len(txnMap), txnGoodEvents, numTxns)
	allTrue := true
	for i := range numTxns {
		allTrue = allTrue && txnMap[i]
		if !txnMap[i] {
			t.Logf("%d", i)
		}
	}
	assert.True(t, allTrue, "all txns successful")
	assert.Equal(t, len(txnMap), int(numTxns), "num txns successful == num txns sent")
}

func TestClient_View(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	payload := &ViewPayload{
		Module:   ModuleId{Address: AccountOne, Name: "coin"},
		Function: "balance",
		ArgTypes: []TypeTag{AptosCoinTypeTag},
		Args:     [][]byte{AccountOne[:]},
	}
	vals, err := client.View(payload)
	require.NoError(t, err)
	assert.Len(t, vals, 1)
	str, ok := vals[0].(string)
	require.True(t, ok)
	_, err = StrToUint64(str)
	require.NoError(t, err)
}

func TestClient_BlockByHeight(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	_, err = client.BlockByHeight(1, true)
	require.NoError(t, err)
}

func TestClient_NodeAPIHealthCheck(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	t.Run("Node API health check default", func(t *testing.T) {
		t.Parallel()
		response, err := client.NodeAPIHealthCheck()
		require.NoError(t, err)
		assert.Contains(t, response.Message, "ok", "Node API health check failed"+response.Message)
	})

	// Now, check node API health check with a future time that should never fail
	t.Run("Node API health check far future", func(t *testing.T) {
		t.Parallel()
		response, err := client.NodeAPIHealthCheck(10000)
		require.NoError(t, err)
		assert.Contains(t, response.Message, "ok", "Node API health check failed"+response.Message)
	})

	// Now, check node API health check with 0
	t.Run("Node API health check fail", func(t *testing.T) {
		t.Parallel()
		// Now, check node API health check with a time that should probably fail
		_, err := client.NodeAPIHealthCheck(0)
		require.Error(t, err)
	})
}

func buildSingleSignerEntryFunction(client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error) {
	return APTTransferTransaction(client, sender, AccountOne, 100, options...)
}

func buildSingleSignerScript(client *Client, sender TransactionSigner, options ...any) (*RawTransaction, error) {
	scriptBytes, err := ParseHex(singleSignerScript)
	if err != nil {
		return nil, err
	}

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

func TestClient_EntryFunctionWithArgs(t *testing.T) {
	t.Parallel()
	client, err := createTestClient()
	require.NoError(t, err)

	// Setup account
	account, err := types.NewEd25519Account()
	require.NoError(t, err)
	err = client.Fund(account.Address, 100000000)
	require.NoError(t, err)

	// Attempt to transfer to an account, using the ABI directly
	payload, err := client.EntryFunctionWithArgs(AccountOne, "aptos_account", "transfer_coins", []any{"0x1::aptos_coin::AptosCoin"}, []any{AccountTwo, 100})
	require.NoError(t, err)
	rawTxn, err := client.BuildTransaction(account.Address, TransactionPayload{Payload: payload})
	require.NoError(t, err)
	signedTxn, err := rawTxn.SignedTransaction(account)
	require.NoError(t, err)
	transaction, err := client.SubmitTransaction(signedTxn)
	require.NoError(t, err)
	txn, err := client.WaitForTransaction(transaction.Hash)
	require.NoError(t, err)

	assert.True(t, txn.Success)
}

func TestEnsureClientInterfacesHold(t *testing.T) {
	t.Parallel()
	var err error
	var c AptosClient
	c, err = NewClient(LocalnetConfig)
	require.NoError(t, err)
	require.NotNil(t, c)
	var rc AptosRpcClient
	rc, err = NewClient(LocalnetConfig)
	require.NoError(t, err)
	require.NotNil(t, rc)
	var fc AptosFaucetClient
	fc, err = NewClient(LocalnetConfig)
	require.NoError(t, err)
	require.NotNil(t, fc)
	var ic AptosIndexerClient
	ic, err = NewClient(LocalnetConfig)
	require.NoError(t, err)
	require.NotNil(t, ic)

	rc, err = NewNodeClient(LocalnetConfig.NodeUrl, 4)
	require.NoError(t, err)
	require.NotNil(t, rc)
	nc, err := NewNodeClient(LocalnetConfig.NodeUrl, 4)
	require.NoError(t, err)
	fc, err = NewFaucetClient(nc, LocalnetConfig.FaucetUrl)
	require.NoError(t, err)
	require.NotNil(t, fc)
	ic = NewIndexerClient(nc.client, LocalnetConfig.IndexerUrl)
	require.NoError(t, err)
	require.NotNil(t, ic)
}
