# aptos-go-sdk
Aptos Go SDK

## TODO List
### Milestone 1 (Reading from chain)
1. [BCS Support](bcs.go)
    - 90%? Works but more interoperability testing is always good.
2. sequence number reading
    - Done. [RestClient.Account()](client.go)
3. Account / object resource listing
    - Good. JSON map[string]any mode and BCS mode.
4. Transaction waiting / reads (by hash)
    - Rough. TransactionByHash() returns map[string]any rather than struct.

### Milestone 2 (Creating accounts)
1. Faucet support
    - Works. See [faucet.go](faucet.go) and usage in [goclient.go](cmd/goclient/goclient.go)

### Milestone 3 (Writing to chain)
1. ED25519 Private key support
    - Works. See [authenticator.go](authenticator.go)
2. Transaction submission (no fee payer)
    - Works. See [util.go](util.go)  and usage in [goclient.go](cmd/goclient/goclient.go)

### Milestone 4 (View functions)
1. 

### Miscellaneous after
1. Object address derivation
2. Indexer reading support
3. Transaction reading by version
4. View functions
