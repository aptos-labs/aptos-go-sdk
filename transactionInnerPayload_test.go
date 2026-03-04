package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionInnerPayload_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	ef := &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{{0x01}},
	}
	nonce := uint64(42)
	addr := AccountTwo
	original := TransactionInnerPayload{
		Payload: &TransactionInnerPayloadV1{
			Executable: TransactionExecutable{Inner: ef},
			ExtraConfig: TransactionExtraConfig{
				Inner: &TransactionExtraConfigV1{
					MultisigAddress:       &addr,
					ReplayProtectionNonce: &nonce,
				},
			},
		},
	}

	assert.Equal(t, TransactionPayloadVariantPayload, original.PayloadType())

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	deserialized := &TransactionInnerPayload{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	// Verify the inner payload is V1
	v1, ok := deserialized.Payload.(*TransactionInnerPayloadV1)
	require.True(t, ok)
	assert.Equal(t, TransactionInnerPayloadVariantV1, v1.InnerPayloadType())

	// Verify the executable is an EntryFunction
	efResult, ok := v1.Executable.Inner.(*EntryFunction)
	require.True(t, ok)
	assert.Equal(t, "transfer", efResult.Function)
	assert.Equal(t, "aptos_account", efResult.Module.Name)
	assert.Equal(t, AccountOne, efResult.Module.Address)
	assert.Equal(t, [][]byte{{0x01}}, efResult.Args)
}

func TestTransactionInnerPayload_NilPayload(t *testing.T) {
	t.Parallel()
	tip := &TransactionInnerPayload{}
	_, err := bcs.Serialize(tip)
	require.Error(t, err)
}

func TestTransactionExecutable_EntryFunction_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	ef := &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{{0x01}},
	}
	original := TransactionExecutable{Inner: ef}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionExecutable{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	efResult, ok := deserialized.Inner.(*EntryFunction)
	require.True(t, ok)
	assert.Equal(t, "transfer", efResult.Function)
	assert.Equal(t, "aptos_account", efResult.Module.Name)
}

func TestTransactionExecutable_Script_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	script := &Script{
		Code:     []byte{0xDE, 0xAD},
		ArgTypes: []TypeTag{},
		Args:     []ScriptArgument{},
	}
	original := TransactionExecutable{Inner: script}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionExecutable{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	scriptResult, ok := deserialized.Inner.(*Script)
	require.True(t, ok)
	assert.Equal(t, []byte{0xDE, 0xAD}, scriptResult.Code)
	assert.Empty(t, scriptResult.ArgTypes)
	assert.Empty(t, scriptResult.Args)
}

func TestTransactionExecutable_Empty_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	original := TransactionExecutable{Inner: &TransactionExecutableEmpty{}}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionExecutable{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	_, ok := deserialized.Inner.(*TransactionExecutableEmpty)
	require.True(t, ok)
	assert.Equal(t, TransactionExecutableVariantEmpty, deserialized.Inner.ExecutableType())
}

func TestTransactionExecutable_NilInner(t *testing.T) {
	t.Parallel()
	te := &TransactionExecutable{}
	_, err := bcs.Serialize(te)
	require.Error(t, err)
}

func TestTransactionExtraConfig_V1_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	nonce := uint64(99)
	addr := AccountThree
	original := TransactionExtraConfig{
		Inner: &TransactionExtraConfigV1{
			MultisigAddress:       &addr,
			ReplayProtectionNonce: &nonce,
		},
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionExtraConfig{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	v1, ok := deserialized.Inner.(*TransactionExtraConfigV1)
	require.True(t, ok)
	assert.Equal(t, TransactionExtraConfigVariantV1, v1.ConfigType())
}

func TestTransactionExtraConfig_V1_NilOptionals_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	original := TransactionExtraConfig{
		Inner: &TransactionExtraConfigV1{
			MultisigAddress:       nil,
			ReplayProtectionNonce: nil,
		},
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionExtraConfig{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	v1, ok := deserialized.Inner.(*TransactionExtraConfigV1)
	require.True(t, ok)
	assert.Equal(t, TransactionExtraConfigVariantV1, v1.ConfigType())
}

func TestTransactionExtraConfig_NilInner(t *testing.T) {
	t.Parallel()
	tec := &TransactionExtraConfig{}
	_, err := bcs.Serialize(tec)
	require.Error(t, err)
}
