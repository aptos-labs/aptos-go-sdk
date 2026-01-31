// Package http provides HTTP client utilities for the Aptos SDK.
//
// This internal package provides composable HTTP client middleware:
//   - Retry with exponential backoff
//   - Rate limiting
//   - Request/response logging
//   - Custom header injection
//   - Timeout management
//
// # HTTPDoer Interface
//
// All middleware implements the HTTPDoer interface:
//
//	type HTTPDoer interface {
//	    Do(req *http.Request) (*http.Response, error)
//	}
//
// This allows composing multiple middleware layers:
//
//	base := &http.Client{}
//	withTimeout := NewTimeoutClient(base, 30*time.Second)
//	withRetry := NewRetryClient(withTimeout, DefaultRetryConfig(), logger)
//	withLogging := NewLoggingClient(withRetry, logger)
//
// # Retry Client
//
// RetryClient provides automatic retry with exponential backoff:
//
//	client := NewRetryClient(inner, RetryConfig{
//	    MaxRetries:     3,
//	    InitialBackoff: 100 * time.Millisecond,
//	    MaxBackoff:     10 * time.Second,
//	    Multiplier:     2.0,
//	    Jitter:         0.1,
//	    RetryableStatusCodes: []int{429, 500, 502, 503, 504},
//	}, logger)
//
// Features:
//   - Respects Retry-After headers
//   - Configurable status codes to retry
//   - Jitter to prevent thundering herd
//   - Context cancellation support
//
// # Rate Limiter
//
// RateLimitedClient implements token bucket rate limiting:
//
//	client := NewRateLimitedClient(inner, 10.0, 20, logger)  // 10 req/s, burst 20
//
// # Header Client
//
// HeaderClient adds custom headers to all requests:
//
//	client := NewHeaderClient(inner, map[string]string{
//	    "Authorization": "Bearer token",
//	    "X-Custom":      "value",
//	})
//
// # Timeout Client
//
// TimeoutClient applies a default timeout to requests without one:
//
//	client := NewTimeoutClient(inner, 30*time.Second)
//
// # Logging Client
//
// LoggingClient logs request/response details using slog:
//
//	client := NewLoggingClient(inner, slog.Default())
package http
