package aptos

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// Transaction prefix used for signing
const (
	RawTransactionSalt         = "APTOS::RawTransaction"
	RawTransactionWithDataSalt = "APTOS::RawTransactionWithData"
)

// rawTransactionPrehash returns the SHA3-256 hash of the raw transaction salt.
func rawTransactionPrehash() []byte {
	hash := sha256.Sum256([]byte(RawTransactionSalt))
	return hash[:]
}

// MarshalBCS serializes a RawTransaction to BCS bytes.
func (txn *RawTransaction) MarshalBCS(ser *bcs.Serializer) {
	txn.Sender.MarshalBCS(ser)
	ser.U64(txn.SequenceNumber)

	// Serialize payload
	serializePayload(ser, txn.Payload)

	ser.U64(txn.MaxGasAmount)
	ser.U64(txn.GasUnitPrice)
	ser.U64(txn.ExpirationTimestampSeconds)
	ser.U8(txn.ChainID)
}

// UnmarshalBCS deserializes a RawTransaction from BCS bytes.
func (txn *RawTransaction) UnmarshalBCS(des *bcs.Deserializer) {
	txn.Sender.UnmarshalBCS(des)
	txn.SequenceNumber = des.U64()
	txn.Payload = deserializePayload(des)
	txn.MaxGasAmount = des.U64()
	txn.GasUnitPrice = des.U64()
	txn.ExpirationTimestampSeconds = des.U64()
	txn.ChainID = des.U8()
}

// SigningMessage returns the message to be signed for this transaction.
func (txn *RawTransaction) SigningMessage() ([]byte, error) {
	prehash := rawTransactionPrehash()

	txnBytes, err := bcs.Serialize(txn)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return append(prehash, txnBytes...), nil
}

// Hash returns the transaction hash.
func (txn *SignedTransaction) Hash() (string, error) {
	// Serialize the signed transaction
	txnBytes, err := bcs.Serialize(txn)
	if err != nil {
		return "", fmt.Errorf("failed to serialize signed transaction: %w", err)
	}

	// Hash with the transaction prefix
	prefix := sha256.Sum256([]byte("APTOS::Transaction"))

	// Combine prefix, variant byte (0 for user transaction), and serialized transaction
	hashInput := append(prefix[:], 0) // 0 = UserTransaction variant
	hashInput = append(hashInput, txnBytes...)

	hash := sha256.Sum256(hashInput)
	return "0x" + hex.EncodeToString(hash[:]), nil
}

// MarshalBCS serializes a SignedTransaction to BCS bytes.
func (txn *SignedTransaction) MarshalBCS(ser *bcs.Serializer) {
	txn.Transaction.MarshalBCS(ser)
	if txn.Authenticator != nil {
		txn.Authenticator.MarshalBCS(ser)
	} else {
		ser.SetError(fmt.Errorf("nil authenticator"))
	}
}

// UnmarshalBCS deserializes a SignedTransaction from BCS bytes.
func (txn *SignedTransaction) UnmarshalBCS(des *bcs.Deserializer) {
	if txn.Transaction == nil {
		txn.Transaction = &RawTransaction{}
	}
	txn.Transaction.UnmarshalBCS(des)
	txn.Authenticator = deserializeTransactionAuthenticator(des)
}

// Payload type variants for BCS serialization
const (
	PayloadVariantScript        = 0
	PayloadVariantModuleBundle  = 1 // Deprecated
	PayloadVariantEntryFunction = 2
	PayloadVariantMultisig      = 3
)

// serializePayload serializes a Payload to BCS.
func serializePayload(ser *bcs.Serializer, payload Payload) {
	switch p := payload.(type) {
	case *EntryFunctionPayload:
		ser.Uleb128(PayloadVariantEntryFunction)
		serializeEntryFunction(ser, p)
	case *ScriptPayload:
		ser.Uleb128(PayloadVariantScript)
		serializeScript(ser, p)
	default:
		ser.SetError(fmt.Errorf("unsupported payload type: %T", payload))
	}
}

// deserializePayload deserializes a Payload from BCS.
func deserializePayload(des *bcs.Deserializer) Payload {
	variant := des.Uleb128()
	switch variant {
	case PayloadVariantEntryFunction:
		return deserializeEntryFunction(des)
	case PayloadVariantScript:
		return deserializeScript(des)
	default:
		des.SetError(fmt.Errorf("unknown payload variant: %d", variant))
		return nil
	}
}

// serializeEntryFunction serializes an EntryFunctionPayload to BCS.
func serializeEntryFunction(ser *bcs.Serializer, ef *EntryFunctionPayload) {
	// Module ID
	ef.Module.Address.MarshalBCS(ser)
	ser.WriteString(ef.Module.Name)

	// Function name
	ser.WriteString(ef.Function)

	// Type arguments
	ser.Uleb128(uint32(len(ef.TypeArgs)))
	for _, typeArg := range ef.TypeArgs {
		typeArg.MarshalBCS(ser)
	}

	// Arguments - serialize each as BCS bytes
	ser.Uleb128(uint32(len(ef.Args)))
	for _, arg := range ef.Args {
		argBytes, err := serializeArg(arg)
		if err != nil {
			ser.SetError(err)
			return
		}
		ser.WriteBytes(argBytes)
	}
}

// deserializeEntryFunction deserializes an EntryFunctionPayload from BCS.
func deserializeEntryFunction(des *bcs.Deserializer) *EntryFunctionPayload {
	ef := &EntryFunctionPayload{}

	// Module ID
	ef.Module.Address.UnmarshalBCS(des)
	ef.Module.Name = des.ReadString()

	// Function name
	ef.Function = des.ReadString()

	// Type arguments
	numTypeArgs := des.Uleb128()
	ef.TypeArgs = make([]TypeTag, numTypeArgs)
	for i := uint32(0); i < numTypeArgs; i++ {
		ef.TypeArgs[i].UnmarshalBCS(des)
	}

	// Arguments (as raw bytes)
	numArgs := des.Uleb128()
	ef.Args = make([]any, numArgs)
	for i := uint32(0); i < numArgs; i++ {
		ef.Args[i] = des.ReadBytes()
	}

	return ef
}

// serializeScript serializes a ScriptPayload to BCS.
func serializeScript(ser *bcs.Serializer, sp *ScriptPayload) {
	ser.WriteBytes(sp.Code)

	// Type arguments
	ser.Uleb128(uint32(len(sp.TypeArgs)))
	for _, typeArg := range sp.TypeArgs {
		typeArg.MarshalBCS(ser)
	}

	// Script arguments - serialize with type tag
	ser.Uleb128(uint32(len(sp.Args)))
	for _, arg := range sp.Args {
		serializeScriptArg(ser, arg)
	}
}

// deserializeScript deserializes a ScriptPayload from BCS.
func deserializeScript(des *bcs.Deserializer) *ScriptPayload {
	sp := &ScriptPayload{}

	sp.Code = des.ReadBytes()

	// Type arguments
	numTypeArgs := des.Uleb128()
	sp.TypeArgs = make([]TypeTag, numTypeArgs)
	for i := uint32(0); i < numTypeArgs; i++ {
		sp.TypeArgs[i].UnmarshalBCS(des)
	}

	// Script arguments
	numArgs := des.Uleb128()
	sp.Args = make([]any, numArgs)
	for i := uint32(0); i < numArgs; i++ {
		sp.Args[i] = deserializeScriptArg(des)
	}

	return sp
}

// serializeArg serializes a single argument to BCS bytes.
func serializeArg(arg any) ([]byte, error) {
	switch v := arg.(type) {
	case []byte:
		return v, nil
	case string:
		// Assume it's a string argument - serialize as BCS string
		ser := bcs.NewSerializer()
		ser.WriteString(v)
		return ser.ToBytes(), ser.Error()
	case uint8:
		ser := bcs.NewSerializer()
		ser.U8(v)
		return ser.ToBytes(), nil
	case uint16:
		ser := bcs.NewSerializer()
		ser.U16(v)
		return ser.ToBytes(), nil
	case uint32:
		ser := bcs.NewSerializer()
		ser.U32(v)
		return ser.ToBytes(), nil
	case uint64:
		ser := bcs.NewSerializer()
		ser.U64(v)
		return ser.ToBytes(), nil
	case bool:
		ser := bcs.NewSerializer()
		ser.Bool(v)
		return ser.ToBytes(), nil
	case AccountAddress:
		return v[:], nil
	case *AccountAddress:
		if v == nil {
			return nil, fmt.Errorf("nil address")
		}
		return (*v)[:], nil
	case bcs.Marshaler:
		return bcs.Serialize(v)
	default:
		return nil, fmt.Errorf("unsupported argument type: %T", arg)
	}
}

// Script argument type variants
const (
	ScriptArgU8       = 0
	ScriptArgU64      = 1
	ScriptArgU128     = 2
	ScriptArgAddress  = 3
	ScriptArgU8Vector = 4
	ScriptArgBool     = 5
	ScriptArgU16      = 6
	ScriptArgU32      = 7
	ScriptArgU256     = 8
)

// serializeScriptArg serializes a script argument with its type tag.
func serializeScriptArg(ser *bcs.Serializer, arg any) {
	switch v := arg.(type) {
	case bool:
		ser.U8(ScriptArgBool)
		ser.Bool(v)
	case uint8:
		ser.U8(ScriptArgU8)
		ser.U8(v)
	case uint16:
		ser.U8(ScriptArgU16)
		ser.U16(v)
	case uint32:
		ser.U8(ScriptArgU32)
		ser.U32(v)
	case uint64:
		ser.U8(ScriptArgU64)
		ser.U64(v)
	case []byte:
		ser.U8(ScriptArgU8Vector)
		ser.WriteBytes(v)
	case AccountAddress:
		ser.U8(ScriptArgAddress)
		ser.FixedBytes(v[:])
	case *AccountAddress:
		ser.U8(ScriptArgAddress)
		ser.FixedBytes((*v)[:])
	default:
		ser.SetError(fmt.Errorf("unsupported script argument type: %T", arg))
	}
}

// deserializeScriptArg deserializes a script argument.
func deserializeScriptArg(des *bcs.Deserializer) any {
	variant := des.U8()
	switch variant {
	case ScriptArgBool:
		return des.Bool()
	case ScriptArgU8:
		return des.U8()
	case ScriptArgU16:
		return des.U16()
	case ScriptArgU32:
		return des.U32()
	case ScriptArgU64:
		return des.U64()
	case ScriptArgU8Vector:
		return des.ReadBytes()
	case ScriptArgAddress:
		var addr AccountAddress
		des.ReadFixedBytesInto(addr[:])
		return addr
	default:
		des.SetError(fmt.Errorf("unknown script argument variant: %d", variant))
		return nil
	}
}

// Transaction authenticator variants (for transaction-level authenticators).
// Note: These are different from account authenticator variants in crypto package.
type TransactionAuthenticatorVariant uint8

const (
	TransactionAuthenticatorVariantEd25519      TransactionAuthenticatorVariant = 0
	TransactionAuthenticatorVariantMultiEd25519 TransactionAuthenticatorVariant = 1
	TransactionAuthenticatorVariantMultiAgent   TransactionAuthenticatorVariant = 2
	TransactionAuthenticatorVariantFeePayer     TransactionAuthenticatorVariant = 3
	TransactionAuthenticatorVariantSingleSender TransactionAuthenticatorVariant = 4
)

// Deprecated variant constants for backwards compatibility
const (
	TransactionAuthenticatorEd25519      = TransactionAuthenticatorVariantEd25519
	TransactionAuthenticatorMultiEd25519 = TransactionAuthenticatorVariantMultiEd25519
	TransactionAuthenticatorMultiAgent   = TransactionAuthenticatorVariantMultiAgent
	TransactionAuthenticatorFeePayer     = TransactionAuthenticatorVariantFeePayer
	TransactionAuthenticatorSingleSender = TransactionAuthenticatorVariantSingleSender
)

// Variant constants for use in NewFeePayerSignedTransaction, etc.
const (
	AccountAuthenticatorVariantFeePayer   = TransactionAuthenticatorVariantFeePayer
	AccountAuthenticatorVariantMultiAgent = TransactionAuthenticatorVariantMultiAgent
)

// TransactionAuthenticator is an interface for all transaction-level authenticators.
// This is different from AccountAuthenticator, which authenticates a single account.
// TransactionAuthenticator handles constructs like FeePayer and MultiAgent.
type TransactionAuthenticator interface {
	bcs.Struct
	// Verify returns true if this authenticator approves the given message.
	Verify(data []byte) bool
}

// SingleSenderAuthenticator wraps a single AccountAuthenticator for simple transactions.
type SingleSenderAuthenticator struct {
	Sender *AccountAuthenticator
}

func (a *SingleSenderAuthenticator) Verify(msg []byte) bool {
	return a.Sender.Verify(msg)
}

func (a *SingleSenderAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(TransactionAuthenticatorVariantSingleSender))
	a.Sender.MarshalBCS(ser)
}

func (a *SingleSenderAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	a.Sender = &AccountAuthenticator{}
	a.Sender.UnmarshalBCS(des)
}

// Ed25519TransactionAuthenticator wraps an Ed25519 AccountAuthenticator (legacy).
type Ed25519TransactionAuthenticator struct {
	Sender *AccountAuthenticator
}

func (a *Ed25519TransactionAuthenticator) Verify(msg []byte) bool {
	return a.Sender.Verify(msg)
}

func (a *Ed25519TransactionAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(TransactionAuthenticatorVariantEd25519))
	// For Ed25519, serialize the inner authenticator without variant prefix
	a.Sender.Auth.MarshalBCS(ser)
}

func (a *Ed25519TransactionAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	a.Sender = &AccountAuthenticator{
		Variant: AccountAuthenticatorEd25519,
	}
	auth := &Ed25519Authenticator{}
	auth.UnmarshalBCS(des)
	a.Sender.Auth = auth
}

// serializeAuthenticator serializes an AccountAuthenticator to BCS.
func serializeAuthenticator(ser *bcs.Serializer, auth *AccountAuthenticator) {
	if auth == nil {
		ser.SetError(fmt.Errorf("nil authenticator"))
		return
	}
	auth.MarshalBCS(ser)
}

// deserializeAuthenticator deserializes an AccountAuthenticator from BCS.
func deserializeAuthenticator(des *bcs.Deserializer) *AccountAuthenticator {
	auth := &AccountAuthenticator{}
	auth.UnmarshalBCS(des)
	return auth
}

// deserializeTransactionAuthenticator deserializes a TransactionAuthenticator from BCS.
func deserializeTransactionAuthenticator(des *bcs.Deserializer) TransactionAuthenticator {
	variant := des.Uleb128()
	switch TransactionAuthenticatorVariant(variant) {
	case TransactionAuthenticatorVariantEd25519:
		auth := &Ed25519TransactionAuthenticator{}
		auth.Sender = &AccountAuthenticator{Variant: AccountAuthenticatorEd25519}
		inner := &Ed25519Authenticator{}
		inner.UnmarshalBCS(des)
		auth.Sender.Auth = inner
		return auth
	case TransactionAuthenticatorVariantSingleSender:
		auth := &SingleSenderAuthenticator{}
		auth.Sender = &AccountAuthenticator{}
		auth.Sender.UnmarshalBCS(des)
		return auth
	case TransactionAuthenticatorVariantMultiAgent:
		auth := &MultiAgentAuthenticator{}
		auth.deserializeInner(des)
		return auth
	case TransactionAuthenticatorVariantFeePayer:
		auth := &FeePayerAuthenticator{}
		auth.deserializeInner(des)
		return auth
	default:
		des.SetError(fmt.Errorf("unknown transaction authenticator variant: %d", variant))
		return nil
	}
}

// MultiAgentAuthenticator represents a multi-agent transaction authenticator.
// This is used for transactions with multiple signers.
type MultiAgentAuthenticator struct {
	Sender                   *AccountAuthenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []*AccountAuthenticator
}

func (a *MultiAgentAuthenticator) Verify(msg []byte) bool {
	if !a.Sender.Verify(msg) {
		return false
	}
	for _, auth := range a.SecondarySigners {
		if !auth.Verify(msg) {
			return false
		}
	}
	return true
}

func (a *MultiAgentAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(TransactionAuthenticatorVariantMultiAgent))
	serializeAuthenticator(ser, a.Sender)

	// Secondary signer addresses
	ser.Uleb128(uint32(len(a.SecondarySignerAddresses)))
	for _, addr := range a.SecondarySignerAddresses {
		addr.MarshalBCS(ser)
	}

	// Secondary signers
	ser.Uleb128(uint32(len(a.SecondarySigners)))
	for _, auth := range a.SecondarySigners {
		serializeAuthenticator(ser, auth)
	}
}

func (a *MultiAgentAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	// Note: variant already consumed by deserializeTransactionAuthenticator
	a.deserializeInner(des)
}

func (a *MultiAgentAuthenticator) deserializeInner(des *bcs.Deserializer) {
	a.Sender = &AccountAuthenticator{}
	a.Sender.UnmarshalBCS(des)

	numAddresses := des.Uleb128()
	a.SecondarySignerAddresses = make([]AccountAddress, numAddresses)
	for i := uint32(0); i < numAddresses; i++ {
		a.SecondarySignerAddresses[i].UnmarshalBCS(des)
	}

	numSigners := des.Uleb128()
	a.SecondarySigners = make([]*AccountAuthenticator, numSigners)
	for i := uint32(0); i < numSigners; i++ {
		a.SecondarySigners[i] = &AccountAuthenticator{}
		a.SecondarySigners[i].UnmarshalBCS(des)
	}
}

// FeePayerAuthenticator represents a fee payer (sponsored) transaction authenticator.
type FeePayerAuthenticator struct {
	Sender                   *AccountAuthenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []*AccountAuthenticator
	FeePayerAddress          AccountAddress
	FeePayerAuth             *AccountAuthenticator
}

func (a *FeePayerAuthenticator) Verify(msg []byte) bool {
	if !a.Sender.Verify(msg) {
		return false
	}
	for _, auth := range a.SecondarySigners {
		if !auth.Verify(msg) {
			return false
		}
	}
	return a.FeePayerAuth.Verify(msg)
}

func (a *FeePayerAuthenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(TransactionAuthenticatorVariantFeePayer))
	serializeAuthenticator(ser, a.Sender)

	// Secondary signer addresses
	ser.Uleb128(uint32(len(a.SecondarySignerAddresses)))
	for _, addr := range a.SecondarySignerAddresses {
		addr.MarshalBCS(ser)
	}

	// Secondary signers
	ser.Uleb128(uint32(len(a.SecondarySigners)))
	for _, auth := range a.SecondarySigners {
		serializeAuthenticator(ser, auth)
	}

	// Fee payer
	a.FeePayerAddress.MarshalBCS(ser)
	serializeAuthenticator(ser, a.FeePayerAuth)
}

func (a *FeePayerAuthenticator) UnmarshalBCS(des *bcs.Deserializer) {
	// Note: variant already consumed by deserializeTransactionAuthenticator
	a.deserializeInner(des)
}

func (a *FeePayerAuthenticator) deserializeInner(des *bcs.Deserializer) {
	a.Sender = &AccountAuthenticator{}
	a.Sender.UnmarshalBCS(des)

	numAddresses := des.Uleb128()
	a.SecondarySignerAddresses = make([]AccountAddress, numAddresses)
	for i := uint32(0); i < numAddresses; i++ {
		a.SecondarySignerAddresses[i].UnmarshalBCS(des)
	}

	numSigners := des.Uleb128()
	a.SecondarySigners = make([]*AccountAuthenticator, numSigners)
	for i := uint32(0); i < numSigners; i++ {
		a.SecondarySigners[i] = &AccountAuthenticator{}
		a.SecondarySigners[i].UnmarshalBCS(des)
	}

	a.FeePayerAddress.UnmarshalBCS(des)
	a.FeePayerAuth = &AccountAuthenticator{}
	a.FeePayerAuth.UnmarshalBCS(des)
}

// MultiAgentTransaction represents a transaction with multiple signers.
type MultiAgentTransaction struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
}

// MarshalBCS serializes a MultiAgentTransaction.
func (txn *MultiAgentTransaction) MarshalBCS(ser *bcs.Serializer) {
	ser.U8(0) // MultiAgent variant
	txn.RawTxn.MarshalBCS(ser)
	ser.Uleb128(uint32(len(txn.SecondarySigners)))
	for _, addr := range txn.SecondarySigners {
		addr.MarshalBCS(ser)
	}
}

// SigningMessage returns the message to be signed for a multi-agent transaction.
func (txn *MultiAgentTransaction) SigningMessage() ([]byte, error) {
	prehash := sha256.Sum256([]byte(RawTransactionWithDataSalt))

	ser := bcs.NewSerializer()
	txn.MarshalBCS(ser)
	if ser.Error() != nil {
		return nil, ser.Error()
	}

	return append(prehash[:], ser.ToBytes()...), nil
}

// FeePayerTransaction represents a sponsored transaction.
type FeePayerTransaction struct {
	RawTxn           *RawTransaction
	SecondarySigners []AccountAddress
	FeePayer         AccountAddress
}

// MarshalBCS serializes a FeePayerTransaction.
func (txn *FeePayerTransaction) MarshalBCS(ser *bcs.Serializer) {
	ser.U8(1) // FeePayer variant
	txn.RawTxn.MarshalBCS(ser)
	ser.Uleb128(uint32(len(txn.SecondarySigners)))
	for _, addr := range txn.SecondarySigners {
		addr.MarshalBCS(ser)
	}
	txn.FeePayer.MarshalBCS(ser)
}

// SigningMessage returns the message to be signed for a fee payer transaction.
func (txn *FeePayerTransaction) SigningMessage() ([]byte, error) {
	prehash := sha256.Sum256([]byte(RawTransactionWithDataSalt))

	ser := bcs.NewSerializer()
	txn.MarshalBCS(ser)
	if ser.Error() != nil {
		return nil, ser.Error()
	}

	return append(prehash[:], ser.ToBytes()...), nil
}

// SignTransaction signs a raw transaction with the given signer.
// The signer can be any type that implements the Signer interface
// (Ed25519PrivateKey, SingleSigner, etc.)
func SignTransaction(signer Signer, txn *RawTransaction) (*SignedTransaction, error) {
	signingMessage, err := txn.SigningMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to get signing message: %w", err)
	}

	auth, err := signer.Sign(signingMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Wrap the account authenticator in a transaction authenticator
	txnAuth := &SingleSenderAuthenticator{Sender: auth}

	return &SignedTransaction{
		Transaction:   txn,
		Authenticator: txnAuth,
	}, nil
}

// SimulationAuthenticator creates an authenticator for simulation.
// This is used when simulating transactions to estimate gas.
func SimulationAuthenticator(signer Signer) *AccountAuthenticator {
	return signer.SimulationAuthenticator()
}

// RawTransactionWithDataPrehash returns the prehash for RawTransactionWithData.
func RawTransactionWithDataPrehash() []byte {
	hash := sha256.Sum256([]byte(RawTransactionWithDataSalt))
	return hash[:]
}

// NewFeePayerSignedTransaction creates a signed transaction with fee payer authentication.
func NewFeePayerSignedTransaction(
	rawTxn *RawTransaction,
	sender *AccountAuthenticator,
	secondarySignerAddresses []AccountAddress,
	secondarySigners []*AccountAuthenticator,
	feePayerAddress AccountAddress,
	feePayerAuth *AccountAuthenticator,
) (*SignedTransaction, error) {
	auth := &FeePayerAuthenticator{
		Sender:                   sender,
		SecondarySignerAddresses: secondarySignerAddresses,
		SecondarySigners:         secondarySigners,
		FeePayerAddress:          feePayerAddress,
		FeePayerAuth:             feePayerAuth,
	}

	return &SignedTransaction{
		Transaction:   rawTxn,
		Authenticator: auth,
	}, nil
}

// NewMultiAgentSignedTransaction creates a signed transaction with multi-agent authentication.
func NewMultiAgentSignedTransaction(
	rawTxn *RawTransaction,
	sender *AccountAuthenticator,
	secondarySignerAddresses []AccountAddress,
	secondarySigners []*AccountAuthenticator,
) (*SignedTransaction, error) {
	auth := &MultiAgentAuthenticator{
		Sender:                   sender,
		SecondarySignerAddresses: secondarySignerAddresses,
		SecondarySigners:         secondarySigners,
	}

	return &SignedTransaction{
		Transaction:   rawTxn,
		Authenticator: auth,
	}, nil
}

// NewSingleSenderSignedTransaction creates a signed transaction with single sender authentication.
func NewSingleSenderSignedTransaction(
	rawTxn *RawTransaction,
	sender *AccountAuthenticator,
) (*SignedTransaction, error) {
	auth := &SingleSenderAuthenticator{
		Sender: sender,
	}

	return &SignedTransaction{
		Transaction:   rawTxn,
		Authenticator: auth,
	}, nil
}

// SignFeePayerTransaction signs a fee payer transaction with the sender and fee payer signers.
// This is a convenience function that handles the signing flow for sponsored transactions.
func SignFeePayerTransaction(
	sender Signer,
	feePayer Signer,
	rawTxn *RawTransaction,
	feePayerAddress AccountAddress,
	secondarySigners []Signer,
	secondaryAddresses []AccountAddress,
) (*SignedTransaction, error) {
	// Create the fee payer transaction for signing
	feePayerTxn := &FeePayerTransaction{
		RawTxn:           rawTxn,
		SecondarySigners: secondaryAddresses,
		FeePayer:         feePayerAddress,
	}

	// Get the signing message
	signingMessage, err := feePayerTxn.SigningMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to get signing message: %w", err)
	}

	// Sign with sender
	senderAuth, err := sender.Sign(signingMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with sender: %w", err)
	}

	// Sign with secondary signers
	secondaryAuths := make([]*AccountAuthenticator, len(secondarySigners))
	for i, signer := range secondarySigners {
		auth, err := signer.Sign(signingMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to sign with secondary signer %d: %w", i, err)
		}
		secondaryAuths[i] = auth
	}

	// Sign with fee payer
	feePayerAuth, err := feePayer.Sign(signingMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with fee payer: %w", err)
	}

	return NewFeePayerSignedTransaction(
		rawTxn,
		senderAuth,
		secondaryAddresses,
		secondaryAuths,
		feePayerAddress,
		feePayerAuth,
	)
}

// SignMultiAgentTransaction signs a multi-agent transaction with multiple signers.
// This is a convenience function that handles the signing flow for multi-agent transactions.
func SignMultiAgentTransaction(
	sender Signer,
	secondarySigners []Signer,
	secondaryAddresses []AccountAddress,
	rawTxn *RawTransaction,
) (*SignedTransaction, error) {
	// Create the multi-agent transaction for signing
	multiAgentTxn := &MultiAgentTransaction{
		RawTxn:           rawTxn,
		SecondarySigners: secondaryAddresses,
	}

	// Get the signing message
	signingMessage, err := multiAgentTxn.SigningMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to get signing message: %w", err)
	}

	// Sign with sender
	senderAuth, err := sender.Sign(signingMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with sender: %w", err)
	}

	// Sign with secondary signers
	secondaryAuths := make([]*AccountAuthenticator, len(secondarySigners))
	for i, signer := range secondarySigners {
		auth, err := signer.Sign(signingMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to sign with secondary signer %d: %w", i, err)
		}
		secondaryAuths[i] = auth
	}

	return NewMultiAgentSignedTransaction(
		rawTxn,
		senderAuth,
		secondaryAddresses,
		secondaryAuths,
	)
}
