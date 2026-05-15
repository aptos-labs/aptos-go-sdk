package ca

import (
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
	"github.com/gtank/ristretto255"
)

// EncryptTwistedElGamal returns C,D compressed bytes for one chunk (TS TwistedElGamal.encryptWithPK).
func EncryptTwistedElGamal(amountChunk uint64, recipientPub32 []byte, random *big.Int) (c32, d32 []byte, err error) {
	var r ristretto255.Scalar
	if err := r.Decode(caed25519.NumberToBytesLE32(random)); err != nil {
		return nil, nil, err
	}
	var rH, mG, D, C ristretto255.Element
	rH.ScalarMult(&r, HRistretto)
	if amountChunk == 0 {
		mG.Zero()
	} else {
		var m ristretto255.Scalar
		if err := m.Decode(caed25519.NumberToBytesLE32(new(big.Int).SetUint64(amountChunk))); err != nil {
			return nil, nil, err
		}
		mG.ScalarBaseMult(&m)
	}
	var ek ristretto255.Element
	if err := ek.Decode(recipientPub32); err != nil {
		return nil, nil, err
	}
	D.ScalarMult(&r, &ek)
	C.Add(&mG, &rH)
	return C.Encode(make([]byte, 0, 32)), D.Encode(make([]byte, 0, 32)), nil
}
