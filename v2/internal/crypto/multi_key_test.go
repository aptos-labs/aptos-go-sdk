package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiKeyBitmap_NewAndContainsKey(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(5)
	assert.False(t, bm.ContainsKey(0))
	assert.False(t, bm.ContainsKey(4))
}

func TestMultiKeyBitmap_AddKey_Valid(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(8)
	err := bm.AddKey(0)
	require.NoError(t, err)
	assert.True(t, bm.ContainsKey(0))

	err = bm.AddKey(3)
	require.NoError(t, err)
	assert.True(t, bm.ContainsKey(3))
}

func TestMultiKeyBitmap_AddKey_Duplicate(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(8)
	require.NoError(t, bm.AddKey(0))
	err := bm.AddKey(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in bitmap")
}

func TestMultiKeyBitmap_AddKey_OutOfRange(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(8)
	err := bm.AddKey(MaxMultiKeySignatures)
	assert.Error(t, err)
}

func TestMultiKeyBitmap_ContainsKey_OutOfRange(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(4)
	assert.False(t, bm.ContainsKey(MaxMultiKeySignatures))
}

func TestMultiKeyBitmap_Indices_Order(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(10)
	require.NoError(t, bm.AddKey(1))
	require.NoError(t, bm.AddKey(3))
	require.NoError(t, bm.AddKey(7))

	indices := bm.Indices()
	assert.Equal(t, []uint8{1, 3, 7}, indices)
}

func TestMultiKeyBitmap_BCSRoundTrip_WithNewBitmap(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(10)
	require.NoError(t, bm.AddKey(0))
	require.NoError(t, bm.AddKey(5))
	require.NoError(t, bm.AddKey(9))

	data, err := bcs.Serialize(bm)
	require.NoError(t, err)

	var bm2 MultiKeyBitmap
	err = bcs.Deserialize(&bm2, data)
	require.NoError(t, err)

	assert.Equal(t, bm.Indices(), bm2.Indices())
}

func TestMultiKey_Scheme_Value(t *testing.T) {
	t.Parallel()

	mk := &MultiKey{}
	assert.Equal(t, MultiKeyScheme, mk.Scheme())
}

func TestMultiKey_AuthKey_NotNil(t *testing.T) {
	t.Parallel()

	edKey := &Ed25519PublicKey{}
	anyKey, _ := ToAnyPublicKey(edKey)
	mk := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyKey},
		SignaturesRequired: 1,
	}
	assert.NotNil(t, mk.AuthKey())
}

func TestMultiKey_BytesFromBytes_RoundTrip(t *testing.T) {
	t.Parallel()

	privKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	anyKey, err := ToAnyPublicKey(privKey.PubKey())
	require.NoError(t, err)
	mk := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyKey},
		SignaturesRequired: 1,
	}

	data := mk.Bytes()
	assert.NotEmpty(t, data)

	var mk2 MultiKey
	err = mk2.FromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, mk.SignaturesRequired, mk2.SignaturesRequired)
}

func TestMultiKey_ToHexFromHex_RoundTrip(t *testing.T) {
	t.Parallel()

	privKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	anyKey, err := ToAnyPublicKey(privKey.PubKey())
	require.NoError(t, err)
	mk := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyKey},
		SignaturesRequired: 1,
	}

	hex := mk.ToHex()
	assert.NotEmpty(t, hex)

	var mk2 MultiKey
	err = mk2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, mk.SignaturesRequired, mk2.SignaturesRequired)
}

func TestMultiKey_FromHex_InvalidHex(t *testing.T) {
	t.Parallel()

	var mk MultiKey
	err := mk.FromHex("0xzzzz")
	assert.Error(t, err)
}

func TestMultiKey_Verify_WrongType(t *testing.T) {
	t.Parallel()

	edKey := &Ed25519PublicKey{}
	anyKey, _ := ToAnyPublicKey(edKey)
	mk := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyKey},
		SignaturesRequired: 1,
	}

	// Wrong signature type
	assert.False(t, mk.Verify([]byte("msg"), &Ed25519Signature{}))
}

func TestNewMultiKeySignature_Valid(t *testing.T) {
	t.Parallel()

	edSig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: &Ed25519Signature{},
	}
	sig, err := NewMultiKeySignature(3, []IndexedAnySignature{
		{Index: 0, Signature: edSig},
		{Index: 2, Signature: edSig},
	})
	require.NoError(t, err)
	assert.Len(t, sig.Signatures, 2)
	assert.Equal(t, []uint8{0, 2}, sig.Bitmap.Indices())
}

func TestNewMultiKeySignature_OutOfBounds(t *testing.T) {
	t.Parallel()

	edSig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: &Ed25519Signature{},
	}
	_, err := NewMultiKeySignature(3, []IndexedAnySignature{
		{Index: 5, Signature: edSig},
	})
	assert.Error(t, err)
}

func TestNewMultiKeySignature_ZeroNumKeys(t *testing.T) {
	t.Parallel()

	_, err := NewMultiKeySignature(0, nil)
	assert.Error(t, err)
}

func TestNewMultiKeySignature_ExceedsMax(t *testing.T) {
	t.Parallel()

	_, err := NewMultiKeySignature(MaxMultiKeySignatures+1, nil)
	assert.Error(t, err)
}

func TestMultiKeySignature_BytesAndHex_RoundTrip(t *testing.T) {
	t.Parallel()

	edSig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: &Ed25519Signature{},
	}
	sig, err := NewMultiKeySignature(2, []IndexedAnySignature{
		{Index: 0, Signature: edSig},
	})
	require.NoError(t, err)

	data := sig.Bytes()
	assert.NotEmpty(t, data)

	var sig2 MultiKeySignature
	err = sig2.FromBytes(data)
	require.NoError(t, err)

	hex := sig.ToHex()
	var sig3 MultiKeySignature
	err = sig3.FromHex(hex)
	require.NoError(t, err)
}

func TestMultiKeySignature_FromHex_InvalidHex(t *testing.T) {
	t.Parallel()

	var sig MultiKeySignature
	err := sig.FromHex("0xzzzz")
	assert.Error(t, err)
}

func TestMultiKeyBitmap_SetNumKeys_Expand(t *testing.T) {
	t.Parallel()

	bm := NewMultiKeyBitmap(4)
	require.NoError(t, bm.AddKey(0))

	bm.SetNumKeys(16) // Expand
	assert.True(t, bm.ContainsKey(0))
	assert.Len(t, bm.inner, 2) // 16 bits = 2 bytes
}

func TestIndexedAnySignature_BCSRoundTrip(t *testing.T) {
	t.Parallel()

	sig := &IndexedAnySignature{
		Index: 2,
		Signature: &AnySignature{
			Variant:   AnySignatureVariantEd25519,
			Signature: &Ed25519Signature{},
		},
	}

	data, err := bcs.Serialize(sig)
	require.NoError(t, err)

	var sig2 IndexedAnySignature
	err = bcs.Deserialize(&sig2, data)
	require.NoError(t, err)
	assert.Equal(t, uint8(2), sig2.Index)
}

func TestNoAuthenticator_BCS(t *testing.T) {
	t.Parallel()

	auth := &NoAuthenticator{}
	ser := bcs.NewSerializer()
	auth.MarshalBCS(ser)
	assert.NoError(t, ser.Error())

	des := bcs.NewDeserializer(nil)
	auth.UnmarshalBCS(des)
}

func TestMultiKeyAuthenticator_BCSRoundTrip_WithGenKey(t *testing.T) {
	t.Parallel()

	privKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey().(*Ed25519PublicKey)
	anyPubKey, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)

	edSig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: &Ed25519Signature{},
	}
	mkSig, err := NewMultiKeySignature(1, []IndexedAnySignature{
		{Index: 0, Signature: edSig},
	})
	require.NoError(t, err)

	auth := &MultiKeyAuthenticator{
		PubKey: &MultiKey{
			PubKeys:            []*AnyPublicKey{anyPubKey},
			SignaturesRequired: 1,
		},
		Sig: mkSig,
	}

	data, err := bcs.Serialize(auth)
	require.NoError(t, err)

	var auth2 MultiKeyAuthenticator
	err = bcs.Deserialize(&auth2, data)
	require.NoError(t, err)
	assert.Equal(t, uint8(1), auth2.PubKey.SignaturesRequired)
}

func TestMultiKeyAuthenticator_Verify_WithRealSig(t *testing.T) {
	t.Parallel()

	privKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey().(*Ed25519PublicKey)
	anyPubKey, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := privKey.SignMessage(msg)
	require.NoError(t, err)

	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: sig.(*Ed25519Signature),
	}

	mkSig, err := NewMultiKeySignature(1, []IndexedAnySignature{
		{Index: 0, Signature: anySig},
	})
	require.NoError(t, err)

	auth := &MultiKeyAuthenticator{
		PubKey: &MultiKey{
			PubKeys:            []*AnyPublicKey{anyPubKey},
			SignaturesRequired: 1,
		},
		Sig: mkSig,
	}

	assert.True(t, auth.Verify(msg))
	assert.False(t, auth.Verify([]byte("wrong")))
}

func TestMultiKeySignature_Verify_WithRealSig(t *testing.T) {
	t.Parallel()

	privKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey().(*Ed25519PublicKey)
	anyPubKey, err := ToAnyPublicKey(pubKey)
	require.NoError(t, err)

	msg := []byte("test message")
	sig, err := privKey.SignMessage(msg)
	require.NoError(t, err)
	anySig := &AnySignature{
		Variant:   AnySignatureVariantEd25519,
		Signature: sig.(*Ed25519Signature),
	}

	mkSig, err := NewMultiKeySignature(1, []IndexedAnySignature{
		{Index: 0, Signature: anySig},
	})
	require.NoError(t, err)

	mk := &MultiKey{
		PubKeys:            []*AnyPublicKey{anyPubKey},
		SignaturesRequired: 1,
	}

	assert.True(t, mk.Verify(msg, mkSig))
	assert.False(t, mk.Verify([]byte("wrong"), mkSig))
}
