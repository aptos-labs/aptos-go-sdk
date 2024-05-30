package api

import (
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"strings"
)

// PrettyJson a simple pretty print for JSON examples
func PrettyJson(x any) string {
	out := strings.Builder{}
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "  ")
	err := enc.Encode(x)
	if err != nil {
		return ""
	}
	return out.String()
}

// UnmarshalFromMap to allow for unmarshalling JSON from a map instead of bytes for nested deserialization
type UnmarshalFromMap interface {
	UnmarshalJSONFromMap(data map[string]any) (err error)
}

func toAccountAddress(data map[string]any, key string) (*types.AccountAddress, error) {
	str, ok := data[key].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert key %s to string", key)
	}
	address := &types.AccountAddress{}
	err := address.ParseStringRelaxed(str)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func toBool(data map[string]any, key string) (bool, error) {
	b, ok := data[key].(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert key %s to bool", key)
	}
	return b, nil
}
func toUint8(data map[string]any, key string) (uint8, error) {
	str, ok := data[key].(int)
	if !ok {
		return 0, fmt.Errorf("failed to convert key %s to uint8", key)
	}
	return uint8(str), nil
}

func toUint64(data map[string]any, key string) (uint64, error) {
	str, ok := data[key].(string)
	if !ok {
		return 0, fmt.Errorf("failed to convert key %s to uint64", key)
	}
	return util.StrToUint64(str)
}

func toString(data map[string]any, key string) (string, error) {
	str, ok := data[key].(string)
	if !ok {
		return "", fmt.Errorf("failed to convert key %s to string", key)
	}
	return str, nil
}

func toBytes(data map[string]any, key string) ([]byte, error) {
	hexStr, err := toString(data, key)
	if err != nil {
		return nil, err
	}
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func toHash(data map[string]any, key string) (string, error) {
	str, ok := data[key].(string)
	if !ok {
		return "", fmt.Errorf("failed to convert key %s to hash", key)
	}
	return str, nil
}

func toMap(data map[string]any, key string) (map[string]any, error) {
	out, ok := data[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to convert key %s to map[string]any", key)
	}
	return out, nil
}

func toGuid(data map[string]any, key string) (*GUID, error) {
	return toStruct(data, key, func() *GUID { return &GUID{} }, func() *GUID { return nil })
}

func toPayload(data map[string]any, key string) (*TransactionPayload, error) {
	return toStruct(data, key, func() *TransactionPayload { return &TransactionPayload{} }, func() *TransactionPayload { return nil })
}

func toSignature(data map[string]any, key string) (*Signature, error) {
	return toStruct(data, key, func() *Signature { return &Signature{} }, func() *Signature { return nil })
}

func toMoveFunction(data map[string]any, key string) (*MoveFunction, error) {
	return toStruct(data, key, func() *MoveFunction { return &MoveFunction{} }, func() *MoveFunction { return nil })
}

func toMoveScript(data map[string]any, key string) (*MoveScript, error) {
	return toStruct(data, key, func() *MoveScript { return &MoveScript{} }, func() *MoveScript { return nil })
}
func toTransactionPayloadScript(data map[string]any, key string) (*TransactionPayloadScript, error) {
	return toStruct(data, key, func() *TransactionPayloadScript { return &TransactionPayloadScript{} }, func() *TransactionPayloadScript { return nil })
}
func toMoveResource(data map[string]any, key string) (*MoveResource, error) {
	return toStruct(data, key, func() *MoveResource { return &MoveResource{} }, func() *MoveResource { return nil })
}
func toMoveBytecode(data map[string]any, key string) (*MoveBytecode, error) {
	return toStruct(data, key, func() *MoveBytecode { return &MoveBytecode{} }, func() *MoveBytecode { return nil })
}

func toDecodedTableData(data map[string]any, key string) (*DecodedTableData, error) {
	// This is optional, and we'll just pass it along
	if data[key] == nil {
		return nil, nil
	}
	return toStruct(data, key, func() *DecodedTableData { return &DecodedTableData{} }, func() *DecodedTableData { return nil })
}
func toDeletedTableData(data map[string]any, key string) (*DeletedTableData, error) {
	// This is optional, and we'll just pass it along
	if data[key] == nil {
		return nil, nil
	}
	return toStruct(data, key, func() *DeletedTableData { return &DeletedTableData{} }, func() *DeletedTableData { return nil })
}
func toStruct[T UnmarshalFromMap](data map[string]any, key string, create func() T, createNil func() T) (out T, err error) {
	data, err = toMap(data, key)
	if err != nil {
		return createNil(), err
	}
	out = create()
	err = out.UnmarshalJSONFromMap(data)
	if err != nil {
		return createNil(), err
	}
	return out, nil
}

func toArray[T any](data map[string]any, key string, convert func(input []any) ([]T, error)) ([]T, error) {
	// If not present, then it's empty
	if data[key] == nil {
		return make([]T, 0), nil
	}
	// Otherwise, it should be an array, or we've got an issue
	val, ok := data[key].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to convert key %s to []any", key)
	}
	return convert(val)
}
func toUint8Array(data map[string]any, key string) ([]uint8, error) {
	return toArray(data, key, func(inputs []any) ([]uint8, error) {
		output := make([]uint8, len(inputs))
		for i, input := range inputs {
			out, ok := input.(float64) // Numbers in JSON are float64 :/
			if !ok {
				return nil, fmt.Errorf("failed to convert key %s to []uint8", key)
			}
			output[i] = uint8(out)
		}
		return output, nil
	})
}
func toUint32Array(data map[string]any, key string) ([]uint32, error) {
	return toArray(data, key, func(inputs []any) ([]uint32, error) {
		output := make([]uint32, len(inputs))
		for i, input := range inputs {
			out, ok := input.(int)
			if !ok {
				return nil, fmt.Errorf("failed to convert key %s to []uint32", key)
			}
			output[i] = uint32(out)
		}
		return output, nil
	})
}

func toStrings(data map[string]any, key string) ([]string, error) {
	return toArray(data, key, func(inputs []any) ([]string, error) {
		output := make([]string, len(inputs))
		for i, input := range inputs {
			out, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("failed to convert key %s to []string", key)
			}
			output[i] = out
		}
		return output, nil
	})
}

func toPointerArray[T UnmarshalFromMap](data map[string]any, key string, create func() T) ([]T, error) {
	return toArray(data, key, func(inputs []any) ([]T, error) {
		var err error
		output := make([]T, len(inputs))
		for i, input := range inputs {
			output[i] = create()
			eventData, ok := input.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("failed to read input %s key as map[string]any", key)
			}
			err = output[i].UnmarshalJSONFromMap(eventData)
			if err != nil {
				return nil, err
			}
		}
		return output, nil
	})
}

func toMoveFunctions(data map[string]any, key string) ([]*MoveFunction, error) {
	return toPointerArray(data, key, func() *MoveFunction { return &MoveFunction{} })
}

func toMoveStructs(data map[string]any, key string) ([]*MoveStruct, error) {
	return toPointerArray(data, key, func() *MoveStruct { return &MoveStruct{} })
}
func toMoveStructFields(data map[string]any, key string) ([]*MoveStructField, error) {
	return toPointerArray(data, key, func() *MoveStructField { return &MoveStructField{} })
}

func toGenericTypeParams(data map[string]any, key string) ([]*GenericTypeParam, error) {
	return toPointerArray(data, key, func() *GenericTypeParam { return &GenericTypeParam{} })
}

func toTransactions(data map[string]any, key string) ([]*Transaction, error) {
	return toPointerArray(data, key, func() *Transaction { return &Transaction{} })
}

func toEvents(data map[string]any, key string) ([]*Event, error) {
	return toPointerArray(data, key, func() *Event { return &Event{} })
}

func toWriteSetChanges(data map[string]any, key string) ([]*WriteSetChange, error) {
	return toPointerArray(data, key, func() *WriteSetChange { return &WriteSetChange{} })
}
func toSignatures(data map[string]any, key string) ([]*Signature, error) {
	return toPointerArray(data, key, func() *Signature { return &Signature{} })
}

func toEd25519Signatures(data map[string]any, key string) ([]*Ed25519Authenticator, error) {
	return toPointerArray(data, key, func() *Ed25519Authenticator { return &Ed25519Authenticator{} })
}

func toAccountAddresses(data map[string]any, key string) ([]*types.AccountAddress, error) {
	return toArray(data, key, func(inputs []any) ([]*types.AccountAddress, error) {
		var err error
		output := make([]*types.AccountAddress, len(inputs))
		for i, input := range inputs {
			output[i] = &types.AccountAddress{}
			input, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("failed to read input %s key as string", key)
			}
			err = output[i].ParseStringRelaxed(input)
			if err != nil {
				return nil, err
			}
		}
		return output, nil
	})
}
