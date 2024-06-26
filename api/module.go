package api

import (
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

// MoveBytecode describes a module, or script, and it's associated ABI
//
// Example 0x1::coin:
//
//	{
//		"bytecode": "0xa11ceb0b123456...",
//		"abi": {
//			"address": "0x1",
//			"name": "coin",
//			"friends": [
//				"0x1::aptos_coin",
//				"0x1::genesis",
//				"0x1::transaction_fee"
//			],
//			"exposed_functions": [
//				{
//					"name": "balance",
//					"visibility": "public",
//					"is_entry": false,
//					"is_view": true,
//					"generic_type_params": [
//						{
//							"constraints": []
//						}
//					],
//					"params": [
//						"address"
//					],
//					"return": [
//						"u64"
//					]
//				}
//			],
//			"structs": [
//				{
//					"name": "Coin",
//					"is_native": false,
//					"abilities": [
//						"store"
//					],
//					"generic_type_params": [
//						{
//							"constraints": []
//						}
//					],
//					"fields": [
//						{
//							"name": "value",
//							"type": "u64"
//						}
//					]
//				},
//			],
//		}
//	}
type MoveBytecode struct {
	Bytecode HexBytes    `json:"bytecode"`      // Bytecode is the hex encoded version of the compiled module
	Abi      *MoveModule `json:"abi,omitempty"` // Abi is the ABI for the module, and is optional
}

// MoveComponentId is an id for a struct, function, or other type e.g. 0x1::aptos_coin::AptosCoin
type MoveComponentId = string

// MoveModule describes the abilities and types associated with a specific module
type MoveModule struct {
	Address          *types.AccountAddress `json:"address"`           // Address is the address of the module e.g. 0x1
	Name             string                `json:"name"`              // Name is the name of the module e.g. coin
	Friends          []MoveComponentId     `json:"friends"`           // Friends are other modules that can access this module in the same package
	ExposedFunctions []*MoveFunction       `json:"exposed_functions"` // ExposedFunctions are the functions that can be called from outside the module
	Structs          []*MoveStruct         `json:"structs"`           // Structs are the structs defined in the module
}

// MoveScript is the representation of a compiled script.  The API may not fill in the ABI field
//
// Example:
//
//	{
//		"bytecode": "0xa11ceb0b123456...",
//		"abi": {
//			"address": "0x1",
//			"name": "coin",
//			"friends": [
//				"0x1::aptos_coin",
//				"0x1::genesis",
//				"0x1::transaction_fee"
//			],
//			"exposed_functions": [
//				{
//					"name": "balance",
//					"visibility": "public",
//					"is_entry": false,
//					"is_view": true,
//					"generic_type_params": [
//						{
//							"constraints": []
//						}
//					],
//					"params": [
//						"address"
//					],
//					"return": [
//						"u64"
//					]
//				}
//			],
//			"structs": [
//				{
//					"name": "Coin",
//					"is_native": false,
//					"abilities": [
//						"store"
//					],
//					"generic_type_params": [
//						{
//							"constraints": []
//						}
//					],
//					"fields": [
//						{
//							"name": "value",
//							"type": "u64"
//						}
//					]
//				}
//			]
//		}
//	}
type MoveScript struct {
	Bytecode HexBytes      `json:"bytecode"`      // Bytecode is the hex encoded version of the compiled script
	Abi      *MoveFunction `json:"abi,omitempty"` // Abi is the ABI for the module, and is optional
}

// MoveFunction describes a move function and its associated properties
//
// Example 0x1::coin::balance:
//
//	{
//		"name": "balance",
//		"visibility": "public",
//		"is_entry": false,
//		"is_view": true,
//		"generic_type_params": [
//			{
//				"constraints": []
//			}
//		],
//		"params": [
//			"address"
//		],
//		"return": [
//			"u64"
//		]
//	}
type MoveFunction struct {
	Name              MoveComponentId     `json:"name"`                // Name is the name of the function e.g. balance
	Visibility        MoveVisibility      `json:"visibility"`          // Visibility is the visibility of the function e.g. public
	IsEntry           bool                `json:"is_entry"`            // IsEntry is true if the function is an entry function
	IsView            bool                `json:"is_view"`             // IsView is true if the function is a view function
	GenericTypeParams []*GenericTypeParam `json:"generic_type_params"` // GenericTypeParams are the generic type parameters for the function
	Params            []string            `json:"params"`              // Params are the parameters for the function in string format for the TypeTag
	Return            []string            `json:"return"`              // Return is the return type for the function in string format for the TypeTag
}

// GenericTypeParam is a set of requirements for a generic.  These can be applied via different
// MoveAbility constraints required on the type
//
// Example:
//
//	{
//		"constraints": [
//			"copy"
//		]
//	}
type GenericTypeParam struct {
	Constraints []MoveAbility `json:"constraints"` // Constraints are the constraints required for the generic type e.g. copy
}

// MoveAbility are the types of abilities applied to structs, the possible types are listed
// as MoveAbilityStore and others
type MoveAbility string

const (
	MoveAbilityStore MoveAbility = "store"
	MoveAbilityDrop  MoveAbility = "drop"
	MoveAbilityKey   MoveAbility = "key"
	MoveAbilityCopy  MoveAbility = "copy"
)

// MoveVisibility is the visibility of a function or struct, the possible types are listed
// as MoveVisibilityPublic and others
type MoveVisibility string

const (
	MoveVisibilityPublic  MoveVisibility = "public"
	MoveVisibilityPrivate MoveVisibility = "private"
	MoveVisibilityFriend  MoveVisibility = "friend"
)

// MoveStruct describes the layout for a struct, and its constraints
//
// Example 0x1::coin::Coin:
//
//	{
//		"name": "Coin",
//		"is_native": false,
//		"abilities": [
//			"store"
//		],
//		"generic_type_params": [
//			{
//				"constraints": []
//			}
//		],
//		"fields": [
//			{
//				"name": "value",
//				"type": "u64"
//			}
//		]
//	}
type MoveStruct struct {
	Name              string              `json:"name"`                // Name is the name of the struct e.g. Coin
	IsNative          bool                `json:"is_native"`           // IsNative is true if the struct is native e.g. u64
	Abilities         []MoveAbility       `json:"abilities"`           // Abilities are the abilities applied to the struct e.g. copy or store
	GenericTypeParams []*GenericTypeParam `json:"generic_type_params"` // GenericTypeParams are the generic type parameters for the struct
	Fields            []*MoveStructField  `json:"fields"`              // Fields are the fields in the struct
}

// MoveStructField represents a single field in a struct, and it's associated type
//
// Example:
//
//	{
//		"name": "value",
//		"type": "u64"
//	}
type MoveStructField struct {
	Name string `json:"name"` // Name of the field e.g. value
	Type string `json:"type"` // Type of the field in string format for the TypeTag e.g. u64
}
