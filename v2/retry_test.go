package aptos

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countingHandler returns a handler that responds with the configured status
// codes in sequence (the last one repeats), tracking how many requests it saw.
func countingHandler(statuses []int, body string) (http.HandlerFunc, *int32) {
	var count int32
	h := func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&count, 1)
		idx := int(n) - 1
		if idx >= len(statuses) {
			idx = len(statuses) - 1
		}
		w.WriteHeader(statuses[idx])
		_, _ = w.Write([]byte(body))
	}
	return h, &count
}

func newRetryTestClient(t *testing.T, handler http.Handler, opts ...ClientOption) *nodeClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	config := &ClientConfig{
		network: NetworkConfig{NodeURL: server.URL, ChainID: 4},
		timeout: 0,
	}
	for _, opt := range opts {
		opt(config)
	}
	client, err := newNodeClient(config)
	require.NoError(t, err)
	return client
}

func TestRetry_NotConfigured_NoRetry(t *testing.T) {
	t.Parallel()
	handler, count := countingHandler([]int{http.StatusServiceUnavailable}, "boom")
	client := newRetryTestClient(t, handler)

	_, err := client.Info(context.Background())
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(count), "no retry should be attempted without WithRetry")
}

func TestRetry_RetriesOnServerError(t *testing.T) {
	t.Parallel()
	// Two 503s then a 200.
	handler, count := countingHandler([]int{
		http.StatusServiceUnavailable,
		http.StatusServiceUnavailable,
		http.StatusOK,
	}, `{"chain_id":4}`)
	client := newRetryTestClient(t, handler, WithRetry(3, time.Millisecond))

	info, err := client.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, int32(3), atomic.LoadInt32(count), "should retry until success")
}

func TestRetry_ExhaustsAndReturnsError(t *testing.T) {
	t.Parallel()
	handler, count := countingHandler([]int{http.StatusInternalServerError}, "nope")
	client := newRetryTestClient(t, handler, WithRetry(2, time.Millisecond))

	_, err := client.Info(context.Background())
	require.Error(t, err)
	// 1 initial + 2 retries = 3 attempts.
	assert.Equal(t, int32(3), atomic.LoadInt32(count))
}

func TestRetry_DoesNotRetryNonRetriableStatus(t *testing.T) {
	t.Parallel()
	handler, count := countingHandler([]int{http.StatusBadRequest}, "bad")
	client := newRetryTestClient(t, handler, WithRetry(3, time.Millisecond))

	_, err := client.Info(context.Background())
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(count), "400 is not retriable")
}

func TestRetry_RetriesPostWithBody(t *testing.T) {
	t.Parallel()
	// View is a POST with a JSON body; ensure the body is re-sent on retry.
	var (
		mu     sync.Mutex
		bodies []string
		count  int32
	)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&count, 1)
		buf, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(buf))
		mu.Unlock()
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`["42"]`))
	})
	client := newRetryTestClient(t, handler, WithRetry(3, time.Millisecond))

	result, err := client.View(context.Background(), &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "coin"},
		Function: "balance",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int32(2), atomic.LoadInt32(&count))
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, bodies, 2)
	assert.Equal(t, bodies[0], bodies[1], "request body must be identical on retry")
	assert.NotEmpty(t, bodies[1])
}

func TestRateLimit_RetriesOn429(t *testing.T) {
	t.Parallel()
	handler, count := countingHandler([]int{
		http.StatusTooManyRequests,
		http.StatusOK,
	}, `{"chain_id":4}`)
	client := newRetryTestClient(t, handler, WithRateLimitHandling(true, time.Second))

	info, err := client.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, int32(2), atomic.LoadInt32(count))
}

func TestRateLimit_HonorsRetryAfterCappedByMaxWait(t *testing.T) {
	t.Parallel()
	var count int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if n < 2 {
			w.Header().Set("Retry-After", "100") // would be 100s uncapped
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chain_id":4}`))
	})
	// Cap the wait at 10ms so the test stays fast; if the cap were ignored
	// this test would hang far beyond its deadline.
	client := newRetryTestClient(t, handler, WithRateLimitHandling(true, 10*time.Millisecond))

	start := time.Now()
	info, err := client.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Less(t, time.Since(start), 5*time.Second, "Retry-After must be capped by MaxWaitTime")
}

func TestRetry_RespectsContextCancellation(t *testing.T) {
	t.Parallel()
	handler, _ := countingHandler([]int{http.StatusServiceUnavailable}, "boom")
	client := newRetryTestClient(t, handler, WithRetry(5, 50*time.Millisecond))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Info(ctx)
	require.Error(t, err)
}

func TestNewRetryHTTPClient_ReturnsInnerWhenInactive(t *testing.T) {
	t.Parallel()
	inner := &defaultHTTPClient{}

	cases := []struct {
		name      string
		retry     *RetryConfig
		rateLimit *RateLimitConfig
	}{
		{"no config", nil, nil},
		{"rate limit disabled", nil, &RateLimitConfig{Enabled: false}},
		{"rate limit enabled but not waiting", nil, &RateLimitConfig{Enabled: true, WaitOnLimit: false}},
		{"retry with zero max retries", &RetryConfig{MaxRetries: 0}, nil},
		{"retry zero and rate limit not waiting", &RetryConfig{MaxRetries: 0}, &RateLimitConfig{Enabled: true, WaitOnLimit: false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := newRetryHTTPClient(inner, tc.retry, tc.rateLimit, nil)
			assert.Same(t, inner, got, "inactive config should return the inner client unchanged")
		})
	}
}

func TestNewRetryHTTPClient_WrapsWhenActive(t *testing.T) {
	t.Parallel()
	inner := &defaultHTTPClient{}

	cases := []struct {
		name      string
		retry     *RetryConfig
		rateLimit *RateLimitConfig
	}{
		{"retries enabled", &RetryConfig{MaxRetries: 1}, nil},
		{"rate-limit waiting enabled", nil, &RateLimitConfig{Enabled: true, WaitOnLimit: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := newRetryHTTPClient(inner, tc.retry, tc.rateLimit, nil)
			assert.NotSame(t, inner, got, "active config should wrap the inner client")
			assert.IsType(t, &retryHTTPClient{}, got)
		})
	}
}

// errDoer is an HTTPDoer that returns a fixed error for the first failN calls,
// then delegates to a success response.
type errDoer struct {
	failN   int
	calls   int32
	mkResp  func() *http.Response
	lastErr error
}

func (d *errDoer) Do(_ context.Context, _ *http.Request) (*http.Response, error) {
	n := atomic.AddInt32(&d.calls, 1)
	if int(n) <= d.failN {
		d.lastErr = errors.New("simulated network error")
		return nil, d.lastErr
	}
	return d.mkResp(), nil
}

func okResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"chain_id":4}`)),
		Header:     make(http.Header),
	}
}

func TestRetry_NetworkErrorIsRetried(t *testing.T) {
	t.Parallel()
	doer := &errDoer{failN: 2, mkResp: okResponse}
	client, err := newNodeClient(&ClientConfig{
		network:     NetworkConfig{NodeURL: "http://example.invalid/v1", ChainID: 4},
		httpClient:  doer,
		retryConfig: &RetryConfig{MaxRetries: 3, InitialBackoff: time.Millisecond, BackoffMultiplier: 2, MaxBackoff: time.Second},
	})
	require.NoError(t, err)

	info, err := client.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, int32(3), atomic.LoadInt32(&doer.calls))
}

func TestRetry_RateLimitOnly_DoesNotRetryNetworkError(t *testing.T) {
	t.Parallel()
	// Rate-limit handling without a retry config must not retry plain network
	// errors (only 429 responses).
	doer := &errDoer{failN: 5, mkResp: okResponse}
	client, err := newNodeClient(&ClientConfig{
		network:         NetworkConfig{NodeURL: "http://example.invalid/v1", ChainID: 4},
		httpClient:      doer,
		rateLimitConfig: &RateLimitConfig{Enabled: true, WaitOnLimit: true, MaxWaitTime: time.Second},
	})
	require.NoError(t, err)

	_, err = client.Info(context.Background())
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&doer.calls))
}

func TestRetryHTTPClient_MaxAttempts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		client   *retryHTTPClient
		expected int
	}{
		{"retry with max", &retryHTTPClient{retry: &RetryConfig{MaxRetries: 3}}, 4},
		{"rate-limit only", &retryHTTPClient{rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: true}}, 6},
		{"rate-limit enabled but not waiting", &retryHTTPClient{rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: false}}, 1},
		{"neither", &retryHTTPClient{}, 1},
		{"retry zero", &retryHTTPClient{retry: &RetryConfig{MaxRetries: 0}}, 1},
		{
			// 1 initial + 0 error/status + 5 rate-limit.
			"retry zero with rate-limit",
			&retryHTTPClient{retry: &RetryConfig{MaxRetries: 0}, rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: true}},
			6,
		},
		{
			// Error/status and rate-limit budgets are independent: the total
			// must reserve room for both (1 + 10 + 5), not max(11, 6).
			"retry plus rate-limit budgets are additive",
			&retryHTTPClient{retry: &RetryConfig{MaxRetries: 10}, rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: true}},
			16,
		},
		{
			// 1 initial + 2 error/status + 5 rate-limit.
			"small retry with rate-limit",
			&retryHTTPClient{retry: &RetryConfig{MaxRetries: 2}, rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: true}},
			8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.client.maxAttempts())
		})
	}
}

func TestRetryHTTPClient_RetryAfter(t *testing.T) {
	t.Parallel()
	c := &retryHTTPClient{rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: true, MaxWaitTime: 5 * time.Second}}

	cases := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{"absent header", "", 0},
		{"unparseable header", "not-a-number", 0},
		{"negative header", "-3", 0},
		{"valid header", "2", 2 * time.Second},
		{"capped at MaxWaitTime", "100", 5 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := make(http.Header)
			if tc.header != "" {
				h.Set("Retry-After", tc.header)
			}
			resp := &http.Response{Header: h, Body: http.NoBody}
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tc.want, c.retryAfter(resp))
		})
	}
}

func TestRetryHTTPClient_ClassifyRetry(t *testing.T) {
	t.Parallel()

	// Network error: classified as error/status only when retries are enabled.
	withRetry := &retryHTTPClient{retry: DefaultRetryConfig()}
	reason, _ := withRetry.classifyRetry(nil, errors.New("boom"))
	assert.Equal(t, retryErrorStatus, reason)

	rateOnly := &retryHTTPClient{rateLimit: &RateLimitConfig{Enabled: true, WaitOnLimit: true}}
	reason, _ = rateOnly.classifyRetry(nil, errors.New("boom"))
	assert.Equal(t, retryNone, reason, "rate-limit-only must not retry network errors")

	// nil response, no error.
	reason, _ = withRetry.classifyRetry(nil, nil)
	assert.Equal(t, retryNone, reason)

	// 429 with rate-limit handling.
	resp429 := &http.Response{StatusCode: http.StatusTooManyRequests, Header: make(http.Header)}
	reason, _ = rateOnly.classifyRetry(resp429, nil)
	assert.Equal(t, retryRateLimit, reason)

	// Retryable 5xx with retries enabled.
	resp503 := &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header)}
	reason, _ = withRetry.classifyRetry(resp503, nil)
	assert.Equal(t, retryErrorStatus, reason)

	// Non-retriable success.
	resp200 := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	reason, _ = withRetry.classifyRetry(resp200, nil)
	assert.Equal(t, retryNone, reason)
}

func TestRetryHTTPClient_Backoff(t *testing.T) {
	t.Parallel()

	// Rate-limit-only mode (no retry config) uses a small fixed base.
	rateOnly := &retryHTTPClient{rateLimit: &RateLimitConfig{Enabled: true}}
	assert.Equal(t, 100*time.Millisecond, rateOnly.backoff(1))

	// Zero/negative config values fall back to defaults without panicking.
	defaulted := &retryHTTPClient{retry: &RetryConfig{InitialBackoff: 0, BackoffMultiplier: 0}}
	got := defaulted.backoff(1)
	assert.Positive(t, got)

	// Growth is capped at MaxBackoff.
	capped := &retryHTTPClient{retry: &RetryConfig{
		InitialBackoff:    time.Second,
		BackoffMultiplier: 10,
		MaxBackoff:        2 * time.Second,
	}}
	assert.LessOrEqual(t, capped.backoff(5), 2*time.Second)
}

func TestRetry_RateLimitDoesNotRaiseServerErrorCap(t *testing.T) {
	t.Parallel()
	// Server returns 503 forever. With MaxRetries=2, enabling rate-limit
	// handling must NOT allow more than 2 retries for the 503s (regression:
	// previously the budget was bumped to the rate-limit default of ~6).
	handler, count := countingHandler([]int{http.StatusServiceUnavailable}, "down")
	client := newRetryTestClient(
		t, handler,
		WithRetry(2, time.Millisecond),
		WithRateLimitHandling(true, time.Second),
	)

	_, err := client.Info(context.Background())
	require.Error(t, err)
	// 1 initial + 2 retries == 3, strictly capped by MaxRetries.
	assert.Equal(t, int32(3), atomic.LoadInt32(count))
}

func TestRetry_ServerErrorAndRateLimitCapsAreIndependent(t *testing.T) {
	t.Parallel()
	// 503 (retryable status) then 429 (rate limit) then 200. With MaxRetries=1
	// the 503 consumes the single error/status retry, and the 429 is handled
	// by the independent rate-limit budget, so the request still succeeds.
	handler, count := countingHandler([]int{
		http.StatusServiceUnavailable,
		http.StatusTooManyRequests,
		http.StatusOK,
	}, `{"chain_id":4}`)
	client := newRetryTestClient(
		t, handler,
		WithRetry(1, time.Millisecond),
		WithRateLimitHandling(true, time.Second),
	)

	info, err := client.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, int32(3), atomic.LoadInt32(count))
}

func TestRetry_MaxRetriesZeroWithRateLimit_OnlyRetries429(t *testing.T) {
	t.Parallel()

	// An explicit MaxRetries == 0 disables error/status retries, but enabling
	// rate-limit handling must still retry 429s.
	zeroRetry := func() *RetryConfig {
		return &RetryConfig{MaxRetries: 0, RetryableStatusCodes: []int{http.StatusServiceUnavailable}}
	}

	// A retryable status (503) must NOT be retried when MaxRetries == 0.
	t.Run("503 not retried", func(t *testing.T) {
		t.Parallel()
		handler, count := countingHandler([]int{
			http.StatusServiceUnavailable,
			http.StatusOK,
		}, `{"chain_id":4}`)
		client := newRetryTestClient(
			t, handler,
			WithRetryConfig(zeroRetry()),
			WithRateLimitHandling(true, time.Second),
		)
		_, err := client.Info(context.Background())
		require.Error(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(count), "503 must not be retried when MaxRetries==0")
	})

	// A 429 must still be retried via rate-limit handling.
	t.Run("429 retried", func(t *testing.T) {
		t.Parallel()
		handler, count := countingHandler([]int{
			http.StatusTooManyRequests,
			http.StatusOK,
		}, `{"chain_id":4}`)
		client := newRetryTestClient(
			t, handler,
			WithRetryConfig(zeroRetry()),
			WithRateLimitHandling(true, time.Second),
		)
		info, err := client.Info(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint8(4), info.ChainID)
		assert.Equal(t, int32(2), atomic.LoadInt32(count), "429 must be retried under rate-limit handling")
	})
}

func TestRetryHTTPClient_ClassifyRetry_MaxRetriesZero(t *testing.T) {
	t.Parallel()

	// retry present but MaxRetries == 0: network errors and retryable statuses
	// must NOT be retried.
	c := &retryHTTPClient{retry: &RetryConfig{MaxRetries: 0, RetryableStatusCodes: []int{http.StatusServiceUnavailable}}}

	reason, _ := c.classifyRetry(nil, errors.New("boom"))
	assert.Equal(t, retryNone, reason, "network error must not be retried when MaxRetries==0")

	resp503 := &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header)}
	reason, _ = c.classifyRetry(resp503, nil)
	assert.Equal(t, retryNone, reason, "503 must not be retried when MaxRetries==0")

	// With rate-limit handling, a 429 is still retried.
	c.rateLimit = &RateLimitConfig{Enabled: true, WaitOnLimit: true}
	resp429 := &http.Response{StatusCode: http.StatusTooManyRequests, Header: make(http.Header)}
	reason, _ = c.classifyRetry(resp429, nil)
	assert.Equal(t, retryRateLimit, reason, "429 must be retried under rate-limit handling even with MaxRetries==0")
}

func TestRetry_WithRetryConfigOption(t *testing.T) {
	t.Parallel()
	handler, count := countingHandler([]int{
		http.StatusBadGateway,
		http.StatusOK,
	}, `{"chain_id":4}`)
	cfg := &RetryConfig{
		MaxRetries:           2,
		InitialBackoff:       time.Millisecond,
		MaxBackoff:           time.Second,
		BackoffMultiplier:    2,
		RetryableStatusCodes: []int{http.StatusBadGateway},
	}
	client := newRetryTestClient(t, handler, WithRetryConfig(cfg))

	info, err := client.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(4), info.ChainID)
	assert.Equal(t, int32(2), atomic.LoadInt32(count))
}
