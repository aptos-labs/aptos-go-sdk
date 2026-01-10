# Aptos Go SDK v2 Benchmarks

This package contains comprehensive benchmarks comparing the v2 SDK performance against the v1 SDK.

## Running Benchmarks

### Run All Benchmarks

```bash
go test -bench=. -benchmem ./benchmark/...
```

### Run Specific Category

```bash
# BCS serialization benchmarks
go test -bench=BenchmarkBCS -benchmem ./benchmark/...

# Address parsing and formatting benchmarks
go test -bench=BenchmarkAddress -benchmem ./benchmark/...

# Cryptographic operation benchmarks
go test -bench=BenchmarkCrypto -benchmem ./benchmark/...

# TypeTag parsing benchmarks
go test -bench=BenchmarkTypeTag -benchmem ./benchmark/...

# Iterator and pattern benchmarks
go test -bench=BenchmarkIter -benchmem ./benchmark/...
```

### Generate Comparison Report

For statistically significant results, run with multiple iterations:

```bash
go test -bench=. -benchmem -count=10 ./benchmark/... > benchmark_results.txt
```

Then use `benchstat` to compare:

```bash
benchstat benchmark_results.txt
```

## Benchmark Categories

### BCS Serialization (`bcs_bench_test.go`)

Compares v1 and v2 BCS serialization/deserialization performance for:
- Primitive types (U64, U128, bytes, strings, ULEB128)
- Struct serialization
- Sequence serialization
- Round-trip operations

**Notable Improvements in v2:**
- U64 serialization: ~5x faster, zero allocations
- U64 deserialization: ~10x faster, zero allocations
- Sequence serialization: ~6x faster, zero allocations
- Struct operations: ~30% faster, fewer allocations

### Address Operations (`address_bench_test.go`)

Compares address parsing, formatting, and serialization:
- Short and full address parsing
- String conversion (short, long, special)
- IsSpecial checks
- BCS and JSON serialization

**Performance:** Both v1 and v2 have comparable performance for address operations.

### Cryptographic Operations (`crypto_bench_test.go`)

Compares Ed25519 cryptographic operations:
- Key generation
- Message signing
- Signature verification
- Authentication key derivation
- Key serialization (hex, AIP-80, BCS)

**Notable Improvements in v2:**
- Sign operations: ~25% faster
- Full Sign (with authenticator): fewer allocations
- AuthKey derivation: ~15% faster

### TypeTag Operations (`typetag_bench_test.go`)

Compares TypeTag parsing and serialization:
- Primitive types
- Vector types
- Struct types
- Nested generics
- BCS serialization round-trips

**Performance:** Similar performance with slight improvements in BCS serialization.

### Patterns (`patterns_bench_test.go`)

Compares idiomatic patterns:
- Functional options vs struct-based configuration
- Go 1.23 iterators vs traditional slice ranges
- Iterator operations (Map, Filter, Take, Collect)
- Chained iterator operations

**Observations:**
- Functional options have ~6x overhead vs direct struct modification, but provide better API ergonomics and type safety
- Iterator operations have minimal overhead compared to direct slice iteration
- Chained operations have some overhead due to allocations but maintain lazy evaluation

## Interpreting Results

Benchmark output format:
```
BenchmarkName-N    iterations    ns/op    bytes/op    allocs/op
```

- `N`: Number of CPU cores used
- `iterations`: Number of times the operation was run
- `ns/op`: Nanoseconds per operation (lower is better)
- `bytes/op`: Bytes allocated per operation (lower is better)
- `allocs/op`: Allocations per operation (lower is better)

## Notes

- Benchmarks are run on the same machine to ensure fair comparison
- Results may vary based on hardware and Go version
- The v2 SDK focuses on reducing allocations for common operations
- Some operations trade memory for speed (or vice versa) based on typical use cases

