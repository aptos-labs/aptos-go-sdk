package aptos

import (
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

//region TransactionPayload

type TransactionPayloadVariant uint32

const (
	TransactionPayloadVariantScript        TransactionPayloadVariant = 0
	TransactionPayloadVariantModuleBundle  TransactionPayloadVariant = 1 // Deprecated
	TransactionPayloadVariantEntryFunction TransactionPayloadVariant = 2
	TransactionPayloadVariantMultisig      TransactionPayloadVariant = 3
)

type TransactionPayloadImpl interface {
	bcs.Struct
	PayloadType() TransactionPayloadVariant // This is specifically to ensure that wrong types don't end up here
}

// TransactionPayload the actual instructions of which functions to call on chain
type TransactionPayload struct {
	Payload TransactionPayloadImpl
}

//region TransactionPayload bcs.Struct

func (txn *TransactionPayload) MarshalBCS(bcs *bcs.Serializer) {
	if txn == nil || txn.Payload == nil {
		bcs.SetError(fmt.Errorf("nil transaction payload"))
		return
	}
	bcs.Uleb128(uint32(txn.Payload.PayloadType()))
	txn.Payload.MarshalBCS(bcs)
}
func (txn *TransactionPayload) UnmarshalBCS(bcs *bcs.Deserializer) {
	payloadType := TransactionPayloadVariant(bcs.Uleb128())
	switch payloadType {
	case TransactionPayloadVariantScript:
		txn.Payload = &Script{}
	case TransactionPayloadVariantModuleBundle:
		// Deprecated, should never be in production
		bcs.SetError(fmt.Errorf("module bundle is not supported as a transaction payload"))
	case TransactionPayloadVariantEntryFunction:
		txn.Payload = &EntryFunction{}
	case TransactionPayloadVariantMultisig:
		txn.Payload = &Multisig{}
	default:
		bcs.SetError(fmt.Errorf("bad txn payload kind, %d", payloadType))
	}

	txn.Payload.UnmarshalBCS(bcs)
}

//endregion
//endregion

//region ModuleBundle

// ModuleBundle is long deprecated and no longer used, but exist as an enum position in TransactionPayload
type ModuleBundle struct{}

func (txn *ModuleBundle) PayloadType() TransactionPayloadVariant {
	return TransactionPayloadVariantModuleBundle
}

func (txn *ModuleBundle) MarshalBCS(bcs *bcs.Serializer) {
	bcs.SetError(errors.New("ModuleBundle unimplemented"))
}
func (txn *ModuleBundle) UnmarshalBCS(bcs *bcs.Deserializer) {
	bcs.SetError(errors.New("ModuleBundle unimplemented"))
}

//endregion ModuleBundle

//region EntryFunction

// EntryFunction call a single published entry function arguments are ordered BCS encoded bytes
type EntryFunction struct {
	Module   ModuleId
	Function string
	ArgTypes []TypeTag
	Args     [][]byte
}

//region EntryFunction TransactionPayloadImpl

func (sf *EntryFunction) PayloadType() TransactionPayloadVariant {
	return TransactionPayloadVariantEntryFunction
}

//endregion

//region EntryFunction bcs.Struct

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

//endregion
//endregion

//region Multisig

// Multisig is an on-chain multisig transaction, that calls an entry function associated
type Multisig struct {
	MultisigAddress AccountAddress
	Payload         *MultisigTransactionPayload // Optional
}

//region Multisig TransactionPayloadImpl

func (sf *Multisig) PayloadType() TransactionPayloadVariant {
	return TransactionPayloadVariantMultisig
}

//endregion

//region Multisig bcs.Struct

func (sf *Multisig) MarshalBCS(serializer *bcs.Serializer) {
	serializer.Struct(&sf.MultisigAddress)
	if sf.Payload == nil {
		serializer.Bool(false)
	} else {
		serializer.Bool(true)
		serializer.Struct(sf.Payload)
	}
}
func (sf *Multisig) UnmarshalBCS(deserializer *bcs.Deserializer) {
	deserializer.Struct(&sf.MultisigAddress)
	if deserializer.Bool() {
		sf.Payload = &MultisigTransactionPayload{}
		deserializer.Struct(sf.Payload)
	}
}

//endregion
//endregion

//region MultisigTransactionPayload

type MultisigTransactionPayloadVariant uint32

const (
	MultisigTransactionPayloadVariantEntryFunction MultisigTransactionPayloadVariant = 0
)

type MultisigTransactionImpl interface {
	bcs.Struct
}

// MultisigTransactionPayload is an enum allowing for multiple types of transactions to be called via multisig
//
// Note this does not implement TransactionPayloadImpl
type MultisigTransactionPayload struct {
	Variant MultisigTransactionPayloadVariant
	Payload MultisigTransactionImpl
}

//region MultisigTransactionPayload bcs.Struct

func (sf *MultisigTransactionPayload) MarshalBCS(serializer *bcs.Serializer) {
	serializer.Uleb128(uint32(sf.Variant))
	serializer.Struct(sf.Payload)
}
func (sf *MultisigTransactionPayload) UnmarshalBCS(deserializer *bcs.Deserializer) {
	variant := MultisigTransactionPayloadVariant(deserializer.Uleb128())
	switch variant {
	case MultisigTransactionPayloadVariantEntryFunction:
		sf.Payload = &EntryFunction{}
	default:
		deserializer.SetError(fmt.Errorf("bad variant %d for MultisigTransactionPayload", variant))
	}
	deserializer.Struct(sf.Payload)
}

//endregion
//endregion
