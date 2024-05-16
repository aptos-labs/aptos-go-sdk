# Aptos Go SDK Changelog

All notable changes to the Aptos Go SDK will be captured in this file. This changelog is written by hand for now. It
adheres to the format set out by [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# Unreleased

- Refactored all pieces into new packages, this may break previous users
- [`Fix`] Misspelling of expiration time
- Added documentation for many functions and structs
- Added all remaining Type tags
- [`Fix`] Improved type tag parsing and printing for all types, including vector
- [`Fix`] Fixed bug in deserializing bools
- Added significantly more test coverage, including for scripts

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