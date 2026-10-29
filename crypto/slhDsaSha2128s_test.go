package crypto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto/hd"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported from aptos-ts-sdk#802 (tests/unit/slhDsaSha2128s.test.ts and the
// cross-check against @noble/post-quantum): the 48-byte seed triple
// 0x00..0x2f must derive this public key.
const (
	testSlhDsaSeedsHex   = "0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f"
	testSlhDsaPublicHex  = "0x202122232425262728292a2b2c2d2e2f990ce6298792b128846a8e4a3a68954c"
	testSlhDsaMessageHex = "hello aptos"
)

func testSlhDsaPrivateKey(t *testing.T) *SlhDsaSha2128sPrivateKey {
	t.Helper()
	key := &SlhDsaSha2128sPrivateKey{}
	require.NoError(t, key.FromHex(testSlhDsaSeedsHex))
	return key
}

func TestSlhDsaSha2128sKeyGeneration(t *testing.T) {
	t.Parallel()
	key, err := GenerateSlhDsaSha2128sPrivateKey()
	require.NoError(t, err)
	assert.Len(t, key.Bytes(), SlhDsaSha2128sPrivateKeyLength)
	assert.Len(t, key.PublicKey().Bytes(), SlhDsaSha2128sPublicKeyLength)

	key2, err := GenerateSlhDsaSha2128sPrivateKey()
	require.NoError(t, err)
	assert.NotEqual(t, key.Bytes(), key2.Bytes())
}

func TestSlhDsaSha2128sDeterministicReader(t *testing.T) {
	t.Parallel()
	seeds, err := util.ParseHex(testSlhDsaSeedsHex)
	require.NoError(t, err)
	key, err := GenerateSlhDsaSha2128sPrivateKey(bytes.NewReader(seeds))
	require.NoError(t, err)
	assert.Equal(t, seeds, key.Bytes())
	// Public key must match the @noble/post-quantum keygen for the same seeds.
	assert.Equal(t, testSlhDsaPublicHex, key.PublicKey().ToHex())
}

func TestSlhDsaSha2128sPrivateKeyRoundTrip(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	assert.Equal(t, testSlhDsaSeedsHex, key.ToHex())

	aip80, err := key.ToAIP80()
	require.NoError(t, err)
	assert.Equal(t, "slh-dsa-sha2-128s-priv-"+testSlhDsaSeedsHex, aip80)

	fromAIP80 := &SlhDsaSha2128sPrivateKey{}
	require.NoError(t, fromAIP80.FromHex(aip80))
	assert.Equal(t, key.Bytes(), fromAIP80.Bytes())
	assert.Equal(t, key.PublicKey().Bytes(), fromAIP80.PublicKey().Bytes())

	assert.Equal(t, "<redacted SlhDsaSha2128sPrivateKey>", key.String())
}

func TestSlhDsaSha2128sPrivateKeyFullSecretKeyImport(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	full, err := key.Inner.MarshalBinary()
	require.NoError(t, err)
	require.Len(t, full, SlhDsaSha2128sSecretKeyLength)

	fromFull := &SlhDsaSha2128sPrivateKey{}
	require.NoError(t, fromFull.FromBytes(full))
	assert.Equal(t, key.Bytes(), fromFull.Bytes())

	bad := make([]byte, len(full))
	copy(bad, full)
	bad[len(bad)-1] ^= 0xff
	require.Error(t, (&SlhDsaSha2128sPrivateKey{}).FromBytes(bad))
}

func TestSlhDsaSha2128sInvalidLengths(t *testing.T) {
	t.Parallel()
	require.Error(t, (&SlhDsaSha2128sPrivateKey{}).FromBytes(make([]byte, 47)))
	require.Error(t, (&SlhDsaSha2128sPrivateKey{}).FromBytes(make([]byte, 49)))
	require.Error(t, (&SlhDsaSha2128sPublicKey{}).FromBytes(make([]byte, 31)))
	require.Error(t, (&SlhDsaSha2128sPublicKey{}).FromBytes(make([]byte, 33)))
	require.Error(t, (&SlhDsaSha2128sSignature{}).FromBytes(make([]byte, SlhDsaSha2128sSignatureLength-1)))
}

func TestSlhDsaSha2128sSignAndVerify(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	msg := []byte(testSlhDsaMessageHex)

	sig, err := key.SignMessage(msg)
	require.NoError(t, err)
	slhSig, ok := sig.(*SlhDsaSha2128sSignature)
	require.True(t, ok)
	assert.Len(t, slhSig.Bytes(), SlhDsaSha2128sSignatureLength)

	pub := key.PublicKey()
	assert.True(t, pub.Verify(msg, sig))
	assert.False(t, pub.Verify([]byte("wrong message"), sig))
	assert.False(t, pub.Verify(msg, &Ed25519Signature{}))

	// Deterministic signing: same key and message produce the same signature.
	sig2, err := key.SignMessage(msg)
	require.NoError(t, err)
	assert.Equal(t, sig.Bytes(), sig2.Bytes())
}

func TestSlhDsaSha2128sPublicKeyAndSignatureBCS(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	pub := key.PublicKey()

	pubBytes, err := bcs.Serialize(pub)
	require.NoError(t, err)
	// BCS: ULEB128(32) || raw key
	assert.Equal(t, byte(SlhDsaSha2128sPublicKeyLength), pubBytes[0])
	assert.Equal(t, pub.Bytes(), pubBytes[1:])

	pub2 := &SlhDsaSha2128sPublicKey{}
	require.NoError(t, bcs.Deserialize(pub2, pubBytes))
	assert.Equal(t, pub.Bytes(), pub2.Bytes())

	sig, err := key.SignMessage([]byte(testSlhDsaMessageHex))
	require.NoError(t, err)
	sigBytes, err := bcs.Serialize(sig)
	require.NoError(t, err)
	sig2 := &SlhDsaSha2128sSignature{}
	require.NoError(t, bcs.Deserialize(sig2, sigBytes))
	assert.Equal(t, sig.Bytes(), sig2.Bytes())
}

func TestSlhDsaSha2128sAnyKeyBCSAndAuthenticator(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	signer := NewSingleSigner(key)
	signerPubKey, ok := signer.PubKey().(*AnyPublicKey)
	require.True(t, ok)
	assert.Equal(t, AnyPublicKeyVariantSlhDsaSha2128s, signerPubKey.Variant)
	assert.Equal(t, AnySignatureVariantSlhDsaSha2128s, signer.SignatureVariant())

	msg := []byte(testSlhDsaMessageHex)
	auth, err := signer.Sign(msg)
	require.NoError(t, err)
	assert.Equal(t, AccountAuthenticatorSingleSender, auth.Variant)
	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong message")))

	authBytes, err := bcs.Serialize(auth)
	require.NoError(t, err)
	// AccountAuthenticator BCS starts with the SingleSender variant (2);
	// inside, AnyPublicKey starts with variant 5 and AnySignature with 4.
	require.Greater(t, len(authBytes), 2)
	assert.Equal(t, byte(AccountAuthenticatorSingleSender), authBytes[0])
	assert.Equal(t, byte(AnyPublicKeyVariantSlhDsaSha2128s), authBytes[1])

	auth2 := &AccountAuthenticator{}
	require.NoError(t, bcs.Deserialize(auth2, authBytes))
	assert.True(t, auth2.Verify(msg))
	reserialized, err := bcs.Serialize(auth2)
	require.NoError(t, err)
	assert.Equal(t, authBytes, reserialized)

	// Simulation authenticator carries an empty (all-zero) SLH-DSA signature.
	sim := signer.SimulationAuthenticator()
	require.NotNil(t, sim)
	single, ok := sim.Auth.(*SingleKeyAuthenticator)
	require.True(t, ok)
	assert.Equal(t, AnySignatureVariantSlhDsaSha2128s, single.Sig.Variant)
}

func TestSlhDsaSha2128sAuthKeyMatchesSingleSigner(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	signer := NewSingleSigner(key)
	// Standalone public key, SingleSigner and the derived address source must agree.
	assert.Equal(t, signer.AuthKey(), key.PublicKey().AuthKey())
	anyPub, err := ToAnyPublicKey(key.PublicKey())
	require.NoError(t, err)
	assert.Equal(t, signer.AuthKey(), anyPub.AuthKey())
	assert.Equal(t, SingleKeyScheme, key.PublicKey().Scheme())
}

func TestSlhDsaSha2128sMultiKey(t *testing.T) {
	t.Parallel()
	slhKey := testSlhDsaPrivateKey(t)
	edKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	slhAny, err := ToAnyPublicKey(slhKey.PublicKey())
	require.NoError(t, err)
	edPubKey, ok := edKey.PubKey().(*Ed25519PublicKey)
	require.True(t, ok)
	edAny, err := ToAnyPublicKey(edPubKey)
	require.NoError(t, err)

	multi := &MultiKey{PubKeys: []*AnyPublicKey{slhAny, edAny}, SignaturesRequired: 1}
	msg := []byte("multikey slh test")
	sig, err := NewSingleSigner(slhKey).SignMessage(msg)
	require.NoError(t, err)
	anySig, ok := sig.(*AnySignature)
	require.True(t, ok)
	multiSig, err := NewMultiKeySignature([]IndexedAnySignature{{Index: 0, Signature: anySig}})
	require.NoError(t, err)
	assert.True(t, multi.Verify(msg, multiSig))
	assert.False(t, multi.Verify([]byte("wrong"), multiSig))

	// BCS round-trip of the MultiKey must preserve the SLH-DSA sub-key.
	multiBytes, err := bcs.Serialize(multi)
	require.NoError(t, err)
	multi2 := &MultiKey{}
	require.NoError(t, bcs.Deserialize(multi2, multiBytes))
	assert.True(t, multi2.Verify(msg, multiSig))
}

func TestSlhDsaSha2128sAuthenticator(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	msg := []byte(testSlhDsaMessageHex)
	sig, err := key.SignMessage(msg)
	require.NoError(t, err)

	slhSig, ok := sig.(*SlhDsaSha2128sSignature)
	require.True(t, ok)
	auth := &SlhDsaSha2128sAuthenticator{PubKey: key.PublicKey(), Sig: slhSig}
	assert.True(t, auth.Verify(msg))
	serialized, err := bcs.Serialize(auth)
	require.NoError(t, err)
	auth2 := &SlhDsaSha2128sAuthenticator{}
	require.NoError(t, bcs.Deserialize(auth2, serialized))
	assert.True(t, auth2.Verify(msg))
}

// Ported from aptos-ts-sdk#802 tests/unit/hdKey.test.ts (SLIP-0010 vectors with
// the custom "SLH-DSA-SHA2-128s seed" HMAC key, seed 0x000102...0f).
func TestSlhDsaSha2128sDerivationVectorsTS(t *testing.T) {
	t.Parallel()
	seed, err := util.ParseHex("0x000102030405060708090a0b0c0d0e0f")
	require.NoError(t, err)
	vectors := []struct{ path, private string }{
		{"m", "53c14b2681fcca8600d3ac7ce3459d6ceece7abc0a3b4c0376fc845c1b6503d66c5107d9484171c02f9eb0409849fceb"},
		{"m/0'", "473049c70d5414379363b94f52de6b194c9f1651cb6a11b80cfee19f6ca67975201cdbec137ff57e2e9cd434fca0680d"},
		{"m/0'/1'", "b609bf21df9baae65045e2bb8c200e97fbf86201f58e4187197a60bfc827158e6ba452ffe50f57849c022b24d01ac123"},
		{"m/0'/1'/2'", "60b55f07c62916f7f27a96be371a9e87adedb9acabd84d25c41e8aa5c8c326e2e64024132f8833181bc2ab008541ef2b"},
		{"m/0'/1'/2'/2'", "fb7242320fc3bc620587cfdfa54c2abed7d034f6c616ade3a9ad1780a099b268ed7a5bea0297711e997ff199b24bb82b"},
		{"m/0'/1'/2'/2'/1000000000'", "c72d4e3fb352a785aa375be20e0767ae20edbce1e18de90e0ea6e9d1677d17e84a410bb833bf091655dfabde17cc2011"},
	}
	for _, v := range vectors {
		derived, err := hd.DeriveSlhDsaSha2128sPrivateKey(v.path, seed)
		require.NoError(t, err, v.path)
		assert.Equal(t, "0x"+v.private, util.BytesToHex(derived), v.path)
	}

	_, err = hd.DeriveSlhDsaSha2128sPrivateKey("m/0/1'", seed)
	require.Error(t, err, "non-hardened segments must be rejected")
	assert.True(t, hd.IsValidSlhDsaHardenedPath(hd.DefaultDerivationPath))
	assert.False(t, hd.IsValidSlhDsaHardenedPath("m/44'/637'/0'/0'/0"))
	assert.NotContains(t, hd.Slip10SlhDsaSha2128sSeed, "ed25519")
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("no randomness available") }

func TestSlhDsaSha2128sGenerateAliasesAndErrors(t *testing.T) {
	t.Parallel()
	key, err := GenerateSlhDsaPrivateKey()
	require.NoError(t, err)
	assert.Len(t, key.Bytes(), SlhDsaSha2128sPrivateKeyLength)

	_, err = GenerateSlhDsaSha2128sPrivateKey(failingReader{})
	require.Error(t, err)

	_, err = GenerateSlhDsaSha2128sPrivateKey(bytes.NewReader(make([]byte, 10)))
	require.Error(t, err, "a reader with fewer than 48 bytes must fail")

	fromSeeds, err := NewSlhDsaSha2128sPrivateKeyFromSeeds(key.Bytes())
	require.NoError(t, err)
	assert.Equal(t, key.PublicKey().Bytes(), fromSeeds.PublicKey().Bytes())

	_, err = NewSlhDsaSha2128sPrivateKeyFromSeeds(make([]byte, 47))
	require.Error(t, err)
	_, err = NewSlhDsaSha2128sPrivateKeyFromSeeds(nil)
	require.Error(t, err)
}

func TestSlhDsaSha2128sPrivateKeyClearAndFromHexErrors(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	key.Clear()
	assert.Empty(t, key.Bytes())

	require.Error(t, (&SlhDsaSha2128sPrivateKey{}).FromHex("not-a-hex-string"))
	require.Error(t, (&SlhDsaSha2128sPrivateKey{}).FromHex("ed25519-priv-0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f"))
	require.Error(t, (&SlhDsaSha2128sPrivateKey{}).FromHex("slh-dsa-sha2-128s-priv-0x00"))
	// A plain (non-AIP-80) hex string must still parse.
	plain := &SlhDsaSha2128sPrivateKey{}
	require.NoError(t, plain.FromHex(testSlhDsaSeedsHex))
	assert.Equal(t, testSlhDsaPublicHex, plain.PublicKey().ToHex())
}

func TestSlhDsaSha2128sPublicKeyFromHex(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	pub := key.PublicKey()

	fromHex := &SlhDsaSha2128sPublicKey{}
	require.NoError(t, fromHex.FromHex(testSlhDsaPublicHex))
	assert.Equal(t, pub.Bytes(), fromHex.Bytes())
	assert.Equal(t, pub.ToHex(), fromHex.ToHex())
	assert.Equal(t, pub.AuthKey(), fromHex.AuthKey())

	require.Error(t, fromHex.FromHex("0xzz"))
	require.Error(t, fromHex.FromHex("0x00"))
	require.Error(t, fromHex.FromBytes(nil))
}

func TestSlhDsaSha2128sSignatureHexAndAny(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	sigI, err := key.SignMessage([]byte(testSlhDsaMessageHex))
	require.NoError(t, err)
	sig, ok := sigI.(*SlhDsaSha2128sSignature)
	require.True(t, ok)

	fromHex := &SlhDsaSha2128sSignature{}
	require.NoError(t, fromHex.FromHex(sig.ToHex()))
	assert.Equal(t, sig.Bytes(), fromHex.Bytes())

	require.Error(t, fromHex.FromHex("0xzz"))
	require.Error(t, fromHex.FromHex("0x00"))

	anySig := sig.ToAnySignature()
	assert.Equal(t, AnySignatureVariantSlhDsaSha2128s, anySig.Variant)
	assert.Equal(t, sig, anySig.Signature)

	anyBytes, err := bcs.Serialize(anySig)
	require.NoError(t, err)
	anySig2 := &AnySignature{}
	require.NoError(t, bcs.Deserialize(anySig2, anyBytes))
	assert.Equal(t, anySig.Variant, anySig2.Variant)
	assert.Equal(t, sig.Bytes(), anySig2.Signature.Bytes())

	// Unknown variants must be rejected on BCS decode.
	badPub := &AnyPublicKey{}
	assert.Error(t, bcs.Deserialize(badPub, []byte{0x09, 0x00}))
	badSig := &AnySignature{}
	assert.Error(t, bcs.Deserialize(badSig, []byte{0x09, 0x00}))
}

func TestSlhDsaSha2128sBCSErrorPaths(t *testing.T) {
	t.Parallel()
	// Truncated BCS public key (claims 32 bytes, supplies 1).
	pub := &SlhDsaSha2128sPublicKey{}
	assert.Error(t, bcs.Deserialize(pub, []byte{0x20, 0x01}))
	// BCS signature with a wrong-length payload.
	sig := &SlhDsaSha2128sSignature{}
	assert.Error(t, bcs.Deserialize(sig, []byte{0x01, 0x00}))
	// Truncated authenticator.
	auth := &SlhDsaSha2128sAuthenticator{}
	assert.Error(t, bcs.Deserialize(auth, []byte{0x20, 0x01}))
	// Empty input.
	assert.Error(t, bcs.Deserialize(&SlhDsaSha2128sPublicKey{}, []byte{}))
	assert.Error(t, bcs.Deserialize(&SlhDsaSha2128sSignature{}, []byte{}))
}

func TestSlhDsaSha2128sAuthenticatorAccessors(t *testing.T) {
	t.Parallel()
	key := testSlhDsaPrivateKey(t)
	msg := []byte(testSlhDsaMessageHex)
	sigI, err := key.SignMessage(msg)
	require.NoError(t, err)
	slhSig, ok := sigI.(*SlhDsaSha2128sSignature)
	require.True(t, ok)

	auth := &SlhDsaSha2128sAuthenticator{PubKey: key.PublicKey(), Sig: slhSig}
	assert.Equal(t, key.PublicKey(), auth.PublicKey())
	assert.Equal(t, slhSig, auth.Signature())
	assert.True(t, auth.Verify(msg))

	// VerifyingKey() must be the same key as PublicKey().
	assert.Equal(t, key.PublicKey().Bytes(), key.VerifyingKey().Bytes())
	// EmptySignature is an all-zero signature of the full length.
	empty := key.EmptySignature()
	assert.Len(t, empty.Bytes(), SlhDsaSha2128sSignatureLength)
	assert.False(t, key.PublicKey().Verify(msg, empty))
}
