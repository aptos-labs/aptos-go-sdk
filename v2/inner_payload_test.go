package aptos

import (
	"context"
	"math"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleEntryFunction() *EntryFunctionPayload {
	return &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		Args:     []any{AccountTwo, uint64(100)},
	}
}

func TestTransactionInnerPayload_RoundTrip(t *testing.T) {
	t.Parallel()
	nonce := uint64(0xdeadbeef)
	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             math.MaxUint64,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &TransactionInnerPayload{
			Executable:            sampleEntryFunction(),
			ReplayProtectionNonce: &nonce,
		},
	}

	raw, err := bcs.Serialize(txn)
	require.NoError(t, err)

	out := &RawTransaction{}
	require.NoError(t, bcs.Deserialize(out, raw))

	inner, ok := out.Payload.(*TransactionInnerPayload)
	require.True(t, ok, "payload should deserialize as *TransactionInnerPayload, got %T", out.Payload)
	require.NotNil(t, inner.ReplayProtectionNonce)
	assert.Equal(t, nonce, *inner.ReplayProtectionNonce)
	assert.Nil(t, inner.MultisigAddress)

	ef, ok := inner.Executable.(*EntryFunctionPayload)
	require.True(t, ok, "executable should be an entry function, got %T", inner.Executable)
	assert.Equal(t, "aptos_account", ef.Module.Name)
	assert.Equal(t, "transfer", ef.Function)

	// Re-serializing the round-tripped transaction must be byte-identical.
	raw2, err := bcs.Serialize(out)
	require.NoError(t, err)
	assert.Equal(t, raw, raw2)
}

func TestTransactionInnerPayload_UsesPayloadVariant4(t *testing.T) {
	t.Parallel()
	txn := &RawTransaction{
		Sender:  AccountOne,
		Payload: &TransactionInnerPayload{Executable: sampleEntryFunction()},
	}
	raw, err := bcs.Serialize(txn)
	require.NoError(t, err)

	// Layout: sender (32 bytes) || sequence_number (8 bytes) || payload variant.
	require.Greater(t, len(raw), 40)
	assert.Equal(t, byte(4), raw[40], "inner payload must use TransactionPayload variant 4")
}

func TestBuildTransaction_Orderless(t *testing.T) {
	t.Parallel()
	// ChainID is cached from the network config (4), and no gas estimation is
	// requested, so BuildTransaction makes no HTTP calls for an orderless txn.
	client := newTestClient(t, jsonHandler(nil))
	nonce := uint64(42)

	txn, err := client.BuildTransaction(
		context.Background(),
		AccountOne,
		sampleEntryFunction(),
		WithReplayProtectionNonce(nonce),
	)
	require.NoError(t, err)

	assert.Equal(t, uint64(math.MaxUint64), txn.SequenceNumber, "orderless txns use u64::MAX as the sequence number")

	inner, ok := txn.Payload.(*TransactionInnerPayload)
	require.True(t, ok, "payload should be wrapped in *TransactionInnerPayload, got %T", txn.Payload)
	require.NotNil(t, inner.ReplayProtectionNonce)
	assert.Equal(t, nonce, *inner.ReplayProtectionNonce)

	_, ok = inner.Executable.(*EntryFunctionPayload)
	assert.True(t, ok, "original entry function should be the executable")
}

func TestBuildTransaction_OrderlessDoesNotDoubleWrap(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, jsonHandler(nil))
	nonce := uint64(7)
	pre := &TransactionInnerPayload{Executable: sampleEntryFunction(), ReplayProtectionNonce: &nonce}

	txn, err := client.BuildTransaction(
		context.Background(),
		AccountOne,
		pre,
		WithReplayProtectionNonce(nonce),
	)
	require.NoError(t, err)

	inner, ok := txn.Payload.(*TransactionInnerPayload)
	require.True(t, ok)
	_, nested := inner.Executable.(*TransactionInnerPayload)
	assert.False(t, nested, "an already-inner payload must not be wrapped again")
}

func TestWrapOrderless_SetsNonceOnExistingInnerWithoutOne(t *testing.T) {
	t.Parallel()
	// An inner payload that arrives without a nonce should have the supplied
	// nonce applied in place, rather than being wrapped again.
	inner := &TransactionInnerPayload{Executable: sampleEntryFunction()}
	nonce := uint64(99)

	wrapped := wrapOrderless(inner, &nonce)

	assert.Same(t, inner, wrapped, "existing inner payload should be reused, not re-wrapped")
	require.NotNil(t, inner.ReplayProtectionNonce)
	assert.Equal(t, nonce, *inner.ReplayProtectionNonce)
}

func TestTransactionInnerPayload_PayloadType(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "inner_payload", (&TransactionInnerPayload{}).payloadType())
}

func TestTransactionInnerPayload_ScriptExecutable_RoundTrip(t *testing.T) {
	t.Parallel()
	nonce := uint64(1)
	txn := &RawTransaction{
		Sender:         AccountOne,
		SequenceNumber: math.MaxUint64,
		Payload: &TransactionInnerPayload{
			Executable: &ScriptPayload{
				Code: []byte{0xa1, 0x02, 0x03},
				Args: []any{uint64(7), AccountTwo},
			},
			ReplayProtectionNonce: &nonce,
		},
	}

	raw, err := bcs.Serialize(txn)
	require.NoError(t, err)

	out := &RawTransaction{}
	require.NoError(t, bcs.Deserialize(out, raw))

	inner, ok := out.Payload.(*TransactionInnerPayload)
	require.True(t, ok)
	script, ok := inner.Executable.(*ScriptPayload)
	require.True(t, ok, "executable should round-trip as a script, got %T", inner.Executable)
	assert.Equal(t, []byte{0xa1, 0x02, 0x03}, script.Code)
	require.Len(t, script.Args, 2)
	assert.Equal(t, uint64(7), script.Args[0])
	assert.Equal(t, AccountTwo, script.Args[1])
}

func TestTransactionInnerPayload_EmptyExecutable_RoundTrip(t *testing.T) {
	t.Parallel()
	txn := &RawTransaction{
		Sender:         AccountOne,
		SequenceNumber: math.MaxUint64,
		Payload:        &TransactionInnerPayload{Executable: nil},
	}

	raw, err := bcs.Serialize(txn)
	require.NoError(t, err)

	out := &RawTransaction{}
	require.NoError(t, bcs.Deserialize(out, raw))

	inner, ok := out.Payload.(*TransactionInnerPayload)
	require.True(t, ok)
	assert.Nil(t, inner.Executable, "empty executable should round-trip as nil")
}

func TestTransactionInnerPayload_MultisigAddress_RoundTrip(t *testing.T) {
	t.Parallel()
	multisig := AccountThree
	nonce := uint64(0x1234)
	txn := &RawTransaction{
		Sender:         AccountOne,
		SequenceNumber: math.MaxUint64,
		Payload: &TransactionInnerPayload{
			Executable:            sampleEntryFunction(),
			MultisigAddress:       &multisig,
			ReplayProtectionNonce: &nonce,
		},
	}

	raw, err := bcs.Serialize(txn)
	require.NoError(t, err)

	out := &RawTransaction{}
	require.NoError(t, bcs.Deserialize(out, raw))

	inner, ok := out.Payload.(*TransactionInnerPayload)
	require.True(t, ok)
	require.NotNil(t, inner.MultisigAddress)
	assert.Equal(t, multisig, *inner.MultisigAddress)
	require.NotNil(t, inner.ReplayProtectionNonce)
	assert.Equal(t, nonce, *inner.ReplayProtectionNonce)
}

func TestSerializeExecutable_UnsupportedTypeErrors(t *testing.T) {
	t.Parallel()
	ser := bcs.NewSerializer()
	// A nested inner payload is a Payload but not a valid executable.
	serializeExecutable(ser, &TransactionInnerPayload{})
	require.Error(t, ser.Error())
	assert.Contains(t, ser.Error().Error(), "unsupported inner executable type")
}

func TestDeserializeInnerPayload_BadInnerVariant(t *testing.T) {
	t.Parallel()
	des := bcs.NewDeserializer([]byte{0x01}) // inner variant 1 is unknown
	deserializeInnerPayload(des)
	require.Error(t, des.Error())
	assert.Contains(t, des.Error().Error(), "unknown inner payload variant")
}

func TestDeserializeInnerPayload_BadExtraConfigVariant(t *testing.T) {
	t.Parallel()
	// inner variant 0, empty executable (2), then an unknown extra-config variant.
	des := bcs.NewDeserializer([]byte{0x00, 0x02, 0x05})
	deserializeInnerPayload(des)
	require.Error(t, des.Error())
	assert.Contains(t, des.Error().Error(), "unknown extra config variant")
}

func TestDeserializeExecutable_BadVariant(t *testing.T) {
	t.Parallel()
	des := bcs.NewDeserializer([]byte{0x09}) // no such executable variant
	assert.Nil(t, deserializeExecutable(des))
	require.Error(t, des.Error())
	assert.Contains(t, des.Error().Error(), "unknown inner executable variant")
}
