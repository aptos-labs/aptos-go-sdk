package ca

import (
	"fmt"
	"math/big"
)

// Chunk layout matches confidential-asset TS chunkedAmount.ts.
const (
	AvailableBalanceChunkCount = 8
	TransferAmountChunkCount   = 4
	ChunkBits                  = 16
)

// MaxChunkValue bounds a single decrypted chunk value. A normalized balance keeps every chunk
// below 2^ChunkBits, but an un-normalized pending balance accumulates deposits homomorphically,
// so a chunk may legitimately exceed 16 bits (recovered via a wider discrete-log solve) while
// still fitting the solver's 32-bit ceiling.
const MaxChunkValue = uint64(1) << 32

// AmountToChunks splits amount into count chunks of ChunkBits bits each (TS ChunkedAmount.amountToChunks).
func AmountToChunks(amount uint64, chunksCount int) []uint64 {
	mask := uint64((1 << ChunkBits) - 1)
	out := make([]uint64, chunksCount)
	for i := 0; i < chunksCount; i++ {
		out[i] = (amount >> (ChunkBits * i)) & mask
	}
	return out
}

// ChunksToAmount combines chunks (TS ChunkedAmount.chunksToAmount).
func ChunksToAmount(chunks []uint64) uint64 {
	var total uint64
	for i, c := range chunks {
		total += c << (ChunkBits * i)
	}
	return total
}

// ChunksToAmountChecked combines chunks like ChunksToAmount but accumulates in big.Int and returns
// an error if the reassembled amount overflows uint64, instead of silently wrapping. Use this on the
// decrypt path where chunk values are not guaranteed to be normalized (e.g. pending balances).
func ChunksToAmountChecked(chunks []uint64) (uint64, error) {
	total := new(big.Int)
	for i, c := range chunks {
		term := new(big.Int).Lsh(new(big.Int).SetUint64(c), uint(ChunkBits*i))
		total.Add(total, term)
	}
	if !total.IsUint64() {
		return 0, fmt.Errorf("ca: reassembled amount %s exceeds uint64", total.String())
	}
	return total.Uint64(), nil
}
