package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
)

func TestTypeTag(t *testing.T) {
	// This is unfortunate with references
	nested := NewTypeTag(NewOptionTag(NewVectorTag(NewObjectTag(NewStringTag()))))

	assert.Equal(t, "0x1::option::Option<vector<0x1::object::Object<0x1::string::String>>>", nested.String())

	bytes, err := bcs.Serialize(&nested)
	assert.NoError(t, err)

	des := bcs.NewDeserializer(bytes)
	tag := &TypeTag{}
	des.Struct(tag)
	assert.NoError(t, des.Error())

	// Check the deserialized is correct
	assert.Equal(t, &nested, tag)
}

func TestTypeTagIdentities(t *testing.T) {
	checkVariant(t, &AddressTag{}, TypeTagAddress, "address")
	checkVariant(t, &SignerTag{}, TypeTagSigner, "signer")
	checkVariant(t, &BoolTag{}, TypeTagBool, "bool")
	checkVariant(t, &U8Tag{}, TypeTagU8, "u8")
	checkVariant(t, &U16Tag{}, TypeTagU16, "u16")
	checkVariant(t, &U32Tag{}, TypeTagU32, "u32")
	checkVariant(t, &U64Tag{}, TypeTagU64, "u64")
	checkVariant(t, &U128Tag{}, TypeTagU128, "u128")
	checkVariant(t, &U256Tag{}, TypeTagU256, "u256")

	checkVariant(t, NewVectorTag(&U8Tag{}), TypeTagVector, "vector<u8>")
	checkVariant(t, NewStringTag(), TypeTagStruct, "0x1::string::String")
}

func checkVariant[T TypeTagImpl](t *testing.T, tag T, expectedType TypeTagVariant, expectedString string) {
	t.Helper()
	assert.Equal(t, expectedType, tag.GetType())
	assert.Equal(t, expectedString, tag.String())

	// Serialize and deserialize test
	typeTag := NewTypeTag(tag)
	bytes, err := bcs.Serialize(&typeTag)
	assert.NoError(t, err)
	var newTag TypeTag
	err = bcs.Deserialize(&newTag, bytes)
	assert.NoError(t, err)
	assert.Equal(t, typeTag, newTag)
}

func TestStructTag(t *testing.T) {
	structTag := StructTag{
		Address: AccountOne,
		Module:  "coin",
		Name:    "CoinStore",
		TypeParams: []TypeTag{
			{Value: &StructTag{
				Address:    AccountOne,
				Module:     "aptos_coin",
				Name:       "AptosCoin",
				TypeParams: nil,
			}},
		},
	}
	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", structTag.String())
	var aa3 AccountAddress
	err := aa3.ParseStringRelaxed("0x3")
	assert.NoError(t, err)

	structTag.TypeParams = append(structTag.TypeParams, TypeTag{Value: &StructTag{
		Address:    aa3,
		Module:     "other",
		Name:       "thing",
		TypeParams: nil,
	}})
	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin,0x3::other::thing>", structTag.String())
}

func TestInvalidTypeTag(t *testing.T) {
	serializer := &bcs.Serializer{}
	serializer.Uleb128(uint32(65535))
	bytes := serializer.ToBytes()
	tag := &TypeTag{}
	err := bcs.Deserialize(tag, bytes)
	assert.Error(t, err)
}

func TestParseTypeTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *TypeTag
		wantErr  bool
	}{
		// Invalid cases
		{"empty string", "", nil, true},
		{"invalid type", "invalid", nil, true},
		{"unclosed vector", "vector<", nil, true},
		{"unopened vector", "vector>", nil, true},
		{"incomplete address", "0x1::string", nil, true},
		{"incomplete module", "0x1::string::", nil, true},
		{"invalid address format", "0x1::::String", nil, true},
		{"missing address", "::string::String", nil, true},
		{"invalid hex address", "dead::string::String", nil, true},
		{"unclosed generic", "0x1::string::String<", nil, true},
		{"unopened generic", "0x1::string::String>", nil, true},
		{"empty generic", "0x1::string::String<>", nil, true},
		{"incomplete generic", "0x1::string::String<u8,", nil, true},
		{"trailing comma", "0x1::string::String<u8,>", nil, true},

		// Primitive types
		{"bool type", "bool", &TypeTag{Value: &BoolTag{}}, false},
		{"u8 type", "u8", &TypeTag{Value: &U8Tag{}}, false},
		{"u16 type", "u16", &TypeTag{Value: &U16Tag{}}, false},
		{"u32 type", "u32", &TypeTag{Value: &U32Tag{}}, false},
		{"u64 type", "u64", &TypeTag{Value: &U64Tag{}}, false},
		{"u128 type", "u128", &TypeTag{Value: &U128Tag{}}, false},
		{"u256 type", "u256", &TypeTag{Value: &U256Tag{}}, false},
		{"address type", "address", &TypeTag{Value: &AddressTag{}}, false},
		{"signer type", "signer", &TypeTag{Value: &SignerTag{}}, false},

		// Handle references
		{"signer reference", "&signer", &TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &SignerTag{}}}}, false},
		{"u8 reference", "&u8", &TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}}, false},

		// Vector types
		{"simple vector", "vector<u8>", &TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}}, false},
		{"nested vector", "vector<vector<u8>>", &TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}}}}, false},
		{"vector of string", "vector<0x1::string::String>", &TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String", TypeParams: []TypeTag{}}}}}, false},

		// Struct types
		{"simple string", "0x1::string::String", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String", TypeParams: []TypeTag{}}}, false},
		{"nested object", "0x1::object::Object<0x1::object::ObjectCore>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "object", Name: "Object", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "object", Name: "ObjectCore", TypeParams: []TypeTag{}}}}}}, false},
		{"option type", "0x1::option::Option<u8>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &U8Tag{}}}}}, false},
		{"coin store", "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "CoinStore", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "aptos_coin", Name: "AptosCoin", TypeParams: []TypeTag{}}}}}}, false},
		{"coin store with multiple params", "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin,0x3::other::thing>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "CoinStore", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "aptos_coin", Name: "AptosCoin", TypeParams: []TypeTag{}}}, {Value: &StructTag{Address: AccountThree, Module: "other", Name: "thing", TypeParams: []TypeTag{}}}}}}, false},

		// Complex nested types
		{"complex nested type", "vector<0x1::option::Option<vector<0x1::object::Object<0x1::object::ObjectCore>>>>", &TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "object", Name: "Object", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "object", Name: "ObjectCore", TypeParams: []TypeTag{}}}}}}}}}}}}}, false},

		// Generic type parameters
		{"generic coin", "0x1::coin::Coin<0x1::aptos_coin::AptosCoin>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "Coin", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "aptos_coin", Name: "AptosCoin", TypeParams: []TypeTag{}}}}}}, false},
		{"generic type", "0x1::coin::Coin<T0>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "Coin", TypeParams: []TypeTag{{Value: &GenericTag{Num: 0}}}}}, false},
		{"generic 2 type", "0x1::pair::Pair<T0,T1>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "pair", Name: "Pair", TypeParams: []TypeTag{{Value: &GenericTag{Num: 0}}, {Value: &GenericTag{Num: 1}}}}}, false},

		// Multiple generic parameters
		{"pair with coins", "0x1::pair::Pair<0x1::coin::Coin<0x1::aptos_coin::AptosCoin>, 0x1::coin::Coin<0x1::usd_coin::UsdCoin>>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "pair", Name: "Pair", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "Coin", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "aptos_coin", Name: "AptosCoin", TypeParams: []TypeTag{}}}}}}, {Value: &StructTag{Address: AccountOne, Module: "coin", Name: "Coin", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "usd_coin", Name: "UsdCoin", TypeParams: []TypeTag{}}}}}}}}}, false},

		// Reference type parameters
		{"vector with reference", "0x1::vector::Vector<&0x1::coin::Coin<0x1::aptos_coin::AptosCoin>>", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "vector", Name: "Vector", TypeParams: []TypeTag{{Value: &ReferenceTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "Coin", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "aptos_coin", Name: "AptosCoin", TypeParams: []TypeTag{}}}}}}}}}}}, false},

		// Mutable reference type parameters
		// &mut not supported atm
		{"vector with mutable reference", "0x1::vector::Vector<&mut 0x1::coin::Coin<0x1::aptos_coin::AptosCoin>>", nil, true},

		// Whitespace handling
		{"spaces in type parameters", "0x1::vector::Vector< 0x1::coin::Coin< 0x1::aptos_coin::AptosCoin > >", &TypeTag{Value: &StructTag{Address: AccountOne, Module: "vector", Name: "Vector", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "coin", Name: "Coin", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "aptos_coin", Name: "AptosCoin", TypeParams: []TypeTag{}}}}}}}}}, false},

		// Invalid cases
		{"unclosed generic", "0x1::vector::Vector<0x1::coin::Coin<0x1::aptos_coin::AptosCoin", nil, true},
		{"unclosed reference", "0x1::vector::Vector<&0x1::coin::Coin<0x1::aptos_coin::AptosCoin", nil, true},
		{"unclosed mutable reference", "0x1::vector::Vector<&mut 0x1::coin::Coin<0x1::aptos_coin::AptosCoin", nil, true},
		{"unclosed spaces", "0x1::vector::Vector< 0x1::coin::Coin< 0x1::aptos_coin::AptosCoin >", nil, true},
		{"unclosed newlines", "0x1::vector::Vector<\n0x1::coin::Coin<\n0x1::aptos_coin::AptosCoin\n>", nil, true},
		{"unclosed multiple params", "0x1::pair::Pair< 0x1::coin::Coin< 0x1::aptos_coin::AptosCoin >, 0x1::coin::Coin< 0x1::usd_coin::UsdCoin", nil, true},
		{"invalid comma before", ",u8", nil, true},
		{"invalid comma after", "u8,", nil, true},
		{"invalid comma in generics before", "0x1::pair::Pair<,u8>", nil, true},
		{"invalid comma in generics after", "0x1::pair::Pair<u8,>", nil, true},
		{"invalid comma in generics only", "0x1::pair::Pair<,>", nil, true},
		{"invalid type in generics before comma", "0x1::pair::Pair<what,u8>", nil, true},
		{"invalid type in generics after comma", "0x1::pair::Pair<u8,what>", nil, true},
		{"invalid type space before comma", " ,", nil, true},
		{"invalid type space after comma", ", ", nil, true},
		{"invalid type space before close angle bracket", " >", nil, true},
		{"invalid type space before open angle bracket", " <", nil, true},
		{"invalid pair", "u8,u8", nil, true},
		{"invalid bool generic", "bool<T>", nil, true},
		{"invalid address generic", "address<T>", nil, true},
		{"invalid u8 generic", "u8<T>", nil, true},
		{"invalid u16 generic", "u16<T>", nil, true},
		{"invalid u32 generic", "u32<T>", nil, true},
		{"invalid u64 generic", "u64<T>", nil, true},
		{"invalid u128 generic", "u128<T>", nil, true},
		{"invalid u256 generic", "u256<T>", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeTag(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypeTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("ParseTypeTag() returned nil without error")
				return
			}
			if tt.wantErr {
				return
			}
			if got.String() != tt.expected.String() {
				t.Errorf("ParseTypeTag() = %v, want %v", got.String(), tt.expected.String())
			}
		})
	}
}
