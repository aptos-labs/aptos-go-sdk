package types

import (
	"log/slog"
	"strconv"
)

type NodeInfo struct {
	ChainId                uint8  `json:"chain_id"`
	EpochStr               string `json:"epoch"`
	LedgerVersionStr       string `json:"ledger_version"`
	OldestLedgerVersionStr string `json:"oldest_ledger_version"`
	NodeRole               string `json:"node_role"`
	BlockHeightStr         string `json:"block_height"`
	OldestBlockHeightStr   string `json:"oldest_block_height"`
	GitHash                string `json:"git_hash"`
}

func (info NodeInfo) Epoch() uint64 {
	value, err := strconv.ParseUint(info.EpochStr, 10, 64)
	if err != nil {
		slog.Error("bad epoch", "v", info.EpochStr, "err", err)
		return 0
	}
	return value
}

func (info NodeInfo) LedgerVersion() uint64 {
	value, err := strconv.ParseUint(info.LedgerVersionStr, 10, 64)
	if err != nil {
		slog.Error("bad ledger_version", "v", info.LedgerVersionStr, "err", err)
		return 0
	}
	return value
}

func (info NodeInfo) OldestLedgerVersion() uint64 {
	value, err := strconv.ParseUint(info.OldestLedgerVersionStr, 10, 64)
	if err != nil {
		slog.Error("bad oldest_ledger_version", "v", info.OldestLedgerVersionStr, "err", err)
		return 0
	}
	return value
}

func (info NodeInfo) BlockHeight() uint64 {
	value, err := strconv.ParseUint(info.BlockHeightStr, 10, 64)
	if err != nil {
		slog.Error("bad block_height", "v", info.BlockHeightStr, "err", err)
		return 0
	}
	return value
}

func (info NodeInfo) OldestBlockHeight() uint64 {
	value, err := strconv.ParseUint(info.OldestBlockHeightStr, 10, 64)
	if err != nil {
		slog.Error("bad oldest_block_height", "v", info.OldestBlockHeightStr, "err", err)
		return 0
	}
	return value
}
