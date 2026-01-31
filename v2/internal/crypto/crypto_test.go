package crypto

import (
	"fmt"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd25519KeyGeneration(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotNil(t, key.PubKey())
	assert.Len(t, key.Bytes(), 32)
}

func TestEd25519SignAndVerify(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.PubKey()
	ed25519PubKey, ok := pubKey.(*Ed25519PublicKey)
	require.True(t, ok, "expected Ed25519PublicKey")
	assert.True(t, ed25519PubKey.Verify(msg, sig))
	assert.False(t, ed25519PubKey.Verify([]byte("wrong message"), sig))
}

func TestEd25519AccountAuthenticator(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	assert.Equal(t, AccountAuthenticatorEd25519, auth.Variant)
	assert.True(t, auth.Verify(msg))
}

func TestEd25519AuthenticationKey(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	authKey := key.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), 32)
}

func TestEd25519HexSerialization(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	hex := key.ToHex()
	assert.NotEmpty(t, hex)
	assert.Equal(t, "0x", hex[:2])

	key2 := &Ed25519PrivateKey{}
	err = key2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, key.Bytes(), key2.Bytes())
}

func TestEd25519AIP80Format(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	aip80, err := key.ToAIP80()
	require.NoError(t, err)
	assert.NotEmpty(t, aip80)
	assert.Contains(t, aip80, "ed25519-priv-")
}

func TestSecp256k1KeyGeneration(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotNil(t, key.VerifyingKey())
	assert.Len(t, key.Bytes(), 32)
}

func TestSecp256k1SignAndVerify(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.VerifyingKey()
	secp256k1Pub, ok := pubKey.(*Secp256k1PublicKey)
	require.True(t, ok, "expected Secp256k1PublicKey")
	assert.True(t, secp256k1Pub.Verify(msg, sig))
	assert.False(t, secp256k1Pub.Verify([]byte("wrong message"), sig))
}

func TestSingleSignerWithSecp256k1(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	assert.NotNil(t, signer)

	msg := []byte("test message")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)

	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)
	assert.True(t, auth.Verify(msg))
}

func TestSingleSignerWithEd25519(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	assert.NotNil(t, signer)

	msg := []byte("test message")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)

	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)
	assert.True(t, auth.Verify(msg))
}

func TestAuthenticationKeyDeriveSchemes(t *testing.T) {
	ed25519Key, _ := GenerateEd25519PrivateKey()
	ed25519AuthKey := ed25519Key.AuthKey()

	// Verify the key is 32 bytes
	assert.Len(t, ed25519AuthKey.Bytes(), 32)

	// SingleSigner uses SingleKeyScheme
	secpKey, _ := GenerateSecp256k1Key()
	singleSigner := NewSingleSigner(secpKey)
	singleAuthKey := singleSigner.AuthKey()
	assert.Len(t, singleAuthKey.Bytes(), 32)

	// Auth keys should be different
	assert.NotEqual(t, ed25519AuthKey, singleAuthKey)
}

func TestEd25519BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok, "expected Ed25519PublicKey")

	// Serialize
	data, err := bcs.Serialize(pubKey)
	require.NoError(t, err)

	// Deserialize
	pubKey2 := &Ed25519PublicKey{}
	err = bcs.Deserialize(pubKey2, data)
	require.NoError(t, err)

	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestAccountAuthenticatorBCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	// Serialize
	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	// Deserialize
	auth2 := &AccountAuthenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)

	assert.Equal(t, auth.Variant, auth2.Variant)
	assert.True(t, auth2.Verify(msg))
}

func TestSimulationAuthenticator(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	simAuth := key.SimulationAuthenticator()
	assert.NotNil(t, simAuth)
	assert.Equal(t, AccountAuthenticatorEd25519, simAuth.Variant)
}

func TestSingleSignerSimulationAuthenticator(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	simAuth := signer.SimulationAuthenticator()
	assert.NotNil(t, simAuth)
	assert.Equal(t, AccountAuthenticatorSingleSender, simAuth.Variant)
}

func TestNoAccountAuthenticator(t *testing.T) {
	auth := NoAccountAuthenticator()
	assert.Equal(t, AccountAuthenticatorNone, auth.Variant)
	assert.Nil(t, auth.PubKey())
	assert.Nil(t, auth.Signature())
	assert.False(t, auth.Verify([]byte("test")))
}

func TestAnyPublicKeyConversion(t *testing.T) {
	// Ed25519
	ed25519Key, _ := GenerateEd25519PrivateKey()
	ed25519PubKey, ok := ed25519Key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok, "expected Ed25519PublicKey")
	anyPubKey, err := ToAnyPublicKey(ed25519PubKey)
	require.NoError(t, err)
	assert.Equal(t, AnyPublicKeyVariantEd25519, anyPubKey.Variant)

	// Secp256k1
	secpKey, _ := GenerateSecp256k1Key()
	secpPubKey, ok := secpKey.VerifyingKey().(*Secp256k1PublicKey)
	require.True(t, ok, "expected Secp256k1PublicKey")
	anyPubKey2, err := ToAnyPublicKey(secpPubKey)
	require.NoError(t, err)
	assert.Equal(t, AnyPublicKeyVariantSecp256k1, anyPubKey2.Variant)
}

func TestAuthenticationKeyHex(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	authKey := key.AuthKey()
	hex := authKey.ToHex()
	assert.NotEmpty(t, hex)

	authKey2 := &AuthenticationKey{}
	err = authKey2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, authKey, authKey2)
}

func TestMultiKeyBitmap(t *testing.T) {
	bm := &MultiKeyBitmap{}

	// Add some keys
	require.NoError(t, bm.AddKey(0))
	require.NoError(t, bm.AddKey(3))
	require.NoError(t, bm.AddKey(7))

	// Check containment
	assert.True(t, bm.ContainsKey(0))
	assert.False(t, bm.ContainsKey(1))
	assert.False(t, bm.ContainsKey(2))
	assert.True(t, bm.ContainsKey(3))
	assert.True(t, bm.ContainsKey(7))

	// Check indices
	indices := bm.Indices()
	assert.Equal(t, []uint8{0, 3, 7}, indices)

	// Cannot add duplicate
	err := bm.AddKey(0)
	require.Error(t, err)

	// Cannot add beyond max
	err = bm.AddKey(32)
	require.Error(t, err)
}

// Additional tests for improved coverage

func TestEd25519PublicKey_AuthKey(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok)

	authKey := pubKey.AuthKey()
	assert.Len(t, authKey.Bytes(), 32)
}

func TestEd25519PublicKey_Scheme(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok)

	assert.Equal(t, Ed25519Scheme, pubKey.Scheme())
}

func TestEd25519PublicKey_ToHex(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok)

	hex := pubKey.ToHex()
	assert.NotEmpty(t, hex)
	assert.True(t, len(hex) > 2)
}

func TestEd25519PublicKey_FromHex(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok)

	hex := pubKey.ToHex()
	pubKey2 := &Ed25519PublicKey{}
	err = pubKey2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestEd25519Signature_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	ed25519Sig, ok := sig.(*Ed25519Signature)
	require.True(t, ok)

	// Serialize
	data, err := bcs.Serialize(ed25519Sig)
	require.NoError(t, err)

	// Deserialize
	sig2 := &Ed25519Signature{}
	err = bcs.Deserialize(sig2, data)
	require.NoError(t, err)

	assert.Equal(t, ed25519Sig.Bytes(), sig2.Bytes())
}

func TestEd25519Signature_ToHex(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	ed25519Sig, ok := sig.(*Ed25519Signature)
	require.True(t, ok)

	hex := ed25519Sig.ToHex()
	assert.NotEmpty(t, hex)
}

func TestEd25519Authenticator_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	ed25519Auth, ok := auth.Auth.(*Ed25519Authenticator)
	require.True(t, ok)

	// Serialize
	data, err := bcs.Serialize(ed25519Auth)
	require.NoError(t, err)

	// Deserialize
	auth2 := &Ed25519Authenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)

	assert.Equal(t, ed25519Auth.PubKey.Bytes(), auth2.PubKey.Bytes())
}

func TestSecp256k1PublicKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	pubKey, ok := key.VerifyingKey().(*Secp256k1PublicKey)
	require.True(t, ok)

	// Serialize
	data, err := bcs.Serialize(pubKey)
	require.NoError(t, err)

	// Deserialize
	pubKey2 := &Secp256k1PublicKey{}
	err = bcs.Deserialize(pubKey2, data)
	require.NoError(t, err)

	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestSecp256k1Signature_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	secpSig, ok := sig.(*Secp256k1Signature)
	require.True(t, ok)

	// Serialize
	data, err := bcs.Serialize(secpSig)
	require.NoError(t, err)

	// Deserialize
	sig2 := &Secp256k1Signature{}
	err = bcs.Deserialize(sig2, data)
	require.NoError(t, err)

	assert.Equal(t, secpSig.Bytes(), sig2.Bytes())
}

func TestSecp256k1PrivateKey_ToHex(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	hex := key.ToHex()
	assert.NotEmpty(t, hex)
	assert.True(t, len(hex) > 2)
}

func TestSecp256k1PrivateKey_FromHex(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	hex := key.ToHex()
	key2 := &Secp256k1PrivateKey{}
	err = key2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, key.Bytes(), key2.Bytes())
}

func TestSecp256k1PrivateKey_ToAIP80(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	aip80, err := key.ToAIP80()
	require.NoError(t, err)
	assert.Contains(t, aip80, "secp256k1-priv-")
}

func TestSingleSigner_AuthKey(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	authKey := signer.AuthKey()
	assert.Len(t, authKey.Bytes(), 32)
}

func TestSingleSigner_PubKey(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	pubKey := signer.PubKey()
	assert.NotNil(t, pubKey)
}

func TestSingleKeyAuthenticator_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	signer := NewSingleSigner(key)
	msg := []byte("test message")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)

	singleKeyAuth, ok := auth.Auth.(*SingleKeyAuthenticator)
	require.True(t, ok)

	// Serialize
	data, err := bcs.Serialize(singleKeyAuth)
	require.NoError(t, err)

	// Deserialize
	auth2 := &SingleKeyAuthenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)

	assert.Equal(t, singleKeyAuth.PubKey.Variant, auth2.PubKey.Variant)
}

func TestAnyPublicKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	require.True(t, ok)

	anyPub, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)

	// Serialize
	data, err := bcs.Serialize(anyPub)
	require.NoError(t, err)

	// Deserialize
	anyPub2 := &AnyPublicKey{}
	err = bcs.Deserialize(anyPub2, data)
	require.NoError(t, err)

	assert.Equal(t, anyPub.Variant, anyPub2.Variant)
}

func TestAnySignature_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	ed25519Sig, ok := sig.(*Ed25519Signature)
	require.True(t, ok)

	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: ed25519Sig,
	}

	// Serialize
	data, err := bcs.Serialize(anySig)
	require.NoError(t, err)

	// Deserialize
	anySig2 := &AnySignature{}
	err = bcs.Deserialize(anySig2, data)
	require.NoError(t, err)

	assert.Equal(t, anySig.Variant, anySig2.Variant)
}

func TestAuthenticationKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	authKey := key.AuthKey()

	// Serialize
	data, err := bcs.Serialize(authKey)
	require.NoError(t, err)

	// Deserialize
	authKey2 := &AuthenticationKey{}
	err = bcs.Deserialize(authKey2, data)
	require.NoError(t, err)

	assert.Equal(t, authKey.Bytes(), authKey2.Bytes())
}

func TestAuthenticationKey_FromBytes(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	authKey := key.AuthKey()
	bytes := authKey.Bytes()

	authKey2 := &AuthenticationKey{}
	err = authKey2.FromBytes(bytes)
	require.NoError(t, err)
	assert.Equal(t, authKey.Bytes(), authKey2.Bytes())
}

func TestAuthenticationKey_FromBytes_InvalidLength(t *testing.T) {
	authKey := &AuthenticationKey{}
	err := authKey.FromBytes([]byte{1, 2, 3}) // Too short
	require.Error(t, err)
}

func TestEd25519PrivateKey_FromBytes_InvalidLength(t *testing.T) {
	key := &Ed25519PrivateKey{}
	err := key.FromBytes([]byte{1, 2, 3}) // Too short
	require.Error(t, err)
}

func TestEd25519PublicKey_FromBytes_InvalidLength(t *testing.T) {
	pubKey := &Ed25519PublicKey{}
	err := pubKey.FromBytes([]byte{1, 2, 3}) // Too short
	require.Error(t, err)
}

func TestEd25519Signature_FromBytes_InvalidLength(t *testing.T) {
	sig := &Ed25519Signature{}
	err := sig.FromBytes([]byte{1, 2, 3}) // Too short
	require.Error(t, err)
}

func TestSecp256k1PrivateKey_FromBytes_InvalidLength(t *testing.T) {
	key := &Secp256k1PrivateKey{}
	err := key.FromBytes([]byte{1, 2, 3}) // Too short
	require.Error(t, err)
}

func TestSecp256k1PublicKey_FromBytes_InvalidLength(t *testing.T) {
	pubKey := &Secp256k1PublicKey{}
	err := pubKey.FromBytes([]byte{1, 2, 3}) // Too short
	require.Error(t, err)
}

func TestEd25519PrivateKey_EmptySignature(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	emptySig := key.EmptySignature()
	assert.NotNil(t, emptySig)
}

func TestEd25519PrivateKey_String(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	str := key.String()
	assert.NotEmpty(t, str)
}

func TestMultiEd25519PublicKey_Bytes(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2},
		SignaturesRequired: 1,
	}

	bytes := multiKey.Bytes()
	assert.NotEmpty(t, bytes)
}

func TestMultiEd25519PublicKey_FromBytes(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2},
		SignaturesRequired: 1,
	}

	bytes := multiKey.Bytes()

	multiKey2 := &MultiEd25519PublicKey{}
	err := multiKey2.FromBytes(bytes)
	require.NoError(t, err)
	assert.Equal(t, multiKey.SignaturesRequired, multiKey2.SignaturesRequired)
	assert.Len(t, multiKey2.PubKeys, 2)
}

func TestFormatPrivateKey(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	formatted, err := FormatPrivateKey(key.Bytes(), PrivateKeyVariantEd25519)
	require.NoError(t, err)
	assert.Contains(t, formatted, "ed25519-priv-")
}

func TestParsePrivateKey(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	formatted, err := FormatPrivateKey(key.Bytes(), PrivateKeyVariantEd25519)
	require.NoError(t, err)

	parsed, err := ParsePrivateKey(formatted, PrivateKeyVariantEd25519)
	require.NoError(t, err)
	assert.Equal(t, key.Bytes(), parsed)
}

// ============ MultiEd25519 Comprehensive Tests ============

func TestMultiEd25519PublicKey_Verify(t *testing.T) {
	// Create 3 keys for a 2-of-3 multisig
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()
	key3, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)
	pub3, _ := key3.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2, pub3},
		SignaturesRequired: 2,
	}

	msg := []byte("test message")

	// Create signatures from key1 and key2
	sig1, _ := key1.SignMessage(msg)
	sig2, _ := key2.SignMessage(msg)

	// Bitmap: 0b11000000 = 0xC0 (keys 0 and 1 signed)
	multiSig := &MultiEd25519Signature{
		Signatures: []*Ed25519Signature{sig1.(*Ed25519Signature), sig2.(*Ed25519Signature)},
		Bitmap:     [MultiEd25519BitmapLen]byte{0xC0, 0, 0, 0},
	}

	assert.True(t, multiKey.Verify(msg, multiSig))

	// Test with wrong signature type
	wrongSig := &Ed25519Signature{}
	assert.False(t, multiKey.Verify(msg, wrongSig))

	// Test with insufficient signatures
	singleSig := &MultiEd25519Signature{
		Signatures: []*Ed25519Signature{sig1.(*Ed25519Signature)},
		Bitmap:     [MultiEd25519BitmapLen]byte{0x80, 0, 0, 0}, // Only key 0
	}
	assert.False(t, multiKey.Verify(msg, singleSig))
}

func TestMultiEd25519PublicKey_AuthKey(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2},
		SignaturesRequired: 1,
	}

	authKey := multiKey.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), AuthenticationKeyLength)
}

func TestMultiEd25519PublicKey_Scheme(t *testing.T) {
	multiKey := &MultiEd25519PublicKey{}
	assert.Equal(t, MultiEd25519Scheme, multiKey.Scheme())
}

func TestMultiEd25519PublicKey_HexRoundTrip(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2},
		SignaturesRequired: 2,
	}

	hex := multiKey.ToHex()
	assert.NotEmpty(t, hex)

	multiKey2 := &MultiEd25519PublicKey{}
	err := multiKey2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, multiKey.SignaturesRequired, multiKey2.SignaturesRequired)
	assert.Len(t, multiKey2.PubKeys, 2)
}

func TestMultiEd25519PublicKey_BCSRoundTrip(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2},
		SignaturesRequired: 1,
	}

	data, err := bcs.Serialize(multiKey)
	require.NoError(t, err)

	multiKey2 := &MultiEd25519PublicKey{}
	err = bcs.Deserialize(multiKey2, data)
	require.NoError(t, err)
	assert.Equal(t, multiKey.SignaturesRequired, multiKey2.SignaturesRequired)
}

func TestMultiEd25519Signature_HexRoundTrip(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	multiSig := &MultiEd25519Signature{
		Signatures: []*Ed25519Signature{sig.(*Ed25519Signature)},
		Bitmap:     [MultiEd25519BitmapLen]byte{0x80, 0, 0, 0},
	}

	hex := multiSig.ToHex()
	assert.NotEmpty(t, hex)

	multiSig2 := &MultiEd25519Signature{}
	err := multiSig2.FromHex(hex)
	require.NoError(t, err)
	assert.Len(t, multiSig2.Signatures, 1)
}

func TestMultiEd25519Signature_BCSRoundTrip(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	multiSig := &MultiEd25519Signature{
		Signatures: []*Ed25519Signature{sig.(*Ed25519Signature)},
		Bitmap:     [MultiEd25519BitmapLen]byte{0x80, 0, 0, 0},
	}

	data, err := bcs.Serialize(multiSig)
	require.NoError(t, err)

	multiSig2 := &MultiEd25519Signature{}
	err = bcs.Deserialize(multiSig2, data)
	require.NoError(t, err)
	assert.Len(t, multiSig2.Signatures, 1)
}

func TestMultiEd25519Authenticator(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateEd25519PrivateKey()

	pub1, _ := key1.PubKey().(*Ed25519PublicKey)
	pub2, _ := key2.PubKey().(*Ed25519PublicKey)

	multiKey := &MultiEd25519PublicKey{
		PubKeys:            []*Ed25519PublicKey{pub1, pub2},
		SignaturesRequired: 1,
	}

	msg := []byte("test message")
	sig1, _ := key1.SignMessage(msg)

	multiSig := &MultiEd25519Signature{
		Signatures: []*Ed25519Signature{sig1.(*Ed25519Signature)},
		Bitmap:     [MultiEd25519BitmapLen]byte{0x80, 0, 0, 0},
	}

	auth := &MultiEd25519Authenticator{
		PubKey: multiKey,
		Sig:    multiSig,
	}

	// Test interface methods
	assert.Equal(t, multiKey, auth.PublicKey())
	assert.Equal(t, multiSig, auth.Signature())
	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong")))

	// Test BCS round-trip
	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	auth2 := &MultiEd25519Authenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)
	assert.True(t, auth2.Verify(msg))
}

// ============ Secp256k1 Authenticator Tests ============

func TestSecp256k1Authenticator(t *testing.T) {
	key, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256k1PublicKey)

	auth := &Secp256k1Authenticator{
		PubKey: pubKey,
		Sig:    sig.(*Secp256k1Signature),
	}

	// Test interface methods
	assert.Equal(t, pubKey, auth.PublicKey())
	assert.Equal(t, sig, auth.Signature())
	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong")))

	// Test BCS round-trip
	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	auth2 := &Secp256k1Authenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)
	assert.True(t, auth2.Verify(msg))
}

func TestSecp256k1Signature_HexRoundTrip(t *testing.T) {
	key, _ := GenerateSecp256k1Key()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	secpSig := sig.(*Secp256k1Signature)
	hex := secpSig.ToHex()
	assert.NotEmpty(t, hex)

	sig2 := &Secp256k1Signature{}
	err := sig2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, secpSig.Bytes(), sig2.Bytes())
}

func TestSecp256k1PublicKey_ToHex(t *testing.T) {
	key, _ := GenerateSecp256k1Key()
	pubKey := key.VerifyingKey().(*Secp256k1PublicKey)

	hex := pubKey.ToHex()
	assert.NotEmpty(t, hex)
	assert.Contains(t, hex, "0x")
}

// ============ Ed25519 Authenticator Tests ============

func TestEd25519Authenticator_PublicKeyAndSignature(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	pubKey := key.PubKey().(*Ed25519PublicKey)

	auth := &Ed25519Authenticator{
		PubKey: pubKey,
		Sig:    sig.(*Ed25519Signature),
	}

	assert.Equal(t, pubKey, auth.PublicKey())
	assert.Equal(t, sig, auth.Signature())
}

func TestEd25519Signature_FromHex(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	ed25519Sig := sig.(*Ed25519Signature)
	hex := ed25519Sig.ToHex()

	sig2 := &Ed25519Signature{}
	err := sig2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, ed25519Sig.Bytes(), sig2.Bytes())
}

// ============ AnyPublicKey / AnySignature Tests ============

func TestAnyPublicKey_HexRoundTrip(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	pubKey := key.PubKey().(*Ed25519PublicKey)

	anyPub := &AnyPublicKey{
		Variant: AnyPublicKeyVariantEd25519,
		PubKey:  pubKey,
	}

	hex := anyPub.ToHex()
	assert.NotEmpty(t, hex)

	anyPub2 := &AnyPublicKey{}
	err := anyPub2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, anyPub.Variant, anyPub2.Variant)
}

func TestAnyPublicKey_AuthKey(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	pubKey := key.PubKey().(*Ed25519PublicKey)

	anyPub := &AnyPublicKey{
		Variant: AnyPublicKeyVariantEd25519,
		PubKey:  pubKey,
	}

	authKey := anyPub.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), AuthenticationKeyLength)
}

func TestAnySignature_HexRoundTrip(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: sig.(*Ed25519Signature),
	}

	hex := anySig.ToHex()
	assert.NotEmpty(t, hex)

	anySig2 := &AnySignature{}
	err := anySig2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, anySig.Variant, anySig2.Variant)
}

// ============ SingleKeyAuthenticator Tests ============

func TestSingleKeyAuthenticator_PublicKeyAndSignature(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	signer := NewSingleSigner(key)

	msg := []byte("test")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)

	singleAuth := auth.Auth.(*SingleKeyAuthenticator)
	assert.NotNil(t, singleAuth.PublicKey())
	assert.NotNil(t, singleAuth.Signature())
}

// ============ MultiKey Tests ============

func TestMultiKey_Verify(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	key2, _ := GenerateSecp256k1Key()

	signer1 := NewSingleSigner(key1)
	signer2 := NewSingleSigner(key2)

	anyPub1, _ := ToAnyPublicKey(signer1.PubKey())
	anyPub2, _ := ToAnyPublicKey(signer2.PubKey())

	multiKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyPub1, anyPub2},
		SignaturesRequired: 1,
	}

	// Test Scheme
	assert.Equal(t, MultiKeyScheme, multiKey.Scheme())

	// Test AuthKey
	authKey := multiKey.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), AuthenticationKeyLength)
}

func TestMultiKey_HexRoundTrip(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	signer1 := NewSingleSigner(key1)
	anyPub1, _ := ToAnyPublicKey(signer1.PubKey())

	multiKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyPub1},
		SignaturesRequired: 1,
	}

	hex := multiKey.ToHex()
	assert.NotEmpty(t, hex)

	multiKey2 := &MultiKey{}
	err := multiKey2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, multiKey.SignaturesRequired, multiKey2.SignaturesRequired)
}

func TestMultiKey_BCSRoundTrip(t *testing.T) {
	key1, _ := GenerateEd25519PrivateKey()
	signer1 := NewSingleSigner(key1)
	anyPub1, _ := ToAnyPublicKey(signer1.PubKey())

	multiKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyPub1},
		SignaturesRequired: 1,
	}

	data, err := bcs.Serialize(multiKey)
	require.NoError(t, err)

	multiKey2 := &MultiKey{}
	err = bcs.Deserialize(multiKey2, data)
	require.NoError(t, err)
	assert.Equal(t, multiKey.SignaturesRequired, multiKey2.SignaturesRequired)
}

func TestMultiKeySignature_BCSRoundTrip(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: sig.(*Ed25519Signature),
	}

	// Use NewMultiKeySignature which properly sets up the bitmap
	multiSig, err := NewMultiKeySignature(1, []IndexedAnySignature{
		{Index: 0, Signature: anySig},
	})
	require.NoError(t, err)

	data, err := bcs.Serialize(multiSig)
	require.NoError(t, err)

	multiSig2 := &MultiKeySignature{}
	err = bcs.Deserialize(multiSig2, data)
	require.NoError(t, err)
	assert.Len(t, multiSig2.Signatures, 1)
}

func TestMultiKeyAuthenticator_BCSRoundTrip(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	signer := NewSingleSigner(key)
	anyPub, _ := ToAnyPublicKey(signer.PubKey())

	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: sig.(*Ed25519Signature),
	}

	multiKey := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyPub},
		SignaturesRequired: 1,
	}

	multiSig, err := NewMultiKeySignature(uint8(len(multiKey.PubKeys)), []IndexedAnySignature{
		{Index: 0, Signature: anySig},
	})
	require.NoError(t, err)

	auth := &MultiKeyAuthenticator{
		PubKey: multiKey,
		Sig:    multiSig,
	}

	// Test interface methods
	assert.Equal(t, multiKey, auth.PublicKey())
	assert.Equal(t, multiSig, auth.Signature())

	// Test BCS round-trip
	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	auth2 := &MultiKeyAuthenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)
}

// ============ AccountAuthenticator FromKeyAndSignature Tests ============

func TestAccountAuthenticator_FromKeyAndSignature_Ed25519(t *testing.T) {
	key, _ := GenerateEd25519PrivateKey()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	signer := NewSingleSigner(key)
	anyPub, _ := ToAnyPublicKey(signer.PubKey())
	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: sig.(*Ed25519Signature),
	}

	auth := &AccountAuthenticator{}
	err := auth.FromKeyAndSignature(anyPub, anySig)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
}

func TestAccountAuthenticator_FromKeyAndSignature_Secp256k1(t *testing.T) {
	key, _ := GenerateSecp256k1Key()
	msg := []byte("test")
	sig, _ := key.SignMessage(msg)

	signer := NewSingleSigner(key)
	anyPub, _ := ToAnyPublicKey(signer.PubKey())
	anySig := &AnySignature{
		Variant:   AnySignatureVariantSecp256k1,
		Signature: sig.(*Secp256k1Signature),
	}

	auth := &AccountAuthenticator{}
	err := auth.FromKeyAndSignature(anyPub, anySig)
	require.NoError(t, err)
	assert.True(t, auth.Verify(msg))
}

// ============ Error Path Tests ============

func TestMultiEd25519PublicKey_FromBytes_TooShort(t *testing.T) {
	multiKey := &MultiEd25519PublicKey{}
	err := multiKey.FromBytes([]byte{1, 2, 3})
	require.Error(t, err)
}

func TestMultiEd25519Signature_FromBytes_TooShort(t *testing.T) {
	multiSig := &MultiEd25519Signature{}
	err := multiSig.FromBytes([]byte{1, 2, 3})
	require.Error(t, err)
}

func TestMultiEd25519PublicKey_FromHex_Invalid(t *testing.T) {
	multiKey := &MultiEd25519PublicKey{}
	err := multiKey.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestMultiEd25519Signature_FromHex_Invalid(t *testing.T) {
	multiSig := &MultiEd25519Signature{}
	err := multiSig.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestSecp256k1Signature_FromHex_Invalid(t *testing.T) {
	sig := &Secp256k1Signature{}
	err := sig.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestAnyPublicKey_FromHex_Invalid(t *testing.T) {
	anyPub := &AnyPublicKey{}
	err := anyPub.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestAnySignature_FromHex_Invalid(t *testing.T) {
	anySig := &AnySignature{}
	err := anySig.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestMultiKey_FromHex_Invalid(t *testing.T) {
	multiKey := &MultiKey{}
	err := multiKey.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestEd25519Signature_FromHex_Invalid(t *testing.T) {
	sig := &Ed25519Signature{}
	err := sig.FromHex("not-valid-hex")
	require.Error(t, err)
}

// ============================================================================
// SLH-DSA Tests
// ============================================================================

func TestSlhDsaKeyGeneration(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotNil(t, key.VerifyingKey())
	assert.Len(t, key.Bytes(), SlhDsaPrivateKeySize)
}

func TestSlhDsaSignAndVerify(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test message for post-quantum signing")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)
	assert.Len(t, sig.Bytes(), SlhDsaSignatureSize)

	pubKey := key.VerifyingKey()
	slhDsaPubKey, ok := pubKey.(*SlhDsaPublicKey)
	require.True(t, ok, "expected SlhDsaPublicKey")
	assert.True(t, slhDsaPubKey.Verify(msg, sig))
	assert.False(t, slhDsaPubKey.Verify([]byte("wrong message"), sig))
}

func TestSlhDsaPublicKey_Bytes(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	bytes := pubKey.Bytes()
	assert.Len(t, bytes, SlhDsaPublicKeySize)
}

func TestSlhDsaPublicKey_FromBytes(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	bytes := pubKey.Bytes()

	pubKey2 := &SlhDsaPublicKey{}
	err = pubKey2.FromBytes(bytes)
	require.NoError(t, err)
	assert.Equal(t, bytes, pubKey2.Bytes())
}

func TestSlhDsaPublicKey_Verify(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	assert.True(t, pubKey.Verify(msg, sig))
	assert.False(t, pubKey.Verify([]byte("different"), sig))
}

func TestSlhDsaPublicKey_AuthKey(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	authKey := pubKey.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), AuthenticationKeyLength)
}

func TestSlhDsaPublicKey_Scheme(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	assert.Equal(t, SingleKeyScheme, pubKey.Scheme())
}

func TestSlhDsaPublicKey_HexRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	hex := pubKey.ToHex()
	assert.True(t, len(hex) > 2) // "0x" + hex bytes

	pubKey2 := &SlhDsaPublicKey{}
	err = pubKey2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestSlhDsaPublicKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	data, err := bcs.Serialize(pubKey)
	require.NoError(t, err)

	pubKey2 := &SlhDsaPublicKey{}
	err = bcs.Deserialize(pubKey2, data)
	require.NoError(t, err)
	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestSlhDsaSignature_HexRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	slhSig := sig.(*SlhDsaSignature)
	hex := slhSig.ToHex()

	slhSig2 := &SlhDsaSignature{}
	err = slhSig2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, slhSig.Bytes(), slhSig2.Bytes())
}

func TestSlhDsaSignature_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	slhSig := sig.(*SlhDsaSignature)
	data, err := bcs.Serialize(slhSig)
	require.NoError(t, err)

	slhSig2 := &SlhDsaSignature{}
	err = bcs.Deserialize(slhSig2, data)
	require.NoError(t, err)
	assert.Equal(t, slhSig.Bytes(), slhSig2.Bytes())
}

func TestSlhDsaAuthenticator(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	slhSig := sig.(*SlhDsaSignature)

	auth := &SlhDsaAuthenticator{
		PubKey: pubKey,
		Sig:    slhSig,
	}

	assert.Equal(t, pubKey, auth.PublicKey())
	assert.Equal(t, slhSig, auth.Signature())
	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong")))
}

func TestSlhDsaAuthenticator_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	auth := &SlhDsaAuthenticator{
		PubKey: key.VerifyingKey().(*SlhDsaPublicKey),
		Sig:    sig.(*SlhDsaSignature),
	}

	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	auth2 := &SlhDsaAuthenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)

	assert.Equal(t, auth.PubKey.Bytes(), auth2.PubKey.Bytes())
	assert.Equal(t, auth.Sig.Bytes(), auth2.Sig.Bytes())
	assert.True(t, auth2.Verify(msg))
}

func TestSlhDsaPrivateKey_HexRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	hex := key.ToHex()
	assert.True(t, len(hex) > 2)

	key2 := &SlhDsaPrivateKey{}
	err = key2.FromHex(hex)
	require.NoError(t, err)

	// Verify they produce the same signatures
	msg := []byte("test")
	sig1, _ := key.SignMessage(msg)
	sig2, _ := key2.SignMessage(msg)
	assert.Equal(t, sig1.Bytes(), sig2.Bytes())
}

func TestSlhDsaPrivateKey_String(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	// String() should return redacted output
	str := key.String()
	assert.Contains(t, str, "REDACTED")
	assert.NotContains(t, str, key.ToHex())
}

func TestSlhDsaPrivateKey_ToAIP80(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	aip80, err := key.ToAIP80()
	require.NoError(t, err)
	assert.True(t, len(aip80) > 0)
	assert.Contains(t, aip80, "slhdsa-priv-")
}

func TestSlhDsa_ToAnyPublicKey(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	anyPub, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)
	assert.Equal(t, AnyPublicKeyVariantSlhDsaSha2_128s, anyPub.Variant)
	assert.Equal(t, pubKey, anyPub.PubKey)
}

func TestSlhDsa_AnyPublicKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*SlhDsaPublicKey)
	anyPub, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)

	data, err := bcs.Serialize(anyPub)
	require.NoError(t, err)

	anyPub2 := &AnyPublicKey{}
	err = bcs.Deserialize(anyPub2, data)
	require.NoError(t, err)

	assert.Equal(t, AnyPublicKeyVariantSlhDsaSha2_128s, anyPub2.Variant)
	slhPub, ok := anyPub2.PubKey.(*SlhDsaPublicKey)
	require.True(t, ok)
	assert.Equal(t, pubKey.Bytes(), slhPub.Bytes())
}

func TestSlhDsa_AnySignature_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	slhSig := sig.(*SlhDsaSignature)
	anySig := slhSig.ToAnySignature()

	data, err := bcs.Serialize(anySig)
	require.NoError(t, err)

	anySig2 := &AnySignature{}
	err = bcs.Deserialize(anySig2, data)
	require.NoError(t, err)

	assert.Equal(t, AnySignatureVariantSlhDsaSha2_128s, anySig2.Variant)
	slhSig2, ok := anySig2.Signature.(*SlhDsaSignature)
	require.True(t, ok)
	assert.Equal(t, slhSig.Bytes(), slhSig2.Bytes())
}

func TestSlhDsa_SingleSigner(t *testing.T) {
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)

	signer := NewSlhDsaSingleSigner(key)
	assert.NotNil(t, signer)

	msg := []byte("test message")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)
	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)
	assert.True(t, auth.Verify(msg))
}

func TestSlhDsaSignature_FromHex_Invalid(t *testing.T) {
	sig := &SlhDsaSignature{}
	err := sig.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestSlhDsaPublicKey_FromHex_Invalid(t *testing.T) {
	pub := &SlhDsaPublicKey{}
	err := pub.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestSlhDsaPrivateKey_FromHex_Invalid(t *testing.T) {
	key := &SlhDsaPrivateKey{}
	err := key.FromHex("not-valid-hex")
	require.Error(t, err)
}

// ============================================================================
// Secp256r1 (P-256) Tests
// ============================================================================

func TestSecp256r1KeyGeneration(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotNil(t, key.VerifyingKey())
	assert.Len(t, key.Bytes(), Secp256r1PrivateKeyLength)
}

func TestSecp256r1SignAndVerify(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test message for P-256 signing")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)
	assert.Len(t, sig.Bytes(), Secp256r1SignatureLength)

	pubKey := key.VerifyingKey()
	r1PubKey, ok := pubKey.(*Secp256r1PublicKey)
	require.True(t, ok, "expected Secp256r1PublicKey")
	assert.True(t, r1PubKey.Verify(msg, sig))
	assert.False(t, r1PubKey.Verify([]byte("wrong message"), sig))
}

func TestSecp256r1PublicKey_Bytes(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	bytes := pubKey.Bytes()
	assert.Len(t, bytes, Secp256r1PublicKeyLength)
	assert.Equal(t, byte(0x04), bytes[0], "uncompressed public key should start with 0x04")
}

func TestSecp256r1PublicKey_FromBytes(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	bytes := pubKey.Bytes()

	pubKey2 := &Secp256r1PublicKey{}
	err = pubKey2.FromBytes(bytes)
	require.NoError(t, err)
	assert.Equal(t, bytes, pubKey2.Bytes())
}

func TestSecp256r1PublicKey_Verify(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	assert.True(t, pubKey.Verify(msg, sig))
	assert.False(t, pubKey.Verify([]byte("different"), sig))
}

func TestSecp256r1PublicKey_AuthKey(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	authKey := pubKey.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), AuthenticationKeyLength)
}

func TestSecp256r1PublicKey_Scheme(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	assert.Equal(t, SingleKeyScheme, pubKey.Scheme())
}

func TestSecp256r1PublicKey_HexRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	hex := pubKey.ToHex()
	assert.True(t, len(hex) > 2) // "0x" + hex bytes

	pubKey2 := &Secp256r1PublicKey{}
	err = pubKey2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestSecp256r1PublicKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	data, err := bcs.Serialize(pubKey)
	require.NoError(t, err)

	pubKey2 := &Secp256r1PublicKey{}
	err = bcs.Deserialize(pubKey2, data)
	require.NoError(t, err)
	assert.Equal(t, pubKey.Bytes(), pubKey2.Bytes())
}

func TestSecp256r1Signature_HexRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	r1Sig := sig.(*Secp256r1Signature)
	hex := r1Sig.ToHex()

	r1Sig2 := &Secp256r1Signature{}
	err = r1Sig2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, r1Sig.Bytes(), r1Sig2.Bytes())
}

func TestSecp256r1Signature_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	r1Sig := sig.(*Secp256r1Signature)
	data, err := bcs.Serialize(r1Sig)
	require.NoError(t, err)

	r1Sig2 := &Secp256r1Signature{}
	err = bcs.Deserialize(r1Sig2, data)
	require.NoError(t, err)
	assert.Equal(t, r1Sig.Bytes(), r1Sig2.Bytes())
}

func TestSecp256r1Authenticator(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	r1Sig := sig.(*Secp256r1Signature)

	auth := &Secp256r1Authenticator{
		PubKey: pubKey,
		Sig:    r1Sig,
	}

	assert.Equal(t, pubKey, auth.PublicKey())
	assert.Equal(t, r1Sig, auth.Signature())
	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong")))
}

func TestSecp256r1Authenticator_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	auth := &Secp256r1Authenticator{
		PubKey: key.VerifyingKey().(*Secp256r1PublicKey),
		Sig:    sig.(*Secp256r1Signature),
	}

	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	auth2 := &Secp256r1Authenticator{}
	err = bcs.Deserialize(auth2, data)
	require.NoError(t, err)

	assert.Equal(t, auth.PubKey.Bytes(), auth2.PubKey.Bytes())
	assert.Equal(t, auth.Sig.Bytes(), auth2.Sig.Bytes())
	assert.True(t, auth2.Verify(msg))
}

func TestSecp256r1PrivateKey_HexRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	hex := key.ToHex()
	assert.True(t, len(hex) > 2)

	key2 := &Secp256r1PrivateKey{}
	err = key2.FromHex(hex)
	require.NoError(t, err)

	// Verify they produce the same public key
	assert.Equal(t, key.VerifyingKey().(*Secp256r1PublicKey).Bytes(),
		key2.VerifyingKey().(*Secp256r1PublicKey).Bytes())
}

func TestSecp256r1PrivateKey_String(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	// String() should return redacted output
	str := key.String()
	assert.Contains(t, str, "REDACTED")
	assert.NotContains(t, str, key.ToHex())
}

func TestSecp256r1PrivateKey_ToAIP80(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	aip80, err := key.ToAIP80()
	require.NoError(t, err)
	assert.True(t, len(aip80) > 0)
	assert.Contains(t, aip80, "secp256r1-priv-")
}

func TestSecp256r1_ToAnyPublicKey(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	anyPub, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)
	assert.Equal(t, AnyPublicKeyVariantSecp256r1, anyPub.Variant)
	assert.Equal(t, pubKey, anyPub.PubKey)
}

func TestSecp256r1_AnyPublicKey_BCSRoundTrip(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	pubKey := key.VerifyingKey().(*Secp256r1PublicKey)
	anyPub, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)

	data, err := bcs.Serialize(anyPub)
	require.NoError(t, err)

	anyPub2 := &AnyPublicKey{}
	err = bcs.Deserialize(anyPub2, data)
	require.NoError(t, err)

	assert.Equal(t, AnyPublicKeyVariantSecp256r1, anyPub2.Variant)
	r1Pub, ok := anyPub2.PubKey.(*Secp256r1PublicKey)
	require.True(t, ok)
	assert.Equal(t, pubKey.Bytes(), r1Pub.Bytes())
}

func TestSecp256r1_SingleSigner(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	signer := NewSecp256r1SingleSigner(key)
	assert.NotNil(t, signer)

	msg := []byte("test message")
	auth, err := signer.Sign(msg)
	require.NoError(t, err)
	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)
	assert.True(t, auth.Verify(msg))
}

func TestSecp256r1Signature_FromHex_Invalid(t *testing.T) {
	sig := &Secp256r1Signature{}
	err := sig.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestSecp256r1PublicKey_FromHex_Invalid(t *testing.T) {
	pub := &Secp256r1PublicKey{}
	err := pub.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestSecp256r1PrivateKey_FromHex_Invalid(t *testing.T) {
	key := &Secp256r1PrivateKey{}
	err := key.FromHex("not-valid-hex")
	require.Error(t, err)
}

func TestSecp256r1Signature_LowS(t *testing.T) {
	// Test that signatures have normalized (low) s values
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	// Sign multiple messages and verify s is always in low form
	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("test message %d", i))
		sig, err := key.SignMessage(msg)
		require.NoError(t, err)

		// Verify the signature can be deserialized (which checks low s)
		r1Sig := sig.(*Secp256r1Signature)
		sig2 := &Secp256r1Signature{}
		err = sig2.FromBytes(r1Sig.Bytes())
		require.NoError(t, err)
	}
}

// ============================================================================
// WebAuthn Signature Tests
// ============================================================================

func TestWebAuthn_AssertionSignature_BCSRoundTrip(t *testing.T) {
	// Create a Secp256r1 signature for WebAuthn
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test message for webauthn")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	r1Sig := sig.(*Secp256r1Signature)

	// Create an assertion signature
	assertionSig := &AssertionSignature{
		Variant:   AssertionSignatureVariantSecp256r1,
		Signature: r1Sig,
	}

	// Serialize
	data, err := bcs.Serialize(assertionSig)
	require.NoError(t, err)

	// Deserialize
	assertionSig2 := &AssertionSignature{}
	err = bcs.Deserialize(assertionSig2, data)
	require.NoError(t, err)

	assert.Equal(t, AssertionSignatureVariantSecp256r1, assertionSig2.Variant)
	assert.Equal(t, r1Sig.Bytes(), assertionSig2.Signature.Bytes())
}

func TestWebAuthn_PartialAuthenticatorAssertionResponse_BCSRoundTrip(t *testing.T) {
	// Create a Secp256r1 signature for WebAuthn
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test message for webauthn")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	r1Sig := sig.(*Secp256r1Signature)

	// Create mock authenticator data
	authenticatorData := []byte{
		73, 150, 13, 229, 136, 14, 140, 104, 116, 52, 23, 15, 100, 118, 96, 91,
		143, 228, 174, 185, 162, 134, 50, 199, 153, 92, 243, 186, 131, 29, 151, 99,
		29, 0, 0, 0, 0,
	}

	// Create mock client data JSON (simplified for testing)
	clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"dGVzdCBjaGFsbGVuZ2U","origin":"http://localhost:4000","crossOrigin":false}`)

	// Create the assertion response
	paar := NewPartialAuthenticatorAssertionResponse(r1Sig, authenticatorData, clientDataJSON)

	// Serialize
	data, err := bcs.Serialize(paar)
	require.NoError(t, err)

	// Deserialize
	paar2 := &PartialAuthenticatorAssertionResponse{}
	err = bcs.Deserialize(paar2, data)
	require.NoError(t, err)

	assert.Equal(t, AssertionSignatureVariantSecp256r1, paar2.Signature.Variant)
	assert.Equal(t, r1Sig.Bytes(), paar2.Signature.Signature.Bytes())
	assert.Equal(t, authenticatorData, paar2.AuthenticatorData)
	assert.Equal(t, clientDataJSON, paar2.ClientDataJSON)
}

func TestWebAuthn_AnySignature_WebAuthnVariant_BCSRoundTrip(t *testing.T) {
	// Create a Secp256r1 signature for WebAuthn
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	msg := []byte("test message for webauthn")
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	r1Sig := sig.(*Secp256r1Signature)

	// Create valid mock authenticator data (minimum 37 bytes: rpIdHash(32) + flags(1) + signCount(4))
	authenticatorData := []byte{
		// rpIdHash (32 bytes)
		73, 150, 13, 229, 136, 14, 140, 104, 116, 52, 23, 15, 100, 118, 96, 91,
		143, 228, 174, 185, 162, 134, 50, 199, 153, 92, 243, 186, 131, 29, 151, 99,
		// flags (1 byte)
		29,
		// signCount (4 bytes)
		0, 0, 0, 0,
	}

	// Create mock client data JSON with a valid 32-byte challenge (base64url encoded)
	// "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" decodes to 32 zero bytes
	clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","origin":"http://localhost"}`)

	// Create the assertion response
	paar := NewPartialAuthenticatorAssertionResponse(r1Sig, authenticatorData, clientDataJSON)

	// Wrap in AnySignature
	anySig := paar.ToAnySignature()
	assert.Equal(t, AnySignatureVariantWebAuthn, anySig.Variant)

	// Serialize
	data, err := bcs.Serialize(anySig)
	require.NoError(t, err)

	// Deserialize
	anySig2 := &AnySignature{}
	err = bcs.Deserialize(anySig2, data)
	require.NoError(t, err)

	assert.Equal(t, AnySignatureVariantWebAuthn, anySig2.Variant)
	paar2, ok := anySig2.Signature.(*PartialAuthenticatorAssertionResponse)
	require.True(t, ok)
	assert.Equal(t, authenticatorData, paar2.AuthenticatorData)
	assert.Equal(t, clientDataJSON, paar2.ClientDataJSON)
}

func TestWebAuthn_GetChallenge(t *testing.T) {
	// Create a signature with known challenge
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	sig, _ := key.SignMessage([]byte("test"))
	r1Sig := sig.(*Secp256r1Signature)

	// Valid authenticator data (37 bytes minimum)
	authenticatorData := []byte{
		// rpIdHash (32 bytes)
		73, 150, 13, 229, 136, 14, 140, 104, 116, 52, 23, 15, 100, 118, 96, 91,
		143, 228, 174, 185, 162, 134, 50, 199, 153, 92, 243, 186, 131, 29, 151, 99,
		// flags (1 byte)
		29,
		// signCount (4 bytes)
		0, 0, 0, 0,
	}

	// Test base64url encoded challenge - must be exactly 32 bytes when decoded
	// Using hex: 0102030405060708091011121314151617181920212223242526272829303132
	// base64url: AQIDBAUGBwgJEBESExQVFhcYGSAhIiMkJSYnKCkwMTI
	expectedChallenge := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
		0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24,
		0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32,
	}
	clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"AQIDBAUGBwgJEBESExQVFhcYGSAhIiMkJSYnKCkwMTI","origin":"http://localhost"}`)

	paar := NewPartialAuthenticatorAssertionResponse(r1Sig, authenticatorData, clientDataJSON)

	challenge, err := paar.GetChallenge()
	require.NoError(t, err)
	assert.Equal(t, expectedChallenge, challenge)
}

func TestWebAuthn_ToHex_FromHex(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	sig, _ := key.SignMessage([]byte("test"))
	r1Sig := sig.(*Secp256r1Signature)

	// Valid authenticator data (37 bytes minimum)
	authenticatorData := []byte{
		// rpIdHash (32 bytes)
		73, 150, 13, 229, 136, 14, 140, 104, 116, 52, 23, 15, 100, 118, 96, 91,
		143, 228, 174, 185, 162, 134, 50, 199, 153, 92, 243, 186, 131, 29, 151, 99,
		// flags (1 byte)
		29,
		// signCount (4 bytes)
		0, 0, 0, 0,
	}
	// Valid 32-byte challenge (base64url encoded)
	clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","origin":"http://localhost"}`)

	paar := NewPartialAuthenticatorAssertionResponse(r1Sig, authenticatorData, clientDataJSON)

	// Convert to hex
	hex := paar.ToHex()
	assert.True(t, len(hex) > 0)
	assert.Equal(t, "0x", hex[:2])

	// Convert back from hex
	paar2 := &PartialAuthenticatorAssertionResponse{}
	err = paar2.FromHex(hex)
	require.NoError(t, err)

	assert.Equal(t, paar.AuthenticatorData, paar2.AuthenticatorData)
	assert.Equal(t, paar.ClientDataJSON, paar2.ClientDataJSON)
}

func TestWebAuthn_BoundsValidation(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	sig, _ := key.SignMessage([]byte("test"))
	r1Sig := sig.(*Secp256r1Signature)

	// Test authenticator data too short (less than 37 bytes)
	t.Run("authenticator_data_too_short", func(t *testing.T) {
		shortAuthData := make([]byte, 36) // One byte less than minimum
		clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","origin":"http://localhost"}`)
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, shortAuthData, clientDataJSON)

		data, _ := bcs.Serialize(paar)
		paar2 := &PartialAuthenticatorAssertionResponse{}
		err := bcs.Deserialize(paar2, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authenticator data too short")
	})

	// Test authenticator data too large
	t.Run("authenticator_data_too_large", func(t *testing.T) {
		largeAuthData := make([]byte, MaxAuthenticatorDataBytes+1)
		clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","origin":"http://localhost"}`)
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, largeAuthData, clientDataJSON)

		data, _ := bcs.Serialize(paar)
		paar2 := &PartialAuthenticatorAssertionResponse{}
		err := bcs.Deserialize(paar2, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authenticator data too large")
	})

	// Test client data JSON too large
	t.Run("client_data_json_too_large", func(t *testing.T) {
		validAuthData := make([]byte, 37)
		largeClientData := make([]byte, MaxClientDataJSONBytes+1)
		copy(largeClientData, []byte(`{"type":"webauthn.get","challenge":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","origin":"http://localhost"}`))
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, validAuthData, largeClientData)

		data, _ := bcs.Serialize(paar)
		paar2 := &PartialAuthenticatorAssertionResponse{}
		err := bcs.Deserialize(paar2, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client data JSON too large")
	})

	// Test empty client data JSON
	t.Run("client_data_json_empty", func(t *testing.T) {
		validAuthData := make([]byte, 37)
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, validAuthData, []byte{})

		data, _ := bcs.Serialize(paar)
		paar2 := &PartialAuthenticatorAssertionResponse{}
		err := bcs.Deserialize(paar2, data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client data JSON is empty")
	})
}

func TestWebAuthn_InvalidChallenge(t *testing.T) {
	key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	sig, _ := key.SignMessage([]byte("test"))
	r1Sig := sig.(*Secp256r1Signature)

	validAuthData := []byte{
		73, 150, 13, 229, 136, 14, 140, 104, 116, 52, 23, 15, 100, 118, 96, 91,
		143, 228, 174, 185, 162, 134, 50, 199, 153, 92, 243, 186, 131, 29, 151, 99,
		29, 0, 0, 0, 0,
	}

	// Test challenge with wrong length (not 32 bytes)
	t.Run("challenge_wrong_length", func(t *testing.T) {
		// "YWJj" = "abc" (3 bytes)
		clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"YWJj","origin":"http://localhost"}`)
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, validAuthData, clientDataJSON)

		_, err := paar.GetChallenge()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid challenge length")
	})

	// Test invalid base64 in challenge
	t.Run("challenge_invalid_base64", func(t *testing.T) {
		clientDataJSON := []byte(`{"type":"webauthn.get","challenge":"!!!invalid!!!","origin":"http://localhost"}`)
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, validAuthData, clientDataJSON)

		_, err := paar.GetChallenge()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})

	// Test malformed JSON
	t.Run("malformed_json", func(t *testing.T) {
		clientDataJSON := []byte(`{not valid json}`)
		paar := NewPartialAuthenticatorAssertionResponse(r1Sig, validAuthData, clientDataJSON)

		_, err := paar.GetChallenge()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse client data JSON")
	})
}

// ============================================================================
// Keyless Authentication Tests
// ============================================================================

func TestKeyless_IdCommitment_BCSRoundTrip(t *testing.T) {
	// Create a 32-byte identity commitment
	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}

	idc, err := NewIdCommitment(idcBytes)
	require.NoError(t, err)
	assert.Equal(t, idcBytes, idc.Bytes())

	// Serialize
	data, err := bcs.Serialize(idc)
	require.NoError(t, err)

	// Deserialize
	idc2 := &IdCommitment{}
	err = bcs.Deserialize(idc2, data)
	require.NoError(t, err)

	assert.Equal(t, idc.Bytes(), idc2.Bytes())
}

func TestKeyless_IdCommitment_InvalidLength(t *testing.T) {
	// Too short
	_, err := NewIdCommitment(make([]byte, 31))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid identity commitment length")

	// Too long
	_, err = NewIdCommitment(make([]byte, 33))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid identity commitment length")
}

func TestKeyless_Pepper_BCSRoundTrip(t *testing.T) {
	var pepper Pepper
	for i := range pepper {
		pepper[i] = byte(i)
	}

	// Serialize
	data, err := bcs.Serialize(&pepper)
	require.NoError(t, err)

	// Deserialize
	var pepper2 Pepper
	err = bcs.Deserialize(&pepper2, data)
	require.NoError(t, err)

	assert.Equal(t, pepper, pepper2)
}

func TestKeyless_KeylessPublicKey_BCSRoundTrip(t *testing.T) {
	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}
	idc, _ := NewIdCommitment(idcBytes)

	pk := &KeylessPublicKey{
		IssVal: "https://accounts.google.com",
		Idc:    *idc,
	}

	// Serialize
	data, err := bcs.Serialize(pk)
	require.NoError(t, err)

	// Deserialize
	pk2 := &KeylessPublicKey{}
	err = bcs.Deserialize(pk2, data)
	require.NoError(t, err)

	assert.Equal(t, pk.IssVal, pk2.IssVal)
	assert.Equal(t, pk.Idc.Bytes(), pk2.Idc.Bytes())
}

func TestKeyless_KeylessPublicKey_HexRoundTrip(t *testing.T) {
	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}
	idc, _ := NewIdCommitment(idcBytes)

	pk := &KeylessPublicKey{
		IssVal: "https://accounts.google.com",
		Idc:    *idc,
	}

	// To hex
	hex := pk.ToHex()
	assert.True(t, len(hex) > 0)
	assert.Equal(t, "0x", hex[:2])

	// From hex
	pk2 := &KeylessPublicKey{}
	err := pk2.FromHex(hex)
	require.NoError(t, err)

	assert.Equal(t, pk.IssVal, pk2.IssVal)
	assert.Equal(t, pk.Idc.Bytes(), pk2.Idc.Bytes())
}

func TestKeyless_FederatedKeylessPublicKey_BCSRoundTrip(t *testing.T) {
	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}
	idc, _ := NewIdCommitment(idcBytes)

	var jwkAddr types.AccountAddress
	jwkAddr[31] = 0x01 // Set to address 0x1

	fedPk := &FederatedKeylessPublicKey{
		JwkAddr: jwkAddr,
		Pk: KeylessPublicKey{
			IssVal: "https://accounts.google.com",
			Idc:    *idc,
		},
	}

	// Serialize
	data, err := bcs.Serialize(fedPk)
	require.NoError(t, err)

	// Deserialize
	fedPk2 := &FederatedKeylessPublicKey{}
	err = bcs.Deserialize(fedPk2, data)
	require.NoError(t, err)

	assert.Equal(t, fedPk.JwkAddr, fedPk2.JwkAddr)
	assert.Equal(t, fedPk.Pk.IssVal, fedPk2.Pk.IssVal)
	assert.Equal(t, fedPk.Pk.Idc.Bytes(), fedPk2.Pk.Idc.Bytes())
}

func TestKeyless_AnyPublicKey_KeylessVariant_BCSRoundTrip(t *testing.T) {
	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}
	idc, _ := NewIdCommitment(idcBytes)

	keylessPk := &KeylessPublicKey{
		IssVal: "https://accounts.google.com",
		Idc:    *idc,
	}

	anyPk := &AnyPublicKey{
		Variant: AnyPublicKeyVariantKeyless,
		PubKey:  keylessPk,
	}

	// Serialize
	data, err := bcs.Serialize(anyPk)
	require.NoError(t, err)

	// Deserialize
	anyPk2 := &AnyPublicKey{}
	err = bcs.Deserialize(anyPk2, data)
	require.NoError(t, err)

	assert.Equal(t, AnyPublicKeyVariantKeyless, anyPk2.Variant)
	keylessPk2, ok := anyPk2.PubKey.(*KeylessPublicKey)
	require.True(t, ok)
	assert.Equal(t, keylessPk.IssVal, keylessPk2.IssVal)
}

func TestKeyless_AnyPublicKey_FederatedKeylessVariant_BCSRoundTrip(t *testing.T) {
	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}
	idc, _ := NewIdCommitment(idcBytes)

	var jwkAddr types.AccountAddress
	jwkAddr[31] = 0x01

	fedPk := &FederatedKeylessPublicKey{
		JwkAddr: jwkAddr,
		Pk: KeylessPublicKey{
			IssVal: "https://accounts.google.com",
			Idc:    *idc,
		},
	}

	anyPk := &AnyPublicKey{
		Variant: AnyPublicKeyVariantFederatedKeyless,
		PubKey:  fedPk,
	}

	// Serialize
	data, err := bcs.Serialize(anyPk)
	require.NoError(t, err)

	// Deserialize
	anyPk2 := &AnyPublicKey{}
	err = bcs.Deserialize(anyPk2, data)
	require.NoError(t, err)

	assert.Equal(t, AnyPublicKeyVariantFederatedKeyless, anyPk2.Variant)
	fedPk2, ok := anyPk2.PubKey.(*FederatedKeylessPublicKey)
	require.True(t, ok)
	assert.Equal(t, fedPk.JwkAddr, fedPk2.JwkAddr)
	assert.Equal(t, fedPk.Pk.IssVal, fedPk2.Pk.IssVal)
}

func TestKeyless_Groth16Proof_BCSRoundTrip(t *testing.T) {
	var a G1Bytes
	var b G2Bytes
	var c G1Bytes

	// Fill with test data
	for i := range a {
		a[i] = byte(i)
	}
	for i := range b {
		b[i] = byte(i + 32)
	}
	for i := range c {
		c[i] = byte(i + 96)
	}

	proof := &Groth16Proof{A: a, B: b, C: c}

	// Serialize
	data, err := bcs.Serialize(proof)
	require.NoError(t, err)

	// Deserialize
	proof2 := &Groth16Proof{}
	err = bcs.Deserialize(proof2, data)
	require.NoError(t, err)

	assert.Equal(t, proof.A, proof2.A)
	assert.Equal(t, proof.B, proof2.B)
	assert.Equal(t, proof.C, proof2.C)
}

func TestKeyless_ZeroKnowledgeSig_BCSRoundTrip(t *testing.T) {
	var a G1Bytes
	var b G2Bytes
	var c G1Bytes

	zkSig := &ZeroKnowledgeSig{
		Proof: ZKP{
			Variant: ZKPVariantGroth16,
			Proof:   &Groth16Proof{A: a, B: b, C: c},
		},
		ExpHorizonSecs:          86400,
		ExtraField:              nil,
		OverrideAudVal:          nil,
		TrainingWheelsSignature: nil,
	}

	// Serialize
	data, err := bcs.Serialize(zkSig)
	require.NoError(t, err)

	// Deserialize
	zkSig2 := &ZeroKnowledgeSig{}
	err = bcs.Deserialize(zkSig2, data)
	require.NoError(t, err)

	assert.Equal(t, zkSig.ExpHorizonSecs, zkSig2.ExpHorizonSecs)
	assert.Nil(t, zkSig2.ExtraField)
	assert.Nil(t, zkSig2.OverrideAudVal)
}

func TestKeyless_ZeroKnowledgeSig_WithOptionalFields_BCSRoundTrip(t *testing.T) {
	var a G1Bytes
	var b G2Bytes
	var c G1Bytes

	extraField := `"key":"value"`
	overrideAud := "custom-aud"

	zkSig := &ZeroKnowledgeSig{
		Proof: ZKP{
			Variant: ZKPVariantGroth16,
			Proof:   &Groth16Proof{A: a, B: b, C: c},
		},
		ExpHorizonSecs:          86400,
		ExtraField:              &extraField,
		OverrideAudVal:          &overrideAud,
		TrainingWheelsSignature: nil,
	}

	// Serialize
	data, err := bcs.Serialize(zkSig)
	require.NoError(t, err)

	// Deserialize
	zkSig2 := &ZeroKnowledgeSig{}
	err = bcs.Deserialize(zkSig2, data)
	require.NoError(t, err)

	require.NotNil(t, zkSig2.ExtraField)
	assert.Equal(t, extraField, *zkSig2.ExtraField)
	require.NotNil(t, zkSig2.OverrideAudVal)
	assert.Equal(t, overrideAud, *zkSig2.OverrideAudVal)
}

func TestKeyless_OpenIdSig_BCSRoundTrip(t *testing.T) {
	var pepper Pepper
	for i := range pepper {
		pepper[i] = byte(i)
	}

	openIdSig := &OpenIdSig{
		JwtSig:         []byte("signature-bytes"),
		JwtPayloadJSON: `{"iss":"https://accounts.google.com","sub":"12345","aud":"app-id","nonce":"abc123"}`,
		UidKey:         "sub",
		EpkBlinder:     make([]byte, EpkBlinderNumBytes),
		Pepper:         pepper,
		IdcAudVal:      nil,
	}

	// Serialize
	data, err := bcs.Serialize(openIdSig)
	require.NoError(t, err)

	// Deserialize
	openIdSig2 := &OpenIdSig{}
	err = bcs.Deserialize(openIdSig2, data)
	require.NoError(t, err)

	assert.Equal(t, openIdSig.JwtSig, openIdSig2.JwtSig)
	assert.Equal(t, openIdSig.JwtPayloadJSON, openIdSig2.JwtPayloadJSON)
	assert.Equal(t, openIdSig.UidKey, openIdSig2.UidKey)
	assert.Equal(t, openIdSig.Pepper, openIdSig2.Pepper)
}

func TestKeyless_EphemeralPublicKey_Ed25519_BCSRoundTrip(t *testing.T) {
	ed25519Key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	ed25519PubKey := ed25519Key.PubKey().(*Ed25519PublicKey)

	epk := &EphemeralPublicKey{
		Variant: EphemeralPublicKeyVariantEd25519,
		PubKey:  ed25519PubKey,
	}

	// Serialize
	data, err := bcs.Serialize(epk)
	require.NoError(t, err)

	// Deserialize
	epk2 := &EphemeralPublicKey{}
	err = bcs.Deserialize(epk2, data)
	require.NoError(t, err)

	assert.Equal(t, EphemeralPublicKeyVariantEd25519, epk2.Variant)
	ed25519Pub2, ok := epk2.PubKey.(*Ed25519PublicKey)
	require.True(t, ok)
	assert.Equal(t, ed25519PubKey.Bytes(), ed25519Pub2.Bytes())
}

func TestKeyless_EphemeralPublicKey_Secp256r1_BCSRoundTrip(t *testing.T) {
	secp256r1Key, err := GenerateSecp256r1Key()
	require.NoError(t, err)

	secp256r1PubKey := secp256r1Key.VerifyingKey().(*Secp256r1PublicKey)

	epk := &EphemeralPublicKey{
		Variant: EphemeralPublicKeyVariantSecp256r1,
		PubKey:  secp256r1PubKey,
	}

	// Serialize
	data, err := bcs.Serialize(epk)
	require.NoError(t, err)

	// Deserialize
	epk2 := &EphemeralPublicKey{}
	err = bcs.Deserialize(epk2, data)
	require.NoError(t, err)

	assert.Equal(t, EphemeralPublicKeyVariantSecp256r1, epk2.Variant)
	secp256r1Pub2, ok := epk2.PubKey.(*Secp256r1PublicKey)
	require.True(t, ok)
	assert.Equal(t, secp256r1PubKey.Bytes(), secp256r1Pub2.Bytes())
}

func TestKeyless_KeylessSignature_BCSRoundTrip(t *testing.T) {
	// Create a minimal keyless signature with OpenIdSig
	var pepper Pepper
	for i := range pepper {
		pepper[i] = byte(i)
	}

	ed25519Key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	ed25519PubKey := ed25519Key.PubKey().(*Ed25519PublicKey)

	msg := []byte("test message")
	ed25519Sig, err := ed25519Key.SignMessage(msg)
	require.NoError(t, err)

	keylessSig := &KeylessSignature{
		Cert: EphemeralCertificate{
			Variant: EphemeralCertificateVariantOpenId,
			Cert: &OpenIdSig{
				JwtSig:         []byte("signature-bytes"),
				JwtPayloadJSON: `{"iss":"https://accounts.google.com","sub":"12345"}`,
				UidKey:         "sub",
				EpkBlinder:     make([]byte, EpkBlinderNumBytes),
				Pepper:         pepper,
				IdcAudVal:      nil,
			},
		},
		JwtHeaderJSON: `{"alg":"RS256","kid":"key-id-1"}`,
		ExpDateSecs:   1700000000,
		EphemeralPubkey: EphemeralPublicKey{
			Variant: EphemeralPublicKeyVariantEd25519,
			PubKey:  ed25519PubKey,
		},
		EphemeralSignature: EphemeralSignature{
			Variant:   EphemeralSignatureVariantEd25519,
			Signature: ed25519Sig,
		},
	}

	// Serialize
	data, err := bcs.Serialize(keylessSig)
	require.NoError(t, err)

	// Deserialize
	keylessSig2 := &KeylessSignature{}
	err = bcs.Deserialize(keylessSig2, data)
	require.NoError(t, err)

	assert.Equal(t, keylessSig.JwtHeaderJSON, keylessSig2.JwtHeaderJSON)
	assert.Equal(t, keylessSig.ExpDateSecs, keylessSig2.ExpDateSecs)
	assert.Equal(t, EphemeralCertificateVariantOpenId, keylessSig2.Cert.Variant)
}

func TestKeyless_AnySignature_KeylessVariant_BCSRoundTrip(t *testing.T) {
	var pepper Pepper

	ed25519Key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	ed25519PubKey := ed25519Key.PubKey().(*Ed25519PublicKey)

	msg := []byte("test message")
	ed25519Sig, err := ed25519Key.SignMessage(msg)
	require.NoError(t, err)

	keylessSig := &KeylessSignature{
		Cert: EphemeralCertificate{
			Variant: EphemeralCertificateVariantOpenId,
			Cert: &OpenIdSig{
				JwtSig:         []byte("sig"),
				JwtPayloadJSON: `{}`,
				UidKey:         "sub",
				EpkBlinder:     make([]byte, EpkBlinderNumBytes),
				Pepper:         pepper,
				IdcAudVal:      nil,
			},
		},
		JwtHeaderJSON: `{}`,
		ExpDateSecs:   1700000000,
		EphemeralPubkey: EphemeralPublicKey{
			Variant: EphemeralPublicKeyVariantEd25519,
			PubKey:  ed25519PubKey,
		},
		EphemeralSignature: EphemeralSignature{
			Variant:   EphemeralSignatureVariantEd25519,
			Signature: ed25519Sig,
		},
	}

	anySig := keylessSig.ToAnySignature()
	assert.Equal(t, AnySignatureVariantKeyless, anySig.Variant)

	// Serialize
	data, err := bcs.Serialize(anySig)
	require.NoError(t, err)

	// Deserialize
	anySig2 := &AnySignature{}
	err = bcs.Deserialize(anySig2, data)
	require.NoError(t, err)

	assert.Equal(t, AnySignatureVariantKeyless, anySig2.Variant)
	keylessSig2, ok := anySig2.Signature.(*KeylessSignature)
	require.True(t, ok)
	assert.Equal(t, keylessSig.ExpDateSecs, keylessSig2.ExpDateSecs)
}
