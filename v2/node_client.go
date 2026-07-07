package aptos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// nodeClient is the concrete implementation of the Client interface.
type nodeClient struct {
	config     *ClientConfig
	httpClient HTTPDoer
	baseURL    *url.URL
	logger     *slog.Logger
	chainID    uint8
	chainIDSet bool
}

// newNodeClient creates a new nodeClient from the configuration.
func newNodeClient(config *ClientConfig) (*nodeClient, error) {
	baseURL, err := url.Parse(config.network.NodeURL)
	if err != nil {
		return nil, fmt.Errorf("invalid node URL: %w", err)
	}

	httpClient := config.httpClient
	if httpClient == nil {
		httpClient = &defaultHTTPClient{
			client:  &http.Client{Timeout: config.timeout},
			headers: config.headers,
		}
	}

	logger := config.logger
	if logger == nil {
		logger = slog.Default()
	}

	// Apply retry / rate-limit handling middleware when configured. This is a
	// no-op wrapper (returns httpClient unchanged) when neither WithRetry,
	// WithRetryConfig, nor WithRateLimitHandling were supplied.
	httpClient = newRetryHTTPClient(httpClient, config.retryConfig, config.rateLimitConfig, logger)

	return &nodeClient{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
		logger:     logger,
		chainID:    config.network.ChainID,
		chainIDSet: config.network.ChainID != 0,
	}, nil
}

// defaultHTTPClient wraps http.Client to implement HTTPDoer.
type defaultHTTPClient struct {
	client  *http.Client
	headers map[string]string
}

func (c *defaultHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return c.client.Do(req)
}

// Info returns information about the connected node.
func (c *nodeClient) Info(ctx context.Context) (*NodeInfo, error) {
	var info NodeInfo
	if err := c.get(ctx, "", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ChainID returns the chain ID of the connected network.
func (c *nodeClient) ChainID(ctx context.Context) (uint8, error) {
	if c.chainIDSet {
		return c.chainID, nil
	}

	info, err := c.Info(ctx)
	if err != nil {
		return 0, err
	}

	c.chainID = info.ChainID
	c.chainIDSet = true
	return c.chainID, nil
}

// Account returns information about an account.
func (c *nodeClient) Account(ctx context.Context, address AccountAddress) (*AccountInfo, error) {
	var info AccountInfo
	path := "accounts/" + address.String()
	if err := c.get(ctx, path, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// AccountResources returns all resources for an account.
func (c *nodeClient) AccountResources(ctx context.Context, address AccountAddress, opts ...ResourceOption) ([]Resource, error) {
	config := &ResourceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	path := fmt.Sprintf("accounts/%s/resources", address.String())
	if config.LedgerVersion != nil {
		path += fmt.Sprintf("?ledger_version=%d", *config.LedgerVersion)
	}

	var resources []Resource
	if err := c.get(ctx, path, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// AccountResource returns a specific resource for an account.
func (c *nodeClient) AccountResource(ctx context.Context, address AccountAddress, resourceType string, opts ...ResourceOption) (*Resource, error) {
	config := &ResourceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	path := fmt.Sprintf("accounts/%s/resource/%s", address.String(), url.PathEscape(resourceType))
	if config.LedgerVersion != nil {
		path += fmt.Sprintf("?ledger_version=%d", *config.LedgerVersion)
	}

	var resource Resource
	if err := c.get(ctx, path, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// AccountBalance returns the APT balance for an account.
func (c *nodeClient) AccountBalance(ctx context.Context, address AccountAddress, opts ...ResourceOption) (uint64, error) {
	// Use the 0x1::coin::balance<AptosCoin>(address) view function.
	//
	// Why a view function rather than reading 0x1::coin::CoinStore<AptosCoin>:
	//   * Newer Aptos networks have migrated APT to the fungible-asset model.
	//     For accounts created post-migration there is no CoinStore resource at
	//     all, only a primary fungible store, and a direct resource read returns
	//     a 404. The view function transparently handles both representations.
	//   * The view function returns the balance as a stringified u64 in the
	//     first slot, which is stable across both representations.
	resourceConfig := &ResourceConfig{}
	for _, opt := range opts {
		opt(resourceConfig)
	}

	viewOpts := []ViewOption(nil)
	if resourceConfig.LedgerVersion != nil {
		viewOpts = append(viewOpts, AtLedgerVersion(*resourceConfig.LedgerVersion))
	}

	values, err := c.View(ctx, &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "coin"},
		Function: "balance",
		TypeArgs: []TypeTag{AptosCoinTypeTag},
		Args:     []any{address.String()},
	}, viewOpts...)
	if err != nil {
		return 0, err
	}
	if len(values) == 0 {
		return 0, errors.New("view function returned no values")
	}

	return parseU64Balance(values[0])
}

// parseU64Balance converts a JSON-decoded balance value (a stringified u64 or,
// defensively, a JSON number) into a uint64.
func parseU64Balance(value any) (uint64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseUint(v, 10, 64)
	case float64:
		// JSON numbers decode to float64. The node *should* return
		// balances as stringified u64 (and does today), but we accept a
		// numeric response defensively. float64 can exactly represent
		// integers only up to 2^53 — about 9.0×10^16 octas, i.e. ~90M
		// APT — so a value beyond that, or one that isn't a
		// non-negative integer, indicates an unexpected response shape
		// and we refuse to silently truncate it.
		if v < 0 || v != float64(uint64(v)) {
			return 0, fmt.Errorf("balance %v is not a non-negative integer", v)
		}
		const maxExactFloat64Int = float64(1 << 53)
		if v > maxExactFloat64Int {
			return 0, fmt.Errorf(
				"balance %v exceeds the float64 exact-integer range (2^53); "+
					"node should return stringified u64",
				v,
			)
		}
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("unexpected balance value type %T", value)
	}
}

// BuildTransaction builds an unsigned transaction.
func (c *nodeClient) BuildTransaction(ctx context.Context, sender AccountAddress, payload Payload, opts ...TransactionOption) (*RawTransaction, error) {
	config := &TransactionConfig{
		MaxGasAmount:       2_000_000,
		ExpirationDuration: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Orderless transactions are validated by a one-time nonce rather than the
	// account's sequence number, so we wrap the payload and use u64::MAX as the
	// sequence number without fetching the account.
	var seqNum uint64
	if config.ReplayProtectionNonce != nil {
		payload = wrapOrderless(payload, config.ReplayProtectionNonce)
		seqNum = math.MaxUint64
	} else if config.SequenceNumber != nil {
		seqNum = *config.SequenceNumber
	} else {
		info, err := c.Account(ctx, sender)
		if err != nil {
			return nil, fmt.Errorf("failed to get account info: %w", err)
		}
		seqNum = info.SequenceNumber
	}

	// Get chain ID
	chainID, err := c.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Get gas price if not provided
	gasPrice := config.GasUnitPrice
	if gasPrice == 0 && config.EstimateGas {
		estimate, err := c.EstimateGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas price: %w", err)
		}
		if config.EstimatePrioritizedGas {
			gasPrice = estimate.PrioritizedGasEstimate
		} else {
			gasPrice = estimate.GasEstimate
		}
	}
	if gasPrice == 0 {
		gasPrice = 100 // Default gas price
	}

	expiration := uint64(time.Now().Add(config.ExpirationDuration).Unix())

	return &RawTransaction{
		Sender:                     sender,
		SequenceNumber:             seqNum,
		Payload:                    payload,
		MaxGasAmount:               config.MaxGasAmount,
		GasUnitPrice:               gasPrice,
		ExpirationTimestampSeconds: expiration,
		ChainID:                    chainID,
	}, nil
}

// SimulateTransaction simulates a single-sender transaction without
// submitting it. To simulate a multi-agent or fee-payer transaction use
// SimulateMultiAgentTransaction or SimulateFeePayerTransaction; this
// entry point intentionally only wires up a SingleSender authenticator.
//
// Gas-estimation query parameters are derived from opts: WithGasEstimation
// turns on both unit-price and max-amount estimation; WithPrioritizedGas
// additionally requests the prioritized estimate. Pass no opts to get the
// default (estimate gas price and max amount, no prioritized estimate),
// which matches v1 behavior.
func (c *nodeClient) SimulateTransaction(ctx context.Context, txn *RawTransaction, signer Signer, opts ...TransactionOption) (*SimulationResult, error) {
	auth := SimulationAuthenticator(signer)
	if auth == nil {
		return nil, errors.New("signer has no public key")
	}

	// Wrap in SingleSenderAuthenticator so the on-the-wire TransactionAuthenticator
	// variant is `SingleSender` (4). Assigning the raw *AccountAuthenticator here
	// would serialize as just the AccountAuthenticator variant byte; for Ed25519
	// (variant 0) that accidentally matches the legacy `TxnAuth::Ed25519` layout,
	// but for Secp256k1/Secp256r1/Keyless it would mis-decode at the node.
	// SignTransaction wraps identically — keeping these paths consistent matters.
	signed := &SignedTransaction{
		Transaction:   txn,
		Authenticator: &SingleSenderAuthenticator{Sender: auth},
	}
	return c.simulateSigned(ctx, signed, opts...)
}

// SimulateMultiAgentTransaction simulates a multi-agent transaction.
// The primary signer signs as the sender; secondarySigners provide the
// remaining authenticators in the order the on-chain multi-agent
// signature expects.
func (c *nodeClient) SimulateMultiAgentTransaction(ctx context.Context, txn *RawTransaction, sender Signer, secondarySigners []Signer, secondaryAddrs []AccountAddress, opts ...TransactionOption) (*SimulationResult, error) {
	if len(secondarySigners) != len(secondaryAddrs) {
		return nil, fmt.Errorf("secondarySigners (%d) and secondaryAddrs (%d) length mismatch", len(secondarySigners), len(secondaryAddrs))
	}

	senderAuth := SimulationAuthenticator(sender)
	if senderAuth == nil {
		return nil, errors.New("sender has no public key")
	}

	secondaryAuths := make([]*AccountAuthenticator, len(secondarySigners))
	for i, s := range secondarySigners {
		a := SimulationAuthenticator(s)
		if a == nil {
			return nil, fmt.Errorf("secondary signer %d has no public key", i)
		}
		secondaryAuths[i] = a
	}

	signed := &SignedTransaction{
		Transaction: txn,
		Authenticator: &MultiAgentAuthenticator{
			Sender:                   senderAuth,
			SecondarySignerAddresses: secondaryAddrs,
			SecondarySigners:         secondaryAuths,
		},
	}
	return c.simulateSigned(ctx, signed, opts...)
}

// SimulateFeePayerTransaction simulates a sponsored (fee-payer)
// transaction. feePayerAddr / feePayer specify who pays gas; pass nil for
// secondarySigners / secondaryAddrs for the common single-sender + sponsor
// case.
func (c *nodeClient) SimulateFeePayerTransaction(ctx context.Context, txn *RawTransaction, sender Signer, secondarySigners []Signer, secondaryAddrs []AccountAddress, feePayerAddr AccountAddress, feePayer Signer, opts ...TransactionOption) (*SimulationResult, error) {
	if len(secondarySigners) != len(secondaryAddrs) {
		return nil, fmt.Errorf("secondarySigners (%d) and secondaryAddrs (%d) length mismatch", len(secondarySigners), len(secondaryAddrs))
	}

	senderAuth := SimulationAuthenticator(sender)
	if senderAuth == nil {
		return nil, errors.New("sender has no public key")
	}

	secondaryAuths := make([]*AccountAuthenticator, len(secondarySigners))
	for i, s := range secondarySigners {
		a := SimulationAuthenticator(s)
		if a == nil {
			return nil, fmt.Errorf("secondary signer %d has no public key", i)
		}
		secondaryAuths[i] = a
	}

	feePayerAuth := SimulationAuthenticator(feePayer)
	if feePayerAuth == nil {
		return nil, errors.New("fee payer has no public key")
	}

	signed := &SignedTransaction{
		Transaction: txn,
		Authenticator: &FeePayerAuthenticator{
			Sender:                   senderAuth,
			SecondarySignerAddresses: secondaryAddrs,
			SecondarySigners:         secondaryAuths,
			FeePayerAddress:          feePayerAddr,
			FeePayerAuth:             feePayerAuth,
		},
	}
	return c.simulateSigned(ctx, signed, opts...)
}

// simulateSigned submits a pre-built simulation SignedTransaction.
func (c *nodeClient) simulateSigned(ctx context.Context, signed *SignedTransaction, opts ...TransactionOption) (*SimulationResult, error) {
	// Default: estimate gas price and max amount, no prioritized estimate.
	// Mirrors v1's behavior so callers migrating don't see a gas-estimation
	// regression. Caller can override via WithGasEstimation / WithPrioritizedGas.
	cfg := &TransactionConfig{
		EstimateGas:            true,
		EstimatePrioritizedGas: false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	txnBytes, err := bcs.Serialize(signed)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	params := url.Values{}
	params.Set("estimate_gas_unit_price", strconv.FormatBool(cfg.EstimateGas))
	params.Set("estimate_max_gas_amount", strconv.FormatBool(cfg.EstimateGas))
	params.Set("estimate_prioritized_gas_unit_price", strconv.FormatBool(cfg.EstimatePrioritizedGas))

	path := "transactions/simulate?" + params.Encode()

	var results []*SimulationResult
	if err := c.postBCS(ctx, path, txnBytes, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("simulation returned no results")
	}

	return results[0], nil
}

// SubmitTransaction submits a signed transaction to the network.
func (c *nodeClient) SubmitTransaction(ctx context.Context, signed *SignedTransaction) (*SubmitResult, error) {
	// Serialize to BCS
	txnBytes, err := bcs.Serialize(signed)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	var result SubmitResult
	if err := c.postBCS(ctx, "transactions", txnBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SignAndSubmitTransaction signs and submits a transaction.
// The signer must implement TransactionSigner to provide both signing and address.
func (c *nodeClient) SignAndSubmitTransaction(ctx context.Context, signer TransactionSigner, payload Payload, opts ...TransactionOption) (*SubmitResult, error) {
	// Build the transaction
	txn, err := c.BuildTransaction(ctx, signer.Address(), payload, opts...)
	if err != nil {
		return nil, err
	}

	// Sign the transaction using the helper function
	signed, err := SignTransaction(signer, txn)
	if err != nil {
		return nil, err
	}

	// Submit
	return c.SubmitTransaction(ctx, signed)
}

// WaitForTransaction waits for a transaction to be confirmed.
func (c *nodeClient) WaitForTransaction(ctx context.Context, hash string, opts ...PollOption) (*Transaction, error) {
	config := &PollConfig{
		PollInterval: 1 * time.Second,
		Timeout:      30 * time.Second,
	}
	for _, opt := range opts {
		opt(config)
	}

	deadline := time.Now().Add(config.Timeout)
	for time.Now().Before(deadline) {
		txn, err := c.Transaction(ctx, hash)
		if err != nil {
			var apiErr *APIError
			if ok := isAPIError(err, &apiErr); ok && apiErr.StatusCode == http.StatusNotFound {
				// Transaction not found yet, keep polling
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(config.PollInterval):
					continue
				}
			}
			return nil, err
		}

		// Check if transaction is committed
		if txn.Type == "pending_transaction" {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(config.PollInterval):
				continue
			}
		}

		return txn, nil
	}

	return nil, fmt.Errorf("%w: transaction %s", ErrTimeout, hash)
}

// Transaction returns a transaction by hash.
func (c *nodeClient) Transaction(ctx context.Context, hash string) (*Transaction, error) {
	var txn Transaction
	path := "transactions/by_hash/" + hash
	if err := c.get(ctx, path, &txn); err != nil {
		return nil, err
	}
	return &txn, nil
}

// TransactionByVersion returns a transaction by version.
func (c *nodeClient) TransactionByVersion(ctx context.Context, version uint64) (*Transaction, error) {
	var txn Transaction
	path := fmt.Sprintf("transactions/by_version/%d", version)
	if err := c.get(ctx, path, &txn); err != nil {
		return nil, err
	}
	return &txn, nil
}

// Transactions returns recent transactions.
func (c *nodeClient) Transactions(ctx context.Context, start *uint64, limit *uint64) ([]*Transaction, error) {
	path := "transactions"
	params := url.Values{}
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var txns []*Transaction
	if err := c.get(ctx, path, &txns); err != nil {
		return nil, err
	}
	return txns, nil
}

// TransactionsIter returns an iterator over transactions.
func (c *nodeClient) TransactionsIter(ctx context.Context, start *uint64) iter.Seq2[*Transaction, error] {
	return func(yield func(*Transaction, error) bool) {
		var currentStart *uint64
		if start != nil {
			v := *start
			currentStart = &v
		}

		limit := uint64(100)
		for {
			txns, err := c.Transactions(ctx, currentStart, &limit)
			if err != nil {
				yield(nil, err)
				return
			}

			if len(txns) == 0 {
				return
			}

			for _, txn := range txns {
				if !yield(txn, nil) {
					return
				}
			}

			// Update start for next batch
			lastVersion := txns[len(txns)-1].Version + 1
			currentStart = &lastVersion
		}
	}
}

func normalizeViewArgs(args []any) []any {
	if args == nil {
		return nil
	}
	normalized := make([]any, len(args))
	for i, arg := range args {
		normalized[i] = normalizeViewArg(arg)
	}
	return normalized
}

func normalizeViewArg(arg any) any {
	if arg == nil {
		return nil
	}

	// *big.Int covers Move's u128/u256/i128/i256 view arguments. The
	// node expects these as decimal strings — passing the *big.Int
	// through would either be JSON-encoded as an unquoted number (too
	// large to fit JSON's safe integer range) or rejected outright.
	if bi, ok := arg.(*big.Int); ok {
		if bi == nil {
			return nil
		}
		return bi.String()
	}

	value := reflect.ValueOf(arg)
	switch value.Kind() {
	case reflect.Uint, reflect.Uint64:
		// Go's platform-sized uint is at least 32 bits; on 64-bit
		// builds it can exceed JSON's safe integer range. Stringify
		// for safety — the node accepts decimal strings for all
		// unsigned widths.
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return arg
		}
		normalized := make([]any, value.Len())
		for i := range normalized {
			normalized[i] = normalizeViewArg(value.Index(i).Interface())
		}
		return normalized
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return arg
		}
		normalized := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			normalized[iter.Key().String()] = normalizeViewArg(iter.Value().Interface())
		}
		return normalized
	default:
		return arg
	}
}

// View executes a view function and returns the results.
func (c *nodeClient) View(ctx context.Context, payload *ViewPayload, opts ...ViewOption) ([]any, error) {
	config := &ViewConfig{}
	for _, opt := range opts {
		opt(config)
	}

	path := "view"
	if config.LedgerVersion != nil {
		path += fmt.Sprintf("?ledger_version=%d", *config.LedgerVersion)
	}

	typeArgs := make([]string, 0, len(payload.TypeArgs))
	for i := range payload.TypeArgs {
		typeArgs = append(typeArgs, payload.TypeArgs[i].String())
	}

	body := map[string]any{
		"function":       fmt.Sprintf("%s::%s", payload.Module.String(), payload.Function),
		"type_arguments": typeArgs,
		"arguments":      normalizeViewArgs(payload.Args),
	}

	var result []any
	if err := c.post(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// BlockByHeight returns a block by height.
func (c *nodeClient) BlockByHeight(ctx context.Context, height uint64, withTransactions bool) (*Block, error) {
	path := fmt.Sprintf("blocks/by_height/%d", height)
	if withTransactions {
		path += "?with_transactions=true"
	}

	var block Block
	if err := c.get(ctx, path, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// BlockByVersion returns the block containing a specific version.
func (c *nodeClient) BlockByVersion(ctx context.Context, version uint64, withTransactions bool) (*Block, error) {
	path := fmt.Sprintf("blocks/by_version/%d", version)
	if withTransactions {
		path += "?with_transactions=true"
	}

	var block Block
	if err := c.get(ctx, path, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// EstimateGasPrice returns gas price estimates.
func (c *nodeClient) EstimateGasPrice(ctx context.Context) (*GasEstimate, error) {
	var estimate GasEstimate
	if err := c.get(ctx, "estimate_gas_price", &estimate); err != nil {
		return nil, err
	}
	return &estimate, nil
}

// EventsByHandle returns events for an event handle.
func (c *nodeClient) EventsByHandle(ctx context.Context, address AccountAddress, handle string, field string, start *uint64, limit *uint64) ([]Event, error) {
	path := fmt.Sprintf("accounts/%s/events/%s/%s", address.String(), handle, field)
	params := url.Values{}
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var events []Event
	if err := c.get(ctx, path, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// HealthCheck checks the health of the node.
func (c *nodeClient) HealthCheck(ctx context.Context, durationSecs ...uint64) (*HealthCheckResponse, error) {
	path := "-/healthy"
	if len(durationSecs) > 0 {
		path += fmt.Sprintf("?duration_secs=%d", durationSecs[0])
	}

	var resp HealthCheckResponse
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AccountModule returns the bytecode and ABI for a module.
func (c *nodeClient) AccountModule(ctx context.Context, address AccountAddress, moduleName string, opts ...ResourceOption) (*ModuleBytecode, error) {
	config := &ResourceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	path := fmt.Sprintf("accounts/%s/module/%s", address.String(), moduleName)
	if config.LedgerVersion != nil {
		path += fmt.Sprintf("?ledger_version=%d", *config.LedgerVersion)
	}

	var module ModuleBytecode
	if err := c.get(ctx, path, &module); err != nil {
		return nil, err
	}
	return &module, nil
}

// AccountTransactions returns transactions sent by an account in a single REST response.
// Unlike [github.com/aptos-labs/aptos-go-sdk.NodeClient.AccountTransactions], this does not
// merge multiple pages when [limit] is larger than one response page; pass an explicit [start]
// and/or issue additional calls if you need more than the node returns in one request.
func (c *nodeClient) AccountTransactions(ctx context.Context, address AccountAddress, start *uint64, limit *uint64) ([]*Transaction, error) {
	path := fmt.Sprintf("accounts/%s/transactions", address.String())
	params := url.Values{}
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var txns []*Transaction
	if err := c.get(ctx, path, &txns); err != nil {
		return nil, err
	}
	return txns, nil
}

// BatchSubmitTransaction submits multiple signed transactions in a single request.
func (c *nodeClient) BatchSubmitTransaction(ctx context.Context, signed []*SignedTransaction) (*BatchSubmitResult, error) {
	// Serialize all transactions
	ser := bcs.NewSerializer()
	ser.Uleb128(uint32(len(signed)))
	for _, txn := range signed {
		txn.MarshalBCS(ser)
		if ser.Error() != nil {
			return nil, fmt.Errorf("failed to serialize transaction: %w", ser.Error())
		}
	}

	var result BatchSubmitResult
	if err := c.postBCS(ctx, "transactions/batch", ser.ToBytes(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// EventsByCreationNumber returns events by creation number.
func (c *nodeClient) EventsByCreationNumber(ctx context.Context, address AccountAddress, creationNumber uint64, start *uint64, limit *uint64) ([]Event, error) {
	path := fmt.Sprintf("accounts/%s/events/%d", address.String(), creationNumber)
	params := url.Values{}
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var events []Event
	if err := c.get(ctx, path, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// Fund requests tokens from the faucet and waits for the resulting funding
// transactions to commit before returning.
//
// The localnet (and aptos-faucet-service in general) responds with a JSON
// array of transaction hashes. Returning before those transactions commit
// makes the call effectively asynchronous: a follow-up balance/account read
// will race the indexer/state and frequently observe a zero balance, which
// silently breaks integration tests and any caller that reasonably expects
// "Fund returned successfully" to mean "the funds are visible". So we wait.
//
// If the faucet response shape is different (for example, some hosted
// faucets return an empty body or an object with no hashes), we fall back to
// best-effort behavior: return nil rather than fail, since the caller can
// always re-poll for the balance.
func (c *nodeClient) Fund(ctx context.Context, address AccountAddress, amount uint64) error {
	if c.config.network.FaucetURL == "" {
		return errors.New("faucet not available for this network")
	}

	faucetURL, err := url.Parse(c.config.network.FaucetURL)
	if err != nil {
		return fmt.Errorf("invalid faucet URL: %w", err)
	}

	faucetURL.Path = "/mint"
	q := faucetURL.Query()
	q.Set("address", address.String())
	q.Set("amount", strconv.FormatUint(amount, 10))
	faucetURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, faucetURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		// We need the body for two things: surfacing the error message
		// on a non-2xx response, and parsing the JSON hashes we have to
		// wait on. Returning here is safer than silently treating Fund
		// as "fire and forget" (which would re-introduce the race the
		// hash-wait was added to fix). Wrap with status context so the
		// caller can tell read-after-success from read-after-error.
		return fmt.Errorf("read faucet response (status %d): %w", resp.StatusCode, readErr)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return &APIError{
			StatusCode:    resp.StatusCode,
			Message:       string(body),
			RequestMethod: req.Method,
			RequestURL:    req.URL.String(),
		}
	}

	// Parse out any transaction hashes the faucet emitted and wait for them
	// to commit. We accept either ["hash", ...] (the localnet/standalone
	// faucet) or {"txn_hashes": ["hash", ...]} (some hosted faucets); on
	// anything else we silently skip the wait.
	hashes := parseFaucetHashes(body)
	for _, h := range hashes {
		if _, err := c.WaitForTransaction(ctx, h); err != nil {
			return fmt.Errorf("waiting for faucet transaction %s: %w", h, err)
		}
	}
	return nil
}

// parseFaucetHashes extracts transaction hashes from a faucet response body.
// Returns nil for any unexpected shape rather than erroring, so unknown
// faucet variants degrade to "fire and forget" rather than to a hard failure.
func parseFaucetHashes(body []byte) []string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil
	}

	var asArray []string
	if err := json.Unmarshal(body, &asArray); err == nil {
		return asArray
	}

	var asObject struct {
		TxnHashes []string `json:"txn_hashes"`
	}
	if err := json.Unmarshal(body, &asObject); err == nil {
		return asObject.TxnHashes
	}

	return nil
}

// HTTP helpers

func (c *nodeClient) get(ctx context.Context, path string, result any) error {
	reqURL := c.buildURL(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	return c.doRequest(ctx, req, result)
}

// buildURL constructs a full URL from a path, handling query strings properly.
func (c *nodeClient) buildURL(path string) string {
	// Check if path contains query string
	if idx := strings.Index(path, "?"); idx != -1 {
		basePath := path[:idx]
		query := path[idx:]
		return c.baseURL.JoinPath(basePath).String() + query
	}
	return c.baseURL.JoinPath(path).String()
}

func (c *nodeClient) post(ctx context.Context, path string, body any, result any) error {
	reqURL := c.buildURL(path)

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.doRequest(ctx, req, result)
}

// postBCS sends a POST request with BCS-encoded body.
func (c *nodeClient) postBCS(ctx context.Context, path string, body []byte, result any) error {
	reqURL := c.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x.aptos.signed_transaction+bcs")
	req.Header.Set("Accept", "application/json")

	return c.doRequest(ctx, req, result)
}

func (c *nodeClient) doRequest(ctx context.Context, req *http.Request, result any) error {
	_, err := c.doRequestReturningHeaders(ctx, req, result)
	return err
}

// doRequestReturningHeaders is like doRequest but also returns the response
// headers on success. It is used by paginated endpoints that carry their next
// cursor in a response header (X-Aptos-Cursor).
func (c *nodeClient) doRequestReturningHeaders(ctx context.Context, req *http.Request, result any) (http.Header, error) {
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode:    resp.StatusCode,
			RequestMethod: req.Method,
			RequestURL:    req.URL.String(),
		}

		// Try to parse error message from body
		var errResp struct {
			Message   string    `json:"message"`
			ErrorCode string    `json:"error_code"`
			VMStatus  *VMStatus `json:"vm_status"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			apiErr.Message = errResp.Message
			apiErr.ErrorCode = errResp.ErrorCode
			apiErr.VMStatus = errResp.VMStatus
		} else {
			apiErr.Message = string(body)
		}

		return nil, apiErr
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return resp.Header, nil
}

// isAPIError checks if err is an APIError and assigns it to target if so.
func isAPIError(err error, target **APIError) bool {
	return errors.As(err, target)
}

// Ensure nodeClient implements Client
var _ Client = (*nodeClient)(nil)
