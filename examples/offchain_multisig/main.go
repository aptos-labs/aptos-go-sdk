// transfer_coin is an example of how to make a coin transfer transaction in the simplest possible way
package main

import (
	"errors"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

const (
	FundAmount     = 100_000_000
	TransferAmount = 1_000
)

type KeyStore interface {
	NumSigners() uint8
	Sign(num uint8, msg []byte) (*crypto.AccountAuthenticator, error)
	SignMessage(num uint8, msg []byte) (crypto.Signature, error)
	PublicKey(num uint8) (crypto.PublicKey, error)
}

// This is an example for a theoretical key store, that has two keys remotely elsewhere
type LocalKeyStore struct {
	Signers []crypto.Signer
}

func NewLocalKeyStore(numKeys uint8) (*LocalKeyStore, error) {
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

	return &LocalKeyStore{
		Signers: signers,
	}, nil
}

func (lk *LocalKeyStore) Sign(num uint8, msg []byte) (*crypto.AccountAuthenticator, error) {
	numSigners, err := lk.numSigners()
	if err != nil {
		return nil, err
	}
	if num > numSigners {
		return nil, errors.New("signer out of range")
	}
	return lk.Signers[num].Sign(msg)
}

func (lk *LocalKeyStore) SignMessage(num uint8, msg []byte) (crypto.Signature, error) {
	numSigners, err := lk.numSigners()
	if err != nil {
		return nil, err
	}
	if num > numSigners {
		return nil, errors.New("signer out of range")
	}
	return lk.Signers[num].SignMessage(msg)
}

func (lk *LocalKeyStore) numSigners() (uint8, error) {
	num, err := util.IntToU8(len(lk.Signers))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (lk *LocalKeyStore) NumSigners() uint8 {
	num, err := lk.numSigners()
	if err != nil {
		return 0
	}
	return num
}

func (lk *LocalKeyStore) PublicKey(num uint8) (crypto.PublicKey, error) {
	numSigners, err := lk.numSigners()
	if err != nil {
		return nil, err
	}
	if num > numSigners {
		return nil, errors.New("signer out of range")
	}
	return lk.Signers[num].PubKey(), nil
}

type MultiKeySigner struct {
	Keystore           KeyStore
	PublicKey          *crypto.MultiKey
	SignaturesRequired uint8
}

func NewMultiKeySigner(keystore KeyStore, signaturesRequired uint8) (*MultiKeySigner, error) {
	numKeys := keystore.NumSigners()
	pubKeys := make([]*crypto.AnyPublicKey, numKeys)
	for i := range numKeys {
		pubkey, err := keystore.PublicKey(i)
		if err != nil {
			return nil, err
		}
		pubkeyTyped, ok := pubkey.(*crypto.AnyPublicKey)
		if !ok {
			return nil, errors.New("public key is not of type AnyPublicKey")
		}
		pubKeys[i] = pubkeyTyped
	}

	totalPubkey := &crypto.MultiKey{
		PubKeys:            pubKeys,
		SignaturesRequired: signaturesRequired,
	}
	return &MultiKeySigner{
		Keystore:           keystore,
		PublicKey:          totalPubkey,
		SignaturesRequired: signaturesRequired,
	}, nil
}

func (s *MultiKeySigner) AccountAddress() aptos.AccountAddress {
	address := aptos.AccountAddress{}
	address.FromAuthKey(s.AuthKey())
	return address
}

func (s *MultiKeySigner) Sign(msg []byte) (*crypto.AccountAuthenticator, error) {
	signature, err := s.SignMessage(msg)
	if err != nil {
		return nil, err
	}

	pubkey, ok := s.PubKey().(*crypto.MultiKey)
	if !ok {
		return nil, errors.New("multi key is not of type MultiKey")
	}
	sig, ok := signature.(*crypto.MultiKeySignature)
	if !ok {
		return nil, errors.New("signature is not of type MultiKeySignature")
	}
	return &crypto.AccountAuthenticator{
		Variant: crypto.AccountAuthenticatorMultiKey,
		Auth: &crypto.MultiKeyAuthenticator{
			PubKey: pubkey,
			Sig:    sig,
		},
	}, nil
}

func (s *MultiKeySigner) SignMessage(msg []byte) (crypto.Signature, error) {
	indexedSigs := make([]crypto.IndexedAnySignature, s.SignaturesRequired)

	for i := range s.SignaturesRequired {
		sig, err := s.Keystore.SignMessage(i, msg)
		if err != nil {
			return nil, err
		}
		sigTyped, ok := sig.(*crypto.AnySignature)
		if !ok {
			return nil, errors.New("signature is not of type *AnySignature")
		}
		indexedSigs[i] = crypto.IndexedAnySignature{Signature: sigTyped, Index: i}
	}

	return crypto.NewMultiKeySignature(indexedSigs)
}

func (s *MultiKeySigner) SimulationAuthenticator() *crypto.AccountAuthenticator {
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

func (s *MultiKeySigner) AuthKey() *crypto.AuthenticationKey {
	return s.PubKey().AuthKey()
}

func (s *MultiKeySigner) PubKey() crypto.PublicKey {
	return s.PublicKey
}

// example This example shows you how to make an APT transfer transaction in the simplest possible way
func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create "remote" keys
	keyStore, err := NewLocalKeyStore(4)
	if err != nil {
		panic("Failed to create keys:" + err.Error())
	}

	// Create account info
	alice, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create alice:" + err.Error())
	}
	multikeySigner, err := NewMultiKeySigner(keyStore, 2)
	if err != nil {
		panic("Failed to create multi key signer:" + err.Error())
	}

	// Fund the sender with the faucet to create it on-chain
	err = client.Fund(alice.AccountAddress(), TransferAmount)
	if err != nil {
		panic("Failed to fund alice:" + err.Error())
	}
	err = client.Fund(multikeySigner.AccountAddress(), FundAmount)
	if err != nil {
		panic("Failed to fund multikey:" + err.Error())
	}

	aliceBalance, err := client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	multikeyBalance, err := client.AccountAPTBalance(multikeySigner.AccountAddress())
	if err != nil {
		panic("Failed to retrieve multikey balance:" + err.Error())
	}
	fmt.Printf("\n=== Initial Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Multikey:%d\n", multikeyBalance)

	// 1. Build transaction
	multikeyAddress := multikeySigner.AccountAddress()
	accountBytes, err := bcs.Serialize(&alice.Address)
	if err != nil {
		panic("Failed to serialize alice's address:" + err.Error())
	}

	amountBytes, err := bcs.SerializeU64(TransferAmount)
	if err != nil {
		panic("Failed to serialize transfer amount:" + err.Error())
	}
	rawTxn, err := client.BuildTransaction(multikeyAddress, aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: aptos.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{
				accountBytes,
				amountBytes,
			},
		},
	},
	)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// TODO: Add generic simulation support in SDK for new types
	/*
		// 2. Simulate transaction (optional)
		// This is useful for understanding how much the transaction will cost
		// and to ensure that the transaction is valid before sending it to the network
		// This is optional, but recommended
		simulationResult, err := client.SimulateTransaction(rawTxn, multikeySigner)
		if err != nil {
			panic("Failed to simulate transaction:" + err.Error())
		}
		fmt.Printf("\n=== Simulation ===\n")
		fmt.Printf("Gas unit price: %d\n", simulationResult[0].GasUnitPrice)
		fmt.Printf("Gas used: %d\n", simulationResult[0].GasUsed)
		fmt.Printf("Total gas fee: %d\n", simulationResult[0].GasUsed*simulationResult[0].GasUnitPrice)
		fmt.Printf("Status: %s\n", simulationResult[0].VmStatus)
	*/
	// 3. Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(multikeySigner)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// 4. Submit transaction
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// 5. Wait for the transaction to complete
	_, err = client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	// Check balances
	aliceBalance, err = client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	multikeyBalance, err = client.AccountAPTBalance(multikeyAddress)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Final Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", multikeyBalance)
}

func main() {
	example(aptos.DevnetConfig)
}
