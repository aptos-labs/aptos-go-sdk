package crypto

import "github.com/aptos-labs/aptos-go-sdk/bcs"

type AccountAuthenticatorNoAccountAuthenticator struct {
}

func (aa *AccountAuthenticatorNoAccountAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {

}

func (aa *AccountAuthenticatorNoAccountAuthenticator) MarshalBCS(ser *bcs.Serializer) {
}

func (aa *AccountAuthenticatorNoAccountAuthenticator) PublicKey() PublicKey {
	var publicKey PublicKey
	err := (publicKey).FromHex("0x0000000000000000000000000000000000000000000000000000000000000000")
	println(publicKey.ToHex())
	if err != nil {

		// Handle error or log it
		// For this case, it should never fail since we're using a valid zero key
	}

	return publicKey
}

// Signature returns the signature of the authenticator
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (ea *AccountAuthenticatorNoAccountAuthenticator) Signature() Signature {
	var signature Signature
	(signature).FromHex("0x0000000000000000000000000000000000000000000000000000000000000000")
	println(signature.ToHex())
	return signature
}

// Verify verifies the signature against the message
//
// Implements:
//   - [AccountAuthenticatorImpl]
func (aa *AccountAuthenticatorNoAccountAuthenticator) Verify(msg []byte) bool {
	return false
}
