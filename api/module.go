package api

import "github.com/aptos-labs/aptos-go-sdk/internal/types"

// MoveBytecode describes a module, or script, and it's associated ABI
type MoveBytecode struct {
	Bytecode []byte
	Abi      *MoveModule // Optional
}

func (o *MoveBytecode) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Bytecode, err = toBytes(data, "bytecode")
	if err != nil {
		return err
	}
	if data["abi"] != nil {
		o.Abi = &MoveModule{}
		abiData, err := toMap(data, "abi")
		if err != nil {
			return err
		}
		err = o.Abi.UnmarshalJSONFromMap(abiData)
	}
	return err
}

// MoveComponentId is an id for a struct, function, or other type e.g. 0x1::aptos_coin::AptosCoin
// TODO: more typing
type MoveComponentId = string

// MoveModule describes the abilities and types associated with a specific module
type MoveModule struct {
	Address          *types.AccountAddress
	Name             string
	Friends          []MoveComponentId // TODO more typing
	ExposedFunctions []*MoveFunction   // TODO: more typing?
	Structs          []*MoveStruct
}

func (o *MoveModule) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Address, err = toAccountAddress(data, "address")
	if err != nil {
		return err
	}
	o.Name, err = toString(data, "name")
	if err != nil {
		return err
	}
	o.Friends, err = toStrings(data, "friends")
	if err != nil {
		return err
	}
	o.ExposedFunctions, err = toMoveFunctions(data, "exposed_functions")
	if err != nil {
		return err
	}
	o.Structs, err = toMoveStructs(data, "structs")
	return err
}

type MoveScript struct {
	Bytecode []byte
	Abi      *MoveFunction // Optional
}

func (o *MoveScript) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Bytecode, err = toBytes(data, "bytecode")
	if err != nil {
		return err
	}
	// ABI is optional
	if data["abi"] != nil {
		o.Abi, err = toMoveFunction(data, "abi")
	}
	return err
}

// MoveFunction describes a move function and its associated properties
type MoveFunction struct {
	Name              MoveComponentId
	Visibility        MoveVisibility // TODO: Typing
	IsEntry           bool
	IsView            bool
	GenericTypeParams []*GenericTypeParam
	Params            []string
	Return            []string
}

func (o *MoveFunction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Name, err = toString(data, "name")
	if err != nil {
		return err
	}
	o.Visibility, err = toString(data, "visibility")
	if err != nil {
		return err
	}
	o.IsEntry, err = toBool(data, "is_entry")
	if err != nil {
		return err
	}
	o.IsView, err = toBool(data, "is_view")
	if err != nil {
		return err
	}
	o.GenericTypeParams, err = toGenericTypeParams(data, "generic_type_params")
	if err != nil {
		return err
	}
	o.Params, err = toStrings(data, "params")
	if err != nil {
		return err
	}
	o.Return, err = toStrings(data, "return")
	return err
}

// GenericTypeParam is a set of requirements for a generic.  These can be applied via different
// MoveAbility constraints required on the type
type GenericTypeParam struct {
	Constraints []MoveAbility
}

func (o *GenericTypeParam) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Constraints, err = toStrings(data, "constraints")
	return err
}

const (
	EnumMoveAbilityStore = "store"
	EnumMoveAbilityDrop  = "drop"
	EnumMoveAbilityKey   = "key"
	EnumMoveAbilityCopy  = "copy"
)

// MoveAbility are the types of abilities applied to structs, the possible types are listed
// as EnumMoveAbilityStore and others
type MoveAbility = string

const (
	EnumMoveVisibilityPublic  = "public"
	EnumMoveVisibilityPrivate = "private"
	EnumMoveVisibilityFriend  = "friend"
)

// MoveVisibility is the visibility of a function or struct, the possible types are listed
// as EnumMoveVisibilityPublic and others
type MoveVisibility = string

// MoveStruct describes the layout for a struct, and its constraints
type MoveStruct struct {
	Name              string
	IsNative          bool
	Abilities         []MoveAbility
	GenericTypeParams []*GenericTypeParam
	Fields            []*MoveStructField
}

func (o *MoveStruct) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Name, err = toString(data, "name")
	if err != nil {
		return err
	}
	o.IsNative, err = toBool(data, "is_native")
	if err != nil {
		return err
	}
	o.Abilities, err = toStrings(data, "abilities")
	if err != nil {
		return err
	}
	o.GenericTypeParams, err = toGenericTypeParams(data, "generic_type_params")
	if err != nil {
		return err
	}
	o.Fields, err = toMoveStructFields(data, "fields")
	return err
}

// MoveStructField represents a single field in a struct, and it's associated type
type MoveStructField struct {
	Name string
	Type string
}

func (o *MoveStructField) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Name, err = toString(data, "name")
	if err != nil {
		return err
	}
	o.Type, err = toString(data, "type")
	return err
}
