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

func (s *Script) MarshalBCS(serializer *bcs.Serializer) {
	serializer.WriteBytes(s.Code)
	bcs.SerializeSequence(s.ArgTypes, serializer)
	bcs.SerializeSequence(s.Args, serializer)
}

func (s *Script) UnmarshalBCS(deserializer *bcs.Deserializer) {
	s.Code = deserializer.ReadBytes()
	s.ArgTypes = bcs.DeserializeSequence[TypeTag](deserializer)
	s.Args = bcs.DeserializeSequence[ScriptArgument](deserializer)
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

func (sa *ScriptArgument) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(sa.Variant))
	switch sa.Variant {
	case ScriptArgumentU8:
		bcs.U8(sa.Value.(uint8))
	case ScriptArgumentU16:
		bcs.U16(sa.Value.(uint16))
	case ScriptArgumentU32:
		bcs.U32(sa.Value.(uint32))
	case ScriptArgumentU64:
		bcs.U64(sa.Value.(uint64))
	case ScriptArgumentU128:
		bcs.U128(sa.Value.(big.Int))
	case ScriptArgumentU256:
		bcs.U256(sa.Value.(big.Int))
	case ScriptArgumentAddress:
		addr := sa.Value.(AccountAddress)
		bcs.Struct(&addr)
	case ScriptArgumentU8Vector:
		bcs.WriteBytes(sa.Value.([]byte))
	case ScriptArgumentBool:
		bcs.Bool(sa.Value.(bool))
	}
}

func (sa *ScriptArgument) UnmarshalBCS(bcs *bcs.Deserializer) {
	sa.Variant = ScriptArgumentVariant(bcs.Uleb128())
	switch sa.Variant {
	case ScriptArgumentU8:
		sa.Value = bcs.U8()
	case ScriptArgumentU16:
		sa.Value = bcs.U16()
	case ScriptArgumentU32:
		sa.Value = bcs.U32()
	case ScriptArgumentU64:
		sa.Value = bcs.U64()
	case ScriptArgumentU128:
		sa.Value = bcs.U128()
	case ScriptArgumentU256:
		sa.Value = bcs.U256()
	case ScriptArgumentAddress:
		aa := AccountAddress{}
		aa.UnmarshalBCS(bcs)
		sa.Value = aa
	case ScriptArgumentU8Vector:
		sa.Value = bcs.ReadBytes()
	case ScriptArgumentBool:
		sa.Value = bcs.Bool()
	}
}

//endregion
//endregion
