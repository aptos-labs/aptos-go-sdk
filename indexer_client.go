package aptos

import "github.com/hasura/go-graphql-client"

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
