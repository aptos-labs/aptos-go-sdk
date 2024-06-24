package crypto

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testSecp256k1PrivateKey     = "0xd107155adf816a0a94c6db3c9489c13ad8a1eda7ada2e558ba3bfa47c020347e"
	testSecp256k1PublicKey      = "0x04acdd16651b839c24665b7e2033b55225f384554949fef46c397b5275f37f6ee95554d70fb5d9f93c5831ebf695c7206e7477ce708f03ae9bb2862dc6c9e033ea"
	testSecp256k1Address        = "0x5792c985bc96f436270bd2a3c692210b09c7febb8889345ceefdbae4bacfe498"
	testSecp256k1MessageEncoded = "0x68656c6c6f20776f726c64"
	testSecp256k1Signature      = "0xd0d634e843b61339473b028105930ace022980708b2855954b977da09df84a770c0b68c29c8ca1b5409a5085b0ec263be80e433c83fcf6debb82f3447e71edca"
)

func TestSecp256k1Keys(t *testing.T) {
	testSecp256k1PrivateKeyBytes, err := util.ParseHex(testSecp256k1PrivateKey)
	assert.NoError(t, err)

	// Either bytes or hex should work
	privateKey := &Secp256k1PrivateKey{}
	err = privateKey.FromHex(testSecp256k1PrivateKey)
	assert.NoError(t, err)
	privateKey2 := &Secp256k1PrivateKey{}
	err = privateKey2.FromBytes(testSecp256k1PrivateKeyBytes)
	assert.NoError(t, err)
	assert.Equal(t, privateKey, privateKey2)

	// The outputs should match as well
	assert.Equal(t, privateKey.Bytes(), testSecp256k1PrivateKeyBytes)
	assert.Equal(t, privateKey.ToHex(), testSecp256k1PrivateKey)

	// Auth key should match
	singleSender := SingleSigner{privateKey}
	assert.Equal(t, testSecp256k1Address, singleSender.AuthKey().ToHex())

	// Test signature
	message, err := util.ParseHex(testSecp256k1MessageEncoded)
	assert.NoError(t, err)
	signature, err := privateKey.SignMessage(message)
	assert.NoError(t, err)

	authenticator, err := singleSender.Sign(message)
	assert.NoError(t, err)

	// We have to wrap the signature in AnySignature
	anySig := &AnySignature{AnySignatureVariantSecp256k1, signature}
	assert.Equal(t, anySig, authenticator.Signature())

	// Check public key
	assert.Equal(t, testSecp256k1PublicKey, privateKey.VerifyingKey().ToHex())
	// We have to wrap the PublicKey in AnyPublicKey, verify authenticator and key are the same
	publicKey := ToAnyPublicKey(privateKey.VerifyingKey())
	assert.Equal(t, publicKey, authenticator.PubKey())

	// Check signature (without a recovery bit)
	actualSignature := authenticator.Signature().(*AnySignature)
	assert.Equal(t, len(testSecp256k1Signature), len(actualSignature.Signature.ToHex()))
	assert.Equal(t, testSecp256k1Signature, actualSignature.Signature.(*Secp256k1Signature).ToHex())

	// Verify signature with the key and the authenticator directly
	assert.True(t, authenticator.Verify(message))
	assert.True(t, publicKey.Verify(message, actualSignature))

	// Verify serialization of public key
	publicKeyBytes, err := bcs.Serialize(publicKey)
	assert.NoError(t, err)
	expectedPublicKeyBytes, err := util.ParseHex(testSecp256k1PublicKey)
	assert.NoError(t, err)
	// Need to prepend the length
	expectedBcsPublicKeyBytes := []byte{Secp256k1PublicKeyLength}
	expectedBcsPublicKeyBytes = append(expectedBcsPublicKeyBytes, expectedPublicKeyBytes[:]...)
	assert.Equal(t, append([]byte{0x1}, expectedBcsPublicKeyBytes...), publicKeyBytes)

	publicKey2 := &AnyPublicKey{}
	err = bcs.Deserialize(publicKey2, publicKeyBytes)
	assert.NoError(t, err)
	assert.Equal(t, publicKey, publicKey2)

	// Check from bytes and from hex
	publicKey3Inner := &Secp256k1PublicKey{}
	err = publicKey3Inner.FromHex(testSecp256k1PublicKey)
	assert.NoError(t, err)
	publicKey3 := ToAnyPublicKey(publicKey3Inner)
	publicKey4Inner := &Secp256k1PublicKey{}
	err = publicKey4Inner.FromBytes(expectedPublicKeyBytes)
	assert.NoError(t, err)
	publicKey4 := ToAnyPublicKey(publicKey4Inner)

	assert.Equal(t, publicKey.ToHex(), publicKey3.ToHex())
	assert.Equal(t, publicKey.ToHex(), publicKey4.ToHex())

	// Test serialization and deserialization of authenticator
	authenticatorBytes, err := bcs.Serialize(authenticator)
	assert.NoError(t, err)
	authenticator2 := &AccountAuthenticator{}
	err = bcs.Deserialize(authenticator2, authenticatorBytes)
	assert.NoError(t, err)
	assert.Equal(t, authenticator.Variant, authenticator2.Variant)
	assert.Equal(t, authenticator.Auth.PublicKey(), authenticator2.Auth.PublicKey())
	assert.Equal(t, authenticator.Auth.Signature().ToHex(), authenticator2.Auth.Signature().ToHex())
	authBytes, err := bcs.Serialize(authenticator)
	assert.NoError(t, err)
	authBytes2, err := bcs.Serialize(authenticator2)
	assert.NoError(t, err)
	assert.Equal(t, authBytes, authBytes2)
}
