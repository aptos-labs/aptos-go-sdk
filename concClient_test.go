package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConcClient_Info(t *testing.T) {
	inner, _ := createTestClient()
	client, _ := NewConcClient(inner)

	result := make(chan ConcResponse[*NodeInfo])
	go client.Info(result)
	resp := <-result
	assert.NoError(t, resp.Err)
}

func TestBlock(t *testing.T) {
	inner, _ := createTestClient()
	client, _ := NewConcClient(inner)

	result := make(chan ConcResponse[*api.Block])
	go client.BlockByHeight(result, 1, true)
	resp := <-result
	assert.NoError(t, resp.Err)

	result2 := make(chan ConcResponse[*api.Block])
	go client.BlockByVersion(result2, 1, true)
	resp2 := <-result2
	assert.NoError(t, resp.Err)

	assert.Equal(t, resp, resp2)
}

func TestConcClient_Transactions(t *testing.T) {
	inner, _ := createTestClient()
	client, _ := NewConcClient(inner)

	infoResult := make(chan ConcResponse[*NodeInfo])
	go client.Info(infoResult)
	infoResp := <-infoResult
	assert.NoError(t, infoResp.Err)

	total := make(chan ConcResponse[[]*api.Transaction], 1)
	totalLimit := min(1000, infoResp.Result.LedgerVersion())
	start := uint64(1)
	limit := totalLimit
	client.Transactions(total, start, limit)
	resp := <-total
	assert.NoError(t, resp.Err)

	assert.Equal(t, totalLimit, uint64(len(resp.Result)))
}

func TestConcClient_SubmitTransaction(t *testing.T) {
	inner, _ := createTestClient()
	client, _ := NewConcClient(inner)

	account1, err := NewEd25519Account()
	assert.NoError(t, err)
	account2, err := NewEd25519Account()
	assert.NoError(t, err)

	const fundAmount = uint64(100_000_000)

	fundChannel1 := make(chan ConcResponse[bool])
	go client.Fund(fundChannel1, account1.AccountAddress(), fundAmount)
	fundChannel2 := make(chan ConcResponse[bool])
	go client.Fund(fundChannel2, account2.AccountAddress(), fundAmount)

	fundResponse1 := <-fundChannel1
	fundResponse2 := <-fundChannel2
	assert.NoError(t, fundResponse1.Err)
	assert.NoError(t, fundResponse2.Err)

	transferAmount, err := bcs.SerializeU64(100)
	assert.NoError(t, err)

	transferChannel1 := make(chan ConcResponse[*api.SubmitTransactionResponse])
	transferChannel2 := make(chan ConcResponse[*api.SubmitTransactionResponse])

	go client.BuildSignAndSubmitTransaction(transferChannel1, account1, TransactionPayload{
		Payload: &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args:     [][]byte{account2.Address[:], transferAmount},
		},
	})
	go client.BuildSignAndSubmitTransaction(transferChannel2, account2, TransactionPayload{
		Payload: &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args:     [][]byte{account1.Address[:], transferAmount},
		},
	})

	transferResponse1 := <-transferChannel1
	assert.NoError(t, transferResponse1.Err)
	transferResponse2 := <-transferChannel2
	assert.NoError(t, transferResponse2.Err)

	waitChannel1 := make(chan ConcResponse[*api.UserTransaction])
	waitChannel2 := make(chan ConcResponse[*api.UserTransaction])
	go client.WaitForTransaction(waitChannel1, transferResponse1.Result.Hash)
	go client.WaitForTransaction(waitChannel2, transferResponse1.Result.Hash)

	waitResponse1 := <-waitChannel1
	assert.NoError(t, waitResponse1.Err)
	waitResponse2 := <-waitChannel2
	assert.NoError(t, waitResponse2.Err)
	assert.True(t, waitResponse1.Result.Success)
	assert.True(t, waitResponse2.Result.Success)
}
