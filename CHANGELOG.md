# Aptos Go SDK Changelog

All notable changes to the Aptos Go SDK will be captured in this file. This changelog is written by hand for now. It
adheres to the format set out by [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# Unreleased

- [`Fix`] Fix all unhandled errors using ineffassign to find them
- [`Doc`] Add examples and documentation in Go doc comments
- [`Breaking change`] Make SignedTransaction.Hash() output string to be consistent with other representations
- Add more functions available from under util
- Add CoinBatchTransferPayload for sending multiple amounts to multiple addresses
- [`Fix`] Block APIs will now pull the rest of the transactions for the block automatically

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
