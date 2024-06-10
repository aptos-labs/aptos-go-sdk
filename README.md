[![Go Reference](https://pkg.go.dev/badge/github.com/aptos-labs/aptos-go-sdk.svg)](https://pkg.go.dev/github.com/aptos-labs/aptos-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptos-labs/aptos-go-sdk)](https://goreportcard.com/report/github.com/aptos-labs/aptos-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aptos-labs/aptos-go-sdk)
![GitHub Tag](https://img.shields.io/github/v/tag/aptos-labs/aptos-go-sdk?label=Latest%20Version)

# aptos-go-sdk

An SDK for the Aptos blockchain in Go. The SDK is currently in Beta.

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
- [x] Secp256k1 Signer support
- [x] On-chain and off-chain multi-sig support
- [x] Sponsored transaction and Multi-agent support
- [x] Fungible Asset support
- [x] Indexer support with limited queries
- [x] Transaction submission and waiting
- [x] External signer support e.g. HSMs or external services
- [x] Move Package publishing support
- [x] Move script support

### TODO
- [ ] Transaction Simulation
- [ ] MultiEd25519 support
- [ ] Predetermined Indexer queries for Fungible Assets and Digital Assets
- [ ] Automated sequence number management for parallel transaction submission

## Examples

- [x] Transaction signing
- [x] Fungible asset usage
- [x] External and alternative signing methods
- [x] On-chain multi-sig
- [x] Performance differences between transaction submission methods
- [x] Move package publishing support

### TODO

- [ ] Multi-agent example
- [ ] Script Example
- [ ] Sponsored transaction example
- [ ] Off-chain multi-sig example
- [ ] Digital assets / NFTs example
- [ ] Transaction parsing example (by blocks)

## Other TODO
- [ ] Ensure blocks fetch all transactions associated
- [ ] More testing around API parsing
- [ ] TypeTag string parsing
- [ ] Add examples into the documentation