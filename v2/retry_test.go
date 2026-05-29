package aptos

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestNewRetryHTTPClient_NoConfigReturnsInner(t *testing.T) {
	t.Parallel()
	inner := &defaultHTTPClient{}
	got := newRetryHTTPClient(inner, nil, nil, nil)
	assert.Same(t, inner, got, "no config should return the inner client unchanged")

	got = newRetryHTTPClient(inner, nil, &RateLimitConfig{Enabled: false}, nil)
	assert.Same(t, inner, got, "disabled rate limiting should return the inner client unchanged")
}
