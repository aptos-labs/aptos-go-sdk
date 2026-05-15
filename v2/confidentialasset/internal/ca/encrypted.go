package ca

import (
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/gtank/ristretto255"
)

// CipherChunk is one Twisted ElGamal chunk (compressed C and D).
type CipherChunk struct {
	C, D [32]byte
}

// EncryptedAmount holds chunked balance + ciphertexts under one public key.
type EncryptedAmount struct {
	PublicKey32  [32]byte
	AmountChunks []uint64
	Amount       uint64
	Cipher       []CipherChunk
	Randomness   []*big.Int // per chunk, may be nil when unknown
}

// NewEncryptedAmountFromAmount encrypts full available balance (8 chunks) with optional randomness.
func NewEncryptedAmountFromAmount(amount uint64, pub32 [32]byte, randomness []*big.Int) (*EncryptedAmount, error) {
	chunks := AmountToChunks(amount, AvailableBalanceChunkCount)
	if randomness == nil {
		var err error
		randomness, err = caed25519.GenListOfRandom(AvailableBalanceChunkCount)
		if err != nil {
			return nil, err
		}
	}
	if len(randomness) != AvailableBalanceChunkCount {
		return nil, fmt.Errorf("ca: need %d random scalars", AvailableBalanceChunkCount)
	}
	cipher := make([]CipherChunk, AvailableBalanceChunkCount)
	for i := range chunks {
		c, d, err := EncryptTwistedElGamal(chunks[i], pub32[:], randomness[i])
		if err != nil {
			return nil, err
		}
		copy(cipher[i].C[:], c)
		copy(cipher[i].D[:], d)
	}
	return &EncryptedAmount{
		PublicKey32:  pub32,
		AmountChunks: chunks,
		Amount:       amount,
		Cipher:       cipher,
		Randomness:   randomness,
	}, nil
}

// NewEncryptedTransferAmount encrypts transfer amount (4 chunks).
func NewEncryptedTransferAmount(amount uint64, pub32 [32]byte, randomness []*big.Int) (*EncryptedAmount, error) {
	chunks := AmountToChunks(amount, TransferAmountChunkCount)
	if randomness == nil {
		var err error
		randomness, err = caed25519.GenListOfRandom(TransferAmountChunkCount)
		if err != nil {
			return nil, err
		}
	}
	if len(randomness) != TransferAmountChunkCount {
		return nil, fmt.Errorf("ca: transfer needs %d random scalars", TransferAmountChunkCount)
	}
	cipher := make([]CipherChunk, TransferAmountChunkCount)
	for i := range chunks {
		c, d, err := EncryptTwistedElGamal(chunks[i], pub32[:], randomness[i])
		if err != nil {
			return nil, err
		}
		copy(cipher[i].C[:], c)
		copy(cipher[i].D[:], d)
	}
	return &EncryptedAmount{
		PublicKey32:  pub32,
		AmountChunks: chunks,
		Amount:       amount,
		Cipher:       cipher,
		Randomness:   randomness,
	}, nil
}

// FromCipherChunks builds EncryptedAmount from on-chain C/D rows + decrypted chunk values.
func FromCipherChunks(pub32 [32]byte, chunks []uint64, c32, d32 [][]byte) (*EncryptedAmount, error) {
	if len(chunks) != len(c32) || len(c32) != len(d32) {
		return nil, fmt.Errorf("ca: chunk length mismatch")
	}
	cc := make([]CipherChunk, len(chunks))
	for i := range chunks {
		if len(c32[i]) != 32 || len(d32[i]) != 32 {
			return nil, fmt.Errorf("ca: chunk %d bad point len", i)
		}
		copy(cc[i].C[:], c32[i])
		copy(cc[i].D[:], d32[i])
	}
	return &EncryptedAmount{
		PublicKey32:  pub32,
		AmountChunks: append([]uint64(nil), chunks...),
		Amount:       ChunksToAmount(chunks),
		Cipher:       cc,
		Randomness:   nil,
	}, nil
}

// RowsCD returns compressed C and D rows (view order).
func (e *EncryptedAmount) RowsCD() (c, d [][]byte) {
	for i := range e.Cipher {
		c = append(c, append([]byte(nil), e.Cipher[i].C[:]...))
		d = append(d, append([]byte(nil), e.Cipher[i].D[:]...))
	}
	return c, d
}

// PublicKeyBytes returns the encryption public key bytes.
func (e *EncryptedAmount) PublicKeyBytes() [32]byte {
	return e.PublicKey32
}

func (e *EncryptedAmount) PointC(i int) (*ristretto255.Element, error) {
	var p ristretto255.Element
	if err := p.Decode(e.Cipher[i].C[:]); err != nil {
		return nil, err
	}
	return &p, nil
}

// PointD returns chunk i D as Ristretto element.
func (e *EncryptedAmount) PointD(i int) (*ristretto255.Element, error) {
	var p ristretto255.Element
	if err := p.Decode(e.Cipher[i].D[:]); err != nil {
		return nil, err
	}
	return &p, nil
}
