// Package api represents all types associated with the Aptos REST API.  It handles JSON packing and un-packing, through
// multiple inner types.
//
// Quick links:
//
//   - [Aptos API Reference] for an interactive OpenAPI documentation experience.
//
// [Aptos API Reference]: https://aptos.dev/en/build/apis/fullnode-rest-api-reference
package api

// HealthCheckResponse is the response to a health check request
//
// Example:
//
//	{
//		"message": "aptos-node:ok"
//	}
type HealthCheckResponse struct {
	Message string `json:"message"` // Message is the human-readable message, usually "aptos-node:ok"
}
