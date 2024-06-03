package api

import (
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// MoveBytecode describes a module, or script, and it's associated ABI
type MoveBytecode struct {
	Bytecode []byte      `json:"bytecode"`
	Abi      *MoveModule `json:"abi,omitempty"` // Optional
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
	Bytecode []byte        `json:"bytecode"`
	Abi      *MoveFunction `json:"abi,omitempty"` // Optional
}

// MoveFunction describes a move function and its associated properties
type MoveFunction struct {
	Name              MoveComponentId     `json:"name"`
	Visibility        MoveVisibility      `json:"visibility"` // TODO: Typing?
	IsEntry           bool                `json:"is_entry"`
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

// MoveAbility are the types of abilities applied to structs, the possible types are listed
// as EnumMoveAbilityStore and others
type MoveAbility string

const (
	MoveAbilityVariantStore MoveAbility = "store"
	MoveAbilityVariantDrop  MoveAbility = "drop"
	MoveAbilityVariantKey   MoveAbility = "key"
	MoveAbilityVariantCopy  MoveAbility = "copy"
)

// MoveVisibility is the visibility of a function or struct, the possible types are listed
// as EnumMoveVisibilityPublic and others
type MoveVisibility string

const (
	MoveVisibilityPublic  MoveVisibility = "public"
	MoveVisibilityPrivate MoveVisibility = "private"
	MoveVisibilityFriend  MoveVisibility = "friend"
)

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
