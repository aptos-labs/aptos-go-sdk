// Package http provides HTTP client utilities for the Aptos SDK.
//
// This package includes:
//   - HTTPDoer interface for custom HTTP clients
//   - Retry middleware with exponential backoff
//   - Rate limiting support
//   - Request/response logging
package http

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// HTTPDoer abstracts HTTP operations for testing and customization.
// This is compatible with *http.Client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RetryConfig defines parameters for HTTP request retries.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 means no retries).
	MaxRetries int

	// InitialBackoff is the initial backoff duration before the first retry.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// Multiplier is the factor by which backoff increases each retry.
	Multiplier float64

	// Jitter adds randomness to backoff to prevent thundering herd.
	Jitter float64

	// RetryableStatusCodes defines which HTTP status codes should be retried.
	// If nil, defaults to 429 (rate limited) and 5xx errors.
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns sensible default retry settings.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.1,
		RetryableStatusCodes: []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	}
}

// RetryClient wraps an HTTPDoer with retry logic.
type RetryClient struct {
	inner  HTTPDoer
	config RetryConfig
	logger *slog.Logger
}

// NewRetryClient creates a new RetryClient.
func NewRetryClient(inner HTTPDoer, config RetryConfig, logger *slog.Logger) *RetryClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &RetryClient{
		inner:  inner,
		config: config,
		logger: logger,
	}
}

// Do executes the request with retries.
func (c *RetryClient) Do(req *http.Request) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff
			backoff := c.calculateBackoff(attempt)

			c.logger.Debug("retrying request",
				"attempt", attempt,
				"max_retries", c.config.MaxRetries,
				"backoff", backoff,
				"url", req.URL.String(),
			)

			// Check for context cancellation
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(backoff):
			}

			// Clone request body if needed for retry
			if req.Body != nil && req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to get request body for retry: %w", err)
				}
				req.Body = body
			}
		}

		resp, err := c.inner.Do(req)
		if err != nil {
			// Network errors are retriable
			lastErr = err
			lastResp = nil
			c.logger.Debug("request failed with error",
				"attempt", attempt,
				"error", err,
				"url", req.URL.String(),
			)
			continue
		}

		// Check if status code is retriable
		if c.isRetriable(resp.StatusCode) {
			// Check for Retry-After header
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					time.Sleep(time.Duration(seconds) * time.Second)
				}
			}

			// Close response body to allow reuse of connection
			if resp.Body != nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}

			lastResp = resp
			lastErr = fmt.Errorf("retriable status code: %d", resp.StatusCode)
			c.logger.Debug("retriable status code",
				"attempt", attempt,
				"status", resp.StatusCode,
				"url", req.URL.String(),
			)
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return lastResp, lastErr
	}
	return lastResp, nil
}

func (c *RetryClient) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.config.InitialBackoff) * math.Pow(c.config.Multiplier, float64(attempt-1))

	// Apply jitter
	jitter := backoff * c.config.Jitter * (rand.Float64()*2 - 1)
	backoff += jitter

	// Cap at max backoff
	if backoff > float64(c.config.MaxBackoff) {
		backoff = float64(c.config.MaxBackoff)
	}

	return time.Duration(backoff)
}

func (c *RetryClient) isRetriable(statusCode int) bool {
	for _, code := range c.config.RetryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// RateLimitedClient wraps an HTTPDoer with rate limiting.
type RateLimitedClient struct {
	inner   HTTPDoer
	limiter *RateLimiter
	logger  *slog.Logger
}

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	rate       float64 // tokens per second
	burst      int     // max tokens
	tokens     float64 // current tokens
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       requestsPerSecond,
		burst:      burst,
		tokens:     float64(burst),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.refill()

		if r.tokens >= 1.0 {
			r.tokens -= 1.0
			return nil
		}

		// Calculate time until next token
		waitTime := time.Duration(float64(time.Second) / r.rate)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}

func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	r.lastRefill = now

	tokensToAdd := elapsed.Seconds() * r.rate
	r.tokens = math.Min(float64(r.burst), r.tokens+tokensToAdd)
}

// NewRateLimitedClient creates a new rate-limited client.
func NewRateLimitedClient(inner HTTPDoer, requestsPerSecond float64, burst int, logger *slog.Logger) *RateLimitedClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &RateLimitedClient{
		inner:   inner,
		limiter: NewRateLimiter(requestsPerSecond, burst),
		logger:  logger,
	}
}

// Do executes the request after waiting for rate limit.
func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	if err := c.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return c.inner.Do(req)
}

// LoggingClient wraps an HTTPDoer with request/response logging.
type LoggingClient struct {
	inner  HTTPDoer
	logger *slog.Logger
}

// NewLoggingClient creates a new logging client.
func NewLoggingClient(inner HTTPDoer, logger *slog.Logger) *LoggingClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingClient{
		inner:  inner,
		logger: logger,
	}
}

// Do executes the request with logging.
func (c *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()

	c.logger.Debug("sending request",
		"method", req.Method,
		"url", req.URL.String(),
	)

	resp, err := c.inner.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Debug("request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"error", err,
			"duration", duration,
		)
		return nil, err
	}

	c.logger.Debug("request completed",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"duration", duration,
	)

	return resp, nil
}

// HeaderClient wraps an HTTPDoer to add custom headers to all requests.
type HeaderClient struct {
	inner   HTTPDoer
	headers map[string]string
}

// NewHeaderClient creates a client that adds headers to all requests.
func NewHeaderClient(inner HTTPDoer, headers map[string]string) *HeaderClient {
	return &HeaderClient{
		inner:   inner,
		headers: headers,
	}
}

// Do executes the request with additional headers.
func (c *HeaderClient) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return c.inner.Do(req)
}

// TimeoutClient wraps an HTTPDoer with a default timeout.
type TimeoutClient struct {
	inner   HTTPDoer
	timeout time.Duration
}

// NewTimeoutClient creates a client with a default timeout.
func NewTimeoutClient(inner HTTPDoer, timeout time.Duration) *TimeoutClient {
	return &TimeoutClient{
		inner:   inner,
		timeout: timeout,
	}
}

// Do executes the request with a timeout if not already set.
func (c *TimeoutClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	return c.inner.Do(req)
}
