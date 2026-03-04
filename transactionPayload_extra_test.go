package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionPayload_Script_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	script := &Script{
		Code:     []byte{0xDE, 0xAD},
		ArgTypes: []TypeTag{},
		Args:     []ScriptArgument{},
	}
	original := TransactionPayload{Payload: script}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionPayload{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	scriptResult, ok := deserialized.Payload.(*Script)
	require.True(t, ok)
	assert.Equal(t, []byte{0xDE, 0xAD}, scriptResult.Code)
	assert.Empty(t, scriptResult.ArgTypes)
	assert.Empty(t, scriptResult.Args)
}

func TestTransactionPayload_Multisig_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	original := TransactionPayload{
		Payload: &Multisig{
			MultisigAddress: AccountTwo,
			Payload:         nil,
		},
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionPayload{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	ms, ok := deserialized.Payload.(*Multisig)
	require.True(t, ok)
	assert.Equal(t, AccountTwo, ms.MultisigAddress)
	assert.Nil(t, ms.Payload)
}

func TestTransactionPayload_Multisig_WithPayload_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	ef := &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{{0x01}},
	}
	original := TransactionPayload{
		Payload: &Multisig{
			MultisigAddress: AccountThree,
			Payload: &MultisigTransactionPayload{
				Variant: MultisigTransactionPayloadVariantEntryFunction,
				Payload: ef,
			},
		},
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionPayload{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	ms, ok := deserialized.Payload.(*Multisig)
	require.True(t, ok)
	assert.Equal(t, AccountThree, ms.MultisigAddress)
	require.NotNil(t, ms.Payload)

	efResult, ok := ms.Payload.Payload.(*EntryFunction)
	require.True(t, ok)
	assert.Equal(t, "transfer", efResult.Function)
	assert.Equal(t, "aptos_account", efResult.Module.Name)
}

func TestTransactionPayload_Payload_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	ef := &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{{0x01}},
	}
	original := TransactionPayload{
		Payload: &TransactionInnerPayload{
			Payload: &TransactionInnerPayloadV1{
				Executable: TransactionExecutable{Inner: ef},
				ExtraConfig: TransactionExtraConfig{
					Inner: &TransactionExtraConfigV1{
						MultisigAddress:       nil,
						ReplayProtectionNonce: nil,
					},
				},
			},
		},
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &TransactionPayload{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	tip, ok := deserialized.Payload.(*TransactionInnerPayload)
	require.True(t, ok)

	v1, ok := tip.Payload.(*TransactionInnerPayloadV1)
	require.True(t, ok)

	efResult, ok := v1.Executable.Inner.(*EntryFunction)
	require.True(t, ok)
	assert.Equal(t, "transfer", efResult.Function)
}

func TestMultisigTransactionPayload_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	ef := &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "coin"},
		Function: "transfer",
		ArgTypes: []TypeTag{{Value: &U64Tag{}}},
		Args:     [][]byte{{0x01}, {0x02}},
	}
	original := MultisigTransactionPayload{
		Variant: MultisigTransactionPayloadVariantEntryFunction,
		Payload: ef,
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &MultisigTransactionPayload{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	efResult, ok := deserialized.Payload.(*EntryFunction)
	require.True(t, ok)
	assert.Equal(t, "transfer", efResult.Function)
	assert.Equal(t, "coin", efResult.Module.Name)
	require.Len(t, efResult.ArgTypes, 1)
	assert.Equal(t, TypeTagU64, efResult.ArgTypes[0].Value.GetType())
	assert.Equal(t, [][]byte{{0x01}, {0x02}}, efResult.Args)
}

func TestModuleBundle_MarshalBCS_Error(t *testing.T) {
	t.Parallel()
	mb := &ModuleBundle{}
	assert.Equal(t, TransactionPayloadVariantModuleBundle, mb.PayloadType())
	_, err := bcs.Serialize(mb)
	require.Error(t, err)
}

func TestModuleBundle_UnmarshalBCS_Error(t *testing.T) {
	t.Parallel()
	// Create bytes that would represent a ModuleBundle variant (1)
	ser := &bcs.Serializer{}
	ser.Uleb128(uint32(TransactionPayloadVariantModuleBundle))
	bytes := ser.ToBytes()

	tp := &TransactionPayload{}
	err := bcs.Deserialize(tp, bytes)
	require.Error(t, err)
}

func TestEntryFunction_BCS_RoundTrip(t *testing.T) {
	t.Parallel()
	original := EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "coin"},
		Function: "transfer",
		ArgTypes: []TypeTag{
			{Value: &U64Tag{}},
			{Value: &BoolTag{}},
		},
		Args: [][]byte{{0x01, 0x02}, {0x03}},
	}

	bytes, err := bcs.Serialize(&original)
	require.NoError(t, err)

	deserialized := &EntryFunction{}
	err = bcs.Deserialize(deserialized, bytes)
	require.NoError(t, err)

	assert.Equal(t, original.Function, deserialized.Function)
	assert.Equal(t, original.Module.Name, deserialized.Module.Name)
	assert.Equal(t, original.Module.Address, deserialized.Module.Address)
	require.Len(t, deserialized.ArgTypes, 2)
	assert.Equal(t, TypeTagU64, deserialized.ArgTypes[0].Value.GetType())
	assert.Equal(t, TypeTagBool, deserialized.ArgTypes[1].Value.GetType())
	assert.Equal(t, original.Args, deserialized.Args)
}

func TestTransactionPayload_NilPayload_Error(t *testing.T) {
	t.Parallel()
	tp := &TransactionPayload{}
	_, err := bcs.Serialize(tp)
	require.Error(t, err)
}

func TestTransactionPayload_UnknownVariant_Error(t *testing.T) {
	t.Parallel()
	ser := &bcs.Serializer{}
	ser.Uleb128(uint32(999))
	bytes := ser.ToBytes()

	tp := &TransactionPayload{}
	err := bcs.Deserialize(tp, bytes)
	require.Error(t, err)
}
