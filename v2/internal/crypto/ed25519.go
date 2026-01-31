package crypto

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
	"github.com/hdevalence/ed25519consensus"
)

// Ed25519PrivateKey is an Ed25519 private key that can sign transactions.
//
// Ed25519PrivateKey implements [Signer] directly (unlike Secp256k1 which needs
// to be wrapped with SingleSigner). This uses the legacy Ed25519Scheme for
// backward compatibility.
//
// Ed25519PrivateKey is safe for concurrent use. Cached values are protected by a mutex.
//
// Implements:
//   - [Signer]
//   - [MessageSigner]
//   - [CryptoMaterial]
type Ed25519PrivateKey struct {
	Inner ed25519.PrivateKey

	// mu protects cached values for concurrent access
	mu sync.RWMutex
	// Cached values to avoid repeated derivation
	cachedPubKey  *Ed25519PublicKey
	cachedAuthKey *AuthenticationKey
}

// GenerateEd25519PrivateKey generates a new random Ed25519 key pair.
//
// An optional [io.Reader] can be provided for deterministic key generation
// (useful for testing). The reader must provide exactly 32 bytes.
func GenerateEd25519PrivateKey(rand ...io.Reader) (*Ed25519PrivateKey, error) {
	var priv ed25519.PrivateKey
	var err error
	if len(rand) > 0 {
		_, priv, err = ed25519.GenerateKey(rand[0])
	} else {
		_, priv, err = ed25519.GenerateKey(nil)
	}
	if err != nil {
		return nil, err
	}
	return &Ed25519PrivateKey{Inner: priv}, nil
}

// Sign signs a message and returns an AccountAuthenticator.
//
// Implements [Signer].
func (key *Ed25519PrivateKey) Sign(msg []byte) (*AccountAuthenticator, error) {
	sig, err := key.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	ed25519Sig, ok := sig.(*Ed25519Signature)
	if !ok {
		return nil, errors.New("expected Ed25519Signature")
	}

	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	if !ok {
		return nil, errors.New("failed to get Ed25519PublicKey")
	}
	return &AccountAuthenticator{
		Variant: AccountAuthenticatorEd25519,
		Auth: &Ed25519Authenticator{
			PubKey: pubKey,
			Sig:    ed25519Sig,
		},
	}, nil
}

// SignMessage signs a message and returns the raw signature.
//
// Implements [Signer] and [MessageSigner].
func (key *Ed25519PrivateKey) SignMessage(msg []byte) (Signature, error) {
	sigBytes := ed25519.Sign(key.Inner, msg)
	return &Ed25519Signature{Inner: [64]byte(sigBytes)}, nil
}

// SimulationAuthenticator creates an authenticator with an empty signature for simulation.
//
// Implements [Signer].
func (key *Ed25519PrivateKey) SimulationAuthenticator() *AccountAuthenticator {
	pubKey, ok := key.PubKey().(*Ed25519PublicKey)
	if !ok {
		return nil
	}
	return &AccountAuthenticator{
		Variant: AccountAuthenticatorEd25519,
		Auth: &Ed25519Authenticator{
			PubKey: pubKey,
			Sig:    &Ed25519Signature{},
		},
	}
}

// AuthKey returns the AuthenticationKey derived from this key.
// The result is cached after the first call. Thread-safe.
//
// Uses double-checked locking pattern which is safe in Go because:
//  1. RLock/RUnlock provides happens-before relationship for reads
//  2. Lock/Unlock provides happens-before relationship for writes
//  3. The double-check after Lock prevents races where another goroutine
//     computed the value between RUnlock and Lock
//
// Implements [Signer].
func (key *Ed25519PrivateKey) AuthKey() *AuthenticationKey {
	// Fast path: check if already cached (read lock)
	key.mu.RLock()
	if key.cachedAuthKey != nil {
		cached := key.cachedAuthKey
		key.mu.RUnlock()
		return cached
	}
	key.mu.RUnlock()

	// Slow path: compute and cache (write lock)
	key.mu.Lock()
	defer key.mu.Unlock()
	// Double-check: another goroutine may have computed while we waited for lock
	if key.cachedAuthKey != nil {
		return key.cachedAuthKey
	}
	out := &AuthenticationKey{}
	out.FromPublicKey(key.pubKeyLocked())
	key.cachedAuthKey = out
	return out
}

// PubKey returns the Ed25519PublicKey for signature verification.
// The result is cached after the first call to avoid repeated derivation. Thread-safe.
//
// Implements [Signer].
func (key *Ed25519PrivateKey) PubKey() PublicKey {
	key.mu.RLock()
	if key.cachedPubKey != nil {
		cached := key.cachedPubKey
		key.mu.RUnlock()
		return cached
	}
	key.mu.RUnlock()

	key.mu.Lock()
	defer key.mu.Unlock()
	return key.pubKeyLocked()
}

// pubKeyLocked returns the public key, must be called with mu held.
func (key *Ed25519PrivateKey) pubKeyLocked() *Ed25519PublicKey {
	if key.cachedPubKey != nil {
		return key.cachedPubKey
	}
	pubKey, ok := key.Inner.Public().(ed25519.PublicKey)
	if !ok {
		return nil
	}
	key.cachedPubKey = &Ed25519PublicKey{Inner: pubKey}
	return key.cachedPubKey
}

// EmptySignature returns an empty signature for simulation.
//
// Implements [MessageSigner].
func (key *Ed25519PrivateKey) EmptySignature() Signature {
	return &Ed25519Signature{}
}

// VerifyingKey returns the public key for verification.
//
// Implements [MessageSigner].
func (key *Ed25519PrivateKey) VerifyingKey() VerifyingKey {
	return key.PubKey()
}

// Bytes returns the seed bytes of the private key.
//
// Implements [CryptoMaterial].
func (key *Ed25519PrivateKey) Bytes() []byte {
	return key.Inner.Seed()
}

// FromBytes loads the private key from seed bytes.
// This clears any cached public key and authentication key. Thread-safe.
//
// Implements [CryptoMaterial].
func (key *Ed25519PrivateKey) FromBytes(bytes []byte) error {
	bytes, err := ParsePrivateKey(bytes, PrivateKeyVariantEd25519, false)
	if err != nil {
		return err
	}
	if len(bytes) != ed25519.SeedSize {
		return fmt.Errorf("ed25519 private key must be %d bytes, got %d", ed25519.SeedSize, len(bytes))
	}

	key.mu.Lock()
	defer key.mu.Unlock()
	key.Inner = ed25519.NewKeyFromSeed(bytes)
	// Clear cached values since key changed
	key.cachedPubKey = nil
	key.cachedAuthKey = nil
	return nil
}

// ToHex returns the hex representation of the seed with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *Ed25519PrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the private key.
//
// Implements [CryptoMaterial].
func (key *Ed25519PrivateKey) FromHex(hexStr string) error {
	bytes, err := ParsePrivateKey(hexStr, PrivateKeyVariantEd25519)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// ToAIP80 formats the private key as an AIP-80 compliant string.
func (key *Ed25519PrivateKey) ToAIP80() (string, error) {
	return FormatPrivateKey(key.ToHex(), PrivateKeyVariantEd25519)
}

// String returns a redacted representation to prevent accidental logging of private keys.
// Use ToAIP80() to get the actual private key string.
func (key *Ed25519PrivateKey) String() string {
	return "<Ed25519PrivateKey:REDACTED>"
}

// Ed25519PublicKey is an Ed25519 public key for signature verification.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type Ed25519PublicKey struct {
	Inner ed25519.PublicKey
}

// Verify verifies a signature against a message.
//
// Implements [VerifyingKey].
func (key *Ed25519PublicKey) Verify(msg []byte, sig Signature) bool {
	ed25519Sig, ok := sig.(*Ed25519Signature)
	if !ok {
		return false
	}
	return ed25519consensus.Verify(key.Inner, msg, ed25519Sig.Bytes())
}

// AuthKey returns the AuthenticationKey for this public key.
//
// Implements [PublicKey].
func (key *Ed25519PublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns Ed25519Scheme.
//
// Implements [PublicKey].
func (key *Ed25519PublicKey) Scheme() uint8 {
	return Ed25519Scheme
}

// Bytes returns the raw public key bytes.
//
// Implements [CryptoMaterial].
func (key *Ed25519PublicKey) Bytes() []byte {
	return key.Inner[:]
}

// FromBytes loads the public key from bytes.
//
// Implements [CryptoMaterial].
func (key *Ed25519PublicKey) FromBytes(bytes []byte) error {
	if len(bytes) != ed25519.PublicKeySize {
		return fmt.Errorf("ed25519 public key must be %d bytes", ed25519.PublicKeySize)
	}
	key.Inner = bytes
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *Ed25519PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *Ed25519PublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *Ed25519PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Inner)
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *Ed25519PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := key.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// Ed25519Signature is an Ed25519 signature.
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type Ed25519Signature struct {
	Inner [ed25519.SignatureSize]byte
}

// Bytes returns the raw signature bytes.
//
// Implements [CryptoMaterial].
func (e *Ed25519Signature) Bytes() []byte {
	return e.Inner[:]
}

// FromBytes loads the signature from bytes.
//
// Implements [CryptoMaterial].
func (e *Ed25519Signature) FromBytes(bytes []byte) error {
	if len(bytes) != ed25519.SignatureSize {
		return fmt.Errorf("ed25519 signature must be %d bytes", ed25519.SignatureSize)
	}
	copy(e.Inner[:], bytes)
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (e *Ed25519Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (e *Ed25519Signature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (e *Ed25519Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(e.Bytes())
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (e *Ed25519Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := e.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// Ed25519Authenticator combines a public key and signature for verification.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type Ed25519Authenticator struct {
	PubKey *Ed25519PublicKey
	Sig    *Ed25519Signature
}

// PublicKey returns the public key.
//
// Implements [AccountAuthenticatorImpl].
func (ea *Ed25519Authenticator) PublicKey() PublicKey {
	return ea.PubKey
}

// Signature returns the signature.
//
// Implements [AccountAuthenticatorImpl].
func (ea *Ed25519Authenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the signature is valid for the message.
//
// Implements [AccountAuthenticatorImpl].
func (ea *Ed25519Authenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (ea *Ed25519Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PubKey)
	ser.Struct(ea.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (ea *Ed25519Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &Ed25519PublicKey{}
	des.Struct(ea.PubKey)
	if des.Error() != nil {
		return
	}
	ea.Sig = &Ed25519Signature{}
	des.Struct(ea.Sig)
}
