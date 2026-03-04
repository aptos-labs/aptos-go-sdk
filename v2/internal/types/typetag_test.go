package types

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeTag_Primitives(t *testing.T) {
	tests := []struct {
		tag     TypeTagImpl
		str     string
		variant TypeTagVariant
	}{
		{&BoolTag{}, "bool", TypeTagBool},
		{&U8Tag{}, "u8", TypeTagU8},
		{&U16Tag{}, "u16", TypeTagU16},
		{&U32Tag{}, "u32", TypeTagU32},
		{&U64Tag{}, "u64", TypeTagU64},
		{&U128Tag{}, "u128", TypeTagU128},
		{&U256Tag{}, "u256", TypeTagU256},
		{&AddressTag{}, "address", TypeTagAddress},
		{&SignerTag{}, "signer", TypeTagSigner},
	}

	for _, tc := range tests {
		t.Run(tc.str, func(t *testing.T) {
			assert.Equal(t, tc.str, tc.tag.String())
			assert.Equal(t, tc.variant, tc.tag.GetType())
		})
	}
}

func TestTypeTag_Vector(t *testing.T) {
	vec := &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}
	assert.Equal(t, "vector<u8>", vec.String())
	assert.Equal(t, TypeTagVector, vec.GetType())
}

func TestTypeTag_NestedVector(t *testing.T) {
	nested := &VectorTag{
		TypeParam: TypeTag{
			Value: &VectorTag{
				TypeParam: TypeTag{Value: &U8Tag{}},
			},
		},
	}
	assert.Equal(t, "vector<vector<u8>>", nested.String())
}

func TestTypeTag_StructTag(t *testing.T) {
	st := &StructTag{
		Address:    AccountOne,
		Module:     "coin",
		Name:       "Coin",
		TypeParams: nil,
	}
	assert.Equal(t, "0x1::coin::Coin", st.String())

	// With type params
	st.TypeParams = []TypeTag{{Value: &U8Tag{}}}
	assert.Equal(t, "0x1::coin::Coin<u8>", st.String())

	// Multiple type params
	st.TypeParams = []TypeTag{{Value: &U8Tag{}}, {Value: &U64Tag{}}}
	assert.Equal(t, "0x1::coin::Coin<u8,u64>", st.String())
}

func TestParseTypeTag_Primitives(t *testing.T) {
	primitives := []string{
		"bool", "u8", "u16", "u32", "u64", "u128", "u256",
		"address", "signer",
	}

	for _, p := range primitives {
		t.Run(p, func(t *testing.T) {
			tt, err := ParseTypeTag(p)
			require.NoError(t, err)
			assert.Equal(t, p, tt.String())
		})
	}
}

func TestParseTypeTag_Vector(t *testing.T) {
	tt, err := ParseTypeTag("vector<u8>")
	require.NoError(t, err)
	assert.Equal(t, "vector<u8>", tt.String())
}

func TestParseTypeTag_NestedVector(t *testing.T) {
	tt, err := ParseTypeTag("vector<vector<u64>>")
	require.NoError(t, err)
	assert.Equal(t, "vector<vector<u64>>", tt.String())
}

func TestParseTypeTag_Struct(t *testing.T) {
	tt, err := ParseTypeTag("0x1::string::String")
	require.NoError(t, err)
	assert.Equal(t, "0x1::string::String", tt.String())
}

func TestParseTypeTag_StructWithParams(t *testing.T) {
	tt, err := ParseTypeTag("0x1::coin::Coin<0x1::aptos_coin::AptosCoin>")
	require.NoError(t, err)
	assert.Equal(t, "0x1::coin::Coin<0x1::aptos_coin::AptosCoin>", tt.String())
}

func TestParseTypeTag_WithWhitespace(t *testing.T) {
	tt, err := ParseTypeTag("vector< u8 >")
	require.NoError(t, err)
	assert.Equal(t, "vector<u8>", tt.String())
}

func TestParseTypeTag_Reference(t *testing.T) {
	tt, err := ParseTypeTag("&u8")
	require.NoError(t, err)
	assert.Equal(t, "&u8", tt.String())
}

func TestParseTypeTag_Generic(t *testing.T) {
	tt, err := ParseTypeTag("T0")
	require.NoError(t, err)
	assert.Equal(t, "T0", tt.String())
}

func TestParseTypeTag_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"invalid",
		"0x1::module",     // Missing name
		"vector<>",        // Empty type param
		"vector<u8, u16>", // Multiple params for vector
		"bool<u8>",        // Primitive with params
		">>",              // Unmatched brackets
	}

	for _, s := range invalid {
		t.Run(s, func(t *testing.T) {
			_, err := ParseTypeTag(s)
			assert.Error(t, err)
		})
	}
}

func TestTypeTag_BCSRoundTrip(t *testing.T) {
	tags := []TypeTag{
		{Value: &BoolTag{}},
		{Value: &U8Tag{}},
		{Value: &U64Tag{}},
		{Value: &AddressTag{}},
		{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}},
		{Value: &StructTag{
			Address:    AccountOne,
			Module:     "coin",
			Name:       "Coin",
			TypeParams: []TypeTag{{Value: &U64Tag{}}},
		}},
	}

	for _, tt := range tags {
		t.Run(tt.String(), func(t *testing.T) {
			// Serialize
			data, err := bcs.Serialize(&tt)
			require.NoError(t, err)

			// Deserialize
			var tt2 TypeTag
			err = bcs.Deserialize(&tt2, data)
			require.NoError(t, err)

			assert.Equal(t, tt.String(), tt2.String())
		})
	}
}

func TestNewTypeTag_Helpers(t *testing.T) {
	// String tag
	stringTag := NewStringTag()
	assert.Equal(t, "0x1::string::String", stringTag.String())

	// Option tag
	optionTag := NewOptionTag(&U64Tag{})
	assert.Equal(t, "0x1::option::Option<u64>", optionTag.String())

	// Object tag
	objectTag := NewObjectTag(&BoolTag{})
	assert.Equal(t, "0x1::object::Object<bool>", objectTag.String())

	// Vector tag
	vectorTag := NewVectorTag(&U8Tag{})
	assert.Equal(t, "vector<u8>", vectorTag.String())
}

func TestAptosCoinTypeTag(t *testing.T) {
	assert.Equal(t, "0x1::aptos_coin::AptosCoin", AptosCoinTypeTag.String())
}

func TestTypeTag_AllPrimitiveBCSRoundTrip(t *testing.T) {
	t.Parallel()

	tags := []struct {
		name string
		tag  TypeTag
	}{
		{"bool", TypeTag{Value: &BoolTag{}}},
		{"u8", TypeTag{Value: &U8Tag{}}},
		{"u16", TypeTag{Value: &U16Tag{}}},
		{"u32", TypeTag{Value: &U32Tag{}}},
		{"u64", TypeTag{Value: &U64Tag{}}},
		{"u128", TypeTag{Value: &U128Tag{}}},
		{"u256", TypeTag{Value: &U256Tag{}}},
		{"address", TypeTag{Value: &AddressTag{}}},
		{"signer", TypeTag{Value: &SignerTag{}}},
	}

	for _, tc := range tags {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data, err := bcs.Serialize(&tc.tag)
			require.NoError(t, err)

			var result TypeTag
			err = bcs.Deserialize(&result, data)
			require.NoError(t, err)
			assert.Equal(t, tc.tag.String(), result.String())
			assert.Equal(t, tc.tag.Value.GetType(), result.Value.GetType())
		})
	}
}

func TestTypeTag_VectorBCSRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("vector<u8>", func(t *testing.T) {
		t.Parallel()
		tag := TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}}
		data, err := bcs.Serialize(&tag)
		require.NoError(t, err)

		var result TypeTag
		err = bcs.Deserialize(&result, data)
		require.NoError(t, err)
		assert.Equal(t, "vector<u8>", result.String())
	})

	t.Run("nested vector<vector<u8>>", func(t *testing.T) {
		t.Parallel()
		inner := TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}}
		tag := TypeTag{Value: &VectorTag{TypeParam: inner}}
		data, err := bcs.Serialize(&tag)
		require.NoError(t, err)

		var result TypeTag
		err = bcs.Deserialize(&result, data)
		require.NoError(t, err)
		assert.Equal(t, "vector<vector<u8>>", result.String())
	})
}

func TestTypeTag_StructBCSRoundTripWithTypeParams(t *testing.T) {
	t.Parallel()

	tag := TypeTag{Value: &StructTag{
		Address: AccountOne,
		Module:  "coin",
		Name:    "Coin",
		TypeParams: []TypeTag{
			{Value: &StructTag{
				Address: AccountOne,
				Module:  "aptos_coin",
				Name:    "AptosCoin",
			}},
		},
	}}

	data, err := bcs.Serialize(&tag)
	require.NoError(t, err)

	var result TypeTag
	err = bcs.Deserialize(&result, data)
	require.NoError(t, err)
	assert.Equal(t, tag.String(), result.String())
}

func TestReferenceTag_StringAndGetType(t *testing.T) {
	t.Parallel()

	ref := &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}
	assert.Equal(t, "&u8", ref.String())
	assert.Equal(t, TypeTagReference, ref.GetType())
}

func TestGenericTag_StringAndGetType(t *testing.T) {
	t.Parallel()

	t.Run("T0", func(t *testing.T) {
		t.Parallel()
		g := &GenericTag{Num: 0}
		assert.Equal(t, "T0", g.String())
		assert.Equal(t, TypeTagGeneric, g.GetType())
	})

	t.Run("T3", func(t *testing.T) {
		t.Parallel()
		g := &GenericTag{Num: 3}
		assert.Equal(t, "T3", g.String())
		assert.Equal(t, TypeTagGeneric, g.GetType())
	})
}

func TestTypeTag_UnmarshalBCS_UnknownVariant(t *testing.T) {
	t.Parallel()

	// Create data with an unknown variant number (99)
	ser := bcs.NewSerializer()
	ser.Uleb128(99)
	data := ser.ToBytes()

	var tag TypeTag
	err := bcs.Deserialize(&tag, data)
	assert.Error(t, err)
}

func TestTypeTag_SignedIntegerBCSRoundTrip(t *testing.T) {
	t.Parallel()

	tags := []struct {
		name string
		tag  TypeTag
	}{
		{"i8", TypeTag{Value: &I8Tag{}}},
		{"i16", TypeTag{Value: &I16Tag{}}},
		{"i32", TypeTag{Value: &I32Tag{}}},
		{"i64", TypeTag{Value: &I64Tag{}}},
		{"i128", TypeTag{Value: &I128Tag{}}},
		{"i256", TypeTag{Value: &I256Tag{}}},
	}

	for _, tc := range tags {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data, err := bcs.Serialize(&tc.tag)
			require.NoError(t, err)

			var result TypeTag
			err = bcs.Deserialize(&result, data)
			require.NoError(t, err)
			assert.Equal(t, tc.tag.String(), result.String())
			assert.Equal(t, tc.tag.Value.GetType(), result.Value.GetType())
		})
	}
}

func TestPrimitiveTag_DirectMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tag     TypeTagImpl
		str     string
		variant TypeTagVariant
	}{
		{&BoolTag{}, "bool", TypeTagBool},
		{&U8Tag{}, "u8", TypeTagU8},
		{&U16Tag{}, "u16", TypeTagU16},
		{&U32Tag{}, "u32", TypeTagU32},
		{&U64Tag{}, "u64", TypeTagU64},
		{&U128Tag{}, "u128", TypeTagU128},
		{&U256Tag{}, "u256", TypeTagU256},
		{&I8Tag{}, "i8", TypeTagI8},
		{&I16Tag{}, "i16", TypeTagI16},
		{&I32Tag{}, "i32", TypeTagI32},
		{&I64Tag{}, "i64", TypeTagI64},
		{&I128Tag{}, "i128", TypeTagI128},
		{&I256Tag{}, "i256", TypeTagI256},
		{&AddressTag{}, "address", TypeTagAddress},
		{&SignerTag{}, "signer", TypeTagSigner},
		{&ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}, "&u8", TypeTagReference},
		{&GenericTag{Num: 0}, "T0", TypeTagGeneric},
	}

	for _, tc := range tests {
		ser := bcs.NewSerializer()
		tc.tag.MarshalBCS(ser)
		require.NoError(t, ser.Error(), "MarshalBCS failed for %s", tc.str)

		des := bcs.NewDeserializer(nil)
		tc.tag.UnmarshalBCS(des)
		require.NoError(t, des.Error(), "UnmarshalBCS failed for %s", tc.str)

		assert.Equal(t, tc.str, tc.tag.String())
		assert.Equal(t, tc.variant, tc.tag.GetType())
	}
}
