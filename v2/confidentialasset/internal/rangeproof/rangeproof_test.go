//go:build cgo

package rangeproof_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/rangeproof"
)

func TestBatchRangeProof_roundTrip(t *testing.T) {
	valBase, _ := hex.DecodeString("e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76")
	randBase, _ := hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134")
	bl := make([]byte, 32)
	if _, err := rand.Read(bl); err != nil {
		t.Fatal(err)
	}
	proof, comms, err := rangeproof.BatchRangeProof([]uint64{42}, bl, valBase, randBase, 32)
	if err != nil {
		t.Fatal(err)
	}
	if len(proof) == 0 || len(comms) == 0 {
		t.Fatal("empty output")
	}
}
