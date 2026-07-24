package confidentialasset

import (
	"context"
)

// GetActivities queries indexer GraphQL (not implemented; see TS getConfidentialAssetActivities).
func (c *Client) GetActivities(_ context.Context, _ map[string]any) ([]any, error) {
	return nil, ErrIndexerNotImplemented
}
