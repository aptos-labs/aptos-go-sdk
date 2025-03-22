package aptos

import (
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/**
 * Behavioral specifications for Ed25519
 */

// Test_Spec_Ed25519_Generation tests the generation of Ed25519 keys.
//
//   - It must be able to generate keys.
//   - It should be able to generate keys with a specific seed, and have a deterministic outcome.
//   - It must not be able to generate the same key twice on default input.
func Test_Spec_Ed25519_Generation(t *testing.T) {
	t.Parallel()
	// It must be able to generate keys
	key1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err, "It must be able to generate keys")

	// It should be able to generate keys with a specific seed, and have a deterministic outcome
	seed := strings.NewReader("abcdefghijklmnopqrstuvwxyzabcdefghijk") // Note that this API on Go only supports 32-byte inputs TODO: Maybe we want to hash this seed?
	specificKey, err := crypto.GenerateEd25519PrivateKey(seed)
	require.NoError(t, err)
	seed = strings.NewReader("abcdefghijklmnopqrstuvwxyzabcdefghijk")
	specificKey2, err := crypto.GenerateEd25519PrivateKey(seed)
	require.NoError(t, err)
	assert.Equal(t, specificKey, specificKey2, "It should be able to generate keys with a specific seed, and have a deterministic outcome")

	// It must be able to not generate the same key twice on default input
	key2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	assert.NotEqual(t, key1, key2, "It must not be able to generate the same key twice on default input")
}

// Test_Spec_Ed25519_PrivateKey tests the Ed25519 private key
//
//   - It must be able to load a private key from a byte array
//   - It must not be able to load a private key from an invalid length byte array
//   - It must be able to load a private key from a 0x prefixed hex string
//   - It must not be able to load a private key from an invalid length hex string
//   - It must not be able to load a private key from an invalid hex string with invalid characters
//   - It must be able to load the same private key and be the same
//   - It must be able to output to a byte array
//   - It must be able to output to a 0x prefixed hex string
func Test_Spec_Ed25519_PrivateKey(t *testing.T) {
	t.Parallel()
	// It must be able to load a private key from a byte array
	key1 := &crypto.Ed25519PrivateKey{}
	err := key1.FromBytes(parseHex(TestEd25519PrivateKeyHex))
	require.NoError(t, err, "It must be able to load a private key from a byte array")

	// It must not be able to load a private key from an invalid length byte array
	err = key1.FromBytes(parseHex(TestInvalidHex))
	require.Error(t, err, "It must not be able to load a private key from an invalid length byte array")

	// It must be able to load a private key from a 0x prefixed hex string
	key2 := &crypto.Ed25519PrivateKey{}
	err = key2.FromHex(TestEd25519PrivateKeyHex)
	require.NoError(t, err, "It must be able to load a private key from a 0x prefixed hex string")

	// It must be able to load the same private key and be the same
	assert.Equal(t, key1, key2, "It must be able to load the same private key and be the same")

	// It must not be able to load a private key from an invalid length hex string
	err = key2.FromHex(TestInvalidHex)
	require.Error(t, err, "It must not be able to load a private key from an invalid length hex string")

	// It must not be able to load a private key from an invalid hex string with invalid characters
	err = key2.FromHex(TestInvalidHexCharacters)
	require.Error(t, err, "It must not be able to load a private key from an invalid hex string with invalid characters")

	// It must be able to output to a byte array
	assert.Equal(t, parseHex(TestEd25519PrivateKeyHex), key1.Bytes(), "It must be able to output to a byte array")

	// It must be able to output to a 0x prefixed hex string
	assert.Equal(t, TestEd25519PrivateKeyHex, key1.ToHex(), "It must be able to output to a 0x prefixed hex string")
}

// Test_Spec_Ed25519_PublicKey tests the Ed25519 public key
//
//   - It must be able to load a public key from a byte array
//   - It must not be able to load a public key from an invalid length byte array
//   - It must be able to load a public key from a 0x prefixed hex string
//   - It must not be able to load a public key from an invalid length hex string
//   - It must not be able to load a private key from an invalid hex string with invalid characters
//   - It must be able to load the same public key and be the same
//   - It must be able to output to a byte array
//   - It must be able to output to a 0x prefixed hex string
//   - It must be able to encode in BCS bytes
//   - It must be able to decode from BCS bytes
//   - It must be able to encode in BCS bytes and decode back to the same
//   - It must be able to catch an invalid byte size from BCS bytes
//   - It must be able to generate an AuthenticationKey
func Test_Spec_Ed25519_PublicKey(t *testing.T) {
	t.Parallel()
	// It must be able to load a public key from a byte array
	key1 := &crypto.Ed25519PublicKey{}
	err := key1.FromBytes(parseHex(TestEd25519PublicKeyHex))
	require.NoError(t, err, "It must be able to load a public key from a byte array")

	// It must not be able to load a public key from an invalid length byte array
	err = key1.FromBytes(parseHex(TestInvalidHex))
	require.Error(t, err, "It must not be able to load a public key from an invalid length byte array")

	// It must be able to load a public key from a 0x prefixed hex string
	key2 := &crypto.Ed25519PublicKey{}
	err = key2.FromHex(TestEd25519PublicKeyHex)
	require.NoError(t, err, "It must be able to load a public key from a 0x prefixed hex string")

	// It must not be able to load a public key from an invalid length hex string
	err = key2.FromHex(TestInvalidHex)
	require.Error(t, err, "It must not be able to load a public key from an invalid length hex string")

	// It must not be able to load a private key from an invalid hex string with invalid characters
	err = key2.FromHex(TestInvalidHexCharacters)
	require.Error(t, err, "It must not be able to load a public key from an invalid hex string with invalid characters")

	// It must be able to load the same public key and be the same
	assert.Equal(t, key1, key2, "It must able to load the same public key and be the same")

	// It must be able to output to a byte array
	assert.Equal(t, parseHex(TestEd25519PublicKeyHex), key1.Bytes(), "It must be able to output to a byte array")

	// It must be able to output to a 0x prefixed hex string
	assert.Equal(t, TestEd25519PublicKeyHex, key1.ToHex(), "It must be able to output to a 0x prefixed hex string")

	// It must be able to encode in BCS bytes
	bcsBytes1, err := bcs.Serialize(key1)
	require.NoError(t, err, "It must be able to encode in BCS bytes")

	// It must be able to decode from BCS bytes
	decodedKey := &crypto.Ed25519PublicKey{}
	err = bcs.Deserialize(decodedKey, bcsBytes1)
	require.NoError(t, err, "It must be able to decode from BCS bytes")

	// It must be able to encode in BCS bytes and decode back to the same
	assert.Equal(t, key1, decodedKey, "It must be able to encode in BCS bytes and decode back to the same")

	// It must be able to catch an invalid byte size from BCS bytes
	err = bcs.Deserialize(decodedKey, parseHex(TestInvalidHex))
	require.Error(t, err, "It must be able to catch an invalid byte size from BCS bytes")

	// It must be able to generate an AuthenticationKey
	authKey := key1.AuthKey()
	assert.Equal(t, TestEd25519AddressHex, authKey.ToHex(), "It must be able to generate an AuthenticationKey")
}

// Test_Spec_Ed25519_Signature tests the Ed25519 signature
//
//   - It must be able to load a signature from a byte array
//   - It must not be able to load a signature from an invalid length byte array
//   - It must be able to load a signature from a 0x prefixed hex string
//   - It must not be able to load a signature from an invalid length hex string
//   - It must not be able to load a signature from an invalid hex string with invalid characters
//   - It must be able to load the same signature and be the same
//   - It must be able to output to a byte array
//   - It must be able to output to a 0x prefixed hex string
//   - It must be able to encode in BCS bytes
//   - It must be able to decode from BCS bytes
//   - It must be able to encode in BCS bytes and decode back to the same
//   - It must be able to catch an invalid byte size from BCS bytes
func Test_Spec_Ed25519_Signature(t *testing.T) {
	t.Parallel()
	// It must be able to load a signature from a byte array
	sig1 := &crypto.Ed25519Signature{}
	err := sig1.FromBytes(parseHex(TestSecp256k1SignatureHex))
	require.NoError(t, err, "It must be able to load a signature from a byte array")

	// It must not be able to load a signature from an invalid length byte array
	err = sig1.FromBytes(parseHex(TestInvalidHex))
	require.Error(t, err, "It must not be able to load a signature from an invalid length byte array")

	// It must be able to load a signature from a 0x prefixed hex string
	sig2 := &crypto.Ed25519Signature{}
	err = sig2.FromHex(TestSecp256k1SignatureHex)
	require.NoError(t, err, "It must be able to load a signature from a 0x prefixed hex string")

	// It must not be able to load a signature from an invalid length hex string
	err = sig2.FromHex(TestInvalidHex)
	require.Error(t, err, "It must not be able to load a signature from an invalid length hex string")

	// It must not be able to load a signature from an invalid hex string with invalid characters
	err = sig2.FromHex(TestInvalidHexCharacters)
	require.Error(t, err, "It must not be able to load a signature from an invalid hex string with invalid characters")

	// It must be able to load the same signature and be the same
	assert.Equal(t, sig1, sig2, "It must able to load the same signature and be the same")

	// It must be able to output to a byte array
	assert.Equal(t, parseHex(TestSecp256k1SignatureHex), sig1.Bytes(), "It must be able to output to a byte array")

	// It must be able to output to a 0x prefixed hex string
	assert.Equal(t, TestSecp256k1SignatureHex, sig1.ToHex(), "It must be able to output to a 0x prefixed hex string")

	// It must be able to encode in BCS bytes
	bcsBytes1, err := bcs.Serialize(sig1)
	require.NoError(t, err, "It must be able to encode in BCS bytes")

	// It must be able to decode from BCS bytes
	decodedSig := &crypto.Ed25519Signature{}
	err = bcs.Deserialize(decodedSig, bcsBytes1)
	require.NoError(t, err, "It must be able to decode from BCS bytes")

	// It must be able to encode in BCS bytes and decode back to the same
	assert.Equal(t, sig1, decodedSig, "It must be able to encode in BCS bytes and decode back to the same")

	// It must be able to catch an invalid byte size from BCS bytes
	err = bcs.Deserialize(decodedSig, parseHex(TestInvalidHex))
	require.Error(t, err, "It must be able to catch an invalid byte size from BCS bytes")
}

// Test_Spec_Ed25519_Authentication tests the Ed25519 authenticator
//
//   - It must be able to generate an AccountAuthenticator
//   - It must be able to verify the message with the AccountAuthenticator
//   - It must be able to not-verify the message with the AccountAuthenticator, with the wrong message
//   - It should be able to generate an empty signature AccountAuthenticator
//   - It must be able to encode in BCS bytes
//   - It must be able to decode from BCS bytes
//   - It must be able to encode in BCS bytes and decode back to the same
//   - It must be able to catch an invalid byte size from BCS bytes
func Test_Spec_Ed25519_Authenticator(t *testing.T) {
	t.Parallel()
	key1 := &crypto.Ed25519PrivateKey{}
	err := key1.FromHex(TestEd25519PrivateKeyHex)
	require.NoError(t, err)
	pubKey1 := key1.PubKey()

	// It must be able to generate an AccountAuthenticator
	message := parseHex(TestEd25519Message)
	auth1, err := key1.Sign(message)
	require.NoError(t, err, "It must be able to generate an AccountAuthenticator")

	// It must be able to verify the message with the AccountAuthenticator
	assert.True(t, auth1.Verify(message), "It must be able to verify the message with the AccountAuthenticator")

	// It must be able to not-verify the message with the AccountAuthenticator, with the wrong message
	assert.False(t, auth1.Verify(parseHex(OtherMessage)), "It must be able to not-verify the message with the AccountAuthenticator, with the wrong message")

	// It should be able to generate an empty signature AccountAuthenticator
	emptySig := key1.EmptySignature()
	assert.Equal(t, &crypto.Ed25519Signature{}, emptySig, "It should be able to generate an empty signature AccountAuthenticator")
	emptyAuth := key1.SimulationAuthenticator()
	assert.Equal(t, &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorEd25519,
		Auth: &crypto.Ed25519Authenticator{
			PubKey: pubKey1.(*crypto.Ed25519PublicKey),
			Sig:    emptySig.(*crypto.Ed25519Signature),
		},
	}, emptyAuth, "It should be able to generate an empty signature AccountAuthenticator")

	// It must be able to encode in BCS bytes
	bcsBytes1, err := bcs.Serialize(auth1)
	require.NoError(t, err, "It must be able to encode in BCS bytes")

	// It must be able to decode from BCS bytes
	decodedAuth := &crypto.AccountAuthenticator{}
	err = bcs.Deserialize(decodedAuth, bcsBytes1)
	require.NoError(t, err, "It must be able to decode from BCS bytes")

	// It must be able to encode in BCS bytes and decode back to the same
	assert.Equal(t, auth1, decodedAuth, "It must be able to encode in BCS bytes and decode back to the same")
}

// Test_Spec_Ed25519_Signing tests the signing of Ed25519 keys
//
//   - It must be able to generate a public key from the private key
//   - It must be able to sign messages
//   - It must have deterministic signing
//   - It must have different signatures for different keys
//   - It must have different signatures for different messages
//   - It must be able to verify the message with the public key
//   - It must be able to not-verify a message with the wrong public key
//   - It must be able to not-verify a message with the wrong signature

func Test_Spec_Ed25519_Signing(t *testing.T) {
	t.Parallel()
	key1 := &crypto.Ed25519PrivateKey{}
	err := key1.FromHex(TestEd25519PrivateKeyHex)
	require.NoError(t, err)
	key2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	// It must be able to generate a public key from the private key
	pubkey1 := key1.PubKey()
	pubkey2 := key2.PubKey()

	// It must be able to sign messages
	message := parseHex(TestEd25519Message)
	signature1, err := key1.SignMessage(message)
	require.NoError(t, err, "It must be able to sign messages")
	assert.Equal(t, TestEd25519SignatureHex, signature1.ToHex(), "It must be able to sign messages")

	signature1Copy, err := key1.SignMessage(message)
	require.NoError(t, err)
	assert.Equal(t, signature1, signature1Copy, "It must have deterministic signing")

	// It must have different signatures for different keys
	signature2, err := key2.SignMessage(message)
	require.NoError(t, err)
	assert.NotEqual(t, signature1, signature2, "It must have different signatures for different keys")

	// It must have different signatures for different messages
	message2 := parseHex(OtherMessage)
	signature1Message2, err := key1.SignMessage(message2)
	require.NoError(t, err)
	assert.NotEqual(t, signature1, signature1Message2, "It must have different signatures for different messages")

	// It must be able to verify the message with the public key
	assert.True(t, pubkey1.Verify(message, signature1), "It must be able to verify the message with the public key")
	assert.True(t, pubkey1.Verify(message, signature1Copy), "It must be able to verify the message with the public key")
	assert.True(t, pubkey1.Verify(message2, signature1Message2), "It must be able to verify the message with the public key")
	assert.True(t, pubkey2.Verify(message, signature2), "It must be able to verify the message with the public key")

	// It must be able to not-verify a message with the wrong public key
	assert.False(t, pubkey2.Verify(message, signature1), "It must be able to not-verify a message with the wrong public key")
	assert.False(t, pubkey1.Verify(message2, signature2), "It must be able to not-verify a message with the wrong public key")

	// It must be able to not-verify a message with the wrong signature
	assert.False(t, pubkey1.Verify(message2, signature1), "It must be able to not-verify a message with the wrong signature")
	assert.False(t, pubkey2.Verify(message2, signature2), "It must be able to not-verify a message with the wrong signature")
}
