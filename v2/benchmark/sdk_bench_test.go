//go:build benchmark

package benchmark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	fasthttpdoer "github.com/aptos-labs/aptos-go-sdk/v2/fasthttp"
)

// newSDKTestServer matches the paths the v2 nodeClient actually hits:
//   - Info():    GET <NodeURL>          (empty path append)
//   - Account(): GET <NodeURL>/accounts/<addr>
//
// NodeURL is set to srv.URL + "/v1", so the bare /v1 endpoint must return
// NodeInfo and /v1/accounts/* must return AccountInfo.
func newSDKTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/v1/accounts/"):
			_ = json.NewEncoder(w).Encode(sampleAccountInfo)
		case r.URL.Path == "/v1" || r.URL.Path == "/v1/":
			_ = json.NewEncoder(w).Encode(sampleNodeInfo)
		default:
			http.NotFound(w, r)
		}
	}))
}

func sdkNetwork(srvURL string) aptos.NetworkConfig {
	return aptos.NetworkConfig{
		Name:    "bench",
		ChainID: 1,
		NodeURL: srvURL + "/v1",
	}
}

// ============================================================================
// SDK-level benchmarks: full v2 Client.Info() call, including URL building,
// header injection, response parsing, and JSON decoding. This is what users
// actually experience — not just raw HTTP.
// ============================================================================

func BenchmarkSDK_NetHTTP_Info(b *testing.B) {
	srv := newSDKTestServer()
	defer srv.Close()

	client, err := aptos.NewClient(sdkNetwork(srv.URL))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	if _, err := client.Info(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := client.Info(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSDK_FastHTTP_Info(b *testing.B) {
	srv := newSDKTestServer()
	defer srv.Close()

	client, err := aptos.NewClient(
		sdkNetwork(srv.URL),
		aptos.WithHTTPClient(fasthttpdoer.New()),
	)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	if _, err := client.Info(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := client.Info(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSDK_NetHTTP_Account(b *testing.B) {
	srv := newSDKTestServer()
	defer srv.Close()

	client, err := aptos.NewClient(sdkNetwork(srv.URL))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	addr := aptos.AccountOne

	if _, err := client.Account(ctx, addr); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := client.Account(ctx, addr); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSDK_FastHTTP_Account(b *testing.B) {
	srv := newSDKTestServer()
	defer srv.Close()

	client, err := aptos.NewClient(
		sdkNetwork(srv.URL),
		aptos.WithHTTPClient(fasthttpdoer.New()),
	)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	addr := aptos.AccountOne

	if _, err := client.Account(ctx, addr); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := client.Account(ctx, addr); err != nil {
			b.Fatal(err)
		}
	}
}

// Burst tests simulate concurrent SDK callers (e.g., a transaction submitter
// fanning out polling requests). Demonstrates per-call allocation and the
// throughput available behind the HTTPDoer abstraction.

func BenchmarkSDK_NetHTTP_InfoBurst100(b *testing.B) {
	srv := newSDKTestServer()
	defer srv.Close()
	client, err := aptos.NewClient(sdkNetwork(srv.URL))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() { defer wg.Done(); _, _ = client.Info(ctx) }()
		}
		wg.Wait()
	}
}

func BenchmarkSDK_FastHTTP_InfoBurst100(b *testing.B) {
	srv := newSDKTestServer()
	defer srv.Close()
	client, err := aptos.NewClient(
		sdkNetwork(srv.URL),
		aptos.WithHTTPClient(fasthttpdoer.New()),
	)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for j := 0; j < 100; j++ {
			go func() { defer wg.Done(); _, _ = client.Info(ctx) }()
		}
		wg.Wait()
	}
}
