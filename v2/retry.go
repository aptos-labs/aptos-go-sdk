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

// maxAttempts returns the total number of attempts (including the first) the
// client will make.
//
// An explicit retry config with MaxRetries == 0 means "no error/status
// retries" and is honored as such (it does not silently inherit the
// rate-limit default). When rate-limit handling is enabled we still allow a
// bounded number of attempts so 429 responses can be retried even if
// error/status retries are disabled.
func (c *retryHTTPClient) maxAttempts() int {
	attempts := 1
	if c.retry != nil && c.retry.MaxRetries > 0 {
		attempts = c.retry.MaxRetries + 1
	}
	// Rate-limit handling needs to retry 429s even when error/status retries
	// are disabled (retry config absent or MaxRetries == 0). Use a sensible
	// bounded default so a misbehaving server can't make us loop forever;
	// overall waiting is additionally bounded by MaxWaitTime per attempt and
	// by the request context.
	if c.rateLimit != nil && c.rateLimit.Enabled && c.rateLimit.WaitOnLimit {
		const rateLimitAttempts = 6
		if rateLimitAttempts > attempts {
			attempts = rateLimitAttempts
		}
	}
	return attempts
}

// retriesEnabled reports whether network-error and retryable-status-code
// retries are active. A retry config with MaxRetries == 0 explicitly disables
// them (only 429 rate-limit handling, if enabled, remains active).
func (c *retryHTTPClient) retriesEnabled() bool {
	return c.retry != nil && c.retry.MaxRetries > 0
}

// Do executes the request, retrying transient failures per the configuration.
func (c *retryHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	attempts := c.maxAttempts()

	// Keep req.Context() consistent with the ctx we pass to the inner doer so
	// cancellation/timeouts propagate even for a custom HTTPDoer that reads
	// req.Context() instead of the explicit ctx argument.
	req = req.WithContext(ctx)

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

		// Last attempt: return whatever we got.
		if attempt == attempts-1 {
			return resp, err
		}

		retry, wait := c.shouldRetry(resp, err)
		if !retry {
			return resp, err
		}

		// Drain and close the response body so the connection can be
		// reused, then back off before retrying.
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		backoff := c.backoff(attempt + 1)
		if wait > backoff {
			backoff = wait
		}

		c.logger.Debug(
			"retrying request",
			"attempt", attempt+1,
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

// shouldRetry reports whether another attempt should be made and, for
// rate-limited responses, how long to wait based on the Retry-After header.
func (c *retryHTTPClient) shouldRetry(resp *http.Response, err error) (bool, time.Duration) {
	// Network/transport errors are retried only when error/status retries are
	// enabled (an explicit MaxRetries == 0 disables them).
	if err != nil {
		return c.retriesEnabled(), 0
	}
	if resp == nil {
		return false, 0
	}

	// Rate-limit handling: honor Retry-After (capped) on 429. This stays
	// active even when error/status retries are disabled.
	if resp.StatusCode == http.StatusTooManyRequests && c.rateLimit != nil && c.rateLimit.Enabled && c.rateLimit.WaitOnLimit {
		return true, c.retryAfter(resp)
	}

	if c.retriesEnabled() && c.isRetriableStatus(resp.StatusCode) {
		return true, 0
	}

	return false, 0
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
