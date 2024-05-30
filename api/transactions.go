package api

import (
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

const (
	EnumPendingTransaction         = "pending_transaction"
	EnumUserTransaction            = "user_transaction"
	EnumGenesisTransaction         = "genesis_transaction"
	EnumBlockMetadataTransaction   = "block_metadata_transaction"
	EnumStateCheckpointTransaction = "state_checkpoint_transaction"
	EnumValidatorTransaction       = "validator_transaction"
)

// Transaction is an enum type for all possible transactions on the blockchain
type Transaction struct {
	Type  string
	Inner TransactionImpl
}

func (o *Transaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Type, err = toString(data, "type")
	if err != nil {
		return err
	}
	switch o.Type {
	case EnumPendingTransaction:
		o.Inner = &PendingTransaction{}
	case EnumUserTransaction:
		o.Inner = &UserTransaction{}
	case EnumGenesisTransaction:
		o.Inner = &GenesisTransaction{}
	case EnumBlockMetadataTransaction:
		o.Inner = &BlockMetadataTransaction{}
	case EnumStateCheckpointTransaction:
		o.Inner = &StateCheckpointTransaction{}
	case EnumValidatorTransaction:
		o.Inner = &ValidatorTransaction{}
	default:
		return fmt.Errorf("unknown transaction type: %s", o.Type)
	}
	return o.Inner.UnmarshalJSONFromMap(data)
}

type TransactionImpl interface {
	UnmarshalFromMap
}

// UserTransaction is a user submitted transaction as an entry function or script
type UserTransaction struct {
	Version                 uint64
	Hash                    string
	AccumulatorRootHash     string
	StateChangeHash         string
	EventRootHash           string
	GasUsed                 uint64
	Success                 bool
	VmStatus                string
	Changes                 []*WriteSetChange
	Events                  []*Event
	Sender                  *types.AccountAddress
	SequenceNumber          uint64
	MaxGasAmount            uint64
	GasUnitPrice            uint64
	ExpirationTimestampSecs uint64
	Payload                 *TransactionPayload
	Signature               *Signature
	Timestamp               uint64 // TODO: native time?
	// TODO: State checkpoint hash is optional
}

func (o *UserTransaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	info := &TransactionInfo{}
	err = info.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}

	o.Version = info.Version
	o.Hash = info.Hash
	o.AccumulatorRootHash = info.AccumulatorRootHash
	o.StateChangeHash = info.StateChangeHash
	o.EventRootHash = info.EventRootHash
	o.GasUsed = info.GasUsed
	o.Success = info.Success
	o.VmStatus = info.VmStatus
	o.Changes = info.Changes
	o.Events = info.Events

	userInfo := &UserTransactionInfo{}
	err = userInfo.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}
	o.Sender = userInfo.Sender
	o.SequenceNumber = userInfo.SequenceNumber
	o.MaxGasAmount = userInfo.MaxGasAmount
	o.GasUnitPrice = userInfo.GasUnitPrice
	o.ExpirationTimestampSecs = userInfo.ExpirationTimestampSecs
	o.Payload = userInfo.Payload
	o.Signature = userInfo.Signature

	o.Timestamp, err = toUint64(data, "timestamp")
	return err
}

type PendingTransaction struct {
	Hash                    string
	Sender                  *types.AccountAddress
	SequenceNumber          uint64
	MaxGasAmount            uint64
	GasUnitPrice            uint64
	ExpirationTimestampSecs uint64
	Payload                 *TransactionPayload
	Signature               *Signature
}

func (o *PendingTransaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Hash, err = toHash(data, "hash")
	if err != nil {
		return err
	}
	userInfo := &UserTransactionInfo{}
	err = userInfo.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}
	o.Sender = userInfo.Sender
	o.SequenceNumber = userInfo.SequenceNumber
	o.MaxGasAmount = userInfo.MaxGasAmount
	o.GasUnitPrice = userInfo.GasUnitPrice
	o.ExpirationTimestampSecs = userInfo.ExpirationTimestampSecs
	o.Payload = userInfo.Payload
	o.Signature = userInfo.Signature
	return nil
}

type GenesisTransaction struct {
	Version             uint64
	Hash                string
	AccumulatorRootHash string
	StateChangeHash     string
	EventRootHash       string
	GasUsed             uint64
	Success             bool
	VmStatus            string
	Changes             []*WriteSetChange
	Events              []*Event
	Payload             *TransactionPayload
}

func (o *GenesisTransaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	info := &TransactionInfo{}
	err = info.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}

	o.Version = info.Version
	o.Hash = info.Hash
	o.AccumulatorRootHash = info.AccumulatorRootHash
	o.StateChangeHash = info.StateChangeHash
	o.EventRootHash = info.EventRootHash
	o.GasUsed = info.GasUsed
	o.Success = info.Success
	o.VmStatus = info.VmStatus
	o.Changes = info.Changes
	o.Events = info.Events

	o.Payload, err = toPayload(data, "payload")
	return err
}

type BlockMetadataTransaction struct {
	Id                       string
	Epoch                    uint64
	Round                    uint64
	PreviousBlockVotesBitvec []uint8
	Proposer                 *types.AccountAddress
	FailedProposerIndices    []uint32
	Version                  uint64
	Hash                     string
	AccumulatorRootHash      string
	StateChangeHash          string
	EventRootHash            string
	GasUsed                  uint64
	Success                  bool
	VmStatus                 string
	Changes                  []*WriteSetChange
	Events                   []*Event
	Timestamp                uint64
	// TODO State checkpoint hash is optional
}

func (o *BlockMetadataTransaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Id, err = toString(data, "id")
	if err != nil {
		return err
	}
	o.Epoch, err = toUint64(data, "epoch")
	if err != nil {
		return err
	}
	o.Round, err = toUint64(data, "round")
	if err != nil {
		return err
	}
	o.PreviousBlockVotesBitvec, err = toUint8Array(data, "previous_block_votes_bitvec")
	if err != nil {
		return err
	}
	o.Proposer, err = toAccountAddress(data, "proposer")
	if err != nil {
		return err
	}
	o.FailedProposerIndices, err = toUint32Array(data, "failed_proposer_indices")
	if err != nil {
		return err
	}

	info := &TransactionInfo{}
	err = info.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}

	o.Version = info.Version
	o.Hash = info.Hash
	o.AccumulatorRootHash = info.AccumulatorRootHash
	o.StateChangeHash = info.StateChangeHash
	o.EventRootHash = info.EventRootHash
	o.GasUsed = info.GasUsed
	o.Success = info.Success
	o.VmStatus = info.VmStatus
	o.Changes = info.Changes
	o.Events = info.Events

	o.Timestamp, err = toUint64(data, "timestamp")
	return err
}

type StateCheckpointTransaction struct {
	Version             uint64
	Hash                string
	AccumulatorRootHash string
	StateChangeHash     string
	EventRootHash       string
	GasUsed             uint64
	Success             bool
	VmStatus            string
	Changes             []*WriteSetChange
	Timestamp           uint64
	StateCheckpointHash string // This is optional
}

func (o *StateCheckpointTransaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	info := &TransactionInfo{}
	err = info.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}

	o.Version = info.Version
	o.Hash = info.Hash
	o.AccumulatorRootHash = info.AccumulatorRootHash
	o.StateChangeHash = info.StateChangeHash
	o.EventRootHash = info.EventRootHash
	o.GasUsed = info.GasUsed
	o.Success = info.Success
	o.VmStatus = info.VmStatus
	o.Changes = info.Changes

	o.Timestamp, err = toUint64(data, "timestamp")
	if err != nil {
		return err
	}
	// Optional Fields
	stateCheckpointHash, ok := data["state_checkpoint_hash"].(string)
	if ok {
		o.StateCheckpointHash = stateCheckpointHash
	}
	return nil
}

type ValidatorTransaction struct {
	Version             uint64
	Hash                string
	AccumulatorRootHash string
	StateChangeHash     string
	EventRootHash       string
	GasUsed             uint64
	Success             bool
	VmStatus            string
	Changes             []*WriteSetChange
	Events              []*Event
	Timestamp           uint64
	// TODO: StateCheckpointHash is optional
}

func (o *ValidatorTransaction) UnmarshalJSONFromMap(data map[string]any) (err error) {
	info := &TransactionInfo{}
	err = info.UnmarshalJSONFromMap(data)
	if err != nil {
		return err
	}

	o.Version = info.Version
	o.Hash = info.Hash
	o.AccumulatorRootHash = info.AccumulatorRootHash
	o.StateChangeHash = info.StateChangeHash
	o.EventRootHash = info.EventRootHash
	o.GasUsed = info.GasUsed
	o.Success = info.Success
	o.VmStatus = info.VmStatus
	o.Changes = info.Changes
	o.Events = info.Events

	o.Timestamp, err = toUint64(data, "timestamp")
	return err
}

func (o *Transaction) UnmarshalJSON(b []byte) error {
	var data map[string]any
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	return o.UnmarshalJSONFromMap(data)
}

type TransactionInfo struct {
	Version             uint64
	Hash                string
	AccumulatorRootHash string
	StateChangeHash     string
	EventRootHash       string
	GasUsed             uint64
	Success             bool
	VmStatus            string
	Changes             []*WriteSetChange
	Events              []*Event
}

func (o *TransactionInfo) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Version, err = toUint64(data, "version")
	if err != nil {
		return err
	}
	o.Hash, err = toHash(data, "hash")
	if err != nil {
		return err
	}
	o.AccumulatorRootHash, err = toHash(data, "accumulator_root_hash")
	if err != nil {
		return err
	}
	o.StateChangeHash, err = toHash(data, "state_change_hash")
	if err != nil {
		return err
	}
	o.EventRootHash, err = toHash(data, "event_root_hash")
	if err != nil {
		return err
	}
	o.GasUsed, err = toUint64(data, "gas_used")
	if err != nil {
		return err
	}
	o.Success, err = toBool(data, "success")
	if err != nil {
		return err
	}
	o.VmStatus, err = toString(data, "vm_status")
	if err != nil {
		return err
	}
	o.Changes, err = toWriteSetChanges(data, "changes")
	if err != nil {
		return err
	}
	o.Events, err = toEvents(data, "events")
	return err
}

type UserTransactionInfo struct {
	SequenceNumber          uint64
	MaxGasAmount            uint64
	GasUnitPrice            uint64
	ExpirationTimestampSecs uint64
	Sender                  *types.AccountAddress
	Payload                 *TransactionPayload
	Signature               *Signature
}

func (o *UserTransactionInfo) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.SequenceNumber, err = toUint64(data, "sequence_number")
	if err != nil {
		return err
	}
	o.MaxGasAmount, err = toUint64(data, "max_gas_amount")
	if err != nil {
		return err
	}
	o.GasUnitPrice, err = toUint64(data, "gas_unit_price")
	if err != nil {
		return err
	}
	o.ExpirationTimestampSecs, err = toUint64(data, "expiration_timestamp_secs")
	if err != nil {
		return err
	}
	o.Sender, err = toAccountAddress(data, "sender")
	if err != nil {
		return err
	}
	o.Payload, err = toPayload(data, "payload")
	if err != nil {
		return err
	}
	o.Signature, err = toSignature(data, "signature")

	return err
}
