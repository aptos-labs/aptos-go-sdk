package aptos

// Version is the current version of the Aptos Go SDK v2.
const Version = "2.0.0-alpha.1"

// Re-export commonly used types from internal packages for convenience.
// Users can import just "github.com/aptos-labs/aptos-go-sdk/v2" for most use cases.

// Network presets for common Aptos networks.
var (
	// Mainnet is the production Aptos network.
	Mainnet = NetworkConfig{
		Name:       "mainnet",
		ChainID:    1,
		NodeURL:    "https://fullnode.mainnet.aptoslabs.com/v1",
		IndexerURL: "https://api.mainnet.aptoslabs.com/v1/graphql",
		FaucetURL:  "", // No faucet on mainnet
	}

	// Testnet is the Aptos test network.
	Testnet = NetworkConfig{
		Name:       "testnet",
		ChainID:    2,
		NodeURL:    "https://fullnode.testnet.aptoslabs.com/v1",
		IndexerURL: "https://api.testnet.aptoslabs.com/v1/graphql",
		FaucetURL:  "https://faucet.testnet.aptoslabs.com",
	}

	// Devnet is the Aptos development network (resets frequently).
	Devnet = NetworkConfig{
		Name:       "devnet",
		ChainID:    0, // Changes on reset
		NodeURL:    "https://fullnode.devnet.aptoslabs.com/v1",
		IndexerURL: "https://api.devnet.aptoslabs.com/v1/graphql",
		FaucetURL:  "https://faucet.devnet.aptoslabs.com",
	}

	// Localnet is a local development network.
	Localnet = NetworkConfig{
		Name:       "localnet",
		ChainID:    4,
		NodeURL:    "http://localhost:8080/v1",
		IndexerURL: "http://localhost:8090/v1/graphql",
		FaucetURL:  "http://localhost:8081",
	}
)

// NetworkConfig contains the configuration for connecting to an Aptos network.
type NetworkConfig struct {
	Name       string // Human-readable name for the network
	ChainID    uint8  // Chain ID for transaction signing
	NodeURL    string // Full node API URL
	IndexerURL string // Indexer GraphQL URL
	FaucetURL  string // Faucet URL (empty for mainnet)
}

// Well-known account addresses.
var (
	// AccountZero is the 0x0 address.
	AccountZero = AccountAddress{}

	// AccountOne is the 0x1 address (core framework).
	AccountOne = AccountAddress{31: 0x01}

	// AccountTwo is the 0x2 address.
	AccountTwo = AccountAddress{31: 0x02}

	// AccountThree is the 0x3 address (token framework).
	AccountThree = AccountAddress{31: 0x03}

	// AccountFour is the 0x4 address (object framework).
	AccountFour = AccountAddress{31: 0x04}
)
