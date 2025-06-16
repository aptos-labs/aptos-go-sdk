# Aptos Go SDK Changelog

All notable changes to the Aptos Go SDK will be captured in this file. This changelog is written by hand for now. It
adheres to the format set out by [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# Unreleased
- [`Feature`] Add orderless transaction support

# v1.9.0 (6/8/2025)

- [`Breaking`] Change how plain text arguments are parsed from inputs. All inputs now have to be []any and will work
  nested properly.
- [`Feature`] Add `CompatibilityMode` to `ConvertArg` so that it can be used to convert arguments similar to the original TS SDK.

# v1.8.0 (6/3/2025)

- [`Feature`] Added `EventsByCreationNumber` API to fetch events by creation number for a given account
- [`Fix`] Not set nil `start` argument for `EventsByHandle` to 0 to allow fetching latest events

# v1.7.0 (4/25/2025)

- [`Feature`] Added `String()` method to `Ed25519PrivateKey` and `Secp256k1PrivateKey` which exports the private key in
  the AIP-80 format
- [`Fix`] Add missing `Version` field to Event struct
- [`Deps`] Update dependencies including go crypto for security purposes
- [`Fix`] Add `MarshalJSON` support for `HexBytes` to match custom `UnmarshalJSON` logic

# v1.6.2 (3/28/2025)

- Limit generics to max 255 in types
- [`Fix`] Add coverage and fix broken view calls in FungibleAssetClient

# v1.6.1 (3/27/2025)

- [`Refactor`] Remove unnecessary atomic swap in submit transactions
- [`Fix`] Enable users to simulate gas fee payer and multi agent transactions

# v1.6.0 (3/24/2025)

- [`Feature`] Add the ability for ABI simple type conversion of entry function arguments with both remote and local ABI
- Add ability to simulate for any transaction including multi-agent and fee payer
- [`Fix`] Ensure proper cleanup of response body on read error to prevent potential memory leak.
- [`Fix`] Fixes possible conflicts between signatures of multiple goroutines
- [`Breaking`] Change ExpirationSeconds from int64 to uint64
- [`Fix`] Fix a self calling function in the client
- [`Fix`/`Breaking`] Fix fungible asset API calls that didn't return a value
- [`Fix`] Fix errors from failed transactions in concurrent submission
- [`Fix`] Fix AptosRpcClient, AptosFaucetClient, AptosIndexerClient interfaces to be enforced on concrete types properly
- [`Fix`] A few error case fixes and lint updates

# v1.5.0 (2/10/2025)

- [`Fix`] Make NodeClient match AptosRpcClient interface
- [`Dependency`] Update `golang.org/x/crypto` to `v0.32.0`
- [`Dependency`] Update `github.com/hasura/go-graphql-client` to `v0.13.1`
- [`Fix`] Make NodeClient satisfy AptosRpcClient interface
- [`Breaking`] Change Account signer input to use Address instead of AuthKey, authkey comes from the private key

# v1.4.1 (01/02/2025)

- [`Fix`] Non struct events, will have a field named `__any_data__` that contains the actual data of the struct

# v1.4.0 (12/09/2024)

- [`Breaking`] SignedTransaction can only contain `RawTransaction`. Note, this is technically breaking, but should not
  change any user behavior or code.
- [`Fix`] SignedTransaction deserialization will properly initialized `nil` fields before deserializing.

# v1.3.0 (11/25/2024)

- Add AIP-80 support for Ed25519 and Secp256k1 private keys
- Add support for optional ledger version in FA functions and APT balance
- [`Breaking`] Change from the `go-ethereum/crypto` to `decred/dcrd` for secp256k1 signing
- [`Breaking`] Add checks for malleability to prevent duplicate secp256k1 signatures in verification and to ensure
  correct on-chain behavior
- Adds functionality to recover public keys from secp256k1 signatures

# v1.2.0 (11/15/2024)

- [`Fix`][`Breaking`] Fix MultiKey implementation to be more consistent with the rest of the SDKs
- Add BCS support for optional values

# v1.1.0 (11/07/2024)

- Add icon uri and project uri to FA client, reduce duplicated code
- Add better error messages around script argument parsing
- Add example for scripts with FA

# v1.0.0 (10/28/2024)

- [`Fix`] Paginate transactions properly if no given start version
- Add account transactions API

# v0.7.0 (8/19/2024)

- [`Fix`] Parse GenesisTransaction properly
- [`Fix`] Ensure if no block transactions are requested, it doesn't fail to fetch a block
- [`Doc`] Fix comment from milliseconds to microseconds
- [`Fix`] Fix GUID parsing for events
- Use ed25519-consensus to ensure signatures are verified in a ZIP215 compatible way
- [`Fix`] Fix MultiKey signature verification and building to work with any keys

# v0.6.0 (6/28/2024)

- [`Breaking`] Change type from Transaction to CommittedTransaction for cases that it's known they're committed
- [`Fix`] Fix secp256k1 signing and verification to be correctly used
- [`Fix`] Fix supply view function for FungibleAssetClient
- [`Breaking`] Rearrange Concurrency and add new types to carry between steps
- [`Fix`] Fix some of the API types that didn't match on-chain representations
- Add Go doc for most functions and types in the codebase
- Add tons more testing
- [`Breaking`] Change ToAnyPublicKey to have an error in the output
- [`Fix`] Properly parse DeleteResourse write sets
- Add batch transaction submit API
- Add ability to set an API bearer key and other arbitrary headers
- Upgrade concurrent APIs to top level
- Add example for comparing concurrent and non-concurrent APIs
- Add BlockEpilogueTransaction support
- [`Breaking`] Make Secp256k1Signature fixed length
- Change signers to have simulation authenticators by default

# v0.5.0 (6/21/2024)

- [`Fix`] Fix sponsored transactions, and add example for sponsored transactions
- Add examples to CI
- Add node health check API to client

# v0.4.1 (6/18/2024)

- [`Fix`] Make examples more portable / not rely on internal packages

# v0.4.0 (6/17/2024)

- [`Fix`] Fix all unhandled errors using ineffassign to find them
- [`Doc`] Add examples and documentation in Go doc comments
- [`Breaking change`] Make SignedTransaction.Hash() output string to be consistent with other representations
- Add more functions available from under util
- Add CoinBatchTransferPayload for sending multiple amounts to multiple addresses
- [`Fix`] Block APIs will now pull the rest of the transactions for the block automatically
- [`Fix`] Fix bytecode JSON parsing in transaction parsing
- Add concurrent APIs, and a transaction handler for a single account
- Add SimulateTransaction API for single-key authenticated transactions

# v0.2.0 (6/10/2024)

- [`Breaking`] Some refactoring of names to be proper camel casing
- [`Fix`] fixed bug with transactions listing
- Refactored all pieces into new packages, this may break previous users
- [`Fix`] Misspelling of expiration time
- Added documentation for many functions and structs
- Added all remaining Type tags
- [`Fix`] Improved type tag parsing and printing for all types, including vector
- [`Fix`] Fixed bug in deserializing bool
- Added significantly more test coverage, including for scripts
- Add block APIs
- Add private key imports, and authentication key
- Add signed transaction hashes
- [`Fix`] Private key imports without public key
- [`Breaking`] Change to structured types over `map[string]any` when possible on outputs
- [`Breaking`] Add types for signatures rather than raw bytes
- [`Fix`] Fix localhost faucet endpoint
- Re-export types from internal account file to external account file
- Add more crypto support for MultiEd25519 and single sender
- [`Breaking`] Change ledgerVersion arg to an uint64 to be more accurate to the possible inputs
- Add single serialize functions to simplify duplicated code
- [`Breaking`] Several transaction authenticator and authenticator types were changed on transaction submission
- Add TransactionSigner
- Fix FungibleAssetClient
- Add predetermined payloads
- Add support for deploying code
- Add support for on-chain multi-sig with an example
- Fix secp256k1 signing, and multikey signing

# v0.1.0 (5/7/2024)

- Ed25519 support
- Ed25519 transaction support
- View function support
- Resource lookup support (via JSON and possibly BCS)
- Transaction lookup by hash and by version
- Faucet support
- BCS support
- Some object address derivation support
- Resource account address derivation support
