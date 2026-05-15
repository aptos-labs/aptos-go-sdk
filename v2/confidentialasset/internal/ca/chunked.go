package ca

// Chunk layout matches confidential-asset TS chunkedAmount.ts.
const (
	AvailableBalanceChunkCount = 8
	TransferAmountChunkCount   = 4
	ChunkBits                  = 16
)

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
