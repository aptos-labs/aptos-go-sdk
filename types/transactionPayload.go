package types

import (
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"math/big"
)

type TransactionPayload struct {
	Payload bcs.Struct
}

const (
	TransactionPayload_Script        = 0
	TransactionPayload_ModuleBundle  = 1 // Deprecated
	TransactionPayload_EntryFunction = 2
	TransactionPayload_Multisig      = 3 // TODO? defined in aptos-core/types/src/transaction/mod.rs
)

func (txn *TransactionPayload) MarshalBCS(bcs *bcs.Serializer) {
	switch p := txn.Payload.(type) {
	case *Script:
		bcs.Uleb128(TransactionPayload_Script)
		p.MarshalBCS(bcs)
	case *ModuleBundle:
		// Deprecated, should never be seen
		bcs.Uleb128(TransactionPayload_ModuleBundle)
		p.MarshalBCS(bcs)
	case *EntryFunction:
		bcs.Uleb128(TransactionPayload_EntryFunction)
		p.MarshalBCS(bcs)
	default:
		bcs.SetError(fmt.Errorf("bad txn payload, %T", txn.Payload))
	}
}
func (txn *TransactionPayload) UnmarshalBCS(bcs *bcs.Deserializer) {
	kind := bcs.Uleb128()
	switch kind {
	case TransactionPayload_Script:
		xs := &Script{}
		xs.UnmarshalBCS(bcs)
		txn.Payload = xs
	case TransactionPayload_ModuleBundle:
		// Deprecated, should never be seen
		xs := &ModuleBundle{}
		xs.UnmarshalBCS(bcs)
		txn.Payload = xs
	case TransactionPayload_EntryFunction:
		xs := &EntryFunction{}
		xs.UnmarshalBCS(bcs)
		txn.Payload = xs
	default:
		bcs.SetError(fmt.Errorf("bad txn payload kind, %d", kind))
	}
}

// Execute a Script literal immediately as a transaction
type Script struct {
	Code     []byte
	ArgTypes []TypeTag
	Args     []ScriptArgument
}

func (sc *Script) MarshalBCS(serializer *bcs.Serializer) {
	serializer.WriteBytes(sc.Code)
	bcs.SerializeSequence(sc.ArgTypes, serializer)
	bcs.SerializeSequence(sc.Args, serializer)
}

func (sc *Script) UnmarshalBCS(deserializer *bcs.Deserializer) {
	sc.Code = deserializer.ReadBytes()
	sc.ArgTypes = bcs.DeserializeSequence[TypeTag](deserializer)
	sc.Args = bcs.DeserializeSequence[ScriptArgument](deserializer)
}

type ScriptArgument struct {
	Variant ScriptArgumentVariant
	Value   any
}

type ScriptArgumentVariant uint8

const (
	ScriptArgument_U8       ScriptArgumentVariant = 0
	ScriptArgument_U64      ScriptArgumentVariant = 1
	ScriptArgument_U128     ScriptArgumentVariant = 2
	ScriptArgument_Address  ScriptArgumentVariant = 3
	ScriptArgument_U8Vector ScriptArgumentVariant = 4
	ScriptArgument_Bool     ScriptArgumentVariant = 5
	ScriptArgument_U16      ScriptArgumentVariant = 6
	ScriptArgument_U32      ScriptArgumentVariant = 7
	ScriptArgument_U256     ScriptArgumentVariant = 8
)

func (sa *ScriptArgument) MarshalBCS(bcs *bcs.Serializer) {
	bcs.U8(uint8(sa.Variant))
	switch sa.Variant {
	case ScriptArgument_U8:
		bcs.U8(sa.Value.(uint8))
	case ScriptArgument_U16:
		bcs.U16(sa.Value.(uint16))
	case ScriptArgument_U32:
		bcs.U32(sa.Value.(uint32))
	case ScriptArgument_U64:
		bcs.U64(sa.Value.(uint64))
	case ScriptArgument_U128:
		bcs.U128(sa.Value.(big.Int))
	case ScriptArgument_U256:
		bcs.U256(sa.Value.(big.Int))
	case ScriptArgument_Address:
		sa.Value.(core.AccountAddress).MarshalBCS(bcs)
	case ScriptArgument_U8Vector:
		bcs.WriteBytes(sa.Value.([]byte))
	case ScriptArgument_Bool:
		bcs.Bool(sa.Value.(bool))
	}
}

// TODO: more like these Set*() accessors?
func (sa *ScriptArgument) SetU8(v uint8) {
	sa.Variant = ScriptArgument_U8
	sa.Value = v
}

// TODO: more like these Set*() accessors?
func (sa *ScriptArgument) SetU128(v big.Int) {
	sa.Variant = ScriptArgument_U128
	sa.Value = v
}

func (sa *ScriptArgument) UnmarshalBCS(bcs *bcs.Deserializer) {
	variant := bcs.U8()
	switch ScriptArgumentVariant(variant) {
	case ScriptArgument_U8:
		sa.Value = bcs.U8()
	case ScriptArgument_U16:
		sa.Value = bcs.U16()
	case ScriptArgument_U32:
		sa.Value = bcs.U32()
	case ScriptArgument_U64:
		sa.Value = bcs.U64()
	case ScriptArgument_U128:
		sa.Value = bcs.U128()
	case ScriptArgument_U256:
		sa.Value = bcs.U256()
	case ScriptArgument_Address:
		aa := core.AccountAddress{}
		aa.UnmarshalBCS(bcs)
		sa.Value = aa
	case ScriptArgument_U8Vector:
		sa.Value = bcs.ReadBytes()
	case ScriptArgument_Bool:
		sa.Value = bcs.Bool()
	}
}

// ModuleBundle is long deprecated and no longer used, but exist as an enum position in TransactionPayload
type ModuleBundle struct {
}

func (txn *ModuleBundle) MarshalBCS(bcs *bcs.Serializer) {
	bcs.SetError(errors.New("ModuleBunidle unimplemented"))
}
func (txn *ModuleBundle) UnmarshalBCS(bcs *bcs.Deserializer) {
	bcs.SetError(errors.New("ModuleBunidle unimplemented"))
}

// Call an existing published function
type EntryFunction struct {
	Module   ModuleId
	Function string
	ArgTypes []TypeTag
	Args     [][]byte
}

func (sf *EntryFunction) MarshalBCS(serializer *bcs.Serializer) {
	sf.Module.MarshalBCS(serializer)
	serializer.WriteString(sf.Function)
	bcs.SerializeSequence(sf.ArgTypes, serializer)
	serializer.Uleb128(uint32(len(sf.Args)))
	for _, a := range sf.Args {
		serializer.WriteBytes(a)
	}
}
func (sf *EntryFunction) UnmarshalBCS(deserializer *bcs.Deserializer) {
	sf.Module.UnmarshalBCS(deserializer)
	sf.Function = deserializer.ReadString()
	sf.ArgTypes = bcs.DeserializeSequence[TypeTag](deserializer)
	alen := deserializer.Uleb128()
	sf.Args = make([][]byte, alen)
	for i := range alen {
		sf.Args[i] = deserializer.ReadBytes()
	}
}
