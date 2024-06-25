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

type NodeClient struct {
	client  *http.Client
	baseUrl *url.URL
	chainId uint8
}

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

func NewNodeClientWithHttpClient(rpcUrl string, chainId uint8, client *http.Client) (*NodeClient, error) {
	baseUrl, err := url.Parse(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RPC url '%s': %w", rpcUrl, err)
	}
	return &NodeClient{
		client:  client,
		baseUrl: baseUrl,
		chainId: chainId,
	}, nil
}

func (rc *NodeClient) Info() (info NodeInfo, err error) {
	info, err = Get[NodeInfo](rc, rc.baseUrl.String())
	if err != nil {
		return info, fmt.Errorf("get node info api err: %w", err)
	}

	// Cache the ChainId for later calls, because performance
	rc.chainId = info.ChainId
	return info, err
}

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
func (rc *NodeClient) TransactionByHash(txnHash string) (data *api.Transaction, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_hash", txnHash)
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) TransactionByVersion(version uint64) (data *api.Transaction, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_version", strconv.FormatUint(version, 10))
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) getTransactionCommon(restUrl *url.URL) (data *api.Transaction, err error) {
	// Fetch transaction
	data, err = Get[*api.Transaction](rc, restUrl.String())
	if err != nil {
		return data, fmt.Errorf("get transaction api err: %w", err)
	}
	return data, nil
}

func (rc *NodeClient) BlockByVersion(ledgerVersion uint64, withTransactions bool) (data *api.Block, err error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_version", strconv.FormatUint(ledgerVersion, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

func (rc *NodeClient) BlockByHeight(blockHeight uint64, withTransactions bool) (data *api.Block, err error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_height", strconv.FormatUint(blockHeight, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

func (rc *NodeClient) getBlockCommon(restUrl *url.URL, withTransactions bool) (block *api.Block, err error) {
	params := url.Values{}
	params.Set("with_transactions", strconv.FormatBool(withTransactions))
	restUrl.RawQuery = params.Encode()

	// Fetch block
	block, err = Get[*api.Block](rc, restUrl.String())
	if err != nil {
		return block, fmt.Errorf("get block api err: %w", err)
	}

	// Now, let's fill in any missing transactions in the block
	numTransactions := block.LastVersion - block.FirstVersion + 1
	retrievedTransactions := uint64(len(block.Transactions))

	// Transaction is always not pending, so it will never be nil
	cursor := block.Transactions[len(block.Transactions)-1].Version()

	// TODO: I maybe should pull these concurrently, but not for now
	for retrievedTransactions < numTransactions {
		numToPull := numTransactions - retrievedTransactions
		transactions, innerError := rc.Transactions(cursor, &numToPull)
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
// Accept option arguments PollPeriod and PollTimeout like PollForTransactions.
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
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
func (rc *NodeClient) Transactions(start *uint64, limit *uint64) (data []*api.Transaction, err error) {
	// Can only pull everything in parallel if a start and a limit is handled
	if start != nil && limit != nil {
		return rc.transactionsConcurrent(*start, *limit)
	} else {
		// TODO: need to pull the first page, then the rest after that / provide similar behavior
		return rc.transactionsInner(start, limit)
	}
}

func (rc *NodeClient) transactionsConcurrent(start uint64, limit uint64) (data []*api.Transaction, err error) {
	const transactionsPageSize = 100
	// If we know both, we can fetch all concurrently
	type Pair struct {
		start uint64
		end   uint64
	}

	// If the limit is  greater than the page size, we need to fetch concurrently, otherwise not
	if limit > transactionsPageSize {
		numChannels := limit / transactionsPageSize
		if limit%transactionsPageSize > 0 {
			numChannels++
		}
		channels := make([]chan ConcResponse[[]*api.Transaction], numChannels)
		for i := uint64(0); i*transactionsPageSize < limit; i += 1 {
			channels[i] = make(chan ConcResponse[[]*api.Transaction], 1)
			st := start + i*100
			li := min(transactionsPageSize, limit-i*transactionsPageSize)
			go fetch(func() ([]*api.Transaction, error) {
				return rc.transactionsConcurrent(st, li)
			}, channels[i])
		}

		responses := make([]*api.Transaction, limit)
		cursor := 0
		for i, ch := range channels {
			response := <-ch
			if response.Err != nil {
				return nil, err
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
		return responses, nil
	} else {
		response, err := rc.transactionsInner(&start, &limit)
		if err != nil {
			return nil, err
		} else {
			return response, nil
		}
	}
}

func (rc *NodeClient) transactionsInner(start *uint64, limit *uint64) (data []*api.Transaction, err error) {
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
	data, err = Get[[]*api.Transaction](rc, au.String())
	if err != nil {
		return data, fmt.Errorf("get transactions api err: %w", err)
	}
	return data, nil
}

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

type EstimateGasUnitPrice bool

type EstimateMaxGasAmount bool

type EstimatePrioritizedGasUnitPrice bool

func (rc *NodeClient) SimulateTransaction(rawTxn *RawTransaction, sender TransactionSigner, options ...any) (data []*api.UserTransaction, err error) {
	// build authenticator for simulation
	var auth *crypto.AccountAuthenticator
	derivationScheme := sender.PubKey().Scheme()
	switch derivationScheme {
	case crypto.Ed25519Scheme:
		auth = &crypto.AccountAuthenticator{
			Variant: crypto.AccountAuthenticatorEd25519,
			Auth: &crypto.Ed25519Authenticator{
				PubKey: sender.PubKey().(*crypto.Ed25519PublicKey),
				Sig:    &crypto.Ed25519Signature{Inner: [64]byte(make([]byte, 64))},
			},
		}
	case crypto.SingleKeyScheme:
		mockSig := &crypto.AnySignature{}
		_ = mockSig.FromBytes(make([]byte, 64))
		auth = &crypto.AccountAuthenticator{
			Variant: crypto.AccountAuthenticatorSingleSender,
			Auth: &crypto.SingleKeyAuthenticator{
				PubKey: sender.PubKey().(*crypto.AnyPublicKey),
				Sig:    mockSig,
			},
		}
	case crypto.MultiEd25519Scheme:
	case crypto.MultiKeyScheme:
		// todo: add support for multikey simulation
		return nil, fmt.Errorf("currently unsupported sender derivation scheme %v", derivationScheme)
	default:
		return nil, fmt.Errorf("unexpected sender derivation scheme %v", derivationScheme)
	}

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

func (rc *NodeClient) GetChainId() (chainId uint8, err error) {
	if rc.chainId == 0 {
		// Calling Info will cache the chain Id
		info, err := rc.Info()
		if err != nil {
			return 0, err
		}
		return info.ChainId, nil
	}
	return rc.chainId, nil
}

type MaxGasAmount uint64

type GasUnitPrice uint64

type ExpirationSeconds int64

type FeePayer *AccountAddress

type AdditionalSigners []AccountAddress

type SequenceNumber uint64

type ChainIdOption uint8

// BuildTransaction builds a raw transaction for signing
// Accepts options: MaxGasAmount, GasUnitPrice, ExpirationSeconds, SequenceNumber, ChainIdOption, FeePayer, AdditionalSigners
func (rc *NodeClient) BuildTransaction(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error) {

	maxGasAmount := DefaultMaxGasAmount
	gasUnitPrice := uint64(0) //DefaultGasUnitPrice
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
// Accepts options: MaxGasAmount, GasUnitPrice, ExpirationSeconds, SequenceNumber, ChainIdOption, FeePayer, AdditionalSigners
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

type ViewPayload struct {
	Module   ModuleId
	Function string
	ArgTypes []TypeTag
	Args     [][]byte
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

// TODO: add caching for some period of time
func (rc *NodeClient) EstimateGasPrice() (info EstimateGasInfo, err error) {
	au := rc.baseUrl.JoinPath("estimate_gas_price")
	info, err = Get[EstimateGasInfo](rc, au.String())
	if err != nil {
		return info, fmt.Errorf("estimate gas price err: %w", err)
	}
	return info, nil
}

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

func (rc *NodeClient) NodeHealthCheck(durationSecs ...uint64) (api.HealthCheckResponse, error) {
	au := rc.baseUrl.JoinPath("-/healthy")
	if len(durationSecs) > 0 {
		params := url.Values{}
		params.Set("duration_secs", strconv.FormatUint(durationSecs[0], 10))
		au.RawQuery = params.Encode()
	}
	return Get[api.HealthCheckResponse](rc, au.String())
}

func Get[T any](rc *NodeClient, getUrl string) (out T, err error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set(ClientHeader, ClientHeaderValue)
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

func (rc *NodeClient) GetBCS(getUrl string) (out []byte, err error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-bcs")
	req.Header.Set(ClientHeader, ClientHeaderValue)
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

// TODO: Support multiple output channels?
func fetch[T any](inner func() (T, error), result chan ConcResponse[T]) {
	response, err := inner()
	if err != nil {
		result <- ConcResponse[T]{Err: err}
	} else {
		result <- ConcResponse[T]{Result: response}
	}
}
