package aptos

// EstimateGasInfo is returned by #EstimateGasPrice()
type EstimateGasInfo struct {
	DeprioritizedGasEstimate uint64 `json:"deprioritized_gas_estimate"`
	GasEstimate              uint64 `json:"gas_estimate"`
	PrioritizedGasEstimate   uint64 `json:"prioritized_gas_estimate"`
}
