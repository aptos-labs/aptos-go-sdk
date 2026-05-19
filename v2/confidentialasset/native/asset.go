//go:build cgo

package native

import (
	"context"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
)

// ConfidentialAsset mirrors the TS class name; embeds native Client for FFI-backed entrypoints.
type ConfidentialAsset struct {
	*Client
}

// NewConfidentialAsset wraps a base confidentialasset.Client with native FFI methods.
func NewConfidentialAsset(c *confidentialasset.Client) *ConfidentialAsset {
	return &ConfidentialAsset{Client: Wrap(c)}
}

// RolloverOpts matches TS rolloverPendingBalance optional behavior.
type RolloverOpts struct {
	TwistedHex        string // for auto-normalize when balance not normalized (same as TS senderDecryptionKey input path)
	WithPauseIncoming bool
	FAMetadataHex     string // FA used for public APT gas balance lookup (e.g. 0xa metadata)
}

// RolloverPendingBalance mirrors TS: if not normalized and TwistedHex is set, normalize first.
func (a *ConfidentialAsset) RolloverPendingBalance(ctx context.Context, signer *account.Account, token aptos.AccountAddress, opts RolloverOpts) ([]*aptos.Transaction, error) {
	var out []*aptos.Transaction
	norm, err := a.IsBalanceNormalized(ctx, signer.Address(), token)
	if err != nil {
		return nil, err
	}
	if !norm {
		if opts.TwistedHex == "" {
			return nil, fmt.Errorf("rollover: balance not normalized and no twisted decryption key provided")
		}
		if _, err := a.NormalizeBalance(ctx, signer, token, opts.TwistedHex, opts.FAMetadataHex); err != nil {
			return nil, err
		}
	}
	tx, err := a.Client.RolloverPendingBalance(ctx, signer, token, opts.WithPauseIncoming, opts.FAMetadataHex)
	if err != nil {
		return nil, err
	}
	out = append(out, tx)
	return out, nil
}
