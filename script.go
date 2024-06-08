package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"math/big"
)

//region Script

// Script A Move script as compiled code as a transaction
type Script struct {
	Code     []byte
	ArgTypes []TypeTag
	Args     []ScriptArgument
}

//region Script TransactionPayloadImpl

func (s *Script) PayloadType() TransactionPayloadVariant {
	return TransactionPayloadVariantScript
}

//endregion

//region Script bcs.Struct

func (s *Script) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(s.Code)
	bcs.SerializeSequence(s.ArgTypes, ser)
	bcs.SerializeSequence(s.Args, ser)
}

func (s *Script) UnmarshalBCS(des *bcs.Deserializer) {
	s.Code = des.ReadBytes()
	s.ArgTypes = bcs.DeserializeSequence[TypeTag](des)
	s.Args = bcs.DeserializeSequence[ScriptArgument](des)
}

//endregion
//endregion

//region ScriptArgument

type ScriptArgumentVariant uint32

const (
	ScriptArgumentU8       ScriptArgumentVariant = 0
	ScriptArgumentU64      ScriptArgumentVariant = 1
	ScriptArgumentU128     ScriptArgumentVariant = 2
	ScriptArgumentAddress  ScriptArgumentVariant = 3
	ScriptArgumentU8Vector ScriptArgumentVariant = 4
	ScriptArgumentBool     ScriptArgumentVariant = 5
	ScriptArgumentU16      ScriptArgumentVariant = 6
	ScriptArgumentU32      ScriptArgumentVariant = 7
	ScriptArgumentU256     ScriptArgumentVariant = 8
)

// ScriptArgument a Move script argument, which encodes its type with it
type ScriptArgument struct {
	Variant ScriptArgumentVariant
	Value   any // TODO: Do we add better typing, or stick with the any
}

//region ScriptArgument bcs.Struct

func (sa *ScriptArgument) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(sa.Variant))
	switch sa.Variant {
	case ScriptArgumentU8:
		ser.U8(sa.Value.(uint8))
	case ScriptArgumentU16:
		ser.U16(sa.Value.(uint16))
	case ScriptArgumentU32:
		ser.U32(sa.Value.(uint32))
	case ScriptArgumentU64:
		ser.U64(sa.Value.(uint64))
	case ScriptArgumentU128:
		ser.U128(sa.Value.(big.Int))
	case ScriptArgumentU256:
		ser.U256(sa.Value.(big.Int))
	case ScriptArgumentAddress:
		addr := sa.Value.(AccountAddress)
		ser.Struct(&addr)
	case ScriptArgumentU8Vector:
		ser.WriteBytes(sa.Value.([]byte))
	case ScriptArgumentBool:
		ser.Bool(sa.Value.(bool))
	}
}

func (sa *ScriptArgument) UnmarshalBCS(des *bcs.Deserializer) {
	sa.Variant = ScriptArgumentVariant(des.Uleb128())
	switch sa.Variant {
	case ScriptArgumentU8:
		sa.Value = des.U8()
	case ScriptArgumentU16:
		sa.Value = des.U16()
	case ScriptArgumentU32:
		sa.Value = des.U32()
	case ScriptArgumentU64:
		sa.Value = des.U64()
	case ScriptArgumentU128:
		sa.Value = des.U128()
	case ScriptArgumentU256:
		sa.Value = des.U256()
	case ScriptArgumentAddress:
		aa := AccountAddress{}
		aa.UnmarshalBCS(des)
		sa.Value = aa
	case ScriptArgumentU8Vector:
		sa.Value = des.ReadBytes()
	case ScriptArgumentBool:
		sa.Value = des.Bool()
	}
}

//endregion
//endregion
