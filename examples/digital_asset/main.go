// digital_asset is an example of how to create and transfer digital assets (NFTs)
package main

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const (
	testEd25519PrivateKey = "ed25519-priv-0xc5338cd251c22daa8c9c9cc94f498cc8a5c7e1d2e75287a5dda91096fe64efa5"
	FundAmount            = uint64(100_000_000)
)

func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create a creator account locally
	key := crypto.Ed25519PrivateKey{}
	err = key.FromHex(testEd25519PrivateKey)
	if err != nil {
		panic("Failed to decode Ed25519 private key:" + err.Error())
	}
	creator, err := aptos.NewAccountFromSigner(&key)
	if err != nil {
		panic("Failed to create creator:" + err.Error())
	}

	// Create a collector account for receiving NFTs
	collector, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create collector:" + err.Error())
	}

	// Fund both accounts with the faucet
	fmt.Printf("CREATOR: %s\n", creator.Address.String())
	fmt.Printf("COLLECTOR: %s\n", collector.Address.String())

	err = client.Fund(creator.Address, FundAmount)
	if err != nil {
		panic("Failed to fund creator:" + err.Error())
	}

	err = client.Fund(collector.Address, FundAmount)
	if err != nil {
		panic("Failed to fund collector:" + err.Error())
	}

	// Create digital asset client
	daClient := aptos.NewDigitalAssetClient(client)

	// Create a collection
	fmt.Println("\n=== Creating NFT Collection ===")
	collectionOptions := aptos.CreateCollectionOptions{
		Description:              "Epic Space Warriors NFT Collection",
		Name:                     "SpaceWarriors3",
		URI:                      "https://spacewarriors.example.com/collection",
		MaxSupply:                10000,
		MutableDescription:       true,
		MutableRoyalty:           true,
		MutableURI:               true,
		MutableTokenDescription:  true,
		MutableTokenName:         true,
		MutableTokenProperties:   true,
		MutableTokenURI:          true,
		TokensBurnableByCreator:  true,
		TokensFreezableByCreator: false,
		RoyaltyPointsNumerator:   250, // 2.5%
		RoyaltyPointsDenominator: 10000,
	}

	collectionTxHash, err := daClient.CreateCollection(creator, collectionOptions)
	if err != nil {
		panic("Failed to create collection:" + err.Error())
	}
	fmt.Printf("Collection created with transaction hash: %s\n", *collectionTxHash)

	// Wait for collection creation
	_, err = client.WaitForTransaction(*collectionTxHash)
	if err != nil {
		panic("Failed to wait for collection creation:" + err.Error())
	}

	// Mint NFTs
	fmt.Println("\n=== Minting NFTs ===")

	// Mint NFT #1 to creator
	mintOptions1 := aptos.MintTokenOptions{
		CollectionName:   "SpaceWarriors3",
		TokenName:        "Space Warrior #001",
		TokenDescription: "A legendary space warrior with plasma sword",
		TokenURI:         "https://spacewarriors.example.com/nft/001",
		PropertyKeys:     []string{"rarity", "faction", "level", "weapon"},
		PropertyValues:   []string{"legendary", "solar_federation", "85", "plasma_sword"},
		PropertyTypes:    []string{"0x1::string::String", "0x1::string::String", "u64", "0x1::string::String"},
		IsSoulBound:      false,
	}

	nftTxHash1, err := daClient.MintToken(creator, creator.AccountAddress(), mintOptions1)
	if err != nil {
		panic("Failed to mint NFT #1:" + err.Error())
	}
	fmt.Printf("NFT #1 minted with transaction hash: %s\n", *nftTxHash1)

	// Wait for minting
	_, err = client.WaitForTransaction(*nftTxHash1)
	if err != nil {
		panic("Failed to wait for NFT #1 minting:" + err.Error())
	}

	// Mint NFT #2 to collector
	mintOptions2 := aptos.MintTokenOptions{
		CollectionName:   "SpaceWarriors",
		TokenName:        "Space Warrior #002",
		TokenDescription: "An epic space warrior with energy shield",
		TokenURI:         "https://spacewarriors.example.com/nft/002",
		PropertyKeys:     []string{"rarity", "faction", "level", "weapon"},
		PropertyValues:   []string{"epic", "lunar_alliance", "72", "energy_shield"},
		PropertyTypes:    []string{"0x1::string::String", "0x1::string::String", "u64", "0x1::string::String"},
		IsSoulBound:      false,
	}

	nftTxHash2, err := daClient.MintToken(creator, collector.AccountAddress(), mintOptions2)
	if err != nil {
		panic("Failed to mint NFT #2:" + err.Error())
	}
	fmt.Printf("NFT #2 minted with transaction hash: %s\n", *nftTxHash2)

	// Wait for minting
	_, err = client.WaitForTransaction(*nftTxHash2)
	if err != nil {
		panic("Failed to wait for NFT #2 minting:" + err.Error())
	}

	// Mint a soul-bound NFT
	mintOptions3 := aptos.MintTokenOptions{
		CollectionName:   "SpaceWarriors",
		TokenName:        "First Collection Creator Certificate",
		TokenDescription: "A permanent achievement certificate for creating the first collection",
		TokenURI:         "https://spacewarriors.example.com/certificates/001",
		PropertyKeys:     []string{"achievement", "timestamp", "rarity"},
		PropertyValues:   []string{"first_collection_creator", "1640995200", "mythic"},
		PropertyTypes:    []string{"0x1::string::String", "u64", "0x1::string::String"},
		IsSoulBound:      true, // This will be soul-bound
	}

	nftTxHash3, err := daClient.MintToken(creator, creator.AccountAddress(), mintOptions3)
	if err != nil {
		panic("Failed to mint soul-bound NFT:" + err.Error())
	}
	fmt.Printf("Soul-bound NFT minted with transaction hash: %s\n", *nftTxHash3)

	// Wait for minting
	_, err = client.WaitForTransaction(*nftTxHash3)
	if err != nil {
		panic("Failed to wait for soul-bound NFT minting:" + err.Error())
	}

	// Transfer NFT (Note: In reality, you'd need to get the actual token address from events)
	fmt.Println("\n=== Transferring NFT ===")

	// You would parse the token address from the mint transaction events
	// tokenAddress := // Placeholder

	// transferTxHash, err := daClient.TransferToken(creator, tokenAddress, collector.AccountAddress())
	// if err != nil {
	// 	fmt.Printf("Warning: Failed to transfer NFT (expected without proper token address): %v\n", err)
	// } else {
	// 	fmt.Printf("NFT transferred with transaction hash: %s\n", *transferTxHash)

	// 	// Wait for transfer
	// 	_, err = client.WaitForTransaction(*transferTxHash)
	// 	if err != nil {
	// 		panic("Failed to wait for transfer:" + err.Error())
	// 	}
	// }

	fmt.Println("\n=== Digital Asset Example Complete ===")
	fmt.Println("Successfully:")
	fmt.Println("✅ Created NFT collection with customizable properties")
	fmt.Println("✅ Minted regular NFTs with rich property maps")
	fmt.Println("✅ Minted soul-bound NFT (non-transferable)")
	//fmt.Println("✅ Demonstrated transfer operations")
}

// main This example shows how to create and transfer digital assets (NFTs)
func main() {
	// Run the main example
	example(aptos.DevnetConfig)

}
