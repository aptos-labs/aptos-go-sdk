// Package fasthttp provides an optional fasthttp-backed [HTTPDoer] for the
// Aptos Go SDK v2. It is a separate module so the fasthttp dependency is never
// pulled in by users who don't opt in.
//
// To use it, add the module to your project:
//
//	go get github.com/aptos-labs/aptos-go-sdk/v2/fasthttp
//
// Then wire it into the SDK client:
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
//
// # Why fasthttp over net/http
//
// fasthttp uses a connection-pooled HTTP/1.1 implementation that avoids most
// per-request allocation overhead in net/http. Benchmarks against an
// Aptos-shaped JSON endpoint show roughly half the allocations and ~30% lower
// wall time per request versus the stdlib default client.
//
// fasthttp does not implement HTTP/2. For typical SDK workloads (sequential
// REST calls against a fullnode) this is fine: Go's HTTP/2 stack carries
// measurable per-request overhead for the small sequential requests the SDK
// issues, and Aptos fullnodes serve HTTP/1.1 over TLS in practice.
package fasthttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

type client interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
	DoDeadline(req *fasthttp.Request, resp *fasthttp.Response, deadline time.Time) error
}

// Doer is an [aptos.HTTPDoer] backed by fasthttp.
//
// The zero value is not usable; construct one with [New].
type Doer struct {
	client client
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
//   - When ctx has a deadline, [fasthttp.Client.DoDeadline] enforces it while
//     the call is also raced against ctx.Done to support earlier cancellation.
//   - When ctx is cancellable without a deadline, [fasthttp.Client.Do] is run
//     on a goroutine and raced against ctx.Done.
//   - When ctx is not cancellable (e.g. [context.Background]), the call is
//     synchronous and allocates no extra goroutine. This is the common SDK
//     case and keeps overhead minimal under high concurrency.
func (d *Doer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	freq := fasthttp.AcquireRequest()
	releaseRequest := true
	defer func() {
		if releaseRequest {
			fasthttp.ReleaseRequest(freq)
		}
	}()

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

	cleanupDone, err := d.executeWithContext(ctx, freq, fresp)
	if cleanupDone != nil {
		// The request is still in flight. The drainer goroutine now owns both
		// fasthttp objects and will release them after the client call returns.
		releaseRequest = false
		return nil, err
	}
	if err != nil {
		fasthttp.ReleaseResponse(fresp)
		return nil, err
	}

	// Hold ownership of fresp until the caller closes the Response.Body so we
	// can hand out fresp's internal []byte directly instead of copying it.
	body := fresp.Body()
	statusCode := fresp.StatusCode()

	resp := &http.Response{
		Status:        strconv.Itoa(statusCode) + " " + string(fresp.Header.StatusMessage()),
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          &fasthttpBody{Reader: bytes.NewReader(body), resp: fresp},
		ContentLength: int64(len(body)),
		Request:       req,
	}
	for k, v := range fresp.Header.All() {
		resp.Header.Add(string(k), string(v))
	}

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
// possible. A non-nil channel means asynchronous cleanup owns freq and fresp;
// the channel closes after both are released. See [Doer.Do] for the cases.
func (d *Doer) executeWithContext(
	ctx context.Context,
	freq *fasthttp.Request,
	fresp *fasthttp.Response,
) (<-chan struct{}, error) {
	// Fast path: non-cancellable context (e.g. context.Background). No
	// goroutine, no channel — direct synchronous call.
	if ctx.Done() == nil {
		return nil, d.client.Do(freq, fresp)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Deadline-bound contexts use fasthttp's native deadline while still
	// racing ctx.Done so an explicit cancel can return before that deadline.
	if deadline, ok := ctx.Deadline(); ok {
		errCh := make(chan error, 1)
		go func() { errCh <- d.client.DoDeadline(freq, fresp, deadline) }()

		select {
		case err := <-errCh:
			return nil, deadlineError(ctx, deadline, err)
		case <-ctx.Done():
			// Prefer a completed client call so the caller can clean up
			// synchronously when both events become ready together.
			select {
			case err := <-errCh:
				return nil, deadlineError(ctx, deadline, err)
			default:
			}
			return cleanupAfter(errCh, freq, fresp), ctx.Err()
		}
	}

	// Cancellable but no deadline: race the call against ctx.Done. On
	// cancellation, hand both pooled objects to a drainer goroutine so they can
	// be released when fasthttp finally returns. fasthttp's ReadTimeout /
	// WriteTimeout cap how long that drainer can block.
	errCh := make(chan error, 1)
	go func() { errCh <- d.client.Do(freq, fresp) }()

	select {
	case <-ctx.Done():
		select {
		case err := <-errCh:
			return nil, contextResult(ctx, err)
		default:
		}
		return cleanupAfter(errCh, freq, fresp), ctx.Err()
	case err := <-errCh:
		return nil, contextResult(ctx, err)
	}
}

func contextResult(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}

func deadlineError(ctx context.Context, deadline time.Time, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	// DoDeadline and the context use independent timers, so the client may
	// report its timeout just before the context timer publishes
	// DeadlineExceeded.
	if err != nil && !time.Now().Before(deadline) {
		return context.DeadlineExceeded
	}
	return err
}

func cleanupAfter(
	errCh <-chan error,
	freq *fasthttp.Request,
	fresp *fasthttp.Response,
) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-errCh
		fasthttp.ReleaseResponse(fresp)
		fasthttp.ReleaseRequest(freq)
	}()
	return done
}
