package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeylessPublicKey_Verify(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	idc, _ := NewIdCommitment(idcBytes)
	key := &KeylessPublicKey{IssVal: "https://accounts.google.com", Idc: *idc}
	assert.False(t, key.Verify([]byte("msg"), &Ed25519Signature{}))
}

func TestKeylessPublicKey_AuthKey(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	idc, _ := NewIdCommitment(idcBytes)
	key := &KeylessPublicKey{IssVal: "https://accounts.google.com", Idc: *idc}
	authKey := key.AuthKey()
	assert.NotNil(t, authKey)
	assert.Len(t, authKey.Bytes(), 32)
}

func TestKeylessPublicKey_Scheme(t *testing.T) {
	t.Parallel()

	key := &KeylessPublicKey{}
	assert.Equal(t, SingleKeyScheme, key.Scheme())
}

func TestKeylessPublicKey_BytesAndFromBytes(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	for i := range idcBytes {
		idcBytes[i] = byte(i)
	}
	idc, _ := NewIdCommitment(idcBytes)
	key := &KeylessPublicKey{IssVal: "https://accounts.google.com", Idc: *idc}

	data := key.Bytes()
	assert.NotEmpty(t, data)

	var key2 KeylessPublicKey
	err := key2.FromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, key.IssVal, key2.IssVal)
}

func TestKeylessPublicKey_ToHexAndFromHex(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	idc, _ := NewIdCommitment(idcBytes)
	key := &KeylessPublicKey{IssVal: "https://accounts.google.com", Idc: *idc}

	hex := key.ToHex()
	assert.NotEmpty(t, hex)

	var key2 KeylessPublicKey
	err := key2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, key.IssVal, key2.IssVal)
}

func TestKeylessPublicKey_FromHex_Invalid(t *testing.T) {
	t.Parallel()

	var key KeylessPublicKey
	err := key.FromHex("0xzzzz")
	assert.Error(t, err)
}

func TestFederatedKeylessPublicKey_Verify(t *testing.T) {
	t.Parallel()

	key := &FederatedKeylessPublicKey{}
	assert.False(t, key.Verify([]byte("msg"), &Ed25519Signature{}))
}

func TestFederatedKeylessPublicKey_AuthKey(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	idc, _ := NewIdCommitment(idcBytes)
	key := &FederatedKeylessPublicKey{
		JwkAddr: types.AccountOne,
		Pk:      KeylessPublicKey{IssVal: "iss", Idc: *idc},
	}
	assert.NotNil(t, key.AuthKey())
}

func TestFederatedKeylessPublicKey_Scheme(t *testing.T) {
	t.Parallel()

	key := &FederatedKeylessPublicKey{}
	assert.Equal(t, SingleKeyScheme, key.Scheme())
}

func TestFederatedKeylessPublicKey_BytesAndFromBytes(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	idc, _ := NewIdCommitment(idcBytes)
	key := &FederatedKeylessPublicKey{
		JwkAddr: types.AccountOne,
		Pk:      KeylessPublicKey{IssVal: "iss", Idc: *idc},
	}

	data := key.Bytes()
	var key2 FederatedKeylessPublicKey
	err := key2.FromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, key.Pk.IssVal, key2.Pk.IssVal)
}

func TestFederatedKeylessPublicKey_ToHexAndFromHex(t *testing.T) {
	t.Parallel()

	idcBytes := make([]byte, IdCommitmentNumBytes)
	idc, _ := NewIdCommitment(idcBytes)
	key := &FederatedKeylessPublicKey{
		JwkAddr: types.AccountOne,
		Pk:      KeylessPublicKey{IssVal: "iss", Idc: *idc},
	}

	hex := key.ToHex()
	var key2 FederatedKeylessPublicKey
	err := key2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, key.Pk.IssVal, key2.Pk.IssVal)
}

func TestFederatedKeylessPublicKey_FromHex_Invalid(t *testing.T) {
	t.Parallel()

	var key FederatedKeylessPublicKey
	err := key.FromHex("0xzzzz")
	assert.Error(t, err)
}

func TestIdCommitment_Bytes_DefensiveCopy(t *testing.T) {
	t.Parallel()

	input := make([]byte, IdCommitmentNumBytes)
	input[0] = 42
	idc, err := NewIdCommitment(input)
	require.NoError(t, err)

	bytes1 := idc.Bytes()
	bytes1[0] = 0xff
	bytes2 := idc.Bytes()
	assert.Equal(t, byte(42), bytes2[0])
}

func TestIdCommitment_Bytes_NilReceiver(t *testing.T) {
	t.Parallel()

	var idc *IdCommitment
	assert.Nil(t, idc.Bytes())
}

func TestIdCommitment_Bytes_NilInner(t *testing.T) {
	t.Parallel()

	idc := &IdCommitment{}
	assert.Nil(t, idc.Bytes())
}

func TestNewIdCommitment_DefensiveCopy(t *testing.T) {
	t.Parallel()

	input := make([]byte, IdCommitmentNumBytes)
	input[0] = 42
	idc, err := NewIdCommitment(input)
	require.NoError(t, err)

	// Modifying input should not affect idc
	input[0] = 0xff
	assert.Equal(t, byte(42), idc.Bytes()[0])
}

func TestZKP_UnknownVariant(t *testing.T) {
	t.Parallel()

	ser := bcs.NewSerializer()
	ser.Uleb128(99) // Unknown variant
	data := ser.ToBytes()

	var zkp ZKP
	err := bcs.Deserialize(&zkp, data)
	assert.Error(t, err)
}

func TestEphemeralSignature_Ed25519_BCSRoundTrip(t *testing.T) {
	t.Parallel()

	sig := &EphemeralSignature{
		Variant:   EphemeralSignatureVariantEd25519,
		Signature: &Ed25519Signature{},
	}

	data, err := bcs.Serialize(sig)
	require.NoError(t, err)

	var sig2 EphemeralSignature
	err = bcs.Deserialize(&sig2, data)
	require.NoError(t, err)
	assert.Equal(t, EphemeralSignatureVariantEd25519, sig2.Variant)
}

func TestEphemeralPublicKey_Bytes(t *testing.T) {
	t.Parallel()

	epk := &EphemeralPublicKey{
		Variant: EphemeralPublicKeyVariantEd25519,
		PubKey:  &Ed25519PublicKey{},
	}
	assert.NotEmpty(t, epk.Bytes())
}

func makeTestKeylessSignature(t *testing.T) *KeylessSignature {
	t.Helper()
	privKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	edPubKey := privKey.PubKey().(*Ed25519PublicKey)

	return &KeylessSignature{
		Cert: EphemeralCertificate{
			Variant: EphemeralCertificateVariantZeroKnowledge,
			Cert: &ZeroKnowledgeSig{
				Proof: ZKP{
					Variant: ZKPVariantGroth16,
					Proof:   &Groth16Proof{},
				},
				ExpHorizonSecs: 1000,
			},
		},
		JwtHeaderJSON: `{"alg":"RS256"}`,
		ExpDateSecs:   9999999999,
		EphemeralPubkey: EphemeralPublicKey{
			Variant: EphemeralPublicKeyVariantEd25519,
			PubKey:  edPubKey,
		},
		EphemeralSignature: EphemeralSignature{
			Variant:   EphemeralSignatureVariantEd25519,
			Signature: &Ed25519Signature{},
		},
	}
}

func TestKeylessSignature_BytesAndFromBytes(t *testing.T) {
	t.Parallel()

	sig := makeTestKeylessSignature(t)

	data := sig.Bytes()
	assert.NotEmpty(t, data)

	var sig2 KeylessSignature
	err := sig2.FromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, sig.JwtHeaderJSON, sig2.JwtHeaderJSON)
}

func TestKeylessSignature_ToHexAndFromHex(t *testing.T) {
	t.Parallel()

	sig := makeTestKeylessSignature(t)

	hex := sig.ToHex()
	assert.NotEmpty(t, hex)

	var sig2 KeylessSignature
	err := sig2.FromHex(hex)
	require.NoError(t, err)
	assert.Equal(t, sig.JwtHeaderJSON, sig2.JwtHeaderJSON)
}

func TestKeylessSignature_FromHex_Invalid(t *testing.T) {
	t.Parallel()

	var sig KeylessSignature
	err := sig.FromHex("0xzzzz")
	assert.Error(t, err)
}

func TestKeylessSignature_ToAnySignature(t *testing.T) {
	t.Parallel()

	sig := &KeylessSignature{
		Cert: EphemeralCertificate{
			Variant: EphemeralCertificateVariantZeroKnowledge,
			Cert: &ZeroKnowledgeSig{
				Proof: ZKP{
					Variant: ZKPVariantGroth16,
					Proof:   &Groth16Proof{},
				},
			},
		},
		JwtHeaderJSON: `{"alg":"RS256"}`,
		EphemeralPubkey: EphemeralPublicKey{
			Variant: EphemeralPublicKeyVariantEd25519,
			PubKey:  &Ed25519PublicKey{},
		},
		EphemeralSignature: EphemeralSignature{
			Variant:   EphemeralSignatureVariantEd25519,
			Signature: &Ed25519Signature{},
		},
	}

	anySig := sig.ToAnySignature()
	assert.Equal(t, AnySignatureVariantKeyless, anySig.Variant)
}

func TestG1Bytes_BCSRoundTrip(t *testing.T) {
	t.Parallel()

	var g1 G1Bytes
	for i := range g1 {
		g1[i] = byte(i)
	}

	data, err := bcs.Serialize(&g1)
	require.NoError(t, err)

	var g1r G1Bytes
	err = bcs.Deserialize(&g1r, data)
	require.NoError(t, err)
	assert.Equal(t, g1, g1r)
}

func TestG2Bytes_BCSRoundTrip(t *testing.T) {
	t.Parallel()

	var g2 G2Bytes
	for i := range g2 {
		g2[i] = byte(i)
	}

	data, err := bcs.Serialize(&g2)
	require.NoError(t, err)

	var g2r G2Bytes
	err = bcs.Deserialize(&g2r, data)
	require.NoError(t, err)
	assert.Equal(t, g2, g2r)
}

func TestEphemeralCertificate_OpenID_BCSRoundTrip(t *testing.T) {
	t.Parallel()

	var pepper Pepper
	cert := &EphemeralCertificate{
		Variant: EphemeralCertificateVariantOpenId,
		Cert: &OpenIdSig{
			JwtSig:         make([]byte, 64),
			JwtPayloadJSON: `{"sub":"123"}`,
			UidKey:         "sub",
			EpkBlinder:     make([]byte, EpkBlinderNumBytes),
			Pepper:         pepper,
		},
	}

	data, err := bcs.Serialize(cert)
	require.NoError(t, err)

	var cert2 EphemeralCertificate
	err = bcs.Deserialize(&cert2, data)
	require.NoError(t, err)
	assert.Equal(t, EphemeralCertificateVariantOpenId, cert2.Variant)
}
