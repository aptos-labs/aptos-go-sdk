package aptos

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedConfig(t *testing.T) {
	names := []string{"mainnet", "devnet", "testnet", "localnet"}
	for _, name := range names {
		assert.Equal(t, name, NamedNetworks[name].Name)
	}
}

func TestAptosClientHeaderValue(t *testing.T) {
	assert.Greater(t, len(AptosClientHeaderValue), 0)
	assert.NotEqual(t, "aptos-go-sdk/unk", AptosClientHeaderValue)
}

func Test_Flow(t *testing.T) {
	if testing.Short() {
		// TODO: only run this in some integration mode set by environment variable?
		// TODO: allow this to be harmlessly flakey if devnet is down?
		// TODO: write a framework to optionally run things against `aptos node run-localnet`
		t.Skip("integration test expects network connection to devnet in cloud")
	}
	// Create a client
	client, err := NewClient(DevnetConfig)
	assert.NoError(t, err)

	// Verify chain id retrieval works
	chainId, err := client.GetChainId()
	assert.NoError(t, err)
	assert.Less(t, uint8(4), chainId)

	// Create an account
	account, err := NewEd25519Account()
	assert.NoError(t, err)

	// Fund the account with 1 APT
	err = client.Fund(account.Address, 100_000_000)
	assert.NoError(t, err)

	// Send money to 0x1
	// Build transaction
	signed_txn, err := APTTransferTransaction(client, account, AccountOne, 100)
	assert.NoError(t, err)

	// Send transaction
	result, err := client.SubmitTransaction(signed_txn)
	assert.NoError(t, err)

	hash := result["hash"].(string)

	// Wait for the transaction
	_, err = client.WaitForTransaction(hash)
	assert.NoError(t, err)

	// Read transaction by hash
	txn, err := client.TransactionByHash(hash)
	assert.NoError(t, err)

	// Read transaction by version
	versionString := txn["version"].(string)

	// Convert string version to uint64
	version, err := strconv.ParseUint(versionString, 10, 64)
	assert.NoError(t, err)

	// Load the transaction again
	txnByVersion, err := client.TransactionByVersion(version)

	// Assert that both are the same
	assert.Equal(t, txn, txnByVersion)
}
