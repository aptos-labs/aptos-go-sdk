# Aptos Go SDK Changelog

All notable changes to the Aptos Go SDK will be captured in this file. This changelog is written by hand for now. It
adheres to the format set out by [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# Unreleased
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
