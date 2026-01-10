// Package keyless provides Keyless (Zero-Knowledge) authentication for Aptos.
//
// Keyless authentication allows users to authenticate using their existing
// OpenID Connect (OIDC) identities from providers like Google, Apple, and others,
// without needing to manage private keys directly.
//
// # Overview
//
// The keyless flow consists of several steps:
//
//  1. Generate an ephemeral key pair for the session
//  2. User authenticates with their OIDC provider (Google, Apple, etc.)
//  3. Get the JWT token from the authentication
//  4. Derive the keyless account address from the JWT
//  5. Generate a ZK proof for transaction signing
//  6. Sign and submit transactions
//
// # Basic Usage
//
//	// 1. Generate ephemeral key pair
//	ephemeralKeyPair, err := keyless.GenerateEphemeralKeyPair()
//	if err != nil {
//		return err
//	}
//
//	// 2. Get nonce for OIDC redirect
//	nonce := ephemeralKeyPair.Nonce()
//
//	// 3. After OIDC authentication, create the account
//	account, err := keyless.DeriveAccount(keyless.DeriveAccountConfig{
//		JWT:              jwtToken,
//		EphemeralKeyPair: ephemeralKeyPair,
//		ProverURL:        "https://prover.aptoslabs.com",
//	})
//
//	// 4. Use the account to sign transactions
//	txn, err := client.BuildTransaction(ctx, account.Address(), payload)
//	signed, err := account.SignTransaction(txn)
//
// # Ephemeral Keys
//
// Ephemeral keys are temporary key pairs used to generate zero-knowledge proofs.
// They should be:
//   - Generated fresh for each session
//   - Stored securely during the session
//   - Discarded after expiration
//
// # Supported OIDC Providers
//
// Currently supported providers:
//   - Google
//   - Apple
//   - Facebook
//   - Discord
//
// # Security Considerations
//
//   - JWTs have expiration times; accounts derived from them will also expire
//   - Ephemeral keys should not be persisted long-term
//   - Use secure storage for ephemeral keys during sessions
//   - The prover service must be trusted
package keyless
