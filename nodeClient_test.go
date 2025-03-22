package aptos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollForTransaction(t *testing.T) {
	t.Parallel()
	// this doesn't need to actually have an aptos-node!
	// API error on every GET is fine, poll for a few milliseconds then return error
	client, err := NewClient(LocalnetConfig)
	require.NoError(t, err)

	start := time.Now()
	err = client.PollForTransactions([]string{"alice", "bob"}, PollTimeout(10*time.Millisecond), PollPeriod(2*time.Millisecond))
	dt := time.Since(start)

	assert.GreaterOrEqual(t, dt, 9*time.Millisecond)
	assert.Less(t, dt, 20*time.Millisecond)
	require.Error(t, err)
}

// TODO: This test has to be rewritten, as it is not parallelizable between subtests
//
//nolint:golint,tparallel
func TestEventsByHandle(t *testing.T) {
	t.Parallel()
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// handle initial request from client
			w.WriteHeader(http.StatusOK)
			return
		}

		assert.Equal(t, "/accounts/0x0/events/0x2/transfer", r.URL.Path)

		start := r.URL.Query().Get("start")
		limit := r.URL.Query().Get("limit")

		startInt, _ := strconv.ParseUint(start, 10, 64)
		limitInt, _ := strconv.ParseUint(limit, 10, 64)

		events := make([]map[string]interface{}, 0, limitInt)
		for i := range limitInt {
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

		err := json.NewEncoder(w).Encode(events)
		if err != nil {
			return
		}
	}))
	defer mockServer.Close()

	client, err := NewClient(NetworkConfig{
		Name:    "mocknet",
		NodeUrl: mockServer.URL,
	})
	require.NoError(t, err)

	//nolint:golint,paralleltest
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

		require.NoError(t, err)
		assert.Len(t, events, 150)
	})

	//nolint:golint,paralleltest
	t.Run("default page size when limit not provided", func(t *testing.T) {
		events, err := client.EventsByHandle(
			AccountZero,
			"0x2",
			"transfer",
			nil,
			nil,
		)

		require.NoError(t, err)
		assert.Len(t, events, 100)
		assert.Equal(t, uint64(99), events[99].SequenceNumber)
	})

	//nolint:golint,paralleltest
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

		require.NoError(t, err)
		assert.Len(t, events, 5)
		assert.Equal(t, uint64(50), events[0].SequenceNumber)
		assert.Equal(t, uint64(54), events[4].SequenceNumber)
	})
}
