package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/api"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	singleSignerScript = "a11ceb0b060000000701000402040a030e0c041a04051e20073e30086e2000000001010204010001000308000104030401000105050601000002010203060c0305010b0001080101080102060c03010b0001090002050b00010900000a6170746f735f636f696e04636f696e04436f696e094170746f73436f696e087769746864726177076465706f7369740000000000000000000000000000000000000000000000000000000000000001000001080b000b0138000c030b020b03380102"
	fundAmount         = 100_000_000
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
	for name, signer := range TestSigners {
		println("Entry function:", name)
		testTransaction(t, signer, submitEntryFunction)
	}
}

func Test_Ed25519_ScriptFlow(t *testing.T) {
	for name, signer := range TestSigners {
		println("Script:", name)
		testTransaction(t, signer, submitScript)
	}
}

func testTransaction(t *testing.T, createAccount func() (TransactionSigner, error), buildAndSignTransaction func(t *testing.T, client *Client, sender TransactionSigner) (*SignedTransaction, error)) {
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

	// Build transaction
	signedTxn, err := buildAndSignTransaction(t, client, account)
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

	for i := uint64(0); i < numToCheck; i++ {
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
	}
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

func TestClient_BlockByHeight(t *testing.T) {
	client, err := createTestClient()
	assert.NoError(t, err)

	_, err = client.BlockByHeight(1, true)
	assert.NoError(t, err)
}

func submitEntryFunction(_ *testing.T, client *Client, sender TransactionSigner) (*SignedTransaction, error) {
	return APTTransferTransaction(client, sender, AccountOne, 100)
}

func submitScript(t *testing.T, client *Client, sender TransactionSigner) (*SignedTransaction, error) {
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
		}})
	if err != nil {
		return nil, err
	}

	return rawTxn.SignedTransaction(sender)
}
