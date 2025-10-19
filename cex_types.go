package aptos

import (
	"fmt"
)

// Constants for CEX functionality
const (
	AssetsSize    = 256
	PerpsSize     = 256
	OrderIDLength = 20
)

// ClobPair trading pair types
type ClobPair uint8

const (
	ClobPairUnspecified      ClobPair = 0
	ClobPairBtcUsdcSpot      ClobPair = 1
	ClobPairBtcUsdcPerpetual ClobPair = 2
	ClobPairEthUsdcSpot      ClobPair = 3
	ClobPairEthUsdcPerpetual ClobPair = 4
)

// AssetPosition asset position information
type AssetPosition struct {
	AssetID  uint32 `json:"asset_id"`
	Quantums uint64 `json:"quantums"`
}

// PerpetualPosition perpetual position information
type PerpetualPosition struct {
	PerpetualID   uint32 `json:"perpetual_id"`
	ShortQuantums int64  `json:"short_quantums"` // Note: using int64 instead of i128 in Go
	LongQuantums  int64  `json:"long_quantums"`  // Note: using int64 instead of i128 in Go
	FundingIndex  int64  `json:"funding_index"`  // Note: using int64 instead of i128 in Go
}

// ConditionType order condition types
type ConditionType uint8

const (
	ConditionTypeUnspecified ConditionType = 0
	ConditionTypeStopLoss    ConditionType = 1
	ConditionTypeTakeProfit  ConditionType = 2
)

// GoodTill validity period types
type GoodTill uint8

const (
	GoodTillBlock GoodTill = 0
	GoodTillGtc   GoodTill = 1
	GoodTillGtd   GoodTill = 2
)

// TimeInForce time validity types
type TimeInForce uint8

const (
	TimeInForceUnspecified TimeInForce = 0
	TimeInForceIoc         TimeInForce = 1 // Immediate or Cancel
	TimeInForceFok         TimeInForce = 2 // Fill or Kill
	TimeInForceAon         TimeInForce = 3 // All or None
	TimeInForceAlo         TimeInForce = 4 // Add Liquidity Only
)

// OrderState order status types
type OrderState uint8

const (
	OrderStateUnspecified     OrderState = 0
	OrderStatePending         OrderState = 1
	OrderStateValidated       OrderState = 2
	OrderStateActive          OrderState = 3
	OrderStatePartiallyFilled OrderState = 4
	OrderStateFilled          OrderState = 5
	OrderStateCancelled       OrderState = 6
	OrderStateRejected        OrderState = 7
)

// Side buy/sell direction
type Side uint8

const (
	SideUnspecified Side = 0
	SideBuy         Side = 1
	SideSell        Side = 2
)

// Operation operation types
type Operation uint8

const (
	OperationUnspecified Operation = 0
	OperationPlace       Operation = 1
	OperationCancel      Operation = 2
	OperationReplace     Operation = 3 // Currently not supported
)

// OrderCateType order category types
type OrderCateType uint8

const (
	OrderCateTypeRegular     OrderCateType = 0
	OrderCateTypeLiquidation OrderCateType = 1
	OrderCateTypeAdl         OrderCateType = 2
	OrderCateTypeFunding     OrderCateType = 3 // Currently not supported
)

// ValidateCexTx validates all enum values in the order
func ValidateCexTx(tx *CexTx) error {
	// Validate ClobPair (0-4)
	if tx.ClobPair > ClobPairEthUsdcPerpetual {
		return fmt.Errorf("invalid ClobPair: %d, must be 0-4", tx.ClobPair)
	}
	// Validate Side (0-2)
	if tx.Side > SideSell {
		return fmt.Errorf("invalid Side: %d, must be 0-2", tx.Side)
	}
	// Validate GoodTill (0-2)
	if tx.GoodTill > GoodTillGtd {
		return fmt.Errorf("invalid GoodTill: %d, must be 0-2", tx.GoodTill)
	}
	// Validate TimeInForce (0-4)
	if tx.TimeInForce > TimeInForceAlo {
		return fmt.Errorf("invalid TimeInForce: %d, must be 0-4", tx.TimeInForce)
	}
	// Validate ConditionType (0-2)
	if tx.ConditionType > ConditionTypeTakeProfit {
		return fmt.Errorf("invalid ConditionType: %d, must be 0-2", tx.ConditionType)
	}
	// Validate Operation (0-3)
	if tx.Operation > OperationReplace {
		return fmt.Errorf("invalid Operation: %d, must be 0-3", tx.Operation)
	}
	// Validate OrderCateType (0-3)
	if tx.CateType > OrderCateTypeFunding {
		return fmt.Errorf("invalid OrderCateType: %d, must be 0-3", tx.CateType)
	}
	return nil
}
