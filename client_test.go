package aptos

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func Test_Flow(t *testing.T) {
	// Create a client
	network := Devnet
	config := NetworkConfig{
		network: &network,
	}
	client, err := NewClient(config)
	assert.NoError(t, err)

	// Verify chain id retrieval works
	chainId, err := client.GetChainId()
	assert.NoError(t, err)
	assert.Less(t, uint8(4), chainId)

	// Create an account
	account, err := NewAccount()
	assert.NoError(t, err)

	// Fund the account with 1 APT
	client.Fund(account.Address, 100_000_000)

	// Send money to 0x1
	// Build transaction
	signed_txn, err := APTTransferTransaction(client, account, Account0x1, 100)
	assert.NoError(t, err)

	// Send transaction
	// TODO: verify response
	result, err := client.SubmitTransaction(signed_txn)
	assert.NoError(t, err)

	hash := result["hash"].(string)

	// TODO Wait on transaction
	err = client.WaitForTransactions([]string{hash})
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
