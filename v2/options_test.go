package aptos

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithTimeout", func(t *testing.T) {
		t.Parallel()
		config := &ClientConfig{}
		WithTimeout(5 * time.Second)(config)
		assert.Equal(t, 5*time.Second, config.timeout)
	})

	t.Run("WithLogger", func(t *testing.T) {
		t.Parallel()
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		config := &ClientConfig{}
		WithLogger(logger)(config)
		assert.Equal(t, logger, config.logger)
	})

	t.Run("WithRetry", func(t *testing.T) {
		t.Parallel()
		config := &ClientConfig{}
		WithRetry(5, 200*time.Millisecond)(config)
		assert.NotNil(t, config.retryConfig)
		assert.Equal(t, 5, config.retryConfig.MaxRetries)
		assert.Equal(t, 200*time.Millisecond, config.retryConfig.InitialBackoff)
	})

	t.Run("WithRetryConfig", func(t *testing.T) {
		t.Parallel()
		retryConfig := &RetryConfig{
			MaxRetries:        10,
			InitialBackoff:    500 * time.Millisecond,
			MaxBackoff:        30 * time.Second,
			BackoffMultiplier: 1.5,
		}
		config := &ClientConfig{}
		WithRetryConfig(retryConfig)(config)
		assert.Equal(t, retryConfig, config.retryConfig)
	})

	t.Run("WithRateLimitHandling", func(t *testing.T) {
		t.Parallel()
		config := &ClientConfig{}
		WithRateLimitHandling(true, 30*time.Second)(config)
		assert.NotNil(t, config.rateLimitConfig)
		assert.True(t, config.rateLimitConfig.Enabled)
		assert.True(t, config.rateLimitConfig.WaitOnLimit)
		assert.Equal(t, 30*time.Second, config.rateLimitConfig.MaxWaitTime)
	})

	t.Run("WithHeader", func(t *testing.T) {
		t.Parallel()
		config := &ClientConfig{}
		WithHeader("X-Custom", "value")(config)
		assert.Equal(t, "value", config.headers["X-Custom"])
	})

	t.Run("WithMultipleHeaders", func(t *testing.T) {
		t.Parallel()
		config := &ClientConfig{}
		WithHeader("X-First", "1")(config)
		WithHeader("X-Second", "2")(config)
		assert.Equal(t, "1", config.headers["X-First"])
		assert.Equal(t, "2", config.headers["X-Second"])
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		t.Parallel()
		config := &ClientConfig{}
		WithAPIKey("my-secret-key")(config)
		assert.Equal(t, "Bearer my-secret-key", config.headers["Authorization"])
	})
}

func TestTransactionOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithMaxGas", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithMaxGas(100000)(config)
		assert.Equal(t, uint64(100000), config.MaxGasAmount)
	})

	t.Run("WithGasPrice", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithGasPrice(150)(config)
		assert.Equal(t, uint64(150), config.GasUnitPrice)
	})

	t.Run("WithExpiration", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithExpiration(2 * time.Minute)(config)
		assert.Equal(t, 2*time.Minute, config.ExpirationDuration)
	})

	t.Run("WithSequenceNumber", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithSequenceNumber(42)(config)
		assert.NotNil(t, config.SequenceNumber)
		assert.Equal(t, uint64(42), *config.SequenceNumber)
	})

	t.Run("WithSimulation", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithSimulation()(config)
		assert.True(t, config.Simulate)
	})

	t.Run("WithGasEstimation", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithGasEstimation()(config)
		assert.True(t, config.EstimateGas)
		assert.False(t, config.EstimatePrioritizedGas)
	})

	t.Run("WithPrioritizedGas", func(t *testing.T) {
		t.Parallel()
		config := &TransactionConfig{}
		WithPrioritizedGas()(config)
		assert.True(t, config.EstimateGas)
		assert.True(t, config.EstimatePrioritizedGas)
	})
}

func TestViewOptions(t *testing.T) {
	t.Parallel()

	t.Run("AtLedgerVersion", func(t *testing.T) {
		t.Parallel()
		config := &ViewConfig{}
		AtLedgerVersion(12345)(config)
		assert.NotNil(t, config.LedgerVersion)
		assert.Equal(t, uint64(12345), *config.LedgerVersion)
	})
}

func TestPollOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithPollInterval", func(t *testing.T) {
		t.Parallel()
		config := &PollConfig{}
		WithPollInterval(500 * time.Millisecond)(config)
		assert.Equal(t, 500*time.Millisecond, config.PollInterval)
	})

	t.Run("WithPollTimeout", func(t *testing.T) {
		t.Parallel()
		config := &PollConfig{}
		WithPollTimeout(30 * time.Second)(config)
		assert.Equal(t, 30*time.Second, config.Timeout)
	})
}

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	config := DefaultRetryConfig()
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.InitialBackoff)
	assert.Equal(t, 10*time.Second, config.MaxBackoff)
	assert.Equal(t, 2.0, config.BackoffMultiplier)
	assert.Contains(t, config.RetryableStatusCodes, 429)
	assert.Contains(t, config.RetryableStatusCodes, 503)
}
