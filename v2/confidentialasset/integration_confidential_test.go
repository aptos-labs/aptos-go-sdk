package confidentialasset

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aptos-labs/aptos-go-sdk/v2"
)

// L3 (optional): live view against Aptos networks. Not run in CI short mode or without env gate.
func TestIntegration_ConfidentialHasConfidentialStoreView(t *testing.T) {
	if testing.Short() {
		t.Skip("set APTOS_CONFIDENTIAL_INTEGRATION=1 and omit -short for live view test")
	}
	if os.Getenv("APTOS_CONFIDENTIAL_INTEGRATION") != "1" {
		t.Skip("set APTOS_CONFIDENTIAL_INTEGRATION=1 to run confidential view integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ac, err := aptos.NewClient(aptos.Testnet)
	require.NoError(t, err)
	cc := NewClient(ac)

	// FA metadata long for native APT (same as examples).
	token := aptos.MustParseAddress("0x000000000000000000000000000000000000000000000000000000000000000a")
	// Framework account is unlikely to have a confidential store; we only assert the view RPC succeeds.
	_, err = cc.HasUserRegistered(ctx, aptos.AccountOne, token)
	require.NoError(t, err)
}

// L2 placeholder gate: extend with devnet simulate of confidential payloads when keys/fixtures exist.
func TestGate_ConfidentialSimulate_envOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("omit -short for optional simulate exploration")
	}
	if os.Getenv("APTOS_CONFIDENTIAL_SIMULATE") != "1" {
		t.Skip("set APTOS_CONFIDENTIAL_SIMULATE=1 for future simulate-based checks")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	ac, err := aptos.NewClient(aptos.Devnet)
	require.NoError(t, err)
	_, err = ac.ChainID(ctx)
	require.NoError(t, err)
}
