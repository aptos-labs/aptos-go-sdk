package api

import (
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

const (
	EnumSignatureEd25519      = "ed25519_signature"
	EnumSignatureMultiEd25519 = "multi_ed25519_signature"
	EnumSignatureMultiAgent   = "multi_agent_signature"
	EnumSignatureFeePayer     = "fee_payer_signature"
	EnumSignatureSingleSender = "single_sender"
	EnumSignatureSingleKey    = "single_key_signature"
	EnumSignatureMultiKey     = "multi_key_signature"
)

// Signature is an enum of all possible signatures on Aptos
type Signature struct {
	Type  string
	Inner SignatureImpl
}

func (o *Signature) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = data.Type
	switch o.Type {
	case EnumSignatureEd25519:
		o.Inner = &Ed25519Signature{}
	case EnumSignatureMultiAgent:
		o.Inner = &MultiAgentSignature{}
	case EnumSignatureFeePayer:
		o.Inner = &FeePayerSignature{}
	case EnumSignatureSingleSender:
		o.Inner = &SingleSenderSignature{}
	case EnumSignatureMultiEd25519:
		o.Inner = &MultiEd25519Signature{}
	default:
		return fmt.Errorf("unknown signature type: %s", o.Type)
	}
	return json.Unmarshal(b, o.Inner)
}

type SignatureImpl interface{}

// Ed25519Signature represents an Ed25519 public key and signature pair, which actually is the authenticator.
// It's poorly named Ed25519Signature in the API spec
type Ed25519Signature crypto.Ed25519Authenticator

// TODO: apply directly to the upstream type?
func (o *Ed25519Signature) UnmarshalJSON(b []byte) error {
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

type FeePayerSignature struct {
	FeePayerAddress          *types.AccountAddress   `json:"fee_payer_address"`
	FeePayerSigner           *Signature              `json:"fee_payer_signer"`
	SecondarySignerAddresses []*types.AccountAddress `json:"secondary_signer_addresses"`
	SecondarySigners         []*Signature            `json:"secondary_signers"`
	Sender                   *Signature              `json:"sender"`
}

type MultiAgentSignature struct {
	SecondarySignerAddresses []*types.AccountAddress `json:"secondary_signer_addresses"`
	SecondarySigners         []*Signature            `json:"secondary_signers"`
	Sender                   *Signature              `json:"sender"`
}

// TODO: add the MultiEd25519 crypto type directly, and remove this extra redirection
// Note that public keys and signatures should be the same length, unless the transaction failed
type MultiEd25519Signature struct {
	PublicKeys []*crypto.Ed25519PublicKey
	Signatures []*crypto.Ed25519Signature
	Threshold  uint8
	Bitmap     []byte
}

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
