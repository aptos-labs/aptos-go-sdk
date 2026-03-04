package aptos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexerClient_GetCoinBalances(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{
			"data": map[string]any{
				"current_fungible_asset_balances": []map[string]any{
					{"asset_type": "0x1::aptos_coin::AptosCoin", "amount": 100, "owner_address": "0x1"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	ic := NewIndexerClient(server.Client(), server.URL)
	balances, err := ic.GetCoinBalances(AccountOne)
	require.NoError(t, err)
	assert.Len(t, balances, 1)
	assert.Equal(t, "0x1::aptos_coin::AptosCoin", balances[0].CoinType)
	assert.Equal(t, uint64(100), balances[0].Amount)
}

func TestIndexerClient_GetProcessorStatus(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{
			"data": map[string]any{
				"processor_status": []map[string]any{
					{"last_success_version": 12345},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	ic := NewIndexerClient(server.Client(), server.URL)
	version, err := ic.GetProcessorStatus("default_processor")
	require.NoError(t, err)
	assert.Equal(t, uint64(12345), version)
}

func TestIndexerClient_QueryIndexer(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{
			"data": map[string]any{
				"current_fungible_asset_balances": []map[string]any{
					{"asset_type": "0x1::aptos_coin::AptosCoin", "amount": 500},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	ic := NewIndexerClient(server.Client(), server.URL)

	var q struct {
		CurrentFungibleAssetBalances []struct {
			AssetType string `graphql:"asset_type"`
			Amount    uint64
		} `graphql:"current_fungible_asset_balances(where: {owner_address: {_eq: $address}})"`
	}

	variables := map[string]any{
		"address": AccountOne.StringLong(),
	}
	err := ic.QueryIndexer(&q, variables)
	require.NoError(t, err)
	require.Len(t, q.CurrentFungibleAssetBalances, 1)
	assert.Equal(t, "0x1::aptos_coin::AptosCoin", q.CurrentFungibleAssetBalances[0].AssetType)
	assert.Equal(t, uint64(500), q.CurrentFungibleAssetBalances[0].Amount)
}
