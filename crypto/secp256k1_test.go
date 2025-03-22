package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecp256k1PrivateKey     = "secp256k1-priv-0xd107155adf816a0a94c6db3c9489c13ad8a1eda7ada2e558ba3bfa47c020347e"
	testSecp256k1PrivateKeyHex  = "0xd107155adf816a0a94c6db3c9489c13ad8a1eda7ada2e558ba3bfa47c020347e"
	testSecp256k1PublicKey      = "0x04acdd16651b839c24665b7e2033b55225f384554949fef46c397b5275f37f6ee95554d70fb5d9f93c5831ebf695c7206e7477ce708f03ae9bb2862dc6c9e033ea"
	testSecp256k1Address        = "0x5792c985bc96f436270bd2a3c692210b09c7febb8889345ceefdbae4bacfe498"
	testSecp256k1MessageEncoded = "0x68656c6c6f20776f726c64"
	testSecp256k1Signature      = "0xd0d634e843b61339473b028105930ace022980708b2855954b977da09df84a770c0b68c29c8ca1b5409a5085b0ec263be80e433c83fcf6debb82f3447e71edca"
)

func TestSecp256k1Keys(t *testing.T) {
	t.Parallel()
	testSecp256k1PrivateKeyBytes, err := util.ParseHex(testSecp256k1PrivateKeyHex)
	require.NoError(t, err)

	// Either bytes or hex should work
	privateKey := &Secp256k1PrivateKey{}
	err = privateKey.FromHex(testSecp256k1PrivateKey)
	require.NoError(t, err)
	privateKey2 := &Secp256k1PrivateKey{}
	err = privateKey2.FromBytes(testSecp256k1PrivateKeyBytes)
	require.NoError(t, err)
	assert.Equal(t, privateKey, privateKey2)

	// The outputs should match as well
	assert.Equal(t, testSecp256k1PrivateKeyBytes, privateKey.Bytes())
	assert.Equal(t, testSecp256k1PrivateKeyHex, privateKey.ToHex())
	formattedString, err := privateKey.ToAIP80()
	require.NoError(t, err)
	assert.Equal(t, testSecp256k1PrivateKey, formattedString)

	// Auth key should match
	singleSender := SingleSigner{privateKey}
	assert.Equal(t, testSecp256k1Address, singleSender.AuthKey().ToHex())

	// Test signature
	message, err := util.ParseHex(testSecp256k1MessageEncoded)
	require.NoError(t, err)
	signature, err := privateKey.SignMessage(message)
	require.NoError(t, err)

	authenticator, err := singleSender.Sign(message)
	require.NoError(t, err)

	// We have to wrap the signature in AnySignature
	anySig := &AnySignature{AnySignatureVariantSecp256k1, signature}
	assert.Equal(t, anySig, authenticator.Signature())

	// Check public key
	assert.Equal(t, testSecp256k1PublicKey, privateKey.VerifyingKey().ToHex())
	// We have to wrap the PublicKey in AnyPublicKey, verify authenticator and key are the same
	publicKey, err := ToAnyPublicKey(privateKey.VerifyingKey())
	require.NoError(t, err)
	assert.Equal(t, publicKey, authenticator.PubKey())

	// Check signature (without a recovery bit)
	actualSignature, ok := authenticator.Signature().(*AnySignature)
	require.True(t, ok)
	assert.Equal(t, len(testSecp256k1Signature), len(actualSignature.Signature.ToHex()))
	secp256k1Signature, ok := actualSignature.Signature.(*Secp256k1Signature)
	require.True(t, ok)
	assert.Equal(t, testSecp256k1Signature, secp256k1Signature.ToHex())

	// Verify signature with the key and the authenticator directly
	assert.True(t, authenticator.Verify(message))
	assert.True(t, publicKey.Verify(message, actualSignature))

	// Verify serialization of public key
	publicKeyBytes, err := bcs.Serialize(publicKey)
	require.NoError(t, err)
	expectedPublicKeyBytes, err := util.ParseHex(testSecp256k1PublicKey)
	require.NoError(t, err)
	// Need to prepend the length
	expectedBcsPublicKeyBytes := []byte{Secp256k1PublicKeyLength}
	expectedBcsPublicKeyBytes = append(expectedBcsPublicKeyBytes, expectedPublicKeyBytes...)
	assert.Equal(t, append([]byte{0x1}, expectedBcsPublicKeyBytes...), publicKeyBytes)

	publicKey2 := &AnyPublicKey{}
	err = bcs.Deserialize(publicKey2, publicKeyBytes)
	require.NoError(t, err)
	assert.Equal(t, publicKey, publicKey2)

	// Check from bytes and from hex
	publicKey3Inner := &Secp256k1PublicKey{}
	err = publicKey3Inner.FromHex(testSecp256k1PublicKey)
	require.NoError(t, err)
	publicKey3, err := ToAnyPublicKey(publicKey3Inner)
	require.NoError(t, err)
	publicKey4Inner := &Secp256k1PublicKey{}
	err = publicKey4Inner.FromBytes(expectedPublicKeyBytes)
	require.NoError(t, err)
	publicKey4, err := ToAnyPublicKey(publicKey4Inner)
	require.NoError(t, err)

	assert.Equal(t, publicKey.ToHex(), publicKey3.ToHex())
	assert.Equal(t, publicKey.ToHex(), publicKey4.ToHex())

	// Test serialization and deserialization of authenticator
	authenticatorBytes, err := bcs.Serialize(authenticator)
	require.NoError(t, err)
	authenticator2 := &AccountAuthenticator{}
	err = bcs.Deserialize(authenticator2, authenticatorBytes)
	require.NoError(t, err)
	assert.Equal(t, authenticator.Variant, authenticator2.Variant)
	assert.Equal(t, authenticator.Auth.PublicKey(), authenticator2.Auth.PublicKey())
	assert.Equal(t, authenticator.Auth.Signature().ToHex(), authenticator2.Auth.Signature().ToHex())
	authBytes, err := bcs.Serialize(authenticator)
	require.NoError(t, err)
	authBytes2, err := bcs.Serialize(authenticator2)
	require.NoError(t, err)
	assert.Equal(t, authBytes, authBytes2)
}

func TestGenerateSecp256k1Key(t *testing.T) {
	t.Parallel()
	privateKey, err := GenerateSecp256k1Key()
	require.NoError(t, err)

	// Sign a message
	msg := []byte("test message")
	sig, err := privateKey.SignMessage(msg)
	require.NoError(t, err)

	assert.True(t, privateKey.VerifyingKey().Verify(msg, sig))
}

func TestSecp256k1Signature_RecoverPublicKey(t *testing.T) {
	t.Parallel()
	privateKey := &Secp256k1PrivateKey{}
	err := privateKey.FromHex(testSecp256k1PrivateKey)
	require.NoError(t, err)
	message := []byte("hello")

	signature, err := privateKey.SignMessage(message)
	require.NoError(t, err)

	// Recover the public key
	sig, ok := signature.(*Secp256k1Signature)
	require.True(t, ok)
	recoveredKey, err := sig.RecoverPublicKey(message, 1)
	require.NoError(t, err)

	// Verify the signature with the key
	assert.True(t, recoveredKey.Verify(message, signature))
	assert.Equal(t, privateKey.VerifyingKey().ToHex(), recoveredKey.ToHex())

	// Also try with the recovery bit attached to the signature
	anyPubKey, err := ToAnyPublicKey(privateKey.VerifyingKey())
	require.NoError(t, err)
	sig2, ok := signature.(*Secp256k1Signature)
	require.True(t, ok)
	recoveredKey2, err := sig2.RecoverSecp256k1PublicKeyWithAuthenticationKey(message, anyPubKey.AuthKey())
	require.NoError(t, err)
	assert.Equal(t, privateKey.VerifyingKey().ToHex(), recoveredKey2.ToHex())
}

func TestSecp256k1Signature_RecoverPublicKeyFromSignature(t *testing.T) {
	t.Parallel()
	privateKey := &Secp256k1PrivateKey{}
	err := privateKey.FromHex(testSecp256k1PrivateKey)
	require.NoError(t, err)
	publicKey := &Secp256k1PublicKey{}
	err = publicKey.FromHex(testSecp256k1PublicKey)
	require.NoError(t, err)
	message, err := util.ParseHex(testSecp256k1MessageEncoded)
	require.NoError(t, err)

	assert.Equal(t, publicKey.ToHex(), privateKey.VerifyingKey().ToHex())

	signature := &Secp256k1Signature{}
	err = signature.FromHex(testSecp256k1Signature)
	require.NoError(t, err)

	// Recover the public key
	recoveryBit := byte(0)
	recoveredKey, err := signature.RecoverPublicKey(message, recoveryBit)
	require.NoError(t, err)

	// Verify the signature with the key
	assert.True(t, recoveredKey.Verify(message, signature))
	assert.Equal(t, publicKey.ToHex(), recoveredKey.ToHex())
}

func TestSecp256k1Signature_RecoverPublicKeyFromSignatureWithRecoveryBit(t *testing.T) {
	t.Parallel()
	privateKey := &Secp256k1PrivateKey{}
	err := privateKey.FromHex(testSecp256k1PrivateKey)
	require.NoError(t, err)
	publicKey := &Secp256k1PublicKey{}
	err = publicKey.FromHex(testSecp256k1PublicKey)
	require.NoError(t, err)
	message, err := util.ParseHex(testSecp256k1MessageEncoded)
	require.NoError(t, err)

	assert.Equal(t, publicKey.ToHex(), privateKey.VerifyingKey().ToHex())

	signature := &Secp256k1Signature{}
	err = signature.FromHex(testSecp256k1Signature)
	require.NoError(t, err)

	// Recover the public key
	recoveryBit := byte(0)
	recoveredKey, err := signature.RecoverPublicKey(message, recoveryBit)
	require.NoError(t, err)

	// Verify the signature with the key
	assert.True(t, recoveredKey.Verify(message, signature))
	assert.Equal(t, publicKey.ToHex(), recoveredKey.ToHex())
}
