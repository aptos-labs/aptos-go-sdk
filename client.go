package aptos

import (
	"errors"
	"net/url"
	"time"
)

// For Content-Type header
const APTOS_SIGNED_BCS = "application/x.aptos.signed_transaction+bcs"

const (
	Localnet = "localnet"
	Devnet   = "devnet"
	Testnet  = "testnet"
	Mainnet  = "mainnet"
)

const (
	localnet_api = "http://localhost:8080/v1"
	devnet_api   = "https://api.devnet.aptoslabs.com/v1"
	testnet_api  = "https://api.testnet.aptoslabs.com/v1"
	mainnet_api  = "https://api.mainnet.aptoslabs.com/v1"
)

const (
	localnet_faucet = "http://localhost:8081/v1"
	devnet_faucet   = "https://faucet.devnet.aptoslabs.com/"
	testnet_faucet  = "https://faucet.testnet.aptoslabs.com/"
)

const (
	localnet_chain_id = 4
	testnet_chain_id  = 2
	mainnet_chain_id  = 1
)

type NetworkConfig struct {
	network *string
	api     *string
	faucet  *string
	indexer *string
}

// Client is a facade over the multiple types of underlying clients, as the user doesn't actually care where the data
// comes from.  It will be then handled underneath
type Client struct {
	nodeClient   NodeClient
	faucetClient FaucetClient
	// TODO: Add indexer client
}

// NewClientFromNetworkName Creates a new client for a specific network name
func NewClientFromNetworkName(network *string) (client *Client, err error) {
	config := NetworkConfig{network: network}
	client, err = NewClient(config)
	return
}

// NewClient Creates a new client with a specific network config that can be extended in the future
func NewClient(config NetworkConfig) (client *Client, err error) {
	var apiUrl *url.URL = nil

	switch {
	case config.api == nil && config.network == nil:
		err = errors.New("aptos api url or network is required")
		return
	case config.api != nil:
		apiUrl, err = url.Parse(*config.api)
		if err != nil {
			return
		}
	case *config.network == Localnet:
		apiUrl, err = url.Parse(localnet_api)
		if err != nil {
			return
		}
	case *config.network == Devnet:
		apiUrl, err = url.Parse(devnet_api)
		if err != nil {
			return
		}
	case *config.network == Testnet:
		apiUrl, err = url.Parse(testnet_api)
		if err != nil {
			return
		}
	case *config.network == Mainnet:
		apiUrl, err = url.Parse(mainnet_api)
		if err != nil {
			return
		}
	default:
		err = errors.New("network name is unknown, please put Localnet, Devnet, Testnet, or Mainnet")
		return
	}
	var faucetUrl *url.URL = nil

	switch {
	case config.faucet == nil && config.network == nil:
		err = errors.New("aptos faucet url or network is required")
		return
	case config.faucet != nil:
		faucetUrl, err = url.Parse(*config.faucet)
		if err != nil {
			return
		}
	case *config.network == Localnet:
		faucetUrl, err = url.Parse(localnet_faucet)
		if err != nil {
			return
		}
	case *config.network == Devnet:
		faucetUrl, err = url.Parse(devnet_faucet)
		if err != nil {
			return
		}
	case *config.network == Testnet:
		faucetUrl, err = url.Parse(testnet_faucet)
		if err != nil {
			return
		}
	case *config.network == Mainnet:
		faucetUrl = nil
	default:
		err = errors.New("network name is unknown, please put Localnet, Devnet, Testnet, or Mainnet")
		return
	}

	// TODO: add indexer

	restClient := new(NodeClient)
	restClient.baseUrl = *apiUrl
	restClient.client.Timeout = 60 * time.Second // TODO: Make configurable
	faucetClient := &FaucetClient{
		restClient,
		faucetUrl,
	}
	client = &Client{
		*restClient,
		*faucetClient,
	}
	return
}

func (client *Client) Info() (info NodeInfo, err error) {
	return client.nodeClient.Info()
}

func (client *Client) Account(address AccountAddress, ledgerVersion ...int) (info AccountInfo, err error) {
	return client.nodeClient.Account(address, ledgerVersion...)
}

// TODO: set HTTP header "x-aptos-client: aptos-go-sdk/{version}"

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

// WaitForTransactions Waits up to 10 seconds for transactions to be done, polling at 10Hz
// TODO: options for polling period and timeout
func (client *Client) WaitForTransactions(txnHashes []string) error {
	return client.nodeClient.WaitForTransactions(txnHashes)
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
