package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestExtractOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path     string
		expected string
	}{
		{"/v1/", "info"},
		{"", "info"},
		{"/", "info"},
		{"/v1/accounts/0x1", "account"},
		{"/accounts/0x1", "account"},
		{"/v1/accounts/0x1/resources", "account_resources"},
		{"/v1/accounts/0x1/resource/0x1::coin::CoinStore", "account_resource"},
		{"/v1/accounts/0x1/modules", "account_modules"},
		{"/v1/transactions/by_hash/0x123", "transaction_by_hash"},
		{"/v1/transactions/by_version/123", "transaction_by_version"},
		{"/v1/transactions/wait_by_hash/0x123", "wait_transaction"},
		{"/v1/transactions/batch", "batch_submit_transaction"},
		{"/v1/transactions/simulate", "simulate_transaction"},
		{"/v1/transactions", "submit_transaction"},
		{"/v1/blocks/by_height/100", "block_by_height"},
		{"/v1/blocks/by_version/200", "block_by_version"},
		{"/v1/view", "view"},
		{"/v1/estimate_gas_price", "estimate_gas_price"},
		{"/v1/events/0x1/0x1::account::CoinRegisterEvent/events", "events"},
		{"/v1/spec", "spec"},
		{"/v1/some/unknown/path", "some"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			result := extractOperation(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewInstrumentedTransport(t *testing.T) {
	t.Parallel()

	// Test with defaults
	transport, err := NewInstrumentedTransport()
	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.NotNil(t, transport.tracer)
	assert.NotNil(t, transport.meter)
	assert.NotNil(t, transport.transport)

	// Test with custom service name
	transport, err = NewInstrumentedTransport(WithServiceName("test-service"))
	require.NoError(t, err)
	assert.Equal(t, "test-service", transport.config.ServiceName)

	// Test with tracing disabled
	transport, err = NewInstrumentedTransport(WithoutTracing())
	require.NoError(t, err)
	assert.True(t, transport.config.DisableTracing)

	// Test with metrics disabled
	transport, err = NewInstrumentedTransport(WithoutMetrics())
	require.NoError(t, err)
	assert.True(t, transport.config.DisableMetrics)
}

func TestNewInstrumentedHTTPClient(t *testing.T) {
	t.Parallel()

	client, err := NewInstrumentedHTTPClient()
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)
	assert.Equal(t, int64(60*1e9), client.Timeout.Nanoseconds())
}

func TestRoundTripWithTracing(t *testing.T) {
	t.Parallel()
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"chain_id": 1}`))
	}))
	defer server.Close()

	// Set up test tracer
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	defer func() { _ = tracerProvider.Shutdown(context.Background()) }()

	// Create instrumented transport
	transport, err := NewInstrumentedTransport(
		WithTracerProvider(tracerProvider),
	)
	require.NoError(t, err)

	client := &http.Client{Transport: transport}

	// Make request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/v1/", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Force flush spans
	err = tracerProvider.ForceFlush(context.Background())
	require.NoError(t, err)

	// Check spans were created
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "aptos.info", spans[0].Name())

	// Check span attributes
	attrs := spans[0].Attributes()
	attrMap := make(map[string]attribute.Value)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value
	}
	assert.Equal(t, "GET", attrMap[AttrHTTPMethod].AsString())
	assert.Equal(t, "info", attrMap[AttrAptosOperation].AsString())
	assert.Equal(t, int64(200), attrMap[AttrHTTPStatusCode].AsInt64())
}

func TestRoundTripWithMetrics(t *testing.T) {
	t.Parallel()
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set up test meter
	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = meterProvider.Shutdown(context.Background()) }()

	// Create instrumented transport
	transport, err := NewInstrumentedTransport(
		WithMeterProvider(meterProvider),
		WithoutTracing(),
	)
	require.NoError(t, err)

	client := &http.Client{Transport: transport}

	// Make request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/v1/accounts/0x1", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Collect metrics
	var rm metricdata.ResourceMetrics
	err = reader.Collect(context.Background(), &rm)
	require.NoError(t, err)

	// Verify metrics were recorded
	assert.NotEmpty(t, rm.ScopeMetrics)

	// Find our metrics
	var foundRequestCount, foundRequestDuration bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case MetricRequestCount:
				foundRequestCount = true
			case MetricRequestDuration:
				foundRequestDuration = true
			}
		}
	}
	assert.True(t, foundRequestCount, "request count metric should be recorded")
	assert.True(t, foundRequestDuration, "request duration metric should be recorded")
}

func TestRoundTripWithError(t *testing.T) {
	t.Parallel()
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	// Set up test tracer
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	defer func() { _ = tracerProvider.Shutdown(context.Background()) }()

	// Set up test meter
	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = meterProvider.Shutdown(context.Background()) }()

	// Create instrumented transport
	transport, err := NewInstrumentedTransport(
		WithTracerProvider(tracerProvider),
		WithMeterProvider(meterProvider),
	)
	require.NoError(t, err)

	client := &http.Client{Transport: transport}

	// Make request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/v1/transactions", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Force flush spans
	err = tracerProvider.ForceFlush(context.Background())
	require.NoError(t, err)

	// Check span recorded error status
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1)

	// Check attributes indicate error
	attrs := spans[0].Attributes()
	attrMap := make(map[string]attribute.Value)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value
	}
	assert.Equal(t, int64(500), attrMap[AttrHTTPStatusCode].AsInt64())
	assert.Equal(t, "http_error", attrMap[AttrErrorType].AsString())

	// Check error metrics were recorded
	var rm metricdata.ResourceMetrics
	err = reader.Collect(context.Background(), &rm)
	require.NoError(t, err)

	var foundErrorCount bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == MetricRequestErrorCount {
				foundErrorCount = true
			}
		}
	}
	assert.True(t, foundErrorCount, "error count metric should be recorded")
}

func TestSpanHelpers(t *testing.T) {
	t.Parallel()

	// Set up test tracer
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	otel.SetTracerProvider(tracerProvider)
	defer func() { _ = tracerProvider.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Test StartSpan
	ctx, span := StartSpan(ctx, "test-operation")
	assert.NotNil(t, span)
	assert.True(t, span.IsRecording())

	// Test SpanFromContext
	extractedSpan := SpanFromContext(ctx)
	assert.Equal(t, span, extractedSpan)

	// Test SetSpanAttributes
	SetSpanAttributes(ctx, attribute.String("test.key", "test-value"))

	// Test RecordError
	RecordError(ctx, assert.AnError)

	span.End()

	// Force flush
	err := tracerProvider.ForceFlush(context.Background())
	require.NoError(t, err)

	// Verify span was recorded
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-operation", spans[0].Name())

	// Check attributes
	attrs := spans[0].Attributes()
	found := false
	for _, attr := range attrs {
		if string(attr.Key) == "test.key" {
			found = true
			assert.Equal(t, "test-value", attr.Value.AsString())
		}
	}
	assert.True(t, found, "custom attribute should be set")
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	// Test custom transport
	customTransport := &http.Transport{}
	transport, err := NewInstrumentedTransport(WithTransport(customTransport))
	require.NoError(t, err)
	assert.Equal(t, customTransport, transport.transport)
}

func TestSanitizeURL(t *testing.T) {
	t.Parallel()

	// For now it just returns the URL as-is
	url := "https://fullnode.mainnet.aptoslabs.com/v1/accounts/0x1"
	result := sanitizeURL(url)
	assert.Equal(t, url, result)
}
