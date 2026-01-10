package keyless

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEphemeralKeyPair(t *testing.T) {
	t.Run("generates valid key pair", func(t *testing.T) {
		kp, err := GenerateEphemeralKeyPair(time.Hour)
		require.NoError(t, err)
		require.NotNil(t, kp)

		// Check public key is 32 bytes (Ed25519)
		assert.Len(t, kp.PublicKey(), 32)

		// Check expiration is set correctly
		assert.WithinDuration(t, time.Now().Add(time.Hour), kp.ExpiresAt(), time.Second)

		// Check not expired
		assert.False(t, kp.IsExpired())
	})

	t.Run("expired key pair", func(t *testing.T) {
		kp, err := GenerateEphemeralKeyPair(-time.Hour) // Already expired
		require.NoError(t, err)
		assert.True(t, kp.IsExpired())
	})
}

func TestEphemeralKeyPair_Nonce(t *testing.T) {
	kp, err := GenerateEphemeralKeyPair(time.Hour)
	require.NoError(t, err)

	nonce := kp.Nonce()

	// Nonce should be non-empty base64url string
	assert.NotEmpty(t, nonce)

	// Should be decodable
	decoded, err := base64.RawURLEncoding.DecodeString(nonce)
	require.NoError(t, err)

	// Should contain: pubkey (32) + expiry (8) + blinder (31) = 71 bytes
	assert.Len(t, decoded, 71)
}

func TestEphemeralKeyPair_Sign(t *testing.T) {
	t.Run("signs successfully", func(t *testing.T) {
		kp, err := GenerateEphemeralKeyPair(time.Hour)
		require.NoError(t, err)

		msg := []byte("hello world")
		sig, err := kp.Sign(msg)
		require.NoError(t, err)

		// Ed25519 signature is 64 bytes
		assert.Len(t, sig, 64)
	})

	t.Run("fails when expired", func(t *testing.T) {
		kp, err := GenerateEphemeralKeyPair(-time.Hour)
		require.NoError(t, err)

		_, err = kp.Sign([]byte("hello"))
		assert.ErrorIs(t, err, ErrEphemeralKeyExpired)
	})
}

func TestParseJWT(t *testing.T) {
	t.Run("parses valid JWT", func(t *testing.T) {
		// Create a test JWT (header.payload.signature)
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{
			"iss": "https://accounts.google.com",
			"sub": "1234567890",
			"aud": "client_id",
			"exp": 9999999999,
			"iat": 1234567890,
			"nonce": "test_nonce",
			"email": "test@example.com",
			"email_verified": true
		}`))
		signature := base64.RawURLEncoding.EncodeToString([]byte("fake_signature"))
		jwt := header + "." + payload + "." + signature

		claims, err := ParseJWT(jwt)
		require.NoError(t, err)

		assert.Equal(t, "https://accounts.google.com", claims.Issuer)
		assert.Equal(t, "1234567890", claims.Subject)
		assert.Equal(t, "client_id", claims.GetAudience())
		assert.Equal(t, int64(9999999999), claims.ExpiresAt)
		assert.Equal(t, "test_nonce", claims.Nonce)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.True(t, claims.EmailVerified)
	})

	t.Run("handles audience array", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{
			"iss": "issuer",
			"sub": "subject",
			"aud": ["client1", "client2"],
			"exp": 9999999999
		}`))
		jwt := header + "." + payload + ".sig"

		claims, err := ParseJWT(jwt)
		require.NoError(t, err)
		assert.Equal(t, "client1", claims.GetAudience())
	})

	t.Run("rejects invalid JWT format", func(t *testing.T) {
		_, err := ParseJWT("not.a.valid.jwt")
		assert.ErrorIs(t, err, ErrInvalidJWT)

		_, err = ParseJWT("invalid")
		assert.ErrorIs(t, err, ErrInvalidJWT)
	})

	t.Run("rejects missing required claims", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub": "test"}`))
		jwt := header + "." + payload + ".sig"

		_, err := ParseJWT(jwt)
		assert.ErrorIs(t, err, ErrMissingClaim)
	})
}

func TestJWTClaims_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		claims := &JWTClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}
		assert.False(t, claims.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		claims := &JWTClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		}
		assert.True(t, claims.IsExpired())
	})
}

func TestDeriveAddress(t *testing.T) {
	claims := &JWTClaims{
		Issuer:   "https://accounts.google.com",
		Subject:  "test_user_id",
		Audience: "test_client_id",
	}

	pepper := make([]byte, 31)

	t.Run("derives consistent address", func(t *testing.T) {
		addr1, err := DeriveAddress(claims, "sub", pepper)
		require.NoError(t, err)

		addr2, err := DeriveAddress(claims, "sub", pepper)
		require.NoError(t, err)

		// Same inputs should produce same address
		assert.Equal(t, addr1, addr2)
	})

	t.Run("different pepper produces different address", func(t *testing.T) {
		addr1, err := DeriveAddress(claims, "sub", pepper)
		require.NoError(t, err)

		differentPepper := make([]byte, 31)
		differentPepper[0] = 1

		addr2, err := DeriveAddress(claims, "sub", differentPepper)
		require.NoError(t, err)

		assert.NotEqual(t, addr1, addr2)
	})

	t.Run("different uid_key produces different address", func(t *testing.T) {
		claims.Email = "test@example.com"

		addr1, err := DeriveAddress(claims, "sub", pepper)
		require.NoError(t, err)

		addr2, err := DeriveAddress(claims, "email", pepper)
		require.NoError(t, err)

		assert.NotEqual(t, addr1, addr2)
	})

	t.Run("rejects unsupported uid_key", func(t *testing.T) {
		_, err := DeriveAddress(claims, "invalid_key", pepper)
		assert.Error(t, err)
	})
}

func TestKeylessAccount_ExpiresAt(t *testing.T) {
	ephKp, _ := GenerateEphemeralKeyPair(time.Hour)
	jwtExpiry := time.Now().Add(30 * time.Minute)

	account := &KeylessAccount{
		ephemeralKeyPair: ephKp,
		claims: &JWTClaims{
			ExpiresAt: jwtExpiry.Unix(),
		},
	}

	// Should return JWT expiry since it's earlier
	expiry := account.ExpiresAt()
	assert.WithinDuration(t, jwtExpiry, expiry, time.Second)
}

func TestProviderFromIssuer(t *testing.T) {
	tests := []struct {
		issuer   string
		expected OIDCProvider
		found    bool
	}{
		{"https://accounts.google.com", ProviderGoogle, true},
		{"https://appleid.apple.com", ProviderApple, true},
		{"https://www.facebook.com", ProviderFacebook, true},
		{"https://discord.com", ProviderDiscord, true},
		{"https://unknown.com", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.issuer, func(t *testing.T) {
			provider, found := ProviderFromIssuer(tt.issuer)
			assert.Equal(t, tt.expected, provider)
			assert.Equal(t, tt.found, found)
		})
	}
}

func TestGenerateRandomPepper(t *testing.T) {
	pepper1, err := GenerateRandomPepper()
	require.NoError(t, err)
	assert.Len(t, pepper1, 31)

	pepper2, err := GenerateRandomPepper()
	require.NoError(t, err)

	// Random peppers should be different
	assert.NotEqual(t, pepper1, pepper2)
}

func TestZKProof_JSON(t *testing.T) {
	proof := &ZKProof{
		A:       []byte{1, 2, 3},
		B:       []byte{4, 5, 6},
		C:       []byte{7, 8, 9},
		Variant: "groth16",
	}

	// Should marshal/unmarshal correctly
	data, err := json.Marshal(proof)
	require.NoError(t, err)

	var decoded ZKProof
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, proof.Variant, decoded.Variant)
}
