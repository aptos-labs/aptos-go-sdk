package benchmark

import (
	"testing"

	aptosv1 "github.com/aptos-labs/aptos-go-sdk"
	bcsv1 "github.com/aptos-labs/aptos-go-sdk/bcs"
	bcsv2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	typesv2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
)

// Benchmark TypeTag parsing

func BenchmarkTypeTag_V1_ParsePrimitive(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aptosv1.ParseTypeTag("u64")
	}
}

func BenchmarkTypeTag_V2_ParsePrimitive(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseTypeTag("u64")
	}
}

func BenchmarkTypeTag_V1_ParseVector(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aptosv1.ParseTypeTag("vector<u8>")
	}
}

func BenchmarkTypeTag_V2_ParseVector(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseTypeTag("vector<u8>")
	}
}

func BenchmarkTypeTag_V1_ParseStruct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aptosv1.ParseTypeTag("0x1::coin::CoinStore")
	}
}

func BenchmarkTypeTag_V2_ParseStruct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseTypeTag("0x1::coin::CoinStore")
	}
}

func BenchmarkTypeTag_V1_ParseNestedGeneric(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aptosv1.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	}
}

func BenchmarkTypeTag_V2_ParseNestedGeneric(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	}
}

func BenchmarkTypeTag_V1_ParseComplexNested(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aptosv1.ParseTypeTag("0x1::fungible_asset::FungibleStore<0x1::aptos_framework::FungibleAsset<0x1::coin::CoinStore<u64>>>")
	}
}

func BenchmarkTypeTag_V2_ParseComplexNested(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseTypeTag("0x1::fungible_asset::FungibleStore<0x1::aptos_framework::FungibleAsset<0x1::coin::CoinStore<u64>>>")
	}
}

func BenchmarkTypeTag_V1_ParseMultipleGenerics(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aptosv1.ParseTypeTag("0x1::pair::Pair<u64, u128>")
	}
}

func BenchmarkTypeTag_V2_ParseMultipleGenerics(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = typesv2.ParseTypeTag("0x1::pair::Pair<u64, u128>")
	}
}

// Benchmark TypeTag String conversion

func BenchmarkTypeTag_V1_StringPrimitive(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("u64")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tag.String()
	}
}

func BenchmarkTypeTag_V2_StringPrimitive(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("u64")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tag.String()
	}
}

func BenchmarkTypeTag_V1_StringStruct(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tag.String()
	}
}

func BenchmarkTypeTag_V2_StringStruct(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tag.String()
	}
}

// Benchmark TypeTag BCS serialization

func BenchmarkTypeTag_V1_BCS_SerializePrimitive(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("u64")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv1.Serialize(tag)
	}
}

func BenchmarkTypeTag_V2_BCS_SerializePrimitive(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("u64")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv2.Serialize(tag)
	}
}

func BenchmarkTypeTag_V1_BCS_SerializeStruct(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv1.Serialize(tag)
	}
}

func BenchmarkTypeTag_V2_BCS_SerializeStruct(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv2.Serialize(tag)
	}
}

func BenchmarkTypeTag_V1_BCS_DeserializePrimitive(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("u64")
	data, _ := bcsv1.Serialize(tag)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &aptosv1.TypeTag{}
		des := bcsv1.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkTypeTag_V2_BCS_DeserializePrimitive(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("u64")
	data, _ := bcsv2.Serialize(tag)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &typesv2.TypeTag{}
		des := bcsv2.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkTypeTag_V1_BCS_DeserializeStruct(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	data, _ := bcsv1.Serialize(tag)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &aptosv1.TypeTag{}
		des := bcsv1.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkTypeTag_V2_BCS_DeserializeStruct(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	data, _ := bcsv2.Serialize(tag)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &typesv2.TypeTag{}
		des := bcsv2.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

// Benchmark TypeTag round-trip

func BenchmarkTypeTag_V1_RoundTrip(b *testing.B) {
	tag, _ := aptosv1.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := bcsv1.Serialize(tag)
		result := &aptosv1.TypeTag{}
		des := bcsv1.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkTypeTag_V2_RoundTrip(b *testing.B) {
	tag, _ := typesv2.ParseTypeTag("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := bcsv2.Serialize(tag)
		result := &typesv2.TypeTag{}
		des := bcsv2.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}
