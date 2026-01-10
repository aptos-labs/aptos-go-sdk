// Package telemetry provides OpenTelemetry instrumentation for the Aptos Go SDK.
//
// This package allows you to add distributed tracing and metrics to your Aptos SDK usage.
// It wraps the HTTP client to automatically create spans and record metrics for all API calls.
//
// # Basic Usage
//
// To enable telemetry, wrap your HTTP client with the instrumented client:
//
//	import (
//		aptos "github.com/aptos-labs/aptos-go-sdk/v2"
//		"github.com/aptos-labs/aptos-go-sdk/v2/telemetry"
//	)
//
//	// Create instrumented HTTP client
//	httpClient := telemetry.NewInstrumentedHTTPClient(
//		telemetry.WithServiceName("my-app"),
//	)
//
//	// Use with Aptos client
//	client, _ := aptos.NewClient(aptos.MainnetConfig,
//		aptos.WithHTTPClient(httpClient),
//	)
//
// # Traces
//
// The package creates spans for each API call with the following attributes:
//   - http.method: GET, POST, etc.
//   - http.url: The full URL of the request
//   - http.status_code: Response status code
//   - aptos.operation: The type of operation (info, account, submit_transaction, etc.)
//   - error: Any error that occurred
//
// # Metrics
//
// The package records the following metrics:
//   - aptos.client.request.duration: Histogram of request durations
//   - aptos.client.request.count: Counter of total requests
//   - aptos.client.request.error.count: Counter of failed requests
//
// # Configuration
//
// You must set up your OpenTelemetry provider before creating the instrumented client.
// The package uses the global tracer and meter providers by default.
//
// Example setup:
//
//	import (
//		"go.opentelemetry.io/otel"
//		"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
//		"go.opentelemetry.io/otel/sdk/trace"
//	)
//
//	// Set up exporter
//	exporter, _ := otlptracehttp.New(ctx)
//	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
//	otel.SetTracerProvider(tp)
//
// # Environment Variables
//
// The package respects standard OpenTelemetry environment variables:
//   - OTEL_SERVICE_NAME: Service name for traces
//   - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP endpoint URL
//   - OTEL_TRACES_SAMPLER: Sampling strategy
package telemetry
