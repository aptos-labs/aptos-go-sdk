package api

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

const (
	EnumTransactionPayloadEntryFunction = "entry_function_payload"
	EnumTransactionPayloadScript        = "script_payload"
	EnumTransactionPayloadMultisig      = "multisig_payload"
	EnumTransactionPayloadWriteSet      = "write_set_payload"
	EnumTransactionModuleBundlePayload  = "module_bundle_payload" // Deprecated
)

// TransactionPayload is an enum of all possible transaction payloads
type TransactionPayload struct {
	Type  string
	Inner TransactionPayloadImpl
}

func (o *TransactionPayload) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	switch o.Type {
	case EnumTransactionPayloadEntryFunction:
		o.Inner = &TransactionPayloadEntryFunction{}
	case EnumTransactionPayloadScript:
		o.Inner = &TransactionPayloadScript{}
	case EnumTransactionPayloadMultisig:
		o.Inner = &TransactionPayloadMultisig{}
	case EnumTransactionPayloadWriteSet:
		o.Inner = &TransactionPayloadWriteSet{}
	case EnumTransactionModuleBundlePayload:
		o.Inner = &TransactionPayloadModuleBundle{}
	default:
		return fmt.Errorf("unknown transaction payload type: %s", o.Type)
	}

	return o.Inner.UnmarshalJSONFromMap(data)
}

// TransactionPayloadImpl is all the interfaces required for all transaction payloads
type TransactionPayloadImpl interface {
	UnmarshalFromMap
}

// TransactionPayloadEntryFunction describes an entry function call by a transaction
type TransactionPayloadEntryFunction struct {
	Function      string
	TypeArguments []string // TODO: Stronger typing?  (needs parser)
	Arguments     []any
}

func (o *TransactionPayloadEntryFunction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Function, err = toString(data, "function")
	if err != nil {
		return err
	}
	o.TypeArguments, err = toStrings(data, "type_arguments")
	if err != nil {
		return err
	}
	o.Arguments = data["arguments"].([]any)
	return nil
}

// TransactionPayloadScript describes a script payload along with associated
type TransactionPayloadScript struct {
	Code          *MoveScript
	TypeArguments []string // TODO: Stronger typing?  (needs parser)
	Arguments     []any
}

func (o *TransactionPayloadScript) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Code, err = toMoveScript(data, "code")
	if err != nil {
		return err
	}
	// Convert all to strings
	o.TypeArguments, err = toStrings(data, "type_arguments")

	o.Arguments = data["arguments"].([]any)
	return nil
}

// TransactionPayloadMultisig describes multisig running an entry function
type TransactionPayloadMultisig struct {
	MultisigAddress    *types.AccountAddress
	TransactionPayload *TransactionPayloadEntryFunction // Optional
}

func (o *TransactionPayloadMultisig) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.MultisigAddress, err = toAccountAddress(data, "multisig_address")
	if err != nil {
		return err
	}
	payload, err := toMap(data, "transaction_payload")
	if err == nil {
		o.TransactionPayload = &TransactionPayloadEntryFunction{}
		err = o.TransactionPayload.UnmarshalJSONFromMap(payload)
		if err != nil {
			return err
		}
	}
	return nil
}

// TransactionPayloadWriteSet  describes a write set transaction, such as genesis
type TransactionPayloadWriteSet struct {
	Inner *DirectWriteSet
}

func (o *TransactionPayloadWriteSet) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Inner = &DirectWriteSet{}
	writeSetData, err := toMap(data, "write_set")
	if err != nil {
		return err
	}
	return o.Inner.UnmarshalJSONFromMap(writeSetData)
}

type TransactionPayloadModuleBundle struct {
}

func (o *TransactionPayloadModuleBundle) UnmarshalJSONFromMap(data map[string]any) (err error) {
	return nil
}
