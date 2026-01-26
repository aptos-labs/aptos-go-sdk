package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPDoer is a simple mock for testing.
type mockHTTPDoer struct {
	responses []*http.Response
	errors    []error
	calls     int32
}

func (m *mockHTTPDoer) Do(*http.Request) (*http.Response, error) {
	call := atomic.AddInt32(&m.calls, 1) - 1
	idx := int(call)
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}

	var resp *http.Response
	var err error

	if m.responses != nil && len(m.responses) > idx {
		resp = m.responses[idx]
	}
	if m.errors != nil && len(m.errors) > idx {
		err = m.errors[idx]
	}

	return resp, err
}

func TestRetryClient_Success(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{
			{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))},
		},
	}

	client := NewRetryClient(mock, DefaultRetryConfig(), nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&mock.calls))
}

func TestRetryClient_RetryOnError(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{nil, nil, {StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}},
		errors:    []error{errors.New("network error"), errors.New("network error"), nil},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 1 * time.Millisecond
	client := NewRetryClient(mock, config, nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&mock.calls))
}

func TestRetryClient_RetryOn429(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{
			{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader(""))},
			{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))},
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 1 * time.Millisecond
	client := NewRetryClient(mock, config, nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&mock.calls))
}

func TestRetryClient_NoRetryOnClientError(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{
			{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(""))},
		},
	}

	config := DefaultRetryConfig()
	client := NewRetryClient(mock, config, nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&mock.calls))
}

func TestRetryClient_ExhaustedRetries(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(""))},
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(""))},
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(""))},
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(""))},
		},
	}

	config := DefaultRetryConfig()
	config.MaxRetries = 2
	config.InitialBackoff = 1 * time.Millisecond
	client := NewRetryClient(mock, config, nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)

	// Should return error and last response after exhausting retries
	require.Error(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&mock.calls)) // initial + 2 retries
}

func TestRetryClient_ContextCancellation(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(""))},
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 1 * time.Second // Long backoff
	client := NewRetryClient(mock, config, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRateLimiter_Basic(t *testing.T) {
	limiter := NewRateLimiter(10, 1) // 10 requests/sec, burst of 1

	ctx := context.Background()

	// First request should succeed immediately
	start := time.Now()
	err := limiter.Wait(ctx)
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 50*time.Millisecond)

	// Second request should wait
	start = time.Now()
	err = limiter.Wait(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), 50*time.Millisecond)
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(0.1, 1) // Very slow rate
	limiter.tokens = 0                // Exhaust tokens

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestHeaderClient(t *testing.T) {
	var capturedReq *http.Request
	mock := &mockHTTPDoer{
		responses: []*http.Response{{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}},
	}
	// Wrap mock to capture request
	wrapper := &struct {
		HTTPDoer
	}{mock}

	headers := map[string]string{
		"X-Custom-Header": "custom-value",
		"Authorization":   "Bearer token",
	}
	client := NewHeaderClient(wrapper, headers)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	capturedReq = req // Capture before Do modifies it

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "custom-value", capturedReq.Header.Get("X-Custom-Header"))
	assert.Equal(t, "Bearer token", capturedReq.Header.Get("Authorization"))
}

func TestTimeoutClient(t *testing.T) {
	// Mock that delays
	mock := &mockHTTPDoer{
		responses: []*http.Response{{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}},
	}

	client := NewTimeoutClient(mock, 5*time.Second)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}

func TestLoggingClient(t *testing.T) {
	mock := &mockHTTPDoer{
		responses: []*http.Response{{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}},
	}

	client := NewLoggingClient(mock, nil)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.InitialBackoff)
	assert.Equal(t, 10*time.Second, config.MaxBackoff)
	assert.InDelta(t, 2.0, config.Multiplier, 0.001)
	assert.Contains(t, config.RetryableStatusCodes, http.StatusTooManyRequests)
	assert.Contains(t, config.RetryableStatusCodes, http.StatusServiceUnavailable)
}
