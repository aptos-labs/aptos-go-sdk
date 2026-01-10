package keyless

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// Common errors
var (
	// ErrJWTExpired indicates the JWT has expired.
	ErrJWTExpired = errors.New("keyless: JWT has expired")

	// ErrInvalidJWT indicates the JWT is malformed or invalid.
	ErrInvalidJWT = errors.New("keyless: invalid JWT format")

	// ErrMissingClaim indicates a required claim is missing from the JWT.
	ErrMissingClaim = errors.New("keyless: missing required claim")

	// ErrProverFailed indicates the prover service returned an error.
	ErrProverFailed = errors.New("keyless: prover service failed")

	// ErrEphemeralKeyExpired indicates the ephemeral key has expired.
	ErrEphemeralKeyExpired = errors.New("keyless: ephemeral key has expired")

	// ErrUnsupportedProvider indicates the OIDC provider is not supported.
	ErrUnsupportedProvider = errors.New("keyless: unsupported OIDC provider")
)

// OIDCProvider represents a supported OIDC identity provider.
type OIDCProvider string

const (
	ProviderGoogle   OIDCProvider = "google"
	ProviderApple    OIDCProvider = "apple"
	ProviderFacebook OIDCProvider = "facebook"
	ProviderDiscord  OIDCProvider = "discord"
)

// ProviderAudiences maps providers to their expected audience values.
var ProviderAudiences = map[OIDCProvider]string{
	ProviderGoogle:   "google",
	ProviderApple:    "apple",
	ProviderFacebook: "facebook",
	ProviderDiscord:  "discord",
}

// EphemeralKeyPair represents a temporary key pair used for keyless authentication.
// The key pair is used to generate ZK proofs and should be kept secure during the session.
type EphemeralKeyPair struct {
	inner     *crypto.Ed25519PrivateKey
	expiresAt time.Time
	blinder   []byte
}

// GenerateEphemeralKeyPair creates a new ephemeral key pair with the specified expiration.
func GenerateEphemeralKeyPair(expiresIn time.Duration) (*EphemeralKeyPair, error) {
	// Generate the underlying Ed25519 key
	privKey, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Generate random blinder (31 bytes for Poseidon commitment)
	blinder := make([]byte, 31)
	if _, err := rand.Read(blinder); err != nil {
		return nil, fmt.Errorf("failed to generate blinder: %w", err)
	}

	return &EphemeralKeyPair{
		inner:     privKey,
		expiresAt: time.Now().Add(expiresIn),
		blinder:   blinder,
	}, nil
}

// Nonce returns the nonce to include in the OIDC authentication request.
// This nonce binds the JWT to this ephemeral key pair.
func (e *EphemeralKeyPair) Nonce() string {
	// Nonce = base64url(ephemeralPublicKey || expiryTimestamp || blinder)
	pubKeyBytes := e.inner.PubKey().Bytes()
	expiryBytes := make([]byte, 8)
	timestamp := uint64(e.expiresAt.Unix())
	for i := 0; i < 8; i++ {
		expiryBytes[i] = byte(timestamp >> (8 * (7 - i)))
	}

	nonceData := append(pubKeyBytes, expiryBytes...)
	nonceData = append(nonceData, e.blinder...)

	return base64.RawURLEncoding.EncodeToString(nonceData)
}

// ExpiresAt returns when this ephemeral key pair expires.
func (e *EphemeralKeyPair) ExpiresAt() time.Time {
	return e.expiresAt
}

// IsExpired returns true if the ephemeral key pair has expired.
func (e *EphemeralKeyPair) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}

// PublicKey returns the public key bytes of the ephemeral key pair.
func (e *EphemeralKeyPair) PublicKey() []byte {
	return e.inner.PubKey().Bytes()
}

// Sign signs a message with the ephemeral private key.
func (e *EphemeralKeyPair) Sign(msg []byte) ([]byte, error) {
	if e.IsExpired() {
		return nil, ErrEphemeralKeyExpired
	}
	sig, err := e.inner.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	return sig.Bytes(), nil
}

// JWTClaims represents the relevant claims from an OIDC JWT.
type JWTClaims struct {
	// Issuer identifies the OIDC provider.
	Issuer string `json:"iss"`

	// Subject is the user's unique identifier at the provider.
	Subject string `json:"sub"`

	// Audience is the client ID.
	Audience interface{} `json:"aud"` // Can be string or []string

	// ExpiresAt is when the JWT expires.
	ExpiresAt int64 `json:"exp"`

	// IssuedAt is when the JWT was issued.
	IssuedAt int64 `json:"iat"`

	// Nonce is the nonce from the authentication request.
	Nonce string `json:"nonce"`

	// Email is the user's email (optional).
	Email string `json:"email,omitempty"`

	// EmailVerified indicates if the email is verified.
	EmailVerified bool `json:"email_verified,omitempty"`
}

// ParseJWT parses a JWT token and extracts the claims.
// Note: This does NOT verify the JWT signature - that's the prover's job.
func ParseJWT(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidJWT
	}

	// Decode the payload (middle part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Try standard base64 with padding
		payload, err = base64.StdEncoding.DecodeString(padBase64(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("%w: failed to decode payload", ErrInvalidJWT)
		}
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("%w: failed to parse claims: %v", ErrInvalidJWT, err)
	}

	// Validate required claims
	if claims.Issuer == "" {
		return nil, fmt.Errorf("%w: iss", ErrMissingClaim)
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("%w: sub", ErrMissingClaim)
	}

	return &claims, nil
}

// GetAudience returns the audience as a string (handles both string and array formats).
func (c *JWTClaims) GetAudience() string {
	switch v := c.Audience.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}
	return ""
}

// IsExpired returns true if the JWT has expired.
func (c *JWTClaims) IsExpired() bool {
	return time.Now().Unix() > c.ExpiresAt
}

// KeylessAccount represents an account derived from keyless authentication.
type KeylessAccount struct {
	address          aptos.AccountAddress
	ephemeralKeyPair *EphemeralKeyPair
	jwt              string
	claims           *JWTClaims
	proof            *ZKProof
	uidKey           string
	uidVal           string
	pepper           []byte
}

// ZKProof represents a zero-knowledge proof for keyless authentication.
type ZKProof struct {
	A       []byte `json:"a"`
	B       []byte `json:"b"`
	C       []byte `json:"c"`
	Variant string `json:"variant"`
}

// DeriveAccountConfig contains configuration for deriving a keyless account.
type DeriveAccountConfig struct {
	// JWT is the OIDC JWT token from the identity provider.
	JWT string

	// EphemeralKeyPair is the ephemeral key pair generated for this session.
	EphemeralKeyPair *EphemeralKeyPair

	// UIDKey is the claim key to use for identity (default: "sub").
	UIDKey string

	// Pepper is optional additional randomness for address derivation.
	// If not provided, it will be fetched from the pepper service.
	Pepper []byte
}

// DeriveAddress derives the keyless account address from JWT claims.
// This can be called before having the full proof to preview the address.
func DeriveAddress(claims *JWTClaims, uidKey string, pepper []byte) (aptos.AccountAddress, error) {
	if uidKey == "" {
		uidKey = "sub"
	}

	// Get the uid value based on the key
	var uidVal string
	switch uidKey {
	case "sub":
		uidVal = claims.Subject
	case "email":
		uidVal = claims.Email
	default:
		return aptos.AccountAddress{}, fmt.Errorf("unsupported uid key: %s", uidKey)
	}

	if uidVal == "" {
		return aptos.AccountAddress{}, fmt.Errorf("%w: %s", ErrMissingClaim, uidKey)
	}

	// Derive address: sha3_256(iss || uid_key || uid_val || aud || pepper || KEYLESS_SCHEME)
	addressInput := [][]byte{
		[]byte(claims.Issuer),
		[]byte(uidKey),
		[]byte(uidVal),
		[]byte(claims.GetAudience()),
		pepper,
		{crypto.SingleKeyScheme},
	}

	addressHash := util.Sha3256Hash(addressInput)

	var addr aptos.AccountAddress
	copy(addr[:], addressHash)
	return addr, nil
}

// Address returns the account address.
func (k *KeylessAccount) Address() aptos.AccountAddress {
	return k.address
}

// Sign signs a message using the keyless account.
// This creates a keyless signature using the ZK proof.
func (k *KeylessAccount) Sign(msg []byte) (*crypto.AccountAuthenticator, error) {
	if k.ephemeralKeyPair.IsExpired() {
		return nil, ErrEphemeralKeyExpired
	}

	if k.claims.IsExpired() {
		return nil, ErrJWTExpired
	}

	// Sign with ephemeral key
	ephSig, err := k.ephemeralKeyPair.Sign(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with ephemeral key: %w", err)
	}

	// Create the keyless signature
	_ = ephSig // Used in the actual keyless signature structure

	// TODO: Implement full keyless signature construction
	// This requires BCS serialization of the keyless signature structure

	return nil, errors.New("keyless signing not fully implemented")
}

// AuthKey returns the authentication key for this keyless account.
func (k *KeylessAccount) AuthKey() *crypto.AuthenticationKey {
	authKey := &crypto.AuthenticationKey{}
	// For keyless accounts, derive auth key from the address components
	// TODO: Implement proper auth key derivation
	copy(authKey[:], k.address[:])
	return authKey
}

// PubKey returns the public key component for this keyless account.
func (k *KeylessAccount) PubKey() crypto.PublicKey {
	// For keyless accounts, this returns the ephemeral public key wrapper
	return k.ephemeralKeyPair.inner.PubKey()
}

// ExpiresAt returns when this keyless account expires (the earlier of JWT or ephemeral key expiry).
func (k *KeylessAccount) ExpiresAt() time.Time {
	jwtExpiry := time.Unix(k.claims.ExpiresAt, 0)
	ephExpiry := k.ephemeralKeyPair.ExpiresAt()

	if jwtExpiry.Before(ephExpiry) {
		return jwtExpiry
	}
	return ephExpiry
}

// IsExpired returns true if the keyless account has expired.
func (k *KeylessAccount) IsExpired() bool {
	return time.Now().After(k.ExpiresAt())
}

// Provider attempts to determine the OIDC provider from the issuer.
func (k *KeylessAccount) Provider() (OIDCProvider, bool) {
	return ProviderFromIssuer(k.claims.Issuer)
}

// ProviderFromIssuer attempts to determine the OIDC provider from an issuer URL.
func ProviderFromIssuer(issuer string) (OIDCProvider, bool) {
	switch {
	case strings.Contains(issuer, "accounts.google.com"):
		return ProviderGoogle, true
	case strings.Contains(issuer, "appleid.apple.com"):
		return ProviderApple, true
	case strings.Contains(issuer, "facebook.com"):
		return ProviderFacebook, true
	case strings.Contains(issuer, "discord.com"):
		return ProviderDiscord, true
	default:
		return "", false
	}
}

// Helper functions

func padBase64(s string) string {
	switch len(s) % 4 {
	case 2:
		return s + "=="
	case 3:
		return s + "="
	}
	return s
}

// GenerateRandomPepper generates a random pepper for address derivation.
func GenerateRandomPepper() ([]byte, error) {
	pepper := make([]byte, 31)
	if _, err := rand.Read(pepper); err != nil {
		return nil, fmt.Errorf("failed to generate pepper: %w", err)
	}
	return pepper, nil
}

// PepperFromBigInt converts a big.Int pepper to bytes.
func PepperFromBigInt(i *big.Int) []byte {
	bytes := i.Bytes()
	// Pad to 31 bytes
	if len(bytes) < 31 {
		padded := make([]byte, 31)
		copy(padded[31-len(bytes):], bytes)
		return padded
	}
	return bytes[:31]
}
