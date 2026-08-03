package ca

import (
	"math/big"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/caed25519"
)

func TestNewEncryptedAmountFromAmount(t *testing.T) {
	t.Parallel()
	var pub [32]byte
	copy(pub[:], mustPub(t))
	enc, err := NewEncryptedAmountFromAmount(1000, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	if enc.Amount != 1000 || len(enc.Cipher) != AvailableBalanceChunkCount {
		t.Fatalf("enc=%+v", enc)
	}
	c, d := enc.RowsCD()
	if len(c) != AvailableBalanceChunkCount || len(d) != AvailableBalanceChunkCount {
		t.Fatalf("rows=%d", len(c))
	}
	_, err = enc.PointC(0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = enc.PointD(0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewEncryptedTransferAmount(t *testing.T) {
	t.Parallel()
	var pub [32]byte
	copy(pub[:], mustPub(t))
	enc, err := NewEncryptedTransferAmount(500, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(enc.Cipher) != TransferAmountChunkCount {
		t.Fatalf("chunks=%d", len(enc.Cipher))
	}
}

func TestFromCipherChunks(t *testing.T) {
	t.Parallel()
	var pub [32]byte
	copy(pub[:], mustPub(t))
	r := caed25519.ModN(big.NewInt(3))
	c, d, err := EncryptTwistedElGamal(10, pub[:], r)
	if err != nil {
		t.Fatal(err)
	}
	chunks := []uint64{10}
	enc, err := FromCipherChunks(pub, chunks, [][]byte{c}, [][]byte{d})
	if err != nil {
		t.Fatal(err)
	}
	if enc.Amount != 10 {
		t.Fatalf("amount=%d", enc.Amount)
	}
}

func mustPub(t *testing.T) []byte {
	t.Helper()
	p, err := TwistedPublicKeyFromPrivateLE32([32]byte{2})
	if err != nil {
		t.Fatal(err)
	}
	return p
}
