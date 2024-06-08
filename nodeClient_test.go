package aptos

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPollForTransaction(t *testing.T) {
	// this doesn't need to actually have an aptos-node!
	// API error on every GET is fine, poll for a few milliseconds then return error
	client, err := NewClient(LocalnetConfig)
	assert.NoError(t, err)

	start := time.Now()
	err = client.PollForTransactions([]string{"alice", "bob"}, PollTimeout(10*time.Millisecond), PollPeriod(2*time.Millisecond))
	dt := time.Now().Sub(start)

	assert.GreaterOrEqual(t, dt, 9*time.Millisecond)
	assert.Less(t, dt, 20*time.Millisecond)
	assert.Error(t, err)
}
