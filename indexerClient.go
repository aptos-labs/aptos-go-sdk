package aptos

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
)

// -- Note: all query parameters must start with capital letters --

type IndexerClient ExposedIndexerClient

// NewIndexerClient creates a new client specifically for requesting data from the indexer
func NewIndexerClient(httpClient *http.Client, url string) *IndexerClient {
	client := NewIndexerClient(httpClient, url)
	return (*IndexerClient)(client)
}

func (ic *IndexerClient) Query(query any, variables map[string]any, options ...graphql.Option) error {
	return (*ExposedIndexerClient)(ic).inner.Query(context.Background(), query, variables, options...)
}

func (ic *IndexerClient) GetCoinBalances(address AccountAddress) ([]CoinBalance, error) {
	return (*ExposedIndexerClient)(ic).GetCoinBalances(context.Background(), address)
}

func (ic *IndexerClient) GetProcessorStatus(ctx context.Context, processorName string) (uint64, error) {
	return (*ExposedIndexerClient)(ic).GetProcessorStatus(context.Background(), processorName)
}

func (ic *IndexerClient) WaitOnIndexer(ctx context.Context, processorName string, requestedVersion uint64) error {
	return (*ExposedIndexerClient)(ic).WaitOnIndexer(context.Background(), processorName, requestedVersion)
}

// ExposedIndexerClient is a GraphQL client specifically for requesting for data from the Aptos indexer
type ExposedIndexerClient struct {
	inner *graphql.Client
}

// NewIndexerClient creates a new client specifically for requesting data from the indexer
func NewExposedIndexerClient(httpClient *http.Client, url string) *ExposedIndexerClient {
	// Reuse the HTTP client in the node client
	client := graphql.NewClient(url, httpClient)
	return &ExposedIndexerClient{
		client,
	}
}

// Query is a generic function for making any GraphQL query against the indexer
func (ic *ExposedIndexerClient) Query(ctx context.Context, query any, variables map[string]any, options ...graphql.Option) error {
	return ic.inner.Query(ctx, query, variables, options...)
}

type CoinBalance struct {
	CoinType string
	Amount   uint64
}

// GetCoinBalances retrieve the coin balances for all coins owned by the address
func (ic *ExposedIndexerClient) GetCoinBalances(ctx context.Context, address AccountAddress) ([]CoinBalance, error) {
	var out []CoinBalance
	var q struct {
		CurrentCoinBalances []struct {
			CoinType     string `graphql:"coin_type"`
			Amount       uint64
			OwnerAddress string `graphql:"owner_address"`
		} `graphql:"current_coin_balances(where: {owner_address: {_eq: $address}})"`
	}

	variables := map[string]any{
		"address": address.StringLong(),
	}
	err := ic.Query(ctx, &q, variables)

	if err != nil {
		return nil, err
	}

	for _, coin := range q.CurrentCoinBalances {
		out = append(out, CoinBalance{
			CoinType: coin.CoinType,
			Amount:   coin.Amount,
		})
	}

	return out, nil
}

// GetProcessorStatus tells the most updated version of the transaction processor.  This helps to determine freshness of data.
func (ic *ExposedIndexerClient) GetProcessorStatus(ctx context.Context, processorName string) (uint64, error) {
	var q struct {
		ProcessorStatus []struct {
			LastSuccessVersion uint64 `graphql:"last_success_version"`
		} `graphql:"processor_status(where: {processor: {_eq: $processor_name}})"`
	}
	variables := map[string]any{
		"processor_name": processorName,
	}
	err := ic.Query(ctx, &q, variables)
	if err != nil {
		return 0, err
	}

	return q.ProcessorStatus[0].LastSuccessVersion, err
}

// WaitOnIndexer waits for the indexer processorName specified to catch up to the requestedVersion
func (ic *ExposedIndexerClient) WaitOnIndexer(ctx context.Context, processorName string, requestedVersion uint64) error {
	// TODO: add customizable timeout and sleep time
	const sleepTime = 100 * time.Millisecond
	const timeout = 5 * time.Second
	startTime := time.Now()
	for {
		version, err := ic.GetProcessorStatus(ctx, processorName)
		if err != nil {
			// TODO: This should probably just retry, depending on the error
			return err
		}

		// If we've caught up, skip out
		if version >= requestedVersion {
			break
		} else if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout waiting on requested version.  last version seen: %d requested: %d", version, requestedVersion)
		}

		// Sleep and try again later
		time.Sleep(sleepTime)
	}
	return nil
}
