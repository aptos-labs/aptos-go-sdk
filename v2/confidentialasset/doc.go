// Package confidentialasset provides a Go SDK for Aptos confidential fungible assets,
// aligned with the TypeScript package @aptos-labs/confidential-asset.
//
// It wraps github.com/aptos-labs/aptos-go-sdk/v2 for views, transaction building, and submission
// against the on-chain confidential_asset Move module (default deployer 0x1).
//
// # Requirements
//
// Balance decryption, range proofs, and proof-bearing entry functions (register_raw, normalize_raw,
// withdraw_to_raw, confidential_transfer_raw, rotate_encryption_key_raw) require:
//
//   - CGO_ENABLED=1
//   - A built libaptos_confidential_asset_ffi from github.com/aptos-labs/confidential-asset-bindings
//     (see confidentialasset/README.md)
//
// Without CGO, views and Deposit / RolloverPendingBalance still work; proof paths return ErrCGODisabled.
//
// # Quick start
//
//	client, err := aptos.NewClient(aptos.Testnet)
//	ca := confidentialasset.NewClient(client,
//	    confidentialasset.WithRESTBaseURL("https://fullnode.testnet.aptoslabs.com/v1"),
//	)
//	ok, _ := ca.HasUserRegistered(ctx, account, token)
//	tx, _ := ca.Deposit(ctx, signer, token, amountOctas, faMetadataHex)
//
// See v2/examples/confidential_asset (balance, register, transfer, deposit_chain, ffismoke).
package confidentialasset
