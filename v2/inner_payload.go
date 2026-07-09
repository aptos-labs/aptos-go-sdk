package aptos

import (
	"fmt"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// Inner (TransactionPayload::Payload) BCS variant constants. These mirror the
// on-chain enum layout for the newer transaction payload format that carries a
// replay-protection nonce (enabling orderless transactions) and an optional
// multisig address.
const (
	// PayloadVariantInner is the TransactionPayload variant for the wrapped
	// TransactionInnerPayload format (AIP-123).
	PayloadVariantInner = 4

	innerPayloadVariantV1 = 0

	executableVariantScript        = 0
	executableVariantEntryFunction = 1
	executableVariantEmpty         = 2

	extraConfigVariantV1 = 0
)

// TransactionInnerPayload wraps an executable (entry function or script) with
// extra configuration using the TransactionPayload::Payload format (variant 4).
//
// Its main use is orderless transactions: setting ReplayProtectionNonce lets a
// transaction be validated by a one-time nonce instead of the account's
// sequence number, so transactions need not be ordered. When built through
// [Client.BuildTransaction] with [WithReplayProtectionNonce], the wrapping and
// the u64::MAX sequence number are handled automatically.
type TransactionInnerPayload struct {
	// Executable is the payload to run: *EntryFunctionPayload, *ScriptPayload,
	// or nil for an empty executable.
	Executable Payload
	// MultisigAddress, when set, targets an on-chain multisig account.
	MultisigAddress *AccountAddress
	// ReplayProtectionNonce, when set, makes the transaction orderless.
	ReplayProtectionNonce *uint64
}

func (p *TransactionInnerPayload) payloadType() string {
	return "inner_payload"
}

// wrapOrderless wraps an entry-function or script payload in a
// TransactionInnerPayload carrying the replay-protection nonce. A payload that
// is already a TransactionInnerPayload is returned unchanged (with its nonce
// set if not already present) to avoid double-wrapping.
func wrapOrderless(payload Payload, nonce *uint64) Payload {
	if inner, ok := payload.(*TransactionInnerPayload); ok {
		if inner.ReplayProtectionNonce == nil {
			inner.ReplayProtectionNonce = nonce
		}
		return inner
	}
	return &TransactionInnerPayload{
		Executable:            payload,
		ReplayProtectionNonce: nonce,
	}
}

// serializeInnerPayload writes the inner-payload body (everything after the
// TransactionPayload variant tag).
func serializeInnerPayload(ser *bcs.Serializer, p *TransactionInnerPayload) {
	ser.Uleb128(innerPayloadVariantV1)
	serializeExecutable(ser, p.Executable)
	serializeExtraConfig(ser, p)
}

func serializeExecutable(ser *bcs.Serializer, executable Payload) {
	switch e := executable.(type) {
	case nil:
		ser.Uleb128(executableVariantEmpty)
	case *EntryFunctionPayload:
		ser.Uleb128(executableVariantEntryFunction)
		serializeEntryFunction(ser, e)
	case *ScriptPayload:
		ser.Uleb128(executableVariantScript)
		serializeScript(ser, e)
	default:
		ser.SetError(fmt.Errorf("unsupported inner executable type: %T", executable))
	}
}

func serializeExtraConfig(ser *bcs.Serializer, p *TransactionInnerPayload) {
	ser.Uleb128(extraConfigVariantV1)
	bcs.SerializeOption(ser, p.MultisigAddress, func(ser *bcs.Serializer, addr AccountAddress) {
		addr.MarshalBCS(ser)
	})
	bcs.SerializeOption(ser, p.ReplayProtectionNonce, func(ser *bcs.Serializer, nonce uint64) {
		ser.U64(nonce)
	})
}

// deserializeInnerPayload reads the inner-payload body (the TransactionPayload
// variant tag has already been consumed).
func deserializeInnerPayload(des *bcs.Deserializer) *TransactionInnerPayload {
	if v := des.Uleb128(); v != innerPayloadVariantV1 {
		des.SetError(fmt.Errorf("unknown inner payload variant: %d", v))
		return nil
	}

	p := &TransactionInnerPayload{}
	p.Executable = deserializeExecutable(des)

	if v := des.Uleb128(); v != extraConfigVariantV1 {
		des.SetError(fmt.Errorf("unknown extra config variant: %d", v))
		return nil
	}
	p.MultisigAddress = bcs.DeserializeOption(des, func(des *bcs.Deserializer) AccountAddress {
		var addr AccountAddress
		addr.UnmarshalBCS(des)
		return addr
	})
	p.ReplayProtectionNonce = bcs.DeserializeOption(des, func(des *bcs.Deserializer) uint64 {
		return des.U64()
	})

	return p
}

func deserializeExecutable(des *bcs.Deserializer) Payload {
	switch v := des.Uleb128(); v {
	case executableVariantEmpty:
		return nil
	case executableVariantEntryFunction:
		return deserializeEntryFunction(des)
	case executableVariantScript:
		return deserializeScript(des)
	default:
		des.SetError(fmt.Errorf("unknown inner executable variant: %d", v))
		return nil
	}
}
