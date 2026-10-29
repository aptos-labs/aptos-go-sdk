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

	// Slip10SlhDsaSha2128sSeed is the custom, non-standard SLIP-0010 HMAC key ("curve"
	// parameter) for SLH-DSA-SHA2-128s, matching the TypeScript SDK (aptos-ts-sdk#802)
	// and AIP-137. It is not a registered SLIP-0010 curve name.
	Slip10SlhDsaSha2128sSeed = "SLH-DSA-SHA2-128s seed"

	// SlhDsaSha2128sPrivateKeyLength is the serialized SLH-DSA-SHA2-128s private-key
	// length: SK seed (16) || PRF seed (16) || PK seed (16). The PK root (the remaining
	// 16 bytes of the FIPS 205 secret key) is derived from these and not stored.
	SlhDsaSha2128sPrivateKeyLength = 48
)

var aptosHardenedPathRegex = regexp.MustCompile(`^m/44'/637'/[0-9]+'/[0-9]+'/[0-9]+'?$`)

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
// The final address segment may omit the trailing apostrophe (e.g. m/44'/637'/0'/0'/0).
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

// IsValidSlhDsaHardenedPath reports whether path is "m" or a fully hardened path
// (every segment suffixed with '). SLH-DSA has no key homomorphism, so only
// hardened derivation is possible. Unlike [IsValidHardenedPath] this is not
// restricted to the Aptos BIP-44 shape, matching the TypeScript SDK.
func IsValidSlhDsaHardenedPath(path string) bool {
	if path == "m" {
		return true
	}
	if !strings.HasPrefix(path, "m/") {
		return false
	}
	for _, part := range strings.Split(path, "/")[1:] {
		if len(part) < 2 || !strings.HasSuffix(part, "'") {
			return false
		}
		if _, err := parseSegment(strings.TrimSuffix(part, "'")); err != nil {
			return false
		}
	}
	return true
}

func parseSegment(cleaned string) (uint32, error) {
	if cleaned == "" {
		return 0, errors.New("empty path segment")
	}
	var value uint64
	for _, ch := range cleaned {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid path segment %q", cleaned)
		}
		value = value*10 + uint64(ch-'0')
	}
	if value >= HardenedOffset {
		return 0, fmt.Errorf("path segment out of range: %q", cleaned)
	}
	return uint32(value), nil
}

// DeriveSlhDsaSha2128sPrivateKey derives the 48-byte SLH-DSA-SHA2-128s private
// key (three 16-byte seeds) from a BIP-39 seed and hardened path, matching
// SlhDsaSha2128sPrivateKey.fromDerivationPath in the TypeScript SDK: SLIP-0010
// with the custom "SLH-DSA-SHA2-128s seed" HMAC key, then
// threeSeeds = derivedKey (32 bytes) || derivedChainCode[0:16].
func DeriveSlhDsaSha2128sPrivateKey(path string, seed []byte) ([]byte, error) {
	if !IsValidSlhDsaHardenedPath(path) {
		return nil, fmt.Errorf("invalid SLH-DSA hardened derivation path: %q", path)
	}
	current := deriveKey([]byte(Slip10SlhDsaSha2128sSeed), seed)
	if path != "m" {
		segments, err := SplitPath(path)
		if err != nil {
			return nil, err
		}
		for i, segment := range segments {
			if segment >= HardenedOffset {
				return nil, fmt.Errorf("path segment at index %d (%d) must be less than %d", i, segment, HardenedOffset)
			}
			current = ckdPriv(current, segment+HardenedOffset)
		}
	}
	threeSeeds := make([]byte, 0, SlhDsaSha2128sPrivateKeyLength)
	threeSeeds = append(threeSeeds, current.key...)
	threeSeeds = append(threeSeeds, current.chainCode[:16]...)
	return threeSeeds, nil
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
	for i, segment := range segments {
		if segment >= HardenedOffset {
			return nil, fmt.Errorf("path segment at index %d (%d) must be less than %d", i, segment, HardenedOffset)
		}
		current = ckdPriv(current, segment+HardenedOffset)
	}
	return current.key, nil
}
