package confidentialasset

import "errors"

// Sentinel errors for features still being ported from TypeScript.
var (
	ErrIndexerNotImplemented = errors.New("confidentialasset: indexer GraphQL getActivities not implemented")
	ErrCGODisabled           = errors.New("confidentialasset: build with CGO_ENABLED=1 and confidential-asset-bindings FFI for balance decrypt")
	ErrInvalidTwistedKey     = errors.New("confidentialasset: twisted decryption key must be 32 bytes")
)
