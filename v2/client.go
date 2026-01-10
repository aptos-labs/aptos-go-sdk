package aptos

import (
	"context"
	"iter"
	"net/http"
	"time"
)

// HTTPDoer is an interface for making HTTP requests.
// This abstraction allows for custom HTTP clients, middleware,
// and easy mocking in tests.
type HTTPDoer interface {
	// Do executes an HTTP request and returns the response.
	// The context should be used for cancellation and timeouts.
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// Note: Signer, PublicKey, Signature, and AuthenticationKey are re-exported
// from internal/crypto in crypto.go

// Client is the main interface for interacting with the Aptos blockchain.
// All methods accept a context.Context for cancellation and timeout control.
type Client interface {
	// Node Information

	// Info returns information about the connected node.
	Info(ctx context.Context) (*NodeInfo, error)

	// ChainID returns the chain ID of the connected network.
	ChainID(ctx context.Context) (uint8, error)

	// HealthCheck checks the health of the node.
	// If durationSecs is provided, checks if node is within that many seconds of current time.
	HealthCheck(ctx context.Context, durationSecs ...uint64) (*HealthCheckResponse, error)

	// Account Operations

	// Account returns information about an account.
	Account(ctx context.Context, address AccountAddress) (*AccountInfo, error)

	// AccountResources returns all resources for an account.
	AccountResources(ctx context.Context, address AccountAddress, opts ...ResourceOption) ([]Resource, error)

	// AccountResource returns a specific resource for an account.
	AccountResource(ctx context.Context, address AccountAddress, resourceType string, opts ...ResourceOption) (*Resource, error)

	// AccountBalance returns the APT balance for an account.
	AccountBalance(ctx context.Context, address AccountAddress, opts ...ResourceOption) (uint64, error)

	// AccountModule returns the bytecode and ABI for a module.
	AccountModule(ctx context.Context, address AccountAddress, moduleName string, opts ...ResourceOption) (*ModuleBytecode, error)

	// AccountTransactions returns transactions sent by an account.
	AccountTransactions(ctx context.Context, address AccountAddress, start *uint64, limit *uint64) ([]*Transaction, error)

	// Transaction Operations

	// BuildTransaction builds an unsigned transaction.
	BuildTransaction(ctx context.Context, sender AccountAddress, payload Payload, opts ...TransactionOption) (*RawTransaction, error)

	// SimulateTransaction simulates a transaction without submitting it.
	SimulateTransaction(ctx context.Context, txn *RawTransaction, signer Signer, opts ...TransactionOption) (*SimulationResult, error)

	// SubmitTransaction submits a signed transaction to the network.
	SubmitTransaction(ctx context.Context, signed *SignedTransaction) (*SubmitResult, error)

	// BatchSubmitTransaction submits multiple signed transactions in a single request.
	BatchSubmitTransaction(ctx context.Context, signed []*SignedTransaction) (*BatchSubmitResult, error)

	// SignAndSubmitTransaction signs and submits a transaction.
	// The signer must implement TransactionSigner to provide both signing and address.
	SignAndSubmitTransaction(ctx context.Context, signer TransactionSigner, payload Payload, opts ...TransactionOption) (*SubmitResult, error)

	// WaitForTransaction waits for a transaction to be confirmed.
	WaitForTransaction(ctx context.Context, hash string, opts ...PollOption) (*Transaction, error)

	// Transaction returns a transaction by hash.
	Transaction(ctx context.Context, hash string) (*Transaction, error)

	// TransactionByVersion returns a transaction by version.
	TransactionByVersion(ctx context.Context, version uint64) (*Transaction, error)

	// Transactions returns recent transactions.
	Transactions(ctx context.Context, start *uint64, limit *uint64) ([]*Transaction, error)

	// TransactionsIter returns an iterator over transactions.
	TransactionsIter(ctx context.Context, start *uint64) iter.Seq2[*Transaction, error]

	// View Functions

	// View executes a view function and returns the results.
	View(ctx context.Context, payload *ViewPayload, opts ...ViewOption) ([]any, error)

	// Block Operations

	// BlockByHeight returns a block by height.
	BlockByHeight(ctx context.Context, height uint64, withTransactions bool) (*Block, error)

	// BlockByVersion returns the block containing a specific version.
	BlockByVersion(ctx context.Context, version uint64, withTransactions bool) (*Block, error)

	// Gas Estimation

	// EstimateGasPrice returns gas price estimates.
	EstimateGasPrice(ctx context.Context) (*GasEstimate, error)

	// Events

	// EventsByHandle returns events for an event handle.
	EventsByHandle(ctx context.Context, address AccountAddress, handle string, field string, start *uint64, limit *uint64) ([]Event, error)

	// EventsByCreationNumber returns events by creation number.
	EventsByCreationNumber(ctx context.Context, address AccountAddress, creationNumber uint64, start *uint64, limit *uint64) ([]Event, error)

	// Faucet (testnet/devnet only)

	// Fund requests tokens from the faucet.
	Fund(ctx context.Context, address AccountAddress, amount uint64) error
}

// Payload is an interface for transaction payloads.
type Payload interface {
	// payloadType returns the payload type identifier.
	payloadType() string
}

// RawTransaction represents an unsigned transaction.
type RawTransaction struct {
	Sender                     AccountAddress
	SequenceNumber             uint64
	Payload                    Payload
	MaxGasAmount               uint64
	GasUnitPrice               uint64
	ExpirationTimestampSeconds uint64
	ChainID                    uint8
}

// SignedTransaction represents a signed transaction ready for submission.
type SignedTransaction struct {
	Transaction   *RawTransaction
	Authenticator TransactionAuthenticator
}

// ViewPayload represents the payload for a view function call.
type ViewPayload struct {
	Module   ModuleID
	Function string
	TypeArgs []TypeTag
	Args     []any
}

func (v *ViewPayload) payloadType() string {
	return "view_function"
}

// EntryFunctionPayload represents an entry function call.
type EntryFunctionPayload struct {
	Module   ModuleID
	Function string
	TypeArgs []TypeTag
	Args     []any
}

func (e *EntryFunctionPayload) payloadType() string {
	return "entry_function_payload"
}

// ScriptPayload represents a script execution.
type ScriptPayload struct {
	Code     []byte
	TypeArgs []TypeTag
	Args     []any
}

func (s *ScriptPayload) payloadType() string {
	return "script_payload"
}

// TypeTag is re-exported from types.go as ConcreteTypeTag (an alias for internal/types.TypeTag).
// The interface version is removed in favor of the concrete BCS-serializable type.

// NewClient creates a new Aptos client for the specified network.
func NewClient(network NetworkConfig, opts ...ClientOption) (Client, error) {
	config := &ClientConfig{
		network: network,
		timeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(config)
	}

	return newNodeClient(config)
}
