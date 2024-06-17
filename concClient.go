package aptos

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/hasura/go-graphql-client"
	"io"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// ConcClient is a concurrent client using Go routines, to wrap the existing [Client].  This is a testing ground, and the future go routine APIs will hopefully be merged directly into the [Client].
type ConcClient struct {
	client *Client
}

// ConcResponse is a concurrent response wrapper as a return type for all APIs.  It is meant to specifically be used in channels.
type ConcResponse[T any] struct {
	Result T
	Err    error
}

// NewConcClient creates a [ConcClient]
func NewConcClient(c *Client) (*ConcClient, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	return &ConcClient{client: c}, nil
}

// SetTimeout adjusts the HTTP client timeout
//
//	client.SetTimeout(5 * time.Millisecond)
func (cc *ConcClient) SetTimeout(timeout time.Duration) {
	cc.client.nodeClient.client.Timeout = timeout
}

// TODO: Support multiple output channels?
func fetch[T any](inner func() (T, error), result chan ConcResponse[T]) {
	response, err := inner()
	if err != nil {
		result <- ConcResponse[T]{Err: err}
	} else {
		result <- ConcResponse[T]{Result: response}
	}
	close(result)
}

func (cc *ConcClient) Info(result chan ConcResponse[*NodeInfo]) {
	go fetch(func() (*NodeInfo, error) {
		// TODO: clean up this * vs not *
		info, err := cc.client.Info()
		return &info, err
	}, result)
}

func (cc *ConcClient) Account(result chan ConcResponse[*AccountInfo], address AccountAddress, ledgerVersion ...uint64) {
	go fetch(func() (*AccountInfo, error) {
		// TODO: clean up this * vs not *
		info, err := cc.client.Account(address, ledgerVersion...)
		return &info, err
	}, result)
}

// TODO account resource calls

func (cc *ConcClient) BlockByVersion(result chan ConcResponse[*api.Block], ledgerVersion uint64, withTransactions bool) {
	restUrl := cc.client.nodeClient.baseUrl.JoinPath("blocks/by_version", strconv.FormatUint(ledgerVersion, 10))
	go cc.getBlockCommon(result, restUrl, withTransactions)
}

func (cc *ConcClient) BlockByHeight(result chan ConcResponse[*api.Block], blockHeight uint64, withTransactions bool) {
	restUrl := cc.client.nodeClient.baseUrl.JoinPath("blocks/by_height", strconv.FormatUint(blockHeight, 10))
	go cc.getBlockCommon(result, restUrl, withTransactions)
}

func (cc *ConcClient) getBlockCommon(result chan ConcResponse[*api.Block], restUrl *url.URL, withTransactions bool) {
	params := url.Values{}
	params.Set("with_transactions", strconv.FormatBool(withTransactions))
	restUrl.RawQuery = params.Encode()

	// Fetch block this has to be done serially anyway
	response, err := cc.client.nodeClient.Get(restUrl.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", restUrl.String(), err)
		result <- ConcResponse[*api.Block]{Err: err}
		return
	}

	// Handle Errors TODO: Handle rate-limits, etc.
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		result <- ConcResponse[*api.Block]{Err: err}
		return
	}

	// Read body to JSON
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		result <- ConcResponse[*api.Block]{Err: err}
		return
	}
	_ = response.Body.Close() // We don't care about the error about closing the body
	block := &api.Block{}
	err = json.Unmarshal(blob, block)

	// Now, let's fill in any missing transactions in the block
	numTransactions := block.LastVersion - block.FirstVersion + 1
	retrievedTransactions := uint64(len(block.Transactions))
	remainingTransactions := numTransactions - retrievedTransactions

	if remainingTransactions > 0 {
		extraTransactions := make(chan ConcResponse[[]*api.Transaction], remainingTransactions)
		nextTransaction := *block.Transactions[len(block.Transactions)-1].Version() + 1
		go cc.Transactions(extraTransactions, nextTransaction, remainingTransactions)

		// When I've pulled them all, they're sorted, so we can just append
		concResponse := <-extraTransactions
		close(extraTransactions)
		if concResponse.Err != nil {
			result <- ConcResponse[*api.Block]{Result: block, Err: concResponse.Err}
			return
		}

		block.Transactions = append(block.Transactions, concResponse.Result...)
	}
	result <- ConcResponse[*api.Block]{Result: block}
}

// Transactions fetches transactions concurrently up to the limit
// TODO make optionals
func (cc *ConcClient) Transactions(result chan ConcResponse[[]*api.Transaction], start uint64, limit uint64) {
	const transactionsPageSize = 100
	// If we know both, we can fetch all concurrently
	type Pair struct {
		start uint64
		end   uint64
	}

	if limit > 100 {
		numChannels := limit / transactionsPageSize
		if limit%transactionsPageSize > 0 {
			numChannels++
		}
		channels := make([]chan ConcResponse[[]*api.Transaction], numChannels)
		for i := uint64(0); i*transactionsPageSize < limit; i += 1 {
			channels[i] = make(chan ConcResponse[[]*api.Transaction], 1)
			st := start + i*100
			li := min(transactionsPageSize, limit-i*transactionsPageSize)
			go cc.Transactions(channels[i], st, li)
		}

		responses := make([]*api.Transaction, limit)
		cursor := 0
		for i, ch := range channels {
			response := <-ch
			if response.Err != nil {
				result <- response
				return
			}
			end := cursor + len(response.Result)

			copy(responses[cursor:end], response.Result)
			cursor = end
			close(channels[i])
		}

		// Sort to keep ordering
		sort.Slice(responses, func(i, j int) bool {
			return *responses[i].Version() < *responses[j].Version()
		})
		result <- ConcResponse[[]*api.Transaction]{Result: responses}
	} else {
		response, err := cc.client.Transactions(&start, &limit)
		if err != nil {
			result <- ConcResponse[[]*api.Transaction]{Err: err}
		} else {
			result <- ConcResponse[[]*api.Transaction]{Result: response}
		}
	}
}

// TransactionByHash gets info on a transaction
// The transaction may be pending or recently committed.
func (cc *ConcClient) TransactionByHash(result chan ConcResponse[*api.Transaction], txnHash string) {
	go fetch(func() (*api.Transaction, error) {
		return cc.client.TransactionByHash(txnHash)
	}, result)
}

// PollForTransactions Waits up to 10 seconds for transactions to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
func (cc *ConcClient) PollForTransactions(result chan ConcResponse[bool], txnHashes []string, options ...any) {
	go fetch(func() (bool, error) {
		err := cc.client.PollForTransactions(txnHashes, options...)
		return err == nil, err
	}, result)
}

// WaitForTransaction Do a long-GET for one transaction and wait for it to complete
func (cc *ConcClient) WaitForTransaction(result chan ConcResponse[*api.UserTransaction], txnHash string) {
	go fetch(func() (*api.UserTransaction, error) {
		return cc.client.WaitForTransaction(txnHash)
	}, result)
}

// SubmitTransaction Submits an already signed transaction to the blockchain
func (cc *ConcClient) SubmitTransaction(result chan ConcResponse[*api.SubmitTransactionResponse], signedTransaction *SignedTransaction) {
	go fetch(func() (*api.SubmitTransactionResponse, error) {
		return cc.client.SubmitTransaction(signedTransaction)
	}, result)
}

// GetChainId Retrieves the ChainId of the network
// Note this will be cached forever, or taken directly from the config
func (cc *ConcClient) GetChainId(result chan ConcResponse[*uint8]) {
	go fetch(func() (*uint8, error) {
		val, err := cc.client.GetChainId()
		return &val, err
	}, result)
}

// Fund Uses the faucet to fund an address, only applies to non-production networks
func (cc *ConcClient) Fund(result chan ConcResponse[bool], address AccountAddress, amount uint64) {
	go fetch(func() (bool, error) {
		err := cc.client.Fund(address, amount)
		return err == nil, err
	}, result)
}

func (cc *ConcClient) BuildTransaction(result chan ConcResponse[*RawTransaction], sender AccountAddress, payload TransactionPayload, options ...any) {

	maxGasAmount := uint64(100_000) // Default to 0.001 APT max gas amount
	gasUnitPrice := uint64(100)     // Default to min gas price
	expirationSeconds := int64(300) // Default to 5 minutes
	sequenceNumber := uint64(0)
	haveSequenceNumber := false
	chainId := uint8(0)
	haveChainId := false

	for opti, option := range options {
		switch ovalue := option.(type) {
		case MaxGasAmount:
			maxGasAmount = uint64(ovalue)
		case GasUnitPrice:
			gasUnitPrice = uint64(ovalue)
		case ExpirationSeconds:
			expirationSeconds = int64(ovalue)
			if expirationSeconds < 0 {
				err := errors.New("ExpirationSeconds cannot be less than 0")
				result <- ConcResponse[*RawTransaction]{Err: err}
				return
			}
		case SequenceNumber:
			sequenceNumber = uint64(ovalue)
			haveSequenceNumber = true
		case ChainIdOption:
			chainId = uint8(ovalue)
			haveChainId = true
		default:
			err := fmt.Errorf("BuildTransaction arg [%d] unknown option type %T", opti+4, option)
			result <- ConcResponse[*RawTransaction]{Err: err}
			return
		}
	}

	// Fetch ChainId which may be cached
	chainIdChannel := make(chan ConcResponse[*uint8], 1)
	if !haveChainId {
		go cc.GetChainId(chainIdChannel)
	} else {
		close(chainIdChannel)
	}

	// Fetch sequence number unless provided
	accountChannel := make(chan ConcResponse[*AccountInfo], 1)
	if !haveSequenceNumber {
		go cc.Account(accountChannel, sender)
	} else {
		close(accountChannel)
	}

	// TODO: fetch gas price on-chain
	// TODO: optionally simulate for max gas

	if !haveChainId {
		chainIdResponse := <-chainIdChannel
		if chainIdResponse.Err != nil {
			result <- ConcResponse[*RawTransaction]{Err: chainIdResponse.Err}
			return
		} else {
			chainId = *chainIdResponse.Result
		}
	}

	if !haveSequenceNumber {
		accountResponse := <-accountChannel
		if accountResponse.Err != nil {
			result <- ConcResponse[*RawTransaction]{Err: accountResponse.Err}
			return
		} else {
			num, err := accountResponse.Result.SequenceNumber()
			sequenceNumber = num
			if err != nil {
				result <- ConcResponse[*RawTransaction]{Err: err}
				return
			}
		}
	}

	expirationTimestampSeconds := uint64(time.Now().Unix() + expirationSeconds)

	// Base raw transaction used for all requests
	result <- ConcResponse[*RawTransaction]{
		Result: &RawTransaction{
			Sender:                     sender,
			SequenceNumber:             sequenceNumber,
			Payload:                    payload,
			MaxGasAmount:               maxGasAmount,
			GasUnitPrice:               gasUnitPrice,
			ExpirationTimestampSeconds: expirationTimestampSeconds,
			ChainId:                    chainId,
		}}

}

func (cc *ConcClient) BuildSignAndSubmitTransaction(result chan ConcResponse[*api.SubmitTransactionResponse], sender TransactionSigner, payload TransactionPayload, options ...any) {
	// Build transaction
	buildResult := make(chan ConcResponse[*RawTransaction], 1)
	go cc.BuildTransaction(buildResult, sender.AccountAddress(), payload, options...)

	// Sign transaction
	buildResponse := <-buildResult
	if buildResponse.Err != nil {
		result <- ConcResponse[*api.SubmitTransactionResponse]{Err: buildResponse.Err}
		return
	}
	signedTxn, err := buildResponse.Result.SignedTransaction(sender)
	if err != nil {
		result <- ConcResponse[*api.SubmitTransactionResponse]{Err: err}
		return
	}

	// Submit transaction
	go cc.SubmitTransaction(result, signedTxn)
}

func (cc *ConcClient) View(result chan ConcResponse[[]any], payload *ViewPayload, ledgerVersion ...uint64) {
	go fetch(func() ([]any, error) {
		return cc.client.View(payload, ledgerVersion...)
	}, result)
}

// EstimateGasPrice Retrieves the gas estimate from the network.
func (cc *ConcClient) EstimateGasPrice(result chan ConcResponse[*EstimateGasInfo]) {
	go fetch(func() (*EstimateGasInfo, error) {
		info, err := cc.client.EstimateGasPrice()
		return &info, err
	}, result)
}

// AccountAPTBalance retrieves the APT balance in the account
func (cc *ConcClient) AccountAPTBalance(result chan ConcResponse[*uint64], address AccountAddress) {
	go fetch(func() (*uint64, error) {
		balance, err := cc.client.AccountAPTBalance(address)
		return &balance, err
	}, result)
}

// QueryIndexer will return the query object filled with a response
func (cc *ConcClient) QueryIndexer(result chan ConcResponse[any], query any, variables map[string]any, options ...graphql.Option) {
	go fetch(func() (any, error) {
		err := cc.client.QueryIndexer(query, variables, options...)
		if err != nil {
			return nil, err
		} else {
			return query, nil
		}
	}, result)
}
