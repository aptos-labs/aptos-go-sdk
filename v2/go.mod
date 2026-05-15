module github.com/aptos-labs/aptos-go-sdk/v2

go 1.24.2

toolchain go1.25.10

require (
	github.com/aptos-labs/aptos-go-sdk v1.13.0
	github.com/aptos-labs/confidential-asset-bindings/bindings/go v1.1.2
	github.com/cloudflare/circl v1.6.3
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1
	github.com/gtank/ristretto255 v0.1.2
	github.com/hdevalence/ed25519consensus v0.2.0
	github.com/stretchr/testify v1.11.1
	github.com/valyala/fasthttp v1.69.0
	go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/otel/metric v1.43.0
	go.opentelemetry.io/otel/sdk v1.43.0
	go.opentelemetry.io/otel/sdk/metric v1.43.0
	go.opentelemetry.io/otel/trace v1.43.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hasura/go-graphql-client v0.15.1 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Local development: point at sibling confidential-asset-bindings checkout.
// Remove or override when bindings publishes bindings/go on the module proxy.
replace github.com/aptos-labs/confidential-asset-bindings/bindings/go => ../../confidential-asset-bindings/bindings/go
