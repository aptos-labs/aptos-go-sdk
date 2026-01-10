package benchmark

import (
	"log/slog"
	"testing"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/iter"
)

// Benchmark functional options pattern

type LegacyConfig struct {
	Timeout    time.Duration
	MaxRetries int
	Headers    map[string]string
	Logger     *slog.Logger
}

func NewLegacyConfig() *LegacyConfig {
	return &LegacyConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		Headers:    make(map[string]string),
		Logger:     slog.Default(),
	}
}

// Legacy configuration via struct modification
func BenchmarkConfig_Legacy_Struct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := NewLegacyConfig()
		cfg.Timeout = 60 * time.Second
		cfg.MaxRetries = 5
		cfg.Headers["X-Custom"] = "value"
		_ = cfg
	}
}

// Functional options pattern
func BenchmarkConfig_V2_FunctionalOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate applying options
		cfg := &aptos.ClientConfig{}
		opts := []aptos.ClientOption{
			aptos.WithTimeout(60 * time.Second),
			aptos.WithRetry(5, 100*time.Millisecond),
			aptos.WithHeader("X-Custom", "value"),
		}
		for _, opt := range opts {
			opt(cfg)
		}
		_ = cfg
	}
}

// Benchmark iterator patterns

func BenchmarkIter_Slice_Range(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range items {
			sum += v
		}
		_ = sum
	}
}

func BenchmarkIter_V2_Seq2(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	it := iter.FromSlice(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for v, err := range it {
			if err != nil {
				b.Fatal(err)
			}
			sum += v
		}
		_ = sum
	}
}

// Benchmark iterator operations

func BenchmarkIter_V2_Map(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := iter.FromSlice(items)
		mapped := iter.Map(it, func(v int) int { return v * 2 })
		sum := 0
		for v, err := range mapped {
			if err != nil {
				b.Fatal(err)
			}
			sum += v
		}
		_ = sum
	}
}

func BenchmarkIter_V2_Filter(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := iter.FromSlice(items)
		filtered := iter.Filter(it, func(v int) bool { return v%2 == 0 })
		sum := 0
		for v, err := range filtered {
			if err != nil {
				b.Fatal(err)
			}
			sum += v
		}
		_ = sum
	}
}

func BenchmarkIter_V2_Take(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := iter.FromSlice(items)
		taken := iter.Take(it, 100)
		sum := 0
		for v, err := range taken {
			if err != nil {
				b.Fatal(err)
			}
			sum += v
		}
		_ = sum
	}
}

// Benchmark collecting iterators

func BenchmarkIter_V2_Collect(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := iter.FromSlice(items)
		result, err := iter.Collect(it)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// Benchmark chained operations

func BenchmarkIter_Slice_ChainedOps(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Manual chained operations
		result := make([]int, 0, 50)
		for _, v := range items {
			if v%2 == 0 { // filter evens
				doubled := v * 2      // map
				if len(result) < 50 { // take 50
					result = append(result, doubled)
				} else {
					break
				}
			}
		}
		_ = result
	}
}

func BenchmarkIter_V2_ChainedOps(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := iter.FromSlice(items)
		filtered := iter.Filter(it, func(v int) bool { return v%2 == 0 })
		mapped := iter.Map(filtered, func(v int) int { return v * 2 })
		taken := iter.Take(mapped, 50)
		result, err := iter.Collect(taken)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}
