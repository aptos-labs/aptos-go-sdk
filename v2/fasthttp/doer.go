// Package fasthttp provides a fasthttp-backed [aptos.HTTPDoer] for the v2 SDK.
//
// fasthttp uses a connection-pooled HTTP/1.1 implementation that avoids most of
// the per-request allocation overhead in net/http. Benchmarks against an Aptos-
// shaped JSON endpoint show roughly half the allocations and ~30% lower wall
// time per request versus the stdlib client. See v2/benchmark for numbers.
//
// fasthttp does not implement HTTP/2; HTTP/2 endpoints are reached over
// HTTP/1.1 via ALPN negotiation. For typical SDK workloads (sequential REST
// calls against a fullnode) this is preferable, because Go's HTTP/2 stack
// carries notable per-request overhead.
//
// Usage:
//
//	import (
//	    aptos "github.com/aptos-labs/aptos-go-sdk/v2"
//	    fasthttpdoer "github.com/aptos-labs/aptos-go-sdk/v2/fasthttp"
//	)
//
//	client, err := aptos.NewClient(
//	    aptos.WithNetwork(aptos.DevnetConfig),
//	    aptos.WithHTTPClient(fasthttpdoer.New()),
//	)
package fasthttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

// errCancelled is a sentinel for "context cancelled mid-flight" so the caller
// path can distinguish it from a network error and avoid double-releasing the
// fasthttp response (a drainer goroutine takes ownership).
var errCancelled = errors.New("fasthttpdoer: cancelled")

// Doer is an [aptos.HTTPDoer] backed by fasthttp.
//
// The zero value is not usable; construct one with [New].
type Doer struct {
	client *fasthttp.Client
}

// Option configures a [Doer].
type Option func(*config)

type config struct {
	maxConnsPerHost     int
	maxIdleConnDuration time.Duration
	readTimeout         time.Duration
	writeTimeout        time.Duration
	tlsConfig           *tls.Config
}

func defaultConfig() config {
	return config{
		maxConnsPerHost:     512,
		maxIdleConnDuration: 90 * time.Second,
		readTimeout:         30 * time.Second,
		writeTimeout:        30 * time.Second,
	}
}

// WithMaxConnsPerHost caps the number of concurrent connections to a single host.
func WithMaxConnsPerHost(n int) Option {
	return func(c *config) { c.maxConnsPerHost = n }
}

// WithIdleConnDuration sets how long idle keep-alive connections are retained.
func WithIdleConnDuration(d time.Duration) Option {
	return func(c *config) { c.maxIdleConnDuration = d }
}

// WithReadTimeout sets the per-request read timeout enforced by fasthttp.
// This is independent of any context deadline.
func WithReadTimeout(d time.Duration) Option {
	return func(c *config) { c.readTimeout = d }
}

// WithWriteTimeout sets the per-request write timeout enforced by fasthttp.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *config) { c.writeTimeout = d }
}

// WithTLSConfig sets the [tls.Config] used for HTTPS endpoints.
func WithTLSConfig(t *tls.Config) Option {
	return func(c *config) { c.tlsConfig = t }
}

// New returns a fasthttp-backed [aptos.HTTPDoer].
func New(opts ...Option) *Doer {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Doer{
		client: &fasthttp.Client{
			MaxConnsPerHost:     cfg.maxConnsPerHost,
			MaxIdleConnDuration: cfg.maxIdleConnDuration,
			ReadTimeout:         cfg.readTimeout,
			WriteTimeout:        cfg.writeTimeout,
			TLSConfig:           cfg.tlsConfig,
		},
	}
}

// Do executes req using the underlying fasthttp client. The returned
// [http.Response] holds a buffered body so the caller can read and Close it as
// with any net/http response.
//
// fasthttp does not natively accept a [context.Context]:
//   - When ctx has a deadline, [fasthttp.Client.DoDeadline] enforces it.
//   - When ctx is cancellable but has no deadline, the request is run on a
//     goroutine and raced against ctx.Done.
//   - When ctx is not cancellable (e.g. [context.Background]), the call is
//     synchronous and allocates no extra goroutine. This is the common SDK
//     case and keeps overhead minimal under high concurrency.
func (d *Doer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)

	freq.SetRequestURI(req.URL.String())
	freq.Header.SetMethod(req.Method)
	for k, vs := range req.Header {
		for _, v := range vs {
			freq.Header.Add(k, v)
		}
	}
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
		freq.SetBody(body)
	}

	fresp := fasthttp.AcquireResponse()

	err := d.executeWithContext(ctx, freq, fresp)
	if err != nil {
		// On error including ctx cancellation, fresp is released here unless
		// it's owned by a drainer goroutine (see cancellation path).
		if err != errCancelled {
			fasthttp.ReleaseResponse(fresp)
		}
		if err == errCancelled {
			return nil, ctx.Err()
		}
		return nil, err
	}

	// Hold ownership of fresp until the caller closes the Response.Body so we
	// can hand out fresp's internal []byte directly instead of copying it.
	body := fresp.Body()

	resp := &http.Response{
		Status:        string(fresp.Header.StatusMessage()),
		StatusCode:    fresp.StatusCode(),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          &fasthttpBody{Reader: bytes.NewReader(body), resp: fresp},
		ContentLength: int64(len(body)),
		Request:       req,
	}
	fresp.Header.VisitAll(func(k, v []byte) {
		resp.Header.Add(string(k), string(v))
	})

	return resp, nil
}

// fasthttpBody adapts a fasthttp response body so it satisfies io.ReadCloser
// while keeping the underlying *fasthttp.Response alive until Close is called.
// On Close, the fasthttp response is returned to its pool.
type fasthttpBody struct {
	*bytes.Reader
	resp *fasthttp.Response
	once bool
}

func (b *fasthttpBody) Close() error {
	if b.once {
		return nil
	}
	b.once = true
	if b.resp != nil {
		fasthttp.ReleaseResponse(b.resp)
		b.resp = nil
	}
	return nil
}

// executeWithContext runs the fasthttp request honoring ctx as cheaply as
// possible. See [Doer.Do] for the three cases.
func (d *Doer) executeWithContext(ctx context.Context, freq *fasthttp.Request, fresp *fasthttp.Response) error {
	// Fast path: non-cancellable context (e.g. context.Background). No
	// goroutine, no channel — direct synchronous call.
	if ctx.Done() == nil {
		return d.client.Do(freq, fresp)
	}

	// Fast path: deadline-bound context. fasthttp has a native API for this.
	if deadline, ok := ctx.Deadline(); ok {
		err := d.client.DoDeadline(freq, fresp, deadline)
		// If the deadline elapsed, surface ctx.Err so callers can distinguish
		// cancellation from network timeouts via errors.Is(err, context.DeadlineExceeded).
		if err != nil && ctx.Err() != nil {
			return errCancelled
		}
		return err
	}

	// Cancellable but no deadline: race the call against ctx.Done. On
	// cancellation, hand the response slot to a drainer goroutine so it can be
	// released when fasthttp finally returns. fasthttp's ReadTimeout /
	// WriteTimeout cap how long that drainer can block.
	errCh := make(chan error, 1)
	go func() { errCh <- d.client.Do(freq, fresp) }()

	select {
	case <-ctx.Done():
		go func() {
			<-errCh
			fasthttp.ReleaseResponse(fresp)
		}()
		return errCancelled
	case err := <-errCh:
		return err
	}
}
