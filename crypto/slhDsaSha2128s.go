package crypto

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/cloudflare/circl/sign/slhdsa"
)

// SLH-DSA-SHA2-128s (FIPS 205, AIP-137) sizes.
const (
	// SlhDsaSha2128sPublicKeyLength is the public key length: PK seed (16) || PK root (16).
	SlhDsaSha2128sPublicKeyLength = 32
	// SlhDsaSha2128sPrivateKeyLength is the serialized private key length, matching the
	// TypeScript SDK: SK seed (16) || PRF seed (16) || PK seed (16). The PK root is
	// derived from these seeds and is not serialized.
	SlhDsaSha2128sPrivateKeyLength = 48
	// SlhDsaSha2128sSecretKeyLength is the full FIPS 205 secret key length
	// (the 48 serialized bytes plus the derived 16-byte PK root). Accepted on import
	// for compatibility with libraries that serialize the full secret key.
	SlhDsaSha2128sSecretKeyLength = 64
	// SlhDsaSha2128sSignatureLength is the signature length for the 128s parameter set.
	SlhDsaSha2128sSignatureLength = 7856
)

// Aliases using the shorter names already used by the v2 module.
const (
	SlhDsaPublicKeySize  = SlhDsaSha2128sPublicKeyLength
	SlhDsaPrivateKeySize = SlhDsaSha2128sPrivateKeyLength
	SlhDsaSignatureSize  = SlhDsaSha2128sSignatureLength
)

// region SlhDsaSha2128sPrivateKey

// SlhDsaSha2128sPrivateKey is an SLH-DSA-SHA2-128s private key. It is a
// post-quantum (stateless hash-based) signing key and, like [Secp256k1PrivateKey],
// cannot stand on its own: wrap it with [NewSingleSigner] to sign transactions.
//
// Its serialized form ([SlhDsaSha2128sPrivateKey.Bytes]) is the 48-byte
// SK seed || PRF seed || PK seed triple used by the TypeScript SDK.
//
// Implements:
//   - [MessageSigner]
//   - [CryptoMaterial]
type SlhDsaSha2128sPrivateKey struct {
	Inner slhdsa.PrivateKey // Inner is the underlying FIPS 205 private key
}

// SlhDsaPrivateKey is an alias for [SlhDsaSha2128sPrivateKey].
type SlhDsaPrivateKey = SlhDsaSha2128sPrivateKey

// GenerateSlhDsaSha2128sPrivateKey generates a random [SlhDsaSha2128sPrivateKey].
// An optional [io.Reader] may be supplied as the randomness source; it must
// provide at least 48 bytes, consumed as SK seed || PRF seed || PK seed.
func GenerateSlhDsaSha2128sPrivateKey(randomness ...io.Reader) (*SlhDsaSha2128sPrivateKey, error) {
	var r io.Reader
	if len(randomness) > 0 {
		r = randomness[0]
	}
	_, priv, err := slhdsa.GenerateKey(r, slhdsa.SHA2_128s)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SLH-DSA-SHA2-128s key: %w", err)
	}
	return &SlhDsaSha2128sPrivateKey{Inner: priv}, nil
}

// GenerateSlhDsaPrivateKey is an alias for [GenerateSlhDsaSha2128sPrivateKey].
func GenerateSlhDsaPrivateKey(randomness ...io.Reader) (*SlhDsaSha2128sPrivateKey, error) {
	return GenerateSlhDsaSha2128sPrivateKey(randomness...)
}

// NewSlhDsaSha2128sPrivateKeyFromSeeds creates a key from its 48-byte serialized
// form (SK seed || PRF seed || PK seed), deriving the PK root.
func NewSlhDsaSha2128sPrivateKeyFromSeeds(threeSeeds []byte) (*SlhDsaSha2128sPrivateKey, error) {
	key := &SlhDsaSha2128sPrivateKey{}
	if err := key.FromBytes(threeSeeds); err != nil {
		return nil, err
	}
	return key, nil
}

func slhDsaPrivateKeyFromThreeSeeds(threeSeeds []byte) (slhdsa.PrivateKey, error) {
	// circl's GenerateKey consumes SK seed, PRF seed and PK seed in order,
	// exactly the 48-byte serialized form, and derives the PK root.
	_, priv, err := slhdsa.GenerateKey(bytes.NewReader(threeSeeds), slhdsa.SHA2_128s)
	if err != nil {
		return slhdsa.PrivateKey{}, fmt.Errorf("failed to derive SLH-DSA-SHA2-128s key from seeds: %w", err)
	}
	return priv, nil
}

// region SlhDsaSha2128sPrivateKey MessageSigner implementation

// SignMessage signs a message and returns the raw [SlhDsaSha2128sSignature].
// Signing is deterministic (FIPS 205 deterministic variant), so the same key
// and message always produce the same signature.
//
// Implements:
//   - [MessageSigner]
func (key *SlhDsaSha2128sPrivateKey) SignMessage(msg []byte) (Signature, error) {
	sigBytes, err := slhdsa.SignDeterministic(&key.Inner, slhdsa.NewMessage(msg), nil)
	if err != nil {
		return nil, fmt.Errorf("SLH-DSA-SHA2-128s signing failed: %w", err)
	}
	sig := &SlhDsaSha2128sSignature{}
	if err := sig.FromBytes(sigBytes); err != nil {
		return nil, err
	}
	return sig, nil
}

// EmptySignature creates an empty signature for use in simulation
//
// Implements:
//   - [MessageSigner]
func (key *SlhDsaSha2128sPrivateKey) EmptySignature() Signature {
	return &SlhDsaSha2128sSignature{}
}

// VerifyingKey returns the [SlhDsaSha2128sPublicKey] associated with the private key
//
// Implements:
//   - [MessageSigner]
func (key *SlhDsaSha2128sPrivateKey) VerifyingKey() VerifyingKey {
	return key.PublicKey()
}

// PublicKey returns the [SlhDsaSha2128sPublicKey] associated with the private key.
func (key *SlhDsaSha2128sPrivateKey) PublicKey() *SlhDsaSha2128sPublicKey {
	return &SlhDsaSha2128sPublicKey{Inner: key.Inner.PublicKey()}
}

// endregion

// region SlhDsaSha2128sPrivateKey CryptoMaterial implementation

// Bytes returns the 48-byte serialized private key (SK seed || PRF seed || PK seed)
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPrivateKey) Bytes() []byte {
	full, err := key.Inner.MarshalBinary()
	if err != nil || len(full) < SlhDsaSha2128sPrivateKeyLength {
		return nil
	}
	out := make([]byte, SlhDsaSha2128sPrivateKeyLength)
	copy(out, full[:SlhDsaSha2128sPrivateKeyLength])
	return out
}

// FromBytes sets the private key from its 48-byte serialized form. A 64-byte
// full FIPS 205 secret key is also accepted if its last 16 bytes (PK root)
// match the root derived from the first 48 bytes.
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPrivateKey) FromBytes(input []byte) error {
	parsed, err := ParsePrivateKey(input, PrivateKeyVariantSlhDsaSha2128s, false)
	if err != nil {
		return err
	}
	switch len(parsed) {
	case SlhDsaSha2128sPrivateKeyLength:
		priv, err := slhDsaPrivateKeyFromThreeSeeds(parsed)
		if err != nil {
			return err
		}
		key.Inner = priv
		return nil
	case SlhDsaSha2128sSecretKeyLength:
		priv, err := slhDsaPrivateKeyFromThreeSeeds(parsed[:SlhDsaSha2128sPrivateKeyLength])
		if err != nil {
			return err
		}
		full, err := priv.MarshalBinary()
		if err != nil {
			return err
		}
		if !bytes.Equal(full, parsed) {
			return errors.New("invalid SLH-DSA-SHA2-128s secret key: PK root does not match the seeds")
		}
		key.Inner = priv
		return nil
	default:
		return fmt.Errorf("invalid SLH-DSA-SHA2-128s private key size %d, expected %d (or %d for a full secret key)", len(parsed), SlhDsaSha2128sPrivateKeyLength, SlhDsaSha2128sSecretKeyLength)
	}
}

// ToHex returns the hex string representation of the private key, with a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex sets the private key from a hex string, with or without a leading 0x,
// or an AIP-80 string ("slh-dsa-sha2-128s-priv-0x...").
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPrivateKey) FromHex(hexStr string) error {
	parsed, err := ParsePrivateKey(hexStr, PrivateKeyVariantSlhDsaSha2128s)
	if err != nil {
		return err
	}
	return key.FromBytes(parsed)
}

// ToAIP80 formats the private key to an AIP-80 compliant string
func (key *SlhDsaSha2128sPrivateKey) ToAIP80() (string, error) {
	return FormatPrivateKey(key.ToHex(), PrivateKeyVariantSlhDsaSha2128s)
}

// String returns a redacted string representation of the private key
func (key *SlhDsaSha2128sPrivateKey) String() string {
	return "<redacted SlhDsaSha2128sPrivateKey>"
}

// Clear zeros out the private key material from memory.
func (key *SlhDsaSha2128sPrivateKey) Clear() {
	key.Inner = slhdsa.PrivateKey{ID: slhdsa.SHA2_128s}
}

// endregion
// endregion

// region SlhDsaSha2128sPublicKey

// SlhDsaSha2128sPublicKey is the public key for [SlhDsaSha2128sPrivateKey].
// On its own its authentication key is the SingleKey authentication key of
// its [AnyPublicKey] form, matching how it is used on-chain.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type SlhDsaSha2128sPublicKey struct {
	Inner slhdsa.PublicKey // Inner is the underlying FIPS 205 public key
}

// SlhDsaPublicKey is an alias for [SlhDsaSha2128sPublicKey].
type SlhDsaPublicKey = SlhDsaSha2128sPublicKey

// Verify verifies a message with the public key and [Signature].
// Returns false if the signature is not a [SlhDsaSha2128sSignature], or verification fails.
//
// Implements:
//   - [VerifyingKey]
func (key *SlhDsaSha2128sPublicKey) Verify(msg []byte, sig Signature) bool {
	switch sig := sig.(type) {
	case *SlhDsaSha2128sSignature:
		return slhdsa.Verify(&key.Inner, slhdsa.NewMessage(msg), sig.Bytes(), nil)
	default:
		return false
	}
}

// AuthKey returns the [AuthenticationKey] for this key as a SingleKey account:
// SHA3-256(BCS(AnyPublicKey) || SingleKeyScheme).
//
// Implements:
//   - [PublicKey]
func (key *SlhDsaSha2128sPublicKey) AuthKey() *AuthenticationKey {
	return key.ToAnyPublicKey().AuthKey()
}

// Scheme returns [SingleKeyScheme], as SLH-DSA keys are only used via SingleKey.
//
// Implements:
//   - [PublicKey]
func (key *SlhDsaSha2128sPublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

// ToAnyPublicKey converts the public key to its [AnyPublicKey] form (variant 5).
func (key *SlhDsaSha2128sPublicKey) ToAnyPublicKey() *AnyPublicKey {
	return &AnyPublicKey{Variant: AnyPublicKeyVariantSlhDsaSha2128s, PubKey: key}
}

// Bytes returns the raw 32-byte public key
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPublicKey) Bytes() []byte {
	out, err := key.Inner.MarshalBinary()
	if err != nil {
		return nil
	}
	return out
}

// FromBytes sets the public key from raw bytes
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPublicKey) FromBytes(input []byte) error {
	if len(input) != SlhDsaSha2128sPublicKeyLength {
		return fmt.Errorf("invalid SLH-DSA-SHA2-128s public key size %d, expected %d", len(input), SlhDsaSha2128sPublicKeyLength)
	}
	key.Inner = slhdsa.PublicKey{ID: slhdsa.SHA2_128s}
	if err := key.Inner.UnmarshalBinary(input); err != nil {
		return fmt.Errorf("failed to parse SLH-DSA-SHA2-128s public key: %w", err)
	}
	return nil
}

// ToHex returns the hex string representation of the public key, with a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex sets the public key from a hex string, with or without a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (key *SlhDsaSha2128sPublicKey) FromHex(hexStr string) error {
	parsed, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(parsed)
}

// MarshalBCS serializes the public key to BCS bytes (ULEB128 length + raw key)
//
// Implements:
//   - [bcs.Marshaler]
func (key *SlhDsaSha2128sPublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}

// UnmarshalBCS deserializes the public key from BCS bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (key *SlhDsaSha2128sPublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	kb := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := key.FromBytes(kb); err != nil {
		des.SetError(err)
	}
}

// endregion

// region SlhDsaSha2128sSignature

// SlhDsaSha2128sSignature is an SLH-DSA-SHA2-128s signature (7,856 bytes)
//
// Implements:
//   - [Signature]
//   - [CryptoMaterial]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type SlhDsaSha2128sSignature struct {
	Inner [SlhDsaSha2128sSignatureLength]byte // Inner is the actual signature
}

// SlhDsaSignature is an alias for [SlhDsaSha2128sSignature].
type SlhDsaSignature = SlhDsaSha2128sSignature

// ToAnySignature converts the signature to its [AnySignature] form (variant 4).
func (s *SlhDsaSha2128sSignature) ToAnySignature() *AnySignature {
	return &AnySignature{Variant: AnySignatureVariantSlhDsaSha2128s, Signature: s}
}

// Bytes returns the raw bytes of the signature
//
// Implements:
//   - [CryptoMaterial]
func (s *SlhDsaSha2128sSignature) Bytes() []byte {
	return s.Inner[:]
}

// FromBytes sets the signature from raw bytes
//
// Implements:
//   - [CryptoMaterial]
func (s *SlhDsaSha2128sSignature) FromBytes(input []byte) error {
	if len(input) != SlhDsaSha2128sSignatureLength {
		return fmt.Errorf("invalid SLH-DSA-SHA2-128s signature size %d, expected %d", len(input), SlhDsaSha2128sSignatureLength)
	}
	copy(s.Inner[:], input)
	return nil
}

// ToHex returns the hex string representation of the signature, with a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (s *SlhDsaSha2128sSignature) ToHex() string {
	return util.BytesToHex(s.Bytes())
}

// FromHex sets the signature from a hex string, with or without a leading 0x
//
// Implements:
//   - [CryptoMaterial]
func (s *SlhDsaSha2128sSignature) FromHex(hexStr string) error {
	parsed, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return s.FromBytes(parsed)
}

// MarshalBCS serializes the signature to BCS bytes (ULEB128 length + raw signature)
//
// Implements:
//   - [bcs.Marshaler]
func (s *SlhDsaSha2128sSignature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(s.Bytes())
}

// UnmarshalBCS deserializes the signature from BCS bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (s *SlhDsaSha2128sSignature) UnmarshalBCS(des *bcs.Deserializer) {
	b := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if err := s.FromBytes(b); err != nil {
		des.SetError(err)
	}
}

// endregion

// region SlhDsaSha2128sAuthenticator

// SlhDsaSha2128sAuthenticator pairs an [SlhDsaSha2128sPublicKey] with an
// [SlhDsaSha2128sSignature]. On-chain this pair is carried by
// [SingleKeyAuthenticator] via [AnyPublicKey] / [AnySignature]; this type is
// provided for direct verification, mirroring [Secp256k1Authenticator].
//
// Implements:
//   - [AccountAuthenticatorImpl]
//   - [bcs.Marshaler]
//   - [bcs.Unmarshaler]
//   - [bcs.Struct]
type SlhDsaSha2128sAuthenticator struct {
	PubKey *SlhDsaSha2128sPublicKey // PubKey is the public key
	Sig    *SlhDsaSha2128sSignature // Sig is the signature
}

// SlhDsaAuthenticator is an alias for [SlhDsaSha2128sAuthenticator].
type SlhDsaAuthenticator = SlhDsaSha2128sAuthenticator

// PublicKey returns the public key of the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *SlhDsaSha2128sAuthenticator) PublicKey() PublicKey {
	return ea.PubKey
}

// Signature returns the signature of the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *SlhDsaSha2128sAuthenticator) Signature() Signature {
	return ea.Sig
}

// Verify returns true if the authenticator can be cryptographically verified
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *SlhDsaSha2128sAuthenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

// MarshalBCS serializes the authenticator to BCS bytes
//
// Implements:
//   - [bcs.Marshaler]
func (ea *SlhDsaSha2128sAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PubKey)
	ser.Struct(ea.Sig)
}

// UnmarshalBCS deserializes the authenticator from BCS bytes
//
// Implements:
//   - [bcs.Unmarshaler]
func (ea *SlhDsaSha2128sAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &SlhDsaSha2128sPublicKey{}
	des.Struct(ea.PubKey)
	if des.Error() != nil {
		return
	}
	ea.Sig = &SlhDsaSha2128sSignature{}
	des.Struct(ea.Sig)
}

// endregion

var (
	_ MessageSigner            = (*SlhDsaSha2128sPrivateKey)(nil)
	_ CryptoMaterial           = (*SlhDsaSha2128sPrivateKey)(nil)
	_ VerifyingKey             = (*SlhDsaSha2128sPublicKey)(nil)
	_ PublicKey                = (*SlhDsaSha2128sPublicKey)(nil)
	_ CryptoMaterial           = (*SlhDsaSha2128sPublicKey)(nil)
	_ bcs.Struct               = (*SlhDsaSha2128sPublicKey)(nil)
	_ Signature                = (*SlhDsaSha2128sSignature)(nil)
	_ CryptoMaterial           = (*SlhDsaSha2128sPublicKey)(nil)
	_ AccountAuthenticatorImpl = (*SlhDsaSha2128sAuthenticator)(nil)
	_ bcs.Struct               = (*SlhDsaSha2128sAuthenticator)(nil)
)
