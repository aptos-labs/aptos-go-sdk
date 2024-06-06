package aptos

import (
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// TransactionPayload the actual instructions of which functions to call on chain
type TransactionPayload struct {
	Payload bcs.Struct
}

const (
	TransactionPayloadScript        = 0
	TransactionPayloadModuleBundle  = 1 // Deprecated
	TransactionPayloadEntryFunction = 2
	TransactionPayloadMultisig      = 3
)

func (txn *TransactionPayload) MarshalBCS(bcs *bcs.Serializer) {
	switch p := txn.Payload.(type) {
	case *Script:
		bcs.Uleb128(TransactionPayloadScript)
		p.MarshalBCS(bcs)
	case *ModuleBundle:
		// Deprecated, this never was used in production, and we will just drop it
		bcs.SetError(fmt.Errorf("module bundle is not supported as a transaction payload"))
	case *EntryFunction:
		bcs.Uleb128(TransactionPayloadEntryFunction)
		p.MarshalBCS(bcs)
	default:
		bcs.SetError(fmt.Errorf("bad txn payload, %T", txn.Payload))
	}
}
func (txn *TransactionPayload) UnmarshalBCS(bcs *bcs.Deserializer) {
	kind := bcs.Uleb128()
	switch kind {
	case TransactionPayloadScript:
		txn.Payload = &Script{}
	case TransactionPayloadModuleBundle:
		// Deprecated, should never be in production
		bcs.SetError(fmt.Errorf("module bundle is not supported as a transaction payload"))
	case TransactionPayloadEntryFunction:
		txn.Payload = &EntryFunction{}
	case TransactionPayloadMultisig:
		txn.Payload = &Multisig{}
	default:
		bcs.SetError(fmt.Errorf("bad txn payload kind, %d", kind))
	}

	txn.Payload.UnmarshalBCS(bcs)
}

// ModuleBundle is long deprecated and no longer used, but exist as an enum position in TransactionPayload
type ModuleBundle struct {
}

func (txn *ModuleBundle) MarshalBCS(bcs *bcs.Serializer) {
	bcs.SetError(errors.New("ModuleBundle unimplemented"))
}
func (txn *ModuleBundle) UnmarshalBCS(bcs *bcs.Deserializer) {
	bcs.SetError(errors.New("ModuleBundle unimplemented"))
}

// EntryFunction call a single published entry function
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

// Multisig is an on-chain multisig transaction, that calls an entry function associated
type Multisig struct {
	MultisigAddress AccountAddress
	Payload         *MultisigTransactionPayload // Optional
}

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

type MultisigTransactionPayloadVariant uint32

const (
	MultisigTransactionPayloadVariantEntryFunction MultisigTransactionPayloadVariant = 0
)

type MultisigTransactionImpl interface {
	bcs.Struct
}

// MultisigTransactionPayload is an enum allowing for multiple types of transactions to be called via multisig
type MultisigTransactionPayload struct {
	Variant MultisigTransactionPayloadVariant
	Payload MultisigTransactionImpl
}

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
