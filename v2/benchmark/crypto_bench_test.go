package benchmark

import (
	"testing"

	bcsv1 "github.com/aptos-labs/aptos-go-sdk/bcs"
	cryptov1 "github.com/aptos-labs/aptos-go-sdk/crypto"
	bcsv2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	cryptov2 "github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
)

// Benchmark Ed25519 key generation

func BenchmarkCrypto_V1_Ed25519_Generate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cryptov1.GenerateEd25519PrivateKey()
	}
}

func BenchmarkCrypto_V2_Ed25519_Generate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cryptov2.GenerateEd25519PrivateKey()
	}
}

// Benchmark Ed25519 signing

func BenchmarkCrypto_V1_Ed25519_Sign(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	msg := []byte("The quick brown fox jumps over the lazy dog. This is a test message for signing benchmarks.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = key.SignMessage(msg)
	}
}

func BenchmarkCrypto_V2_Ed25519_Sign(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	msg := []byte("The quick brown fox jumps over the lazy dog. This is a test message for signing benchmarks.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = key.SignMessage(msg)
	}
}

// Benchmark Ed25519 signature verification

func BenchmarkCrypto_V1_Ed25519_Verify(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	msg := []byte("The quick brown fox jumps over the lazy dog. This is a test message for signing benchmarks.")
	sig, _ := key.SignMessage(msg)
	pubKey := key.PubKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pubKey.Verify(msg, sig)
	}
}

func BenchmarkCrypto_V2_Ed25519_Verify(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	msg := []byte("The quick brown fox jumps over the lazy dog. This is a test message for signing benchmarks.")
	sig, _ := key.SignMessage(msg)
	pubKey := key.PubKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pubKey.Verify(msg, sig)
	}
}

// Benchmark Ed25519 full Sign (with authenticator creation)

func BenchmarkCrypto_V1_Ed25519_FullSign(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	msg := []byte("The quick brown fox jumps over the lazy dog. This is a test message for signing benchmarks.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = key.Sign(msg)
	}
}

func BenchmarkCrypto_V2_Ed25519_FullSign(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	msg := []byte("The quick brown fox jumps over the lazy dog. This is a test message for signing benchmarks.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = key.Sign(msg)
	}
}

// Benchmark public key derivation

func BenchmarkCrypto_V1_Ed25519_PubKey(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = key.PubKey()
	}
}

func BenchmarkCrypto_V2_Ed25519_PubKey(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = key.PubKey()
	}
}

// Benchmark AuthKey derivation

func BenchmarkCrypto_V1_Ed25519_AuthKey(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = key.AuthKey()
	}
}

func BenchmarkCrypto_V2_Ed25519_AuthKey(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = key.AuthKey()
	}
}

// Benchmark key serialization

func BenchmarkCrypto_V1_Ed25519_ToHex(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = key.ToHex()
	}
}

func BenchmarkCrypto_V2_Ed25519_ToHex(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = key.ToHex()
	}
}

func BenchmarkCrypto_V1_Ed25519_FromHex(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	hexStr := key.ToHex()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newKey := &cryptov1.Ed25519PrivateKey{}
		_ = newKey.FromHex(hexStr)
	}
}

func BenchmarkCrypto_V2_Ed25519_FromHex(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	hexStr := key.ToHex()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newKey := &cryptov2.Ed25519PrivateKey{}
		_ = newKey.FromHex(hexStr)
	}
}

// Benchmark AIP-80 formatting

func BenchmarkCrypto_V1_Ed25519_ToAIP80(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = key.ToAIP80()
	}
}

func BenchmarkCrypto_V2_Ed25519_ToAIP80(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = key.ToAIP80()
	}
}

// Benchmark BCS serialization of keys

func BenchmarkCrypto_V1_Ed25519_PubKey_BCS(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	pubKey, ok := key.PubKey().(*cryptov1.Ed25519PublicKey)
	if !ok {
		b.Fatal("expected Ed25519PublicKey")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv1.Serialize(pubKey)
	}
}

func BenchmarkCrypto_V2_Ed25519_PubKey_BCS(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	pubKey, ok := key.PubKey().(*cryptov2.Ed25519PublicKey)
	if !ok {
		b.Fatal("expected Ed25519PublicKey")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv2.Serialize(pubKey)
	}
}

func BenchmarkCrypto_V1_Ed25519_Signature_BCS(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	msg := []byte("test message")
	sig, _ := key.SignMessage(msg)
	ed25519Sig, ok := sig.(*cryptov1.Ed25519Signature)
	if !ok {
		b.Fatal("expected Ed25519Signature")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv1.Serialize(ed25519Sig)
	}
}

func BenchmarkCrypto_V2_Ed25519_Signature_BCS(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	msg := []byte("test message")
	sig, _ := key.SignMessage(msg)
	ed25519Sig, ok := sig.(*cryptov2.Ed25519Signature)
	if !ok {
		b.Fatal("expected Ed25519Signature")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv2.Serialize(ed25519Sig)
	}
}

// Benchmark Authenticator BCS serialization

func BenchmarkCrypto_V1_Ed25519_Authenticator_BCS(b *testing.B) {
	key, _ := cryptov1.GenerateEd25519PrivateKey()
	msg := []byte("test message")
	auth, _ := key.Sign(msg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv1.Serialize(auth)
	}
}

func BenchmarkCrypto_V2_Ed25519_Authenticator_BCS(b *testing.B) {
	key, _ := cryptov2.GenerateEd25519PrivateKey()
	msg := []byte("test message")
	auth, _ := key.Sign(msg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcsv2.Serialize(auth)
	}
}
