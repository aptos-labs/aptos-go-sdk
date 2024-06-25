// Package aptos is a Go interface into the Aptos blockchain.
//
// You can create a client and send a transfer transaction with the below example:
//
//	// Create a Client
//	client := NewClient(DevnetConfig)
//
//	// Create an account, and fund it
//	account := NewEd25519Account()
//	err := client.Fund(account.AccountAddress())
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to fund account %s %w", account.AccountAddress().ToString(), err))
//	}
//
//	// Send funds to a different address
//	receiver := &AccountAddress{}
//	receiver.ParseStringRelaxed("0xcafe")
//
//	// Build a transaction to send 1 APT to the receiver
//	amount := 100_000_000 // 1 APT
//	transferTransaction, err := APTTransferTransaction(client, account, receiver, amount)
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to build transaction %w", err))
//	}
//
//	// Submit transaction to the blockchain
//	submitResponse, err := client.SubmitTransaction(transferTransaction)
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to submit transaction %w", err))
//	}
//
//	// Wait for transaction to complete
//	err := client.WaitForTransaction(submitResponse.Hash)
//	if err != nil {
//	  panic(fmt.Sprintf("Failed to wait for transaction %w", err))
//	}
package aptos

import (
	"time"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/hasura/go-graphql-client"
)

// NetworkConfig a configuration for the Client and which network to use.  Use one of the preconfigured [LocalnetConfig], [DevnetConfig], [TestnetConfig], or [MainnetConfig] unless you have your own full node.
//
// Name, ChainId, IndexerUrl, FaucetUrl are not required.
//
// If ChainId is 0, the ChainId wil be fetched on-chain
// If IndexerUrl or FaucetUrl are an empty string "", clients will not be made for them.
type NetworkConfig struct {
	Name       string
	ChainId    uint8
	NodeUrl    string
	IndexerUrl string
	FaucetUrl  string
}

// LocalnetConfig is for use with a localnet, created by the [Aptos CLI](https://aptos.dev/tools/aptos-cli)
//
// To start a localnet, install the Aptos CLI then run:
//
//	aptos node run-localnet --with-indexer-api
var LocalnetConfig = NetworkConfig{
	Name:    "localnet",
	ChainId: 4,
	// We use 127.0.0.1 as it is more foolproof than localhost
	NodeUrl:    "http://127.0.0.1:8080/v1",
	IndexerUrl: "http://127.0.0.1:8090/v1/graphql",
	FaucetUrl:  "http://127.0.0.1:8081",
}

// DevnetConfig is for use with devnet.  Note devnet resets at least weekly.  ChainId differs after each reset.
var DevnetConfig = NetworkConfig{
	Name:       "devnet",
	NodeUrl:    "https://api.devnet.aptoslabs.com/v1",
	IndexerUrl: "https://api.devnet.aptoslabs.com/v1/graphql",
	FaucetUrl:  "https://faucet.devnet.aptoslabs.com/",
}

// TestnetConfig is for use with testnet. Testnet does not reset.
var TestnetConfig = NetworkConfig{
	Name:       "testnet",
	ChainId:    2,
	NodeUrl:    "https://api.testnet.aptoslabs.com/v1",
	IndexerUrl: "https://api.testnet.aptoslabs.com/v1/graphql",
	FaucetUrl:  "https://faucet.testnet.aptoslabs.com/",
}

// MainnetConfig is for use with mainnet.  There is no faucet for Mainnet, as these are real user assets.
var MainnetConfig = NetworkConfig{
	Name:       "mainnet",
	ChainId:    1,
	NodeUrl:    "https://api.mainnet.aptoslabs.com/v1",
	IndexerUrl: "https://api.mainnet.aptoslabs.com/v1/graphql",
	FaucetUrl:  "",
}

// NamedNetworks Map from network name to NetworkConfig
var NamedNetworks map[string]NetworkConfig

func init() {
	NamedNetworks = make(map[string]NetworkConfig, 4)
	setNN := func(nc NetworkConfig) {
		NamedNetworks[nc.Name] = nc
	}
	setNN(LocalnetConfig)
	setNN(DevnetConfig)
	setNN(TestnetConfig)
	setNN(MainnetConfig)
}

// Client is a facade over the multiple types of underlying clients, as the user doesn't actually care where the data
// comes from.  It will be then handled underneath
//
// To create a new client, please use [NewClient].  An example below for Devnet:
//
//	client := NewClient(DevnetConfig)
type Client struct {
	nodeClient    *NodeClient
	faucetClient  *FaucetClient
	indexerClient *IndexerClient
}

// NewClient Creates a new client with a specific network config that can be extended in the future
func NewClient(config NetworkConfig) (client *Client, err error) {
	nodeClient, err := NewNodeClient(config.NodeUrl, config.ChainId)
	if err != nil {
		return nil, err
	}
	// Indexer may not be present
	var indexerClient *IndexerClient = nil
	if config.IndexerUrl != "" {
		indexerClient = NewIndexerClient(nodeClient.client, config.IndexerUrl)
	}

	// Faucet may not be present
	var faucetClient *FaucetClient = nil
	if config.FaucetUrl != "" {
		faucetClient, err = NewFaucetClient(nodeClient, config.FaucetUrl)
		if err != nil {
			return nil, err
		}
	}

	// Fetch the chain Id if it isn't in the config
	if config.ChainId == 0 {
		_, _ = nodeClient.GetChainId()
	}

	client = &Client{
		nodeClient,
		faucetClient,
		indexerClient,
	}
	return
}

// SetTimeout adjusts the HTTP client timeout
//
//	client.SetTimeout(5 * time.Millisecond)
func (client *Client) SetTimeout(timeout time.Duration) {
	client.nodeClient.client.Timeout = timeout
}

// Info Retrieves the node info about the network and it's current state
func (client *Client) Info() (info NodeInfo, err error) {
	return client.nodeClient.Info()
}

// Account Retrieves information about the account such as [SequenceNumber] and [crypto.AuthenticationKey]
func (client *Client) Account(address AccountAddress, ledgerVersion ...uint64) (info AccountInfo, err error) {
	return client.nodeClient.Account(address, ledgerVersion...)
}

// AccountResource Retrieves a single resource given its struct name.
//
//	address := AccountOne
//	dataMap, _ := client.AccountResource(address, "0x1::coin::CoinStore")
//
// Can also fetch at a specific ledger version
//
//	address := AccountOne
//	dataMap, _ := client.AccountResource(address, "0x1::coin::CoinStore", 1)
func (client *Client) AccountResource(address AccountAddress, resourceType string, ledgerVersion ...uint64) (data map[string]any, err error) {
	return client.nodeClient.AccountResource(address, resourceType, ledgerVersion...)
}

// AccountResources fetches resources for an account into a JSON-like map[string]any in AccountResourceInfo.Data
// For fetching raw Move structs as BCS, See #AccountResourcesBCS
//
//	address := AccountOne
//	dataMap, _ := client.AccountResources(address)
//
// Can also fetch at a specific ledger version
//
//	address := AccountOne
//	dataMap, _ := client.AccountResource(address, 1)
func (client *Client) AccountResources(address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceInfo, err error) {
	return client.nodeClient.AccountResources(address, ledgerVersion...)
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (client *Client) AccountResourcesBCS(address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceRecord, err error) {
	return client.nodeClient.AccountResourcesBCS(address, ledgerVersion...)
}

// BlockByHeight fetches a block by height
//
//	block, _ := client.BlockByHeight(1, false)
//
// Can also fetch with transactions
//
//	block, _ := client.BlockByHeight(1, true)
func (client *Client) BlockByHeight(blockHeight uint64, withTransactions bool) (data *api.Block, err error) {
	return client.nodeClient.BlockByHeight(blockHeight, withTransactions)
}

// BlockByVersion fetches a block by ledger version
//
//	block, _ := client.BlockByVersion(123, false)
//
// Can also fetch with transactions
//
//	block, _ := client.BlockByVersion(123, true)
func (client *Client) BlockByVersion(ledgerVersion uint64, withTransactions bool) (data *api.Block, err error) {
	return client.nodeClient.BlockByVersion(ledgerVersion, withTransactions)
}

// TransactionByHash gets info on a transaction
// The transaction may be pending or recently committed.
//
//	data, err := client.TransactionByHash("0xabcd")
//	if err != nil {
//		if httpErr, ok := err.(aptos.HttpError) {
//			if httpErr.StatusCode == 404 {
//				// if we're sure this has been submitted, assume it is still pending elsewhere in the mempool
//			}
//		}
//	} else {
//		if data["type"] == "pending_transaction" {
//			// known to local mempool, but not committed yet
//		}
//	}
func (client *Client) TransactionByHash(txnHash string) (data *api.Transaction, err error) {
	return client.nodeClient.TransactionByHash(txnHash)
}

// TransactionByVersion gets info on a transaction from its LedgerVersion.  It must have been
// committed to have a ledger version
//
//	data, err := client.TransactionByVersion("0xabcd")
//	if err != nil {
//		if httpErr, ok := err.(aptos.HttpError) {
//			if httpErr.StatusCode == 404 {
//				// if we're sure this has been submitted, the full node might not be caught up to this version yet
//			}
//		}
//	}
func (client *Client) TransactionByVersion(version uint64) (data *api.Transaction, err error) {
	return client.nodeClient.TransactionByVersion(version)
}

// PollForTransactions Waits up to 10 seconds for transactions to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
//
//	hashes := []string{"0x1234", "0x4567"}
//	err := client.PollForTransactions(hashes)
//
// Can additionally configure different options
//
//	hashes := []string{"0x1234", "0x4567"}
//	err := client.PollForTransactions(hashes, PollPeriod(500 * time.Milliseconds), PollTimeout(5 * time.Seconds))
func (client *Client) PollForTransactions(txnHashes []string, options ...any) error {
	return client.nodeClient.PollForTransactions(txnHashes, options...)
}

// WaitForTransaction Do a long-GET for one transaction and wait for it to complete
func (client *Client) WaitForTransaction(txnHash string, options ...any) (data *api.UserTransaction, err error) {
	return client.nodeClient.WaitForTransaction(txnHash, options...)
}

// Transactions Get recent transactions.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
//
//	client.Transactions(0, 2)   // Returns 2 transactions
//	client.Transactions(1, 100) // Returns 100 transactions
func (client *Client) Transactions(start *uint64, limit *uint64) (data []*api.Transaction, err error) {
	return client.nodeClient.Transactions(start, limit)
}

// SubmitTransaction Submits an already signed transaction to the blockchain
func (client *Client) SubmitTransaction(signedTransaction *SignedTransaction) (data *api.SubmitTransactionResponse, err error) {
	return client.nodeClient.SubmitTransaction(signedTransaction)
}

// SimulateTransaction Simulates a raw transaction without sending it to the blockchain
func (client *Client) SimulateTransaction(rawTxn *RawTransaction, sender TransactionSigner, options ...any) (data []*api.UserTransaction, err error) {
	return client.nodeClient.SimulateTransaction(rawTxn, sender, options...)
}

// GetChainId Retrieves the ChainId of the network
// Note this will be cached forever, or taken directly from the config
func (client *Client) GetChainId() (chainId uint8, err error) {
	return client.nodeClient.GetChainId()
}

// Fund Uses the faucet to fund an address, only applies to non-production networks
func (client *Client) Fund(address AccountAddress, amount uint64) error {
	return client.faucetClient.Fund(address, amount)
}

// BuildTransaction Builds a raw transaction from the payload and fetches any necessary information from on-chain
//
//	sender := NewEd25519Account()
//	txnPayload := TransactionPayload{
//		Payload: &EntryFunction{
//			Module: ModuleId{
//				Address: AccountOne,
//				Name: "aptos_account",
//			},
//			Function: "transfer",
//			ArgTypes: []TypeTag{},
//			Args: [][]byte{
//				dest[:],
//				amountBytes,
//			},
//		}
//	}
//	rawTxn, err := client.BuildTransaction(sender.AccountAddress(), txnPayload)
func (client *Client) BuildTransaction(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error) {
	return client.nodeClient.BuildTransaction(sender, payload, options...)
}

// BuildTransactionMultiAgent Builds a raw transaction for MultiAgent or FeePayer from the payload and fetches any necessary information from on-chain
//
//	sender := NewEd25519Account()
//	txnPayload := TransactionPayload{
//		Payload: &EntryFunction{
//			Module: ModuleId{
//				Address: AccountOne,
//				Name: "aptos_account",
//			},
//			Function: "transfer",
//			ArgTypes: []TypeTag{},
//			Args: [][]byte{
//				dest[:],
//				amountBytes,
//			},
//		}
//	}
//	rawTxn, err := client.BuildTransactionMultiAgent(sender.AccountAddress(), txnPayload, FeePayer(AccountZero))
func (client *Client) BuildTransactionMultiAgent(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransactionWithData, err error) {
	return client.nodeClient.BuildTransactionMultiAgent(sender, payload, options...)
}

// BuildSignAndSubmitTransaction Convenience function to do all three in one
// for more configuration, please use them separately
//
//	sender := NewEd25519Account()
//	txnPayload := TransactionPayload{
//		Payload: &EntryFunction{
//			Module: ModuleId{
//				Address: AccountOne,
//				Name: "aptos_account",
//			},
//			Function: "transfer",
//			ArgTypes: []TypeTag{},
//			Args: [][]byte{
//				dest[:],
//				amountBytes,
//			},
//		}
//	}
//	submitResponse, err := client.BuildSignAndSubmitTransaction(sender, txnPayload)
func (client *Client) BuildSignAndSubmitTransaction(sender *Account, payload TransactionPayload, options ...any) (data *api.SubmitTransactionResponse, err error) {
	return client.nodeClient.BuildSignAndSubmitTransaction(sender, payload, options...)
}

// View Runs a view function on chain returning a list of return values.
//
//	 address := AccountOne
//		payload := &ViewPayload{
//			Module: ModuleId{
//				Address: AccountOne,
//				Name:    "coin",
//			},
//			Function: "balance",
//			ArgTypes: []TypeTag{AptosCoinTypeTag},
//			Args:     [][]byte{address[:]},
//		}
//		vals, err := client.aptosClient.View(payload)
//		balance := StrToU64(vals.(any[])[0].(string))
func (client *Client) View(payload *ViewPayload, ledgerVersion ...uint64) (vals []any, err error) {
	return client.nodeClient.View(payload, ledgerVersion...)
}

// EstimateGasPrice Retrieves the gas estimate from the network.
func (client *Client) EstimateGasPrice() (info EstimateGasInfo, err error) {
	return client.nodeClient.EstimateGasPrice()
}

// AccountAPTBalance retrieves the APT balance in the account
func (client *Client) AccountAPTBalance(address AccountAddress) (uint64, error) {
	return client.nodeClient.AccountAPTBalance(address)
}

// QueryIndexer queries the indexer using GraphQL to fill the `query` struct with data.  See examples in the indexer client on how to make queries
//
//	var out []CoinBalance
//	var q struct {
//		Current_coin_balances []struct {
//			CoinType     string `graphql:"coin_type"`
//			Amount       uint64
//			OwnerAddress string `graphql:"owner_address"`
//		} `graphql:"current_coin_balances(where: {owner_address: {_eq: $address}})"`
//	}
//	variables := map[string]any{
//		"address": address.StringLong(),
//	}
//	err := client.QueryIndexer(&q, variables)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, coin := range q.Current_coin_balances {
//		out = append(out, CoinBalance{
//			CoinType: coin.CoinType,
//			Amount:   coin.Amount,
//	})
//	}
//
//	return out, nil
func (client *Client) QueryIndexer(query any, variables map[string]any, options ...graphql.Option) error {
	return client.indexerClient.Query(query, variables, options...)
}

// GetProcessorStatus returns the ledger version up to which the processor has processed
func (client *Client) GetProcessorStatus(processorName string) (uint64, error) {
	return client.indexerClient.GetProcessorStatus(processorName)
}

// GetCoinBalances gets the balances of all coins associated with a given address
func (client *Client) GetCoinBalances(address AccountAddress) ([]CoinBalance, error) {
	return client.indexerClient.GetCoinBalances(address)
}

// NodeAPIHealthCheck checks if the node is within durationSecs of the current time, if not provided the node default is used
func (client *Client) NodeAPIHealthCheck(durationSecs ...uint64) (api.HealthCheckResponse, error) {
	return client.nodeClient.NodeHealthCheck(durationSecs...)
}
