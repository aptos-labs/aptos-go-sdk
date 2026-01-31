package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
	"github.com/cloudflare/circl/sign/slhdsa"
)

// SLH-DSA-SHA2-128s key and signature sizes per FIPS 205.
const (
	SlhDsaPublicKeySize  = 32
	SlhDsaPrivateKeySize = 64
	SlhDsaSignatureSize  = 7856
)

// SlhDsaPrivateKey is an SLH-DSA-SHA2-128s private key that can sign transactions.
//
// SLH-DSA (Stateless Hash-Based Digital Signature Algorithm) is a post-quantum
// signature scheme standardized in FIPS 205. The SHA2-128s variant provides
// NIST Level 1 security with smaller signatures (7,856 bytes).
//
// Note: SlhDsaPrivateKey must be wrapped with [SingleSigner] to implement [Signer].
//
// Implements:
//   - [MessageSigner]
//   - [CryptoMaterial]
type SlhDsaPrivateKey struct {
	Inner slhdsa.PrivateKey
}

// GenerateSlhDsaPrivateKey generates a new random SLH-DSA-SHA2-128s key pair.
//
// An optional [io.Reader] can be provided for deterministic key generation
// (useful for testing).
func GenerateSlhDsaPrivateKey(randReader ...io.Reader) (*SlhDsaPrivateKey, error) {
	var r io.Reader = rand.Reader
	if len(randReader) > 0 {
		r = randReader[0]
	}

	_, priv, err := slhdsa.GenerateKey(r, slhdsa.SHA2_128s)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SLH-DSA key: %w", err)
	}
	return &SlhDsaPrivateKey{Inner: priv}, nil
}

// SignMessage signs a message and returns the raw signature.
//
// Uses deterministic signing for reproducible signatures.
//
// Implements [MessageSigner].
func (key *SlhDsaPrivateKey) SignMessage(msg []byte) (Signature, error) {
	message := slhdsa.NewMessage(msg)
	sigBytes, err := slhdsa.SignDeterministic(&key.Inner, message, nil)
	if err != nil {
		return nil, fmt.Errorf("SLH-DSA signing failed: %w", err)
	}
	sig := &SlhDsaSignature{}
	if err := sig.FromBytes(sigBytes); err != nil {
		return nil, err
	}
	return sig, nil
}

// EmptySignature returns an empty signature for simulation.
//
// Implements [MessageSigner].
func (key *SlhDsaPrivateKey) EmptySignature() Signature {
	return &SlhDsaSignature{}
}

// VerifyingKey returns the public key for verification.
//
// Implements [MessageSigner].
func (key *SlhDsaPrivateKey) VerifyingKey() VerifyingKey {
	pubKey := key.Inner.PublicKey()
	return &SlhDsaPublicKey{Inner: pubKey}
}

// Bytes returns the raw private key bytes.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPrivateKey) Bytes() []byte {
	bytes, err := key.Inner.MarshalBinary()
	if err != nil {
		return nil
	}
	return bytes
}

// FromBytes loads the private key from bytes.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPrivateKey) FromBytes(bytes []byte) error {
	bytes, err := ParsePrivateKey(bytes, PrivateKeyVariantSlhDsa, false)
	if err != nil {
		return err
	}
	if len(bytes) != SlhDsaPrivateKeySize {
		return fmt.Errorf("SLH-DSA private key must be %d bytes, got %d", SlhDsaPrivateKeySize, len(bytes))
	}

	key.Inner = slhdsa.PrivateKey{ID: slhdsa.SHA2_128s}
	if err := key.Inner.UnmarshalBinary(bytes); err != nil {
		return fmt.Errorf("failed to parse SLH-DSA private key: %w", err)
	}
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the private key.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPrivateKey) FromHex(hexStr string) error {
	bytes, err := ParsePrivateKey(hexStr, PrivateKeyVariantSlhDsa)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// ToAIP80 formats the private key as an AIP-80 compliant string.
func (key *SlhDsaPrivateKey) ToAIP80() (string, error) {
	return FormatPrivateKey(key.ToHex(), PrivateKeyVariantSlhDsa)
}

// String returns a redacted representation to prevent accidental logging.
func (key *SlhDsaPrivateKey) String() string {
	return "<SlhDsaPrivateKey:REDACTED>"
}

// SlhDsaPublicKey is an SLH-DSA-SHA2-128s public key for signature verification.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type SlhDsaPublicKey struct {
	Inner slhdsa.PublicKey
}

// Verify verifies a signature against a message.
//
// Implements [VerifyingKey].
func (key *SlhDsaPublicKey) Verify(msg []byte, sig Signature) bool {
	slhDsaSig, ok := sig.(*SlhDsaSignature)
	if !ok {
		return false
	}
	message := slhdsa.NewMessage(msg)
	// Verify(key, message, signature, context)
	return slhdsa.Verify(&key.Inner, message, slhDsaSig.Inner[:], nil)
}

// AuthKey returns the AuthenticationKey for this public key.
//
// For SLH-DSA keys used with SingleKey, the authentication key is derived
// from the BCS-serialized AnyPublicKey with the SingleKeyScheme.
//
// Implements [PublicKey].
func (key *SlhDsaPublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns SingleKeyScheme since SLH-DSA is only used with SingleKey.
//
// Implements [PublicKey].
func (key *SlhDsaPublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

// Bytes returns the raw public key bytes.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPublicKey) Bytes() []byte {
	bytes, err := key.Inner.MarshalBinary()
	if err != nil {
		return nil
	}
	return bytes
}

// FromBytes loads the public key from bytes.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPublicKey) FromBytes(bytes []byte) error {
	if len(bytes) != SlhDsaPublicKeySize {
		return fmt.Errorf("SLH-DSA public key must be %d bytes, got %d", SlhDsaPublicKeySize, len(bytes))
	}

	key.Inner = slhdsa.PublicKey{ID: slhdsa.SHA2_128s}
	if err := key.Inner.UnmarshalBinary(bytes); err != nil {
		return fmt.Errorf("failed to parse SLH-DSA public key: %w", err)
	}
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *SlhDsaPublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *SlhDsaPublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *SlhDsaPublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := key.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// SlhDsaSignature is an SLH-DSA-SHA2-128s signature.
//
// Note: SLH-DSA signatures are large (7,856 bytes for SHA2-128s).
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type SlhDsaSignature struct {
	Inner [SlhDsaSignatureSize]byte
}

// Bytes returns the raw signature bytes.
//
// Implements [CryptoMaterial].
func (s *SlhDsaSignature) Bytes() []byte {
	return s.Inner[:]
}

// FromBytes loads the signature from bytes.
//
// Implements [CryptoMaterial].
func (s *SlhDsaSignature) FromBytes(bytes []byte) error {
	if len(bytes) != SlhDsaSignatureSize {
		return fmt.Errorf("SLH-DSA signature must be %d bytes, got %d", SlhDsaSignatureSize, len(bytes))
	}
	copy(s.Inner[:], bytes)
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (s *SlhDsaSignature) ToHex() string {
	return util.BytesToHex(s.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (s *SlhDsaSignature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return s.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (s *SlhDsaSignature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(s.Bytes())
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (s *SlhDsaSignature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := s.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// SlhDsaAuthenticator combines an SLH-DSA public key and signature for verification.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type SlhDsaAuthenticator struct {
	PubKey *SlhDsaPublicKey
	Sig    *SlhDsaSignature
}

// PublicKey returns the public key.
//
// Implements [AccountAuthenticatorImpl].
func (auth *SlhDsaAuthenticator) PublicKey() PublicKey {
	return auth.PubKey
}

// Signature returns the signature.
//
// Implements [AccountAuthenticatorImpl].
func (auth *SlhDsaAuthenticator) Signature() Signature {
	return auth.Sig
}

// Verify returns true if the signature is valid for the message.
//
// Implements [AccountAuthenticatorImpl].
func (auth *SlhDsaAuthenticator) Verify(msg []byte) bool {
	return auth.PubKey.Verify(msg, auth.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (auth *SlhDsaAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(auth.PubKey)
	ser.Struct(auth.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (auth *SlhDsaAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	auth.PubKey = &SlhDsaPublicKey{}
	des.Struct(auth.PubKey)
	if des.Error() != nil {
		return
	}
	auth.Sig = &SlhDsaSignature{}
	des.Struct(auth.Sig)
}

// ToAnyPublicKey converts the SLH-DSA public key to an AnyPublicKey.
func (key *SlhDsaPublicKey) ToAnyPublicKey() *AnyPublicKey {
	return &AnyPublicKey{
		Variant: AnyPublicKeyVariantSlhDsaSha2_128s,
		PubKey:  key,
	}
}

// ToAnySignature converts the SLH-DSA signature to an AnySignature.
func (s *SlhDsaSignature) ToAnySignature() *AnySignature {
	return &AnySignature{
		Variant:   AnySignatureVariantSlhDsaSha2_128s,
		Signature: s,
	}
}

// NewSlhDsaSingleSigner creates a SingleSigner wrapping an SLH-DSA private key.
func NewSlhDsaSingleSigner(key *SlhDsaPrivateKey) *SingleSigner {
	return NewSingleSigner(key)
}

// Compile-time interface checks
var (
	_ MessageSigner            = (*SlhDsaPrivateKey)(nil)
	_ CryptoMaterial           = (*SlhDsaPrivateKey)(nil)
	_ VerifyingKey             = (*SlhDsaPublicKey)(nil)
	_ PublicKey                = (*SlhDsaPublicKey)(nil)
	_ CryptoMaterial           = (*SlhDsaPublicKey)(nil)
	_ bcs.Struct               = (*SlhDsaPublicKey)(nil)
	_ Signature                = (*SlhDsaSignature)(nil)
	_ CryptoMaterial           = (*SlhDsaSignature)(nil)
	_ bcs.Struct               = (*SlhDsaSignature)(nil)
	_ AccountAuthenticatorImpl = (*SlhDsaAuthenticator)(nil)
	_ bcs.Struct               = (*SlhDsaAuthenticator)(nil)
)
