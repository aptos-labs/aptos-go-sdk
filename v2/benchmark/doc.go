// Package benchmark provides performance benchmarks comparing v2 SDK operations
// against v1 SDK implementations.
//
// To run all benchmarks:
//
//	go test -bench=. -benchmem ./benchmark/...
//
// To run specific benchmarks:
//
//	go test -bench=BenchmarkBCS -benchmem ./benchmark/...
//	go test -bench=BenchmarkAddress -benchmem ./benchmark/...
//	go test -bench=BenchmarkCrypto -benchmem ./benchmark/...
//
// To generate comparison reports:
//
//	go test -bench=. -benchmem -count=10 ./benchmark/... > benchmark_results.txt
//
// The benchmarks cover:
//   - BCS serialization/deserialization performance
//   - Address parsing and formatting
//   - Ed25519 key generation and signing
//   - TypeTag parsing
//   - HTTP client operations (mocked)
package benchmark
