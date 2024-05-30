package api

import (
	"encoding/json"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// MoveBytecode describes a module, or script, and it's associated ABI
type MoveBytecode struct {
	Bytecode []byte
	Abi      *MoveModule // Optional
}

func (o *MoveBytecode) UnmarshalJSON(b []byte) error {
	type inner struct {
		Bytecode []byte      `json:"bytecode"`
		Abi      *MoveModule `json:"abi,omitempty"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Bytecode = data.Bytecode
	o.Abi = data.Abi
	return nil
}

// MoveComponentId is an id for a struct, function, or other type e.g. 0x1::aptos_coin::AptosCoin
// TODO: more typing
type MoveComponentId = string

// MoveModule describes the abilities and types associated with a specific module
type MoveModule struct {
	Address          *types.AccountAddress `json:"address"`
	Name             string                `json:"name"`
	Friends          []MoveComponentId     `json:"friends"`           // TODO more typing
	ExposedFunctions []*MoveFunction       `json:"exposed_functions"` // TODO: more typing?
	Structs          []*MoveStruct         `json:"structs"`
}

type MoveScript struct {
	Bytecode []byte
	Abi      *MoveFunction // Optional
}

func (o *MoveScript) UnmarshalJSON(b []byte) error {
	type inner struct {
		Bytecode []byte        `json:"bytecode"`
		Abi      *MoveFunction `json:"abi,omitempty"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Bytecode = data.Bytecode
	o.Abi = data.Abi
	return nil
}

// MoveFunction describes a move function and its associated properties
type MoveFunction struct {
	Name              MoveComponentId     `json:"name"`
	Visibility        MoveVisibility      `json:"visibility"` // TODO: Typing?
	IsEntry           bool                `json:"is_entry""`
	IsView            bool                `json:"is_view"`
	GenericTypeParams []*GenericTypeParam `json:"generic_type_params"`
	Params            []string            `json:"params"`
	Return            []string            `json:"return"`
}

// GenericTypeParam is a set of requirements for a generic.  These can be applied via different
// MoveAbility constraints required on the type
type GenericTypeParam struct {
	Constraints []MoveAbility `json:"constraints"`
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
	Name              string              `json:"name"`
	IsNative          bool                `json:"is_native"`
	Abilities         []MoveAbility       `json:"abilities"`
	GenericTypeParams []*GenericTypeParam `json:"generic_type_params"`
	Fields            []*MoveStructField  `json:"fields"`
}

// MoveStructField represents a single field in a struct, and it's associated type
type MoveStructField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
