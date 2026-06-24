package crypto

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto/hd"
)

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
