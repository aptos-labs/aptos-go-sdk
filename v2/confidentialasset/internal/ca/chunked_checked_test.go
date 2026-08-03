package ca

import "testing"

func TestChunksToAmountChecked(t *testing.T) {
	t.Parallel()
	// Normalized amount (every chunk < 2^16) round-trips like ChunksToAmount.
	amount := uint64(123456789)
	got, err := ChunksToAmountChecked(AmountToChunks(amount, AvailableBalanceChunkCount))
	if err != nil || got != amount {
		t.Fatalf("normalized round-trip: got %d err %v, want %d", got, err, amount)
	}

	// Un-normalized pending balance: a low chunk exceeds 2^16 from homomorphic accumulation while
	// higher chunks stay zero. This must reassemble correctly (the old 16-bit check rejected it).
	acc := uint64(6_000_000) // ~2^22.5
	got, err = ChunksToAmountChecked([]uint64{acc, 0, 0, 0, 0, 0, 0, 0})
	if err != nil || got != acc {
		t.Fatalf("pending low chunk: got %d err %v, want %d", got, err, acc)
	}

	// A wide value at a higher chunk index still reassembles as long as the total fits uint64.
	got, err = ChunksToAmountChecked([]uint64{0, 1 << 20, 0, 0, 0, 0, 0, 0})
	if err != nil {
		t.Fatalf("high chunk err: %v", err)
	}
	if want := uint64(1) << 36; got != want {
		t.Fatalf("high chunk: got %d want %d", got, want)
	}

	// Overflow past uint64 must error rather than silently wrap (chunk 4 => value 2^64).
	if _, err := ChunksToAmountChecked([]uint64{0, 0, 0, 0, 1, 0, 0, 0}); err == nil {
		t.Fatal("expected overflow error for chunk4=1 (value 2^64)")
	}
}
