package aptos

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/hasura/go-graphql-client"
)

var _ ExposedAptosClient = (*ExposedClient)(nil)

// ExposedAptosClient is an interface for all functionality on the Client.
// It is a combination of [AptosRpcClient], [AptosIndexerClient], and [AptosFaucetClient] for the purposes
// of mocking and convenience.
type ExposedAptosClient interface {
	ExposedAptosRpcClient
	ExposedAptosIndexerClient
	ExposedAptosFaucetClient
}

// ExposedAptosRpcClient is an interface for all functionality on the Client that is Node RPC related.  Its main implementation
// is [WrappedNodeClient]
type ExposedAptosRpcClient interface {
	// SetHeader sets the header for all future requests
	//
	//	client.SetHeader("Authorization", "Bearer abcde")
	SetHeader(key string, value string)

	// RemoveHeader removes the header from being automatically set all future requests.
	//
	//	client.RemoveHeader("Authorization")
	RemoveHeader(key string)

	// Info Retrieves the node info about the network and it's current state
	Info(ctx context.Context) (info NodeInfo, err error)

	// Account Retrieves information about the account such as [SequenceNumber] and [crypto.AuthenticationKey]
	Account(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (info AccountInfo, err error)

	// AccountResource Retrieves a single resource given its struct name.
	//
	//	address := AccountOne
	//	dataMap, _ := client.AccountResource(address, "0x1::coin::CoinStore")
	//
	// Can also fetch at a specific ledger version
	//
	//	address := AccountOne
	//	dataMap, _ := client.AccountResource(address, "0x1::coin::CoinStore", 1)
	AccountResource(ctx context.Context, address AccountAddress, resourceType string, ledgerVersion ...uint64) (data map[string]any, err error)

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
	AccountResources(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceInfo, err error)

	// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
	AccountResourcesBCS(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceRecord, err error)

	// BlockByHeight fetches a block by height
	//
	//	block, _ := client.BlockByHeight(1, false)
	//
	// Can also fetch with transactions
	//
	//	block, _ := client.BlockByHeight(1, true)
	BlockByHeight(ctx context.Context, blockHeight uint64, withTransactions bool) (data *api.Block, err error)

	// BlockByVersion fetches a block by ledger version
	//
	//	block, _ := client.BlockByVersion(123, false)
	//
	// Can also fetch with transactions
	//
	//	block, _ := client.BlockByVersion(123, true)
	BlockByVersion(ctx context.Context, ledgerVersion uint64, withTransactions bool) (data *api.Block, err error)

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
	TransactionByHash(ctx context.Context, txnHash string) (data *api.Transaction, err error)

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
	TransactionByVersion(ctx context.Context, version uint64) (data *api.CommittedTransaction, err error)

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
	PollForTransactions(ctx context.Context, txnHashes []string, options ...any) error

	// WaitForTransaction Do a long-GET for one transaction and wait for it to complete
	//
	//	data, err := client.WaitForTransaction("0x1234")
	WaitForTransaction(ctx context.Context, txnHash string, options ...any) (data *api.UserTransaction, err error)

	// Transactions Get recent transactions.
	// Start is a version number. Nil for most recent transactions.
	// Limit is a number of transactions to return. 'about a hundred' by default.
	//
	//	client.Transactions(0, 2)   // Returns 2 transactions
	//	client.Transactions(1, 100) // Returns 100 transactions
	Transactions(ctx context.Context, start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error)

	// AccountTransactions Get transactions associated with an account.
	// Start is a version number. Nil for most recent transactions.
	// Limit is a number of transactions to return. 'about a hundred' by default.
	//
	//	client.AccountTransactions(AccountOne, 0, 2)   // Returns 2 transactions for 0x1
	//	client.AccountTransactions(AccountOne, 1, 100) // Returns 100 transactions for 0x1
	AccountTransactions(ctx context.Context, address AccountAddress, start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error)

	// SubmitTransaction Submits an already signed transaction to the blockchain
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
	//	rawTxn, _ := client.BuildTransaction(sender.AccountAddress(), txnPayload)
	//	signedTxn, _ := sender.SignTransaction(rawTxn)
	//	submitResponse, err := client.SubmitTransaction(signedTxn)
	SubmitTransaction(ctx context.Context, signedTransaction *SignedTransaction) (data *api.SubmitTransactionResponse, err error)

	// BatchSubmitTransaction submits a collection of signed transactions to the network in a single request
	//
	// It will return the responses in the same order as the input transactions that failed.  If the response is empty, then
	// all transactions succeeded.
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
	//	rawTxn, _ := client.BuildTransaction(sender.AccountAddress(), txnPayload)
	//	signedTxn, _ := sender.SignTransaction(rawTxn)
	//	submitResponse, err := client.BatchSubmitTransaction([]*SignedTransaction{signedTxn})
	BatchSubmitTransaction(ctx context.Context, signedTxns []*SignedTransaction) (response *api.BatchSubmitTransactionResponse, err error)

	// SimulateTransaction Simulates a raw transaction without sending it to the blockchain
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
	//	rawTxn, _ := client.BuildTransaction(sender.AccountAddress(), txnPayload)
	//	simResponse, err := client.SimulateTransaction(rawTxn, sender)
	SimulateTransaction(ctx context.Context, rawTxn *RawTransaction, sender TransactionSigner, options ...any) (data []*api.UserTransaction, err error)

	// GetChainId Retrieves the ChainId of the network
	// Note this will be cached forever, or taken directly from the config
	GetChainId(ctx context.Context) (chainId uint8, err error)

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
	BuildTransaction(ctx context.Context, sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error)

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
	BuildTransactionMultiAgent(ctx context.Context, sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransactionWithData, err error)

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
	BuildSignAndSubmitTransaction(ctx context.Context, sender TransactionSigner, payload TransactionPayload, options ...any) (data *api.SubmitTransactionResponse, err error)

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
	View(ctx context.Context, payload *ViewPayload, ledgerVersion ...uint64) (vals []any, err error)

	// EstimateGasPrice Retrieves the gas estimate from the network.
	EstimateGasPrice(ctx context.Context) (info EstimateGasInfo, err error)

	// AccountAPTBalance retrieves the APT balance in the account
	AccountAPTBalance(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (uint64, error)

	// NodeAPIHealthCheck checks if the node is within durationSecs of the current time, if not provided the node default is used
	NodeAPIHealthCheck(ctx context.Context, durationSecs ...uint64) (api.HealthCheckResponse, error)
}

// AptosFaucetClient is an interface for all functionality on the Client that is Faucet related.  Its main implementation
// is [FaucetClient]
type ExposedAptosFaucetClient interface {
	// Fund Uses the faucet to fund an address, only applies to non-production networks
	Fund(ctx context.Context, address AccountAddress, amount uint64) error
}

// AptosIndexerClient is an interface for all functionality on the Client that is Indexer related.  Its main implementation
// is [ExposedIndexerClient]
type ExposedAptosIndexerClient interface {

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
	QueryIndexer(ctx context.Context, query any, variables map[string]any, options ...graphql.Option) error

	// GetProcessorStatus returns the ledger version up to which the processor has processed
	GetProcessorStatus(ctx context.Context, processorName string) (uint64, error)

	// GetCoinBalances gets the balances of all coins associated with a given address
	GetCoinBalances(ctx context.Context, address AccountAddress) ([]CoinBalance, error)
}

// Client is a facade over the multiple types of underlying clients, as the user doesn't actually care where the data
// comes from.  It will be then handled underneath
//
// To create a new client, please use [NewClient].  An example below for Devnet:
//
//	client := NewClient(DevnetConfig)
//
// Implements AptosClient
type ExposedClient struct {
	nodeClient    *ExposedNodeClient
	faucetClient  *ExposedFaucetClient
	indexerClient *ExposedIndexerClient
}

// NewClient Creates a new client with a specific network config that can be extended in the future
func NewExposedClient(config NetworkConfig, options ...any) (client *ExposedClient, err error) {
	var httpClient *http.Client = nil
	for i, arg := range options {
		switch value := arg.(type) {
		case *http.Client:
			if httpClient != nil {
				err = fmt.Errorf("NewClient only accepts one http.Client")
				return
			}
			httpClient = value
		default:
			err = fmt.Errorf("NewClient arg %d bad type %T", i+1, arg)
			return
		}
	}
	var nodeClient *ExposedNodeClient
	if httpClient == nil {
		nodeClient, err = NewExposedNodeClient(config.NodeUrl, config.ChainId)
	} else {
		nodeClient, err = NewExposedNodeClientWithHttpClient(config.NodeUrl, config.ChainId, httpClient)
	}
	if err != nil {
		return nil, err
	}
	// Indexer may not be present
	var indexerClient *ExposedIndexerClient = nil
	if config.IndexerUrl != "" {
		indexerClient = NewExposedIndexerClient(nodeClient.client, config.IndexerUrl)
	}

	// Faucet may not be present
	var faucetClient *ExposedFaucetClient = nil
	if config.FaucetUrl != "" {
		faucetClient, err = NewExposedFaucetClient(nodeClient, config.FaucetUrl)
		if err != nil {
			return nil, err
		}
	}

	// Fetch the chain Id if it isn't in the config
	if config.ChainId == 0 {
		_, _ = nodeClient.GetChainId(context.Background())
	}

	client = &ExposedClient{
		nodeClient,
		faucetClient,
		indexerClient,
	}
	return
}

// SetHeader sets the header for all future requests
//
//	client.SetHeader("Authorization", "Bearer abcde")
func (client *ExposedClient) SetHeader(key string, value string) {
	client.nodeClient.SetHeader(key, value)
}

// RemoveHeader removes the header from being automatically set all future requests.
//
//	client.RemoveHeader("Authorization")
func (client *ExposedClient) RemoveHeader(key string) {
	client.nodeClient.RemoveHeader(key)
}

// Info Retrieves the node info about the network and it's current state
func (client *ExposedClient) Info(ctx context.Context) (info NodeInfo, err error) {
	return client.nodeClient.Info(ctx)
}

// Account Retrieves information about the account such as [SequenceNumber] and [crypto.AuthenticationKey]
func (client *ExposedClient) Account(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (info AccountInfo, err error) {
	return client.nodeClient.Account(ctx, address, ledgerVersion...)
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
func (client *ExposedClient) AccountResource(ctx context.Context, address AccountAddress, resourceType string, ledgerVersion ...uint64) (data map[string]any, err error) {
	return client.nodeClient.AccountResource(ctx, address, resourceType, ledgerVersion...)
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
func (client *ExposedClient) AccountResources(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceInfo, err error) {
	return client.nodeClient.AccountResources(ctx, address, ledgerVersion...)
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (client *ExposedClient) AccountResourcesBCS(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceRecord, err error) {
	return client.nodeClient.AccountResourcesBCS(ctx, address, ledgerVersion...)
}

// BlockByHeight fetches a block by height
//
//	block, _ := client.BlockByHeight(1, false)
//
// Can also fetch with transactions
//
//	block, _ := client.BlockByHeight(1, true)
func (client *ExposedClient) BlockByHeight(ctx context.Context, blockHeight uint64, withTransactions bool) (data *api.Block, err error) {
	return client.nodeClient.BlockByHeight(ctx, blockHeight, withTransactions)
}

// BlockByVersion fetches a block by ledger version
//
//	block, _ := client.BlockByVersion(123, false)
//
// Can also fetch with transactions
//
//	block, _ := client.BlockByVersion(123, true)
func (client *ExposedClient) BlockByVersion(ctx context.Context, ledgerVersion uint64, withTransactions bool) (data *api.Block, err error) {
	return client.nodeClient.BlockByVersion(ctx, ledgerVersion, withTransactions)
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
func (client *ExposedClient) TransactionByHash(ctx context.Context, txnHash string) (data *api.Transaction, err error) {
	return client.nodeClient.TransactionByHash(ctx, txnHash)
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
func (client *ExposedClient) TransactionByVersion(ctx context.Context, version uint64) (data *api.CommittedTransaction, err error) {
	return client.nodeClient.TransactionByVersion(ctx, version)
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
func (client *ExposedClient) PollForTransactions(ctx context.Context, txnHashes []string, options ...any) error {
	return client.nodeClient.PollForTransactions(ctx, txnHashes, options...)
}

// WaitForTransaction Do a long-GET for one transaction and wait for it to complete
//
//	data, err := client.WaitForTransaction("0x1234")
func (client *ExposedClient) WaitForTransaction(ctx context.Context, txnHash string, options ...any) (data *api.UserTransaction, err error) {
	return client.nodeClient.WaitForTransaction(ctx, txnHash, options...)
}

// Transactions Get recent transactions.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
//
//	client.Transactions(0, 2)   // Returns 2 transactions
//	client.Transactions(1, 100) // Returns 100 transactions
func (client *ExposedClient) Transactions(ctx context.Context, start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error) {
	return client.nodeClient.Transactions(ctx, start, limit)
}

// AccountTransactions Get transactions associated with an account.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
//
//	client.AccountTransactions(AccountOne, 0, 2)   // Returns 2 transactions for 0x1
//	client.AccountTransactions(AccountOne, 1, 100) // Returns 100 transactions for 0x1
func (client *ExposedClient) AccountTransactions(ctx context.Context, address AccountAddress, start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error) {
	return client.nodeClient.AccountTransactions(ctx, address, start, limit)
}

// SubmitTransaction Submits an already signed transaction to the blockchain
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
//	rawTxn, _ := client.BuildTransaction(sender.AccountAddress(), txnPayload)
//	signedTxn, _ := sender.SignTransaction(rawTxn)
//	submitResponse, err := client.SubmitTransaction(signedTxn)
func (client *ExposedClient) SubmitTransaction(ctx context.Context, signedTransaction *SignedTransaction) (data *api.SubmitTransactionResponse, err error) {
	return client.nodeClient.SubmitTransaction(ctx, signedTransaction)
}

// BatchSubmitTransaction submits a collection of signed transactions to the network in a single request
//
// It will return the responses in the same order as the input transactions that failed.  If the response is empty, then
// all transactions succeeded.
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
//	rawTxn, _ := client.BuildTransaction(sender.AccountAddress(), txnPayload)
//	signedTxn, _ := sender.SignTransaction(rawTxn)
//	submitResponse, err := client.BatchSubmitTransaction([]*SignedTransaction{signedTxn})
func (client *ExposedClient) BatchSubmitTransaction(ctx context.Context, signedTxns []*SignedTransaction) (response *api.BatchSubmitTransactionResponse, err error) {
	return client.nodeClient.BatchSubmitTransaction(ctx, signedTxns)
}

// SimulateTransaction Simulates a raw transaction without sending it to the blockchain
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
//	rawTxn, _ := client.BuildTransaction(sender.AccountAddress(), txnPayload)
//	simResponse, err := client.SimulateTransaction(rawTxn, sender)
func (client *ExposedClient) SimulateTransaction(ctx context.Context, rawTxn *RawTransaction, sender TransactionSigner, options ...any) (data []*api.UserTransaction, err error) {
	return client.nodeClient.SimulateTransaction(ctx, rawTxn, sender, options...)
}

// GetChainId Retrieves the ChainId of the network
// Note this will be cached forever, or taken directly from the config
func (client *ExposedClient) GetChainId(ctx context.Context) (chainId uint8, err error) {
	return client.nodeClient.GetChainId(ctx)
}

// Fund Uses the faucet to fund an address, only applies to non-production networks
func (client *ExposedClient) Fund(ctx context.Context, address AccountAddress, amount uint64) error {
	return client.faucetClient.Fund(ctx, address, amount)
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
func (client *ExposedClient) BuildTransaction(ctx context.Context, sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error) {
	return client.nodeClient.BuildTransaction(ctx, sender, payload, options...)
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
func (client *ExposedClient) BuildTransactionMultiAgent(ctx context.Context, sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransactionWithData, err error) {
	return client.nodeClient.BuildTransactionMultiAgent(ctx, sender, payload, options...)
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
func (client *ExposedClient) BuildSignAndSubmitTransaction(ctx context.Context, sender TransactionSigner, payload TransactionPayload, options ...any) (data *api.SubmitTransactionResponse, err error) {
	return client.nodeClient.BuildSignAndSubmitTransaction(ctx, sender, payload, options...)
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
func (client *ExposedClient) View(ctx context.Context, payload *ViewPayload, ledgerVersion ...uint64) (vals []any, err error) {
	return client.nodeClient.View(ctx, payload, ledgerVersion...)
}

// EstimateGasPrice Retrieves the gas estimate from the network.
func (client *ExposedClient) EstimateGasPrice(ctx context.Context) (info EstimateGasInfo, err error) {
	return client.nodeClient.EstimateGasPrice(ctx)
}

// AccountAPTBalance retrieves the APT balance in the account
func (client *ExposedClient) AccountAPTBalance(ctx context.Context, address AccountAddress, ledgerVersion ...uint64) (uint64, error) {
	return client.nodeClient.AccountAPTBalance(ctx, address, ledgerVersion...)
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
func (client *ExposedClient) QueryIndexer(ctx context.Context, query any, variables map[string]any, options ...graphql.Option) error {
	return client.indexerClient.Query(ctx, query, variables, options...)
}

// GetProcessorStatus returns the ledger version up to which the processor has processed
func (client *ExposedClient) GetProcessorStatus(ctx context.Context, processorName string) (uint64, error) {
	return client.indexerClient.GetProcessorStatus(ctx, processorName)
}

// GetCoinBalances gets the balances of all coins associated with a given address
func (client *ExposedClient) GetCoinBalances(ctx context.Context, address AccountAddress) ([]CoinBalance, error) {
	return client.indexerClient.GetCoinBalances(ctx, address)
}

// NodeAPIHealthCheck checks if the node is within durationSecs of the current time, if not provided the node default is used
func (client *ExposedClient) NodeAPIHealthCheck(ctx context.Context, durationSecs ...uint64) (api.HealthCheckResponse, error) {
	return client.nodeClient.NodeAPIHealthCheck(ctx, durationSecs...)
}
