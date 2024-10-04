package aptos

import (
	"context"
	"fmt"
	"github.com/hasura/go-graphql-client"
	"net/http"
	"time"
)

// -- Note: all query parameters must start with capital letters --

// IndexerClient is a GraphQL client specifically for requesting for data from the Aptos indexer
type IndexerClient struct {
	inner *graphql.Client
}

// NewIndexerClient creates a new client specifically for requesting data from the indexer
func NewIndexerClient(httpClient *http.Client, url string) *IndexerClient {
	// Reuse the HTTP client in the node client
	client := graphql.NewClient(url, httpClient)
	return &IndexerClient{
		client,
	}
}

// Query is a generic function for making any GraphQL query against the indexer
func (ic *IndexerClient) Query(query any, variables map[string]any, options ...graphql.Option) error {
	return ic.inner.Query(context.Background(), query, variables, options...)
}

type CoinBalance struct {
	CoinType string
	Amount   uint64
}

// GetCoinBalances retrieve the coin balances for all coins owned by the address
func (ic *IndexerClient) GetCoinBalances(address AccountAddress) ([]CoinBalance, error) {
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
	err := ic.Query(&q, variables)

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
func (ic *IndexerClient) GetProcessorStatus(processorName string) (uint64, error) {
	var q struct {
		ProcessorStatus []struct {
			LastSuccessVersion uint64 `graphql:"last_success_version"`
		} `graphql:"processor_status(where: {processor: {_eq: $processor_name}})"`
	}
	variables := map[string]any{
		"processor_name": processorName,
	}
	err := ic.Query(&q, variables)
	if err != nil {
		return 0, err
	}

	return q.ProcessorStatus[0].LastSuccessVersion, err
}

// WaitOnIndexer waits for the indexer processorName specified to catch up to the requestedVersion
func (ic *IndexerClient) WaitOnIndexer(processorName string, requestedVersion uint64) error {
	// TODO: add customizable timeout and sleep time
	const sleepTime = 100 * time.Millisecond
	const timeout = 5 * time.Second
	startTime := time.Now()
	for {
		version, err := ic.GetProcessorStatus(processorName)
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
