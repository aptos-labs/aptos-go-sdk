package types

// EstimateGasInfo is returned by #EstimateGasPrice()
type EstimateGasInfo struct {
	DeprioritizedGasEstimate uint64 `json:"deprioritized_gas_estimate"` // DeprioritizedGasEstimate is the gas estimate for a transaction that is willing to be deprioritized and pay less
	GasEstimate              uint64 `json:"gas_estimate"`               // GasEstimate is the gas estimate for a transaction that is willing to pay close to the median gas price
	PrioritizedGasEstimate   uint64 `json:"prioritized_gas_estimate"`   // PrioritizedGasEstimate is the gas estimate for a transaction that is willing to pay more to be prioritized
}
