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
