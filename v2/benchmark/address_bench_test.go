package benchmark

import (
	"testing"

	bcsv1 "github.com/aptos-labs/aptos-go-sdk/bcs"
	typesv1 "github.com/aptos-labs/aptos-go-sdk/internal/types"
	bcsv2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	typesv2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
)

// Benchmark address parsing

func BenchmarkAddress_V1_ParseShort(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.ParseStringRelaxed("0x1")
	}
}

func BenchmarkAddress_V2_ParseShort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseAddress("0x1")
	}
}

func BenchmarkAddress_V1_ParseFull(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.ParseStringRelaxed("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	}
}

func BenchmarkAddress_V2_ParseFull(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	}
}

func BenchmarkAddress_V1_ParseNoPrefix(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.ParseStringRelaxed("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	}
}

func BenchmarkAddress_V2_ParseNoPrefix(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseAddress("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	}
}

// Benchmark address to string conversion

func BenchmarkAddress_V1_StringSpecial(b *testing.B) {
	addr := typesv1.AccountOne
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.String()
	}
}

func BenchmarkAddress_V2_StringSpecial(b *testing.B) {
	addr := typesv2.MustParseAddress("0x1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.String()
	}
}

func BenchmarkAddress_V1_StringFull(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	_ = addr.ParseStringRelaxed("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.String()
	}
}

func BenchmarkAddress_V2_StringFull(b *testing.B) {
	addr, _ := typesv2.ParseAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.String()
	}
}

func BenchmarkAddress_V1_StringLong(b *testing.B) {
	addr := typesv1.AccountOne
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.StringLong()
	}
}

func BenchmarkAddress_V2_StringLong(b *testing.B) {
	addr := typesv2.MustParseAddress("0x1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.StringLong()
	}
}

// Benchmark IsSpecial check

func BenchmarkAddress_V1_IsSpecial_True(b *testing.B) {
	addr := typesv1.AccountOne
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.IsSpecial()
	}
}

func BenchmarkAddress_V2_IsSpecial_True(b *testing.B) {
	addr := typesv2.MustParseAddress("0x1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.IsSpecial()
	}
}

func BenchmarkAddress_V1_IsSpecial_False(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	_ = addr.ParseStringRelaxed("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.IsSpecial()
	}
}

func BenchmarkAddress_V2_IsSpecial_False(b *testing.B) {
	addr, _ := typesv2.ParseAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.IsSpecial()
	}
}

// Benchmark BCS serialization of addresses

func BenchmarkAddress_V1_BCS_Serialize(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	_ = addr.ParseStringRelaxed("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv1.Serialize(addr)
	}
}

func BenchmarkAddress_V2_BCS_Serialize(b *testing.B) {
	addr, _ := typesv2.ParseAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv2.Serialize(&addr)
	}
}

func BenchmarkAddress_V1_BCS_Deserialize(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	_ = addr.ParseStringRelaxed("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	data, _ := bcsv1.Serialize(addr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &typesv1.AccountAddress{}
		des := bcsv1.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkAddress_V2_BCS_Deserialize(b *testing.B) {
	addr, _ := typesv2.ParseAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	data, _ := bcsv2.Serialize(&addr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := typesv2.AccountAddress{}
		des := bcsv2.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

// Benchmark JSON marshaling

func BenchmarkAddress_V1_JSON_Marshal(b *testing.B) {
	addr := &typesv1.AccountAddress{}
	_ = addr.ParseStringRelaxed("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = addr.MarshalJSON()
	}
}

func BenchmarkAddress_V2_JSON_Marshal(b *testing.B) {
	addr, _ := typesv2.ParseAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = addr.MarshalJSON()
	}
}

func BenchmarkAddress_V1_JSON_Unmarshal(b *testing.B) {
	data := []byte(`"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := &typesv1.AccountAddress{}
		_ = addr.UnmarshalJSON(data)
	}
}

func BenchmarkAddress_V2_JSON_Unmarshal(b *testing.B) {
	data := []byte(`"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := typesv2.AccountAddress{}
		_ = addr.UnmarshalJSON(data)
	}
}
