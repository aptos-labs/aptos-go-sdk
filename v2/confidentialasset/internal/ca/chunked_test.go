package ca

import "testing"

func TestAmountToChunks_ChunksToAmount_roundTrip(t *testing.T) {
	t.Parallel()
	amount := uint64(0x1234567890abcdef)
	chunks := AmountToChunks(amount, AvailableBalanceChunkCount)
	got := ChunksToAmount(chunks)
	if got != amount {
		t.Fatalf("got %#x want %#x", got, amount)
	}
}

func TestAmountToChunks_zero(t *testing.T) {
	t.Parallel()
	chunks := AmountToChunks(0, 4)
	for i, c := range chunks {
		if c != 0 {
			t.Fatalf("chunk %d=%d", i, c)
		}
	}
}
