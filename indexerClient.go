package aptos

import (
	"context"
	"encoding/json"
	"github.com/hasura/go-graphql-client"
	"os"
)

// -- Note: all query parameters must start with capital letters --

type IndexerClient struct {
	inner *graphql.Client
}

func NewIndexerClient(nodeClient *NodeClient, url string) *IndexerClient {
	// Reuse the HTTP client in the node client
	client := graphql.NewClient(url, nodeClient.client)
	return &IndexerClient{
		client,
	}
}

func (ic *IndexerClient) Query(query any, variables map[string]any, options ...graphql.Option) error {
	return ic.inner.Query(context.Background(), query, variables, options...)
}

type CoinBalance struct {
	CoinType string
	Amount   uint64
}

func (ic *IndexerClient) GetCoinBalances(address AccountAddress) ([]CoinBalance, error) {
	var out []CoinBalance
	var q struct {
		Current_coin_balances []struct {
			CoinType      string `graphql:"coin_type"`
			Amount        uint64
			Owner_address string
		} `graphql:"current_coin_balances(where: {owner_address: {_eq: $address}})"`
	}

	variables := map[string]any{
		"address": address.StringLong(),
	}
	err := ic.Query(&q, variables)

	if err != nil {
		return nil, err
	}

	for _, coin := range q.Current_coin_balances {
		out = append(out, CoinBalance{
			CoinType: coin.CoinType,
			Amount:   coin.Amount,
		})
	}

	return out, nil
}

func (ic *IndexerClient) GetProcessorStatus(processorName string) (uint64, error) {
	var q struct {
		Processor_status []struct {
			Last_success_version uint64
		} `graphql:"processor_status(where: {processor: {_eq: $processor_name}})"`
	}
	variables := map[string]any{
		"processor_name": processorName,
	}
	err := ic.Query(&q, variables)
	if err != nil {
		return 0, err
	}

	return q.Processor_status[0].Last_success_version, err
}

// print Helper function for debugging and printing out query outputs
func print(v any) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		panic(err)
	}
}