package testutil

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk"
)

// TestClients holds the clients needed for testing
type TestClients struct {
	NodeClient *aptos.NodeClient
	Client     *aptos.Client
}

func CreateTestClient() (*aptos.Client, error) {
	return aptos.NewClient(aptos.DevnetConfig)
}

func CreateTestNodeClient() (*aptos.NodeClient, error) {
	return aptos.NewNodeClient(aptos.DevnetConfig.NodeUrl, aptos.DevnetConfig.ChainId)
}

func SetupTestClients(t *testing.T) *TestClients {
	t.Helper()
	nodeClient, err := CreateTestNodeClient()
	if err != nil {
		t.Fatalf("Failed to create NodeClient: %v", err)
	}

	client, err := CreateTestClient()
	if err != nil {
		t.Fatalf("Failed to create Client: %v", err)
	}

	return &TestClients{
		NodeClient: nodeClient,
		Client:     client,
	}
}

func CreateTransferPayload(t *testing.T, receiver aptos.AccountAddress, amount uint64) aptos.TransactionPayload {
	t.Helper()
	p, err := aptos.CoinTransferPayload(nil, receiver, amount)
	if err != nil {
		t.Fatalf("Failed to create transfer payload: %v", err)
	}
	return aptos.TransactionPayload{Payload: p}
}

// TestAccount represents a funded account for testing
type TestAccount struct {
	Account        *aptos.Account
	InitialBalance uint64
}

func SetupTestAccount(t *testing.T, client *aptos.Client, funding uint64) TestAccount {
	t.Helper()
	account, err := aptos.NewEd25519Account()
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	err = client.Fund(account.Address, funding)
	if err != nil {
		t.Fatalf("Failed to fund account: %v", err)
	}

	balance, err := client.AccountAPTBalance(account.Address)
	if err != nil {
		t.Fatalf("Failed to get initial balance: %v", err)
	}

	return TestAccount{
		Account:        account,
		InitialBalance: balance,
	}
}
