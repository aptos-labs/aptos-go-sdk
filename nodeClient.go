package aptos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const (
	DefaultMaxGasAmount      = uint64(100_000) // Default to 0.001 APT max gas amount
	DefaultGasUnitPrice      = uint64(100)     // Default to min gas price
	DefaultExpirationSeconds = int64(300)      // Default to 5 minutes
)

// For Content-Type header when POST-ing a Transaction

// ContentTypeAptosSignedTxnBcs header for sending BCS transaction payloads
const ContentTypeAptosSignedTxnBcs = "application/x.aptos.signed_transaction+bcs"

// ContentTypeAptosViewFunctionBcs header for sending BCS view function payloads
const ContentTypeAptosViewFunctionBcs = "application/x.aptos.view_function+bcs"

// NodeClient is a client for interacting with an Aptos node API
type NodeClient struct {
	client  *http.Client      // HTTP client to use for requests
	baseUrl *url.URL          // Base URL of the node e.g. https://fullnode.testnet.aptoslabs.com/v1
	chainId uint8             // Chain ID of the network e.g. 2 for Testnet
	headers map[string]string // Headers to be added to every transaction
}

// NewNodeClient creates a new client for interacting with an Aptos node API
func NewNodeClient(rpcUrl string, chainId uint8) (*NodeClient, error) {
	// Set cookie jar so cookie stickiness applies to connections
	// TODO Add appropriate suffix list
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	defaultClient := &http.Client{
		Jar:     jar,
		Timeout: 60 * time.Second,
	}

	return NewNodeClientWithHttpClient(rpcUrl, chainId, defaultClient)
}

// NewNodeClientWithHttpClient creates a new client for interacting with an Aptos node API with a custom http.Client
func NewNodeClientWithHttpClient(rpcUrl string, chainId uint8, client *http.Client) (*NodeClient, error) {
	baseUrl, err := url.Parse(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RPC url '%s': %w", rpcUrl, err)
	}
	return &NodeClient{
		client:  client,
		baseUrl: baseUrl,
		chainId: chainId,
		headers: make(map[string]string),
	}, nil
}

// SetTimeout adjusts the HTTP client timeout
//
//	client.SetTimeout(5 * time.Millisecond)
func (rc *NodeClient) SetTimeout(timeout time.Duration) {
	rc.client.Timeout = timeout
}

// SetHeader sets the header for all future requests
//
//	client.SetHeader("Authorization", "Bearer abcde")
func (rc *NodeClient) SetHeader(key string, value string) {
	rc.headers[key] = value
}

// RemoveHeader removes the header from being automatically set all future requests.
//
//	client.RemoveHeader("Authorization")
func (rc *NodeClient) RemoveHeader(key string) {
	delete(rc.headers, key)
}

// Info gets general information about the blockchain
func (rc *NodeClient) Info() (info NodeInfo, err error) {
	info, err = Get[NodeInfo](rc, rc.baseUrl.String())
	if err != nil {
		return info, fmt.Errorf("get node info api err: %w", err)
	}

	// Cache the ChainId for later calls, because performance
	rc.chainId = info.ChainId
	return info, err
}

// Account gets information about an account for a given address
//
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
func (rc *NodeClient) Account(address AccountAddress, ledgerVersion ...uint64) (info AccountInfo, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String())
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	info, err = Get[AccountInfo](rc, au.String())
	if err != nil {
		return info, fmt.Errorf("get account info api err: %w", err)
	}
	return info, nil
}

// AccountResource fetches a resource for an account into a JSON-like map[string]any.
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
//
// For fetching raw Move structs as BCS, See #AccountResourceBCS
func (rc *NodeClient) AccountResource(address AccountAddress, resourceType string, ledgerVersion ...uint64) (data map[string]any, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resource", resourceType)
	// TODO: offer a list of known-good resourceType string constants
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	data, err = Get[map[string]any](rc, au.String())
	if err != nil {
		return nil, fmt.Errorf("get resource api err: %w", err)
	}
	return data, nil
}

// AccountResources fetches resources for an account into a JSON-like map[string]any in AccountResourceInfo.Data
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
// For fetching raw Move structs as BCS, See #AccountResourcesBCS
func (rc *NodeClient) AccountResources(address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceInfo, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	resources, err = Get[[]AccountResourceInfo](rc, au.String())
	if err != nil {
		return nil, fmt.Errorf("get resources api err: %w", err)
	}
	return resources, err
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
func (rc *NodeClient) AccountResourcesBCS(address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceRecord, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	blob, err := rc.GetBCS(au.String())
	if err != nil {
		return nil, err
	}

	deserializer := bcs.NewDeserializer(blob)
	// See resource_test.go TestMoveResourceBCS
	resources = bcs.DeserializeSequence[AccountResourceRecord](deserializer)
	return
}

// TransactionByHash gets info on a transaction
// The transaction may be pending or recently committed.  If the transaction is a [api.PendingTransaction], then it is
// still in the mempool.  If the transaction is any other type, it has been committed.
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
func (rc *NodeClient) TransactionByHash(txnHash string) (data *api.Transaction, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_hash", txnHash)
	data, err = Get[*api.Transaction](rc, restUrl.String())
	if err != nil {
		return data, fmt.Errorf("get transaction api err: %w", err)
	}
	return data, nil
}

// TransactionByVersion gets info on a transaction by version number
// The transaction will have been committed.  The response will not be of the type [api.PendingTransaction].
func (rc *NodeClient) TransactionByVersion(version uint64) (data *api.CommittedTransaction, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_version", strconv.FormatUint(version, 10))
	data, err = Get[*api.CommittedTransaction](rc, restUrl.String())
	if err != nil {
		return data, fmt.Errorf("get transaction api err: %w", err)
	}
	return data, nil
}

// BlockByVersion gets a block by a transaction's version number
//
// Note that this is not the same as a block's height.
//
// The function will fetch all transactions in the block if withTransactions is true.
func (rc *NodeClient) BlockByVersion(ledgerVersion uint64, withTransactions bool) (data *api.Block, err error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_version", strconv.FormatUint(ledgerVersion, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

// BlockByHeight gets a block by block height
//
// The function will fetch all transactions in the block if withTransactions is true.
func (rc *NodeClient) BlockByHeight(blockHeight uint64, withTransactions bool) (data *api.Block, err error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_height", strconv.FormatUint(blockHeight, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

// getBlockCommon is a helper function for fetching a block by version or height
//
// It will fetch all the transactions associated with the block if withTransactions is true.
func (rc *NodeClient) getBlockCommon(restUrl *url.URL, withTransactions bool) (block *api.Block, err error) {
	params := url.Values{}
	params.Set("with_transactions", strconv.FormatBool(withTransactions))
	restUrl.RawQuery = params.Encode()

	// Fetch block
	block, err = Get[*api.Block](rc, restUrl.String())
	if err != nil {
		return block, fmt.Errorf("get block api err: %w", err)
	}

	// Return early if we don't need transactions
	if withTransactions == false {
		return block, nil
	}

	// Now, let's fill in any missing transactions in the block
	if block.Transactions == nil {
		block.Transactions = make([]*api.CommittedTransaction, 0)
	}

	// Now, let's fill in any missing transactions in the block
	numTransactions := block.LastVersion - block.FirstVersion + 1
	retrievedTransactions := uint64(len(block.Transactions))

	// Transaction is always not pending, so it will never be nil
	cursor := block.Transactions[len(block.Transactions)-1].Version()

	// TODO: I maybe should pull these concurrently, but not for now
	for retrievedTransactions < numTransactions {
		numToPull := numTransactions - retrievedTransactions
		transactions, innerError := rc.Transactions(&cursor, &numToPull)
		if innerError != nil {
			// We will still return the block, since we did so much work for it
			return block, innerError
		}

		// Add transactions to the list
		block.Transactions = append(block.Transactions, transactions...)
		retrievedTransactions = uint64(len(block.Transactions))
		cursor = block.Transactions[len(block.Transactions)-1].Version()
	}
	return
}

// WaitForTransaction does a long-GET for one transaction and wait for it to complete.
// Initially poll at 10 Hz for up to 1 second if node replies with 404 (wait for txn to propagate).
//
// Optional arguments:
//   - PollPeriod: time.Duration, how often to poll for the transaction. Default 100ms.
//   - PollTimeout: time.Duration, how long to wait for the transaction. Default 10s.
func (rc *NodeClient) WaitForTransaction(txnHash string, options ...any) (data *api.UserTransaction, err error) {
	return rc.PollForTransaction(txnHash, options...)
}

// PollPeriod is an option to PollForTransactions
type PollPeriod time.Duration

// PollTimeout is an option to PollForTransactions
type PollTimeout time.Duration

func getTransactionPollOptions(defaultPeriod, defaultTimeout time.Duration, options ...any) (period time.Duration, timeout time.Duration, err error) {
	period = defaultPeriod
	timeout = defaultTimeout
	for i, arg := range options {
		switch value := arg.(type) {
		case PollPeriod:
			period = time.Duration(value)
		case PollTimeout:
			timeout = time.Duration(value)
		default:
			err = fmt.Errorf("PollForTransactions arg %d bad type %T", i+1, arg)
			return
		}
	}
	return
}

// PollForTransaction waits up to 10 seconds for a transaction to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
// Not just a degenerate case of PollForTransactions, it may return additional information for the single transaction polled.
func (rc *NodeClient) PollForTransaction(hash string, options ...any) (*api.UserTransaction, error) {
	period, timeout, err := getTransactionPollOptions(100*time.Millisecond, 10*time.Second, options...)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	deadline := start.Add(timeout)
	for {
		if time.Now().After(deadline) {
			return nil, errors.New("PollForTransaction timeout")
		}
		time.Sleep(period)
		txn, err := rc.TransactionByHash(hash)
		if err == nil {
			if txn.Type == api.TransactionVariantPending {
				// not done yet!
			} else if txn.Type == api.TransactionVariantUser {
				// done!
				slog.Debug("txn done", "hash", hash)
				return txn.UserTransaction()
			}
		}
	}
}

// PollForTransactions waits up to 10 seconds for transactions to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
func (rc *NodeClient) PollForTransactions(txnHashes []string, options ...any) error {
	period, timeout, err := getTransactionPollOptions(100*time.Millisecond, 10*time.Second, options...)
	if err != nil {
		return err
	}
	hashSet := make(map[string]bool, len(txnHashes))
	for _, hash := range txnHashes {
		hashSet[hash] = true
	}
	start := time.Now()
	deadline := start.Add(timeout)
	for len(hashSet) > 0 {
		if time.Now().After(deadline) {
			return errors.New("PollForTransactions timeout")
		}
		time.Sleep(period)
		for _, hash := range txnHashes {
			if !hashSet[hash] {
				// already done
				continue
			}
			txn, err := rc.TransactionByHash(hash)
			if err == nil {
				if txn.Type == api.TransactionVariantPending {
					// not done yet!
				} else if txn.Type == api.TransactionVariantUser {
					// done!
					delete(hashSet, hash)
					slog.Debug("txn done", "hash", hash)
				}
			}
		}
	}
	return nil
}

// Transactions Get recent transactions.
//
// Arguments:
//   - start is a version number. Nil for most recent transactions.
//   - limit is a number of transactions to return. 'about a hundred' by default.
func (rc *NodeClient) Transactions(start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error) {
	return rc.handleTransactions(start, limit, func(txns *[]*api.CommittedTransaction) uint64 {
		txn := (*txns)[len(*txns)-1]
		return txn.Version()
	}, func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
		return rc.transactionsInner(start, limit)
	})
}

// AccountTransactions Get recent transactions for an account
//
// Arguments:
//   - start is a version number. Nil for most recent transactions.
//   - limit is a number of transactions to return. 'about a hundred' by default.
func (rc *NodeClient) AccountTransactions(account AccountAddress, start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error) {
	return rc.handleTransactions(start, limit, func(txns *[]*api.CommittedTransaction) uint64 {
		// It will always be a UserTransaction, no other type will come from the API
		userTxn, _ := ((*txns)[0]).UserTransaction()
		return userTxn.SequenceNumber - 1
	}, func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
		return rc.accountTransactionsInner(account, start, limit)
	})
}

// handleTransactions is a helper function for fetching transactions
//
// It will fetch the transactions from the node in a single request if possible, otherwise it will fetch them concurrently.
func (rc *NodeClient) handleTransactions(
	start *uint64,
	limit *uint64,
	getNext func(txns *[]*api.CommittedTransaction) uint64,
	getTxns func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error),
) (data []*api.CommittedTransaction, err error) {
	// Can only pull everything in parallel if a start and a limit is handled
	if start != nil && limit != nil {
		return rc.transactionsConcurrent(*start, *limit, getTxns)
	} else if limit != nil {
		// If we don't know the start, we can only pull one page first, then handle the rest
		// Note that, this actually pulls the last page first, then goes backwards
		actualLimit := *limit
		txns, err := getTxns(nil, limit)
		if err != nil {
			return nil, err
		}

		// If we have enough transactions, return otherwise, pull the rest
		numTxns := uint64(len(txns))
		if numTxns >= actualLimit {
			return txns, nil
		} else {
			newStart := getNext(&txns)
			newLength := actualLimit - numTxns
			extra, err := rc.transactionsConcurrent(newStart, newLength, getTxns)
			if err != nil {
				return nil, err
			}

			return append(extra, txns...), nil
		}
	} else {
		// If we know the start, just pull one page
		return getTxns(start, nil)
	}
}

// transactionsConcurrent fetches the transactions from the node concurrently
//
// It will fetch the transactions concurrently if the limit is greater than the page size, otherwise it will fetch them in a single request.
func (rc *NodeClient) transactionsConcurrent(
	start uint64,
	limit uint64,
	getTxns func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error),
) (data []*api.CommittedTransaction, err error) {
	const transactionsPageSize = 100
	// If we know both, we can fetch all concurrently
	type Pair struct {
		start uint64 // inclusive
		end   uint64 // exclusive
	}

	// If the limit is  greater than the page size, we need to fetch concurrently, otherwise not
	if limit > transactionsPageSize {
		numChannels := limit / transactionsPageSize
		if limit%transactionsPageSize > 0 {
			numChannels++
		}

		// Concurrently fetch all the transactions by the page size
		channels := make([]chan ConcResponse[[]*api.CommittedTransaction], numChannels)
		for i := uint64(0); i*transactionsPageSize < limit; i += 1 {
			channels[i] = make(chan ConcResponse[[]*api.CommittedTransaction], 1)
			st := start + i*100 // TODO: allow page size to be configured
			li := min(transactionsPageSize, limit-i*transactionsPageSize)
			go fetch(func() ([]*api.CommittedTransaction, error) {
				return rc.transactionsConcurrent(st, li, getTxns)
			}, channels[i])
		}

		// Collect all the responses
		responses := make([]*api.CommittedTransaction, 0)
		for i, ch := range channels {
			response := <-ch
			if response.Err != nil {
				return nil, err
			}
			responses = append(responses, response.Result...)
			close(channels[i])
		}

		// Sort to keep ordering
		sort.Slice(responses, func(i, j int) bool {
			return responses[i].Version() < responses[j].Version()
		})
		return responses, nil
	} else {
		response, err := getTxns(&start, &limit)
		if err != nil {
			return nil, err
		} else {
			return response, nil
		}
	}
}

// transactionsInner fetches the transactions from the node in a single request
func (rc *NodeClient) transactionsInner(start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error) {
	au := rc.baseUrl.JoinPath("transactions")
	params := url.Values{}
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) != 0 {
		au.RawQuery = params.Encode()
	}
	data, err = Get[[]*api.CommittedTransaction](rc, au.String())
	if err != nil {
		return data, fmt.Errorf("get transactions api err: %w", err)
	}
	return data, nil
}

// accountTransactionsInner fetches the transactions from the node in a single request for a single account
func (rc *NodeClient) accountTransactionsInner(account AccountAddress, start *uint64, limit *uint64) (data []*api.CommittedTransaction, err error) {
	au := rc.baseUrl.JoinPath(fmt.Sprintf("accounts/%s/transactions", account.String()))
	params := url.Values{}
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) != 0 {
		au.RawQuery = params.Encode()
	}

	data, err = Get[[]*api.CommittedTransaction](rc, au.String())
	if err != nil {
		return data, fmt.Errorf("get account transactions api err: %w", err)
	}
	return data, nil
}

// SubmitTransaction submits a signed transaction to the network
func (rc *NodeClient) SubmitTransaction(signedTxn *SignedTransaction) (data *api.SubmitTransactionResponse, err error) {
	sblob, err := bcs.Serialize(signedTxn)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions")
	data, err = Post[*api.SubmitTransactionResponse](rc, au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("submit transaction api err: %w", err)
	}
	return data, nil
}

// BatchSubmitTransaction submits a collection of signed transactions to the network in a single request
//
// It will return the responses in the same order as the input transactions that failed.  If the response is empty, then
// all transactions succeeded.
func (rc *NodeClient) BatchSubmitTransaction(signedTxns []*SignedTransaction) (response *api.BatchSubmitTransactionResponse, err error) {
	sblob, err := bcs.SerializeSequenceOnly(signedTxns)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions/batch")
	response, err = Post[*api.BatchSubmitTransactionResponse](rc, au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("submit transaction api err: %w", err)
	}
	return response, nil
}

// EstimateGasUnitPrice estimates the gas unit price for a transaction
type EstimateGasUnitPrice bool

// EstimateMaxGasAmount estimates the max gas amount for a transaction
type EstimateMaxGasAmount bool

// EstimatePrioritizedGasUnitPrice estimates the prioritized gas unit price for a transaction
type EstimatePrioritizedGasUnitPrice bool

// SimulateTransaction simulates a transaction
//
// TODO: This needs to support RawTransactionWithData
// TODO: Support multikey simulation
func (rc *NodeClient) SimulateTransaction(rawTxn *RawTransaction, sender TransactionSigner, options ...any) (data []*api.UserTransaction, err error) {
	// build authenticator for simulation
	derivationScheme := sender.PubKey().Scheme()
	switch derivationScheme {
	case crypto.MultiEd25519Scheme:
	case crypto.MultiKeyScheme:
		// todo: add support for multikey simulation on the node
		return nil, fmt.Errorf("currently unsupported sender derivation scheme %v", derivationScheme)
	}
	auth := sender.SimulationAuthenticator()

	// generate signed transaction for simulation (with zero signature)
	signedTxn, err := rawTxn.SignedTransactionWithAuthenticator(auth)
	if err != nil {
		return nil, err
	}

	sblob, err := bcs.Serialize(signedTxn)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions/simulate")

	// parse simulate tx options
	params := url.Values{}
	for i, arg := range options {
		switch value := arg.(type) {
		case EstimateGasUnitPrice:
			params.Set("estimate_gas_unit_price", strconv.FormatBool(bool(value)))
		case EstimateMaxGasAmount:
			params.Set("estimate_max_gas_amount", strconv.FormatBool(bool(value)))
		case EstimatePrioritizedGasUnitPrice:
			params.Set("estimate_prioritized_gas_unit_price", strconv.FormatBool(bool(value)))
		default:
			err = fmt.Errorf("SimulateTransaction arg %d bad type %T", i+1, arg)
			return
		}
	}
	if len(params) != 0 {
		au.RawQuery = params.Encode()
	}

	data, err = Post[[]*api.UserTransaction](rc, au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("simulate transaction api err: %w", err)
	}

	return data, nil
}

// GetChainId gets the chain ID of the network
func (rc *NodeClient) GetChainId() (chainId uint8, err error) {
	if rc.chainId == 0 {
		// Calling Info will cache the ChainId
		info, err := rc.Info()
		if err != nil {
			return 0, err
		}
		return info.ChainId, nil
	}
	return rc.chainId, nil
}

// MaxGasAmount will set the max gas amount in gas units for a transaction
type MaxGasAmount uint64

// GasUnitPrice will set the gas unit price in octas (1/10^8 APT) for a transaction
type GasUnitPrice uint64

// ExpirationSeconds will set the number of seconds from the current time to expire a transaction
type ExpirationSeconds int64

// FeePayer will set the fee payer for a transaction
type FeePayer *AccountAddress

// AdditionalSigners will set the additional signers for a transaction
type AdditionalSigners []AccountAddress

// SequenceNumber will set the sequence number for a transaction
type SequenceNumber uint64

// ChainIdOption will set the chain ID for a transaction
// TODO: This one may want to be removed / renamed?
type ChainIdOption uint8

// BuildTransaction builds a raw transaction for signing for a single signer
//
// For MultiAgent and FeePayer transactions use [NodeClient.BuildTransactionMultiAgent]
//
// Accepts options:
//   - [MaxGasAmount]
//   - [GasUnitPrice]
//   - [ExpirationSeconds]
//   - [SequenceNumber]
//   - [ChainIdOption]
func (rc *NodeClient) BuildTransaction(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error) {

	maxGasAmount := DefaultMaxGasAmount
	gasUnitPrice := DefaultGasUnitPrice
	expirationSeconds := DefaultExpirationSeconds
	sequenceNumber := uint64(0)
	haveSequenceNumber := false
	chainId := uint8(0)
	haveChainId := false
	haveGasUnitPrice := false

	for opti, option := range options {
		switch ovalue := option.(type) {
		case MaxGasAmount:
			maxGasAmount = uint64(ovalue)
		case GasUnitPrice:
			gasUnitPrice = uint64(ovalue)
			haveGasUnitPrice = true
		case ExpirationSeconds:
			expirationSeconds = int64(ovalue)
			if expirationSeconds < 0 {
				err = errors.New("ExpirationSeconds cannot be less than 0")
				return nil, err
			}
		case SequenceNumber:
			sequenceNumber = uint64(ovalue)
			haveSequenceNumber = true
		case ChainIdOption:
			chainId = uint8(ovalue)
			haveChainId = true
		default:
			err = fmt.Errorf("BuildTransaction arg [%d] unknown option type %T", opti+4, option)
			return nil, err
		}
	}

	return rc.buildTransactionInner(sender, payload, maxGasAmount, gasUnitPrice, haveGasUnitPrice, expirationSeconds, sequenceNumber, haveSequenceNumber, chainId, haveChainId)
}

// BuildTransactionMultiAgent builds a raw transaction for signing with fee payer or multi-agent
//
// For single signer transactions use [NodeClient.BuildTransaction]
//
// Accepts options:
//   - [MaxGasAmount]
//   - [GasUnitPrice]
//   - [ExpirationSeconds]
//   - [SequenceNumber]
//   - [ChainIdOption]
//   - [FeePayer]
//   - [AdditionalSigners]
func (rc *NodeClient) BuildTransactionMultiAgent(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxnImpl *RawTransactionWithData, err error) {

	maxGasAmount := DefaultMaxGasAmount
	gasUnitPrice := DefaultGasUnitPrice
	expirationSeconds := DefaultExpirationSeconds
	sequenceNumber := uint64(0)
	haveSequenceNumber := false
	chainId := uint8(0)
	haveChainId := false
	haveGasUnitPrice := false

	var feePayer *AccountAddress
	var additionalSigners []AccountAddress

	for opti, option := range options {
		switch ovalue := option.(type) {
		case MaxGasAmount:
			maxGasAmount = uint64(ovalue)
		case GasUnitPrice:
			gasUnitPrice = uint64(ovalue)
			haveGasUnitPrice = true
		case ExpirationSeconds:
			expirationSeconds = int64(ovalue)
			if expirationSeconds < 0 {
				err = errors.New("ExpirationSeconds cannot be less than 0")
				return nil, err
			}
		case SequenceNumber:
			sequenceNumber = uint64(ovalue)
			haveSequenceNumber = true
		case ChainIdOption:
			chainId = uint8(ovalue)
			haveChainId = true
		case FeePayer:
			feePayer = ovalue
		case AdditionalSigners:
			additionalSigners = ovalue
		default:
			err = fmt.Errorf("APTTransferTransaction arg [%d] unknown option type %T", opti+4, option)
			return nil, err
		}
	}

	// Build the base raw transaction
	rawTxn, err := rc.buildTransactionInner(sender, payload, maxGasAmount, gasUnitPrice, haveGasUnitPrice, expirationSeconds, sequenceNumber, haveSequenceNumber, chainId, haveChainId)
	if err != nil {
		return nil, err
	}

	// Based on the options, choose which to use
	if feePayer != nil {
		return &RawTransactionWithData{
			Variant: MultiAgentWithFeePayerRawTransactionWithDataVariant,
			Inner: &MultiAgentWithFeePayerRawTransactionWithData{
				RawTxn:           rawTxn,
				FeePayer:         feePayer,
				SecondarySigners: additionalSigners,
			},
		}, nil
	} else {
		return &RawTransactionWithData{
			Variant: MultiAgentRawTransactionWithDataVariant,
			Inner: &MultiAgentRawTransactionWithData{
				RawTxn:           rawTxn,
				SecondarySigners: additionalSigners,
			},
		}, nil
	}
}

func (rc *NodeClient) buildTransactionInner(
	sender AccountAddress,
	payload TransactionPayload,
	maxGasAmount uint64,
	gasUnitPrice uint64,
	haveGasUnitPrice bool,
	expirationSeconds int64,
	sequenceNumber uint64,
	haveSequenceNumber bool,
	chainId uint8,
	haveChainId bool,
) (rawTxn *RawTransaction, err error) {
	// Fetch requirements concurrently, and then consume them

	// Fetch GasUnitPrice which may be cached
	var gasPriceErrChannel chan error
	if !haveGasUnitPrice {
		gasPriceErrChannel = make(chan error, 1)
		go func() {
			gasPriceEstimation, innerErr := rc.EstimateGasPrice()
			if innerErr != nil {
				gasPriceErrChannel <- innerErr
			} else {
				gasUnitPrice = gasPriceEstimation.GasEstimate
				gasPriceErrChannel <- nil
			}
			close(gasPriceErrChannel)
		}()
	}

	// Fetch ChainId which may be cached
	var chainIdErrChannel chan error
	if !haveChainId {
		if rc.chainId == 0 {
			chainIdErrChannel = make(chan error, 1)
			go func() {
				chain, innerErr := rc.GetChainId()
				if innerErr != nil {
					chainIdErrChannel <- innerErr
				} else {
					chainId = chain
					chainIdErrChannel <- nil
				}
				close(chainIdErrChannel)
			}()
		} else {
			chainId = rc.chainId
		}
	}

	// Fetch sequence number unless provided
	var accountErrChannel chan error
	if !haveSequenceNumber {
		accountErrChannel = make(chan error, 1)
		go func() {
			account, innerErr := rc.Account(sender)
			if innerErr != nil {
				accountErrChannel <- innerErr
				close(accountErrChannel)
				return
			}
			seqNo, innerErr := account.SequenceNumber()
			if innerErr != nil {
				accountErrChannel <- innerErr
				close(accountErrChannel)
				return
			}
			sequenceNumber = seqNo
			accountErrChannel <- nil
			close(accountErrChannel)
		}()
	}

	// TODO: optionally simulate for max gas
	// Wait on the errors
	if chainIdErrChannel != nil {
		chainIdErr := <-chainIdErrChannel
		if chainIdErr != nil {
			return nil, chainIdErr
		}
	}
	if accountErrChannel != nil {
		accountErr := <-accountErrChannel
		if accountErr != nil {
			return nil, accountErr
		}
	}
	if gasPriceErrChannel != nil {
		gasPriceErr := <-gasPriceErrChannel
		if gasPriceErr != nil {
			return nil, gasPriceErr
		}
	}

	expirationTimestampSeconds := uint64(time.Now().Unix() + expirationSeconds)

	// Base raw transaction used for all requests
	rawTxn = &RawTransaction{
		Sender:                     sender,
		SequenceNumber:             sequenceNumber,
		Payload:                    payload,
		MaxGasAmount:               maxGasAmount,
		GasUnitPrice:               gasUnitPrice,
		ExpirationTimestampSeconds: expirationTimestampSeconds,
		ChainId:                    chainId,
	}
	return rawTxn, nil
}

// ViewPayload is a payload for a view function
type ViewPayload struct {
	Module   ModuleId  // ModuleId of the View function e.g. 0x1::coin
	Function string    // Name of the View function e.g. balance
	ArgTypes []TypeTag // TypeTags of the type arguments
	Args     [][]byte  // Arguments to the function encoded in BCS
}

func (vp *ViewPayload) MarshalBCS(ser *bcs.Serializer) {
	vp.Module.MarshalBCS(ser)
	ser.WriteString(vp.Function)
	bcs.SerializeSequence(vp.ArgTypes, ser)
	ser.Uleb128(uint32(len(vp.Args)))
	for _, a := range vp.Args {
		ser.WriteBytes(a)
	}
}

// View calls a view function on the blockchain and returns the return value of the function
func (rc *NodeClient) View(payload *ViewPayload, ledgerVersion ...uint64) (data []any, err error) {
	serializer := bcs.Serializer{}
	payload.MarshalBCS(&serializer)
	err = serializer.Error()
	if err != nil {
		return
	}
	sblob := serializer.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("view")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}

	data, err = Post[[]any](rc, au.String(), ContentTypeAptosViewFunctionBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("view function api err: %w", err)
	}
	return data, nil
}

// EstimateGasPrice estimates the gas price given on-chain data
// TODO: add caching for some period of time
func (rc *NodeClient) EstimateGasPrice() (info EstimateGasInfo, err error) {
	au := rc.baseUrl.JoinPath("estimate_gas_price")
	info, err = Get[EstimateGasInfo](rc, au.String())
	if err != nil {
		return info, fmt.Errorf("estimate gas price err: %w", err)
	}
	return info, nil
}

// AccountAPTBalance fetches the balance of an account of APT.  Response is in octas or 1/10^8 APT.
func (rc *NodeClient) AccountAPTBalance(account AccountAddress) (balance uint64, err error) {
	accountBytes, err := bcs.Serialize(&account)
	if err != nil {
		return 0, err
	}
	values, err := rc.View(&ViewPayload{Module: ModuleId{
		Address: AccountOne,
		Name:    "coin",
	},
		Function: "balance",
		ArgTypes: []TypeTag{AptosCoinTypeTag},
		Args:     [][]byte{accountBytes},
	})
	if err != nil {
		return 0, err
	}
	return StrToUint64(values[0].(string))
}

// BuildSignAndSubmitTransaction builds, signs, and submits a transaction to the network
func (rc *NodeClient) BuildSignAndSubmitTransaction(sender TransactionSigner, payload TransactionPayload, options ...any) (data *api.SubmitTransactionResponse, err error) {
	rawTxn, err := rc.BuildTransaction(sender.AccountAddress(), payload, options...)
	if err != nil {
		return nil, err
	}
	signedTxn, err := rawTxn.SignedTransaction(sender)
	if err != nil {
		return nil, err
	}
	return rc.SubmitTransaction(signedTxn)
}

// NodeHealthCheck performs a health check on the node
//
// Returns a HealthCheckResponse if successful, returns error if not.
func (rc *NodeClient) NodeHealthCheck(durationSecs ...uint64) (api.HealthCheckResponse, error) {
	au := rc.baseUrl.JoinPath("-/healthy")
	if len(durationSecs) > 0 {
		params := url.Values{}
		params.Set("duration_secs", strconv.FormatUint(durationSecs[0], 10))
		au.RawQuery = params.Encode()
	}
	return Get[api.HealthCheckResponse](rc, au.String())
}

// Get makes a GET request to the endpoint and parses the response into the given type with JSON
func Get[T any](rc *NodeClient, getUrl string) (out T, err error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set(ClientHeader, ClientHeaderValue)

	// Set all preset headers
	for key, value := range rc.headers {
		req.Header.Set(key, value)
	}

	response, err := rc.client.Do(req)
	if err != nil {
		err = fmt.Errorf("GET %s, %w", getUrl, err)
		return out, err
	}

	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return out, err
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		return out, fmt.Errorf("error getting response data, %w", err)
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &out)
	if err != nil {
		return out, err
	}
	return out, nil
}

// GetBCS makes a GET request to the endpoint and parses the response into the given type with BCS
func (rc *NodeClient) GetBCS(getUrl string) (out []byte, err error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-bcs")
	req.Header.Set(ClientHeader, ClientHeaderValue)

	// Set all preset headers
	for key, value := range rc.headers {
		req.Header.Set(key, value)
	}

	response, err := rc.client.Do(req)
	if err != nil {
		err = fmt.Errorf("GET %s, %w", getUrl, err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	return blob, nil
}

// Post makes a POST request to the endpoint with the given body and parses the response into the given type with JSON
func Post[T any](rc *NodeClient, postUrl string, contentType string, body io.Reader) (data T, err error) {
	if body == nil {
		body = http.NoBody
	}
	req, err := http.NewRequest("POST", postUrl, body)
	if err != nil {
		return data, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(ClientHeader, ClientHeaderValue)

	// Set all preset headers
	for key, value := range rc.headers {
		req.Header.Set(key, value)
	}

	response, err := rc.client.Do(req)
	if err != nil {
		err = fmt.Errorf("POST %s, %w", postUrl, err)
		return data, err
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return data, err
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return data, err
	}
	_ = response.Body.Close()

	err = json.Unmarshal(blob, &data)
	return data, err
}

// ConcResponse is a concurrent response wrapper as a return type for all APIs.  It is meant to specifically be used in channels.
type ConcResponse[T any] struct {
	Result T
	Err    error
}

// fetch is a helper function to fetch data concurrently with a channel and support error and value together
// TODO: Support multiple output channels?
func fetch[T any](inner func() (T, error), result chan ConcResponse[T]) {
	response, err := inner()
	if err != nil {
		result <- ConcResponse[T]{Err: err}
	} else {
		result <- ConcResponse[T]{Result: response}
	}
}
