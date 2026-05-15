# internal/crypto (planned)

Port TwistedEd25519, TwistedElGamal, `EncryptedAmount`, and chunked amount helpers from:

- `confidential-asset/src/crypto/twistedEd25519.ts`
- `confidential-asset/src/crypto/twistedElGamal.ts`
- `confidential-asset/src/crypto/encryptedAmount.ts`
- `confidential-asset/src/crypto/chunkedAmount.ts`

This directory is reserved for pure-Go crypto that must byte-match the TypeScript SDK for on-chain proofs.
