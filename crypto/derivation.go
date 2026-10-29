package crypto

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/crypto/hd"
)

// DefaultDerivationPath is the standard Petra-style Aptos Ed25519 BIP-44 path.
const DefaultDerivationPath = hd.DefaultDerivationPath

// ValidateMnemonic reports whether the mnemonic is a valid BIP-39 phrase.
func ValidateMnemonic(mnemonic string) bool {
	return hd.ValidateMnemonic(mnemonic)
}

// SlhDsaSha2128sPrivateKeyFromDerivationPath derives an SLH-DSA-SHA2-128s private
// key from a BIP-39 mnemonic and a fully hardened derivation path, matching
// the TypeScript SDK: SLIP-0010 with the custom "SLH-DSA-SHA2-128s seed" HMAC
// key, then threeSeeds = derivedKey || derivedChainCode[0:16]. Only hardened
// paths are supported, as SLH-DSA has no key homomorphism.
func SlhDsaSha2128sPrivateKeyFromDerivationPath(mnemonic, path, passphrase string) (*SlhDsaSha2128sPrivateKey, error) {
	seed, err := hd.MnemonicToSeed(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}
	privateKeyBytes, err := hd.DeriveSlhDsaSha2128sPrivateKey(path, seed)
	if err != nil {
		return nil, err
	}
	key := &SlhDsaSha2128sPrivateKey{}
	if err := key.FromBytes(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to load derived SLH-DSA-SHA2-128s key: %w", err)
	}
	return key, nil
}

// Ed25519PrivateKeyFromDerivationPath derives a legacy Ed25519 private key from a
// BIP-39 mnemonic and Aptos BIP-44 hardened path (e.g. m/44'/637'/0'/0'/0').
func Ed25519PrivateKeyFromDerivationPath(mnemonic, path, passphrase string) (*Ed25519PrivateKey, error) {
	seed, err := hd.MnemonicToSeed(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := hd.DeriveEd25519PrivateKey(path, seed)
	if err != nil {
		return nil, err
	}

	key := &Ed25519PrivateKey{}
	if err := key.FromBytes(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to load derived Ed25519 key: %w", err)
	}
	return key, nil
}
