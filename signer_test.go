package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"math/rand/v2"
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
	return key
}

// This is an example for testing, a real signer would be signing and collecting the signatures in a different way
type MultiKeyTestSigner struct {
	Signers            []crypto.Signer
	SignaturesRequired uint8
	MultiKey           *crypto.MultiKey
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

	pubKeys := make([]*crypto.AnyPublicKey, len(signers))
	for i, key := range signers {
		pubKeys[i] = key.PubKey().(*crypto.AnyPublicKey)
	}

	pubKey := &crypto.MultiKey{
		PubKeys:            pubKeys,
		SignaturesRequired: signaturesRequired,
	}

	return &MultiKeyTestSigner{
		Signers:            signers,
		SignaturesRequired: signaturesRequired,
		MultiKey:           pubKey,
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
	indexedSigs := make([]crypto.IndexedAnySignature, s.SignaturesRequired)

	alreadyUsed := make(map[int]bool)

	for i := uint8(0); i < s.SignaturesRequired; i++ {
		// Find a random key
		index := 0
		for {
			index = rand.IntN(len(s.Signers))
			_, present := alreadyUsed[index]
			if !present {
				alreadyUsed[index] = true
				break
			}
		}

		sig, err := s.Signers[index].SignMessage(msg)
		if err != nil {
			return nil, err
		}
		indexedSigs[i] = crypto.IndexedAnySignature{Signature: sig.(*crypto.AnySignature), Index: uint8(index)}
	}

	return crypto.NewMultiKeySignature(indexedSigs)
}

func (s *MultiKeyTestSigner) SimulationAuthenticator() *crypto.AccountAuthenticator {
	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiKey,
		Auth: &crypto.MultiKeyAuthenticator{
			PubKey: s.PubKey().(*crypto.MultiKey),
			Sig:    &crypto.MultiKeySignature{},
		},
	}
}

func (s *MultiKeyTestSigner) AuthKey() *crypto.AuthenticationKey {
	return s.PubKey().AuthKey()
}

func (s *MultiKeyTestSigner) PubKey() crypto.PublicKey {
	return s.MultiKey
}
