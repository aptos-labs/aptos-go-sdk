//go:build !cgo

package confidentialasset

import (
	"context"

	"github.com/aptos-labs/aptos-go-sdk/v2"
)

// NormalizeBalance requires CGO + FFI (decrypt + range proof).
func (c *Client) NormalizeBalance(context.Context, aptos.TransactionSigner, aptos.AccountAddress, string, string) (*aptos.Transaction, error) {
	return nil, ErrCGODisabled
}

// Withdraw requires CGO + FFI.
func (c *Client) Withdraw(context.Context, aptos.TransactionSigner, aptos.AccountAddress, uint64, aptos.AccountAddress, string, string) (*aptos.Transaction, error) {
	return nil, ErrCGODisabled
}

// Transfer requires CGO + FFI.
func (c *Client) Transfer(context.Context, aptos.TransactionSigner, aptos.AccountAddress, uint64, aptos.AccountAddress, string, string) (*aptos.Transaction, error) {
	return nil, ErrCGODisabled
}
