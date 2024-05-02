package aptos

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

// TODO: write NodeInfo accessors to ParseUint on *Str which work around 53 bit float64 limit in JavaScript
