[![Go Reference](https://pkg.go.dev/badge/github.com/aptos-labs/aptos-go-sdk.svg)](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptos-labs/aptos-go-sdk)](https://goreportcard.com/report/github.com/aptos-labs/aptos-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aptos-labs/aptos-go-sdk)
[![GitHub Tag](https://img.shields.io/github/v/tag/aptos-labs/aptos-go-sdk?label=Latest%20Version)](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk)

# aptos-go-sdk

An SDK for the Aptos blockchain in Go.

## Getting started

Add go to your `go.mod` file

```bash
go get -u  github.com/aptos-labs/aptos-go-sdk
```

## Where can I see examples?

Take a look at `examples/` for some examples of how to write clients.

## Where can I learn more?

You can read more about the Go SDK documentation on [aptos.dev](https://aptos.dev/sdks/go-sdk/)

## Feature support

- [x] BCS encoding and decoding
- [x] Structured API parsing
- [x] Ed25519 Signer support
- [x] MultiEd25519 support
- [x] Secp256k1 Signer support
- [x] On-chain and off-chain multi-sig support
- [x] Sponsored transaction and Multi-agent support
- [x] Fungible Asset support
- [x] Indexer support with limited queries
- [x] Transaction submission and waiting
- [x] External signer support e.g. HSMs or external services
- [x] Move Package publishing support
- [x] Move script support
- [x] Transaction Simulation
- [x] Automated sequence number management for parallel transaction submission

### TODO

- [ ] Predetermined Indexer queries for Fungible Assets and Digital Assets

## Examples

- [x] Transaction signing
- [x] Fungible asset usage
- [x] External and alternative signing methods
- [x] On-chain multi-sig
- [x] Performance differences between transaction submission methods
- [x] Move package publishing support
- [x] Sponsored transaction example
- [x] Off-chain multi-sig example

### TODO

- [ ] Multi-agent example
- [x] Script Example
- [ ] Digital assets / NFTs example
- [ ] Transaction parsing example (by blocks)

## Other TODO

- [x] Ensure blocks fetch all transactions associated
- [ ] More testing around API parsing
- [ ] TypeTag string parsing
- [x] Add examples into the documentation


# How to publish
1. Update changelog with a pull request
2. Create a new tag via e.g. v1.1.0 with the list of changes