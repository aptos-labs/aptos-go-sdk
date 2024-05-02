package aptos

// Signer a generic interface for any kind of signing
type Signer interface {
	Sign(msg []byte) (authenticator Authenticator, err error)
}

// PrivateKey a generic interface for a signing private key
type PrivateKey interface {
	Signer

	/// PubKey Retrieve the public key for signature verification
	PubKey() PublicKey

	Bytes() []byte
}

// PublicKey a generic interface for a public key associated with the private key
type PublicKey interface {
	// BCSStruct The public key must be serializable or it will not be used in transactions
	BCSStruct

	// Bytes the raw bytes for an authenticator
	Bytes() []byte
	// Scheme The scheme used for address derivation
	Scheme() uint8

	// TODO: add verify
}
