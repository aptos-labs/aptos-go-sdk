// Package testutil provides test utilities for the Aptos Go SDK.
//
// This package includes:
//   - FakeClient: A mock implementation of the Client interface for testing
//   - Test helpers for creating test data
//   - Fixtures for common test scenarios
//
// # FakeClient Usage
//
// The FakeClient allows you to test code that depends on the Aptos client
// without making actual network requests:
//
//	func TestMyFunction(t *testing.T) {
//		// Create a fake client with predefined responses
//		client := testutil.NewFakeClient().
//			WithAccount(addr, &aptos.AccountInfo{SequenceNumber: 5}).
//			WithBalance(addr, 100_000_000)
//
//		// Use the fake client in your code under test
//		result, err := myFunction(client, addr)
//		require.NoError(t, err)
//	}
//
// # Request Recording
//
// FakeClient can record all requests for verification:
//
//	client := testutil.NewFakeClient().WithRecording()
//
//	// ... use client ...
//
//	calls := client.RecordedCalls()
//	require.Len(t, calls, 2)
//	assert.Equal(t, "Account", calls[0].Method)
//
// # Error Simulation
//
// Simulate errors for specific operations:
//
//	client := testutil.NewFakeClient().
//		WithError("Account", aptos.ErrNotFound)
//
//	_, err := client.Account(ctx, addr)
//	assert.ErrorIs(t, err, aptos.ErrNotFound)
//
// # Test Helpers
//
// Helper functions for creating test data:
//
//	addr := testutil.RandomAddress()
//	signer := testutil.RandomSigner()
//	txn := testutil.SampleTransaction()
package testutil
