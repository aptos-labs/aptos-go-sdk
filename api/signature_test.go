package api

import (
	"encoding/json"
	"testing"

	"github.com/qimeila/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountAuthenticator_Unknown(t *testing.T) {
	t.Parallel()
	testJson := `{
      "type": "something"
	}`
	data := &Signature{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, SignatureVariantUnknown, data.Type)
	auth, ok := data.Inner.(*UnknownSignature)
	require.True(t, ok)

	assert.Equal(t, "something", auth.Type)
}

func TestAccountAuthenticator_Ed25519(t *testing.T) {
	t.Parallel()
	testJson := `{
      "public_key": "0xfc0947a61275f90ed089e1584143362eb236b11d72f901b8c2a5ca546f7fa34f",
      "signature": "0x0ba0310b8dad7053259b956f088779a59dc4a913e997678b4c8fb2da9a9d13d39736ad3a713ca300e7c8fcc98e483d829a8ddcf99df873038e3558ee982f6609",
      "type": "ed25519_signature"
	}`
	data := &Signature{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, SignatureVariantEd25519, data.Type)
	auth, ok := data.Inner.(*Ed25519Signature)
	require.True(t, ok)

	expectedPubKey := crypto.Ed25519PublicKey{}
	err = expectedPubKey.FromHex("0xfc0947a61275f90ed089e1584143362eb236b11d72f901b8c2a5ca546f7fa34f")
	require.NoError(t, err)

	expectedSignature := crypto.Ed25519Signature{}
	err = expectedSignature.FromHex("0x0ba0310b8dad7053259b956f088779a59dc4a913e997678b4c8fb2da9a9d13d39736ad3a713ca300e7c8fcc98e483d829a8ddcf99df873038e3558ee982f6609")
	require.NoError(t, err)
	assert.Equal(t, expectedSignature, *auth.Sig)
}

func TestAccountAuthenticator_FeePayer(t *testing.T) {
	t.Parallel()
	testJson := `{
  "sender": {
    "public_key": "0xfc0947a61275f90ed089e1584143362eb236b11d72f901b8c2a5ca546f7fa34f",
    "signature": "0x0ba0310b8dad7053259b956f088779a59dc4a913e997678b4c8fb2da9a9d13d39736ad3a713ca300e7c8fcc98e483d829a8ddcf99df873038e3558ee982f6609",
    "type": "ed25519_signature"
  },
  "secondary_signer_addresses": [],
  "secondary_signers": [],
  "fee_payer_address": "0xc1d18520beffe36d104232f455d5cc83b991bde0d1425a735aea1c0c2df60e0b",
  "fee_payer_signer": {
    "public_key": "0xcfbeb24598919df85ecb827b24bf70e082fd08fdefef8a4b470da16e633a8dee",
    "signature": "0x82d46bfb63d774fc724ed85b9822d318a79b9ec9a8d5cc1c56f4bd6964e13273e3f53962e5a2b75184544343adff70a9920167d9b1b84f8e5ad74dc8882b7707",
    "type": "ed25519_signature"
  },
  "type": "fee_payer_signature"
}`
	data := &Signature{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, SignatureVariantFeePayer, data.Type)

	// TODO: verify some parsing
	// auth := data.Inner.(*FeePayerSignature)
}
