package benchmark

import (
	"math/big"
	"testing"

	bcsv1 "github.com/aptos-labs/aptos-go-sdk/bcs"
	bcsv2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// SampleStructV1 is a test struct for v1 BCS benchmarks
type SampleStructV1 struct {
	Num     uint64
	Enabled bool
	Data    []byte
	Name    string
}

func (s *SampleStructV1) MarshalBCS(ser *bcsv1.Serializer) {
	ser.U64(s.Num)
	ser.Bool(s.Enabled)
	ser.WriteBytes(s.Data)
	ser.WriteString(s.Name)
}

func (s *SampleStructV1) UnmarshalBCS(des *bcsv1.Deserializer) {
	s.Num = des.U64()
	s.Enabled = des.Bool()
	s.Data = des.ReadBytes()
	s.Name = des.ReadString()
}

// SampleStructV2 is a test struct for v2 BCS benchmarks
type SampleStructV2 struct {
	Num     uint64
	Enabled bool
	Data    []byte
	Name    string
}

func (s *SampleStructV2) MarshalBCS(ser *bcsv2.Serializer) {
	ser.U64(s.Num)
	ser.Bool(s.Enabled)
	ser.WriteBytes(s.Data)
	ser.WriteString(s.Name)
}

func (s *SampleStructV2) UnmarshalBCS(des *bcsv2.Deserializer) {
	s.Num = des.U64()
	s.Enabled = des.Bool()
	s.Data = des.ReadBytes()
	s.Name = des.ReadString()
}

// Benchmark BCS primitive serialization

func BenchmarkBCS_V1_SerializeU64(b *testing.B) {
	ser := &bcsv1.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.U64(12345678901234567890)
		ser.Reset()
	}
}

func BenchmarkBCS_V2_SerializeU64(b *testing.B) {
	ser := &bcsv2.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.U64(12345678901234567890)
		ser.Reset()
	}
}

func BenchmarkBCS_V1_SerializeU128(b *testing.B) {
	val := big.NewInt(0)
	val.SetString("340282366920938463463374607431768211455", 10) // max u128
	ser := &bcsv1.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.U128(*val)
		ser.Reset()
	}
}

func BenchmarkBCS_V2_SerializeU128(b *testing.B) {
	val := big.NewInt(0)
	val.SetString("340282366920938463463374607431768211455", 10) // max u128
	ser := &bcsv2.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.U128(val)
		ser.Reset()
	}
}

func BenchmarkBCS_V1_SerializeBytes(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	ser := &bcsv1.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.WriteBytes(data)
		ser.Reset()
	}
}

func BenchmarkBCS_V2_SerializeBytes(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	ser := &bcsv2.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.WriteBytes(data)
		ser.Reset()
	}
}

func BenchmarkBCS_V1_SerializeString(b *testing.B) {
	str := "The quick brown fox jumps over the lazy dog. This is a reasonably long string for benchmarking purposes."
	ser := &bcsv1.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.WriteString(str)
		ser.Reset()
	}
}

func BenchmarkBCS_V2_SerializeString(b *testing.B) {
	str := "The quick brown fox jumps over the lazy dog. This is a reasonably long string for benchmarking purposes."
	ser := &bcsv2.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.WriteString(str)
		ser.Reset()
	}
}

func BenchmarkBCS_V1_SerializeUleb128(b *testing.B) {
	ser := &bcsv1.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.Uleb128(16383) // 2-byte ULEB128
		ser.Reset()
	}
}

func BenchmarkBCS_V2_SerializeUleb128(b *testing.B) {
	ser := &bcsv2.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.Uleb128(16383) // 2-byte ULEB128
		ser.Reset()
	}
}

// Benchmark struct serialization

func BenchmarkBCS_V1_SerializeStruct(b *testing.B) {
	s := &SampleStructV1{
		Num:     12345678901234567890,
		Enabled: true,
		Data:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Name:    "TestAccount",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bcsv1.Serialize(s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBCS_V2_SerializeStruct(b *testing.B) {
	s := &SampleStructV2{
		Num:     12345678901234567890,
		Enabled: true,
		Data:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Name:    "TestAccount",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bcsv2.Serialize(s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark deserialization

func BenchmarkBCS_V1_DeserializeU64(b *testing.B) {
	data, _ := bcsv1.SerializeU64(12345678901234567890)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		des := bcsv1.NewDeserializer(data)
		_ = des.U64()
	}
}

func BenchmarkBCS_V2_DeserializeU64(b *testing.B) {
	data := []byte{0xD2, 0x02, 0x96, 0x49, 0x1B, 0x3C, 0xF8, 0xAB} // 12345678901234567890 in LE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		des := bcsv2.NewDeserializer(data)
		_ = des.U64()
	}
}

func BenchmarkBCS_V1_DeserializeStruct(b *testing.B) {
	s := &SampleStructV1{
		Num:     12345678901234567890,
		Enabled: true,
		Data:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Name:    "TestAccount",
	}
	data, _ := bcsv1.Serialize(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &SampleStructV1{}
		des := bcsv1.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkBCS_V2_DeserializeStruct(b *testing.B) {
	s := &SampleStructV2{
		Num:     12345678901234567890,
		Enabled: true,
		Data:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Name:    "TestAccount",
	}
	data, _ := bcsv2.Serialize(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &SampleStructV2{}
		des := bcsv2.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

// Benchmark sequence serialization

func BenchmarkBCS_V1_SerializeSequence(b *testing.B) {
	items := make([]uint64, 100)
	for i := range items {
		items[i] = uint64(i * 12345)
	}
	ser := &bcsv1.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcsv1.SerializeSequenceWithFunction(items, ser, func(ser *bcsv1.Serializer, item uint64) {
			ser.U64(item)
		})
		ser.Reset()
	}
}

func BenchmarkBCS_V2_SerializeSequence(b *testing.B) {
	items := make([]uint64, 100)
	for i := range items {
		items[i] = uint64(i * 12345)
	}
	ser := &bcsv2.Serializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcsv2.SerializeSequenceFunc(ser, items, func(ser *bcsv2.Serializer, item uint64) {
			ser.U64(item)
		})
		ser.Reset()
	}
}

// Benchmark combined serialize + deserialize round-trip

func BenchmarkBCS_V1_RoundTrip(b *testing.B) {
	s := &SampleStructV1{
		Num:     12345678901234567890,
		Enabled: true,
		Data:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Name:    "TestAccount",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := bcsv1.Serialize(s)
		if err != nil {
			b.Fatal(err)
		}
		result := &SampleStructV1{}
		des := bcsv1.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}

func BenchmarkBCS_V2_RoundTrip(b *testing.B) {
	s := &SampleStructV2{
		Num:     12345678901234567890,
		Enabled: true,
		Data:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Name:    "TestAccount",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := bcsv2.Serialize(s)
		if err != nil {
			b.Fatal(err)
		}
		result := &SampleStructV2{}
		des := bcsv2.NewDeserializer(data)
		result.UnmarshalBCS(des)
	}
}
