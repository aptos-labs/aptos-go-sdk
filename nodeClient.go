package aptos

import (
	"bytes"
	"context"
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
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

const (
	DefaultMaxGasAmount      = uint64(100_000) // Default to 0.001 APT max gas amount
	DefaultGasUnitPrice      = uint64(100)     // Default to min gas price
	DefaultExpirationSeconds = uint64(300)     // Default to 5 minutes
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
func (rc *NodeClient) Info() (NodeInfo, error) {
	info, err := Get[NodeInfo](rc, rc.baseUrl.String())
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
func (rc *NodeClient) Account(address AccountAddress, ledgerVersion ...uint64) (AccountInfo, error) {
	au := rc.baseUrl.JoinPath("accounts", address.String())
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	info, err := Get[AccountInfo](rc, au.String())
	if err != nil {
		return info, fmt.Errorf("get account info api err: %w", err)
	}
	return info, nil
}

// AccountResource fetches a resource for an account into a JSON-like map[string]any.
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
//
// For fetching raw Move structs as BCS, See #AccountResourceBCS
func (rc *NodeClient) AccountResource(address AccountAddress, resourceType string, ledgerVersion ...uint64) (map[string]any, error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resource", resourceType)
	// TODO: offer a list of known-good resourceType string constants
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	data, err := Get[map[string]any](rc, au.String())
	if err != nil {
		return nil, fmt.Errorf("get resource api err: %w", err)
	}
	return data, nil
}

// AccountResources fetches resources for an account into a JSON-like map[string]any in AccountResourceInfo.Data
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
// For fetching raw Move structs as BCS, See #AccountResourcesBCS
func (rc *NodeClient) AccountResources(address AccountAddress, ledgerVersion ...uint64) ([]AccountResourceInfo, error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	resources, err := Get[[]AccountResourceInfo](rc, au.String())
	if err != nil {
		return nil, fmt.Errorf("get resources api err: %w", err)
	}
	return resources, err
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
// Optionally, a ledgerVersion can be given to get the account state at a specific ledger version
func (rc *NodeClient) AccountResourcesBCS(address AccountAddress, ledgerVersion ...uint64) ([]AccountResourceRecord, error) {
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
	resources := bcs.DeserializeSequence[AccountResourceRecord](deserializer)
	err = deserializer.Error()
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// AccountModule fetches a single account module's bytecode and ABI from on-chain state.
func (rc *NodeClient) AccountModule(address AccountAddress, moduleName string, ledgerVersion ...uint64) (*api.MoveBytecode, error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "module", moduleName)
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	data, err := Get[*api.MoveBytecode](rc, au.String())
	if err != nil {
		return nil, fmt.Errorf("get module api err: %w", err)
	}
	return data, nil
}

// EntryFunctionWithArgs generates an EntryFunction from on-chain Module ABI, and converts simple inputs to BCS encoded ones.
func (rc *NodeClient) EntryFunctionWithArgs(moduleAddress AccountAddress, moduleName string, functionName string, typeArgs []any, args []any, options ...any) (*EntryFunction, error) {
	// TODO: This should be cached / we should be able to take in an ABI
	module, err := rc.AccountModule(moduleAddress, moduleName)
	if err != nil {
		return nil, err
	}

	return EntryFunctionFromAbi(module.Abi, moduleAddress, moduleName, functionName, typeArgs, args, options...)
}

// BlockByVersion gets a block by a transaction's version number
//
// Note that this is not the same as a block's height.
//
// The function will fetch all transactions in the block if withTransactions is true.
func (rc *NodeClient) BlockByVersion(ledgerVersion uint64, withTransactions bool) (*api.Block, error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_version", strconv.FormatUint(ledgerVersion, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

// BlockByHeight gets a block by block height
//
// The function will fetch all transactions in the block if withTransactions is true.
func (rc *NodeClient) BlockByHeight(blockHeight uint64, withTransactions bool) (*api.Block, error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_height", strconv.FormatUint(blockHeight, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
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
func (rc *NodeClient) TransactionByHash(txnHash string) (*api.Transaction, error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_hash", txnHash)
	data, err := Get[*api.Transaction](rc, restUrl.String())
	if err != nil {
		return data, fmt.Errorf("get transaction api err: %w", err)
	}
	return data, nil
}

// WaitTransactionByHash waits for a transaction to be confirmed by its hash.
// This function allows you to monitor the status of a transaction until it is finalized.
func (rc *NodeClient) WaitTransactionByHash(txnHash string) (*api.Transaction, error) {
	restUrl := rc.baseUrl.JoinPath("transactions/wait_by_hash", txnHash)
	data, err := Get[*api.Transaction](rc, restUrl.String())
	if err != nil {
		return data, fmt.Errorf("get transaction api err: %w", err)
	}
	return data, nil
}

// TransactionByVersion gets info on a transaction by version number
// The transaction will have been committed.  The response will not be of the type [api.PendingTransaction].
func (rc *NodeClient) TransactionByVersion(version uint64) (*api.CommittedTransaction, error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_version", strconv.FormatUint(version, 10))
	data, err := Get[*api.CommittedTransaction](rc, restUrl.String())
	if err != nil {
		return data, fmt.Errorf("get transaction api err: %w", err)
	}
	return data, nil
}

// getBlockCommon is a helper function for fetching a block by version or height
//
// It will fetch all the transactions associated with the block if withTransactions is true.
func (rc *NodeClient) getBlockCommon(restUrl *url.URL, withTransactions bool) (*api.Block, error) {
	params := url.Values{}
	params.Set("with_transactions", strconv.FormatBool(withTransactions))
	restUrl.RawQuery = params.Encode()

	// Fetch block
	block, err := Get[*api.Block](rc, restUrl.String())
	if err != nil {
		return block, fmt.Errorf("get block api err: %w", err)
	}

	// Return early if we don't need transactions
	if !withTransactions {
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
	return block, nil
}

func getTransactionPollOptions(defaultPeriod, defaultTimeout time.Duration, options ...any) (time.Duration, time.Duration, error) {
	period := defaultPeriod
	timeout := defaultTimeout
	for i, arg := range options {
		switch value := arg.(type) {
		case PollPeriod:
			period = time.Duration(value)
		case PollTimeout:
			timeout = time.Duration(value)
		default:
			return period, timeout, fmt.Errorf("PollForTransactions arg %d bad type %T", i+1, arg)
		}
	}
	return period, timeout, nil
}

// PollForTransaction waits up to 10 seconds for a transaction to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
// Not just a degenerate case of PollForTransactions, it may return additional information for the single transaction polled.
func (rc *NodeClient) PollForTransaction(hash string, options ...any) (*api.UserTransaction, error) {
	// Wait for the transaction to be done
	txn, err := rc.WaitTransactionByHash(hash)
	if err == nil && txn.Type == api.TransactionVariantUser {
		return txn.UserTransaction()
	}

	// Poll for the transaction to be done
	period, timeout, err := getTransactionPollOptions(100*time.Millisecond, 10*time.Second, options...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("PollForTransaction timeout")
		case <-ticker.C:
			txn, err := rc.TransactionByHash(hash)
			if err != nil {
				continue
			}
			switch txn.Type {
			case api.TransactionVariantPending:
				// not done yet!
				continue
			case api.TransactionVariantUser:
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
				if txn.Type != api.TransactionVariantPending {
					// done!
					delete(hashSet, hash)
					slog.Debug("txn done", "hash", hash)
				}
			}
		}
	}
	return nil
}

// WaitForTransaction does a long-GET for one transaction and wait for it to complete.
// Initially poll at 10 Hz for up to 1 second if node replies with 404 (wait for txn to propagate).
//
// Optional arguments:
//   - PollPeriod: time.Duration, how often to poll for the transaction. Default 100ms.
//   - PollTimeout: time.Duration, how long to wait for the transaction. Default 10s.
func (rc *NodeClient) WaitForTransaction(txnHash string, options ...any) (*api.UserTransaction, error) {
	return rc.PollForTransaction(txnHash, options...)
}

// Transactions Get recent transactions.
//
// Arguments:
//   - start is a version number. Nil for most recent transactions.
//   - limit is a number of transactions to return. 'about a hundred' by default.
func (rc *NodeClient) Transactions(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
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
func (rc *NodeClient) AccountTransactions(account AccountAddress, start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
	return rc.handleTransactions(start, limit, func(txns *[]*api.CommittedTransaction) uint64 {
		// It will always be a UserTransaction, no other type will come from the API
		userTxn, _ := ((*txns)[0]).UserTransaction()
		return userTxn.SequenceNumber - 1
	}, func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
		return rc.accountTransactionsInner(account, start, limit)
	})
}

// EventsByHandle retrieves events by event handle and field name for a given account.
//
// Arguments:
//   - account - The account address to get events for
//   - eventHandle - The event handle struct tag
//   - fieldName - The field in the event handle struct
//   - start - The starting sequence number. nil for most recent events
//   - limit - The number of events to return, 100 by default
func (rc *NodeClient) EventsByHandle(
	account AccountAddress,
	eventHandle string,
	fieldName string,
	start *uint64,
	limit *uint64,
) ([]*api.Event, error) {
	basePath := fmt.Sprintf("accounts/%s/events/%s/%s",
		account.String(),
		eventHandle,
		fieldName)

	baseUrl := rc.baseUrl.JoinPath(basePath)

	const eventsPageSize = 100
	var effectiveLimit uint64
	if limit == nil {
		effectiveLimit = eventsPageSize
	} else {
		effectiveLimit = *limit
	}

	if effectiveLimit <= eventsPageSize {
		params := url.Values{}
		if start != nil {
			params.Set("start", strconv.FormatUint(*start, 10))
		}
		params.Set("limit", strconv.FormatUint(effectiveLimit, 10))

		requestUrl := *baseUrl
		requestUrl.RawQuery = params.Encode()

		data, err := Get[[]*api.Event](rc, requestUrl.String())
		if err != nil {
			return nil, fmt.Errorf("get events api err: %w", err)
		}
		return data, nil
	}

	pages := (effectiveLimit + eventsPageSize - 1) / eventsPageSize
	channels := make([]chan ConcResponse[[]*api.Event], pages)

	for i := range pages {
		channels[i] = make(chan ConcResponse[[]*api.Event], 1)
		var pageStart *uint64
		if start != nil {
			value := *start + (i * eventsPageSize)
			pageStart = &value
		}
		pageLimit := min(eventsPageSize, effectiveLimit-(i*eventsPageSize))

		go fetch(func() ([]*api.Event, error) {
			params := url.Values{}
			if pageStart != nil {
				params.Set("start", strconv.FormatUint(*pageStart, 10))
			}
			params.Set("limit", strconv.FormatUint(pageLimit, 10))

			requestUrl := *baseUrl
			requestUrl.RawQuery = params.Encode()

			events, err := Get[[]*api.Event](rc, requestUrl.String())
			if err != nil {
				return nil, fmt.Errorf("get events api err: %w", err)
			}
			return events, nil
		}, channels[i])
	}

	events := make([]*api.Event, 0, effectiveLimit)
	for i, ch := range channels {
		response := <-ch
		if response.Err != nil {
			return nil, response.Err
		}
		events = append(events, response.Result...)
		close(channels[i])
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].SequenceNumber < events[j].SequenceNumber
	})

	return events, nil
}

// EventsByCreationNumber retrieves events by creation number for a given account.
//
// Arguments:
//   - account - The account address to get events for
//   - creationNumber - The creation number of the event
//   - start - The starting sequence number. nil for most recent events
//   - limit - The number of events to return, 100 by default
func (rc *NodeClient) EventsByCreationNumber(
	account AccountAddress,
	creationNumber string,
	start *uint64,
	limit *uint64,
) ([]*api.Event, error) {
	basePath := fmt.Sprintf("accounts/%s/events/%s",
		account.String(),
		creationNumber)

	baseUrl := rc.baseUrl.JoinPath(basePath)

	const eventsPageSize = 100
	var effectiveLimit uint64
	if limit == nil {
		effectiveLimit = eventsPageSize
	} else {
		effectiveLimit = *limit
	}

	if effectiveLimit <= eventsPageSize {
		params := url.Values{}
		if start != nil {
			params.Set("start", strconv.FormatUint(*start, 10))
		}
		params.Set("limit", strconv.FormatUint(effectiveLimit, 10))

		requestUrl := *baseUrl
		requestUrl.RawQuery = params.Encode()

		data, err := Get[[]*api.Event](rc, requestUrl.String())
		if err != nil {
			return nil, fmt.Errorf("get events api err: %w", err)
		}
		return data, nil
	}

	pages := (effectiveLimit + eventsPageSize - 1) / eventsPageSize
	channels := make([]chan ConcResponse[[]*api.Event], pages)

	for i := range pages {
		channels[i] = make(chan ConcResponse[[]*api.Event], 1)
		var pageStart *uint64
		if start != nil {
			value := *start + (i * eventsPageSize)
			pageStart = &value
		}
		pageLimit := min(eventsPageSize, effectiveLimit-(i*eventsPageSize))

		go fetch(func() ([]*api.Event, error) {
			params := url.Values{}
			if pageStart != nil {
				params.Set("start", strconv.FormatUint(*pageStart, 10))
			}
			params.Set("limit", strconv.FormatUint(pageLimit, 10))

			requestUrl := *baseUrl
			requestUrl.RawQuery = params.Encode()

			events, err := Get[[]*api.Event](rc, requestUrl.String())
			if err != nil {
				return nil, fmt.Errorf("get events api err: %w", err)
			}
			return events, nil
		}, channels[i])
	}

	events := make([]*api.Event, 0, effectiveLimit)
	for i, ch := range channels {
		response := <-ch
		if response.Err != nil {
			return nil, response.Err
		}
		events = append(events, response.Result...)
		close(channels[i])
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].SequenceNumber < events[j].SequenceNumber
	})

	return events, nil
}

// handleTransactions is a helper function for fetching transactions
//
// It will fetch the transactions from the node in a single request if possible, otherwise it will fetch them concurrently.
func (rc *NodeClient) handleTransactions(
	start *uint64,
	limit *uint64,
	getNext func(txns *[]*api.CommittedTransaction) uint64,
	getTxns func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error),
) ([]*api.CommittedTransaction, error) {
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
		}

		newStart := getNext(&txns)
		newLength := actualLimit - numTxns
		extra, err := rc.transactionsConcurrent(newStart, newLength, getTxns)
		if err != nil {
			return nil, err
		}

		return append(extra, txns...), nil
	}

	// If we know the start, just pull one page
	return getTxns(start, nil)
}

// transactionsConcurrent fetches the transactions from the node concurrently
//
// It will fetch the transactions concurrently if the limit is greater than the page size, otherwise it will fetch them in a single request.
func (rc *NodeClient) transactionsConcurrent(
	start uint64,
	limit uint64,
	getTxns func(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error),
) ([]*api.CommittedTransaction, error) {
	const transactionsPageSize = 100
	// If the limit is  greater than the page size, we need to fetch concurrently, otherwise not
	if limit > transactionsPageSize {
		numChannels := limit / transactionsPageSize
		if limit%transactionsPageSize > 0 {
			numChannels++
		}

		// Concurrently fetch all the transactions by the page size
		channels := make([]chan ConcResponse[[]*api.CommittedTransaction], numChannels)
		for i := uint64(0); i*transactionsPageSize < limit; i++ {
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
				return nil, response.Err
			}
			responses = append(responses, response.Result...)
			close(channels[i])
		}

		// Sort to keep ordering
		sort.Slice(responses, func(i, j int) bool {
			return responses[i].Version() < responses[j].Version()
		})
		return responses, nil
	}

	response, err := getTxns(&start, &limit)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// transactionsInner fetches the transactions from the node in a single request
func (rc *NodeClient) transactionsInner(start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
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
	data, err := Get[[]*api.CommittedTransaction](rc, au.String())
	if err != nil {
		return data, fmt.Errorf("get transactions api err: %w", err)
	}
	return data, nil
}

// accountTransactionsInner fetches the transactions from the node in a single request for a single account
func (rc *NodeClient) accountTransactionsInner(account AccountAddress, start *uint64, limit *uint64) ([]*api.CommittedTransaction, error) {
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

	data, err := Get[[]*api.CommittedTransaction](rc, au.String())
	if err != nil {
		return data, fmt.Errorf("get account transactions api err: %w", err)
	}
	return data, nil
}

// SubmitTransaction submits a signed transaction to the network
func (rc *NodeClient) SubmitTransaction(signedTxn *SignedTransaction) (*api.SubmitTransactionResponse, error) {
	sblob, err := bcs.Serialize(signedTxn)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions")
	data, err := Post[*api.SubmitTransactionResponse](rc, au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("submit transaction api err: %w", err)
	}
	return data, nil
}

// BatchSubmitTransaction submits a collection of signed transactions to the network in a single request
//
// It will return the responses in the same order as the input transactions that failed.  If the response is empty, then
// all transactions succeeded.
func (rc *NodeClient) BatchSubmitTransaction(signedTxns []*SignedTransaction) (*api.BatchSubmitTransactionResponse, error) {
	sblob, err := bcs.SerializeSequenceOnly(signedTxns)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions/batch")
	response, err := Post[*api.BatchSubmitTransactionResponse](rc, au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("submit transaction api err: %w", err)
	}
	return response, nil
}

// SimulateTransaction simulates a transaction
func (rc *NodeClient) SimulateTransaction(rawTxn *RawTransaction, sender TransactionSigner, options ...any) ([]*api.UserTransaction, error) {
	// build authenticator for simulation
	auth := sender.SimulationAuthenticator()

	// generate signed transaction for simulation (with zero signature)
	signedTxn, err := rawTxn.SignedTransactionWithAuthenticator(auth)
	if err != nil {
		return nil, err
	}

	return rc.simulateTransactionInner(signedTxn, options...)
}

// SimulateTransactionMultiAgent simulates a transaction as fee payer or multi agent
func (rc *NodeClient) SimulateTransactionMultiAgent(rawTxn *RawTransactionWithData, sender TransactionSigner, options ...any) ([]*api.UserTransaction, error) {
	var feePayer *AccountAddress
	var additionalSigners []AccountAddress

	for _, option := range options {
		switch ovalue := option.(type) {
		case FeePayer:
			feePayer = ovalue
		case AdditionalSigners:
			additionalSigners = ovalue
		default:
			// Silently ignore unknown arguments, as there are types that are passed down further
		}
	}

	var signedTxn *SignedTransaction
	var ok bool
	if feePayer != nil {
		senderAuth := sender.SimulationAuthenticator()
		feePayerAuth := crypto.NoAccountAuthenticator()
		additionalSignersAuth := make([]crypto.AccountAuthenticator, len(additionalSigners))
		for i := range additionalSigners {
			additionalSignersAuth[i] = *crypto.NoAccountAuthenticator()
		}
		signedTxn, ok = rawTxn.ToFeePayerSignedTransaction(senderAuth, feePayerAuth, additionalSignersAuth)
		if !ok {
			return nil, errors.New("failed to convert fee payer signer to signed transaction")
		}
	} else {
		senderAuth := sender.SimulationAuthenticator()
		additionalSignersAuth := make([]crypto.AccountAuthenticator, len(additionalSigners))
		for i := range additionalSigners {
			additionalSignersAuth[i] = *crypto.NoAccountAuthenticator()
		}
		signedTxn, ok = rawTxn.ToMultiAgentSignedTransaction(senderAuth, additionalSignersAuth)
		if !ok {
			return nil, errors.New("failed to convert multi agent signer to signed transaction")
		}
	}

	return rc.simulateTransactionInner(signedTxn, options...)
}

func (rc *NodeClient) simulateTransactionInner(signedTxn *SignedTransaction, options ...any) ([]*api.UserTransaction, error) {
	sblob, err := bcs.Serialize(signedTxn)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions/simulate")

	// parse simulate tx options
	params := url.Values{}
	for _, arg := range options {
		switch value := arg.(type) {
		case EstimateGasUnitPrice:
			params.Set("estimate_gas_unit_price", strconv.FormatBool(bool(value)))
		case EstimateMaxGasAmount:
			params.Set("estimate_max_gas_amount", strconv.FormatBool(bool(value)))
		case EstimatePrioritizedGasUnitPrice:
			params.Set("estimate_prioritized_gas_unit_price", strconv.FormatBool(bool(value)))
		default:
			// Silently ignore unknown arguments, as there are multiple intakes
		}
	}
	if len(params) != 0 {
		au.RawQuery = params.Encode()
	}

	data, err := Post[[]*api.UserTransaction](rc, au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("simulate transaction api err: %w", err)
	}

	return data, nil
}

// GetChainId gets the chain ID of the network
func (rc *NodeClient) GetChainId() (uint8, error) {
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
func (rc *NodeClient) BuildTransaction(sender AccountAddress, payload TransactionPayload, options ...any) (*RawTransaction, error) {
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
			expirationSeconds = uint64(ovalue)
		case SequenceNumber:
			sequenceNumber = uint64(ovalue)
			haveSequenceNumber = true
		case ChainIdOption:
			chainId = uint8(ovalue)
			haveChainId = true
		default:
			return nil, fmt.Errorf("BuildTransaction arg [%d] unknown option type %T", opti+4, option)
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
func (rc *NodeClient) BuildTransactionMultiAgent(sender AccountAddress, payload TransactionPayload, options ...any) (*RawTransactionWithData, error) {
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
			expirationSeconds = uint64(ovalue)
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
			return nil, fmt.Errorf("APTTransferTransaction arg [%d] unknown option type %T", opti+4, option)
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
	}

	return &RawTransactionWithData{
		Variant: MultiAgentRawTransactionWithDataVariant,
		Inner: &MultiAgentRawTransactionWithData{
			RawTxn:           rawTxn,
			SecondarySigners: additionalSigners,
		},
	}, nil
}

func (rc *NodeClient) buildTransactionInner(
	sender AccountAddress,
	payload TransactionPayload,
	maxGasAmount uint64,
	gasUnitPrice uint64,
	haveGasUnitPrice bool,
	expirationSeconds uint64,
	sequenceNumber uint64,
	haveSequenceNumber bool,
	chainId uint8,
	haveChainId bool,
) (*RawTransaction, error) {
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

	now, err := util.IntToU64(int(time.Now().Unix()))
	if err != nil {
		return nil, err
	}
	expirationTimestampSeconds := now + expirationSeconds

	// Base raw transaction used for all requests
	return &RawTransaction{
		Sender:                     sender,
		SequenceNumber:             sequenceNumber,
		Payload:                    payload,
		MaxGasAmount:               maxGasAmount,
		GasUnitPrice:               gasUnitPrice,
		ExpirationTimestampSeconds: expirationTimestampSeconds,
		ChainId:                    chainId,
	}, nil
}

// View calls a view function on the blockchain and returns the return value of the function
func (rc *NodeClient) View(payload *ViewPayload, ledgerVersion ...uint64) ([]any, error) {
	serializer := bcs.Serializer{}
	payload.MarshalBCS(&serializer)
	err := serializer.Error()
	if err != nil {
		return nil, err
	}
	sblob := serializer.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("view")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}

	data, err := Post[[]any](rc, au.String(), ContentTypeAptosViewFunctionBcs, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("view function api err: %w", err)
	}
	return data, nil
}

// EstimateGasPrice estimates the gas price given on-chain data
// TODO: add caching for some period of time
func (rc *NodeClient) EstimateGasPrice() (EstimateGasInfo, error) {
	au := rc.baseUrl.JoinPath("estimate_gas_price")
	info, err := Get[EstimateGasInfo](rc, au.String())
	if err != nil {
		return info, fmt.Errorf("estimate gas price err: %w", err)
	}
	return info, nil
}

// AccountAPTBalance fetches the balance of an account of APT.  Response is in octas or 1/10^8 APT.
func (rc *NodeClient) AccountAPTBalance(account AccountAddress, ledgerVersion ...uint64) (uint64, error) {
	accountBytes, err := bcs.Serialize(&account)
	if err != nil {
		return 0, err
	}
	values, err := rc.View(&ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "coin",
		},
		Function: "balance",
		ArgTypes: []TypeTag{AptosCoinTypeTag},
		Args:     [][]byte{accountBytes},
	}, ledgerVersion...)
	if err != nil {
		return 0, err
	}
	str, ok := values[0].(string)
	if !ok {
		return 0, errors.New("account balance err: could not convert account bytes")
	}
	return StrToUint64(str)
}

// NodeAPIHealthCheck performs a health check on the node
//
// Returns a HealthCheckResponse if successful, returns error if not.
func (rc *NodeClient) NodeAPIHealthCheck(durationSecs ...uint64) (api.HealthCheckResponse, error) {
	au := rc.baseUrl.JoinPath("-/healthy")
	if len(durationSecs) > 0 {
		params := url.Values{}
		params.Set("duration_secs", strconv.FormatUint(durationSecs[0], 10))
		au.RawQuery = params.Encode()
	}
	return Get[api.HealthCheckResponse](rc, au.String())
}

// NodeHealthCheck performs a health check on the node
//
// Returns a HealthCheckResponse if successful, returns error if not.
//
// Deprecated: Use NodeAPIHealthCheck instead
func (rc *NodeClient) NodeHealthCheck(durationSecs ...uint64) (api.HealthCheckResponse, error) {
	return rc.NodeAPIHealthCheck(durationSecs...)
}

// BuildSignAndSubmitTransaction builds, signs, and submits a transaction to the network
func (rc *NodeClient) BuildSignAndSubmitTransaction(sender TransactionSigner, payload TransactionPayload, options ...any) (*api.SubmitTransactionResponse, error) {
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

// Get makes a GET request to the endpoint and parses the response into the given type with JSON
func Get[T any](rc *NodeClient, getUrl string) (T, error) {
	var out T
	req, err := http.NewRequest(http.MethodGet, getUrl, nil)
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
	defer response.Body.Close()
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		return out, fmt.Errorf("error getting response data, %w", err)
	}
	err = json.Unmarshal(blob, &out)
	if err != nil {
		return out, err
	}
	return out, nil
}

// GetBCS makes a GET request to the endpoint and parses the response into the given type with BCS
func (rc *NodeClient) GetBCS(getUrl string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, getUrl, nil)
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
		return nil, fmt.Errorf("GET %s, %w", getUrl, err)
	}
	if response.StatusCode >= 400 {
		return nil, NewHttpError(response)
	}
	defer response.Body.Close()
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error getting response data, %w", err)
	}
	return blob, nil
}

// Post makes a POST request to the endpoint with the given body and parses the response into the given type with JSON
func Post[T any](rc *NodeClient, postUrl string, contentType string, body io.Reader) (T, error) {
	var data T
	if body == nil {
		body = http.NoBody
	}
	req, err := http.NewRequest(http.MethodPost, postUrl, body)
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
	defer response.Body.Close()
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return data, err
	}

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
