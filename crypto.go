package aptos

// Signer a generic interface for any kind of signing
type Signer interface {
	Sign(msg []byte) (authenticator Authenticator, err error)

	// String if a private key, it's bytes, if it's not a private key
	// then a placeholder
	String() string
}

// PrivateKey a generic interface for a signing private key
type PrivateKey interface {
	Signer

	/// PubKey Retrieve the public key for signature verification
	PubKey() PublicKey

	Bytes() []byte

	String() string
}

// PublicKey a generic interface for a public key associated with the private key
type PublicKey interface {
	// Bytes the raw bytes for an authenticator
	Bytes() []byte

	// Scheme The scheme used for address derivation
	Scheme() uint8

	String() string
	// TODO: add verify
}
