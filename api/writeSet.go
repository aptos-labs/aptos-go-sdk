package api

import (
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

func (o *WriteSet) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}

	switch o.Type {
	case EnumWriteSetDirect:
		o.Inner = &DirectWriteSet{}
	case EnumWriteSetScript:
		o.Inner = &ScriptWriteSet{}
	default:
		return fmt.Errorf("unknown writeset type: %s", o.Type)
	}

	return o.Inner.UnmarshalJSONFromMap(data)
}

type WriteSetImpl interface {
	UnmarshalFromMap
}

type DirectWriteSet struct {
	Changes []*WriteSetChange
	Events  []*Event
}

func (o *DirectWriteSet) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Changes, err = toWriteSetChanges(data, "changes")
	if err != nil {
		return err
	}
	o.Events, err = toEvents(data, "events")
	return err
}

type ScriptWriteSet struct {
	ExecuteAs *types.AccountAddress
	Script    *TransactionPayloadScript
}

func (o *ScriptWriteSet) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.ExecuteAs, err = toAccountAddress(data, "execute_as")
	if err != nil {
		return err
	}
	o.Script, err = toTransactionPayloadScript(data, "script")
	return err
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

func (o *WriteSetChange) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	// TODO: Implement these
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
	return o.Inner.UnmarshalJSONFromMap(data)
}

type WriteSetChangeImpl interface {
	UnmarshalFromMap
}

type WriteSetChangeWriteResource struct {
	Type         string
	Address      *types.AccountAddress
	StateKeyHash string
	Data         *MoveResource
}

func (o *WriteSetChangeWriteResource) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.Address, err = toAccountAddress(data, "address")
	if err != nil {
		return err
	}
	o.StateKeyHash, err = toHash(data, "state_key_hash")
	if err != nil {
		return err
	}
	o.Data, err = toMoveResource(data, "data")
	return err
}

type WriteSetChangeDeleteResource struct {
	Type         string
	Address      *types.AccountAddress
	StateKeyHash string
	// TODO: Resource is required, but doesn't seem to always show up
}

func (o *WriteSetChangeDeleteResource) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.Address, err = toAccountAddress(data, "address")
	if err != nil {
		return err
	}
	o.StateKeyHash, err = toHash(data, "state_key_hash")
	return err
}

type WriteSetChangeWriteModule struct {
	Type         string
	Address      *types.AccountAddress
	StateKeyHash string
	Data         *MoveBytecode
}

func (o *WriteSetChangeWriteModule) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.Address, err = toAccountAddress(data, "address")
	if err != nil {
		return err
	}
	o.StateKeyHash, err = toHash(data, "state_key_hash")
	if err != nil {
		return err
	}
	o.Data, err = toMoveBytecode(data, "data")
	return err
}

type WriteSetChangeDeleteModule struct {
	Type         string
	Address      *types.AccountAddress
	StateKeyHash string
	Module       string
}

func (o *WriteSetChangeDeleteModule) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.Address, err = toAccountAddress(data, "address")
	if err != nil {
		return err
	}
	o.StateKeyHash, err = toHash(data, "state_key_hash")
	if err != nil {
		return err
	}
	o.Module, err = toString(data, "module")
	return err
}

type WriteSetChangeWriteTableItem struct {
	Type         string
	StateKeyHash string
	Handle       string
	Key          string
	Value        string
	Data         *DecodedTableData // Optional, doesn't seem to be used
}

func (o *WriteSetChangeWriteTableItem) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.StateKeyHash, err = toHash(data, "state_key_hash")
	if err != nil {
		return err
	}
	o.Handle, err = toHash(data, "handle")
	if err != nil {
		return err
	}
	o.Key, err = toString(data, "key")
	if err != nil {
		return err
	}
	o.Value, err = toString(data, "value")
	if err != nil {
		return err
	}
	o.Data, err = toDecodedTableData(data, "data")
	return err
}

type WriteSetChangeDeleteTableItem struct {
	Type         string
	StateKeyHash string
	Handle       string
	Key          string
	Data         *DeletedTableData // This is optional, and never seems to be filled
}

func (o *WriteSetChangeDeleteTableItem) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.StateKeyHash, err = toHash(data, "state_key_hash")
	if err != nil {
		return err
	}
	o.Handle, err = toHash(data, "handle")
	if err != nil {
		return err
	}
	o.Key, err = toString(data, "key")
	if err != nil {
		return err
	}
	o.Data, err = toDeletedTableData(data, "data")
	return err
}

type DecodedTableData struct {
	Key       any
	KeyType   string
	Value     any
	ValueType string
}

func (o *DecodedTableData) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Key = data["key"]
	o.KeyType, err = toString(data, "key_type")
	if err != nil {
		return err
	}
	o.Value = data["value"]
	o.ValueType, err = toString(data, "value_type")
	return err
}

type DeletedTableData struct {
	Key     any
	KeyType string
}

func (o *DeletedTableData) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Key = data["key"]
	o.KeyType, err = toString(data, "key_type")
	return err
}

type MoveResource struct {
	Type string
	Data map[string]any
}

func (o *MoveResource) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	o.Data, err = toMap(data, "data")
	return err
}
