package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

type WriteSetVariant string

const (
	WriteSetVariantDirect  WriteSetVariant = "direct_write_set"
	WriteSetVariantScript  WriteSetVariant = "script_write_set"
	WriteSetVariantUnknown WriteSetVariant = "unknown"
)

type WriteSet struct {
	Type  WriteSetVariant
	Inner WriteSetImpl
}

func (o *WriteSet) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = WriteSetVariant(data.Type)
	switch o.Type {
	case WriteSetVariantDirect:
		o.Inner = &DirectWriteSet{}
	case WriteSetVariantScript:
		o.Inner = &ScriptWriteSet{}
	case WriteSetVariantUnknown:
		o.Inner = &ScriptWriteSet{}
	default:
		o.Inner = &UnknownWriteSet{Type: string(o.Type)}
		o.Type = WriteSetVariantUnknown
		return json.Unmarshal(b, &o.Inner.(*UnknownWriteSet).Payload)
	}
	return json.Unmarshal(b, o.Inner)
}

type WriteSetImpl interface {
}

type UnknownWriteSet struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type DirectWriteSet struct {
	Changes []*WriteSetChange `json:"changes"`
	Events  []*Event          `json:"events"`
}

type ScriptWriteSet struct {
	ExecuteAs *types.AccountAddress     `json:"execute_as"`
	Script    *TransactionPayloadScript `json:"script"`
}

type WriteSetChangeVariant string

const (
	WriteSetChangeVariantWriteResource   WriteSetChangeVariant = "write_resource"
	WriteSetChangeVariantDeleteResource  WriteSetChangeVariant = "delete_resource"
	WriteSetChangeVariantWriteModule     WriteSetChangeVariant = "write_module"
	WriteSetChangeVariantDeleteModule    WriteSetChangeVariant = "delete_module"
	WriteSetChangeVariantWriteTableItem  WriteSetChangeVariant = "write_table_item"
	WriteSetChangeVariantDeleteTableItem WriteSetChangeVariant = "delete_table_item"
	WriteSetChangeVariantUnknown         WriteSetChangeVariant = "unknown"
)

type WriteSetChange struct {
	Type  WriteSetChangeVariant
	Inner WriteSetChangeImpl
}

func (o *WriteSetChange) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = WriteSetChangeVariant(data.Type)
	switch o.Type {
	case WriteSetChangeVariantWriteResource:
		o.Inner = &WriteSetChangeWriteResource{}
	case WriteSetChangeVariantDeleteResource:
		o.Inner = &WriteSetChangeDeleteResource{}
	case WriteSetChangeVariantWriteModule:
		o.Inner = &WriteSetChangeWriteModule{}
	case WriteSetChangeVariantDeleteModule:
		o.Inner = &WriteSetChangeDeleteModule{}
	case WriteSetChangeVariantWriteTableItem:
		o.Inner = &WriteSetChangeWriteTableItem{}
	case WriteSetChangeVariantDeleteTableItem:
		o.Inner = &WriteSetChangeDeleteTableItem{}
	default:
		o.Inner = &WriteSetChangeUnknown{Type: string(o.Type)}
		o.Type = WriteSetChangeVariantUnknown
		return json.Unmarshal(b, &o.Inner.(*WriteSetChangeUnknown).Payload)
	}
	return json.Unmarshal(b, o.Inner)
}

type WriteSetChangeImpl interface {
}

type WriteSetChangeUnknown struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type WriteSetChangeWriteResource struct {
	Type         string                `json:"type"`
	Address      *types.AccountAddress `json:"address"`
	StateKeyHash Hash                  `json:"state_key_hash"`
	Data         *MoveResource         `json:"data"`
}

type WriteSetChangeDeleteResource struct {
	Type         string                `json:"type"`
	Address      *types.AccountAddress `json:"address"`
	StateKeyHash Hash                  `json:"state_key_hash"`
	Resource     *MoveResource         `json:"resource"`
}

type WriteSetChangeWriteModule struct {
	Type         string                `json:"type"`
	Address      *types.AccountAddress `json:"address"`
	StateKeyHash Hash                  `json:"state_key_hash"`
	Data         *MoveBytecode         `json:"data"`
}

type WriteSetChangeDeleteModule struct {
	Type         string                `json:"type"`
	Address      *types.AccountAddress `json:"address"`
	StateKeyHash Hash                  `json:"state_key_hash"`
	Module       string                `json:"module"`
}

type WriteSetChangeWriteTableItem struct {
	Type         string            `json:"type"`
	StateKeyHash Hash              `json:"state_key_hash"`
	Handle       string            `json:"handle"`
	Key          string            `json:"key"`
	Value        string            `json:"value"`
	Data         *DecodedTableData `json:"data,omitempty"` // Optional, doesn't seem to be used
}

type WriteSetChangeDeleteTableItem struct {
	Type         string            `json:"type"`
	StateKeyHash string            `json:"state_key_hash"`
	Handle       string            `json:"handle"`
	Key          string            `json:"key"`
	Data         *DeletedTableData `json:"data,omitempty"` // This is optional, and never seems to be filled
}

type DecodedTableData struct {
	Key       any    `json:"key"`
	KeyType   string `json:"key_type"`
	Value     any    `json:"value"`
	ValueType string `json:"value_type"`
}

type DeletedTableData struct {
	Key     any    `json:"key"`
	KeyType string `json:"key_type"`
}

type MoveResource struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}
