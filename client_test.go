package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	singleSignerScript = "a11ceb0b060000000701000402040a030e0c041a04051e20073e30086e2000000001010204010001000308000104030401000105050601000002010203060c0305010b0001080101080102060c03010b0001090002050b00010900000a6170746f735f636f696e04636f696e04436f696e094170746f73436f696e087769746864726177076465706f7369740000000000000000000000000000000000000000000000000000000000000001000001080b000b0138000c030b020b03380102"
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

func Test_EntryFunctionFlow(t *testing.T) {
	testTransaction(t, func(client *Client, sender *Account) (*SignedTransaction, error) {
		return APTTransferTransaction(client, sender, AccountOne, 100)
	})
}

func Test_ScriptFlow(t *testing.T) {
	testTransaction(t, func(client *Client, sender *Account) (*SignedTransaction, error) {
		scriptBytes, err := ParseHex(singleSignerScript)
		assert.NoError(t, err)

		amount := uint64(1)
		dest := AccountOne

		rawTxn, err := client.BuildTransaction(sender.Address,
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
		return rawTxn.Sign(sender)
	})
}
func testTransaction(t *testing.T, buildAndSignTransaction func(client *Client, sender *Account) (*SignedTransaction, error)) {
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

	// Build transaction
	signedTxn, err := buildAndSignTransaction(client, account)
	assert.NoError(t, err)

	serializer := bcs.Serializer{}
	signedTxn.MarshalBCS(&serializer)

	// Send transaction
	result, err := client.SubmitTransaction(signedTxn)
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

func TestAPTTransferTransaction(t *testing.T) {
	sender, err := NewEd25519Account()
	assert.NoError(t, err)
	dest, err := NewEd25519Account()
	assert.NoError(t, err)

	client, err := NewClient(DevnetConfig)
	assert.NoError(t, err)
	stxn, err := APTTransferTransaction(client, sender, dest.Address, 1337, MaxGasAmount(123123), GasUnitPrice(111), ExpirationSeconds(42), ChainIdOption(71), SequenceNumber(31337))
	assert.NoError(t, err)
	assert.NotNil(t, stxn)

	// use defaults for: max gas amount, gas unit price
	stxn, err = APTTransferTransaction(client, sender, dest.Address, 1337, ExpirationSeconds(42), ChainIdOption(71), SequenceNumber(31337))
	assert.NoError(t, err)
	assert.NotNil(t, stxn)
}
