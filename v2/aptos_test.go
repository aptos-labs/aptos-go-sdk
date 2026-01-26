package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	// Version should be a non-empty string
	assert.NotEmpty(t, Version)
}

func TestNetworkConfigs(t *testing.T) {
	t.Run("Mainnet", func(t *testing.T) {
		assert.Equal(t, "mainnet", Mainnet.Name)
		assert.Equal(t, uint8(1), Mainnet.ChainID)
		assert.NotEmpty(t, Mainnet.NodeURL)
		assert.NotEmpty(t, Mainnet.IndexerURL)
		assert.Empty(t, Mainnet.FaucetURL) // No faucet on mainnet
	})

	t.Run("Testnet", func(t *testing.T) {
		assert.Equal(t, "testnet", Testnet.Name)
		assert.Equal(t, uint8(2), Testnet.ChainID)
		assert.NotEmpty(t, Testnet.NodeURL)
		assert.NotEmpty(t, Testnet.IndexerURL)
		assert.NotEmpty(t, Testnet.FaucetURL)
	})

	t.Run("Devnet", func(t *testing.T) {
		assert.Equal(t, "devnet", Devnet.Name)
		// ChainID can be 0 (changes on reset)
		assert.NotEmpty(t, Devnet.NodeURL)
		assert.NotEmpty(t, Devnet.IndexerURL)
		assert.NotEmpty(t, Devnet.FaucetURL)
	})

	t.Run("Localnet", func(t *testing.T) {
		assert.Equal(t, "localnet", Localnet.Name)
		assert.Equal(t, uint8(4), Localnet.ChainID)
		assert.Contains(t, Localnet.NodeURL, "localhost")
		assert.Contains(t, Localnet.IndexerURL, "localhost")
		assert.Contains(t, Localnet.FaucetURL, "localhost")
	})
}

func TestWellKnownAddresses(t *testing.T) {
	t.Run("AccountZero", func(t *testing.T) {
		expected := AccountAddress{}
		assert.Equal(t, expected, AccountZero)
		assert.Equal(t, "0x0", AccountZero.String())
	})

	t.Run("AccountOne", func(t *testing.T) {
		assert.Equal(t, "0x1", AccountOne.String())
		assert.Equal(t, byte(0x01), AccountOne[31])
	})

	t.Run("AccountTwo", func(t *testing.T) {
		assert.Equal(t, "0x2", AccountTwo.String())
		assert.Equal(t, byte(0x02), AccountTwo[31])
	})

	t.Run("AccountThree", func(t *testing.T) {
		assert.Equal(t, "0x3", AccountThree.String())
		assert.Equal(t, byte(0x03), AccountThree[31])
	})

	t.Run("AccountFour", func(t *testing.T) {
		assert.Equal(t, "0x4", AccountFour.String())
		assert.Equal(t, byte(0x04), AccountFour[31])
	})
}

func TestCryptoReExports(t *testing.T) {
	// Test that key generation functions work
	t.Run("GenerateEd25519PrivateKey", func(t *testing.T) {
		key, err := GenerateEd25519PrivateKey()
		assert.NoError(t, err)
		assert.NotNil(t, key)
	})

	t.Run("GenerateSecp256k1Key", func(t *testing.T) {
		key, err := GenerateSecp256k1Key()
		assert.NoError(t, err)
		assert.NotNil(t, key)
	})

	t.Run("NewSingleSigner", func(t *testing.T) {
		key, _ := GenerateSecp256k1Key()
		signer := NewSingleSigner(key)
		assert.NotNil(t, signer)
	})

	t.Run("NoAccountAuthenticator", func(t *testing.T) {
		auth := NoAccountAuthenticator()
		assert.NotNil(t, auth)
	})
}

func TestTypeReExports(t *testing.T) {
	t.Run("ParseAddress", func(t *testing.T) {
		addr, err := ParseAddress("0x1")
		assert.NoError(t, err)
		assert.Equal(t, AccountOne, addr)
	})

	t.Run("MustParseAddress", func(t *testing.T) {
		addr := MustParseAddress("0x1")
		assert.Equal(t, AccountOne, addr)
	})

	t.Run("NewTypeTag", func(t *testing.T) {
		tag := NewTypeTag(&BoolTag{})
		assert.NotNil(t, tag)
	})

	t.Run("NewVectorTag", func(t *testing.T) {
		tag := NewVectorTag(&U8Tag{})
		assert.NotNil(t, tag)
	})

	t.Run("NewStringTag", func(t *testing.T) {
		tag := NewStringTag()
		assert.NotNil(t, tag)
	})

	t.Run("NewOptionTag", func(t *testing.T) {
		tag := NewOptionTag(&U64Tag{})
		assert.NotNil(t, tag)
	})

	t.Run("NewObjectTag", func(t *testing.T) {
		tag := NewObjectTag(&AddressTag{})
		assert.NotNil(t, tag)
	})

	t.Run("ParseTypeTag", func(t *testing.T) {
		tag, err := ParseTypeTag("u64")
		assert.NoError(t, err)
		assert.NotNil(t, tag)
	})

	t.Run("AptosCoinTypeTag", func(t *testing.T) {
		tag := AptosCoinTypeTag
		assert.NotNil(t, tag)
		assert.Contains(t, tag.String(), "aptos_coin")
	})
}

func TestAuthenticatorConstants(t *testing.T) {
	// Verify authenticator type constants are defined
	assert.Equal(t, uint8(0), uint8(AccountAuthenticatorEd25519))
	assert.Equal(t, uint8(1), uint8(AccountAuthenticatorMultiEd25519))
	assert.Equal(t, uint8(2), uint8(AccountAuthenticatorSingleSender))
	assert.Equal(t, uint8(3), uint8(AccountAuthenticatorMultiKey))
}

func TestDeriveSchemeConstants(t *testing.T) {
	// Verify derive scheme constants are defined
	assert.Equal(t, uint8(0), Ed25519Scheme)
	assert.Equal(t, uint8(1), MultiEd25519Scheme)
	assert.Equal(t, uint8(2), SingleKeyScheme)
	assert.Equal(t, uint8(3), MultiKeyScheme)
	assert.Equal(t, uint8(252), DeriveObjectScheme)
	assert.Equal(t, uint8(254), NamedObjectScheme)
	assert.Equal(t, uint8(255), ResourceAccountScheme)
}
