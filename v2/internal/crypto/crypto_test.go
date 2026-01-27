package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
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
	multiSig, err := NewMultiKeySignature([]IndexedAnySignature{
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

	multiSig, err := NewMultiKeySignature([]IndexedAnySignature{
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
