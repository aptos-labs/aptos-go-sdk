package client

import (
	"log/slog"
	"strconv"
)

// NodeInfo information retrieved about the current state of the blockchain on API requests
type NodeInfo struct {
	ChainId                uint8  `json:"chain_id"`              // ChainId is the chain ID of the network
	EpochStr               string `json:"epoch"`                 // EpochStr is the current epoch of the network.  On Mainnet, this is usually every 2 hours.
	LedgerTimestampStr     string `json:"ledger_timestamp"`      // LedgerTimestampStr is the timestamp the block was committed
	LedgerVersionStr       string `json:"ledger_version"`        // LedgerVersionStr is the newest transaction available on the full node
	OldestLedgerVersionStr string `json:"oldest_ledger_version"` // OldestLedgerVersionStr is the oldest ledger version not pruned on the full node
	NodeRole               string `json:"node_role"`             // NodeRole is the role of the node in the network
	BlockHeightStr         string `json:"block_height"`          // BlockHeightStr is the newest block available on the full node (by the time you call this there's already a new one!)
	OldestBlockHeightStr   string `json:"oldest_block_height"`   // OldestBlockHeightStr is the oldest block note pruned on the full node
	GitHash                string `json:"git_hash"`              // GitHash is the git hash of the node
}

// Epoch the current epoch of the network.  On Mainnet, this is usually every 2 hours.
func (info NodeInfo) Epoch() uint64 {
	value, err := strconv.ParseUint(info.EpochStr, 10, 64)
	if err != nil {
		slog.Error("bad epoch", "v", info.EpochStr, "err", err)
		return 0
	}
	return value
}

// LedgerTimestamp is the timestamp the block was committed
func (info NodeInfo) LedgerTimestamp() uint64 {
	value, err := strconv.ParseUint(info.LedgerTimestampStr, 10, 64)
	if err != nil {
		slog.Error("bad ledger_timestamp", "v", info.LedgerTimestampStr, "err", err)
		return 0
	}
	return value
}

// LedgerVersion the newest transaction available on the full node
func (info NodeInfo) LedgerVersion() uint64 {
	value, err := strconv.ParseUint(info.LedgerVersionStr, 10, 64)
	if err != nil {
		slog.Error("bad ledger_version", "v", info.LedgerVersionStr, "err", err)
		return 0
	}
	return value
}

// OldestLedgerVersion the oldest ledger version not pruned on the full node
func (info NodeInfo) OldestLedgerVersion() uint64 {
	value, err := strconv.ParseUint(info.OldestLedgerVersionStr, 10, 64)
	if err != nil {
		slog.Error("bad oldest_ledger_version", "v", info.OldestLedgerVersionStr, "err", err)
		return 0
	}
	return value
}

// BlockHeight the newest block available on the full node (by the time you call this there's already a new one!)
func (info NodeInfo) BlockHeight() uint64 {
	value, err := strconv.ParseUint(info.BlockHeightStr, 10, 64)
	if err != nil {
		slog.Error("bad block_height", "v", info.BlockHeightStr, "err", err)
		return 0
	}
	return value
}

// OldestBlockHeight the oldest block note pruned on the full node
func (info NodeInfo) OldestBlockHeight() uint64 {
	value, err := strconv.ParseUint(info.OldestBlockHeightStr, 10, 64)
	if err != nil {
		slog.Error("bad oldest_block_height", "v", info.OldestBlockHeightStr, "err", err)
		return 0
	}
	return value
}
