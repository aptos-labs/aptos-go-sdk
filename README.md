[![Go Reference](https://pkg.go.dev/badge/github.com/qimeila/aptos-go-sdk.svg)](https://pkg.go.dev/github.com/qimeila/aptos-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/qimeila/aptos-go-sdk)](https://goreportcard.com/report/github.com/qimeila/aptos-go-sdk)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aptos-labs/aptos-go-sdk)
[![GitHub Tag](https://img.shields.io/github/v/tag/aptos-labs/aptos-go-sdk?label=Latest%20Version)](https://pkg.go.dev/github.com/qimeila/aptos-go-sdk)

# aptos-go-sdk

An SDK for the Aptos blockchain in Go.

## Getting started

Add go to your `go.mod` file

```bash
go get -u  github.com/qimeila/aptos-go-sdk
```

## Where can I see examples?

Take a look at `examples/` for some examples of how to write clients.

## Where can I learn more?

You can read more about the Go SDK documentation on [aptos.dev](https://aptos.dev/sdks/go-sdk/)

## Development

1. Make your changes
2. Update the CHANGELOG.md
3. Run `gofumpt -l -w .`
4. Run `golangci-lint run`
5. Commit with a good description
6. Submit a PR

# How to publish

1. Update changelog with a pull request
2. Create a new tag via e.g. v1.1.0 with the list of changes