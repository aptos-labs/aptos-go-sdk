package aptos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test_Spec_ClientConfig tests the client configuration
//
//   - It must be able to create a devnet client
//   - It must be able to create a testnet client
//   - It must be able to create a mainnet client
//   - It must be able to create a localnet client
//   - It must be able to create a client with a custom configuration
//   - It must be able to create a client with a custom configuration and custom headers
func Test_Spec_ClientConfig(t *testing.T) {
	// It must be able to create a devnet client
	_, err := NewClient(DevnetConfig)
	assert.NoError(t, err)

	// It must be able to create a testnet client
	_, err = NewClient(TestnetConfig)
	assert.NoError(t, err)

	// It must be able to create a mainnet client
	_, err = NewClient(MainnetConfig)
	assert.NoError(t, err)

	// It must be able to create a localnet client
	_, err = NewClient(LocalnetConfig)
	assert.NoError(t, err)

	// It must be able to create a client with a custom configuration
	_, err = NewClient(NetworkConfig{
		Name:       "Localnet",
		NodeUrl:    "http://localhost:8080/v1",
		IndexerUrl: "http://localhost:8090/v1/graphql",
		FaucetUrl:  "http://localhost:8081",
		ChainId:    4,
	})
	assert.NoError(t, err)

	// It must be able to create a client with a custom configuration and custom headers
	// TODO: Do we put this in the setup directly rather than adding later?
	client, err := NewClient(NetworkConfig{
		Name:       "Localnet",
		NodeUrl:    "http://localhost:8080/v1",
		IndexerUrl: "http://localhost:8090/v1/graphql",
		FaucetUrl:  "http://localhost:8081",
		ChainId:    4,
	})
	assert.NoError(t, err)
	client.SetHeader("Authorization", "Bearer abcdefg")
}
