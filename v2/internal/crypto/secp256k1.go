package crypto

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Secp256k1 key and signature sizes.
const (
	Secp256k1PrivateKeyLength = 32
	Secp256k1PublicKeyLength  = 65 // Uncompressed format
	Secp256k1SignatureLength  = 64 // Without recovery bit
)

// Secp256k1PrivateKey is a Secp256k1 private key.
//
// Unlike Ed25519PrivateKey, Secp256k1PrivateKey does not implement [Signer] directly.
// It must be wrapped with [SingleSigner] for transaction signing:
//
//	key, _ := crypto.GenerateSecp256k1Key()
//	signer := crypto.NewSingleSigner(key)
//	auth, _ := signer.Sign(txnBytes)
//
// Secp256k1PrivateKey supports concurrent read-only use after initialization.
// The mutex protects cached values (such as the derived public key), but callers
// must not mutate the underlying key material (e.g., via FromBytes) concurrently
// with signing or other operations that use the key.
//
// Implements:
//   - [MessageSigner]
//   - [CryptoMaterial]
type Secp256k1PrivateKey struct {
	Inner *secp256k1.PrivateKey

	// mu protects cached values for concurrent access
	mu sync.RWMutex
	// Cached public key to avoid repeated derivation
	cachedPubKey *Secp256k1PublicKey
}

// GenerateSecp256k1Key generates a new random Secp256k1 key pair.
func GenerateSecp256k1Key() (*Secp256k1PrivateKey, error) {
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	return &Secp256k1PrivateKey{Inner: priv}, nil
}

// SignMessage signs a message and returns the signature.
// The message is hashed with SHA3-256 before signing.
//
// Implements [MessageSigner].
func (key *Secp256k1PrivateKey) SignMessage(msg []byte) (Signature, error) {
	hash := util.Sha3256Hash([][]byte{msg})
	sig := ecdsa.Sign(key.Inner, hash)
	return &Secp256k1Signature{Inner: sig}, nil
}

// EmptySignature returns an empty signature for simulation.
//
// Implements [MessageSigner].
func (key *Secp256k1PrivateKey) EmptySignature() Signature {
	return &Secp256k1Signature{
		Inner: ecdsa.NewSignature(&secp256k1.ModNScalar{}, &secp256k1.ModNScalar{}),
	}
}

// VerifyingKey returns the public key for verification.
// The result is cached after the first call to avoid repeated derivation. Thread-safe.
//
// Implements [MessageSigner].
func (key *Secp256k1PrivateKey) VerifyingKey() VerifyingKey {
	key.mu.RLock()
	if key.cachedPubKey != nil {
		cached := key.cachedPubKey
		key.mu.RUnlock()
		return cached
	}
	key.mu.RUnlock()

	key.mu.Lock()
	defer key.mu.Unlock()
	// Double-check after acquiring write lock
	if key.cachedPubKey != nil {
		return key.cachedPubKey
	}
	key.cachedPubKey = &Secp256k1PublicKey{Inner: key.Inner.PubKey()}
	return key.cachedPubKey
}

// Bytes returns the raw private key bytes.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PrivateKey) Bytes() []byte {
	return key.Inner.Serialize()
}

// FromBytes loads the private key from bytes.
// This clears any cached public key. Thread-safe.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PrivateKey) FromBytes(bytes []byte) error {
	bytes, err := ParsePrivateKey(bytes, PrivateKeyVariantSecp256k1, false)
	if err != nil {
		return err
	}
	if len(bytes) != Secp256k1PrivateKeyLength {
		return fmt.Errorf("secp256k1 private key must be %d bytes", Secp256k1PrivateKeyLength)
	}

	key.mu.Lock()
	defer key.mu.Unlock()
	key.Inner = secp256k1.PrivKeyFromBytes(bytes)
	// Clear cached value since key changed
	key.cachedPubKey = nil
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the private key.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PrivateKey) FromHex(hexStr string) error {
	bytes, err := ParsePrivateKey(hexStr, PrivateKeyVariantSecp256k1)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// ToAIP80 formats the private key as an AIP-80 compliant string.
func (key *Secp256k1PrivateKey) ToAIP80() (string, error) {
	return FormatPrivateKey(key.ToHex(), PrivateKeyVariantSecp256k1)
}

// String returns a redacted representation to prevent accidental logging of private keys.
// Use ToAIP80() to get the actual private key string.
func (key *Secp256k1PrivateKey) String() string {
	return "<Secp256k1PrivateKey:REDACTED>"
}

// Secp256k1PublicKey is a Secp256k1 public key for signature verification.
//
// Note: Secp256k1PublicKey cannot be used directly as an on-chain PublicKey.
// It must be wrapped with [AnyPublicKey] via [SingleSigner].
//
// Implements:
//   - [VerifyingKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type Secp256k1PublicKey struct {
	Inner *secp256k1.PublicKey
}

// Verify verifies a signature against a message.
// The message is hashed with SHA3-256 before verification.
//
// Implements [VerifyingKey].
func (key *Secp256k1PublicKey) Verify(msg []byte, sig Signature) bool {
	secpSig, ok := sig.(*Secp256k1Signature)
	if !ok {
		return false
	}
	hash := util.Sha3256Hash([][]byte{msg})
	return secpSig.Inner.Verify(hash, key.Inner)
}

// Bytes returns the uncompressed public key bytes.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PublicKey) Bytes() []byte {
	return key.Inner.SerializeUncompressed()
}

// FromBytes loads the public key from bytes.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PublicKey) FromBytes(bytes []byte) error {
	pubKey, err := secp256k1.ParsePubKey(bytes)
	if err != nil {
		return err
	}
	key.Inner = pubKey
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *Secp256k1PublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *Secp256k1PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *Secp256k1PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	pubKey, err := secp256k1.ParsePubKey(bytes)
	if err != nil {
		des.SetError(err)
		return
	}
	key.Inner = pubKey
}

// Secp256k1Signature is a Secp256k1 ECDSA signature without recovery bit.
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type Secp256k1Signature struct {
	Inner *ecdsa.Signature
}

// Bytes returns the signature bytes (r || s, 64 bytes total).
//
// Implements [CryptoMaterial].
func (e *Secp256k1Signature) Bytes() []byte {
	r := e.Inner.R()
	s := e.Inner.S()
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Pre-allocate the exact size needed to avoid append reallocation
	result := make([]byte, Secp256k1SignatureLength)
	copy(result[0:32], rBytes[:])
	copy(result[32:64], sBytes[:])
	return result
}

// FromBytes loads the signature from bytes.
//
// Implements [CryptoMaterial].
func (e *Secp256k1Signature) FromBytes(bytes []byte) error {
	if len(bytes) != Secp256k1SignatureLength {
		return fmt.Errorf("secp256k1 signature must be %d bytes, got %d", Secp256k1SignatureLength, len(bytes))
	}

	var rBytes, sBytes [32]byte
	copy(rBytes[:], bytes[0:32])
	copy(sBytes[:], bytes[32:64])

	r := &secp256k1.ModNScalar{}
	r.SetBytes(&rBytes)
	s := &secp256k1.ModNScalar{}
	s.SetBytes(&sBytes)

	sig := ecdsa.NewSignature(r, s)

	// Check that s is in low order (required by Aptos)
	sVal := sig.S()
	if sVal.IsOverHalfOrder() {
		return errors.New("secp256k1 signature: s is over half order")
	}

	e.Inner = sig
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (e *Secp256k1Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (e *Secp256k1Signature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (e *Secp256k1Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(e.Bytes())
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (e *Secp256k1Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := e.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// RecoverPublicKey recovers the public key from the signature and message.
// The recovery bit (0-3) must be known.
func (e *Secp256k1Signature) RecoverPublicKey(message []byte, recoveryBit byte) (*Secp256k1PublicKey, error) {
	hash := util.Sha3256Hash([][]byte{message})
	return e.recoverPublicKey(hash, recoveryBit)
}

// RecoverPublicKeyWithAuthKey recovers the public key by trying all recovery bits
// and matching against the authentication key.
func (e *Secp256k1Signature) RecoverPublicKeyWithAuthKey(message []byte, authKey *AuthenticationKey) (*Secp256k1PublicKey, error) {
	hash := util.Sha3256Hash([][]byte{message})

	for i := byte(0); i < 4; i++ {
		key, err := e.recoverPublicKey(hash, i)
		if err != nil {
			continue
		}

		// Wrap and check auth key
		anyPubKey, err := ToAnyPublicKey(key)
		if err != nil {
			continue
		}

		if *anyPubKey.AuthKey() == *authKey {
			return key, nil
		}
	}

	return nil, errors.New("unable to recover public key from signature")
}

func (e *Secp256k1Signature) recoverPublicKey(messageHash []byte, recoveryBit byte) (*Secp256k1PublicKey, error) {
	// Bitcoin magic number 27 + recovery bit
	sigWithRecovery := append([]byte{recoveryBit + 27}, e.Bytes()...)
	publicKey, _, err := ecdsa.RecoverCompact(sigWithRecovery, messageHash)
	if err != nil {
		return nil, err
	}
	return &Secp256k1PublicKey{Inner: publicKey}, nil
}

// Secp256k1Authenticator combines a public key and signature for verification.
// This is used internally by SingleKeyAuthenticator.
//
// Implements:
//   - [bcs.Struct]
type Secp256k1Authenticator struct {
	PubKey *Secp256k1PublicKey
	Sig    *Secp256k1Signature
}

// PublicKey returns the public key as a VerifyingKey.
func (ea *Secp256k1Authenticator) PublicKey() VerifyingKey {
	return ea.PubKey
}

// Signature returns the signature.
func (ea *Secp256k1Authenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the signature is valid for the message.
func (ea *Secp256k1Authenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
func (ea *Secp256k1Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PubKey)
	ser.Struct(ea.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
func (ea *Secp256k1Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &Secp256k1PublicKey{}
	des.Struct(ea.PubKey)
	if des.Error() != nil {
		return
	}
	ea.Sig = &Secp256k1Signature{}
	des.Struct(ea.Sig)
}
