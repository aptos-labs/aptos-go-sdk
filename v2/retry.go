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
// when neither retry nor rate-limit handling is configured so the common
// path adds no overhead.
func newRetryHTTPClient(inner HTTPDoer, retry *RetryConfig, rateLimit *RateLimitConfig, logger *slog.Logger) HTTPDoer {
	if retry == nil && (rateLimit == nil || !rateLimit.Enabled) {
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
func (c *retryHTTPClient) maxAttempts() int {
	if c.retry != nil && c.retry.MaxRetries > 0 {
		return c.retry.MaxRetries + 1
	}
	// Rate-limit handling without an explicit retry config still needs to
	// retry 429s. Use a sensible bounded default so a misbehaving server
	// can't make us loop forever; overall waiting is additionally bounded
	// by MaxWaitTime per attempt and by the request context.
	if c.rateLimit != nil && c.rateLimit.Enabled {
		return 6
	}
	return 1
}

// Do executes the request, retrying transient failures per the configuration.
func (c *retryHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	attempts := c.maxAttempts()

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

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	return resp, err
}

// shouldRetry reports whether another attempt should be made and, for
// rate-limited responses, how long to wait based on the Retry-After header.
func (c *retryHTTPClient) shouldRetry(resp *http.Response, err error) (bool, time.Duration) {
	// Network/transport errors are retried only when retry is configured.
	if err != nil {
		return c.retry != nil, 0
	}
	if resp == nil {
		return false, 0
	}

	// Rate-limit handling: honor Retry-After (capped) on 429.
	if resp.StatusCode == http.StatusTooManyRequests && c.rateLimit != nil && c.rateLimit.Enabled && c.rateLimit.WaitOnLimit {
		return true, c.retryAfter(resp)
	}

	if c.retry != nil && c.isRetriableStatus(resp.StatusCode) {
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
