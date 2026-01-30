package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

// Sample response data similar to Aptos node responses
var sampleNodeInfo = map[string]interface{}{
	"chain_id":              1,
	"epoch":                 "1000",
	"ledger_version":        "12345678",
	"oldest_ledger_version": "0",
	"ledger_timestamp":      "1700000000000000",
	"node_role":             "full_node",
	"oldest_block_height":   "0",
	"block_height":          "5000000",
	"git_hash":              "abc123def456",
}

var sampleAccountInfo = map[string]interface{}{
	"sequence_number":    "42",
	"authentication_key": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
}

var largeResourceResponse []byte

func init() {
	// Create a larger response to simulate resource queries
	resources := make([]map[string]interface{}, 50)
	for i := range resources {
		resources[i] = map[string]interface{}{
			"type": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
			"data": map[string]interface{}{
				"coin": map[string]interface{}{
					"value": "100000000",
				},
				"deposit_events": map[string]interface{}{
					"counter": "10",
					"guid": map[string]interface{}{
						"id": map[string]interface{}{
							"addr":         "0x1",
							"creation_num": "2",
						},
					},
				},
			},
		}
	}
	largeResourceResponse, _ = json.Marshal(resources)
}

// newTestServer creates a test HTTP server that simulates Aptos node responses
func newTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Node info endpoint (small response)
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleNodeInfo)
	})

	// Account endpoint (medium response)
	mux.HandleFunc("/v1/accounts/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleAccountInfo)
	})

	// Resources endpoint (large response)
	mux.HandleFunc("/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(largeResourceResponse)
	})

	return httptest.NewServer(mux)
}

// DefaultHTTPClient uses default http.Client settings
type DefaultHTTPClient struct {
	client *http.Client
}

func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *DefaultHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.client.Do(req.WithContext(ctx))
}

// TunedHTTPClient uses optimized http.Client settings
type TunedHTTPClient struct {
	client *http.Client
}

func NewTunedHTTPClient() *TunedHTTPClient {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
		DisableKeepAlives:   false,
	}

	return &TunedHTTPClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

func (c *TunedHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.client.Do(req.WithContext(ctx))
}

// FastHTTPClient wraps fasthttp for comparison
type FastHTTPClient struct {
	client *fasthttp.Client
}

func NewFastHTTPClient() *FastHTTPClient {
	return &FastHTTPClient{
		client: &fasthttp.Client{
			MaxConnsPerHost:     100,
			MaxIdleConnDuration: 90 * time.Second,
			ReadTimeout:         30 * time.Second,
			WriteTimeout:        30 * time.Second,
		},
	}
}

// DoFastHTTP performs a request using fasthttp native API (no conversion overhead)
func (c *FastHTTPClient) DoFastHTTP(url string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("Accept", "application/json")

	if err := c.client.Do(req, resp); err != nil {
		return nil, err
	}

	// Copy body since it's released after this function
	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())
	return body, nil
}

// Do implements HTTPDoer interface (with conversion overhead)
func (c *FastHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	frequest := fasthttp.AcquireRequest()
	fresponse := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(frequest)

	// Convert net/http request to fasthttp request
	frequest.SetRequestURI(req.URL.String())
	frequest.Header.SetMethod(req.Method)
	for k, v := range req.Header {
		for _, val := range v {
			frequest.Header.Add(k, val)
		}
	}

	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		frequest.SetBody(body)
	}

	if err := c.client.Do(frequest, fresponse); err != nil {
		fasthttp.ReleaseResponse(fresponse)
		return nil, err
	}

	// Convert fasthttp response to net/http response
	resp := &http.Response{
		StatusCode: fresponse.StatusCode(),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(fresponse.Body())),
		Request:    req,
	}

	fresponse.Header.VisitAll(func(key, value []byte) {
		resp.Header.Add(string(key), string(value))
	})

	// Note: We need to keep fresponse alive until body is read
	// In a real implementation, you'd need a custom ReadCloser
	return resp, nil
}

// ============================================================================
// Single Request Benchmarks (measures per-request latency)
// ============================================================================

func BenchmarkHTTP_Default_SmallResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewDefaultHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkHTTP_Tuned_SmallResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewTunedHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkHTTP_FastHTTP_SmallResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewFastHTTPClient()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.DoFastHTTP(url)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP_FastHTTP_WithConversion_SmallResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewFastHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// ============================================================================
// Large Response Benchmarks
// ============================================================================

func BenchmarkHTTP_Default_LargeResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewDefaultHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/resources"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkHTTP_Tuned_LargeResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewTunedHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/resources"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkHTTP_FastHTTP_LargeResponse(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewFastHTTPClient()
	url := server.URL + "/v1/resources"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.DoFastHTTP(url)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Burst Request Benchmarks (simulates transaction submission bursts)
// ============================================================================

func BenchmarkHTTP_Default_Burst100(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewDefaultHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("GET", url, nil)
				resp, err := client.Do(ctx, req)
				if err != nil {
					return
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkHTTP_Tuned_Burst100(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewTunedHTTPClient()
	ctx := context.Background()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("GET", url, nil)
				resp, err := client.Do(ctx, req)
				if err != nil {
					return
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkHTTP_FastHTTP_Burst100(b *testing.B) {
	server := newTestServer()
	defer server.Close()

	client := NewFastHTTPClient()
	url := server.URL + "/v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() {
				defer wg.Done()
				client.DoFastHTTP(url)
			}()
		}
		wg.Wait()
	}
}
