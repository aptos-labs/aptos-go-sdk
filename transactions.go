package aptos

import (
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

type RawTransaction struct {
	Sender         AccountAddress
	SequenceNumber uint64
	Payload        TransactionPayload
	MaxGasAmount   uint64
	GasUnitPrice   uint64

	// ExpirationTimetampSeconds is seconds since Unix epoch
	ExpirationTimetampSeconds uint64

	ChainId uint8
}

func (txn *RawTransaction) MarshalBCS(bcs *Serializer) {
	txn.Sender.MarshalBCS(bcs)
	bcs.U64(txn.SequenceNumber)
	txn.Payload.MarshalBCS(bcs)
	bcs.U64(txn.MaxGasAmount)
	bcs.U64(txn.GasUnitPrice)
	bcs.U64(txn.ExpirationTimetampSeconds)
	bcs.U8(txn.ChainId)
}

func (txn *RawTransaction) UnmarshalBCS(bcs *Deserializer) {
	txn.Sender.UnmarshalBCS(bcs)
	txn.SequenceNumber = bcs.U64()
	txn.Payload.UnmarshalBCS(bcs)
	txn.MaxGasAmount = bcs.U64()
	txn.GasUnitPrice = bcs.U64()
	txn.ExpirationTimetampSeconds = bcs.U64()
	txn.ChainId = bcs.U8()
}

func (txn *RawTransaction) SignableBytes() (signableBytes []byte, err error) {
	ser := Serializer{}
	txn.MarshalBCS(&ser)
	err = ser.Error()
	if err != nil {
		return
	}
	prehash := RawTransactionPrehash()
	txnbytes := ser.ToBytes()
	signableBytes = make([]byte, len(prehash)+len(txnbytes))
	copy(signableBytes, prehash)
	copy(signableBytes[len(prehash):], txnbytes)
	return signableBytes, nil
}

func (txn *RawTransaction) Sign(sender *Account) (stxn *SignedTransaction, err error) {
	signableBytes, err := txn.SignableBytes()
	if err != nil {
		return
	}
	authenticator, err := sender.Sign(signableBytes)
	if err != nil {
		return
	}

	stxn = &SignedTransaction{
		Transaction:   *txn,
		Authenticator: authenticator,
	}
	return
}

var rawTransactionPrehash []byte

const rawTransactionPrehashStr = "APTOS::RawTransaction"

// Return the sha3-256 prehash for RawTransaction
// Do not write to the []byte returned
func RawTransactionPrehash() []byte {
	if rawTransactionPrehash == nil {
		b32 := sha3.Sum256([]byte(rawTransactionPrehashStr))
		out := make([]byte, len(b32))
		copy(out, b32[:])
		rawTransactionPrehash = out
		return out
	}
	return rawTransactionPrehash
}

type TransactionPayload struct {
	Payload BCSStruct
}

const (
	TransactionPayload_Script        = 0
	TransactionPayload_ModuleBundle  = 1 // Deprecated
	TransactionPayload_EntryFunction = 2
	TransactionPayload_Multisig      = 3 // TODO? defined in aptos-core/types/src/transaction/mod.rs
)

func (txn *TransactionPayload) MarshalBCS(bcs *Serializer) {
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
func (txn *TransactionPayload) UnmarshalBCS(bcs *Deserializer) {
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

func (sc *Script) MarshalBCS(bcs *Serializer) {
	bcs.WriteBytes(sc.Code)
	SerializeSequence(sc.ArgTypes, bcs)
	SerializeSequence(sc.Args, bcs)
}
func (sc *Script) UnmarshalBCS(bcs *Deserializer) {
	sc.Code = bcs.ReadBytes()
	sc.ArgTypes = DeserializeSequence[TypeTag](bcs)
	sc.Args = DeserializeSequence[ScriptArgument](bcs)
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

func (sa *ScriptArgument) MarshalBCS(bcs *Serializer) {
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
		sa.Value.(AccountAddress).MarshalBCS(bcs)
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

func (sa *ScriptArgument) UnmarshalBCS(bcs *Deserializer) {
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
		aa := AccountAddress{}
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

func (txn *ModuleBundle) MarshalBCS(bcs *Serializer) {
	bcs.SetError(errors.New("ModuleBunidle unimplemented"))
}
func (txn *ModuleBundle) UnmarshalBCS(bcs *Deserializer) {
	bcs.SetError(errors.New("ModuleBunidle unimplemented"))
}

// Call an existing published function
type EntryFunction struct {
	Module   ModuleId
	Function string
	ArgTypes []TypeTag
	Args     [][]byte
}

func (sf *EntryFunction) MarshalBCS(bcs *Serializer) {
	sf.Module.MarshalBCS(bcs)
	bcs.WriteString(sf.Function)
	SerializeSequence(sf.ArgTypes, bcs)
	bcs.Uleb128(uint32(len(sf.Args)))
	for _, a := range sf.Args {
		bcs.WriteBytes(a)
	}
}
func (sf *EntryFunction) UnmarshalBCS(bcs *Deserializer) {
	sf.Module.UnmarshalBCS(bcs)
	sf.Function = bcs.ReadString()
	sf.ArgTypes = DeserializeSequence[TypeTag](bcs)
	alen := bcs.Uleb128()
	sf.Args = make([][]byte, alen)
	for i := range alen {
		sf.Args[i] = bcs.ReadBytes()
	}
}

type ModuleId struct {
	Address AccountAddress
	Name    string
}

func (mod *ModuleId) MarshalBCS(bcs *Serializer) {
	mod.Address.MarshalBCS(bcs)
	bcs.WriteString(mod.Name)
}
func (mod *ModuleId) UnmarshalBCS(bcs *Deserializer) {
	mod.Address.UnmarshalBCS(bcs)
	mod.Name = bcs.ReadString()
}

type SignedTransaction struct {
	Transaction   RawTransaction
	Authenticator Authenticator
}

func (txn *SignedTransaction) MarshalBCS(bcs *Serializer) {
	txn.Transaction.MarshalBCS(bcs)
	txn.Authenticator.MarshalBCS(bcs)
}
func (txn *SignedTransaction) UnmarshalBCS(bcs *Deserializer) {
	txn.Transaction.UnmarshalBCS(bcs)
	txn.Authenticator.UnmarshalBCS(bcs)
}

func (txn *SignedTransaction) Verify() error {
	tbytes, err := txn.Transaction.SignableBytes()
	if err != nil {
		return err
	}
	if txn.Authenticator.Verify(tbytes) {
		return nil
	}
	return errors.New("Bad Signature")
}
