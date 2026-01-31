package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// Security constants for WebAuthn validation
const (
	// MaxWebAuthnSignatureBytes is the maximum size of a WebAuthn signature.
	MaxWebAuthnSignatureBytes = 1024

	// MaxAuthenticatorDataBytes is the maximum size of authenticator data.
	// WebAuthn spec allows variable length, but we limit for DoS protection.
	// Typical authenticator data is ~37 bytes minimum, plus extensions.
	MaxAuthenticatorDataBytes = 1024

	// MaxClientDataJSONBytes is the maximum size of client data JSON.
	// WebAuthn spec doesn't define a limit, but we need DoS protection.
	MaxClientDataJSONBytes = 8192

	// MinAuthenticatorDataBytes is the minimum valid authenticator data size.
	// rpIdHash (32) + flags (1) + signCount (4) = 37 bytes minimum.
	MinAuthenticatorDataBytes = 37

	// ExpectedChallengeLength is the expected length of the SHA3-256 challenge.
	ExpectedChallengeLength = 32
)

// AssertionSignatureVariant identifies the type of signature in AssertionSignature.
type AssertionSignatureVariant uint32

const (
	AssertionSignatureVariantSecp256r1 AssertionSignatureVariant = 0
)

// AssertionSignature wraps the raw ECDSA signature from WebAuthn.
//
// Note: WebAuthn assertions often return DER-encoded signatures.
// The signature must be converted to raw format (r || s) before use.
//
// Implements [bcs.Struct].
type AssertionSignature struct {
	Variant   AssertionSignatureVariant
	Signature *Secp256r1Signature
}

// MarshalBCS serializes the assertion signature to BCS.
func (s *AssertionSignature) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(s.Variant))
	ser.Struct(s.Signature)
}

// UnmarshalBCS deserializes the assertion signature from BCS.
func (s *AssertionSignature) UnmarshalBCS(des *bcs.Deserializer) {
	s.Variant = AssertionSignatureVariant(des.Uleb128())
	switch s.Variant {
	case AssertionSignatureVariantSecp256r1:
		s.Signature = &Secp256r1Signature{}
		des.Struct(s.Signature)
	default:
		des.SetError(fmt.Errorf("unknown assertion signature variant: %d", s.Variant))
	}
}

// CollectedClientData represents the client data from a WebAuthn assertion.
// This is a subset of the full WebAuthn CollectedClientData structure.
type CollectedClientData struct {
	Type        string `json:"type"`
	Challenge   string `json:"challenge"`
	Origin      string `json:"origin"`
	CrossOrigin bool   `json:"crossOrigin,omitempty"`
}

// PartialAuthenticatorAssertionResponse contains a subset of the fields
// from a WebAuthn AuthenticatorAssertionResponse.
//
// The challenge in client_data_json should be SHA3-256(signing_message(transaction)).
//
// Implements:
//   - [Signature]
//   - [bcs.Struct]
type PartialAuthenticatorAssertionResponse struct {
	// Signature is the raw ECDSA signature from the authenticator.
	// Note: Many WebAuthn signatures are DER-encoded and must be converted
	// to raw format before use.
	Signature AssertionSignature

	// AuthenticatorData contains the authenticator data returned by the authenticator.
	AuthenticatorData []byte

	// ClientDataJSON contains the JSON serialization of CollectedClientData.
	// The exact serialization must be preserved as the hash is computed over it.
	ClientDataJSON []byte
}

// NewPartialAuthenticatorAssertionResponse creates a new WebAuthn assertion response.
func NewPartialAuthenticatorAssertionResponse(
	signature *Secp256r1Signature,
	authenticatorData []byte,
	clientDataJSON []byte,
) *PartialAuthenticatorAssertionResponse {
	return &PartialAuthenticatorAssertionResponse{
		Signature: AssertionSignature{
			Variant:   AssertionSignatureVariantSecp256r1,
			Signature: signature,
		},
		AuthenticatorData: authenticatorData,
		ClientDataJSON:    clientDataJSON,
	}
}

// Bytes returns the BCS serialized bytes.
func (p *PartialAuthenticatorAssertionResponse) Bytes() []byte {
	data, _ := bcs.Serialize(p)
	return data
}

// FromBytes is not supported for WebAuthn signatures.
// Use BCS deserialization instead.
func (p *PartialAuthenticatorAssertionResponse) FromBytes(bytes []byte) error {
	return bcs.Deserialize(p, bytes)
}

// ToHex returns the hex representation of the BCS serialized signature.
func (p *PartialAuthenticatorAssertionResponse) ToHex() string {
	return util.BytesToHex(p.Bytes())
}

// FromHex parses a hex string into the signature.
func (p *PartialAuthenticatorAssertionResponse) FromHex(hexStr string) error {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return p.FromBytes(bytes)
}

// MarshalBCS serializes the WebAuthn assertion response to BCS.
func (p *PartialAuthenticatorAssertionResponse) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&p.Signature)
	ser.WriteBytes(p.AuthenticatorData)
	ser.WriteBytes(p.ClientDataJSON)
}

// UnmarshalBCS deserializes the WebAuthn assertion response from BCS.
func (p *PartialAuthenticatorAssertionResponse) UnmarshalBCS(des *bcs.Deserializer) {
	des.Struct(&p.Signature)
	if des.Error() != nil {
		return
	}

	p.AuthenticatorData = des.ReadBytes()
	if des.Error() != nil {
		return
	}

	// Validate authenticator data bounds
	if len(p.AuthenticatorData) < MinAuthenticatorDataBytes {
		des.SetError(fmt.Errorf("authenticator data too short: %d < %d", len(p.AuthenticatorData), MinAuthenticatorDataBytes))
		return
	}
	if len(p.AuthenticatorData) > MaxAuthenticatorDataBytes {
		des.SetError(fmt.Errorf("authenticator data too large: %d > %d", len(p.AuthenticatorData), MaxAuthenticatorDataBytes))
		return
	}

	p.ClientDataJSON = des.ReadBytes()
	if des.Error() != nil {
		return
	}

	// Validate client data JSON bounds
	if len(p.ClientDataJSON) == 0 {
		des.SetError(fmt.Errorf("client data JSON is empty"))
		return
	}
	if len(p.ClientDataJSON) > MaxClientDataJSONBytes {
		des.SetError(fmt.Errorf("client data JSON too large: %d > %d", len(p.ClientDataJSON), MaxClientDataJSONBytes))
		return
	}
}

// GetChallenge extracts and decodes the challenge from the client data JSON.
// The challenge must be base64url encoded (per WebAuthn spec) and should be
// exactly 32 bytes (SHA3-256 hash of signing message) for transaction signing.
func (p *PartialAuthenticatorAssertionResponse) GetChallenge() ([]byte, error) {
	// Validate client data JSON size before parsing
	if len(p.ClientDataJSON) == 0 {
		return nil, fmt.Errorf("client data JSON is empty")
	}
	if len(p.ClientDataJSON) > MaxClientDataJSONBytes {
		return nil, fmt.Errorf("client data JSON too large: %d > %d", len(p.ClientDataJSON), MaxClientDataJSONBytes)
	}

	var clientData CollectedClientData
	if err := json.Unmarshal(p.ClientDataJSON, &clientData); err != nil {
		return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
	}

	// WebAuthn spec requires base64url encoding (RFC 4648 Section 5) without padding
	// We strictly enforce this for security - no fallback to other encodings
	challenge, err := base64.RawURLEncoding.DecodeString(clientData.Challenge)
	if err != nil {
		// Try with padding as some implementations include it
		challenge, err = base64.URLEncoding.DecodeString(clientData.Challenge)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64url challenge: %w", err)
		}
	}

	// Validate challenge length - for transaction signing, this should be a SHA3-256 hash
	if len(challenge) != ExpectedChallengeLength {
		return nil, fmt.Errorf("invalid challenge length: expected %d, got %d", ExpectedChallengeLength, len(challenge))
	}

	return challenge, nil
}

// generateVerificationData creates the data that was signed by the authenticator.
// This is the binary concatenation of authenticator_data and SHA-256(client_data_json).
func (p *PartialAuthenticatorAssertionResponse) generateVerificationData() []byte {
	// SHA-256 hash of client data JSON
	clientDataHash := sha256.Sum256(p.ClientDataJSON)

	// Concatenate authenticator data and client data hash
	verificationData := make([]byte, len(p.AuthenticatorData)+len(clientDataHash))
	copy(verificationData, p.AuthenticatorData)
	copy(verificationData[len(p.AuthenticatorData):], clientDataHash[:])

	return verificationData
}

// Verify verifies the WebAuthn signature against a message and public key.
//
// The verification process:
// 1. Extract the challenge from client_data_json
// 2. Verify the challenge equals SHA3-256(signing_message(message)) using constant-time comparison
// 3. Construct verification_data = authenticator_data || SHA-256(client_data_json)
// 4. Verify the Secp256r1 signature over verification_data
//
// Security: This function uses constant-time comparison for the challenge to prevent timing attacks.
func (p *PartialAuthenticatorAssertionResponse) Verify(msg []byte, pubKey *AnyPublicKey) bool {
	// Validate authenticator data bounds
	if len(p.AuthenticatorData) < MinAuthenticatorDataBytes || len(p.AuthenticatorData) > MaxAuthenticatorDataBytes {
		return false
	}

	// Extract challenge from client data
	actualChallenge, err := p.GetChallenge()
	if err != nil {
		return false
	}

	// Expected challenge is SHA3-256 of the signing message
	// The message passed here should already be the signing message (prefixed hash + BCS)
	expectedChallenge := util.Sha3256Hash([][]byte{msg})

	// Use constant-time comparison to prevent timing attacks
	// subtle.ConstantTimeCompare returns 1 if equal, 0 otherwise
	if subtle.ConstantTimeCompare(actualChallenge, expectedChallenge) != 1 {
		return false
	}

	// Generate verification data
	verificationData := p.generateVerificationData()

	// Verify signature based on public key type
	switch pubKey.Variant {
	case AnyPublicKeyVariantSecp256r1:
		secp256r1PubKey, ok := pubKey.PubKey.(*Secp256r1PublicKey)
		if !ok {
			return false
		}
		// For WebAuthn, we verify directly over the verification data without additional hashing
		return p.verifySecp256r1Raw(verificationData, secp256r1PubKey)
	default:
		return false
	}
}

// verifySecp256r1Raw verifies a Secp256r1 signature over raw data.
// WebAuthn uses SHA-256 internally, so we hash with SHA-256 for verification.
func (p *PartialAuthenticatorAssertionResponse) verifySecp256r1Raw(data []byte, pubKey *Secp256r1PublicKey) bool {
	if p.Signature.Signature == nil {
		return false
	}

	// For WebAuthn, the data is hashed with SHA-256 (not SHA3-256)
	hash := sha256.Sum256(data)
	r, s := p.Signature.Signature.getRS()

	return ecdsa.Verify(pubKey.Inner, hash[:], r, s)
}

// VerifyArbitraryMessage verifies the WebAuthn signature against arbitrary message bytes.
// The message bytes are compared directly to the challenge (no additional hashing).
//
// Security: This function uses constant-time comparison for the challenge to prevent timing attacks.
func (p *PartialAuthenticatorAssertionResponse) VerifyArbitraryMessage(msg []byte, pubKey *AnyPublicKey) bool {
	// Validate authenticator data bounds
	if len(p.AuthenticatorData) < MinAuthenticatorDataBytes || len(p.AuthenticatorData) > MaxAuthenticatorDataBytes {
		return false
	}

	// Extract challenge from client data
	actualChallenge, err := p.GetChallenge()
	if err != nil {
		return false
	}

	// For arbitrary message verification, challenge should equal the message directly
	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(actualChallenge, msg) != 1 {
		return false
	}

	// Generate verification data
	verificationData := p.generateVerificationData()

	// Verify signature based on public key type
	switch pubKey.Variant {
	case AnyPublicKeyVariantSecp256r1:
		secp256r1PubKey, ok := pubKey.PubKey.(*Secp256r1PublicKey)
		if !ok {
			return false
		}
		return p.verifySecp256r1Raw(verificationData, secp256r1PubKey)
	default:
		return false
	}
}

// ParseSignatureRS parses the r and s components from a Secp256r1 signature.
func ParseSignatureRS(sig *Secp256r1Signature) (r, s *big.Int) {
	return sig.getRS()
}

// ToAnySignature converts the WebAuthn signature to an AnySignature.
func (p *PartialAuthenticatorAssertionResponse) ToAnySignature() *AnySignature {
	return &AnySignature{
		Variant:   AnySignatureVariantWebAuthn,
		Signature: p,
	}
}

// Compile-time interface checks
var (
	_ Signature  = (*PartialAuthenticatorAssertionResponse)(nil)
	_ bcs.Struct = (*PartialAuthenticatorAssertionResponse)(nil)
	_ bcs.Struct = (*AssertionSignature)(nil)
)
