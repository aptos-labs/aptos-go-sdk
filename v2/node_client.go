package aptos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
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
	path := fmt.Sprintf("accounts/%s", address.String())
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
	resource, err := c.AccountResource(ctx, address, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", opts...)
	if err != nil {
		return 0, err
	}

	coin, ok := resource.Data["coin"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("unexpected coin data format")
	}

	valueStr, ok := coin["value"].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value format")
	}

	return strconv.ParseUint(valueStr, 10, 64)
}

// BuildTransaction builds an unsigned transaction.
func (c *nodeClient) BuildTransaction(ctx context.Context, sender AccountAddress, payload Payload, opts ...TransactionOption) (*RawTransaction, error) {
	config := &TransactionConfig{
		MaxGasAmount:       200000,
		ExpirationDuration: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Get sequence number if not provided
	var seqNum uint64
	if config.SequenceNumber != nil {
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

// SimulateTransaction simulates a transaction without submitting it.
func (c *nodeClient) SimulateTransaction(ctx context.Context, txn *RawTransaction, signer Signer, opts ...TransactionOption) (*SimulationResult, error) {
	// Create simulation authenticator (zero signature)
	auth := SimulationAuthenticator(signer)

	// Create signed transaction for simulation
	signed := &SignedTransaction{
		Transaction:   txn,
		Authenticator: auth,
	}

	// Serialize to BCS
	txnBytes, err := bcs.Serialize(signed)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("estimate_gas_unit_price", "true")
	params.Set("estimate_max_gas_amount", "true")
	params.Set("estimate_prioritized_gas_unit_price", "false")

	path := "transactions/simulate?" + params.Encode()

	var results []*SimulationResult
	if err := c.postBCS(ctx, path, txnBytes, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("simulation returned no results")
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
	path := fmt.Sprintf("transactions/by_hash/%s", hash)
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

	// Build request body
	body := map[string]any{
		"function":       fmt.Sprintf("%s::%s", payload.Module.String(), payload.Function),
		"type_arguments": []string{}, // TODO: Convert TypeTags to strings
		"arguments":      payload.Args,
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

// AccountTransactions returns transactions sent by an account.
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

// Fund requests tokens from the faucet.
func (c *nodeClient) Fund(ctx context.Context, address AccountAddress, amount uint64) error {
	if c.config.network.FaucetURL == "" {
		return fmt.Errorf("faucet not available for this network")
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode:    resp.StatusCode,
			Message:       string(body),
			RequestMethod: req.Method,
			RequestURL:    req.URL.String(),
		}
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
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
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

		return apiErr
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// isAPIError checks if err is an APIError and assigns it to target if so.
func isAPIError(err error, target **APIError) bool {
	apiErr, ok := err.(*APIError)
	if ok {
		*target = apiErr
	}
	return ok
}

// Ensure nodeClient implements Client
var _ Client = (*nodeClient)(nil)
