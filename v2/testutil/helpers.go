package testutil

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
)

// RandomAddress generates a random AccountAddress for testing.
func RandomAddress() aptos.AccountAddress {
	var addr aptos.AccountAddress
	if _, err := rand.Read(addr[:]); err != nil {
		panic(fmt.Sprintf("failed to generate random address: %v", err))
	}
	return addr
}

// RandomSigner generates a random Ed25519 signer for testing.
func RandomSigner() *crypto.Ed25519PrivateKey {
	key, err := crypto.GenerateEd25519PrivateKey()
	if err != nil {
		panic(fmt.Sprintf("failed to generate random signer: %v", err))
	}
	return key
}

// SampleAccountInfo creates a sample AccountInfo for testing.
func SampleAccountInfo(seqNum uint64) *aptos.AccountInfo {
	return &aptos.AccountInfo{
		SequenceNumber:    seqNum,
		AuthenticationKey: "0x" + RandomAddress().String()[2:],
	}
}

// SampleTransaction creates a sample successful transaction for testing.
func SampleTransaction() *aptos.Transaction {
	return &aptos.Transaction{
		Type:                    "user_transaction",
		Version:                 1000,
		Hash:                    fmt.Sprintf("0x%064x", time.Now().UnixNano()),
		StateChangeHash:         "0x0",
		EventRootHash:           "0x0",
		GasUsed:                 1000,
		Success:                 true,
		VMStatus:                "Executed successfully",
		AccumulatorRootHash:     "0x0",
		Timestamp:               uint64(time.Now().Unix()),
		Sender:                  RandomAddress().String(),
		SequenceNumber:          0,
		MaxGasAmount:            200000,
		GasUnitPrice:            100,
		ExpirationTimestampSecs: uint64(time.Now().Add(30 * time.Second).Unix()),
	}
}

// SampleFailedTransaction creates a sample failed transaction for testing.
func SampleFailedTransaction() *aptos.Transaction {
	txn := SampleTransaction()
	txn.Success = false
	txn.VMStatus = "Move abort in 0x1::coin: EINSUFFICIENT_BALANCE"
	return txn
}

// SampleBlock creates a sample block for testing.
func SampleBlock(height uint64, numTxns int) *aptos.Block {
	block := &aptos.Block{
		BlockHeight:    height,
		BlockHash:      fmt.Sprintf("0x%064x", height),
		BlockTimestamp: uint64(time.Now().Unix() * 1000000),
		FirstVersion:   height * 100,
		LastVersion:    height*100 + uint64(numTxns-1),
	}

	if numTxns > 0 {
		block.Transactions = make([]aptos.Transaction, numTxns)
		for i := 0; i < numTxns; i++ {
			txn := SampleTransaction()
			txn.Version = block.FirstVersion + uint64(i)
			block.Transactions[i] = *txn
		}
	}

	return block
}

// SampleNodeInfo creates sample node info for testing.
func SampleNodeInfo() *aptos.NodeInfo {
	return &aptos.NodeInfo{
		ChainID:             4,
		Epoch:               100,
		LedgerVersion:       50000,
		OldestLedgerVersion: 0,
		LedgerTimestamp:     uint64(time.Now().Unix() * 1000000),
		NodeRole:            "full_node",
		OldestBlockHeight:   0,
		BlockHeight:         500,
		GitHash:             "abc123",
	}
}

// SampleGasEstimate creates sample gas estimate for testing.
func SampleGasEstimate() *aptos.GasEstimate {
	return &aptos.GasEstimate{
		GasEstimate:              100,
		DeprioritizedGasEstimate: 50,
		PrioritizedGasEstimate:   150,
	}
}

// SampleResource creates a sample resource for testing.
func SampleResource(resourceType string, data map[string]any) aptos.Resource {
	if data == nil {
		data = make(map[string]any)
	}
	return aptos.Resource{
		Type: resourceType,
		Data: data,
	}
}

// SampleCoinStoreResource creates a coin store resource for testing.
func SampleCoinStoreResource(coinType string, balance uint64) aptos.Resource {
	return aptos.Resource{
		Type: fmt.Sprintf("0x1::coin::CoinStore<%s>", coinType),
		Data: map[string]any{
			"coin": map[string]any{
				"value": strconv.FormatUint(balance, 10),
			},
			"deposit_events": map[string]any{
				"counter": "0",
				"guid":    map[string]any{"id": map[string]any{"addr": "0x1", "creation_num": "0"}},
			},
			"withdraw_events": map[string]any{
				"counter": "0",
				"guid":    map[string]any{"id": map[string]any{"addr": "0x1", "creation_num": "1"}},
			},
			"frozen": false,
		},
	}
}

// SampleEvent creates a sample event for testing.
func SampleEvent(eventType string, data any) aptos.Event {
	return aptos.Event{
		GUID: aptos.EventGUID{
			CreationNumber: 0,
			AccountAddress: RandomAddress().String(),
		},
		SequenceNumber: 0,
		Type:           eventType,
		Data:           data,
	}
}

// WellKnownAddresses contains commonly used addresses for testing.
var WellKnownAddresses = struct {
	Zero  aptos.AccountAddress
	One   aptos.AccountAddress
	Two   aptos.AccountAddress
	Three aptos.AccountAddress
}{
	Zero:  aptos.AccountZero,
	One:   aptos.AccountOne,
	Two:   aptos.AccountTwo,
	Three: aptos.AccountThree,
}

// PredefinedSigners contains pre-generated signers for deterministic tests.
// These should only be used for testing purposes.
var PredefinedSigners = func() []*crypto.Ed25519PrivateKey {
	signers := make([]*crypto.Ed25519PrivateKey, 5)
	for i := range signers {
		seed := make([]byte, 32)
		for j := range seed {
			seed[j] = byte(i*32 + j)
		}
		key := &crypto.Ed25519PrivateKey{}
		if err := key.FromBytes(seed); err != nil {
			panic(err)
		}
		signers[i] = key
	}
	return signers
}()

// AssertEventually retries an assertion until it passes or times out.
// Useful for testing async operations.
func AssertEventually(fn func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}
