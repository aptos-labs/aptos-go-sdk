package aptos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPollForTransaction(t *testing.T) {
	// this doesn't need to actually have an aptos-node!
	// API error on every GET is fine, poll for a few milliseconds then return error
	client, err := NewClient(LocalnetConfig)
	assert.NoError(t, err)

	start := time.Now()
	err = client.PollForTransactions([]string{"alice", "bob"}, PollTimeout(10*time.Millisecond), PollPeriod(2*time.Millisecond))
	dt := time.Since(start)

	assert.GreaterOrEqual(t, dt, 9*time.Millisecond)
	assert.Less(t, dt, 20*time.Millisecond)
	assert.Error(t, err)
}

func TestEventsByHandle(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			// handle initial request from client
			writer.WriteHeader(http.StatusOK)
			return
		}

		assert.Equal(t, "/accounts/0x0/events/0x2/transfer", request.URL.Path)

		start := request.URL.Query().Get("start")
		limit := request.URL.Query().Get("limit")

		startInt, _ := strconv.ParseUint(start, 10, 64)
		limitInt, _ := strconv.ParseUint(limit, 10, 64)

		events := make([]map[string]interface{}, 0, limitInt)
		for i := uint64(0); i < limitInt; i++ {
			events = append(events, map[string]interface{}{
				"type": "0x1::coin::TransferEvent",
				"guid": map[string]interface{}{
					"creation_number": "1",
					"account_address": AccountZero.String(),
				},
				"sequence_number": strconv.FormatUint(startInt+i, 10),
				"data": map[string]interface{}{
					"amount": strconv.FormatUint((startInt+i)*100, 10),
				},
			})
		}

		err := json.NewEncoder(writer).Encode(events)
		assert.NoError(t, err)
	}))
	defer mockServer.Close()

	client, err := NewClient(NetworkConfig{
		Name:    "mocknet",
		NodeUrl: mockServer.URL,
	})
	assert.NoError(t, err)

	t.Run("pagination with concurrent fetching", func(t *testing.T) {
		start := uint64(0)
		limit := uint64(150)
		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			&start,
			&limit,
		)

		assert.NoError(t, err)
		assert.Len(t, events, 150)
	})

	t.Run("default page size when limit not provided", func(t *testing.T) {
		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			nil,
			nil,
		)

		assert.NoError(t, err)
		assert.Len(t, events, 100)
		assert.Equal(t, uint64(99), events[99].SequenceNumber)
	})

	t.Run("single page fetch", func(t *testing.T) {
		start := uint64(50)
		limit := uint64(5)
		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			&start,
			&limit,
		)

		jsonBytes, _ := json.MarshalIndent(events, "", "  ")
		t.Logf("JSON Response: %s", string(jsonBytes))

		assert.NoError(t, err)
		assert.Len(t, events, 5)
		assert.Equal(t, uint64(50), events[0].SequenceNumber)
		assert.Equal(t, uint64(54), events[4].SequenceNumber)
	})
}
