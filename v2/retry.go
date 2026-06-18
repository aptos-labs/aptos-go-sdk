package aptos

import (
	"context"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// retryHTTPClient wraps an HTTPDoer with automatic retry and rate-limit
// handling. It is installed by newNodeClient when WithRetry, WithRetryConfig,
// or WithRateLimitHandling are supplied.
//
// Retries are attempted on transient failures (network errors and the
// configured retryable status codes) using exponential backoff with jitter.
// When rate-limit handling is enabled the client additionally honors the
// Retry-After header returned with HTTP 429 responses, capped at MaxWaitTime.
type retryHTTPClient struct {
	inner     HTTPDoer
	retry     *RetryConfig
	rateLimit *RateLimitConfig
	logger    *slog.Logger
}

// newRetryHTTPClient builds a retrying HTTPDoer. It returns inner unchanged
// unless at least one retry path can actually trigger — i.e. error/status
// retries are enabled (retry.MaxRetries > 0) or 429 rate-limit handling is
// enabled (rateLimit.Enabled && rateLimit.WaitOnLimit). This keeps the common
// path (and any effectively-disabled configuration) free of wrapper overhead.
func newRetryHTTPClient(inner HTTPDoer, retry *RetryConfig, rateLimit *RateLimitConfig, logger *slog.Logger) HTTPDoer {
	retryActive := retry != nil && retry.MaxRetries > 0
	rateLimitActive := rateLimit != nil && rateLimit.Enabled && rateLimit.WaitOnLimit
	if !retryActive && !rateLimitActive {
		return inner
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &retryHTTPClient{
		inner:     inner,
		retry:     retry,
		rateLimit: rateLimit,
		logger:    logger,
	}
}

// rateLimitMaxRetries bounds how many times a 429 response is retried when
// rate-limit handling is enabled, so a misbehaving server can't make us loop
// forever. Overall waiting is additionally bounded by MaxWaitTime per attempt
// and by the request context.
const rateLimitMaxRetries = 5

// errorStatusCap is the strict maximum number of retries for network errors
// and retryable status codes. It is exactly MaxRetries (0 when retries are
// disabled or no retry config is set) and is never inflated by enabling
// rate-limit handling.
func (c *retryHTTPClient) errorStatusCap() int {
	if c.retry != nil && c.retry.MaxRetries > 0 {
		return c.retry.MaxRetries
	}
	return 0
}

// rateLimitCap is the maximum number of 429 rate-limit retries.
func (c *retryHTTPClient) rateLimitCap() int {
	if c.rateLimitActive() {
		return rateLimitMaxRetries
	}
	return 0
}

// rateLimitActive reports whether 429 Retry-After handling is enabled.
func (c *retryHTTPClient) rateLimitActive() bool {
	return c.rateLimit != nil && c.rateLimit.Enabled && c.rateLimit.WaitOnLimit
}

// retriesEnabled reports whether network-error and retryable-status-code
// retries are active. A retry config with MaxRetries == 0 explicitly disables
// them (only 429 rate-limit handling, if enabled, remains active).
func (c *retryHTTPClient) retriesEnabled() bool {
	return c.errorStatusCap() > 0
}

// maxAttempts returns the total number of attempts (including the first) the
// client will make. Error/status retries and 429 retries are capped
// independently, so the worst case is one initial attempt plus both caps.
// This keeps MaxRetries a strict cap on error/status retries even when
// rate-limit handling reserves additional attempts for 429s.
func (c *retryHTTPClient) maxAttempts() int {
	return 1 + c.errorStatusCap() + c.rateLimitCap()
}

// retryReason classifies why (or whether) a response/error warrants a retry.
type retryReason int

const (
	retryNone retryReason = iota
	retryErrorStatus
	retryRateLimit
)

// classifyRetry determines the retry category for a response/error and, for
// rate-limited responses, the Retry-After wait. The two categories are tracked
// against separate caps by Do.
func (c *retryHTTPClient) classifyRetry(resp *http.Response, err error) (retryReason, time.Duration) {
	// Network/transport errors are retried only when error/status retries are
	// enabled (an explicit MaxRetries == 0 disables them).
	if err != nil {
		if c.retriesEnabled() {
			return retryErrorStatus, 0
		}
		return retryNone, 0
	}
	if resp == nil {
		return retryNone, 0
	}

	// Rate-limit handling: honor Retry-After (capped) on 429. This stays
	// active even when error/status retries are disabled.
	if resp.StatusCode == http.StatusTooManyRequests && c.rateLimitActive() {
		return retryRateLimit, c.retryAfter(resp)
	}

	if c.retriesEnabled() && c.isRetriableStatus(resp.StatusCode) {
		return retryErrorStatus, 0
	}

	return retryNone, 0
}

// Do executes the request, retrying transient failures per the configuration.
func (c *retryHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	attempts := c.maxAttempts()

	// Keep req.Context() consistent with the ctx we pass to the inner doer so
	// cancellation/timeouts propagate even for a custom HTTPDoer that reads
	// req.Context() instead of the explicit ctx argument.
	req = req.WithContext(ctx)

	errorStatusCap := c.errorStatusCap()
	rateLimitCap := c.rateLimitCap()
	var errorStatusRetries, rateLimitRetries int

	var resp *http.Response
	var err error
	for attempt := range attempts {
		// Re-buffer the request body for every attempt after the first.
		// http.NewRequestWithContext populates GetBody for the in-memory
		// body types the SDK uses (bytes.Reader), so this is reliable.
		if attempt > 0 && req.Body != nil && req.GetBody != nil {
			body, gerr := req.GetBody()
			if gerr != nil {
				return nil, gerr
			}
			req.Body = body
		}

		resp, err = c.inner.Do(ctx, req)

		reason, wait := c.classifyRetry(resp, err)

		// Each retry category is capped independently. MaxRetries is a strict
		// cap on error/status retries and is not raised by enabling rate-limit
		// handling; 429s have their own bounded budget.
		proceed := false
		switch reason {
		case retryErrorStatus:
			if errorStatusRetries < errorStatusCap {
				errorStatusRetries++
				proceed = true
			}
		case retryRateLimit:
			if rateLimitRetries < rateLimitCap {
				rateLimitRetries++
				proceed = true
			}
		case retryNone:
		}
		if !proceed {
			return resp, err
		}

		// Drain and close the response body so the connection can be
		// reused, then back off before retrying.
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		retryNum := errorStatusRetries + rateLimitRetries
		backoff := c.backoff(retryNum)
		if wait > backoff {
			backoff = wait
		}

		c.logger.Debug(
			"retrying request",
			"attempt", retryNum,
			"max_attempts", attempts,
			"backoff", backoff,
			"url", req.URL.String(),
		)

		// Use an explicit timer (stopped on cancellation) rather than
		// time.After so a cancelled wait doesn't leave a timer running until
		// it fires.
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return resp, err
}

// retryAfter parses the Retry-After header (seconds form) and caps it at
// MaxWaitTime. Returns 0 when the header is absent or unparseable.
func (c *retryHTTPClient) retryAfter(resp *http.Response) time.Duration {
	header := resp.Header.Get("Retry-After")
	if header == "" {
		return 0
	}
	seconds, perr := strconv.Atoi(header)
	if perr != nil || seconds < 0 {
		return 0
	}
	wait := time.Duration(seconds) * time.Second
	if c.rateLimit.MaxWaitTime > 0 && wait > c.rateLimit.MaxWaitTime {
		wait = c.rateLimit.MaxWaitTime
	}
	return wait
}

func (c *retryHTTPClient) isRetriableStatus(statusCode int) bool {
	for _, code := range c.retry.RetryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// backoff computes the exponential backoff (with +/-10% jitter) for the given
// attempt number (1-based), capped at MaxBackoff.
func (c *retryHTTPClient) backoff(attempt int) time.Duration {
	if c.retry == nil {
		// Rate-limit-only mode: a small fixed base so non-Retry-After 429s
		// (or empty headers) still pace themselves.
		return 100 * time.Millisecond
	}

	base := float64(c.retry.InitialBackoff)
	if base <= 0 {
		base = float64(100 * time.Millisecond)
	}
	mult := c.retry.BackoffMultiplier
	if mult <= 0 {
		mult = 2.0
	}

	backoff := base * math.Pow(mult, float64(attempt-1))
	// +/-10% jitter to avoid synchronized retries (thundering herd).
	backoff += backoff * 0.1 * (rand.Float64()*2 - 1) //nolint:gosec // non-crypto jitter
	if maxB := float64(c.retry.MaxBackoff); maxB > 0 && backoff > maxB {
		backoff = maxB
	}
	if backoff < 0 {
		backoff = 0
	}
	return time.Duration(backoff)
}
