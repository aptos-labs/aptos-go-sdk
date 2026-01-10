package telemetry

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// InstrumentationName is the name used to identify this instrumentation.
	InstrumentationName = "github.com/aptos-labs/aptos-go-sdk/v2/telemetry"

	// InstrumentationVersion is the version of this instrumentation.
	InstrumentationVersion = "2.0.0"
)

// Metric names
const (
	MetricRequestDuration   = "aptos.client.request.duration"
	MetricRequestCount      = "aptos.client.request.count"
	MetricRequestErrorCount = "aptos.client.request.error.count"
)

// Attribute keys
const (
	AttrHTTPMethod     = "http.method"
	AttrHTTPURL        = "http.url"
	AttrHTTPStatusCode = "http.status_code"
	AttrAptosOperation = "aptos.operation"
	AttrAptosNetwork   = "aptos.network"
	AttrErrorType      = "error.type"
)

// Config holds configuration for the instrumented HTTP client.
type Config struct {
	// ServiceName identifies your application in traces.
	ServiceName string

	// TracerProvider is the OpenTelemetry tracer provider to use.
	// If nil, uses the global provider.
	TracerProvider trace.TracerProvider

	// MeterProvider is the OpenTelemetry meter provider to use.
	// If nil, uses the global provider.
	MeterProvider metric.MeterProvider

	// Propagator is the context propagator to use for distributed tracing.
	// If nil, uses the global propagator.
	Propagator propagation.TextMapPropagator

	// Transport is the underlying HTTP transport to use.
	// If nil, uses http.DefaultTransport.
	Transport http.RoundTripper

	// DisableTracing disables trace creation.
	DisableTracing bool

	// DisableMetrics disables metric recording.
	DisableMetrics bool
}

// Option configures the instrumented client.
type Option func(*Config)

// WithServiceName sets the service name for traces.
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithTracerProvider sets a custom tracer provider.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Config) {
		c.TracerProvider = tp
	}
}

// WithMeterProvider sets a custom meter provider.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return func(c *Config) {
		c.MeterProvider = mp
	}
}

// WithPropagator sets a custom context propagator.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return func(c *Config) {
		c.Propagator = p
	}
}

// WithTransport sets the underlying HTTP transport.
func WithTransport(t http.RoundTripper) Option {
	return func(c *Config) {
		c.Transport = t
	}
}

// WithoutTracing disables tracing.
func WithoutTracing() Option {
	return func(c *Config) {
		c.DisableTracing = true
	}
}

// WithoutMetrics disables metrics.
func WithoutMetrics() Option {
	return func(c *Config) {
		c.DisableMetrics = true
	}
}

// InstrumentedTransport wraps an HTTP transport with telemetry.
type InstrumentedTransport struct {
	transport  http.RoundTripper
	tracer     trace.Tracer
	meter      metric.Meter
	propagator propagation.TextMapPropagator

	// Metrics
	requestDuration   metric.Float64Histogram
	requestCount      metric.Int64Counter
	requestErrorCount metric.Int64Counter

	config *Config
}

// NewInstrumentedTransport creates a new instrumented HTTP transport.
func NewInstrumentedTransport(opts ...Option) (*InstrumentedTransport, error) {
	cfg := &Config{
		ServiceName: "aptos-go-sdk",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Set defaults
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	if cfg.MeterProvider == nil {
		cfg.MeterProvider = otel.GetMeterProvider()
	}
	if cfg.Propagator == nil {
		cfg.Propagator = otel.GetTextMapPropagator()
	}
	if cfg.Transport == nil {
		cfg.Transport = http.DefaultTransport
	}

	t := &InstrumentedTransport{
		transport:  cfg.Transport,
		tracer:     cfg.TracerProvider.Tracer(InstrumentationName, trace.WithInstrumentationVersion(InstrumentationVersion)),
		meter:      cfg.MeterProvider.Meter(InstrumentationName, metric.WithInstrumentationVersion(InstrumentationVersion)),
		propagator: cfg.Propagator,
		config:     cfg,
	}

	// Initialize metrics
	if !cfg.DisableMetrics {
		var err error
		t.requestDuration, err = t.meter.Float64Histogram(
			MetricRequestDuration,
			metric.WithDescription("Duration of HTTP requests to Aptos nodes"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			return nil, err
		}

		t.requestCount, err = t.meter.Int64Counter(
			MetricRequestCount,
			metric.WithDescription("Total number of HTTP requests to Aptos nodes"),
		)
		if err != nil {
			return nil, err
		}

		t.requestErrorCount, err = t.meter.Int64Counter(
			MetricRequestErrorCount,
			metric.WithDescription("Total number of failed HTTP requests to Aptos nodes"),
		)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

// RoundTrip implements http.RoundTripper.
func (t *InstrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	operation := extractOperation(req.URL.Path)

	// Create span attributes
	attrs := []attribute.KeyValue{
		attribute.String(AttrHTTPMethod, req.Method),
		attribute.String(AttrHTTPURL, sanitizeURL(req.URL.String())),
		attribute.String(AttrAptosOperation, operation),
	}

	// Start span
	var span trace.Span
	if !t.config.DisableTracing {
		ctx, span = t.tracer.Start(ctx, "aptos."+operation,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)
		defer span.End()

		// Inject trace context into request headers
		t.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	}

	// Execute request
	start := time.Now()
	req = req.WithContext(ctx)
	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(start)

	// Record metrics
	if !t.config.DisableMetrics {
		metricAttrs := metric.WithAttributes(
			attribute.String(AttrHTTPMethod, req.Method),
			attribute.String(AttrAptosOperation, operation),
		)

		t.requestCount.Add(ctx, 1, metricAttrs)
		t.requestDuration.Record(ctx, float64(duration.Milliseconds()), metricAttrs)

		if err != nil || (resp != nil && resp.StatusCode >= 400) {
			t.requestErrorCount.Add(ctx, 1, metricAttrs)
		}
	}

	// Update span with response info
	if span != nil && span.IsRecording() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(attribute.String(AttrErrorType, "transport_error"))
		} else if resp != nil {
			span.SetAttributes(attribute.Int(AttrHTTPStatusCode, resp.StatusCode))
			if resp.StatusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
				span.SetAttributes(attribute.String(AttrErrorType, "http_error"))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		}
	}

	return resp, err
}

// NewInstrumentedHTTPClient creates an http.Client with telemetry instrumentation.
func NewInstrumentedHTTPClient(opts ...Option) (*http.Client, error) {
	transport, err := NewInstrumentedTransport(opts...)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}, nil
}

// extractOperation extracts the operation name from the URL path.
func extractOperation(path string) string {
	// Remove version prefix like /v1/
	path = strings.TrimPrefix(path, "/v1/")
	path = strings.TrimPrefix(path, "/")

	// Common Aptos API operations
	switch {
	case path == "" || path == "/":
		return "info"
	case strings.HasPrefix(path, "accounts/") && strings.Contains(path, "/resources"):
		return "account_resources"
	case strings.HasPrefix(path, "accounts/") && strings.Contains(path, "/resource"):
		return "account_resource"
	case strings.HasPrefix(path, "accounts/") && strings.Contains(path, "/modules"):
		return "account_modules"
	case strings.HasPrefix(path, "accounts/"):
		return "account"
	case strings.HasPrefix(path, "transactions/by_hash"):
		return "transaction_by_hash"
	case strings.HasPrefix(path, "transactions/by_version"):
		return "transaction_by_version"
	case strings.HasPrefix(path, "transactions/wait_by_hash"):
		return "wait_transaction"
	case strings.HasPrefix(path, "transactions/batch"):
		return "batch_submit_transaction"
	case strings.HasPrefix(path, "transactions/simulate"):
		return "simulate_transaction"
	case path == "transactions":
		return "submit_transaction"
	case strings.HasPrefix(path, "transactions"):
		return "transactions"
	case strings.HasPrefix(path, "blocks/by_height"):
		return "block_by_height"
	case strings.HasPrefix(path, "blocks/by_version"):
		return "block_by_version"
	case path == "view":
		return "view"
	case path == "estimate_gas_price":
		return "estimate_gas_price"
	case strings.HasPrefix(path, "events"):
		return "events"
	case strings.HasPrefix(path, "spec"):
		return "spec"
	default:
		// Return first path segment
		if idx := strings.Index(path, "/"); idx > 0 {
			return path[:idx]
		}
		if path != "" {
			return path
		}
		return "unknown"
	}
}

// sanitizeURL removes sensitive information from URLs for logging.
func sanitizeURL(urlStr string) string {
	// For now, just return as-is. In future, could strip query params, etc.
	return urlStr
}

// SpanFromContext returns the current span from context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan creates a new context with the given span.
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// StartSpan starts a new span for custom operations.
// The caller is responsible for ending the span with defer span.End().
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(InstrumentationName).Start(ctx, name, opts...) //nolint:spancheck
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}
