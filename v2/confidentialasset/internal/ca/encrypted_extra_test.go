package ca

import "testing"

func TestPublicKeyBytes(t *testing.T) {
	t.Parallel()
	ek, err := TwistedPublicKeyFromPrivateLE32([32]byte{2})
	if err != nil {
		t.Fatal(err)
	}
	var pub [32]byte
	copy(pub[:], ek)
	enc, err := NewEncryptedAmountFromAmount(0, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	if enc.PublicKeyBytes() != pub {
		t.Fatal("pub mismatch")
	}
}
