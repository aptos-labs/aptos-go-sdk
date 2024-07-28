package types

// AccountResourceInfo is returned by #AccountResource() and #AccountResources()
type AccountResourceInfo struct {
	// e.g. "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>"
	Type string `json:"type"`

	// Decoded from Move contract data, could really be anything
	Data map[string]any `json:"data"`
}
