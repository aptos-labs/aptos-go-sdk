package crypto

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/util"
)

// PrivateKeyVariant identifies the type of private key for AIP-80 formatting.
type PrivateKeyVariant string

const (
	PrivateKeyVariantEd25519   PrivateKeyVariant = "ed25519"
	PrivateKeyVariantSecp256k1 PrivateKeyVariant = "secp256k1"
)

// AIP80Prefixes maps key variants to their AIP-80 prefix strings.
// AIP-80 defines a standard format for human-readable private keys.
var AIP80Prefixes = map[PrivateKeyVariant]string{
	PrivateKeyVariantEd25519:   "ed25519-priv-",
	PrivateKeyVariantSecp256k1: "secp256k1-priv-",
}

// FormatPrivateKey formats a private key to AIP-80 compliant string.
// See: https://github.com/aptos-foundation/AIPs/blob/main/aips/aip-80.md
func FormatPrivateKey(privateKey any, keyType PrivateKeyVariant) (string, error) {
	prefix := AIP80Prefixes[keyType]

	var hexStr string
	switch v := privateKey.(type) {
	case string:
		// Remove AIP-80 prefix if present
		if strings.HasPrefix(v, prefix) {
			parts := strings.Split(v, "-")
			v = parts[2]
		}

		bytes, err := util.ParseHex(v)
		if err != nil {
			return "", err
		}
		hexStr = util.BytesToHex(bytes)
	case []byte:
		hexStr = util.BytesToHex(v)
	default:
		return "", errors.New("private key must be string or []byte")
	}

	return fmt.Sprintf("%s%s", prefix, hexStr), nil
}

// ParsePrivateKey parses a private key from hex or AIP-80 format to bytes.
//
// The strict parameter controls AIP-80 enforcement:
//   - nil: Warn about non-AIP-80 format (default)
//   - false: Accept any format silently
//   - true: Require AIP-80 format
func ParsePrivateKey(value any, keyType PrivateKeyVariant, strict ...bool) ([]byte, error) {
	prefix := AIP80Prefixes[keyType]

	var strictness *bool
	if len(strict) > 1 {
		return nil, errors.New("only one strict argument allowed")
	} else if len(strict) == 1 {
		strictness = &strict[0]
	}

	switch v := value.(type) {
	case string:
		if (strictness == nil || !*strictness) && !strings.HasPrefix(v, prefix) {
			bytes, err := util.ParseHex(v)
			if err != nil {
				return nil, err
			}
			// Warn about non-AIP-80 compliance by default
			if strictness == nil {
				// TODO: Use proper logging instead of println
				println("[Aptos SDK] It is recommended that private keys are AIP-80 compliant (https://github.com/aptos-foundation/AIPs/blob/main/aips/aip-80.md)")
			}
			return bytes, nil
		} else if strings.HasPrefix(v, prefix) {
			parts := strings.Split(v, "-")
			if len(parts) != 3 {
				return nil, errors.New("invalid AIP-80 private key format")
			}
			return util.ParseHex(parts[2])
		}
		return nil, errors.New("private key must be AIP-80 compliant when strict mode is enabled")
	case []byte:
		return v, nil
	default:
		return nil, errors.New("private key must be string or []byte")
	}
}
