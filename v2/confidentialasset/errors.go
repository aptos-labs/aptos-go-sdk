package confidentialasset

import "errors"

// Sentinel errors for features still being ported from TypeScript.
var (
	ErrIndexerNotImplemented = errors.New("confidentialasset: indexer GraphQL getActivities not implemented")
	ErrInvalidTwistedKey     = errors.New("confidentialasset: twisted decryption key must be 32 bytes")
)
