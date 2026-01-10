// Package aptos faucet integration tests verify faucet functionality.
//
// These tests require network connectivity and fund accounts on testnet/devnet.
// They should be run sparingly and are skipped by default unless APTOS_TEST_FAUCET=1.
//
// Run with: APTOS_TEST_FAUCET=1 go test -v -run IntegrationFaucet ./v2/...
package aptos

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoFaucetTests skips unless APTOS_TEST_FAUCET=1 is set.
func skipIfNoFaucetTests(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping faucet integration test in short mode")
	}
	if os.Getenv("APTOS_TEST_FAUCET") != "1" {
		t.Skip("skipping faucet test (set APTOS_TEST_FAUCET=1 to enable)")
	}
}

// testNetworkWithFaucet returns a network that has a faucet (testnet or devnet).
func testNetworkWithFaucet() NetworkConfig {
	network := os.Getenv("APTOS_NETWORK")
	switch network {
	case "devnet":
		return Devnet
	default:
		return Testnet
	}
}

// generateRandomAddress generates a random-looking address for testing.
// In real usage, this would come from a private key, but for faucet tests
// we just need a unique address.
func generateRandomAddress(t *testing.T) AccountAddress {
	t.Helper()
	// Use test name hash to generate deterministic but unique address
	hash := []byte(t.Name())
	var addr AccountAddress
	for i := 0; i < 32 && i < len(hash); i++ {
		addr[i] = hash[i]
	}
	// Ensure it's not a special address
	addr[0] = 0xAB
	return addr
}

// TestIntegrationFaucet_Fund tests funding an account from the faucet.
func TestIntegrationFaucet_Fund(t *testing.T) {
	skipIfNoFaucetTests(t)

	client, err := NewClient(testNetworkWithFaucet())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Generate a fresh address for this test
	addr := generateRandomAddress(t)
	t.Logf("Testing faucet fund for address: %s", addr.String())

	// Fund the account
	amount := uint64(100_000_000) // 1 APT
	err = client.Fund(ctx, addr, amount)
	require.NoError(t, err, "faucet should fund the account")

	// Wait a bit for the transaction to be processed
	time.Sleep(2 * time.Second)

	// Verify the account now exists
	_, err = client.Account(ctx, addr)
	require.NoError(t, err, "account should exist after funding")

	// Verify the balance
	balance, err := client.AccountBalance(ctx, addr)
	require.NoError(t, err, "should be able to get balance")
	assert.GreaterOrEqual(t, balance, amount, "balance should be at least the funded amount")

	t.Logf("Successfully funded account %s with %d octas (balance: %d)", addr.String(), amount, balance)
}

// TestIntegrationFaucet_FundExistingAccount tests funding an account that already exists.
func TestIntegrationFaucet_FundExistingAccount(t *testing.T) {
	skipIfNoFaucetTests(t)

	client, err := NewClient(testNetworkWithFaucet())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Fund 0x1 (always exists)
	// This might not actually add funds (the faucet might reject well-funded accounts)
	// but it shouldn't error
	err = client.Fund(ctx, AccountOne, 100_000_000)
	// Some faucets may reject funding 0x1, which is fine
	if err != nil {
		t.Logf("Faucet rejected funding 0x1 (expected for some faucets): %v", err)
	} else {
		t.Log("Successfully funded 0x1")
	}
}

// TestIntegrationFaucet_MainnetShouldFail tests that faucet fails on mainnet.
func TestIntegrationFaucet_MainnetShouldFail(t *testing.T) {
	skipIfNoFaucetTests(t)

	client, err := NewClient(Mainnet)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Fund(ctx, AccountOne, 100_000_000)
	require.Error(t, err, "mainnet should not have a faucet")
	assert.Contains(t, err.Error(), "faucet not available", "error should mention faucet unavailability")
}

// TestIntegrationFaucet_MultipleFunds tests funding the same account multiple times.
func TestIntegrationFaucet_MultipleFunds(t *testing.T) {
	skipIfNoFaucetTests(t)

	client, err := NewClient(testNetworkWithFaucet())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	addr := generateRandomAddress(t)
	amount := uint64(100_000_000)

	// First fund
	err = client.Fund(ctx, addr, amount)
	require.NoError(t, err, "first fund should succeed")
	time.Sleep(2 * time.Second)

	balance1, err := client.AccountBalance(ctx, addr)
	require.NoError(t, err)

	// Second fund
	err = client.Fund(ctx, addr, amount)
	require.NoError(t, err, "second fund should succeed")
	time.Sleep(2 * time.Second)

	balance2, err := client.AccountBalance(ctx, addr)
	require.NoError(t, err)

	// Balance should have increased (approximately doubled, minus any fees)
	assert.Greater(t, balance2, balance1, "balance should increase after second fund")
	t.Logf("Balance increased from %d to %d after second fund", balance1, balance2)
}

// TestIntegrationFaucet_CancelledContext tests that a cancelled context stops the faucet request.
func TestIntegrationFaucet_CancelledContext(t *testing.T) {
	skipIfNoFaucetTests(t)

	client, err := NewClient(testNetworkWithFaucet())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.Fund(ctx, generateRandomAddress(t), 100_000_000)
	require.Error(t, err, "cancelled context should cause error")
}
