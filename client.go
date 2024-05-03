package aptos

import (
	"errors"
	"net/url"
	"time"
)

type NetworkConfig struct {
	Name       string
	ChainId    uint8
	NodeUrl    string
	IndexerUrl string
	FaucetUrl  string
}

var LocalnetConfig = NetworkConfig{
	Name:       "localnet",
	ChainId:    4,
	NodeUrl:    "http://localhost:8080/v1",
	IndexerUrl: "",
	FaucetUrl:  "http://localhost:8081/v1",
}
var DevnetConfig = NetworkConfig{
	Name:       "devnet",
	ChainId:    3,
	NodeUrl:    "https://api.devnet.aptoslabs.com/v1",
	IndexerUrl: "",
	FaucetUrl:  "https://faucet.devnet.aptoslabs.com/",
}
var TestnetConfig = NetworkConfig{
	Name:       "testnet",
	ChainId:    2,
	NodeUrl:    "https://api.testnet.aptoslabs.com/v1",
	IndexerUrl: "",
	FaucetUrl:  "https://faucet.testnet.aptoslabs.com/",
}
var MainnetConfig = NetworkConfig{
	Name:       "mainnet",
	ChainId:    1,
	NodeUrl:    "https://api.mainnet.aptoslabs.com/v1",
	IndexerUrl: "",
	FaucetUrl:  "",
}

// Map from network name to NetworkConfig
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
type Client struct {
	nodeClient   NodeClient
	faucetClient FaucetClient
	// TODO: Add indexer client
}

var ErrUnknownNetworkName = errors.New("Unknown network name")

// NewClientFromNetworkName Creates a new client for a specific network name
func NewClientFromNetworkName(networkName string) (client *Client, err error) {
	config, ok := NamedNetworks[networkName]
	if !ok {
		return nil, ErrUnknownNetworkName

	}
	return NewClient(config)
}

// NewClient Creates a new client with a specific network config that can be extended in the future
func NewClient(config NetworkConfig) (client *Client, err error) {
	nodeUrl, err := url.Parse(config.NodeUrl)
	if err != nil {
		return nil, err
	}

	faucetUrl, err := url.Parse(config.FaucetUrl)
	if err != nil {
		return nil, err
	}

	// TODO: add indexer

	nodeClient := new(NodeClient)
	nodeClient.baseUrl = *nodeUrl
	nodeClient.client.Timeout = 60 * time.Second
	faucetClient := &FaucetClient{
		nodeClient,
		*faucetUrl,
	}
	client = &Client{
		*nodeClient,
		*faucetClient,
	}
	return
}

func (client *Client) SetTimeout(timeout time.Duration) {
	client.nodeClient.client.Timeout = timeout
}

func (client *Client) Info() (info NodeInfo, err error) {
	return client.nodeClient.Info()
}

func (client *Client) Account(address AccountAddress, ledgerVersion ...int) (info AccountInfo, err error) {
	return client.nodeClient.Account(address, ledgerVersion...)
}

func (client *Client) AccountResource(address AccountAddress, resourceType string, ledgerVersion ...int) (data map[string]any, err error) {
	return client.nodeClient.AccountResource(address, resourceType, ledgerVersion...)
}

// AccountResources fetches resources for an account into a JSON-like map[string]any in AccountResourceInfo.Data
// For fetching raw Move structs as BCS, See #AccountResourcesBCS
func (client *Client) AccountResources(address AccountAddress, ledgerVersion ...int) (resources []AccountResourceInfo, err error) {
	return client.nodeClient.AccountResources(address, ledgerVersion...)
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (client *Client) AccountResourcesBCS(address AccountAddress, ledgerVersion ...int) (resources []AccountResourceRecord, err error) {
	return client.nodeClient.AccountResourcesBCS(address, ledgerVersion...)
}

// TransactionByHash gets info on a transaction
// The transaction may be pending or recently committed.
//
//	data, err := c.TransactionByHash("0xabcd")
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
func (client *Client) TransactionByHash(txnHash string) (data map[string]any, err error) {
	return client.nodeClient.TransactionByHash(txnHash)
}

func (client *Client) TransactionByVersion(version uint64) (data map[string]any, err error) {
	return client.nodeClient.TransactionByVersion(version)
}

// PollForTransactions Waits up to 10 seconds for transactions to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
func (client *Client) PollForTransactions(txnHashes []string, options ...any) error {
	return client.nodeClient.PollForTransactions(txnHashes, options...)
}

// Do a long-GET for one transaction and wait for it to complete
func (client *Client) WaitForTransaction(txnHash string) (data map[string]any, err error) {
	return client.nodeClient.WaitForTransaction(txnHash)
}

// Transactions Get recent transactions.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
func (client *Client) Transactions(start *uint64, limit *uint64) (data []map[string]any, err error) {
	return client.nodeClient.Transactions(start, limit)
}

func (client *Client) SubmitTransaction(signedTransaction *SignedTransaction) (data map[string]any, err error) {
	return client.nodeClient.SubmitTransaction(signedTransaction)
}

func (client *Client) GetChainId() (chainId uint8, err error) {
	return client.nodeClient.GetChainId()
}

func (client *Client) Fund(address AccountAddress, amount uint64) error {
	return client.faucetClient.Fund(address, amount)
}

func (client *Client) BuildTransaction(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error) {
	return client.nodeClient.BuildTransaction(sender, payload, options...)
}

func (client *Client) BuildSignAndSubmitTransaction(sender Account, payload TransactionPayload, options ...any) (hash string, err error) {
	return client.nodeClient.BuildSignAndSubmitTransaction(sender, payload, options...)
}

// TODO: support ledger version
func (client *Client) View(payload *ViewPayload) (vals []any, err error) {
	return client.nodeClient.View(payload)
}
