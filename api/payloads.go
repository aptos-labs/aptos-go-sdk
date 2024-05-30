package api

import (
	"encoding/json"
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

func (o *TransactionPayload) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = data.Type
	// TODO: for all enums, we will likely have to add an "unknown type" so it doesn't just crash
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
	return json.Unmarshal(b, o.Inner)
}

// TransactionPayloadImpl is all the interfaces required for all transaction payloads
type TransactionPayloadImpl interface {
}

// TransactionPayloadEntryFunction describes an entry function call by a transaction
type TransactionPayloadEntryFunction struct {
	Function      string   `json:"function"`
	TypeArguments []string `json:"type_arguments"` // TODO: Stronger typing?  (needs parser)
	Arguments     []any    `json:"arguments"`
}

// TransactionPayloadScript describes a script payload along with associated
type TransactionPayloadScript struct {
	Code          *MoveScript `json:"code"`
	TypeArguments []string    `json:"type_arguments"`
	Arguments     []any       `json:"arguments"`
}

// TransactionPayloadMultisig describes multisig running an entry function
type TransactionPayloadMultisig struct {
	MultisigAddress    *types.AccountAddress            `json:"multisig_address"`
	TransactionPayload *TransactionPayloadEntryFunction `json:"transaction_payload,omitempty"` // Optional
}

// TransactionPayloadWriteSet  describes a write set transaction, such as genesis
type TransactionPayloadWriteSet = DirectWriteSet

type TransactionPayloadModuleBundle struct{}
