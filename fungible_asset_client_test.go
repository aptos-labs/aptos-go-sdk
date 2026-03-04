package aptos

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFAMockClient creates a mock client that handles View calls for fungible asset testing.
func newFAMockClient(t *testing.T, viewResponse []any) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
		case r.URL.Path == "/view" && r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(viewResponse)
		default:
			// For AccountResource calls (metadata check), return a valid resource
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type": "0x1::fungible_asset::Metadata",
				"data": map[string]any{},
			})
		}
	}))

	client, err := NewClient(NetworkConfig{
		Name:    "mocknet",
		NodeUrl: server.URL,
		ChainId: 4,
	})
	require.NoError(t, err)
	return client, server
}

func TestFungibleAssetClient_Name_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"Aptos Coin"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	name, err := faClient.Name()
	require.NoError(t, err)
	assert.Equal(t, "Aptos Coin", name)
}

func TestFungibleAssetClient_Symbol_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"APT"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	symbol, err := faClient.Symbol()
	require.NoError(t, err)
	assert.Equal(t, "APT", symbol)
}

func TestFungibleAssetClient_Decimals_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{float64(8)})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	decimals, err := faClient.Decimals()
	require.NoError(t, err)
	assert.Equal(t, uint8(8), decimals)
}

func TestFungibleAssetClient_PrimaryBalance_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"500000000"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	owner := AccountTwo
	balance, err := faClient.PrimaryBalance(&owner)
	require.NoError(t, err)
	assert.Equal(t, uint64(500000000), balance)
}

func TestFungibleAssetClient_PrimaryStoreExists_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{true})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	owner := AccountTwo
	exists, err := faClient.PrimaryStoreExists(&owner)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFungibleAssetClient_PrimaryIsFrozen_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{false})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	owner := AccountTwo
	frozen, err := faClient.PrimaryIsFrozen(&owner)
	require.NoError(t, err)
	assert.False(t, frozen)
}

func TestFungibleAssetClient_PrimaryStoreAddress_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"0x0000000000000000000000000000000000000000000000000000000000000002"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	owner := AccountTwo
	addr, err := faClient.PrimaryStoreAddress(&owner)
	require.NoError(t, err)
	assert.Equal(t, AccountTwo, *addr)
}

func TestFungibleAssetClient_Balance_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"1000"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	store := AccountTwo
	balance, err := faClient.Balance(&store)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), balance)
}

func TestFungibleAssetClient_IsFrozen_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{false})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	store := AccountTwo
	frozen, err := faClient.IsFrozen(&store)
	require.NoError(t, err)
	assert.False(t, frozen)
}

func TestFungibleAssetClient_StoreExists_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{true})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	store := AccountTwo
	exists, err := faClient.StoreExists(&store)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFungibleAssetClient_StoreMetadata_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{map[string]any{"inner": "0x1"}})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	store := AccountTwo
	addr, err := faClient.StoreMetadata(&store)
	require.NoError(t, err)
	assert.Equal(t, AccountOne, *addr)
}

func TestFungibleAssetClient_Supply_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{map[string]any{"vec": []any{"1000000"}}})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	supply, err := faClient.Supply()
	require.NoError(t, err)
	assert.Equal(t, int64(1000000), supply.Int64())
}

func TestFungibleAssetClient_Maximum_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{map[string]any{"vec": []any{}}})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	maximum, err := faClient.Maximum()
	require.NoError(t, err)
	assert.Equal(t, int64(-1), maximum.Int64())
}

func TestFungibleAssetClient_MaxSupply_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{map[string]any{"vec": []any{"5000000"}}})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	maximum, err := faClient.Maximum()
	require.NoError(t, err)
	assert.Equal(t, int64(5000000), maximum.Int64())
}

func TestFungibleAssetClient_IconUri_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"https://example.com/icon.png"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	uri, err := faClient.IconUri()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/icon.png", uri)
}

func TestFungibleAssetClient_ProjectUri_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{"https://example.com"})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	uri, err := faClient.ProjectUri()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", uri)
}

func TestFungibleAssetClient_IsUntransferable_Mock(t *testing.T) {
	t.Parallel()
	client, server := newFAMockClient(t, []any{false})
	defer server.Close()

	metadata := AccountTen
	faClient, err := NewFungibleAssetClient(client, &metadata)
	require.NoError(t, err)

	store := AccountTwo
	result, err := faClient.IsUntransferable(&store)
	require.NoError(t, err)
	assert.False(t, result)
}

func TestClient(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("integration test expects network connection to mainnet")
	}

	// Create a new Aptos client
	aptosClient, err := NewClient(LocalnetConfig)
	require.NoError(t, err)

	// Fund sender
	sender, err := NewEd25519Account()
	require.NoError(t, err)

	err = aptosClient.Fund(sender.AccountAddress(), 100000000)
	require.NoError(t, err)

	// Convert to FA APT
	module, err := aptosClient.AccountModule(AccountOne, "coin")
	require.NoError(t, err)
	convertPayload, err := EntryFunctionFromAbi(module.Abi, AccountOne, "coin", "migrate_to_fungible_store", []any{AptosCoinTypeTag}, []any{})
	require.NoError(t, err)

	// TODO: verify that it worked
	_, err = aptosClient.BuildSignAndSubmitTransaction(sender, TransactionPayload{Payload: convertPayload})
	require.NoError(t, err)

	// Owner address
	receiver, err := NewEd25519Account()
	require.NoError(t, err)

	ownerAddress := &receiver.Address
	require.NoError(t, err)

	metadataAddress := &types.AccountTen

	// Create a fungible asset client
	faClient, err := NewFungibleAssetClient(aptosClient, metadataAddress)
	require.NoError(t, err)

	// Retrieve primary store address
	primaryStoreAddress, err := faClient.PrimaryStoreAddress(ownerAddress)
	require.NoError(t, err)

	// Check store exists (it won't)
	storeExists, err := faClient.StoreExists(primaryStoreAddress)
	require.NoError(t, err)
	assert.False(t, storeExists)

	// Send to that address

	transferTxn, err := faClient.TransferPrimaryStore(sender, receiver.Address, 100, SequenceNumber(1))
	require.NoError(t, err)
	submittedTxn, err := aptosClient.SubmitTransaction(transferTxn)
	require.NoError(t, err)
	transaction, err := aptosClient.WaitForTransaction(submittedTxn.Hash)
	require.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.True(t, transaction.Success)

	// Primary store by direct access
	balance, err := faClient.Balance(primaryStoreAddress)
	require.NoError(t, err)
	println("BALANCE: ", balance)

	name, err := faClient.Name()
	require.NoError(t, err)
	println("NAME: ", name)
	symbol, err := faClient.Symbol()
	require.NoError(t, err)
	println("Symbol: ", symbol)

	supply, err := faClient.Supply()
	require.NoError(t, err)
	assert.NotNil(t, supply)

	// TODO: need a custom coin contract to check more
	maximum, err := faClient.Maximum()
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(-1), maximum)

	// This should hold
	storeNotExist, err := faClient.StoreExists(&types.AccountOne)
	require.NoError(t, err)
	assert.False(t, storeNotExist)

	storeMetadataAddress, err := faClient.StoreMetadata(primaryStoreAddress)
	require.NoError(t, err)
	assert.Equal(t, metadataAddress, storeMetadataAddress)

	decimals, err := faClient.Decimals()
	require.NoError(t, err)
	assert.Equal(t, uint8(8), decimals)

	icon, err := faClient.IconUri()
	require.NoError(t, err)
	assert.Empty(t, icon)

	project, err := faClient.ProjectUri()
	require.NoError(t, err)
	assert.Empty(t, project)

	storePrimaryStoreAddress, err := faClient.PrimaryStoreAddress(ownerAddress)
	require.NoError(t, err)
	assert.Equal(t, primaryStoreAddress, storePrimaryStoreAddress)

	primaryStoreExists, err := faClient.PrimaryStoreExists(ownerAddress)
	require.NoError(t, err)
	assert.True(t, primaryStoreExists)

	primaryFrozen, err := faClient.PrimaryIsFrozen(ownerAddress)
	require.NoError(t, err)
	assert.False(t, primaryFrozen)

	primaryUntransferrable, err := faClient.IsUntransferable(primaryStoreAddress)
	require.NoError(t, err)
	assert.False(t, primaryUntransferrable)

	// Primary store by default
	primaryBalance, err := faClient.PrimaryBalance(ownerAddress)
	require.NoError(t, err)
	println("PRIMARY BALANCE: ", primaryBalance)

	isFrozen, err := faClient.IsFrozen(primaryStoreAddress)
	require.NoError(t, err)
	assert.False(t, isFrozen)
}
