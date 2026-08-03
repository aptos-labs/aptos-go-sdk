//go:build cgo

// Package native provides FFI-backed confidential asset operations (balance decrypt,
// range proofs). Importing this package requires CGO_ENABLED=1 and libaptos_confidential_asset_ffi.
package native

import (
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
)

// Client wraps confidentialasset.Client with Rust FFI-backed methods.
type Client struct {
	*confidentialasset.Client
}

// Wrap returns a native client for decrypt and proof-bearing transactions.
func Wrap(c *confidentialasset.Client) *Client {
	if c == nil {
		return nil
	}
	return &Client{Client: c}
}
