package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// Secp256r1 (P-256/prime256v1) key and signature sizes.
const (
	Secp256r1PrivateKeyLength = 32
	Secp256r1PublicKeyLength  = 65 // Uncompressed format: 0x04 + 32 bytes X + 32 bytes Y
	Secp256r1SignatureLength  = 64 // r (32 bytes) || s (32 bytes)
)

// Secp256r1PrivateKey is a Secp256r1 (P-256) private key.
//
// Secp256r1 is commonly used with WebAuthn/passkeys and is widely supported
// by hardware security modules.
//
// Unlike Ed25519PrivateKey, Secp256r1PrivateKey does not implement [Signer] directly.
// It must be wrapped with [SingleSigner] for transaction signing:
//
//	key, _ := crypto.GenerateSecp256r1Key()
//	signer := crypto.NewSingleSigner(key)
//	auth, _ := signer.Sign(txnBytes)
//
// Implements:
//   - [MessageSigner]
//   - [CryptoMaterial]
type Secp256r1PrivateKey struct {
	Inner *ecdsa.PrivateKey
}

// GenerateSecp256r1Key generates a new random Secp256r1 (P-256) key pair.
func GenerateSecp256r1Key() (*Secp256r1PrivateKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Secp256r1 key: %w", err)
	}
	return &Secp256r1PrivateKey{Inner: priv}, nil
}

// SignMessage signs a message and returns the signature.
// The message is hashed with SHA3-256 before signing.
//
// Implements [MessageSigner].
func (key *Secp256r1PrivateKey) SignMessage(msg []byte) (Signature, error) {
	hash := util.Sha3256Hash([][]byte{msg})

	r, s, err := ecdsa.Sign(rand.Reader, key.Inner, hash)
	if err != nil {
		return nil, fmt.Errorf("Secp256r1 signing failed: %w", err)
	}

	// Normalize s to low order (required by Aptos)
	s = normalizeS(s, elliptic.P256().Params().N)

	sig := &Secp256r1Signature{}
	sig.setRS(r, s)
	return sig, nil
}

// EmptySignature returns an empty signature for simulation.
//
// Implements [MessageSigner].
func (key *Secp256r1PrivateKey) EmptySignature() Signature {
	return &Secp256r1Signature{}
}

// VerifyingKey returns the public key for verification.
//
// Implements [MessageSigner].
func (key *Secp256r1PrivateKey) VerifyingKey() VerifyingKey {
	return &Secp256r1PublicKey{Inner: &key.Inner.PublicKey}
}

// Bytes returns the raw private key bytes (32 bytes).
//
// Implements [CryptoMaterial].
func (key *Secp256r1PrivateKey) Bytes() []byte {
	bytes := key.Inner.D.Bytes()
	// Pad to 32 bytes if needed
	if len(bytes) < Secp256r1PrivateKeyLength {
		padded := make([]byte, Secp256r1PrivateKeyLength)
		copy(padded[Secp256r1PrivateKeyLength-len(bytes):], bytes)
		return padded
	}
	return bytes
}

// FromBytes loads the private key from bytes.
//
// Implements [CryptoMaterial].
func (key *Secp256r1PrivateKey) FromBytes(bytes []byte) error {
	bytes, err := ParsePrivateKey(bytes, PrivateKeyVariantSecp256r1, false)
	if err != nil {
		return err
	}
	if len(bytes) != Secp256r1PrivateKeyLength {
		return fmt.Errorf("secp256r1 private key must be %d bytes, got %d", Secp256r1PrivateKeyLength, len(bytes))
	}

	curve := elliptic.P256()
	d := new(big.Int).SetBytes(bytes)

	// Validate that d is in valid range
	if d.Cmp(big.NewInt(1)) < 0 || d.Cmp(curve.Params().N) >= 0 {
		return errors.New("secp256r1 private key out of range")
	}

	// Derive public key
	x, y := curve.ScalarBaseMult(bytes)

	key.Inner = &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *Secp256r1PrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the private key.
//
// Implements [CryptoMaterial].
func (key *Secp256r1PrivateKey) FromHex(hexStr string) error {
	bytes, err := ParsePrivateKey(hexStr, PrivateKeyVariantSecp256r1)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// ToAIP80 formats the private key as an AIP-80 compliant string.
func (key *Secp256r1PrivateKey) ToAIP80() (string, error) {
	return FormatPrivateKey(key.ToHex(), PrivateKeyVariantSecp256r1)
}

// String returns a redacted representation to prevent accidental logging.
func (key *Secp256r1PrivateKey) String() string {
	return "<Secp256r1PrivateKey:REDACTED>"
}

// Secp256r1PublicKey is a Secp256r1 (P-256) public key for signature verification.
//
// Note: Secp256r1PublicKey cannot be used directly as an on-chain PublicKey.
// It must be wrapped with [AnyPublicKey] via [SingleSigner].
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type Secp256r1PublicKey struct {
	Inner *ecdsa.PublicKey
}

// AuthKey returns the AuthenticationKey for this public key.
//
// For Secp256r1 keys used with SingleKey, the authentication key is derived
// from the BCS-serialized AnyPublicKey with the SingleKeyScheme.
//
// Implements [PublicKey].
func (key *Secp256r1PublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns SingleKeyScheme since Secp256r1 is only used with SingleKey.
//
// Implements [PublicKey].
func (key *Secp256r1PublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

// Verify verifies a signature against a message.
// The message is hashed with SHA3-256 before verification.
//
// Implements [VerifyingKey].
func (key *Secp256r1PublicKey) Verify(msg []byte, sig Signature) bool {
	r1Sig, ok := sig.(*Secp256r1Signature)
	if !ok {
		return false
	}

	hash := util.Sha3256Hash([][]byte{msg})
	r, s := r1Sig.getRS()

	return ecdsa.Verify(key.Inner, hash, r, s)
}

// Bytes returns the uncompressed public key bytes (65 bytes: 0x04 + X + Y).
//
// Implements [CryptoMaterial].
func (key *Secp256r1PublicKey) Bytes() []byte {
	return elliptic.Marshal(key.Inner.Curve, key.Inner.X, key.Inner.Y)
}

// FromBytes loads the public key from uncompressed bytes.
//
// Implements [CryptoMaterial].
func (key *Secp256r1PublicKey) FromBytes(bytes []byte) error {
	if len(bytes) != Secp256r1PublicKeyLength {
		return fmt.Errorf("secp256r1 public key must be %d bytes, got %d", Secp256r1PublicKeyLength, len(bytes))
	}
	if bytes[0] != 0x04 {
		return errors.New("secp256r1 public key must be in uncompressed format (starting with 0x04)")
	}

	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, bytes)
	if x == nil || y == nil {
		return errors.New("invalid secp256r1 public key")
	}

	key.Inner = &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (key *Secp256r1PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
//
// Implements [CryptoMaterial].
func (key *Secp256r1PublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the public key to BCS.
//
// Implements [bcs.Marshaler].
func (key *Secp256r1PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}

// UnmarshalBCS deserializes the public key from BCS.
//
// Implements [bcs.Unmarshaler].
func (key *Secp256r1PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := key.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// Secp256r1Signature is a Secp256r1 ECDSA signature (r || s, 64 bytes).
//
// Note: For WebAuthn usage, this signature would be wrapped in a WebAuthn
// assertion structure. This type represents the raw ECDSA signature.
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Struct]
type Secp256r1Signature struct {
	Inner [Secp256r1SignatureLength]byte
}

// Bytes returns the signature bytes (r || s, 64 bytes total).
//
// Implements [CryptoMaterial].
func (s *Secp256r1Signature) Bytes() []byte {
	return s.Inner[:]
}

// FromBytes loads the signature from bytes.
//
// Implements [CryptoMaterial].
func (s *Secp256r1Signature) FromBytes(bytes []byte) error {
	if len(bytes) != Secp256r1SignatureLength {
		return fmt.Errorf("secp256r1 signature must be %d bytes, got %d", Secp256r1SignatureLength, len(bytes))
	}

	// Validate r and s are in valid range
	r := new(big.Int).SetBytes(bytes[0:32])
	sVal := new(big.Int).SetBytes(bytes[32:64])
	n := elliptic.P256().Params().N

	if r.Cmp(big.NewInt(1)) < 0 || r.Cmp(n) >= 0 {
		return errors.New("secp256r1 signature: r out of range")
	}
	if sVal.Cmp(big.NewInt(1)) < 0 || sVal.Cmp(n) >= 0 {
		return errors.New("secp256r1 signature: s out of range")
	}

	// Check that s is in low order (required by Aptos)
	halfN := new(big.Int).Rsh(n, 1)
	if sVal.Cmp(halfN) > 0 {
		return errors.New("secp256r1 signature: s is over half order")
	}

	copy(s.Inner[:], bytes)
	return nil
}

// ToHex returns the hex representation with "0x" prefix.
//
// Implements [CryptoMaterial].
func (s *Secp256r1Signature) ToHex() string {
	return util.BytesToHex(s.Bytes())
}

// FromHex parses a hex string into the signature.
//
// Implements [CryptoMaterial].
func (s *Secp256r1Signature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return s.FromBytes(bytes)
}

// MarshalBCS serializes the signature to BCS.
//
// Implements [bcs.Marshaler].
func (s *Secp256r1Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(s.Bytes())
}

// UnmarshalBCS deserializes the signature from BCS.
//
// Implements [bcs.Unmarshaler].
func (s *Secp256r1Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := s.FromBytes(bytes); err != nil {
		des.SetError(err)
	}
}

// setRS sets the r and s values in the signature.
func (s *Secp256r1Signature) setRS(r, sVal *big.Int) {
	rBytes := r.Bytes()
	sBytes := sVal.Bytes()

	// Pad r to 32 bytes
	copy(s.Inner[32-len(rBytes):32], rBytes)
	// Pad s to 32 bytes
	copy(s.Inner[64-len(sBytes):64], sBytes)
}

// getRS extracts r and s from the signature.
func (s *Secp256r1Signature) getRS() (*big.Int, *big.Int) {
	r := new(big.Int).SetBytes(s.Inner[0:32])
	sVal := new(big.Int).SetBytes(s.Inner[32:64])
	return r, sVal
}

// normalizeS ensures s is in the lower half of the curve order.
// This prevents signature malleability.
func normalizeS(s *big.Int, n *big.Int) *big.Int {
	halfN := new(big.Int).Rsh(n, 1)
	if s.Cmp(halfN) > 0 {
		return new(big.Int).Sub(n, s)
	}
	return s
}

// Secp256r1Authenticator combines a public key and signature for verification.
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Struct]
type Secp256r1Authenticator struct {
	PubKey *Secp256r1PublicKey
	Sig    *Secp256r1Signature
}

// PublicKey returns the public key.
//
// Implements [AccountAuthenticatorImpl].
func (auth *Secp256r1Authenticator) PublicKey() PublicKey {
	return auth.PubKey
}

// Signature returns the signature.
//
// Implements [AccountAuthenticatorImpl].
func (auth *Secp256r1Authenticator) Signature() Signature {
	return auth.Sig
}

// Verify returns true if the signature is valid for the message.
//
// Implements [AccountAuthenticatorImpl].
func (auth *Secp256r1Authenticator) Verify(msg []byte) bool {
	return auth.PubKey.Verify(msg, auth.Sig)
}

// MarshalBCS serializes the authenticator to BCS.
//
// Implements [bcs.Marshaler].
func (auth *Secp256r1Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(auth.PubKey)
	ser.Struct(auth.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS.
//
// Implements [bcs.Unmarshaler].
func (auth *Secp256r1Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	auth.PubKey = &Secp256r1PublicKey{}
	des.Struct(auth.PubKey)
	if des.Error() != nil {
		return
	}
	auth.Sig = &Secp256r1Signature{}
	des.Struct(auth.Sig)
}

// ToAnyPublicKey converts the Secp256r1 public key to an AnyPublicKey.
func (key *Secp256r1PublicKey) ToAnyPublicKey() *AnyPublicKey {
	return &AnyPublicKey{
		Variant: AnyPublicKeyVariantSecp256r1,
		PubKey:  key,
	}
}

// NewSecp256r1SingleSigner creates a SingleSigner wrapping a Secp256r1 private key.
func NewSecp256r1SingleSigner(key *Secp256r1PrivateKey) *SingleSigner {
	return NewSingleSigner(key)
}

// Compile-time interface checks
var (
	_ MessageSigner            = (*Secp256r1PrivateKey)(nil)
	_ CryptoMaterial           = (*Secp256r1PrivateKey)(nil)
	_ VerifyingKey             = (*Secp256r1PublicKey)(nil)
	_ PublicKey                = (*Secp256r1PublicKey)(nil)
	_ CryptoMaterial           = (*Secp256r1PublicKey)(nil)
	_ bcs.Struct               = (*Secp256r1PublicKey)(nil)
	_ Signature                = (*Secp256r1Signature)(nil)
	_ CryptoMaterial           = (*Secp256r1Signature)(nil)
	_ bcs.Struct               = (*Secp256r1Signature)(nil)
	_ AccountAuthenticatorImpl = (*Secp256r1Authenticator)(nil)
	_ bcs.Struct               = (*Secp256r1Authenticator)(nil)
)
