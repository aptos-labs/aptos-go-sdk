//go:build cgo

package native

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

func TestWrap(t *testing.T) {
	if Wrap(nil) != nil {
		t.Fatal("nil wrap")
	}
	cc := confidentialasset.NewClient(testutil.NewFakeClient())
	if Wrap(cc) == nil {
		t.Fatal("expected client")
	}
}
