package api

import (
	"encoding/json"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// TransactionVariant is the type of transaction, all transactions submitted by this SDK are [TransactionVariantUser]
type TransactionVariant string

const (
	TransactionVariantPending         TransactionVariant = "pending_transaction"          // TransactionVariantPending maps to PendingTransaction
	TransactionVariantUser            TransactionVariant = "user_transaction"             // TransactionVariantUser maps to UserTransaction
	TransactionVariantGenesis         TransactionVariant = "genesis_transaction"          // TransactionVariantGenesis maps to GenesisTransaction
	TransactionVariantBlockMetadata   TransactionVariant = "block_metadata_transaction"   // TransactionVariantBlockMetadata maps to BlockMetadataTransaction
	TransactionVariantBlockEpilogue   TransactionVariant = "block_epilogue_transaction"   // TransactionVariantBlockEpilogue maps to BlockEpilogueTransaction
	TransactionVariantStateCheckpoint TransactionVariant = "state_checkpoint_transaction" // TransactionVariantStateCheckpoint maps to StateCheckpointTransaction
	TransactionVariantValidator       TransactionVariant = "validator_transaction"        // TransactionVariantValidator maps to ValidatorTransaction
	TransactionVariantUnknown         TransactionVariant = "unknown"                      // TransactionVariantUnknown maps to UnknownTransaction for unknown types
)

// CommittedTransaction is an enum type for all possible committed transactions on the blockchain
// This is the same as [Transaction] but with the Success and Version functions always confirmed.
type CommittedTransaction struct {
	Type  TransactionVariant // Type of the transaction
	Inner TransactionImpl    // Inner is the actual transaction
}

// Hash of the transaction for lookup on-chain
func (o *CommittedTransaction) Hash() Hash {
	return o.Inner.TxnHash()
}

// Success of the transaction.  Pending transactions, and genesis may not have a success field.
// If this is the case, it will be nil
func (o *CommittedTransaction) Success() bool {
	return *o.Inner.TxnSuccess()
}

// Version of the transaction on chain, will be nil if it is a PendingTransaction
func (o *CommittedTransaction) Version() uint64 {
	return *o.Inner.TxnVersion()
}

// UnmarshalJSON unmarshals the [Transaction] from JSON handling conversion between types
func (o *CommittedTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = TransactionVariant(data.Type)
	switch o.Type {
	case TransactionVariantPending:
		return fmt.Errorf("transaction type is not committed: %s, this is unexpected for the API to return", o.Type)
	case TransactionVariantUser:
		o.Inner = &UserTransaction{}
	case TransactionVariantGenesis:
		o.Inner = &GenesisTransaction{}
	case TransactionVariantBlockMetadata:
		o.Inner = &BlockMetadataTransaction{}
	case TransactionVariantBlockEpilogue:
		o.Inner = &BlockEpilogueTransaction{}
	case TransactionVariantStateCheckpoint:
		o.Inner = &StateCheckpointTransaction{}
	case TransactionVariantValidator:
		o.Inner = &ValidatorTransaction{}
	default:
		sig := &UnknownTransaction{Type: string(o.Type)}
		o.Inner = sig
		o.Type = TransactionVariantUnknown
		return json.Unmarshal(b, &sig.Payload)
	}
	return json.Unmarshal(b, o.Inner)
}

// UserTransaction changes the transaction to a [UserTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) UserTransaction() (*UserTransaction, error) {
	switch inner := o.Inner.(type) {
	case *UserTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not user: %s", o.Type)
	}
}

// GenesisTransaction changes the transaction to a [GenesisTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) GenesisTransaction() (*GenesisTransaction, error) {
	switch inner := o.Inner.(type) {
	case *GenesisTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not genesis: %s", o.Type)
	}
}

// BlockMetadataTransaction changes the transaction to a [BlockMetadataTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) BlockMetadataTransaction() (*BlockMetadataTransaction, error) {
	switch inner := o.Inner.(type) {
	case *BlockMetadataTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not block_metadata: %s", o.Type)
	}
}

// BlockEpilogueTransaction changes the transaction to a [BlockEpilogueTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) BlockEpilogueTransaction() (*BlockEpilogueTransaction, error) {
	switch inner := o.Inner.(type) {
	case *BlockEpilogueTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not block_epilogue: %s", o.Type)
	}
}

// StateCheckpointTransaction changes the transaction to a [StateCheckpointTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) StateCheckpointTransaction() (*StateCheckpointTransaction, error) {
	switch inner := o.Inner.(type) {
	case *StateCheckpointTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not state_checkpoint: %s", o.Type)
	}
}

// ValidatorTransaction changes the transaction to a [ValidatorTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) ValidatorTransaction() (*ValidatorTransaction, error) {
	switch inner := o.Inner.(type) {
	case *ValidatorTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not validator: %s", o.Type)
	}
}

// UnknownTransaction changes the transaction to a [UnknownTransaction]; however, it will fail if it's not one.
func (o *CommittedTransaction) UnknownTransaction() (*UnknownTransaction, error) {
	switch inner := o.Inner.(type) {
	case *UnknownTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not unknown: %s", o.Type)
	}
}

// Transaction is an enum type for all possible transactions on the blockchain
type Transaction struct {
	Type  TransactionVariant // Type of the transaction
	Inner TransactionImpl    // Inner is the actual transaction
}

// Hash of the transaction for lookup on-chain
func (o *Transaction) Hash() Hash {
	return o.Inner.TxnHash()
}

// Success of the transaction.  Pending transactions, and genesis may not have a success field.
// If this is the case, it will be nil
func (o *Transaction) Success() *bool {
	return o.Inner.TxnSuccess()
}

// Version of the transaction on chain, will be nil if it is a PendingTransaction
func (o *Transaction) Version() *uint64 {
	return o.Inner.TxnVersion()
}

// UnmarshalJSON unmarshals the [Transaction] from JSON handling conversion between types
func (o *Transaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type string `json:"type"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = TransactionVariant(data.Type)
	switch o.Type {
	case TransactionVariantPending:
		o.Inner = &PendingTransaction{}
	case TransactionVariantUser:
		o.Inner = &UserTransaction{}
	case TransactionVariantGenesis:
		o.Inner = &GenesisTransaction{}
	case TransactionVariantBlockMetadata:
		o.Inner = &BlockMetadataTransaction{}
	case TransactionVariantBlockEpilogue:
		o.Inner = &BlockEpilogueTransaction{}
	case TransactionVariantStateCheckpoint:
		o.Inner = &StateCheckpointTransaction{}
	case TransactionVariantValidator:
		o.Inner = &ValidatorTransaction{}
	default:
		txn := &UnknownTransaction{Type: string(o.Type)}
		o.Inner = txn
		o.Type = TransactionVariantUnknown
		return json.Unmarshal(b, &txn.Payload)
	}
	return json.Unmarshal(b, o.Inner)
}

// UserTransaction changes the transaction to a [UserTransaction]; however, it will fail if it's not one.
func (o *Transaction) UserTransaction() (*UserTransaction, error) {
	switch inner := o.Inner.(type) {
	case *UserTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not user: %s", o.Type)
	}
}

// PendingTransaction changes the transaction to a [PendingTransaction]; however, it will fail if it's not one.
func (o *Transaction) PendingTransaction() (*PendingTransaction, error) {
	switch inner := o.Inner.(type) {
	case *PendingTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not pending: %s", o.Type)
	}
}

// GenesisTransaction changes the transaction to a [GenesisTransaction]; however, it will fail if it's not one.
func (o *Transaction) GenesisTransaction() (*GenesisTransaction, error) {
	switch inner := o.Inner.(type) {
	case *GenesisTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not genesis: %s", o.Type)
	}
}

// BlockMetadataTransaction changes the transaction to a [BlockMetadataTransaction]; however, it will fail if it's not one.
func (o *Transaction) BlockMetadataTransaction() (*BlockMetadataTransaction, error) {
	switch inner := o.Inner.(type) {
	case *BlockMetadataTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not block_metadata: %s", o.Type)
	}
}

// BlockEpilogueTransaction changes the transaction to a [BlockEpilogueTransaction]; however, it will fail if it's not one.
func (o *Transaction) BlockEpilogueTransaction() (*BlockEpilogueTransaction, error) {
	switch inner := o.Inner.(type) {
	case *BlockEpilogueTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not block_epilogue: %s", o.Type)
	}
}

// StateCheckpointTransaction changes the transaction to a [StateCheckpointTransaction]; however, it will fail if it's not one.
func (o *Transaction) StateCheckpointTransaction() (*StateCheckpointTransaction, error) {
	switch inner := o.Inner.(type) {
	case *StateCheckpointTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not state_checkpoint: %s", o.Type)
	}
}

// ValidatorTransaction changes the transaction to a [ValidatorTransaction]; however, it will fail if it's not one.
func (o *Transaction) ValidatorTransaction() (*ValidatorTransaction, error) {
	switch inner := o.Inner.(type) {
	case *ValidatorTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not validator: %s", o.Type)
	}
}

// UnknownTransaction changes the transaction to a [UnknownTransaction]; however, it will fail if it's not one.
func (o *Transaction) UnknownTransaction() (*UnknownTransaction, error) {
	switch inner := o.Inner.(type) {
	case *UnknownTransaction:
		return inner, nil
	default:
		return nil, fmt.Errorf("transaction type is not unknown: %s", o.Type)
	}
}

// TransactionImpl is an interface for all transactions
type TransactionImpl interface {
	// TxnSuccess tells us if the transaction is a success.  It will be nil if the transaction is not committed.
	TxnSuccess() *bool

	// TxnHash gives us the hash of the transaction.
	TxnHash() Hash

	// TxnVersion gives us the ledger version of the transaction. It will be nil if the transaction is not committed.
	TxnVersion() *uint64
}

// UnknownTransaction is a transaction type that is not recognized by the SDK
type UnknownTransaction struct {
	Type    string         // Type is the type of the unknown transaction
	Payload map[string]any // Payload is the raw JSON payload
}

// TxnSuccess tells us if the transaction is a success.  It will be nil if the transaction is not committed.
func (u *UnknownTransaction) TxnSuccess() *bool {
	success := u.Payload["success"]
	if success == nil {
		return nil
	}
	successBool, ok := success.(bool)
	if !ok {
		return nil
	}
	return &successBool
}

// TxnHash gives us the hash of the transaction.
func (u *UnknownTransaction) TxnHash() Hash {
	if hash, ok := u.Payload["hash"].(string); ok {
		return hash
	}
	return ""
}

// TxnVersion gives us the ledger version of the transaction. It will be nil if the transaction is not committed.
func (u *UnknownTransaction) TxnVersion() *uint64 {
	versionStr, ok := u.Payload["version"].(string)
	if !ok {
		return nil
	}
	num, err := util.StrToUint64(versionStr)
	if err != nil {
		return nil
	}

	return &num
}

// UserTransaction is a user submitted transaction as an entry function, script, or more.
//
// These transactions are the only transactions submitted by users to the blockchain.
type UserTransaction struct {
	Version                 uint64                // Version of the transaction, starts at 0 and increments per transaction.
	Hash                    Hash                  // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	AccumulatorRootHash     Hash                  // AccumulatorRootHash of the transaction.
	StateChangeHash         Hash                  // StateChangeHash of the transaction.
	EventRootHash           Hash                  // EventRootHash of the transaction.
	GasUsed                 uint64                // GasUsed by the transaction, will be in gas units.
	Success                 bool                  // Success of the transaction.
	VmStatus                string                // VmStatus of the transaction, this will contain the error if any.
	Changes                 []*WriteSetChange     // Changes to the ledger from the transaction, should never be empty.
	Events                  []*Event              // Events emitted by the transaction, may be empty.
	Sender                  *types.AccountAddress // Sender of the transaction, will never be nil.
	SequenceNumber          uint64                // SequenceNumber of the transaction, starts at 0 and increments per transaction submitted by the sender.
	ReplayProtectionNonce   *uint64               // ReplayProtectionNonce of the transaction, if this value is non-null, then the sequence number does not apply.  Must be unique for 60 seconds for between user transactions for an addredss.
	MaxGasAmount            uint64                // MaxGasAmount of the transaction, this is the max amount of gas units that the user is willing to pay.
	GasUnitPrice            uint64                // GasUnitPrice of the transaction, this is the multiplier per unit of gas to tokens.
	ExpirationTimestampSecs uint64                // ExpirationTimestampSecs of the transaction, this is the Unix timestamp in seconds when the transaction expires.
	Payload                 *TransactionPayload   // Payload of the transaction, this is the actual transaction data.
	Signature               *Signature            // Signature is the AccountAuthenticator of the sender.
	Timestamp               uint64                // Timestamp is the Unix timestamp in microseconds when the block of the transaction was committed.
	StateCheckpointHash     Hash                  // StateCheckpointHash of the transaction. Optional, and will be "" if not set.
}

// TxnHash gives us the hash of the transaction.
func (o *UserTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will never be nil.
func (o *UserTransaction) TxnSuccess() *bool {
	return &o.Success
}

// TxnVersion gives us the ledger version of the transaction. It will never be nil.
func (o *UserTransaction) TxnVersion() *uint64 {
	return &o.Version
}

// UnmarshalJSON unmarshals the [UserTransaction] from JSON handling conversion between types
func (o *UserTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Version                 U64                   `json:"version"`
		Hash                    Hash                  `json:"hash"`
		AccumulatorRootHash     Hash                  `json:"accumulator_root_hash"`
		StateChangeHash         Hash                  `json:"state_change_hash"`
		EventRootHash           Hash                  `json:"event_root_hash"`
		GasUsed                 U64                   `json:"gas_used"`
		Success                 bool                  `json:"success"`
		VmStatus                string                `json:"vm_status"`
		Changes                 []*WriteSetChange     `json:"changes"`
		Events                  []*Event              `json:"events"`
		Sender                  *types.AccountAddress `json:"sender"`
		SequenceNumber          U64                   `json:"sequence_number"`
		ReplayProtectionNonce   *U64                  `json:"replay_protection_nonce"`
		MaxGasAmount            U64                   `json:"max_gas_amount"`
		GasUnitPrice            U64                   `json:"gas_unit_price"`
		ExpirationTimestampSecs U64                   `json:"expiration_timestamp_secs"`
		Payload                 *TransactionPayload   `json:"payload"`
		Signature               *Signature            `json:"signature"`
		Timestamp               U64                   `json:"timestamp"`
		StateCheckpointHash     Hash                  `json:"state_checkpoint_hash"` // Optional
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Version = data.Version.ToUint64()
	o.Hash = data.Hash
	o.AccumulatorRootHash = data.AccumulatorRootHash
	o.StateChangeHash = data.StateChangeHash
	o.EventRootHash = data.EventRootHash
	o.GasUsed = data.GasUsed.ToUint64()
	o.Success = data.Success
	o.VmStatus = data.VmStatus
	o.Changes = data.Changes
	o.Events = data.Events
	o.Sender = data.Sender
	o.SequenceNumber = data.SequenceNumber.ToUint64()
	o.MaxGasAmount = data.MaxGasAmount.ToUint64()
	o.GasUnitPrice = data.GasUnitPrice.ToUint64()
	o.ExpirationTimestampSecs = data.ExpirationTimestampSecs.ToUint64()
	o.Payload = data.Payload
	o.Signature = data.Signature
	o.Timestamp = data.Timestamp.ToUint64()
	o.StateCheckpointHash = data.StateCheckpointHash

	if data.ReplayProtectionNonce != nil {
		replayNonce := (data.ReplayProtectionNonce).ToUint64()
		o.ReplayProtectionNonce = &replayNonce
	}
	return nil
}

// PendingTransaction is a transaction that is not yet committed to the blockchain.
type PendingTransaction struct {
	Hash                    Hash                  // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	Sender                  *types.AccountAddress // Sender of the transaction, will never be nil.
	SequenceNumber          uint64                // SequenceNumber of the transaction, starts at 0 and increments per transaction submitted by the sender.
	ReplayProtectionNonce   *uint64               // ReplayProtectionNonce of the transaction, if this value is non-null, then the sequence number does not apply.  Must be unique for 60 seconds for between user transactions for an addredss.
	MaxGasAmount            uint64                // MaxGasAmount of the transaction, this is the max amount of gas units that the user is willing to pay.
	GasUnitPrice            uint64                // GasUnitPrice of the transaction, this is the multiplier per unit of gas to tokens.
	ExpirationTimestampSecs uint64                // ExpirationTimestampSecs of the transaction, this is the Unix timestamp in seconds when the transaction expires.
	Payload                 *TransactionPayload   // Payload of the transaction, this is the actual transaction data.
	Signature               *Signature            // Signature is the AccountAuthenticator of the sender.
}

// TxnHash gives us the hash of the transaction.
func (o *PendingTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will always be nil.
func (o *PendingTransaction) TxnSuccess() *bool {
	return nil
}

// TxnVersion gives us the ledger version of the transaction. It will always be nil.
func (o *PendingTransaction) TxnVersion() *uint64 {
	return nil
}

// UnmarshalJSON unmarshals the [PendingTransaction] from JSON handling conversion between types
func (o *PendingTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Hash                    Hash                  `json:"hash"`
		Sender                  *types.AccountAddress `json:"sender"`
		SequenceNumber          U64                   `json:"sequence_number"`
		ReplayProtectionNonce   *U64                  `json:"replay_protection_nonce"`
		MaxGasAmount            U64                   `json:"max_gas_amount"`
		GasUnitPrice            U64                   `json:"gas_unit_price"`
		ExpirationTimestampSecs U64                   `json:"expiration_timestamp_secs"`
		Payload                 *TransactionPayload   `json:"payload"`
		Signature               *Signature            `json:"signature"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Hash = data.Hash
	o.Sender = data.Sender
	o.SequenceNumber = data.SequenceNumber.ToUint64()
	o.MaxGasAmount = data.MaxGasAmount.ToUint64()
	o.GasUnitPrice = data.GasUnitPrice.ToUint64()
	o.ExpirationTimestampSecs = data.ExpirationTimestampSecs.ToUint64()
	o.Payload = data.Payload
	o.Signature = data.Signature

	if data.ReplayProtectionNonce != nil {
		replayNonce := data.ReplayProtectionNonce.ToUint64()
		o.ReplayProtectionNonce = &replayNonce
	}
	return nil
}

// GenesisTransaction is a transaction that is the first transaction on the blockchain.
type GenesisTransaction struct {
	Version             uint64              // Version of the transaction, starts at 0 and increments per transaction.
	Hash                Hash                // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	AccumulatorRootHash Hash                // AccumulatorRootHash of the transaction.
	StateChangeHash     Hash                // StateChangeHash of the transaction.
	EventRootHash       Hash                // EventRootHash of the transaction.
	GasUsed             uint64              // GasUsed by the transaction, will be in gas units.
	Success             bool                // Success of the transaction.
	VmStatus            string              // VmStatus of the transaction, this will contain the error if any.
	Changes             []*WriteSetChange   // Changes to the ledger from the transaction, should never be empty.
	Events              []*Event            // Events emitted by the transaction, may be empty.
	Payload             *TransactionPayload // Payload of the transaction, this is the actual transaction data.
	StateCheckpointHash Hash                // StateCheckpointHash of the transaction. Optional, and will be "" if not set.
}

// TxnHash gives us the hash of the transaction.
func (o *GenesisTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will never be nil.
func (o *GenesisTransaction) TxnSuccess() *bool {
	return &o.Success
}

// TxnVersion gives us the ledger version of the transaction. It will never be nil.
func (o *GenesisTransaction) TxnVersion() *uint64 {
	return &o.Version
}

// UnmarshalJSON unmarshals the [GenesisTransaction] from JSON handling conversion between types
func (o *GenesisTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Version             U64                 `json:"version"`
		Hash                Hash                `json:"hash"`
		AccumulatorRootHash Hash                `json:"accumulator_root_hash"`
		StateChangeHash     Hash                `json:"state_change_hash"`
		EventRootHash       Hash                `json:"event_root_hash"`
		GasUsed             U64                 `json:"gas_used"`
		Success             bool                `json:"success"`
		VmStatus            string              `json:"vm_status"`
		Changes             []*WriteSetChange   `json:"changes"`
		Events              []*Event            `json:"events"`
		Payload             *TransactionPayload `json:"payload"`
		StateCheckpointHash Hash                `json:"state_checkpoint_hash"` // Optional
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Version = data.Version.ToUint64()
	o.Hash = data.Hash
	o.AccumulatorRootHash = data.AccumulatorRootHash
	o.StateChangeHash = data.StateChangeHash
	o.EventRootHash = data.EventRootHash
	o.GasUsed = data.GasUsed.ToUint64()
	o.Success = data.Success
	o.VmStatus = data.VmStatus
	o.Changes = data.Changes
	o.Events = data.Events
	o.Payload = data.Payload
	o.StateCheckpointHash = data.StateCheckpointHash
	return nil
}

// BlockMetadataTransaction is a transaction that is metadata about a block.
type BlockMetadataTransaction struct {
	Id                       string                // Id of the block, is the Hash of the block.
	Epoch                    uint64                // Epoch of the block, starts at 0 and increments per epoch.  Epoch is roughly 2 hours, and subject to change.
	Round                    uint64                // Round of the block, starts at 0 and increments per round in the epoch.
	PreviousBlockVotesBitvec []uint8               // PreviousBlockVotesBitvec of the block, this is a bit vector of the votes of the previous block.
	Proposer                 *types.AccountAddress // Proposer of the block, will never be nil.
	FailedProposerIndices    []uint32              // FailedProposerIndices of the block, this is the indices of the proposers that failed to propose a block.
	Version                  uint64                // Version of the transaction, starts at 0 and increments per transaction.
	Hash                     string                // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	AccumulatorRootHash      Hash                  // AccumulatorRootHash of the transaction.
	StateChangeHash          Hash                  // StateChangeHash of the transaction.
	EventRootHash            Hash                  // EventRootHash of the transaction.
	GasUsed                  uint64                // GasUsed by the transaction, will be in gas units. Should always be 0.
	Success                  bool                  // Success of the transaction.
	VmStatus                 string                // VmStatus of the transaction, this will contain the error if any.
	Changes                  []*WriteSetChange     // Changes to the ledger from the transaction, should never be empty.
	Events                   []*Event              // Events emitted by the transaction, may be empty.
	Timestamp                uint64                // Timestamp is the Unix timestamp in microseconds when the block of the transaction was committed.
	StateCheckpointHash      Hash                  // StateCheckpointHash of the transaction. Optional, and will be "" if not set.
}

// TxnHash gives us the hash of the transaction.
func (o *BlockMetadataTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will never be nil.
func (o *BlockMetadataTransaction) TxnSuccess() *bool {
	return &o.Success
}

// TxnVersion gives us the ledger version of the transaction. It will never be nil.
func (o *BlockMetadataTransaction) TxnVersion() *uint64 {
	return &o.Version
}

// UnmarshalJSON unmarshals the [BlockMetadataTransaction] from JSON handling conversion between types
func (o *BlockMetadataTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Id                       string                `json:"id"`
		Epoch                    U64                   `json:"epoch"`
		Round                    U64                   `json:"round"`
		PreviousBlockVotesBitvec []byte                `json:"previous_block_votes_bitvec"` // TODO: this had to be float64 earlier
		Proposer                 *types.AccountAddress `json:"proposer"`
		FailedProposerIndices    []uint32              `json:"failed_proposer_indices"` // TODO: verify
		Version                  U64                   `json:"version"`
		Hash                     Hash                  `json:"hash"`
		AccumulatorRootHash      Hash                  `json:"accumulator_root_hash"`
		StateChangeHash          Hash                  `json:"state_change_hash"`
		EventRootHash            Hash                  `json:"event_root_hash"`
		GasUsed                  U64                   `json:"gas_used"`
		Success                  bool                  `json:"success"`
		VmStatus                 string                `json:"vm_status"`
		Changes                  []*WriteSetChange     `json:"changes"`
		Events                   []*Event              `json:"events"`
		Timestamp                U64                   `json:"timestamp"`
		StateCheckpointHash      Hash                  `json:"state_checkpoint_hash,omitempty"` // Optional
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	o.Id = data.Id
	o.Epoch = data.Epoch.ToUint64()
	o.Round = data.Round.ToUint64()
	o.PreviousBlockVotesBitvec = data.PreviousBlockVotesBitvec
	o.Proposer = data.Proposer
	o.FailedProposerIndices = data.FailedProposerIndices
	o.Version = data.Version.ToUint64()
	o.Hash = data.Hash
	o.AccumulatorRootHash = data.AccumulatorRootHash
	o.StateChangeHash = data.StateChangeHash
	o.EventRootHash = data.EventRootHash
	o.GasUsed = data.GasUsed.ToUint64()
	o.Success = data.Success
	o.VmStatus = data.VmStatus
	o.Changes = data.Changes
	o.Events = data.Events
	o.Timestamp = data.Timestamp.ToUint64()
	o.StateCheckpointHash = data.StateCheckpointHash
	return nil
}

// BlockEpilogueTransaction is a transaction at the end of the block.  It is not necessarily at the end of a block prior to being enabled as a feature.
type BlockEpilogueTransaction struct {
	Version             uint64            // Version of the transaction, starts at 0 and increments per transaction.
	Hash                Hash              // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	AccumulatorRootHash Hash              // AccumulatorRootHash of the transaction.
	StateChangeHash     Hash              // StateChangeHash of the transaction.
	EventRootHash       Hash              // EventRootHash of the transaction.
	GasUsed             uint64            // GasUsed by the transaction, will be in gas units.  It should be 0.
	Success             bool              // Success of the transaction.
	VmStatus            string            // VmStatus of the transaction, this will contain the error if any.
	Changes             []*WriteSetChange // Changes to the ledger from the transaction, should never be empty.
	Events              []*Event          // Events emitted by the transaction, may be empty.
	Timestamp           uint64            // Timestamp is the Unix timestamp in microseconds when the block of the transaction was committed.
	BlockEndInfo        *BlockEndInfo     // BlockEndInfo of the transaction, this will contain information about block gas.
	StateCheckpointHash Hash              // StateCheckpointHash of the transaction. Optional, and will be "" if not set.
}

// TxnHash gives us the hash of the transaction.
func (o *BlockEpilogueTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will never be nil.
func (o *BlockEpilogueTransaction) TxnSuccess() *bool {
	return &o.Success
}

// TxnVersion gives us the ledger version of the transaction. It will never be nil.
func (o *BlockEpilogueTransaction) TxnVersion() *uint64 {
	return &o.Version
}

// UnmarshalJSON unmarshals the [BlockEpilogueTransaction] from JSON handling conversion between types
func (o *BlockEpilogueTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Version             U64               `json:"version"`
		Hash                Hash              `json:"hash"`
		AccumulatorRootHash Hash              `json:"accumulator_root_hash"`
		StateChangeHash     Hash              `json:"state_change_hash"`
		EventRootHash       Hash              `json:"event_root_hash"`
		GasUsed             U64               `json:"gas_used"`
		Success             bool              `json:"success"`
		VmStatus            string            `json:"vm_status"`
		Changes             []*WriteSetChange `json:"changes"`
		Timestamp           U64               `json:"timestamp"`
		BlockEndInfo        *BlockEndInfo     `json:"block_end_info"`
		StateCheckpointHash Hash              `json:"state_checkpoint_hash"` // Optional
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	o.Version = data.Version.ToUint64()
	o.Hash = data.Hash
	o.AccumulatorRootHash = data.AccumulatorRootHash
	o.StateChangeHash = data.StateChangeHash
	o.EventRootHash = data.EventRootHash
	o.GasUsed = data.GasUsed.ToUint64()
	o.Success = data.Success
	o.VmStatus = data.VmStatus
	o.Changes = data.Changes
	o.Timestamp = data.Timestamp.ToUint64()
	o.BlockEndInfo = data.BlockEndInfo
	o.StateCheckpointHash = data.StateCheckpointHash
	return nil
}

// StateCheckpointTransaction is a transaction that is a checkpoint of the state of the blockchain.  It is not necessarily at the end of a block.
type StateCheckpointTransaction struct {
	Version             uint64            // Version of the transaction, starts at 0 and increments per transaction.
	Hash                Hash              // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	AccumulatorRootHash Hash              // AccumulatorRootHash of the transaction.
	StateChangeHash     Hash              // StateChangeHash of the transaction.
	EventRootHash       Hash              // EventRootHash of the transaction.
	GasUsed             uint64            // GasUsed by the transaction, will be in gas units.  It should be 0.
	Success             bool              // Success of the transaction.
	VmStatus            string            // VmStatus of the transaction, this will contain the error if any.
	Changes             []*WriteSetChange // Changes to the ledger from the transaction, should never be empty.
	Timestamp           uint64            // Timestamp is the Unix timestamp in microseconds when the block of the transaction was committed.
	StateCheckpointHash Hash              // StateCheckpointHash of the transaction. Optional, and will be "" if not set.
}

// TxnHash gives us the hash of the transaction.
func (o *StateCheckpointTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will never be nil.
func (o *StateCheckpointTransaction) TxnSuccess() *bool {
	return &o.Success
}

// TxnVersion gives us the ledger version of the transaction. It will never be nil.
func (o *StateCheckpointTransaction) TxnVersion() *uint64 {
	return &o.Version
}

// UnmarshalJSON unmarshals the [StateCheckpointTransaction] from JSON handling conversion between types
func (o *StateCheckpointTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Version             U64               `json:"version"`
		Hash                Hash              `json:"hash"`
		AccumulatorRootHash Hash              `json:"accumulator_root_hash"`
		StateChangeHash     Hash              `json:"state_change_hash"`
		EventRootHash       Hash              `json:"event_root_hash"`
		GasUsed             U64               `json:"gas_used"`
		Success             bool              `json:"success"`
		VmStatus            string            `json:"vm_status"`
		Changes             []*WriteSetChange `json:"changes"`
		Timestamp           U64               `json:"timestamp"`
		StateCheckpointHash Hash              `json:"state_checkpoint_hash"` // Optional
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	o.Version = data.Version.ToUint64()
	o.Hash = data.Hash
	o.AccumulatorRootHash = data.AccumulatorRootHash
	o.StateChangeHash = data.StateChangeHash
	o.EventRootHash = data.EventRootHash
	o.GasUsed = data.GasUsed.ToUint64()
	o.Success = data.Success
	o.VmStatus = data.VmStatus
	o.Changes = data.Changes
	o.Timestamp = data.Timestamp.ToUint64()
	o.StateCheckpointHash = data.StateCheckpointHash
	return nil
}

// ValidatorTransaction is a transaction that is metadata about a block.  It's additional information from [BlockMetadataTransaction]
type ValidatorTransaction struct {
	Version             uint64            // Version of the transaction, starts at 0 and increments per transaction.
	Hash                Hash              // Hash of the transaction, it is a SHA3-256 hash in hexadecimal format with a leading 0x.
	AccumulatorRootHash Hash              // AccumulatorRootHash of the transaction.
	StateChangeHash     Hash              // StateChangeHash of the transaction.
	EventRootHash       Hash              // EventRootHash of the transaction.
	GasUsed             uint64            // GasUsed by the transaction, will be in gas units.  It should be 0.
	Success             bool              // Success of the transaction.
	VmStatus            string            // VmStatus of the transaction, this will contain the error if any.
	Changes             []*WriteSetChange // Changes to the ledger from the transaction, should never be empty.
	Events              []*Event          // Events emitted by the transaction, may be empty.
	Timestamp           uint64            // Timestamp is the Unix timestamp in microseconds when the block of the transaction was committed.
	StateCheckpointHash Hash              // StateCheckpointHash of the transaction. Optional, and will be "" if not set.
}

// TxnHash gives us the hash of the transaction.
func (o *ValidatorTransaction) TxnHash() Hash {
	return o.Hash
}

// TxnSuccess tells us if the transaction is a success.  It will never be nil.
func (o *ValidatorTransaction) TxnSuccess() *bool {
	return &o.Success
}

// TxnVersion gives us the ledger version of the transaction. It will never be nil.
func (o *ValidatorTransaction) TxnVersion() *uint64 {
	return &o.Version
}

// UnmarshalJSON unmarshals the [ValidatorTransaction] from JSON handling conversion between types
func (o *ValidatorTransaction) UnmarshalJSON(b []byte) error {
	type inner struct {
		Version             U64               `json:"version"`
		Hash                Hash              `json:"hash"`
		AccumulatorRootHash Hash              `json:"accumulator_root_hash"`
		StateChangeHash     Hash              `json:"state_change_hash"`
		EventRootHash       Hash              `json:"event_root_hash"`
		GasUsed             U64               `json:"gas_used"`
		Success             bool              `json:"success"`
		VmStatus            string            `json:"vm_status"`
		Changes             []*WriteSetChange `json:"changes"`
		Events              []*Event          `json:"events"`
		Timestamp           U64               `json:"timestamp"`
		StateCheckpointHash Hash              `json:"state_checkpoint_hash"` // Optional
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Version = data.Version.ToUint64()
	o.Hash = data.Hash
	o.AccumulatorRootHash = data.AccumulatorRootHash
	o.StateChangeHash = data.StateChangeHash
	o.EventRootHash = data.EventRootHash
	o.GasUsed = data.GasUsed.ToUint64()
	o.Success = data.Success
	o.VmStatus = data.VmStatus
	o.Changes = data.Changes
	o.Events = data.Events
	o.Timestamp = data.Timestamp.ToUint64()
	o.StateCheckpointHash = data.StateCheckpointHash

	return nil
}

// SubmitTransactionResponse is the response from submitting a transaction to the blockchain, it is the same
// as a [PendingTransaction]
type SubmitTransactionResponse = PendingTransaction

// BatchSubmitTransactionResponse is the response from submitting a batch of transactions to the blockchain
type BatchSubmitTransactionResponse struct {
	// TransactionFailures is the list of transactions that failed to submit, if it is empty, all were successful
	TransactionFailures []BatchSubmitTransactionFailure `json:"transaction_failures"`
}

// BatchSubmitTransactionFailure is a failure of a transaction in a batch submission,
type BatchSubmitTransactionFailure struct {
	// Error is the error that occurred when submitting the transaction
	Error Error
	// TransactionIndex is the index of submitted transactions that failed
	TransactionIndex uint32 `json:"transaction_index"`
}

// BlockEndInfo is the information about the block gas
type BlockEndInfo struct {
	BlockGasLimitReached        bool   `json:"block_gas_limit_reached"`         // BlockGasLimitReached is true if the block gas limit was reached.
	BlockOutputLimitReached     bool   `json:"block_output_limit_reached"`      // BlockOutputLimitReached is true if the block output limit was reached.
	BlockEffectiveBlockGasUnits uint64 `json:"block_effective_block_gas_units"` // BlockEffectiveBlockGasUnits is the effective gas units used in the block.
	BlockApproxOutputSize       uint64 `json:"block_approx_output_size"`        // BlockApproxOutputSize is the approximate output size of the block.
}
