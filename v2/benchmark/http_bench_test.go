//go:build benchmark

// Package benchmark contains HTTP client benchmarks for the v2 SDK.
//
// This mirrors benchmark/http_bench_test.go (v1) so the two SDKs can be
// compared apples-to-apples. The structure of the cases (Default / Tuned,
// over HTTP/1.1 plain, HTTP/1.1+TLS, and HTTP/2+TLS) matches the v1 file.
//
// Run with: go test -tags=benchmark -bench=. ./benchmark/...
//
// macOS note: the TLS-bearing cases open many short-lived sockets. Running the
// full suite in a single process can exhaust ephemeral ports (sockets enter
// TIME_WAIT for ~60s). If that happens, run the TLS cases in a separate
// invocation, e.g.
//
//	go test -tags=benchmark -bench='^BenchmarkHTTP_'      -run=^$ ./benchmark/...
//	go test -tags=benchmark -bench='^BenchmarkHTTP[12]'   -run=^$ ./benchmark/...
package benchmark

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var sampleNodeInfo = map[string]any{
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

var sampleAccountInfo = map[string]any{
	"sequence_number":    "42",
	"authentication_key": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
}

var largeResourceResponse []byte

func init() {
	resources := make([]map[string]any, 50)
	for i := range resources {
		resources[i] = map[string]any{
			"type": "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>",
			"data": map[string]any{
				"coin": map[string]any{"value": "100000000"},
				"deposit_events": map[string]any{
					"counter": "10",
					"guid": map[string]any{
						"id": map[string]any{
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

func mux() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleNodeInfo)
	})
	m.HandleFunc("/v1/accounts/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleAccountInfo)
	})
	m.HandleFunc("/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(largeResourceResponse)
	})
	return m
}

func newTestServer() *httptest.Server { return httptest.NewServer(mux()) }

// newTestServerH2 returns a TLS httptest server with HTTP/2 enabled via ALPN.
// Requires Go 1.24+ for httptest.Server.EnableHTTP2.
func newTestServerH2() *httptest.Server {
	srv := httptest.NewUnstartedServer(mux())
	srv.EnableHTTP2 = true
	srv.StartTLS()
	return srv
}

func insecureTLS() *tls.Config { return &tls.Config{InsecureSkipVerify: true} }

type netHTTPClient struct{ c *http.Client }

func newDefault() *netHTTPClient {
	return &netHTTPClient{c: &http.Client{Timeout: 30 * time.Second}}
}

func newTuned() *netHTTPClient {
	return &netHTTPClient{c: &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}}
}

func newTunedTLS() *netHTTPClient {
	return &netHTTPClient{c: &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
			TLSClientConfig:     insecureTLS(),
		},
	}}
}

// newTunedH1Only forces HTTP/1.1 even when the TLS server offers HTTP/2.
// Passing a non-nil empty TLSNextProto map disables h2 upgrade.
func newTunedH1Only() *netHTTPClient {
	return &netHTTPClient{c: &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig:     insecureTLS(),
			TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
		},
	}}
}

func (n *netHTTPClient) get(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := n.c.Do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

// ============================================================================
// HTTP/1.1 plain
// ============================================================================

func BenchmarkHTTP_Default_SmallResponse(b *testing.B) {
	srv := newTestServer()
	defer srv.Close()
	c := newDefault()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP_Tuned_SmallResponse(b *testing.B) {
	srv := newTestServer()
	defer srv.Close()
	c := newTuned()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP_Default_LargeResponse(b *testing.B) {
	srv := newTestServer()
	defer srv.Close()
	c := newDefault()
	ctx := context.Background()
	url := srv.URL + "/v1/resources"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP_Tuned_LargeResponse(b *testing.B) {
	srv := newTestServer()
	defer srv.Close()
	c := newTuned()
	ctx := context.Background()
	url := srv.URL + "/v1/resources"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP_Default_Burst100(b *testing.B) {
	srv := newTestServer()
	defer srv.Close()
	c := newDefault()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() {
				defer wg.Done()
				_ = c.get(ctx, url)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkHTTP_Tuned_Burst100(b *testing.B) {
	srv := newTestServer()
	defer srv.Close()
	c := newTuned()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() {
				defer wg.Done()
				_ = c.get(ctx, url)
			}()
		}
		wg.Wait()
	}
}

// ============================================================================
// TLS / HTTP-2
// ============================================================================

func BenchmarkHTTP2_NetHTTP_SmallResponse(b *testing.B) {
	srv := newTestServerH2()
	defer srv.Close()
	c := newTunedTLS()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	_ = c.get(ctx, url) // warm
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP1TLS_NetHTTP_SmallResponse(b *testing.B) {
	srv := newTestServerH2()
	defer srv.Close()
	c := newTunedH1Only()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	_ = c.get(ctx, url)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP2_NetHTTP_LargeResponse(b *testing.B) {
	srv := newTestServerH2()
	defer srv.Close()
	c := newTunedTLS()
	ctx := context.Background()
	url := srv.URL + "/v1/resources"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP1TLS_NetHTTP_LargeResponse(b *testing.B) {
	srv := newTestServerH2()
	defer srv.Close()
	c := newTunedH1Only()
	ctx := context.Background()
	url := srv.URL + "/v1/resources"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.get(ctx, url); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTP2_NetHTTP_Burst100(b *testing.B) {
	srv := newTestServerH2()
	defer srv.Close()
	c := newTunedTLS()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() { defer wg.Done(); _ = c.get(ctx, url) }()
		}
		wg.Wait()
	}
}

func BenchmarkHTTP1TLS_NetHTTP_Burst100(b *testing.B) {
	srv := newTestServerH2()
	defer srv.Close()
	c := newTunedH1Only()
	ctx := context.Background()
	url := srv.URL + "/v1/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() { defer wg.Done(); _ = c.get(ctx, url) }()
		}
		wg.Wait()
	}
}
