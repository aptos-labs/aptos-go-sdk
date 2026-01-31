package crypto

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// Keyless authentication constants
const (
	// PepperNumBytes is the size of the pepper used for identity commitment.
	// This matches poseidon_bn254::keyless::BYTES_PACKED_PER_SCALAR (31 bytes).
	PepperNumBytes = 31

	// IdCommitmentNumBytes is the size of the identity commitment (32 bytes).
	IdCommitmentNumBytes = 32

	// G1ProjectiveCompressedNumBytes is the size of a compressed BN254 G1 point.
	G1ProjectiveCompressedNumBytes = 32

	// G2ProjectiveCompressedNumBytes is the size of a compressed BN254 G2 point.
	G2ProjectiveCompressedNumBytes = 64

	// MaxKeylessPublicKeyLen is the maximum length of a keyless public key.
	MaxKeylessPublicKeyLen = 200 + IdCommitmentNumBytes

	// MaxKeylessSignatureLen is the maximum length of a keyless signature.
	MaxKeylessSignatureLen = 4000

	// EpkBlinderNumBytes is the size of the EPK blinder.
	EpkBlinderNumBytes = 31
)

// Pepper is used to create a hiding identity commitment (IDC) when deriving a keyless address.
type Pepper [PepperNumBytes]byte

// MarshalBCS serializes the pepper to BCS.
func (p *Pepper) MarshalBCS(ser *bcs.Serializer) {
	ser.FixedBytes(p[:])
}

// UnmarshalBCS deserializes the pepper from BCS.
func (p *Pepper) UnmarshalBCS(des *bcs.Deserializer) {
	des.ReadFixedBytesInto(p[:])
}

// IdCommitment is a SNARK-friendly commitment to the user's identity.
// It commits to: H(pepper || aud_val || uid_val || uid_key)
type IdCommitment struct {
	inner []byte
}

// NewIdCommitment creates a new identity commitment from bytes.
func NewIdCommitment(bytes []byte) (*IdCommitment, error) {
	if len(bytes) != IdCommitmentNumBytes {
		return nil, fmt.Errorf("invalid identity commitment length: expected %d, got %d", IdCommitmentNumBytes, len(bytes))
	}
	return &IdCommitment{inner: bytes}, nil
}

// Bytes returns the raw bytes of the identity commitment.
func (idc *IdCommitment) Bytes() []byte {
	return idc.inner
}

// MarshalBCS serializes the identity commitment to BCS.
func (idc *IdCommitment) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(idc.inner)
}

// UnmarshalBCS deserializes the identity commitment from BCS.
func (idc *IdCommitment) UnmarshalBCS(des *bcs.Deserializer) {
	idc.inner = des.ReadBytes()
	if des.Error() != nil {
		return
	}
	if len(idc.inner) != IdCommitmentNumBytes {
		des.SetError(fmt.Errorf("invalid identity commitment length: expected %d, got %d", IdCommitmentNumBytes, len(idc.inner)))
	}
}

// KeylessPublicKey represents a keyless account public key.
// It contains the OIDC issuer and the identity commitment.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [bcs.Struct]
type KeylessPublicKey struct {
	// IssVal is the value of the `iss` field from the JWT, indicating the OIDC provider.
	// e.g., "https://accounts.google.com"
	IssVal string

	// Idc is the identity commitment - a SNARK-friendly commitment to:
	// 1. The application's ID (aud field)
	// 2. The OIDC provider's internal identifier for the user (sub or email field)
	Idc IdCommitment
}

// Verify is not directly supported for KeylessPublicKey.
// Keyless verification requires additional context (JWK, configuration).
func (key *KeylessPublicKey) Verify(msg []byte, sig Signature) bool {
	// Keyless verification is more complex and requires:
	// 1. The JWK from the OIDC provider
	// 2. The keyless configuration
	// 3. ZK proof verification or OpenID signature verification
	// This is typically done at the blockchain level, not client-side.
	return false
}

// AuthKey returns the AuthenticationKey for this public key.
func (key *KeylessPublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns SingleKeyScheme (Keyless uses AnyPublicKey wrapper).
func (key *KeylessPublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

// Bytes returns the BCS-serialized public key.
func (key *KeylessPublicKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

// FromBytes deserializes from BCS bytes.
func (key *KeylessPublicKey) FromBytes(bytes []byte) error {
	return bcs.Deserialize(key, bytes)
}

// ToHex returns the hex representation with "0x" prefix.
func (key *KeylessPublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
func (key *KeylessPublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the keyless public key to BCS.
func (key *KeylessPublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteString(key.IssVal)
	ser.Struct(&key.Idc)
}

// UnmarshalBCS deserializes the keyless public key from BCS.
func (key *KeylessPublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	key.IssVal = des.ReadString()
	if des.Error() != nil {
		return
	}
	des.Struct(&key.Idc)
}

// FederatedKeylessPublicKey represents a federated keyless account.
// Unlike a normal keyless account, a federated keyless account accepts
// JWKs published at a specific contract address.
//
// Implements:
//   - [VerifyingKey]
//   - [PublicKey]
//   - [bcs.Struct]
type FederatedKeylessPublicKey struct {
	// JwkAddr is the address where the JWKs are published.
	JwkAddr types.AccountAddress

	// Pk is the underlying keyless public key.
	Pk KeylessPublicKey
}

// Verify is not directly supported for FederatedKeylessPublicKey.
func (key *FederatedKeylessPublicKey) Verify(msg []byte, sig Signature) bool {
	return false
}

// AuthKey returns the AuthenticationKey for this public key.
func (key *FederatedKeylessPublicKey) AuthKey() *AuthenticationKey {
	out := &AuthenticationKey{}
	out.FromPublicKey(key)
	return out
}

// Scheme returns SingleKeyScheme.
func (key *FederatedKeylessPublicKey) Scheme() uint8 {
	return SingleKeyScheme
}

// Bytes returns the BCS-serialized public key.
func (key *FederatedKeylessPublicKey) Bytes() []byte {
	val, _ := bcs.Serialize(key)
	return val
}

// FromBytes deserializes from BCS bytes.
func (key *FederatedKeylessPublicKey) FromBytes(bytes []byte) error {
	return bcs.Deserialize(key, bytes)
}

// ToHex returns the hex representation with "0x" prefix.
func (key *FederatedKeylessPublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

// FromHex parses a hex string into the public key.
func (key *FederatedKeylessPublicKey) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

// MarshalBCS serializes the federated keyless public key to BCS.
func (key *FederatedKeylessPublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&key.JwkAddr)
	ser.Struct(&key.Pk)
}

// UnmarshalBCS deserializes the federated keyless public key from BCS.
func (key *FederatedKeylessPublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	des.Struct(&key.JwkAddr)
	if des.Error() != nil {
		return
	}
	des.Struct(&key.Pk)
}

// G1Bytes represents a compressed BN254 G1 point.
type G1Bytes [G1ProjectiveCompressedNumBytes]byte

// MarshalBCS serializes the G1 point to BCS.
func (g *G1Bytes) MarshalBCS(ser *bcs.Serializer) {
	ser.FixedBytes(g[:])
}

// UnmarshalBCS deserializes the G1 point from BCS.
func (g *G1Bytes) UnmarshalBCS(des *bcs.Deserializer) {
	des.ReadFixedBytesInto(g[:])
}

// G2Bytes represents a compressed BN254 G2 point.
type G2Bytes [G2ProjectiveCompressedNumBytes]byte

// MarshalBCS serializes the G2 point to BCS.
func (g *G2Bytes) MarshalBCS(ser *bcs.Serializer) {
	ser.FixedBytes(g[:])
}

// UnmarshalBCS deserializes the G2 point from BCS.
func (g *G2Bytes) UnmarshalBCS(des *bcs.Deserializer) {
	des.ReadFixedBytesInto(g[:])
}

// Groth16Proof represents a Groth16 zero-knowledge proof.
type Groth16Proof struct {
	A G1Bytes
	B G2Bytes
	C G1Bytes
}

// MarshalBCS serializes the Groth16 proof to BCS.
func (p *Groth16Proof) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&p.A)
	ser.Struct(&p.B)
	ser.Struct(&p.C)
}

// UnmarshalBCS deserializes the Groth16 proof from BCS.
func (p *Groth16Proof) UnmarshalBCS(des *bcs.Deserializer) {
	des.Struct(&p.A)
	if des.Error() != nil {
		return
	}
	des.Struct(&p.B)
	if des.Error() != nil {
		return
	}
	des.Struct(&p.C)
}

// ZKPVariant identifies the type of zero-knowledge proof.
type ZKPVariant uint32

const (
	ZKPVariantGroth16 ZKPVariant = 0
)

// ZKP wraps different zero-knowledge proof types.
type ZKP struct {
	Variant ZKPVariant
	Proof   *Groth16Proof
}

// MarshalBCS serializes the ZKP to BCS.
func (z *ZKP) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(z.Variant))
	ser.Struct(z.Proof)
}

// UnmarshalBCS deserializes the ZKP from BCS.
func (z *ZKP) UnmarshalBCS(des *bcs.Deserializer) {
	z.Variant = ZKPVariant(des.Uleb128())
	switch z.Variant {
	case ZKPVariantGroth16:
		z.Proof = &Groth16Proof{}
		des.Struct(z.Proof)
	default:
		des.SetError(fmt.Errorf("unknown ZKP variant: %d", z.Variant))
	}
}

// ZeroKnowledgeSig contains a ZK proof for keyless authentication.
type ZeroKnowledgeSig struct {
	// Proof is the zero-knowledge proof.
	Proof ZKP

	// ExpHorizonSecs is the expiration horizon the circuit enforces.
	ExpHorizonSecs uint64

	// ExtraField is an optional extra field matched publicly in the JWT.
	ExtraField *string

	// OverrideAudVal allows users to recover keyless accounts bound to
	// an application that is no longer online.
	OverrideAudVal *string

	// TrainingWheelsSignature is a signature to mitigate against circuit flaws.
	TrainingWheelsSignature *EphemeralSignature
}

// MarshalBCS serializes the ZeroKnowledgeSig to BCS.
func (z *ZeroKnowledgeSig) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&z.Proof)
	ser.U64(z.ExpHorizonSecs)

	// Optional extra field
	if z.ExtraField != nil {
		ser.Bool(true)
		ser.WriteString(*z.ExtraField)
	} else {
		ser.Bool(false)
	}

	// Optional override aud val
	if z.OverrideAudVal != nil {
		ser.Bool(true)
		ser.WriteString(*z.OverrideAudVal)
	} else {
		ser.Bool(false)
	}

	// Optional training wheels signature
	if z.TrainingWheelsSignature != nil {
		ser.Bool(true)
		ser.Struct(z.TrainingWheelsSignature)
	} else {
		ser.Bool(false)
	}
}

// UnmarshalBCS deserializes the ZeroKnowledgeSig from BCS.
func (z *ZeroKnowledgeSig) UnmarshalBCS(des *bcs.Deserializer) {
	des.Struct(&z.Proof)
	if des.Error() != nil {
		return
	}

	z.ExpHorizonSecs = des.U64()

	// Optional extra field
	if des.Bool() {
		s := des.ReadString()
		z.ExtraField = &s
	}

	// Optional override aud val
	if des.Bool() {
		s := des.ReadString()
		z.OverrideAudVal = &s
	}

	// Optional training wheels signature
	if des.Bool() {
		z.TrainingWheelsSignature = &EphemeralSignature{}
		des.Struct(z.TrainingWheelsSignature)
	}
}

// OpenIdSig contains an OpenID signature for keyless authentication.
type OpenIdSig struct {
	// JwtSig is the decoded bytes of the JWS signature in the JWT.
	JwtSig []byte

	// JwtPayloadJSON is the decoded/plaintext JSON payload of the JWT.
	JwtPayloadJSON string

	// UidKey is the name of the key in the claim that maps to the user identifier.
	// e.g., "sub" or "email"
	UidKey string

	// EpkBlinder is the random value used to obfuscate the EPK from OIDC providers.
	EpkBlinder []byte

	// Pepper is the privacy-preserving value used to calculate the identity commitment.
	Pepper Pepper

	// IdcAudVal is set when an override aud_val is used.
	IdcAudVal *string
}

// MarshalBCS serializes the OpenIdSig to BCS.
func (o *OpenIdSig) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(o.JwtSig)
	ser.WriteString(o.JwtPayloadJSON)
	ser.WriteString(o.UidKey)
	ser.WriteBytes(o.EpkBlinder)
	ser.Struct(&o.Pepper)

	// Optional idc_aud_val
	if o.IdcAudVal != nil {
		ser.Bool(true)
		ser.WriteString(*o.IdcAudVal)
	} else {
		ser.Bool(false)
	}
}

// UnmarshalBCS deserializes the OpenIdSig from BCS.
func (o *OpenIdSig) UnmarshalBCS(des *bcs.Deserializer) {
	o.JwtSig = des.ReadBytes()
	if des.Error() != nil {
		return
	}

	o.JwtPayloadJSON = des.ReadString()
	if des.Error() != nil {
		return
	}

	o.UidKey = des.ReadString()
	if des.Error() != nil {
		return
	}

	o.EpkBlinder = des.ReadBytes()
	if des.Error() != nil {
		return
	}

	des.Struct(&o.Pepper)
	if des.Error() != nil {
		return
	}

	// Optional idc_aud_val
	if des.Bool() {
		s := des.ReadString()
		o.IdcAudVal = &s
	}
}

// EphemeralCertificateVariant identifies the type of ephemeral certificate.
type EphemeralCertificateVariant uint32

const (
	EphemeralCertificateVariantZeroKnowledge EphemeralCertificateVariant = 0
	EphemeralCertificateVariantOpenId        EphemeralCertificateVariant = 1
)

// EphemeralCertificate is a certificate binding the ephemeral public key
// to the keyless account. It's either a ZK proof or an OpenID signature.
type EphemeralCertificate struct {
	Variant EphemeralCertificateVariant
	Cert    bcs.Struct // Either *ZeroKnowledgeSig or *OpenIdSig
}

// MarshalBCS serializes the ephemeral certificate to BCS.
func (e *EphemeralCertificate) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(e.Variant))
	ser.Struct(e.Cert)
}

// UnmarshalBCS deserializes the ephemeral certificate from BCS.
func (e *EphemeralCertificate) UnmarshalBCS(des *bcs.Deserializer) {
	e.Variant = EphemeralCertificateVariant(des.Uleb128())
	switch e.Variant {
	case EphemeralCertificateVariantZeroKnowledge:
		cert := &ZeroKnowledgeSig{}
		des.Struct(cert)
		e.Cert = cert
	case EphemeralCertificateVariantOpenId:
		cert := &OpenIdSig{}
		des.Struct(cert)
		e.Cert = cert
	default:
		des.SetError(fmt.Errorf("unknown ephemeral certificate variant: %d", e.Variant))
	}
}

// EphemeralSignatureVariant identifies the type of ephemeral signature.
type EphemeralSignatureVariant uint32

const (
	EphemeralSignatureVariantEd25519  EphemeralSignatureVariant = 0
	EphemeralSignatureVariantWebAuthn EphemeralSignatureVariant = 1
)

// EphemeralSignature is a signature under an ephemeral key.
type EphemeralSignature struct {
	Variant   EphemeralSignatureVariant
	Signature Signature
}

// MarshalBCS serializes the ephemeral signature to BCS.
func (e *EphemeralSignature) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(e.Variant))
	ser.Struct(e.Signature)
}

// UnmarshalBCS deserializes the ephemeral signature from BCS.
func (e *EphemeralSignature) UnmarshalBCS(des *bcs.Deserializer) {
	e.Variant = EphemeralSignatureVariant(des.Uleb128())
	switch e.Variant {
	case EphemeralSignatureVariantEd25519:
		sig := &Ed25519Signature{}
		des.Struct(sig)
		e.Signature = sig
	case EphemeralSignatureVariantWebAuthn:
		sig := &PartialAuthenticatorAssertionResponse{}
		des.Struct(sig)
		e.Signature = sig
	default:
		des.SetError(fmt.Errorf("unknown ephemeral signature variant: %d", e.Variant))
	}
}

// EphemeralPublicKeyVariant identifies the type of ephemeral public key.
type EphemeralPublicKeyVariant uint32

const (
	EphemeralPublicKeyVariantEd25519   EphemeralPublicKeyVariant = 0
	EphemeralPublicKeyVariantSecp256r1 EphemeralPublicKeyVariant = 1
)

// EphemeralPublicKey is a short-lived public key for keyless authentication.
type EphemeralPublicKey struct {
	Variant EphemeralPublicKeyVariant
	PubKey  VerifyingKey
}

// MarshalBCS serializes the ephemeral public key to BCS.
func (e *EphemeralPublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(e.Variant))
	ser.Struct(e.PubKey)
}

// UnmarshalBCS deserializes the ephemeral public key from BCS.
func (e *EphemeralPublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	e.Variant = EphemeralPublicKeyVariant(des.Uleb128())
	switch e.Variant {
	case EphemeralPublicKeyVariantEd25519:
		pk := &Ed25519PublicKey{}
		des.Struct(pk)
		e.PubKey = pk
	case EphemeralPublicKeyVariantSecp256r1:
		pk := &Secp256r1PublicKey{}
		des.Struct(pk)
		e.PubKey = pk
	default:
		des.SetError(fmt.Errorf("unknown ephemeral public key variant: %d", e.Variant))
	}
}

// Bytes returns the BCS-serialized ephemeral public key.
func (e *EphemeralPublicKey) Bytes() []byte {
	val, _ := bcs.Serialize(e)
	return val
}

// KeylessSignature contains the signature data for keyless authentication.
//
// Implements:
//   - [Signature]
//   - [bcs.Struct]
type KeylessSignature struct {
	// Cert is the ephemeral certificate (ZK proof or OpenID signature).
	Cert EphemeralCertificate

	// JwtHeaderJSON is the decoded/plaintext JWT header.
	// Contains `kid` (key ID) and `alg` (algorithm) fields.
	JwtHeaderJSON string

	// ExpDateSecs is the expiry time of the ephemeral public key as UNIX timestamp.
	ExpDateSecs uint64

	// EphemeralPubkey is the short-lived public key.
	EphemeralPubkey EphemeralPublicKey

	// EphemeralSignature is a signature over the transaction and ZKP.
	EphemeralSignature EphemeralSignature
}

// Bytes returns the BCS serialized bytes.
func (s *KeylessSignature) Bytes() []byte {
	data, _ := bcs.Serialize(s)
	return data
}

// FromBytes deserializes from BCS bytes.
func (s *KeylessSignature) FromBytes(bytes []byte) error {
	return bcs.Deserialize(s, bytes)
}

// ToHex returns the hex representation.
func (s *KeylessSignature) ToHex() string {
	return util.BytesToHex(s.Bytes())
}

// FromHex parses a hex string.
func (s *KeylessSignature) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return s.FromBytes(bytes)
}

// MarshalBCS serializes the keyless signature to BCS.
func (s *KeylessSignature) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&s.Cert)
	ser.WriteString(s.JwtHeaderJSON)
	ser.U64(s.ExpDateSecs)
	ser.Struct(&s.EphemeralPubkey)
	ser.Struct(&s.EphemeralSignature)
}

// UnmarshalBCS deserializes the keyless signature from BCS.
func (s *KeylessSignature) UnmarshalBCS(des *bcs.Deserializer) {
	des.Struct(&s.Cert)
	if des.Error() != nil {
		return
	}

	s.JwtHeaderJSON = des.ReadString()
	if des.Error() != nil {
		return
	}

	s.ExpDateSecs = des.U64()

	des.Struct(&s.EphemeralPubkey)
	if des.Error() != nil {
		return
	}

	des.Struct(&s.EphemeralSignature)
}

// ToAnySignature converts the keyless signature to an AnySignature.
func (s *KeylessSignature) ToAnySignature() *AnySignature {
	return &AnySignature{
		Variant:   AnySignatureVariantKeyless,
		Signature: s,
	}
}

// Compile-time interface checks
var (
	_ VerifyingKey = (*KeylessPublicKey)(nil)
	_ PublicKey    = (*KeylessPublicKey)(nil)
	_ bcs.Struct   = (*KeylessPublicKey)(nil)

	_ VerifyingKey = (*FederatedKeylessPublicKey)(nil)
	_ PublicKey    = (*FederatedKeylessPublicKey)(nil)
	_ bcs.Struct   = (*FederatedKeylessPublicKey)(nil)

	_ Signature  = (*KeylessSignature)(nil)
	_ bcs.Struct = (*KeylessSignature)(nil)

	_ bcs.Struct = (*IdCommitment)(nil)
	_ bcs.Struct = (*Pepper)(nil)
	_ bcs.Struct = (*G1Bytes)(nil)
	_ bcs.Struct = (*G2Bytes)(nil)
	_ bcs.Struct = (*Groth16Proof)(nil)
	_ bcs.Struct = (*ZKP)(nil)
	_ bcs.Struct = (*ZeroKnowledgeSig)(nil)
	_ bcs.Struct = (*OpenIdSig)(nil)
	_ bcs.Struct = (*EphemeralCertificate)(nil)
	_ bcs.Struct = (*EphemeralSignature)(nil)
	_ bcs.Struct = (*EphemeralPublicKey)(nil)
)
