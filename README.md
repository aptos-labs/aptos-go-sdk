# aptos-go-sdk

Aptos Go SDK

## How does it work?

Take a look at `examples/` for some examples of how to write clients.

## TODO List

### Important

1. `DONE` Indexer support
2. `DONE` Struct support for well known types like transactions e.g. TransactionByHash
3. `DONE` general signer support (should be proven by ^)
4. Basic documentation
5. Additional examples
6. `DONE` External signing by implementing a signer (example)
7. Fee payer support, with example
8. Multi-agent support? And an example?
9. `DONE` On-chain Multi-sig support? and an example?
10. Off-chain multi-sig support? and an example?
11. Simulation
12. Possibly more BCS handling / an easier way to create payloads

### Less important

1. See if there's a better way to handle collection serialization, may need to wrap the collection with a function added
2. Move remaining files into packages, they are partially moved now to keep some separation of code
3. Additional test coverage and examples
4. Parallel submission? Instead of just serially (sequence number handling etc.)
5. Secp256k1 support -> Currently broken, signatures don't come out correctly
