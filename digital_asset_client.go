package aptos

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// DigitalAssetClient provides methods for interacting with Aptos Digital Assets (NFTs) and Fungible Assets
type DigitalAssetClient struct {
	client *Client
}

// NewDigitalAssetClient creates a new digital asset client
func NewDigitalAssetClient(client *Client) *DigitalAssetClient {
	return &DigitalAssetClient{
		client: client,
	}
}

// Collection represents a collection of digital assets
type Collection struct {
	Address                  string `json:"address"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	URI                      string `json:"uri"`
	Creator                  string `json:"creator"`
	MaxSupply                uint64 `json:"max_supply"`
	CurrentSupply            uint64 `json:"current_supply"`
	MutableDescription       bool   `json:"mutable_description"`
	MutableRoyalty           bool   `json:"mutable_royalty"`
	MutableURI               bool   `json:"mutable_uri"`
	MutableTokenDesc         bool   `json:"mutable_token_description"`
	MutableTokenName         bool   `json:"mutable_token_name"`
	MutableTokenProps        bool   `json:"mutable_token_properties"`
	MutableTokenURI          bool   `json:"mutable_token_uri"`
	TokensBurnedIdsSupported bool   `json:"tokens_burnable_by_creator"`
	TokensFreezableByCreator bool   `json:"tokens_freezable_by_creator"`
	RoyaltyPayeeAddress      string `json:"royalty_payee_address"`
	RoyaltyPointsNumerator   uint64 `json:"royalty_points_numerator"`
	RoyaltyPointsDenominator uint64 `json:"royalty_points_denominator"`
}

// Token represents a digital asset (NFT)
type Token struct {
	Address                string            `json:"address"`
	Name                   string            `json:"name"`
	Description            string            `json:"description"`
	URI                    string            `json:"uri"`
	CollectionAddress      string            `json:"collection_address"`
	CollectionName         string            `json:"collection_name"`
	CreatorAddress         string            `json:"creator_address"`
	OwnerAddress           string            `json:"owner_address"`
	PropertyMap            map[string]string `json:"property_map"`
	PropertyMapKeys        []string          `json:"property_map_keys"`
	PropertyMapValues      []string          `json:"property_map_values"`
	PropertyMapTypes       []string          `json:"property_map_types"`
	LastTransactionHash    string            `json:"last_transaction_hash"`
	LastTransactionVersion uint64            `json:"last_transaction_version"`
	IsSoulBound            bool              `json:"is_soul_bound"`
	IsBurned               bool              `json:"is_burned"`
	IsFrozen               bool              `json:"is_frozen"`
}

// FungibleAsset represents a fungible asset balance
type FungibleAsset struct {
	AssetAddress     string `json:"asset_address"`
	AssetName        string `json:"asset_name"`
	AssetSymbol      string `json:"asset_symbol"`
	AssetDecimals    uint8  `json:"asset_decimals"`
	AssetURI         string `json:"asset_uri"`
	Balance          uint64 `json:"balance"`
	OwnerAddress     string `json:"owner_address"`
	LastTransaction  string `json:"last_transaction"`
	IsFrozen         bool   `json:"is_frozen"`
	SupplyAggregator string `json:"supply_aggregator"`
}

// CreateCollectionOptions represents options for creating a collection
type CreateCollectionOptions struct {
	Description              string
	Name                     string
	URI                      string
	MaxSupply                uint64
	MutableDescription       bool
	MutableRoyalty           bool
	MutableURI               bool
	MutableTokenDescription  bool
	MutableTokenName         bool
	MutableTokenProperties   bool
	MutableTokenURI          bool
	TokensBurnableByCreator  bool
	TokensFreezableByCreator bool
	RoyaltyPointsNumerator   uint64
	RoyaltyPointsDenominator uint64
}

// MintTokenOptions represents options for minting a token
type MintTokenOptions struct {
	CollectionName   string
	TokenName        string
	TokenDescription string
	TokenURI         string
	PropertyKeys     []string
	PropertyValues   []string
	PropertyTypes    []string
	IsSoulBound      bool
}

// === COLLECTION OPERATIONS ===

// CreateCollection creates a new collection using the Digital Asset standard
func (dac *DigitalAssetClient) CreateCollection(creator *Account, options CreateCollectionOptions) (string, error) {
	// Serialize arguments for BCS
	descriptionBytes, err := bcs.SerializeString(options.Description)
	if err != nil {
		return CreateErrorMessage("description", err)
	}

	nameBytes, err := bcs.SerializeString(options.Name)
	if err != nil {
		return CreateErrorMessage("name", err)
	}

	uriBytes, err := bcs.SerializeString(options.URI)
	if err != nil {
		return CreateErrorMessage("URI", err)
	}

	maxSupplyBytes, err := bcs.SerializeU64(options.MaxSupply)
	if err != nil {
		return CreateErrorMessage("max_supply", err)
	}

	mutableDescBytes, err := bcs.SerializeBool(options.MutableDescription)
	if err != nil {
		return CreateErrorMessage("mutable_description", err)
	}

	mutableRoyaltyBytes, err := bcs.SerializeBool(options.MutableRoyalty)
	if err != nil {
		return CreateErrorMessage("mutable_royalty", err)
	}

	mutableURIBytes, err := bcs.SerializeBool(options.MutableURI)
	if err != nil {
		return CreateErrorMessage("mutable_uri", err)
	}

	mutableTokenDescBytes, err := bcs.SerializeBool(options.MutableTokenDescription)
	if err != nil {
		return CreateErrorMessage("mutable_token_description", err)
	}

	mutableTokenNameBytes, err := bcs.SerializeBool(options.MutableTokenName)
	if err != nil {
		return CreateErrorMessage("mutable_token_name", err)
	}

	mutableTokenPropsBytes, err := bcs.SerializeBool(options.MutableTokenProperties)
	if err != nil {
		return CreateErrorMessage("mutable_token_properties", err)
	}

	mutableTokenURIBytes, err := bcs.SerializeBool(options.MutableTokenURI)
	if err != nil {
		return CreateErrorMessage("mutable_token_uri", err)
	}

	tokensBurnableBytes, err := bcs.SerializeBool(options.TokensBurnableByCreator)
	if err != nil {
		return CreateErrorMessage("tokens_burnable_by_creator", err)
	}

	tokensFreezableBytes, err := bcs.SerializeBool(options.TokensFreezableByCreator)
	if err != nil {
		return CreateErrorMessage("tokens_freezable_by_creator", err)
	}

	royaltyNumeratorBytes, err := bcs.SerializeU64(options.RoyaltyPointsNumerator)
	if err != nil {
		return CreateErrorMessage("royalty_points_numerator", err)
	}

	royaltyDenominatorBytes, err := bcs.SerializeU64(options.RoyaltyPointsDenominator)
	if err != nil {
		return CreateErrorMessage("royalty_points_denominator", err)
	}

	// Build the transaction
	rawTxn, err := dac.client.BuildTransaction(creator.AccountAddress(), TransactionPayload{
		Payload: &EntryFunction{
			Module: ModuleId{
				Address: AccountFour, // 0x4
				Name:    "aptos_token",
			},
			Function: "create_collection",
			ArgTypes: []TypeTag{},
			Args: [][]byte{
				descriptionBytes,
				maxSupplyBytes,
				nameBytes,
				uriBytes,
				mutableDescBytes,
				mutableRoyaltyBytes,
				mutableURIBytes,
				mutableTokenDescBytes,
				mutableTokenNameBytes,
				mutableTokenPropsBytes,
				mutableTokenURIBytes,
				tokensBurnableBytes,
				tokensFreezableBytes,
				royaltyNumeratorBytes,
				royaltyDenominatorBytes,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign and submit the transaction
	signedTxn, err := rawTxn.SignedTransaction(creator)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	submitResult, err := dac.client.SubmitTransaction(signedTxn)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	return submitResult.Hash, nil
}

func CreateErrorMessage(attr string, err error) (string, error) {
	return "", fmt.Errorf("failed to serialize %s: %w", attr, err)
}

// === TOKEN OPERATIONS ===

// MintToken mints a new token (NFT) to the specified collection
func (dac *DigitalAssetClient) MintToken(creator *Account, recipient AccountAddress, options MintTokenOptions) (string, error) {
	// Serialize arguments
	collectionBytes, err := bcs.SerializeString(options.CollectionName)
	if err != nil {
		return CreateErrorMessage("collection", err)
	}

	descriptionBytes, err := bcs.SerializeString(options.TokenDescription)
	if err != nil {
		return CreateErrorMessage("description", err)
	}

	nameBytes, err := bcs.SerializeString(options.TokenName)
	if err != nil {
		return CreateErrorMessage("name", err)
	}

	uriBytes, err := bcs.SerializeString(options.TokenURI)
	if err != nil {
		return CreateErrorMessage("URI", err)
	}

	// Serialize property keys
	propertyKeysBytes, err := SerializeVectorString(options.PropertyKeys)
	if err != nil {
		return CreateErrorMessage("property_keys", err)
	}

	// Serialize property types
	propertyTypesBytes, err := SerializeVectorString(options.PropertyTypes)
	if err != nil {
		return CreateErrorMessage("property_types", err)
	}

	// Serialize property values (as bytes)
	propertyValuesBytes, err := SerializePropertyValues(options.PropertyValues, options.PropertyTypes)
	if err != nil {
		return CreateErrorMessage("property_values", err)
	}

	// Build the transaction
	var functionName string
	var args [][]byte

	if options.IsSoulBound {
		functionName, args, err = mintSoulboundArgs(recipient, collectionBytes, descriptionBytes, nameBytes, uriBytes, propertyKeysBytes, propertyTypesBytes, propertyValuesBytes)
		if err != nil {
			return "", err
		}
	} else {
		functionName, args = mintArgs(collectionBytes, descriptionBytes, nameBytes, uriBytes, propertyKeysBytes, propertyTypesBytes, propertyValuesBytes)
	}

	rawTxn, err := dac.client.BuildTransaction(creator.AccountAddress(), TransactionPayload{
		Payload: &EntryFunction{
			Module: ModuleId{
				Address: AccountFour, // 0x4
				Name:    "aptos_token",
			},
			Function: functionName,
			ArgTypes: []TypeTag{},
			Args:     args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign and submit the transaction
	signedTxn, err := rawTxn.SignedTransaction(creator)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	submitResult, err := dac.client.SubmitTransaction(signedTxn)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	return submitResult.Hash, nil
}

func mintSoulboundArgs(recipient AccountAddress, collectionBytes []byte, descriptionBytes []byte, nameBytes []byte, uriBytes []byte, propertyKeysBytes []byte, propertyTypesBytes []byte, propertyValuesBytes []byte) (string, [][]byte, error) {
	functionName := "mint_soul_bound"

	// Serialize the soul_bound_to address
	recipientBytes, err := bcs.Serialize(&recipient)
	if err != nil {
		return "", nil, fmt.Errorf("failed to serialize recipient: %w", err)
	}

	// mint_soul_bound requires the recipient address as the last argument
	args := [][]byte{
		collectionBytes,
		descriptionBytes,
		nameBytes,
		uriBytes,
		propertyKeysBytes,
		propertyTypesBytes,
		propertyValuesBytes,
		recipientBytes, // soul_bound_to: address
	}

	return functionName, args, err
}

func mintArgs(collectionBytes []byte, descriptionBytes []byte, nameBytes []byte, uriBytes []byte, propertyKeysBytes []byte, propertyTypesBytes []byte, propertyValuesBytes []byte) (string, [][]byte) {
	functionName := "mint"

	// mint_soul_bound requires the recipient address as the last argument
	args := [][]byte{
		collectionBytes,
		descriptionBytes,
		nameBytes,
		uriBytes,
		propertyKeysBytes,
		propertyTypesBytes,
		propertyValuesBytes,
	}

	return functionName, args
}

// TransferToken transfers a token from one account to another
func (dac *DigitalAssetClient) TransferToken(sender *Account, tokenAddress AccountAddress, recipient AccountAddress) (*string, error) {
	// Serialize arguments
	tokenAddressBytes, err := bcs.Serialize(&tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize token address: %w", err)
	}

	recipientBytes, err := bcs.Serialize(&recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recipient: %w", err)
	}

	// Build the transaction
	rawTxn, err := dac.client.BuildTransaction(sender.AccountAddress(), TransactionPayload{
		Payload: &EntryFunction{
			Module: ModuleId{
				Address: AccountOne, // 0x1
				Name:    "object",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{ObjectTypeTag},
			Args: [][]byte{
				tokenAddressBytes,
				recipientBytes,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign and submit the transaction
	signedTxn, err := rawTxn.SignedTransaction(sender)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	submitResult, err := dac.client.SubmitTransaction(signedTxn)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	return &submitResult.Hash, nil
}

// === HELPER FUNCTIONS FOR BCS SERIALIZATION ===

// SerializeVectorString serializes a vector of strings for BCS

func SerializeVectorString(strings []string) ([]byte, error) {
	return bcs.SerializeSingle(func(ser *bcs.Serializer) {
		bcs.SerializeSequenceWithFunction(strings, ser, func(ser *bcs.Serializer, item string) {
			ser.WriteString(item)
		})
	})
}

// SerializeVectorBytes serializes a vector of byte arrays for BCS
func SerializeVectorBytes(values []string) ([]byte, error) {
	return bcs.SerializeSingle(func(ser *bcs.Serializer) {
		byteArrays := make([][]byte, len(values))
		for i, value := range values {
			byteArrays[i] = []byte(value)
		}

		bcs.SerializeSequenceWithFunction(byteArrays, ser, func(ser *bcs.Serializer, item []byte) {
			ser.WriteBytes(item)
		})
	})
}

func SerializePropertyValues(values []string, types []string) ([]byte, error) {
	if len(values) != len(types) {
		return nil, errors.New("property values and types must have the same length")
	}

	// Convert each value to BCS-serialized bytes according to its type
	byteArrays := make([][]byte, len(values))

	for i, value := range values {
		propertyType := types[i]

		var valueBytes []byte
		var err error

		switch propertyType {
		case "0x1::string::String":
			// For Move String type, BCS serialize the string
			valueBytes, err = bcs.SerializeSingle(func(s *bcs.Serializer) {
				s.WriteString(value)
			})

		case "u8":
			// For u8, parse and BCS serialize as u8
			numValue, parseErr := ConvertToU8(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse u8 value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.SerializeU8(numValue)

		case "u16":
			// For u8, parse and BCS serialize as u8
			numValue, parseErr := ConvertToU16(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse u8 value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.SerializeU16(numValue)

		case "u32":
			// For u8, parse and BCS serialize as u8
			numValue, parseErr := ConvertToU32(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse u8 value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.SerializeU32(numValue)

		case "u64":
			// For u64, parse and BCS serialize as u64
			numValue, parseErr := ConvertToU64(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse u64 value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.SerializeU64(numValue)

		case "u128":
			// For u128, parse as bigInt and BCS serialize as u128
			bigIntValue, parseErr := StrToBigInt(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse bigInt value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.SerializeU128(*bigIntValue)

		case "bool":
			// For bool, parse and BCS serialize as bool
			boolValue, parseErr := strconv.ParseBool(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse bool value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.SerializeBool(boolValue)

		case "address":
			// For address, decode hex and use as 32-byte array
			addressBytes, parseErr := ConvertToAddress(value)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse address value '%s': %w", value, parseErr)
			}
			valueBytes, err = bcs.Serialize(addressBytes)

		case "vector<u8>":
			// For byte vector, decode from hex or use as UTF-8, then BCS serialize
			var bytesValue []byte
			if strings.HasPrefix(value, "0x") {
				bytesValue, err = hex.DecodeString(value[2:])
				if err != nil {
					return nil, fmt.Errorf("failed to parse hex bytes value '%s': %w", value, err)
				}
			} else {
				bytesValue = []byte(value)
			}
			// BCS serialize the byte vector
			valueBytes, err = bcs.SerializeSingle(func(s *bcs.Serializer) {
				s.WriteBytes(bytesValue)
			})

		default:
			// For unknown types, fall back to BCS string
			return nil, fmt.Errorf("failed to serialize, unknown property type '%s'", propertyType)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to serialize property value '%s' of type '%s': %w", value, propertyType, err)
		}

		byteArrays[i] = valueBytes
	}

	// Serialize as vector<vector<u8>> using BCS
	return bcs.SerializeSingle(func(ser *bcs.Serializer) {
		bcs.SerializeSequenceWithFunction(byteArrays, ser, func(ser *bcs.Serializer, item []byte) {
			ser.WriteBytes(item)
		})
	})
}
