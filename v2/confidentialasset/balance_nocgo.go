//go:build !cgo

package confidentialasset

import (
	"context"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
)

// GetBalance requires CGO (aptosconfidential FFI). Build with CGO_ENABLED=1.
func (c *Client) GetBalance(_ context.Context, _ *account.Account, _ aptos.AccountAddress, _ string) (*ConfidentialBalance, error) {
	return nil, ErrCGODisabled
}
