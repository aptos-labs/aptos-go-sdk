package confidentialasset

import (
	"context"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/movearg"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigma"
)

// RegisterBalance submits register_raw (sigma registration proof).
func (c *Client) RegisterBalance(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, twistedHex, faMetadataHex string) (*aptos.Transaction, error) {
	acct, ok := signer.(*account.Account)
	if !ok {
		return nil, fmt.Errorf("register: signer must be *account.Account for twisted key derivation")
	}
	dk32, err := TwistedDecryptionKey32(acct, twistedHex)
	if err != nil {
		return nil, err
	}
	ch, err := c.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	sender := signer.Address()
	tok := token
	proof, err := sigma.ProveRegistration(dk32, sender, tok, ch)
	if err != nil {
		return nil, err
	}
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk32)
	if err != nil {
		return nil, err
	}
	payload := &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: "register_raw",
		TypeArgs: nil,
		Args: []any{
			&token,
			movearg.VectorU8(ek),
			movearg.VectorVectorU8(proof.Commitment),
			movearg.VectorVectorU8(proof.Response),
		},
	}
	return c.SubmitWithSimulatedGas(ctx, signer, "register_raw", payload, faMetadataHex)
}

// RotateEncryptionKey submits rotate_encryption_key_raw (sigma key rotation + re-encrypted D).
func (c *Client) RotateEncryptionKey(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, oldTwistedHex, newTwistedHex, faMetadataHex string) (*aptos.Transaction, error) {
	acct, ok := signer.(*account.Account)
	if !ok {
		return nil, fmt.Errorf("rotate: signer must be *account.Account for twisted key derivation")
	}
	oldDK, err := TwistedDecryptionKey32(acct, oldTwistedHex)
	if err != nil {
		return nil, err
	}
	newDK, err := TwistedDecryptionKey32(acct, newTwistedHex)
	if err != nil {
		return nil, fmt.Errorf("new twisted key: %w", err)
	}

	_, dRows, err := c.FetchBalanceCipherChunks(ctx, signer.Address(), token, "get_available_balance")
	if err != nil {
		return nil, err
	}
	ch, err := c.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	sender := signer.Address()
	tok := token
	kr, err := sigma.ProveKeyRotation(oldDK, newDK, dRows, sender, tok, ch)
	if err != nil {
		return nil, err
	}
	payload := &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: "rotate_encryption_key_raw",
		TypeArgs: nil,
		Args: []any{
			&token,
			movearg.VectorU8(kr.NewEkBytes),
			true, // unpause (TS default)
			movearg.VectorVectorU8(kr.NewDBytes),
			movearg.VectorVectorU8(kr.Proof.Commitment),
			movearg.VectorVectorU8(kr.Proof.Response),
		},
	}
	return c.SubmitWithSimulatedGas(ctx, signer, "rotate_encryption_key_raw", payload, faMetadataHex)
}
