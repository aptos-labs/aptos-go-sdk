package aptos

import (
	"log/slog"
	"time"
)

// ClientConfig holds the configuration for an Aptos client.
type ClientConfig struct {
	// Network configuration
	network NetworkConfig

	// HTTP client configuration
	httpClient HTTPDoer
	timeout    time.Duration
	headers    map[string]string

	// Retry configuration
	retryConfig *RetryConfig

	// Logging
	logger *slog.Logger

	// Rate limiting
	rateLimitConfig *RateLimitConfig
}

// RetryConfig configures automatic retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff.
	BackoffMultiplier float64

	// RetryableStatusCodes are HTTP status codes that trigger a retry.
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        10 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableStatusCodes: []int{
			429, // Too Many Requests
			500, // Internal Server Error
			502, // Bad Gateway
			503, // Service Unavailable
			504, // Gateway Timeout
		},
	}
}

// RateLimitConfig configures rate limiting behavior.
type RateLimitConfig struct {
	// Enabled enables automatic rate limit handling.
	Enabled bool

	// WaitOnLimit determines whether to wait when rate limited.
	WaitOnLimit bool

	// MaxWaitTime is the maximum time to wait when rate limited.
	MaxWaitTime time.Duration
}

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*ClientConfig)

// WithHTTPClient sets a custom HTTP client.
// The client must implement the HTTPDoer interface.
func WithHTTPClient(client HTTPDoer) ClientOption {
	return func(c *ClientConfig) {
		c.httpClient = client
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.timeout = timeout
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *ClientConfig) {
		c.logger = logger
	}
}

// WithRetry configures automatic retry behavior.
func WithRetry(maxRetries int, initialBackoff time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.retryConfig = &RetryConfig{
			MaxRetries:           maxRetries,
			InitialBackoff:       initialBackoff,
			MaxBackoff:           10 * time.Second,
			BackoffMultiplier:    2.0,
			RetryableStatusCodes: []int{429, 500, 502, 503, 504},
		}
	}
}

// WithRetryConfig sets a custom retry configuration.
func WithRetryConfig(config *RetryConfig) ClientOption {
	return func(c *ClientConfig) {
		c.retryConfig = config
	}
}

// WithRateLimitHandling enables automatic rate limit handling.
func WithRateLimitHandling(waitOnLimit bool, maxWait time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.rateLimitConfig = &RateLimitConfig{
			Enabled:     true,
			WaitOnLimit: waitOnLimit,
			MaxWaitTime: maxWait,
		}
	}
}

// WithHeader adds a custom header to all requests.
func WithHeader(key, value string) ClientOption {
	return func(c *ClientConfig) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		c.headers[key] = value
	}
}

// WithAPIKey sets the API key header for authenticated requests.
func WithAPIKey(apiKey string) ClientOption {
	return WithHeader("Authorization", "Bearer "+apiKey)
}

// TransactionOption is a functional option for transaction building.
type TransactionOption func(*TransactionConfig)

// TransactionConfig holds configuration for building a transaction.
type TransactionConfig struct {
	// MaxGasAmount is the maximum gas units to use.
	MaxGasAmount uint64

	// GasUnitPrice is the gas unit price in octas.
	GasUnitPrice uint64

	// ExpirationDuration is how long until the transaction expires.
	ExpirationDuration time.Duration

	// SequenceNumber overrides automatic sequence number fetching.
	SequenceNumber *uint64

	// Simulate determines whether to simulate before submitting.
	Simulate bool

	// EstimateGas enables automatic gas estimation.
	EstimateGas bool

	// EstimatePrioritizedGas uses prioritized gas estimation.
	EstimatePrioritizedGas bool

	// SecondarySigners are additional signers for multi-agent transactions.
	SecondarySigners []AccountAddress

	// FeePayer is the fee payer address for sponsored transactions.
	FeePayer *AccountAddress
}

// WithMaxGas sets the maximum gas amount.
func WithMaxGas(amount uint64) TransactionOption {
	return func(c *TransactionConfig) {
		c.MaxGasAmount = amount
	}
}

// WithGasPrice sets the gas unit price.
func WithGasPrice(price uint64) TransactionOption {
	return func(c *TransactionConfig) {
		c.GasUnitPrice = price
	}
}

// WithExpiration sets the transaction expiration duration.
func WithExpiration(d time.Duration) TransactionOption {
	return func(c *TransactionConfig) {
		c.ExpirationDuration = d
	}
}

// WithSequenceNumber sets a specific sequence number.
func WithSequenceNumber(seqNum uint64) TransactionOption {
	return func(c *TransactionConfig) {
		c.SequenceNumber = &seqNum
	}
}

// WithSimulation enables transaction simulation before submission.
func WithSimulation() TransactionOption {
	return func(c *TransactionConfig) {
		c.Simulate = true
	}
}

// WithGasEstimation enables automatic gas estimation.
func WithGasEstimation() TransactionOption {
	return func(c *TransactionConfig) {
		c.EstimateGas = true
	}
}

// WithPrioritizedGas uses prioritized gas estimation for faster inclusion.
func WithPrioritizedGas() TransactionOption {
	return func(c *TransactionConfig) {
		c.EstimateGas = true
		c.EstimatePrioritizedGas = true
	}
}

// WithSecondarySigners sets secondary signers for multi-agent transactions.
func WithSecondarySigners(signers ...AccountAddress) TransactionOption {
	return func(c *TransactionConfig) {
		c.SecondarySigners = signers
	}
}

// WithFeePayer sets the fee payer for sponsored transactions.
func WithFeePayer(feePayer AccountAddress) TransactionOption {
	return func(c *TransactionConfig) {
		c.FeePayer = &feePayer
	}
}

// ViewOption is a functional option for view function calls.
type ViewOption func(*ViewConfig)

// ViewConfig holds configuration for view function calls.
type ViewConfig struct {
	// LedgerVersion specifies a specific ledger version to query.
	LedgerVersion *uint64
}

// AtLedgerVersion queries at a specific ledger version.
func AtLedgerVersion(version uint64) ViewOption {
	return func(c *ViewConfig) {
		c.LedgerVersion = &version
	}
}

// ResourceOption is a functional option for resource queries.
type ResourceOption func(*ResourceConfig)

// ResourceConfig holds configuration for resource queries.
type ResourceConfig struct {
	// LedgerVersion specifies a specific ledger version to query.
	LedgerVersion *uint64
}

// PollOption is a functional option for polling operations.
type PollOption func(*PollConfig)

// PollConfig holds configuration for polling operations.
type PollConfig struct {
	// PollInterval is the time between poll attempts.
	PollInterval time.Duration

	// Timeout is the maximum time to wait.
	Timeout time.Duration
}

// WithPollInterval sets the polling interval.
func WithPollInterval(d time.Duration) PollOption {
	return func(c *PollConfig) {
		c.PollInterval = d
	}
}

// WithPollTimeout sets the polling timeout.
func WithPollTimeout(d time.Duration) PollOption {
	return func(c *PollConfig) {
		c.Timeout = d
	}
}
