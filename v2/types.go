package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/types"
)

// AccountAddress is a 32-byte address on the Aptos blockchain.
// It can represent an account, an object, or other on-chain entities.
type AccountAddress = types.AccountAddress

// ParseAddress parses a hex string into an AccountAddress.
// It accepts addresses with or without the "0x" prefix, and handles
// both short (special) addresses like "0x1" and full 64-character addresses.
var ParseAddress = types.ParseAddress

// MustParseAddress parses a hex string into an AccountAddress, panicking on error.
// This is useful for compile-time constants.
var MustParseAddress = types.MustParseAddress

// TypeTag re-exports from internal/types

// TypeTag is the concrete wrapper for Move type representations.
// It wraps a TypeTagImpl and handles BCS serialization.
// Use the helper functions (NewTypeTag, ParseTypeTag, etc.) to create TypeTags.
type TypeTag = types.TypeTag

// TypeTagImpl is the interface implemented by all TypeTag variants.
type TypeTagImpl = types.TypeTagImpl

// TypeTagVariant is the BCS discriminant for TypeTag variants.
type TypeTagVariant = types.TypeTagVariant

// Type tag variant constants
const (
	TypeTagBool      = types.TypeTagBool
	TypeTagU8        = types.TypeTagU8
	TypeTagU64       = types.TypeTagU64
	TypeTagU128      = types.TypeTagU128
	TypeTagAddress   = types.TypeTagAddress
	TypeTagSigner    = types.TypeTagSigner
	TypeTagVector    = types.TypeTagVector
	TypeTagStruct    = types.TypeTagStruct
	TypeTagU16       = types.TypeTagU16
	TypeTagU32       = types.TypeTagU32
	TypeTagU256      = types.TypeTagU256
	TypeTagGeneric   = types.TypeTagGeneric
	TypeTagReference = types.TypeTagReference
)

// Primitive type tags
type (
	BoolTag    = types.BoolTag
	U8Tag      = types.U8Tag
	U16Tag     = types.U16Tag
	U32Tag     = types.U32Tag
	U64Tag     = types.U64Tag
	U128Tag    = types.U128Tag
	U256Tag    = types.U256Tag
	AddressTag = types.AddressTag
	SignerTag  = types.SignerTag
)

// Compound type tags
type (
	VectorTag    = types.VectorTag
	StructTag    = types.StructTag
	ReferenceTag = types.ReferenceTag
	GenericTag   = types.GenericTag
)

// NewTypeTag creates a new TypeTag wrapping the given implementation.
var NewTypeTag = types.NewTypeTag

// NewVectorTag creates a vector<T> TypeTag.
var NewVectorTag = types.NewVectorTag

// NewStringTag creates a 0x1::string::String TypeTag.
var NewStringTag = types.NewStringTag

// NewOptionTag creates a 0x1::option::Option<T> TypeTag.
var NewOptionTag = types.NewOptionTag

// NewObjectTag creates a 0x1::object::Object<T> TypeTag.
var NewObjectTag = types.NewObjectTag

// ParseTypeTag parses a Move type string into a TypeTag.
var ParseTypeTag = types.ParseTypeTag

// AptosCoinTypeTag is the TypeTag for 0x1::aptos_coin::AptosCoin.
var AptosCoinTypeTag = types.AptosCoinTypeTag

// ModuleID identifies a Move module by its address and name.
type ModuleID struct {
	Address AccountAddress
	Name    string
}

// String returns the canonical string representation of the module ID.
func (m ModuleID) String() string {
	return m.Address.String() + "::" + m.Name
}

// AccountInfo contains information about an on-chain account.
type AccountInfo struct {
	SequenceNumber    uint64 `json:"sequence_number,string"`
	AuthenticationKey string `json:"authentication_key"`
}

// Resource represents an on-chain Move resource.
type Resource struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// NodeInfo contains information about an Aptos node.
type NodeInfo struct {
	ChainID             uint8  `json:"chain_id"`
	Epoch               uint64 `json:"epoch,string"`
	LedgerVersion       uint64 `json:"ledger_version,string"`
	OldestLedgerVersion uint64 `json:"oldest_ledger_version,string"`
	LedgerTimestamp     uint64 `json:"ledger_timestamp,string"`
	NodeRole            string `json:"node_role"`
	OldestBlockHeight   uint64 `json:"oldest_block_height,string"`
	BlockHeight         uint64 `json:"block_height,string"`
	GitHash             string `json:"git_hash"`
}

// GasEstimate contains gas price estimates from the network.
type GasEstimate struct {
	GasEstimate              uint64 `json:"gas_estimate"`
	DeprioritizedGasEstimate uint64 `json:"deprioritized_gas_estimate"`
	PrioritizedGasEstimate   uint64 `json:"prioritized_gas_estimate"`
}

// SubmitResult contains the result of a transaction submission.
type SubmitResult struct {
	Hash string `json:"hash"`
}

// SimulationResult contains the result of a transaction simulation.
type SimulationResult struct {
	Success      bool     `json:"success"`
	VMStatus     string   `json:"vm_status"`
	GasUsed      uint64   `json:"gas_used,string"`
	GasUnitPrice uint64   `json:"gas_unit_price,string"`
	Changes      []Change `json:"changes"`
	Events       []Event  `json:"events"`
}

// Change represents a state change from a transaction.
type Change struct {
	Type         string `json:"type"`
	Address      string `json:"address,omitempty"`
	StateKeyHash string `json:"state_key_hash,omitempty"`
	Data         any    `json:"data,omitempty"`
}

// Event represents an event emitted by a transaction.
type Event struct {
	GUID           EventGUID `json:"guid"`
	SequenceNumber uint64    `json:"sequence_number,string"`
	Type           string    `json:"type"`
	Data           any       `json:"data"`
}

// EventGUID identifies an event stream.
type EventGUID struct {
	CreationNumber uint64 `json:"creation_number,string"`
	AccountAddress string `json:"account_address"`
}

// Block represents a block on the Aptos blockchain.
type Block struct {
	BlockHeight    uint64        `json:"block_height,string"`
	BlockHash      string        `json:"block_hash"`
	BlockTimestamp uint64        `json:"block_timestamp,string"`
	FirstVersion   uint64        `json:"first_version,string"`
	LastVersion    uint64        `json:"last_version,string"`
	Transactions   []Transaction `json:"transactions,omitempty"`
}

// Transaction represents a transaction on the Aptos blockchain.
type Transaction struct {
	Type                    string   `json:"type"`
	Version                 uint64   `json:"version,string"`
	Hash                    string   `json:"hash"`
	StateChangeHash         string   `json:"state_change_hash"`
	EventRootHash           string   `json:"event_root_hash"`
	StateCheckpointHash     *string  `json:"state_checkpoint_hash"`
	GasUsed                 uint64   `json:"gas_used,string"`
	Success                 bool     `json:"success"`
	VMStatus                string   `json:"vm_status"`
	AccumulatorRootHash     string   `json:"accumulator_root_hash"`
	Timestamp               uint64   `json:"timestamp,string"`
	Sender                  string   `json:"sender,omitempty"`
	SequenceNumber          uint64   `json:"sequence_number,string,omitempty"`
	MaxGasAmount            uint64   `json:"max_gas_amount,string,omitempty"`
	GasUnitPrice            uint64   `json:"gas_unit_price,string,omitempty"`
	ExpirationTimestampSecs uint64   `json:"expiration_timestamp_secs,string,omitempty"`
	Payload                 any      `json:"payload,omitempty"`
	Signature               any      `json:"signature,omitempty"`
	Events                  []Event  `json:"events,omitempty"`
	Changes                 []Change `json:"changes,omitempty"`
}

// BatchSubmitResult contains results from batch transaction submission.
type BatchSubmitResult struct {
	// TransactionFailures contains any failures during batch submission.
	// Index corresponds to the position in the original batch.
	TransactionFailures []BatchSubmitFailure `json:"transaction_failures,omitempty"`
}

// BatchSubmitFailure represents a failure in batch transaction submission.
type BatchSubmitFailure struct {
	Error struct {
		Message   string `json:"message"`
		ErrorCode string `json:"error_code"`
		VMStatus  string `json:"vm_error_code,omitempty"`
	} `json:"error"`
	TransactionIndex int `json:"transaction_index"`
}

// HealthCheckResponse contains the result of a health check.
type HealthCheckResponse struct {
	Message string `json:"message"`
}

// ModuleBytecode contains the bytecode and ABI for a module.
type ModuleBytecode struct {
	Bytecode string     `json:"bytecode"`
	ABI      *ModuleABI `json:"abi,omitempty"`
}

// ModuleABI contains the ABI for a module.
type ModuleABI struct {
	Address          string               `json:"address"`
	Name             string               `json:"name"`
	Friends          []string             `json:"friends"`
	ExposedFunctions []ExposedFunctionABI `json:"exposed_functions"`
	Structs          []StructABI          `json:"structs"`
}

// ExposedFunctionABI contains the ABI for an exposed function.
type ExposedFunctionABI struct {
	Name              string   `json:"name"`
	Visibility        string   `json:"visibility"`
	IsEntry           bool     `json:"is_entry"`
	IsView            bool     `json:"is_view"`
	GenericTypeParams []any    `json:"generic_type_params"`
	Params            []string `json:"params"`
	Return            []string `json:"return"`
}

// StructABI contains the ABI for a struct.
type StructABI struct {
	Name              string     `json:"name"`
	IsNative          bool       `json:"is_native"`
	Abilities         []string   `json:"abilities"`
	GenericTypeParams []any      `json:"generic_type_params"`
	Fields            []FieldABI `json:"fields"`
}

// FieldABI contains the ABI for a struct field.
type FieldABI struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
