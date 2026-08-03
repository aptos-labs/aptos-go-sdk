// Package confidentialasset provides a Go SDK for Aptos confidential fungible assets,
// aligned with the TypeScript package @aptos-labs/confidential-asset.
//
// It wraps github.com/aptos-labs/aptos-go-sdk/v2 for views, transaction building, and submission
// against the on-chain confidential_asset Move module (default deployer 0x1).
//
// # CGO-free usage
//
// Views, Deposit, RolloverPendingBalance, RegisterBalance (sigma-only), and RotateEncryptionKey
// work with CGO_ENABLED=0.
//
// # FFI-backed usage
//
// Balance decryption, range proofs, and normalize/withdraw/transfer entry functions live in
// subpackage confidentialasset/native. Importing native requires CGO_ENABLED=1 and a built
// libaptos_confidential_asset_ffi from github.com/aptos-labs/confidential-asset-bindings
// (see confidentialasset/README.md).
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
//	nc := native.Wrap(ca)
//	bal, _ := nc.GetBalance(ctx, acct, token, twistedHex)
//
// See v2/examples/confidential_asset.
package confidentialasset
