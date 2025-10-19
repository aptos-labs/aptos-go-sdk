package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// CexTx: Flat transaction structure aligned with on-chain CexTx
type CexTx struct {
	Subaccount      uint32
	Nonce           uint64
	ClobPair        ClobPair
	Side            Side
	Quantums        uint64
	Subticks        uint64
	OrderBasicType  uint32 //  1: limit order, 2: market order
	GoodTill        GoodTill
	TimeInForce     TimeInForce
	ReduceOnly      bool
	ConditionType   ConditionType
	TriggerSubticks uint64
	Operation       Operation
	TargetNonce     uint64
	CateType        OrderCateType
}

// region TransactionPayloadImpl
func (tx *CexTx) PayloadType() TransactionPayloadVariant {
	return TransactionPayloadVariantCEX
}

// region bcs.Struct
func (tx *CexTx) MarshalBCS(ser *bcs.Serializer) {
	ser.U32(tx.Subaccount)
	ser.U64(tx.Nonce)
	ser.U8(uint8(tx.ClobPair))
	ser.U8(uint8(tx.Side))
	ser.U64(tx.Quantums)
	ser.U64(tx.Subticks)
	ser.U32(tx.OrderBasicType)
	ser.U8(uint8(tx.GoodTill))
	ser.U8(uint8(tx.TimeInForce))
	ser.Bool(tx.ReduceOnly)
	ser.U8(uint8(tx.ConditionType))
	ser.U64(tx.TriggerSubticks)
	ser.U8(uint8(tx.Operation))
	ser.U64(tx.TargetNonce)
	ser.U8(uint8(tx.CateType))
}

func (tx *CexTx) UnmarshalBCS(des *bcs.Deserializer) {
	tx.Subaccount = des.U32()
	tx.Nonce = des.U64()
	tx.ClobPair = ClobPair(des.U8())
	tx.Side = Side(des.U8())
	tx.Quantums = des.U64()
	tx.Subticks = des.U64()
	tx.OrderBasicType = des.U32()
	tx.GoodTill = GoodTill(des.U8())
	tx.TimeInForce = TimeInForce(des.U8())
	tx.ReduceOnly = des.Bool()
	tx.ConditionType = ConditionType(des.U8())
	tx.TriggerSubticks = des.U64()
	tx.Operation = Operation(des.U8())
	tx.TargetNonce = des.U64()
	tx.CateType = OrderCateType(des.U8())
}
