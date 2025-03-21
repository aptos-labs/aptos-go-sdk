package crypto

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAuthKey = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

func TestAuthenticationKey_CryptoMaterial(t *testing.T) {
	authKeyBytes, err := util.ParseHex(testAuthKey)
	require.NoError(t, err)

	authKeyFromString := &AuthenticationKey{}
	err = authKeyFromString.FromHex(testAuthKey)
	require.NoError(t, err)

	authKeyFromBytes := &AuthenticationKey{}
	err = authKeyFromBytes.FromBytes(authKeyBytes)
	require.NoError(t, err)

	assert.Equal(t, authKeyFromString, authKeyFromBytes)

	derivedBytes := authKeyFromString.Bytes()
	assert.Equal(t, authKeyBytes, derivedBytes)
	hexString := authKeyFromString.ToHex()
	assert.Equal(t, testAuthKey, hexString)
}

func TestAuthenticationKey_CryptoMaterialError(t *testing.T) {
	authKey := &AuthenticationKey{}
	err := authKey.FromHex("0x123456")
	require.Error(t, err) // Not long enough

	err = authKey.FromHex("abcde")
	require.Error(t, err) // Not a string
}
