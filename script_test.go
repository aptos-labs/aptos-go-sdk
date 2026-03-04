package aptos

import (
	"math/big"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScript_PayloadType(t *testing.T) {
	t.Parallel()

	s := &Script{}
	assert.Equal(t, TransactionPayloadVariantScript, s.PayloadType())
}

func TestScript_ExecutableType(t *testing.T) {
	t.Parallel()

	s := &Script{}
	assert.Equal(t, TransactionExecutableVariantScript, s.ExecutableType())
}

func TestScript_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	script := &Script{
		Code: []byte{0xDE, 0xAD, 0xBE, 0xEF},
		ArgTypes: []TypeTag{
			{Value: &U8Tag{}},
			{Value: &BoolTag{}},
		},
		Args: []ScriptArgument{
			{Variant: ScriptArgumentU8, Value: uint8(42)},
			{Variant: ScriptArgumentBool, Value: true},
		},
	}

	serialized, err := bcs.Serialize(script)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	deserialized := &Script{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, script.Code, deserialized.Code)
	require.Len(t, deserialized.ArgTypes, 2)
	assert.Equal(t, TypeTagU8, deserialized.ArgTypes[0].Value.GetType())
	assert.Equal(t, TypeTagBool, deserialized.ArgTypes[1].Value.GetType())
	require.Len(t, deserialized.Args, 2)
	assert.Equal(t, ScriptArgumentU8, deserialized.Args[0].Variant)
	assert.Equal(t, uint8(42), deserialized.Args[0].Value)
	assert.Equal(t, ScriptArgumentBool, deserialized.Args[1].Variant)
	assert.Equal(t, true, deserialized.Args[1].Value)

	// Full round-trip byte comparison.
	reserialized, err := bcs.Serialize(deserialized)
	require.NoError(t, err)
	assert.Equal(t, serialized, reserialized)
}

func TestScriptArgument_U8_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sa := &ScriptArgument{Variant: ScriptArgumentU8, Value: uint8(42)}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU8, deserialized.Variant)
	assert.Equal(t, uint8(42), deserialized.Value)
}

func TestScriptArgument_U16_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sa := &ScriptArgument{Variant: ScriptArgumentU16, Value: uint16(1000)}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU16, deserialized.Variant)
	assert.Equal(t, uint16(1000), deserialized.Value)
}

func TestScriptArgument_U32_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sa := &ScriptArgument{Variant: ScriptArgumentU32, Value: uint32(100000)}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU32, deserialized.Variant)
	assert.Equal(t, uint32(100000), deserialized.Value)
}

func TestScriptArgument_U64_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sa := &ScriptArgument{Variant: ScriptArgumentU64, Value: uint64(1000000)}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU64, deserialized.Variant)
	assert.Equal(t, uint64(1000000), deserialized.Value)
}

func TestScriptArgument_U128_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	val := big.NewInt(0).SetUint64(999999999999999999)
	sa := &ScriptArgument{Variant: ScriptArgumentU128, Value: *val}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU128, deserialized.Variant)
	deserializedVal, ok := deserialized.Value.(big.Int)
	require.True(t, ok, "deserialized value should be big.Int")
	assert.Equal(t, 0, val.Cmp(&deserializedVal), "U128 values should match")
}

func TestScriptArgument_U256_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	val := big.NewInt(0).SetUint64(12345678901234567)
	sa := &ScriptArgument{Variant: ScriptArgumentU256, Value: *val}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU256, deserialized.Variant)
	deserializedVal, ok := deserialized.Value.(big.Int)
	require.True(t, ok, "deserialized value should be big.Int")
	assert.Equal(t, 0, val.Cmp(&deserializedVal), "U256 values should match")
}

func TestScriptArgument_Address_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	addr := AccountOne
	sa := &ScriptArgument{Variant: ScriptArgumentAddress, Value: addr}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentAddress, deserialized.Variant)
	deserializedAddr, ok := deserialized.Value.(AccountAddress)
	require.True(t, ok, "deserialized value should be AccountAddress")
	assert.Equal(t, addr, deserializedAddr)
}

func TestScriptArgument_U8Vector_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	sa := &ScriptArgument{Variant: ScriptArgumentU8Vector, Value: data}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentU8Vector, deserialized.Variant)
	deserializedBytes, ok := deserialized.Value.([]byte)
	require.True(t, ok, "deserialized value should be []byte")
	assert.Equal(t, data, deserializedBytes)
}

func TestScriptArgument_Bool_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	sa := &ScriptArgument{Variant: ScriptArgumentBool, Value: true}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentBool, deserialized.Variant)
	assert.Equal(t, true, deserialized.Value)
}

func TestScriptArgument_Serialized_BCS_RoundTrip(t *testing.T) {
	t.Parallel()

	serializedData := bcs.NewSerialized([]byte{0x01, 0x02, 0x03})
	sa := &ScriptArgument{Variant: ScriptArgumentSerialized, Value: serializedData}

	serialized, err := bcs.Serialize(sa)
	require.NoError(t, err)

	deserialized := &ScriptArgument{}
	err = bcs.Deserialize(deserialized, serialized)
	require.NoError(t, err)

	assert.Equal(t, ScriptArgumentSerialized, deserialized.Variant)
	deserializedSerialized, ok := deserialized.Value.(*bcs.Serialized)
	require.True(t, ok, "deserialized value should be *bcs.Serialized")
	assert.Equal(t, serializedData.Value, deserializedSerialized.Value)
}
