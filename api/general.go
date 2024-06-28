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
