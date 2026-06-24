// Package hd implements BIP-39 mnemonic handling and SLIP-0010 Ed25519 key derivation
// for Aptos hierarchical deterministic wallets.
package hd

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

const (
	// Slip10Ed25519Seed is the HMAC-SHA512 key used for SLIP-0010 Ed25519 master derivation.
	Slip10Ed25519Seed = "ed25519 seed"

	// HardenedOffset is added to each path segment index for hardened derivation.
	HardenedOffset = 0x80000000

	// AptosCoinType is the BIP-44 registered coin type for Aptos (SLIP-0044).
	AptosCoinType = 637

	// DefaultDerivationPath is the standard Petra-style Aptos Ed25519 derivation path.
	DefaultDerivationPath = "m/44'/637'/0'/0'/0'"
)

var aptosHardenedPathRegex = regexp.MustCompile(`^m/44'/637'/[0-9]+'/[0-9]+'/[0-9]+'?`)

// NormalizeMnemonic trims whitespace, lowercases each word, and joins with single spaces.
func NormalizeMnemonic(mnemonic string) string {
	parts := strings.Fields(strings.TrimSpace(mnemonic))
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	return strings.Join(parts, " ")
}

// ValidateMnemonic reports whether the mnemonic is a valid BIP-39 phrase.
func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(NormalizeMnemonic(mnemonic))
}

// MnemonicToSeed derives a 64-byte BIP-39 seed from a mnemonic and optional passphrase.
func MnemonicToSeed(mnemonic, passphrase string) ([]byte, error) {
	normalized := NormalizeMnemonic(mnemonic)
	if !bip39.IsMnemonicValid(normalized) {
		return nil, errors.New("invalid BIP-39 mnemonic")
	}
	return bip39.NewSeed(normalized, passphrase), nil
}

// IsValidHardenedPath reports whether path matches the Aptos SLIP-0010 BIP-44 format:
// m/44'/637'/{account}'/{change}'/{address}'.
func IsValidHardenedPath(path string) bool {
	return aptosHardenedPathRegex.MatchString(path)
}

// SplitPath parses a BIP-44 derivation path into numeric segments (without the leading "m").
func SplitPath(path string) ([]uint32, error) {
	if !strings.HasPrefix(path, "m/") {
		return nil, fmt.Errorf("derivation path must start with m/: %q", path)
	}

	parts := strings.Split(path, "/")[1:]
	if len(parts) == 0 {
		return nil, errors.New("derivation path has no segments")
	}

	segments := make([]uint32, len(parts))
	for i, part := range parts {
		cleaned := strings.TrimSuffix(part, "'")
		if cleaned == "" {
			return nil, fmt.Errorf("invalid path segment %q", part)
		}

		var value uint64
		for _, ch := range cleaned {
			if ch < '0' || ch > '9' {
				return nil, fmt.Errorf("invalid path segment %q", part)
			}
			value = value*10 + uint64(ch-'0')
		}
		if value > 0xFFFFFFFF {
			return nil, fmt.Errorf("path segment out of range: %q", part)
		}
		segments[i] = uint32(value)
	}
	return segments, nil
}

type derivedKeys struct {
	key       []byte
	chainCode []byte
}

func deriveKey(hmacKey, data []byte) derivedKeys {
	mac := hmac.New(sha512.New, hmacKey)
	_, _ = mac.Write(data)
	digest := mac.Sum(nil)
	return derivedKeys{
		key:       digest[:32],
		chainCode: digest[32:],
	}
}

func ckdPriv(parent derivedKeys, index uint32) derivedKeys {
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)

	data := make([]byte, 1+len(parent.key)+4)
	data[0] = 0x00
	copy(data[1:], parent.key)
	copy(data[1+len(parent.key):], indexBytes)

	return deriveKey(parent.chainCode, data)
}

// DeriveEd25519PrivateKey derives a 32-byte Ed25519 seed from a BIP-39 seed and hardened path.
func DeriveEd25519PrivateKey(path string, seed []byte) ([]byte, error) {
	if !IsValidHardenedPath(path) {
		return nil, fmt.Errorf("invalid Aptos hardened derivation path: %q", path)
	}

	segments, err := SplitPath(path)
	if err != nil {
		return nil, err
	}

	master := deriveKey([]byte(Slip10Ed25519Seed), seed)
	current := master
	for _, segment := range segments {
		current = ckdPriv(current, segment+HardenedOffset)
	}
	return current.key, nil
}
