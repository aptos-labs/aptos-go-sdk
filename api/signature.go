package api

import (
	"encoding/json"

	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// SignatureVariant is the JSON representation of the signature types
type SignatureVariant string

const (
	SignatureVariantEd25519      SignatureVariant = "ed25519_signature"       // SignatureVariantEd25519 maps to Ed25519Signature
	SignatureVariantMultiEd25519 SignatureVariant = "multi_ed25519_signature" // SignatureVariantMultiEd25519 maps to MultiEd25519Signature
	SignatureVariantMultiAgent   SignatureVariant = "multi_agent_signature"   // SignatureVariantMultiAgent maps to MultiAgentSignature
	SignatureVariantFeePayer     SignatureVariant = "fee_payer_signature"     // SignatureVariantFeePayer maps to FeePayerSignature
	SignatureVariantSingleSender SignatureVariant = "single_sender"           // SignatureVariantSingleSender maps to SingleSenderSignature
	SignatureVariantUnknown      SignatureVariant = "unknown"                 // SignatureVariantUnknown maps to UnknownSignature for unknown types
)

// Signature is an enum of all possible signatures on Aptos
//
// Unknown types will have the Type set to [SignatureVariantUnknown] and the Inner set to [UnknownSignature]
type Signature struct {
	Type  SignatureVariant // Type of the signature, if the signature isn't recognized, it will be [SignatureVariantUnknown]
	Inner SignatureImpl    // Inner is the actual signature
}

// UnmarshalJSON unmarshals the [Signature] from JSON handling conversion between types
func (o *Signature) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	o.Type = SignatureVariant(data.Type)

	switch o.Type {
	case SignatureVariantEd25519:
		o.Inner = &Ed25519Signature{}
	case SignatureVariantMultiAgent:
		o.Inner = &MultiAgentSignature{}
	case SignatureVariantFeePayer:
		o.Inner = &FeePayerSignature{}
	case SignatureVariantSingleSender:
		o.Inner = &SingleSenderSignature{}
	case SignatureVariantMultiEd25519:
		o.Inner = &MultiEd25519Signature{}
	default:
		o.Inner = &UnknownSignature{Type: string(o.Type)}
		o.Type = SignatureVariantUnknown
		return json.Unmarshal(b, &o.Inner.(*UnknownSignature).Payload)
	}
	return json.Unmarshal(b, o.Inner)
}

// SignatureImpl is an interface for all signatures in their JSON formats
type SignatureImpl interface{}

// UnknownSignature is a signature type that is not recognized by the SDK
//
// This is a fallback type for unknown signatures.
type UnknownSignature struct {
	Type    string         // Type is the type of the unknown signature
	Payload map[string]any // Payload is the raw JSON payload
}

// Ed25519Signature represents an Ed25519 public key and signature pair, which actually is the [crypto.AccountAuthenticator].
// It's poorly named Ed25519Signature in the API spec
type Ed25519Signature crypto.Ed25519Authenticator

// UnmarshalJSON deserializes a JSON data blob into an [Ed25519Signature]
func (o *Ed25519Signature) UnmarshalJSON(b []byte) error {
	// TODO: apply directly to the upstream type?
	type inner struct {
		PublicKey HexBytes `json:"public_key"`
		Signature HexBytes `json:"signature"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.PubKey = &crypto.Ed25519PublicKey{}
	err = o.PubKey.FromBytes(data.PublicKey)
	if err != nil {
		return err
	}
	o.Sig = &crypto.Ed25519Signature{}
	return o.Sig.FromBytes(data.Signature)
}

// SingleSenderSignature is a map of the possible keys, the API needs an update to describe the type of key
// TODO: Implement single sender crypto properly, needs updates on the API side
type SingleSenderSignature map[string]any

// FeePayerSignature is a sponsored transaction, the sender can be different from the fee payer of the transaction.
// It can also be multi-agent like [MultiAgentSignature]
type FeePayerSignature struct {
	FeePayerAddress          *types.AccountAddress   `json:"fee_payer_address"`
	FeePayerSigner           *Signature              `json:"fee_payer_signer"`
	SecondarySignerAddresses []*types.AccountAddress `json:"secondary_signer_addresses"`
	SecondarySigners         []*Signature            `json:"secondary_signers"`
	Sender                   *Signature              `json:"sender"`
}

// MultiAgentSignature is a transaction with multiple unique signers, acting on behalf of multiple accounts
type MultiAgentSignature struct {
	SecondarySignerAddresses []*types.AccountAddress `json:"secondary_signer_addresses"`
	SecondarySigners         []*Signature            `json:"secondary_signers"`
	Sender                   *Signature              `json:"sender"`
}

// MultiEd25519Signature is off-chain multi-sig with only Ed25519 keys
type MultiEd25519Signature struct {
	// TODO: add the MultiEd25519 crypto type directly, and remove this extra redirection
	// Note that public keys and signatures should be the same length, unless the transaction failed
	PublicKeys []*crypto.Ed25519PublicKey
	Signatures []*crypto.Ed25519Signature
	Threshold  uint8
	Bitmap     []byte
}

// UnmarshalJSON deserializes a JSON data blob into a [MultiEd25519Signature]
func (o *MultiEd25519Signature) UnmarshalJSON(b []byte) error {
	type inner struct {
		PublicKeys []HexBytes `json:"public_keys"`
		Signatures []HexBytes `json:"signatures"`
		Threshold  uint8      `json:"threshold"`
		Bitmap     HexBytes   `json:"bitmap"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	// For some reason, this is different in structure from Ed25519Signature, making it need custom logic
	o.PublicKeys = make([]*crypto.Ed25519PublicKey, len(data.PublicKeys))
	for i, key := range data.PublicKeys {
		o.PublicKeys[i] = &crypto.Ed25519PublicKey{}
		err = o.PublicKeys[i].FromBytes(key)
		if err != nil {
			return err
		}
	}

	o.Signatures = make([]*crypto.Ed25519Signature, len(data.Signatures))
	for i, signature := range data.Signatures {
		o.Signatures[i] = &crypto.Ed25519Signature{}
		err = o.Signatures[i].FromBytes(signature)
		if err != nil {
			return err
		}
	}
	o.Threshold = data.Threshold
	o.Bitmap = data.Bitmap
	return nil
}
