package api

import (
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// TransactionPayloadVariant is the type of payload represented in JSON
type TransactionPayloadVariant string

const (
	TransactionPayloadVariantEntryFunction       TransactionPayloadVariant = "entry_function_payload"
	TransactionPayloadVariantScript              TransactionPayloadVariant = "script_payload"
	TransactionPayloadVariantMultisig            TransactionPayloadVariant = "multisig_payload"
	TransactionPayloadVariantWriteSet            TransactionPayloadVariant = "write_set_payload"
	TransactionPayloadVariantModuleBundlePayload TransactionPayloadVariant = "module_bundle_payload" // Deprecated
)

// TransactionPayload is an enum of all possible transaction payloads
type TransactionPayload struct {
	Type  TransactionPayloadVariant
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
	o.Type = TransactionPayloadVariant(data.Type)
	// TODO: for all enums, we will likely have to add an "unknown type" so it doesn't just crash
	switch o.Type {
	case TransactionPayloadVariantEntryFunction:
		o.Inner = &TransactionPayloadEntryFunction{}
	case TransactionPayloadVariantScript:
		o.Inner = &TransactionPayloadScript{}
	case TransactionPayloadVariantMultisig:
		o.Inner = &TransactionPayloadMultisig{}
	case TransactionPayloadVariantWriteSet:
		o.Inner = &TransactionPayloadWriteSet{}
	case TransactionPayloadVariantModuleBundlePayload:
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

// TransactionPayloadWriteSet describes a write set transaction, such as genesis
type TransactionPayloadWriteSet = DirectWriteSet

// TransactionPayloadModuleBundle is a deprecated type that does not exist on mainnet
type TransactionPayloadModuleBundle struct{}
