package api

import (
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

func (o *Signature) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	switch o.Type {
	case EnumSignatureEd25519:
		o.Inner = &Ed25519Authenticator{}
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
	return o.Inner.UnmarshalJSONFromMap(data)
}

type SignatureImpl interface {
	UnmarshalFromMap
}

// Ed25519Authenticator represents an Ed25519 signature, which actually is the authenticator.
type Ed25519Authenticator struct {
	Inner *crypto.Ed25519Authenticator
}

func (o *Ed25519Authenticator) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Inner = &crypto.Ed25519Authenticator{}
	pubKey := &Ed25519PublicKey{}
	err = pubKey.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}
	o.Inner.PubKey = pubKey.Inner
	signature := &Ed25519Signature{}
	err = pubKey.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}
	o.Inner.Sig = signature.Inner
	return nil
}

type Ed25519PublicKey struct {
	Inner *crypto.Ed25519PublicKey
}

func (o *Ed25519PublicKey) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Inner = &crypto.Ed25519PublicKey{}
	hexStr, err := toString(data, "public_key")
	if err != nil {
		return err
	}
	err = o.Inner.FromHex(hexStr)
	return err
}

type Ed25519Signature struct {
	Inner *crypto.Ed25519Signature
}

func (o *Ed25519Signature) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Inner = &crypto.Ed25519Signature{}
	hexStr, err := toString(data, "signature")
	if err != nil {
		return err
	}
	err = o.Inner.FromHex(hexStr)
	return err
}

// SingleSenderSignature is a map of the possible keys, the API needs an update to describe the type of key
// TODO: Implement single sender crypto properly, needs updates on the API side
type SingleSenderSignature struct {
	Inner map[string]any
}

func (o *SingleSenderSignature) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Inner = data
	return nil
}

type FeePayerSignature struct {
	FeePayerAddress          *types.AccountAddress
	FeePayerSigner           *Signature
	SecondarySignerAddresses []*types.AccountAddress
	SecondarySigners         []*Signature
	Sender                   *Signature
}

func (o *FeePayerSignature) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.FeePayerAddress, err = toAccountAddress(data, "fee_payer_address")
	if err != nil {
		return err
	}
	o.FeePayerSigner, err = toSignature(data, "fee_payer_signer")
	if err != nil {
		return err
	}
	multiAgent := &MultiAgentSignature{}
	err = multiAgent.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}
	o.SecondarySignerAddresses = multiAgent.SecondarySignerAddresses
	o.SecondarySigners = multiAgent.SecondarySigners
	o.Sender = multiAgent.Sender
	return nil
}

type MultiAgentSignature struct {
	SecondarySignerAddresses []*types.AccountAddress
	SecondarySigners         []*Signature
	Sender                   *Signature
}

func (o *MultiAgentSignature) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.SecondarySignerAddresses, err = toAccountAddresses(data, "secondary_signer_addresses")
	if err != nil {
		return err
	}
	o.SecondarySigners, err = toSignatures(data, "secondary_signers")
	if err != nil {
		return err
	}
	o.Sender, err = toSignature(data, "sender")
	return err
}

type MultiEd25519Signature struct {
	PublicKeys []*crypto.Ed25519PublicKey
	Signatures []*Ed25519Authenticator
	Threshold  uint8
	Bitmap     []byte
}

func (o *MultiEd25519Signature) UnmarshalJSONFromMap(data map[string]any) (err error) {
	publicKeyData, ok := data["public_keys"].([]any)
	if !ok {
		return fmt.Errorf("public_keys is not an array")
	}
	signatureData, ok := data["authenticators"].([]any)
	if !ok {
		return fmt.Errorf("public_keys is not an array")
	}
	numSignatures := len(signatureData)
	if len(publicKeyData) != numSignatures {
		return fmt.Errorf("public_keys and authenticators length mismatch")
	}
	authenticators := make([]*Ed25519Authenticator, numSignatures)
	for i := 0; i < len(publicKeyData); i++ {
		authenticators[i] = &Ed25519Authenticator{}
		pubKey := &Ed25519PublicKey{}
		pubKeyEntry, ok := publicKeyData[i].(map[string]any)
		if !ok {
			return fmt.Errorf("public_key[%d] is invalid", i)
		}
		err = pubKey.UnmarshalJSONFromMap(pubKeyEntry)
		if err != nil {
			return err
		}
		authenticators[i].Inner.PubKey = pubKey.Inner
		sig := &Ed25519Signature{}
		sigEntry, ok := signatureData[i].(map[string]any)
		if !ok {
			return fmt.Errorf("signature[%d] is invalid", i)
		}
		err = sig.UnmarshalJSONFromMap(sigEntry)
		if err != nil {
			return err
		}
		authenticators[i].Inner.Sig = sig.Inner
	}

	o.Threshold, err = toUint8(data, "threshold")
	if err != nil {
		return err
	}
	o.Bitmap, err = toBytes(data, "bitmap")
	return err
}
