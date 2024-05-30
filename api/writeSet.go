package api

import (
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

const (
	EnumWriteSetDirect = "direct_write_set"
	EnumWriteSetScript = "script_write_set"
)

type WriteSet struct {
	Type  string
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
	o.Type = data.Type
	switch o.Type {
	case EnumWriteSetDirect:
		o.Inner = &DirectWriteSet{}
	case EnumWriteSetScript:
		o.Inner = &ScriptWriteSet{}
	default:
		return fmt.Errorf("unknown writeset type: %s", o.Type)
	}
	return json.Unmarshal(b, o.Inner)
}

type WriteSetImpl interface {
}

type DirectWriteSet struct {
	Changes []*WriteSetChange `json:"changes"`
	Events  []*Event          `json:"events"`
}

type ScriptWriteSet struct {
	ExecuteAs *types.AccountAddress     `json:"execute_as"`
	Script    *TransactionPayloadScript `json:"script"`
}

const (
	EnumWriteSetChangeWriteResource   = "write_resource"
	EnumWriteSetChangeDeleteResource  = "delete_resource"
	EnumWriteSetChangeWriteModule     = "write_module"
	EnumWriteSetChangeDeleteModule    = "delete_module"
	EnumWriteSetChangeWriteTableItem  = "write_table_item"
	EnumWriteSetChangeDeleteTableItem = "delete_table_item"
)

type WriteSetChange struct {
	Type  string
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
	o.Type = data.Type
	switch o.Type {
	case EnumWriteSetChangeWriteResource:
		o.Inner = &WriteSetChangeWriteResource{}
	case EnumWriteSetChangeDeleteResource:
		o.Inner = &WriteSetChangeDeleteResource{}
	case EnumWriteSetChangeWriteModule:
		o.Inner = &WriteSetChangeWriteModule{}
	case EnumWriteSetChangeDeleteModule:
		o.Inner = &WriteSetChangeDeleteModule{}
	case EnumWriteSetChangeWriteTableItem:
		o.Inner = &WriteSetChangeWriteTableItem{}
	case EnumWriteSetChangeDeleteTableItem:
		o.Inner = &WriteSetChangeDeleteTableItem{}
	default:
		return fmt.Errorf("unknown writeset change type: %s", o.Type)
	}
	return json.Unmarshal(b, o.Inner)
}

type WriteSetChangeImpl interface {
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
	StateKeyHash string                `json:"state_key_hash"`
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
