package confidentialasset

import (
	"context"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/cipherparse"
)

// ConfidentialBalance holds decrypted available and pending balances in octas (TS ConfidentialBalance.availableBalance / pendingBalance).
type ConfidentialBalance struct {
	AvailableOctas uint64
	PendingOctas   uint64
}

const twistedDecryptionDerivationMessage = "Sign this message to derive decryption key from your private key"

// TwistedDecryptionKey32 returns 32-byte TwistedEd25519 decryption key from env-style hex or from Ed25519 account signature (TS default).
func TwistedDecryptionKey32(acct *account.Account, twistedHex string) ([32]byte, error) {
	var out [32]byte
	if twistedHex != "" {
		raw, err := decodeHex32(twistedHex)
		if err != nil {
			return out, fmt.Errorf("twisted key hex: %w", err)
		}
		if len(raw) != 32 {
			return out, ErrInvalidTwistedKey
		}
		copy(out[:], raw)
		return out, nil
	}
	sig, err := acct.SignMessage([]byte(twistedDecryptionDerivationMessage))
	if err != nil {
		return out, err
	}
	b := sig.Bytes()
	return twistedKeyFromSigBig(b)
}

// FetchBalanceCipherChunks loads compressed C,D chunks from a balance view (no CGO).
func (c *Client) FetchBalanceCipherChunks(ctx context.Context, account, token aptos.AccountAddress, viewFunction string) (cChunks, dChunks [][]byte, err error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: viewFunction,
		TypeArgs: nil,
		Args:     []any{account.String(), token.String()},
	})
	if err != nil {
		return nil, nil, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil {
		return nil, nil, err
	}
	return cipherparse.ParseCipherChunks(arr)
}
