package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

/* This is a collection of test signers, that don't make sense in the real world, but are used for testing */

type MultiEd25519TestSigner struct {
	Keys               []*crypto.Ed25519PrivateKey
	SignaturesRequired uint8
}

func NewMultiEd25519Signer(numKeys uint8, signaturesRequired uint8) (*MultiEd25519TestSigner, error) {
	keys := make([]*crypto.Ed25519PrivateKey, numKeys)
	for i := range keys {
		key, err := crypto.GenerateEd25519PrivateKey(nil)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}

	return &MultiEd25519TestSigner{
		Keys:               keys,
		SignaturesRequired: signaturesRequired,
	}, nil
}

func (s *MultiEd25519TestSigner) AccountAddress() AccountAddress {
	address := AccountAddress{}
	address.FromAuthKey(s.AuthKey())
	return address
}

func (s *MultiEd25519TestSigner) Sign(msg []byte) (authenticator *crypto.AccountAuthenticator, err error) {
	signature, err := s.SignMessage(msg)
	if err != nil {
		return nil, err
	}

	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiEd25519,
		Auth: &crypto.MultiEd25519Authenticator{
			PubKey: s.PubKey().(*crypto.MultiEd25519PublicKey),
			Sig:    signature.(*crypto.MultiEd25519Signature),
		},
	}, nil
}

func (s *MultiEd25519TestSigner) SignMessage(msg []byte) (crypto.Signature, error) {
	signatures := make([]*crypto.Ed25519Signature, s.SignaturesRequired)
	for i := 0; i < int(s.SignaturesRequired); i++ {
		sig, err := s.Keys[i].SignMessage(msg)
		if err != nil {
			return nil, err
		}
		signatures[i] = sig.(*crypto.Ed25519Signature)
	}

	return &crypto.MultiEd25519Signature{
		Signatures: signatures,
		Bitmap:     [4]byte{},
	}, nil
}

func (s *MultiEd25519TestSigner) AuthKey() *crypto.AuthenticationKey {
	return s.PubKey().AuthKey()
}

func (s *MultiEd25519TestSigner) PubKey() crypto.PublicKey {
	pubKeys := make([]*crypto.Ed25519PublicKey, len(s.Keys))
	for i, key := range s.Keys {
		pubKeys[i] = key.PubKey().(*crypto.Ed25519PublicKey)
	}

	key := &crypto.MultiEd25519PublicKey{
		PubKeys:            pubKeys,
		SignaturesRequired: s.SignaturesRequired,
	}
	println(util.PrettyJson(key))
	return key
}

type MultiKeyTestSigner struct {
	Signers            []crypto.Signer
	SignaturesRequired uint8
}

func NewMultiKeyTestSigner(numKeys uint8, signaturesRequired uint8) (*MultiKeyTestSigner, error) {
	signers := make([]crypto.Signer, numKeys)
	for i := range signers {
		switch i % 2 {
		case 0:
			signer, err := crypto.GenerateEd25519PrivateKey()
			if err != nil {
				return nil, err
			}
			// Wrap in a SingleSigner
			signers[i] = &crypto.SingleSigner{Signer: signer}
		case 1:
			signer, err := crypto.GenerateSecp256k1Key()
			if err != nil {
				return nil, err
			}
			// Wrap in a SingleSigner
			signers[i] = &crypto.SingleSigner{Signer: signer}
		}
	}

	return &MultiKeyTestSigner{
		Signers:            signers,
		SignaturesRequired: signaturesRequired,
	}, nil
}

func (s *MultiKeyTestSigner) AccountAddress() AccountAddress {
	address := AccountAddress{}
	address.FromAuthKey(s.AuthKey())
	return address
}

func (s *MultiKeyTestSigner) Sign(msg []byte) (authenticator *crypto.AccountAuthenticator, err error) {
	signature, err := s.SignMessage(msg)
	if err != nil {
		return nil, err
	}

	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiKey,
		Auth: &crypto.MultiKeyAuthenticator{
			PubKey: s.PubKey().(*crypto.MultiKey),
			Sig:    signature.(*crypto.MultiKeySignature),
		},
	}, nil
}

func (s *MultiKeyTestSigner) SignMessage(msg []byte) (crypto.Signature, error) {
	totalSignature := &crypto.MultiKeySignature{}
	totalSignature.Signatures = make([]*crypto.AnySignature, s.SignaturesRequired)

	// TODO: randomize keys used for testing
	for i := uint8(0); i < s.SignaturesRequired; i++ {
		sig, err := s.Signers[i].SignMessage(msg)
		if err != nil {
			return nil, err
		}
		totalSignature.Signatures[i] = sig.(*crypto.AnySignature)

		err = totalSignature.Bitmap.AddKey(i)
		if err != nil {
			return nil, err
		}
	}

	return totalSignature, nil
}

func (s *MultiKeyTestSigner) AuthKey() *crypto.AuthenticationKey {
	return s.PubKey().AuthKey()
}

func (s *MultiKeyTestSigner) PubKey() crypto.PublicKey {
	pubKeys := make([]*crypto.AnyPublicKey, len(s.Signers))
	for i, key := range s.Signers {
		pubKeys[i] = key.PubKey().(*crypto.AnyPublicKey)
	}

	return &crypto.MultiKey{
		PubKeys:            pubKeys,
		SignaturesRequired: s.SignaturesRequired,
	}
}
