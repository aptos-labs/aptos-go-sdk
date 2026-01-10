package testutil

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

// FakeClient is a mock implementation of the aptos.Client interface for testing.
// It allows you to configure responses and verify interactions.
type FakeClient struct {
	mu sync.RWMutex

	// Configured data
	accounts     map[aptos.AccountAddress]*aptos.AccountInfo
	resources    map[string][]aptos.Resource
	balances     map[aptos.AccountAddress]uint64
	nodeInfo     *aptos.NodeInfo
	gasEstimate  *aptos.GasEstimate
	transactions map[string]*aptos.Transaction
	blocks       map[uint64]*aptos.Block

	// Error simulation
	errors map[string]error

	// Request recording
	recording bool
	calls     []RecordedCall
}

// RecordedCall represents a recorded method call.
type RecordedCall struct {
	Method string
	Args   []any
	Time   time.Time
}

// NewFakeClient creates a new FakeClient with default values.
func NewFakeClient() *FakeClient {
	return &FakeClient{
		accounts:     make(map[aptos.AccountAddress]*aptos.AccountInfo),
		resources:    make(map[string][]aptos.Resource),
		balances:     make(map[aptos.AccountAddress]uint64),
		transactions: make(map[string]*aptos.Transaction),
		blocks:       make(map[uint64]*aptos.Block),
		errors:       make(map[string]error),
		nodeInfo: &aptos.NodeInfo{
			ChainID:       4,
			Epoch:         1,
			LedgerVersion: 1000,
			BlockHeight:   100,
			NodeRole:      "full_node",
		},
		gasEstimate: &aptos.GasEstimate{
			GasEstimate:              100,
			DeprioritizedGasEstimate: 50,
			PrioritizedGasEstimate:   150,
		},
	}
}

// Configuration methods

// WithAccount sets account info for an address.
func (c *FakeClient) WithAccount(addr aptos.AccountAddress, info *aptos.AccountInfo) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accounts[addr] = info
	return c
}

// WithBalance sets the APT balance for an address.
func (c *FakeClient) WithBalance(addr aptos.AccountAddress, balance uint64) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.balances[addr] = balance
	return c
}

// WithResources sets resources for an address.
func (c *FakeClient) WithResources(addr aptos.AccountAddress, resources []aptos.Resource) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resources[addr.String()] = resources
	return c
}

// WithNodeInfo sets the node info response.
func (c *FakeClient) WithNodeInfo(info *aptos.NodeInfo) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodeInfo = info
	return c
}

// WithGasEstimate sets the gas estimate response.
func (c *FakeClient) WithGasEstimate(estimate *aptos.GasEstimate) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gasEstimate = estimate
	return c
}

// WithTransaction adds a transaction to the store.
func (c *FakeClient) WithTransaction(txn *aptos.Transaction) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transactions[txn.Hash] = txn
	return c
}

// WithBlock adds a block to the store.
func (c *FakeClient) WithBlock(block *aptos.Block) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blocks[block.BlockHeight] = block
	return c
}

// WithError configures an error to be returned for a specific method.
func (c *FakeClient) WithError(method string, err error) *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors[method] = err
	return c
}

// WithRecording enables call recording.
func (c *FakeClient) WithRecording() *FakeClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recording = true
	return c
}

// RecordedCalls returns all recorded method calls.
func (c *FakeClient) RecordedCalls() []RecordedCall {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]RecordedCall, len(c.calls))
	copy(result, c.calls)
	return result
}

// ClearRecordedCalls clears all recorded calls.
func (c *FakeClient) ClearRecordedCalls() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = nil
}

// Internal helpers

func (c *FakeClient) record(method string, args ...any) {
	if c.recording {
		c.mu.Lock()
		c.calls = append(c.calls, RecordedCall{
			Method: method,
			Args:   args,
			Time:   time.Now(),
		})
		c.mu.Unlock()
	}
}

func (c *FakeClient) getError(method string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.errors[method]
}

// Client interface implementation

// Info returns information about the node.
func (c *FakeClient) Info(ctx context.Context) (*aptos.NodeInfo, error) {
	c.record("Info")
	if err := c.getError("Info"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodeInfo, nil
}

// ChainID returns the chain ID.
func (c *FakeClient) ChainID(ctx context.Context) (uint8, error) {
	c.record("ChainID")
	if err := c.getError("ChainID"); err != nil {
		return 0, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodeInfo.ChainID, nil
}

// Account returns account information.
func (c *FakeClient) Account(ctx context.Context, address aptos.AccountAddress) (*aptos.AccountInfo, error) {
	c.record("Account", address)
	if err := c.getError("Account"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, ok := c.accounts[address]
	if !ok {
		return nil, aptos.ErrNotFound
	}
	return info, nil
}

// AccountResources returns all resources for an account.
func (c *FakeClient) AccountResources(ctx context.Context, address aptos.AccountAddress, opts ...aptos.ResourceOption) ([]aptos.Resource, error) {
	c.record("AccountResources", address)
	if err := c.getError("AccountResources"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	resources, ok := c.resources[address.String()]
	if !ok {
		return []aptos.Resource{}, nil
	}
	return resources, nil
}

// AccountResource returns a specific resource.
func (c *FakeClient) AccountResource(ctx context.Context, address aptos.AccountAddress, resourceType string, opts ...aptos.ResourceOption) (*aptos.Resource, error) {
	c.record("AccountResource", address, resourceType)
	if err := c.getError("AccountResource"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	resources, ok := c.resources[address.String()]
	if !ok {
		return nil, aptos.ErrNotFound
	}
	for _, r := range resources {
		if r.Type == resourceType {
			return &r, nil
		}
	}
	return nil, aptos.ErrNotFound
}

// AccountBalance returns the APT balance.
func (c *FakeClient) AccountBalance(ctx context.Context, address aptos.AccountAddress, opts ...aptos.ResourceOption) (uint64, error) {
	c.record("AccountBalance", address)
	if err := c.getError("AccountBalance"); err != nil {
		return 0, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	balance, ok := c.balances[address]
	if !ok {
		return 0, nil
	}
	return balance, nil
}

// BuildTransaction builds a transaction.
func (c *FakeClient) BuildTransaction(ctx context.Context, sender aptos.AccountAddress, payload aptos.Payload, opts ...aptos.TransactionOption) (*aptos.RawTransaction, error) {
	c.record("BuildTransaction", sender, payload)
	if err := c.getError("BuildTransaction"); err != nil {
		return nil, err
	}

	// Get sequence number from account or return error
	c.mu.RLock()
	info, ok := c.accounts[sender]
	chainID := c.nodeInfo.ChainID
	c.mu.RUnlock()

	var seqNum uint64
	if ok {
		seqNum = info.SequenceNumber
	}

	return &aptos.RawTransaction{
		Sender:                     sender,
		SequenceNumber:             seqNum,
		Payload:                    payload,
		MaxGasAmount:               200000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: uint64(time.Now().Add(30 * time.Second).Unix()),
		ChainID:                    chainID,
	}, nil
}

// SimulateTransaction simulates a transaction.
func (c *FakeClient) SimulateTransaction(ctx context.Context, txn *aptos.RawTransaction, signer aptos.Signer, opts ...aptos.TransactionOption) (*aptos.SimulationResult, error) {
	c.record("SimulateTransaction", txn)
	if err := c.getError("SimulateTransaction"); err != nil {
		return nil, err
	}
	return &aptos.SimulationResult{
		Success:  true,
		VMStatus: "Executed successfully",
		GasUsed:  1000,
	}, nil
}

// SubmitTransaction submits a signed transaction.
func (c *FakeClient) SubmitTransaction(ctx context.Context, signed *aptos.SignedTransaction) (*aptos.SubmitResult, error) {
	c.record("SubmitTransaction", signed)
	if err := c.getError("SubmitTransaction"); err != nil {
		return nil, err
	}
	// Generate a fake hash
	hash := fmt.Sprintf("0x%064x", time.Now().UnixNano())
	return &aptos.SubmitResult{Hash: hash}, nil
}

// SignAndSubmitTransaction signs and submits a transaction.
func (c *FakeClient) SignAndSubmitTransaction(ctx context.Context, signer aptos.TransactionSigner, payload aptos.Payload, opts ...aptos.TransactionOption) (*aptos.SubmitResult, error) {
	c.record("SignAndSubmitTransaction", signer, payload)
	if err := c.getError("SignAndSubmitTransaction"); err != nil {
		return nil, err
	}
	hash := fmt.Sprintf("0x%064x", time.Now().UnixNano())
	return &aptos.SubmitResult{Hash: hash}, nil
}

// WaitForTransaction waits for a transaction to be confirmed.
func (c *FakeClient) WaitForTransaction(ctx context.Context, hash string, opts ...aptos.PollOption) (*aptos.Transaction, error) {
	c.record("WaitForTransaction", hash)
	if err := c.getError("WaitForTransaction"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	txn, ok := c.transactions[hash]
	c.mu.RUnlock()
	if !ok {
		// Return a successful transaction
		return &aptos.Transaction{
			Type:     "user_transaction",
			Hash:     hash,
			Success:  true,
			VMStatus: "Executed successfully",
			Version:  1000,
		}, nil
	}
	return txn, nil
}

// Transaction returns a transaction by hash.
func (c *FakeClient) Transaction(ctx context.Context, hash string) (*aptos.Transaction, error) {
	c.record("Transaction", hash)
	if err := c.getError("Transaction"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	txn, ok := c.transactions[hash]
	if !ok {
		return nil, aptos.ErrNotFound
	}
	return txn, nil
}

// TransactionByVersion returns a transaction by version.
func (c *FakeClient) TransactionByVersion(ctx context.Context, version uint64) (*aptos.Transaction, error) {
	c.record("TransactionByVersion", version)
	if err := c.getError("TransactionByVersion"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, txn := range c.transactions {
		if txn.Version == version {
			return txn, nil
		}
	}
	return nil, aptos.ErrNotFound
}

// Transactions returns recent transactions.
func (c *FakeClient) Transactions(ctx context.Context, start *uint64, limit *uint64) ([]*aptos.Transaction, error) {
	c.record("Transactions", start, limit)
	if err := c.getError("Transactions"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*aptos.Transaction
	for _, txn := range c.transactions {
		result = append(result, txn)
	}
	return result, nil
}

// TransactionsIter returns an iterator over transactions.
func (c *FakeClient) TransactionsIter(ctx context.Context, start *uint64) iter.Seq2[*aptos.Transaction, error] {
	c.record("TransactionsIter", start)
	return func(yield func(*aptos.Transaction, error) bool) {
		if err := c.getError("TransactionsIter"); err != nil {
			yield(nil, err)
			return
		}
		c.mu.RLock()
		defer c.mu.RUnlock()
		for _, txn := range c.transactions {
			if !yield(txn, nil) {
				return
			}
		}
	}
}

// View executes a view function.
func (c *FakeClient) View(ctx context.Context, payload *aptos.ViewPayload, opts ...aptos.ViewOption) ([]any, error) {
	c.record("View", payload)
	if err := c.getError("View"); err != nil {
		return nil, err
	}
	// Return empty result by default
	return []any{}, nil
}

// BlockByHeight returns a block by height.
func (c *FakeClient) BlockByHeight(ctx context.Context, height uint64, withTransactions bool) (*aptos.Block, error) {
	c.record("BlockByHeight", height, withTransactions)
	if err := c.getError("BlockByHeight"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	block, ok := c.blocks[height]
	if !ok {
		return nil, aptos.ErrNotFound
	}
	return block, nil
}

// BlockByVersion returns the block containing a version.
func (c *FakeClient) BlockByVersion(ctx context.Context, version uint64, withTransactions bool) (*aptos.Block, error) {
	c.record("BlockByVersion", version, withTransactions)
	if err := c.getError("BlockByVersion"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, block := range c.blocks {
		if version >= block.FirstVersion && version <= block.LastVersion {
			return block, nil
		}
	}
	return nil, aptos.ErrNotFound
}

// EstimateGasPrice returns gas price estimates.
func (c *FakeClient) EstimateGasPrice(ctx context.Context) (*aptos.GasEstimate, error) {
	c.record("EstimateGasPrice")
	if err := c.getError("EstimateGasPrice"); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gasEstimate, nil
}

// EventsByHandle returns events for an event handle.
func (c *FakeClient) EventsByHandle(ctx context.Context, address aptos.AccountAddress, handle string, field string, start *uint64, limit *uint64) ([]aptos.Event, error) {
	c.record("EventsByHandle", address, handle, field)
	if err := c.getError("EventsByHandle"); err != nil {
		return nil, err
	}
	return []aptos.Event{}, nil
}

// Fund requests tokens from the faucet.
func (c *FakeClient) Fund(ctx context.Context, address aptos.AccountAddress, amount uint64) error {
	c.record("Fund", address, amount)
	if err := c.getError("Fund"); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.balances[address] += amount
	// Create account if it doesn't exist
	if _, ok := c.accounts[address]; !ok {
		c.accounts[address] = &aptos.AccountInfo{
			SequenceNumber:    0,
			AuthenticationKey: address.String(),
		}
	}
	return nil
}

// HealthCheck checks the health of the node.
func (c *FakeClient) HealthCheck(ctx context.Context, durationSecs ...uint64) (*aptos.HealthCheckResponse, error) {
	c.record("HealthCheck", durationSecs)
	if err := c.getError("HealthCheck"); err != nil {
		return nil, err
	}
	return &aptos.HealthCheckResponse{Message: "aptos-node:ok"}, nil
}

// AccountModule returns module bytecode and ABI.
func (c *FakeClient) AccountModule(ctx context.Context, address aptos.AccountAddress, moduleName string, opts ...aptos.ResourceOption) (*aptos.ModuleBytecode, error) {
	c.record("AccountModule", address, moduleName)
	if err := c.getError("AccountModule"); err != nil {
		return nil, err
	}
	return &aptos.ModuleBytecode{
		Bytecode: "0x",
		ABI: &aptos.ModuleABI{
			Address: address.String(),
			Name:    moduleName,
		},
	}, nil
}

// AccountTransactions returns transactions sent by an account.
func (c *FakeClient) AccountTransactions(ctx context.Context, address aptos.AccountAddress, start *uint64, limit *uint64) ([]*aptos.Transaction, error) {
	c.record("AccountTransactions", address, start, limit)
	if err := c.getError("AccountTransactions"); err != nil {
		return nil, err
	}
	return []*aptos.Transaction{}, nil
}

// BatchSubmitTransaction submits multiple transactions.
func (c *FakeClient) BatchSubmitTransaction(ctx context.Context, signed []*aptos.SignedTransaction) (*aptos.BatchSubmitResult, error) {
	c.record("BatchSubmitTransaction", signed)
	if err := c.getError("BatchSubmitTransaction"); err != nil {
		return nil, err
	}
	return &aptos.BatchSubmitResult{}, nil
}

// EventsByCreationNumber returns events by creation number.
func (c *FakeClient) EventsByCreationNumber(ctx context.Context, address aptos.AccountAddress, creationNumber uint64, start *uint64, limit *uint64) ([]aptos.Event, error) {
	c.record("EventsByCreationNumber", address, creationNumber, start, limit)
	if err := c.getError("EventsByCreationNumber"); err != nil {
		return nil, err
	}
	return []aptos.Event{}, nil
}

// Ensure FakeClient implements Client
var _ aptos.Client = (*FakeClient)(nil)
