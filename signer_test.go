package aptos

import (
	"errors"
	"math/rand/v2"

	"github.com/qimeila/aptos-go-sdk/crypto"
	"github.com/qimeila/aptos-go-sdk/internal/util"
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

func (s *MultiEd25519TestSigner) Sign(msg []byte) (*crypto.AccountAuthenticator, error) {
	signature, err := s.SignMessage(msg)
	if err != nil {
		return nil, err
	}
	pubkey, ok := s.PubKey().(*crypto.MultiEd25519PublicKey)
	if !ok {
		return nil, errors.New("invalid MultiEd25519 public key")
	}
	sig, ok := signature.(*crypto.MultiEd25519Signature)
	if !ok {
		return nil, errors.New("invalid MultiEd25519 signature")
	}

	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiEd25519,
		Auth: &crypto.MultiEd25519Authenticator{
			PubKey: pubkey,
			Sig:    sig,
		},
	}, nil
}

func (s *MultiEd25519TestSigner) SignMessage(msg []byte) (crypto.Signature, error) {
	signatures := make([]*crypto.Ed25519Signature, s.SignaturesRequired)
	for i := range s.SignaturesRequired {
		sig, err := s.Keys[i].SignMessage(msg)
		if err != nil {
			return nil, err
		}
		typedSig, ok := sig.(*crypto.Ed25519Signature)
		if !ok {
			return nil, errors.New("invalid Ed25519 signature")
		}
		signatures[i] = typedSig
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
		pubkey, ok := key.PubKey().(*crypto.Ed25519PublicKey)
		if !ok {
			return nil
		}
		pubKeys[i] = pubkey
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
		pubkey, ok := key.PubKey().(*crypto.AnyPublicKey)
		if !ok {
			return nil, errors.New("invalid MultiEd25519 public key")
		}
		pubKeys[i] = pubkey
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

func (s *MultiKeyTestSigner) Sign(msg []byte) (*crypto.AccountAuthenticator, error) {
	signature, err := s.SignMessage(msg)
	if err != nil {
		return nil, err
	}

	pubkey, ok := s.PubKey().(*crypto.MultiKey)
	if !ok {
		return nil, errors.New("invalid Multikey public key")
	}
	sig, ok := signature.(*crypto.MultiKeySignature)
	if !ok {
		return nil, errors.New("invalid Multikey signature")
	}

	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiKey,
		Auth: &crypto.MultiKeyAuthenticator{
			PubKey: pubkey,
			Sig:    sig,
		},
	}, nil
}

func (s *MultiKeyTestSigner) SignMessage(msg []byte) (crypto.Signature, error) {
	indexedSigs := make([]crypto.IndexedAnySignature, s.SignaturesRequired)

	alreadyUsed := make(map[int]bool)

	for i := range s.SignaturesRequired {
		// Find a random key
		var index int
		for {
			// Note, this is just for testing, no reason to bring a crypto randomness in here
			//nolint:gosec
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
		u8Index, err := util.IntToU8(index)
		if err != nil {
			return nil, err
		}

		typedSig, ok := sig.(*crypto.AnySignature)
		if !ok {
			return nil, errors.New("invalid AnySignature")
		}
		indexedSigs[i] = crypto.IndexedAnySignature{Signature: typedSig, Index: u8Index}
	}

	return crypto.NewMultiKeySignature(indexedSigs)
}

func (s *MultiKeyTestSigner) SimulationAuthenticator() *crypto.AccountAuthenticator {
	pubkey, ok := s.PubKey().(*crypto.MultiKey)
	if !ok {
		return nil
	}
	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiKey,
		Auth: &crypto.MultiKeyAuthenticator{
			PubKey: pubkey,
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
