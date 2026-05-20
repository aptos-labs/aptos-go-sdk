package sigma

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/ca"
)

func TestBCSRegistrationSession_golden(t *testing.T) {
	t.Parallel()
	var sender, token [32]byte
	sender[0] = 1
	token[0] = 2
	got := BCSRegistrationSession(sender, token)
	if len(got) != 64 {
		t.Fatalf("len=%d", len(got))
	}
}

func TestProveRegistration(t *testing.T) {
	t.Parallel()
	var dk, sender, token [32]byte
	dk[0] = 5
	sender[0] = 1
	token[0] = 2
	proof, err := ProveRegistration(dk, sender, token, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(proof.Commitment) == 0 || len(proof.Response) == 0 {
		t.Fatal("empty proof")
	}
	ek, err := ca.TwistedPublicKeyFromPrivateLE32(dk)
	if err != nil {
		t.Fatal(err)
	}
	if len(ek) != 32 {
		t.Fatalf("ek len=%d", len(ek))
	}
}
