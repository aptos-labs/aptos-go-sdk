package confidentialasset

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

func TestNewClient_options(t *testing.T) {
	t.Parallel()
	fc := testutil.NewFakeClient()
	mod := aptos.MustParseAddress("0x2")
	cc := NewClient(
		fc,
		WithModuleAddress(mod),
		WithFeePayer(true),
		WithRESTBaseURL("https://example.com/v1/"),
	)
	if cc.ModuleAddress != mod {
		t.Fatalf("module=%s", cc.ModuleAddress)
	}
	if !cc.WithFeePayer {
		t.Fatal("fee payer")
	}
	if cc.RESTBaseURL != "https://example.com/v1" {
		t.Fatalf("rest=%q", cc.RESTBaseURL)
	}
	mid := cc.ViewModule()
	if mid.Address != mod || mid.Name != ModuleName {
		t.Fatalf("module id %+v", mid)
	}
}
