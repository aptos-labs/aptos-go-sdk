package aptos

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ... existing code ...

func TestConvertTypeTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		wantErr  bool
		validate func(*testing.T, *TypeTag)
	}{
		{
			name:    "TypeTag value",
			input:   TypeTag{Value: &U8Tag{}},
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				if _, ok := tag.Value.(*U8Tag); !ok {
					t.Error("Expected U8Tag")
				}
			},
		},
		{
			name:    "TypeTag pointer",
			input:   &TypeTag{Value: &U8Tag{}},
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				if _, ok := tag.Value.(*U8Tag); !ok {
					t.Error("Expected U8Tag")
				}
			},
		},
		{
			name:    "nil TypeTag pointer",
			input:   (*TypeTag)(nil),
			wantErr: true,
		},
		{
			name:    "string type tag",
			input:   "u8",
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				if _, ok := tag.Value.(*U8Tag); !ok {
					t.Error("Expected U8Tag")
				}
			},
		},
		{
			name:    "vector type tag string",
			input:   "vector<u8>",
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				vecTag, ok := tag.Value.(*VectorTag)
				if !ok {
					t.Error("Expected VectorTag")
				}
				if _, ok := vecTag.TypeParam.Value.(*U8Tag); !ok {
					t.Error("Expected U8Tag in VectorTag")
				}
			},
		},
		{
			name:    "reference type tag string",
			input:   "&u8",
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				refTag, ok := tag.Value.(*ReferenceTag)
				if !ok {
					t.Error("Expected ReferenceTag")
				}
				if _, ok := refTag.TypeParam.Value.(*U8Tag); !ok {
					t.Error("Expected U8Tag in ReferenceTag")
				}
			},
		},
		{
			name:    "generic type tag string",
			input:   "T0",
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				genTag, ok := tag.Value.(*GenericTag)
				if !ok {
					t.Error("Expected GenericTag")
				}
				if genTag.Num != 0 {
					t.Error("Expected generic number 0")
				}
			},
		},
		{
			name:    "struct type tag string",
			input:   "0x1::string::String",
			wantErr: false,
			validate: func(t *testing.T, tag *TypeTag) {
				t.Helper()
				structTag, ok := tag.Value.(*StructTag)
				if !ok {
					t.Error("Expected StructTag")
				}
				if structTag.Address != AccountOne {
					t.Error("Expected AccountOne address")
				}
				if structTag.Module != "string" {
					t.Error("Expected string module")
				}
				if structTag.Name != "String" {
					t.Error("Expected String name")
				}
			},
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
		{
			name:    "invalid string type tag",
			input:   "invalid_type",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertTypeTag(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTypeTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

// ... existing code ...

func TestConvertArg(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		typeArg  TypeTag
		arg      any
		generics []TypeTag
		expected []byte
		wantErr  bool
	}{
		{
			name:     "u8",
			typeArg:  TypeTag{Value: &U8Tag{}},
			arg:      uint8(42),
			expected: []byte{42},
			wantErr:  false,
		},
		{
			name:     "u16",
			typeArg:  TypeTag{Value: &U16Tag{}},
			arg:      uint16(42),
			expected: []byte{42, 0},
			wantErr:  false,
		},
		{
			name:     "u32",
			typeArg:  TypeTag{Value: &U32Tag{}},
			arg:      uint32(42),
			expected: []byte{42, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "u64",
			typeArg:  TypeTag{Value: &U64Tag{}},
			arg:      uint64(42),
			expected: []byte{42, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "u128",
			typeArg:  TypeTag{Value: &U128Tag{}},
			arg:      big.NewInt(42),
			expected: []byte{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "u256",
			typeArg:  TypeTag{Value: &U256Tag{}},
			arg:      big.NewInt(42),
			expected: []byte{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "i8 positive",
			typeArg:  TypeTag{Value: &I8Tag{}},
			arg:      int8(42),
			expected: []byte{42},
			wantErr:  false,
		},
		{
			name:     "i8 negative",
			typeArg:  TypeTag{Value: &I8Tag{}},
			arg:      int8(-1),
			expected: []byte{0xff},
			wantErr:  false,
		},
		{
			name:     "i16 positive",
			typeArg:  TypeTag{Value: &I16Tag{}},
			arg:      int16(1000),
			expected: []byte{0xe8, 0x03},
			wantErr:  false,
		},
		{
			name:     "i16 negative",
			typeArg:  TypeTag{Value: &I16Tag{}},
			arg:      int16(-1),
			expected: []byte{0xff, 0xff},
			wantErr:  false,
		},
		{
			name:     "i32 positive",
			typeArg:  TypeTag{Value: &I32Tag{}},
			arg:      int32(100000),
			expected: []byte{0xa0, 0x86, 0x01, 0x00},
			wantErr:  false,
		},
		{
			name:     "i32 negative",
			typeArg:  TypeTag{Value: &I32Tag{}},
			arg:      int32(-1),
			expected: []byte{0xff, 0xff, 0xff, 0xff},
			wantErr:  false,
		},
		{
			name:     "i64 positive",
			typeArg:  TypeTag{Value: &I64Tag{}},
			arg:      int64(100000000000),
			expected: []byte{0x00, 0xe8, 0x76, 0x48, 0x17, 0x00, 0x00, 0x00},
			wantErr:  false,
		},
		{
			name:     "i64 negative",
			typeArg:  TypeTag{Value: &I64Tag{}},
			arg:      int64(-1),
			expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantErr:  false,
		},
		{
			name:     "i128 positive",
			typeArg:  TypeTag{Value: &I128Tag{}},
			arg:      big.NewInt(42),
			expected: []byte{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "i128 negative",
			typeArg:  TypeTag{Value: &I128Tag{}},
			arg:      big.NewInt(-1),
			expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantErr:  false,
		},
		{
			name:     "i256 positive",
			typeArg:  TypeTag{Value: &I256Tag{}},
			arg:      big.NewInt(42),
			expected: []byte{42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "i256 negative",
			typeArg:  TypeTag{Value: &I256Tag{}},
			arg:      big.NewInt(-1),
			expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantErr:  false,
		},
		{
			name:     "bool",
			typeArg:  TypeTag{Value: &BoolTag{}},
			arg:      true,
			expected: []byte{0x01},
			wantErr:  false,
		},
		{
			name:     "address",
			typeArg:  TypeTag{Value: &AddressTag{}},
			arg:      AccountOne,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01},
			wantErr:  false,
		},
		{
			name:     "vector<u8> from bytes",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}},
			arg:      []byte{0x42},
			expected: []byte{0x1, 0x42},
			wantErr:  false,
		},
		{
			name:     "option",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &U8Tag{}}}}},
			arg:      uint8(0x42),
			expected: []byte{0x1, 0x42},
			wantErr:  false,
		},
		{
			name:     "option none",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &U8Tag{}}}}},
			arg:      nil,
			expected: []byte{0x0},
			wantErr:  false,
		},
		{
			name:     "option<string> some",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}}},
			arg:      "hello",
			expected: []byte{0x1, 5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o')},
			wantErr:  false,
		},
		{
			name:     "option<string> none",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}}},
			arg:      nil,
			expected: []byte{0x0},
			wantErr:  false,
		},
		{
			name:     "option<vector<string>> some",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}}}}},
			arg:      []string{"hello", "goodbye"},
			expected: []byte{0x1, 2, 5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o'), 7, byte('g'), byte('o'), byte('o'), byte('d'), byte('b'), byte('y'), byte('e')},
			wantErr:  false,
		},
		{
			name:     "option<vector<string>> none",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}}}}},
			arg:      nil,
			expected: []byte{0x0},
			wantErr:  false,
		},
		{
			name:     "vector<option<u8>> mixed",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &U8Tag{}}}}}}},
			arg:      []any{nil, uint8(2)},
			expected: []byte{2, 0, 1, 2},
			wantErr:  false,
		},
		{
			name:     "vector<option<string>> mixed",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}}}}},
			arg:      []any{nil, "hello"},
			expected: []byte{2, 0, 1, 5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o')},
			wantErr:  false,
		},
		{
			name:     "string",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}},
			arg:      "hello",
			expected: []byte{5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o')},
			wantErr:  false,
		},
		{
			name:     "reference",
			typeArg:  TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}},
			arg:      uint8(0x42),
			expected: []byte{0x42},
			wantErr:  false,
		},
		{
			name:     "generic",
			typeArg:  TypeTag{Value: &GenericTag{Num: 0}},
			arg:      uint8(0x42),
			expected: []byte{0x42},
			generics: []TypeTag{
				{Value: &U8Tag{}},
			},
			wantErr: false,
		},
		{
			name:    "generic out of bounds",
			typeArg: TypeTag{Value: &GenericTag{Num: 1}},
			arg:     uint8(42),
			generics: []TypeTag{
				{Value: &U8Tag{}},
			},
			wantErr: true,
		},
		{
			name:     "object",
			typeArg:  TypeTag{Value: &StructTag{Address: AccountOne, Module: "object", Name: "Object"}},
			arg:      AccountOne,
			expected: AccountOne[:],
			wantErr:  false,
		},
		{
			name:    "invalid type",
			typeArg: TypeTag{Value: &U8Tag{}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:    "invalid option type",
			typeArg: TypeTag{Value: &StructTag{Address: AccountOne, Module: "option", Name: "Option", TypeParams: []TypeTag{{Value: &U8Tag{}}}}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:    "invalid string type",
			typeArg: TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}},
			arg:     42,
			wantErr: true,
		},
		{
			name:    "invalid reference type",
			typeArg: TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:    "invalid generic type",
			typeArg: TypeTag{Value: &GenericTag{Num: 0}},
			arg:     "invalid",
			generics: []TypeTag{
				{Value: &U8Tag{}},
			},
			wantErr: true,
		},
		{
			name:    "invalid object type",
			typeArg: TypeTag{Value: &StructTag{Address: AccountOne, Module: "object", Name: "Object"}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:     "vector<u64>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U64Tag{}}}},
			arg:      []uint64{0, 1, 2},
			expected: []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<bool>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &BoolTag{}}}},
			arg:      []bool{true, false, true},
			expected: []byte{3, 1, 0, 1},
			wantErr:  false,
		},
		{
			name:     "vector<address>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &AddressTag{}}}},
			arg:      []AccountAddress{AccountOne, AccountTwo},
			expected: []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			wantErr:  false,
		},
		{
			name:     "vector<address> with pointers",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &AddressTag{}}}},
			arg:      []*AccountAddress{&AccountOne, &AccountTwo},
			expected: []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			wantErr:  false,
		},
		{
			name:     "vector<generic>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &GenericTag{Num: 0}}}},
			arg:      []any{uint8(42), uint8(43)},
			generics: []TypeTag{{Value: &U8Tag{}}},
			expected: []byte{2, 42, 43},
			wantErr:  false,
		},
		{
			name:     "vector<&u8>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}}}},
			arg:      []any{uint8(42), uint8(43)},
			expected: []byte{2, 42, 43},
			wantErr:  false,
		},
		{
			name:     "vector<u8> with various types",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}},
			arg:      []any{uint8(0), uint(1), byte(2), 3, "4", big.NewInt(5), *big.NewInt(6)},
			expected: []byte{7, 0, 1, 2, 3, 4, 5, 6},
			wantErr:  false,
		},
		{
			name:     "vector<u16> with various types",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U16Tag{}}}},
			arg:      []any{uint16(0), uint(1), 2, 3, "4", big.NewInt(5), *big.NewInt(6)},
			expected: []byte{7, 0, 0, 1, 0, 2, 0, 3, 0, 4, 0, 5, 0, 6, 0},
			wantErr:  false,
		},
		{
			name:     "vector<u32> with various types",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U32Tag{}}}},
			arg:      []any{uint32(0), uint(1), 2, 3, "4", big.NewInt(5), *big.NewInt(6)},
			expected: []byte{7, 0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 0, 0, 0, 5, 0, 0, 0, 6, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<u64> with various types",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U64Tag{}}}},
			arg:      []any{uint64(0), uint(1), 2, 3, "4", big.NewInt(5), *big.NewInt(6)},
			expected: []byte{7, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<u128> with various types",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U128Tag{}}}},
			arg:      []any{0, uint(1), 2, 3, "4", big.NewInt(5), *big.NewInt(6)},
			expected: []byte{7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<u256> with big.Int",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U256Tag{}}}},
			arg:      []big.Int{*big.NewInt(2)},
			expected: []byte{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<u256> with various types",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U256Tag{}}}},
			arg:      []any{0, uint(1), 2, 3, "4", big.NewInt(5), *big.NewInt(6)},
			expected: []byte{7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<vector<u8>>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &U8Tag{}}}}}},
			arg:      []any{[]any{}, []any{1}, []int{1, 2, 3}, []uint{1, 2, 3}, []uint{}, []int{}},
			expected: []byte{6, 0, 1, 1, 3, 1, 2, 3, 3, 1, 2, 3, 0, 0},
			wantErr:  false,
		},
		{
			name:     "vector<vector<vector<u8>>>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &VectorTag{TypeParam: TypeTag{&U8Tag{}}}}}}}},
			arg:      []any{[]any{[]any{1}, []any{2, 3}}, []any{[]any{4, 5}, []any{}}},
			expected: []byte{2, 2, 1, 1, 2, 2, 3, 2, 2, 4, 5, 0},
			wantErr:  false,
		},
		{
			name:     "vector<0x1::string::String>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}},
			arg:      []string{"hello", "goodbye"},
			expected: []byte{2, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 7, 0x67, 0x6f, 0x6f, 0x64, 0x62, 0x79, 0x65},
		},
		{
			name:     "vector<0x1::string::String> with []any",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &StructTag{Address: AccountOne, Module: "string", Name: "String"}}}},
			arg:      []any{"hello", "goodbye"},
			expected: []byte{2, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 7, 0x67, 0x6f, 0x6f, 0x64, 0x62, 0x79, 0x65},
		},
		{
			name:    "nil vector<bool>",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &BoolTag{}}}},
			arg:     []bool(nil),
			wantErr: true,
		},
		{
			name:    "nil vector<address>",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &AddressTag{}}}},
			arg:     []AccountAddress(nil),
			wantErr: true,
		},
		{
			name:    "nil vector<address> with pointers",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &AddressTag{}}}},
			arg:     []*AccountAddress(nil),
			wantErr: true,
		},
		{
			name:     "nil vector<generic>",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &GenericTag{Num: 0}}}},
			arg:      []any(nil),
			generics: []TypeTag{{Value: &U8Tag{}}},
			wantErr:  true,
		},
		{
			name:    "nil vector<reference>",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}}}},
			arg:     []any(nil),
			wantErr: true,
		},
		{
			name:    "invalid vector<bool> type",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &BoolTag{}}}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:    "invalid vector<address> type",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &AddressTag{}}}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:     "invalid vector<generic> type",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &GenericTag{Num: 0}}}},
			arg:      "invalid",
			generics: []TypeTag{{Value: &U8Tag{}}},
			wantErr:  true,
		},
		{
			name:    "invalid vector<reference> type",
			typeArg: TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &ReferenceTag{TypeParam: TypeTag{Value: &U8Tag{}}}}}},
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:     "generic vector out of bounds",
			typeArg:  TypeTag{Value: &VectorTag{TypeParam: TypeTag{Value: &GenericTag{Num: 1}}}},
			arg:      []any{uint8(42)},
			generics: []TypeTag{{Value: &U8Tag{}}},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val, err := ConvertArg(tt.typeArg, tt.arg, tt.generics)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertArg() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equalf(t, tt.expected, val, "ConvertArg() failed to convert to correct bytes %v != %v", tt.expected, val)
		})
	}
}

func TestConvertArg_Special(t *testing.T) {
	t.Parallel()
	type Test struct {
		strTag            string
		arg               any
		generics          []TypeTag
		wantErr           bool
		expected          []byte
		compatibilityMode bool
	}
	tests := []Test{
		{
			strTag:   "0x1::option::Option<vector<vector<vector<u8>>>>",
			arg:      nil,
			expected: []byte{0},
		},
		{
			strTag:   "0x1::option::Option<vector<vector<vector<u8>>>>",
			arg:      []any{[]any{[]any{}, []any{22}}, []any{}, []any{[]any{42}}},
			expected: []byte{1, 3, 2, 0, 1, 22, 0, 1, 1, 42},
		},
		{
			strTag:   "vector<vector<vector<bool>>>",
			arg:      []any{[]any{[]any{}, []any{false}}, []any{}, []any{[]any{true}}},
			expected: []byte{3, 2, 0, 1, 0, 0, 1, 1, 1},
		},
		{
			strTag:   "vector<vector<vector<u16>>>",
			arg:      []any{[]any{[]any{}, []any{22}}, []any{}, []any{[]any{42}}},
			expected: []byte{3, 2, 0, 1, 22, 0, 0, 1, 1, 0x2a, 0},
		},
		{
			strTag:   "vector<vector<vector<u8>>>",
			arg:      []any{[]any{"0x4222"}, []any{}, []string{"0x32"}},
			expected: []byte{3, 1, 6, 48, 120, 52, 50, 50, 50, 0, 1, 4, 48, 120, 51, 50}, // "0x4222" and "0x32" as bytes
		},
		{ // Special case, difference in behavior with compatibility mode
			strTag:            "0x1::option::Option<signer>",
			arg:               "0x00",
			expected:          []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			compatibilityMode: false,
		},
		{ // Special case in compatibility mode
			strTag:            "0x1::option::Option<signer>",
			arg:               "0x00",
			expected:          []byte{0}, // Option with "0x00" as None
			compatibilityMode: true,
		},
		{ // Special case in compatibility mode
			strTag:            "0x1::option::Option<vector<u8>>",
			arg:               "0x00",
			expected:          []byte{0}, // Option with "0x00" as None
			compatibilityMode: true,
		},
		{ // Special case in compatibility mode
			strTag:            "0x1::option::Option<vector<u8>>",
			arg:               "0x0100",
			expected:          []byte{1, 0}, // Option with "0x0100" as Some([]byte{})
			compatibilityMode: true,
		},
		{ // Special case in compatibility mode
			strTag:            "0x1::option::Option<vector<u8>>",
			arg:               "0x010102",
			expected:          []byte{1, 1, 2}, // Option with "0x010102" as Some([]byte{2})
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<0x1::string::String>",
			arg:               "0x00",
			expected:          []byte{0},
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<0x1::string::String>",
			arg:               "0x010568656c6c6f",
			expected:          []byte{1, 5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o')},
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<vector<0x1::string::String>>",
			arg:               "0x00",
			expected:          []byte{0},
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<vector<0x1::string::String>>",
			arg:               "0x01020568656c6c6f07676f6f64627965",
			expected:          []byte{1, 2, 5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o'), 7, byte('g'), byte('o'), byte('o'), byte('d'), byte('b'), byte('y'), byte('e')},
			compatibilityMode: true,
		},
		{
			strTag:            "vector<0x1::option::Option<u8>>",
			arg:               []any{"0x00", "0x0102"},
			expected:          []byte{2, 0, 1, 2},
			compatibilityMode: true,
		},
		{
			strTag:            "vector<0x1::option::Option<0x1::string::String>>",
			arg:               []any{"0x00", "0x010568656c6c6f"},
			expected:          []byte{2, 0, 1, 5, byte('h'), byte('e'), byte('l'), byte('l'), byte('o')},
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<0x1::string::String>",
			arg:               "0x0100",
			expected:          []byte{1, 0}, // Some("")
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<vector<0x1::string::String>>",
			arg:               "0x0100",
			expected:          []byte{1, 0}, // Some([])
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<0x1::string::String>",
			arg:               "not_hex",
			wantErr:           true,
			compatibilityMode: true,
		},
		{
			strTag:            "0x1::option::Option<vector<0x1::string::String>>",
			arg:               "not_hex",
			wantErr:           true,
			compatibilityMode: true,
		},
		{
			strTag:            "vector<u8>",
			arg:               "0x00",
			expected:          []byte{4, 48, 120, 48, 48}, // "0x00" as bytes
			compatibilityMode: false,
		},
		{
			strTag:            "vector<u8>",
			arg:               "0x00",
			expected:          []byte{4, 48, 120, 48, 48}, // "0x00" as bytes
			compatibilityMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.strTag, func(t *testing.T) {
			t.Parallel()
			typeArg, err := ParseTypeTag(tt.strTag)
			require.NoError(t, err)

			val, err := ConvertArg(*typeArg, tt.arg, tt.generics, CompatibilityMode(tt.compatibilityMode))
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertArg() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equalf(t, tt.expected, val, "ConvertArg() failed to convert to correct bytes %v != %v", tt.expected, val)
		})
	}
}

func TestConvertToU8(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    uint8
		wantErr bool
	}{
		{
			name:    "int",
			input:   42,
			want:    42,
			wantErr: false,
		},
		{
			name:    "uint",
			input:   uint(42),
			want:    42,
			wantErr: false,
		},
		{
			name:    "uint8",
			input:   uint8(42),
			want:    42,
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(42),
			want:    42,
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    0,
			wantErr: true,
		},
		{
			name:    "string",
			input:   "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToU8(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToU8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertToU8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    bool
		wantErr bool
	}{
		{
			name:    "true bool",
			input:   true,
			want:    true,
			wantErr: false,
		},
		{
			name:    "false bool",
			input:   false,
			want:    false,
			wantErr: false,
		},
		{
			name:    "true string",
			input:   "true",
			want:    true,
			wantErr: false,
		},
		{
			name:    "false string",
			input:   "false",
			want:    false,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    false,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   42,
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    *AccountAddress
		wantErr bool
	}{
		{
			name:    "AccountAddress",
			input:   AccountOne,
			want:    &AccountOne,
			wantErr: false,
		},
		{
			name:    "AccountAddress pointer",
			input:   &AccountOne,
			want:    &AccountOne,
			wantErr: false,
		},
		{
			name:    "nil AccountAddress pointer",
			input:   (*AccountAddress)(nil),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid string",
			input:   "0x1",
			want:    &AccountOne,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   42,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want.String() {
				t.Errorf("ConvertToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToVectorU8(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    []byte
		wantErr bool
	}{
		{
			name:    "hex string",
			input:   "0x42",
			want:    []byte{0x04, 48, 120, 52, 50},
			wantErr: false,
		},
		{
			name:    "bytes",
			input:   []byte{0x42},
			want:    []byte{0x01, 0x42},
			wantErr: false,
		},
		{
			name:    "nil bytes",
			input:   []byte(nil),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   42,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToVectorU8(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToVectorU8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("ConvertToVectorU8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToI8(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    int8
		wantErr bool
	}{
		{
			name:    "int positive",
			input:   42,
			want:    42,
			wantErr: false,
		},
		{
			name:    "int negative",
			input:   -42,
			want:    -42,
			wantErr: false,
		},
		{
			name:    "int8",
			input:   int8(-128),
			want:    -128,
			wantErr: false,
		},
		{
			name:    "int8 max",
			input:   int8(127),
			want:    127,
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(-42),
			want:    -42,
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    0,
			wantErr: true,
		},
		{
			name:    "string",
			input:   "-42",
			want:    -42,
			wantErr: false,
		},
		{
			name:    "out of range positive",
			input:   128,
			want:    0,
			wantErr: true,
		},
		{
			name:    "out of range negative",
			input:   -129,
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToI8(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToI8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertToI8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToI16(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    int16
		wantErr bool
	}{
		{
			name:    "int positive",
			input:   1000,
			want:    1000,
			wantErr: false,
		},
		{
			name:    "int negative",
			input:   -1000,
			want:    -1000,
			wantErr: false,
		},
		{
			name:    "int16 min",
			input:   int16(-32768),
			want:    -32768,
			wantErr: false,
		},
		{
			name:    "int16 max",
			input:   int16(32767),
			want:    32767,
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(-1000),
			want:    -1000,
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    0,
			wantErr: true,
		},
		{
			name:    "string",
			input:   "-1000",
			want:    -1000,
			wantErr: false,
		},
		{
			name:    "out of range positive",
			input:   32768,
			want:    0,
			wantErr: true,
		},
		{
			name:    "out of range negative",
			input:   -32769,
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToI16(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToI16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertToI16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToI32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    int32
		wantErr bool
	}{
		{
			name:    "int positive",
			input:   100000,
			want:    100000,
			wantErr: false,
		},
		{
			name:    "int negative",
			input:   -100000,
			want:    -100000,
			wantErr: false,
		},
		{
			name:    "int32 min",
			input:   int32(-2147483648),
			want:    -2147483648,
			wantErr: false,
		},
		{
			name:    "int32 max",
			input:   int32(2147483647),
			want:    2147483647,
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(-100000),
			want:    -100000,
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    0,
			wantErr: true,
		},
		{
			name:    "string",
			input:   "-100000",
			want:    -100000,
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToI32(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToI32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertToI32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    int64
		wantErr bool
	}{
		{
			name:    "int positive",
			input:   100000000000,
			want:    100000000000,
			wantErr: false,
		},
		{
			name:    "int negative",
			input:   -100000000000,
			want:    -100000000000,
			wantErr: false,
		},
		{
			name:    "int64",
			input:   int64(-9223372036854775808),
			want:    -9223372036854775808,
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(-100000000000),
			want:    -100000000000,
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    0,
			wantErr: true,
		},
		{
			name:    "string",
			input:   "-100000000000",
			want:    -100000000000,
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToI64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToI64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertToI64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToI128(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    *big.Int
		wantErr bool
	}{
		{
			name:    "int positive",
			input:   42,
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name:    "int negative",
			input:   -42,
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name:    "int64",
			input:   int64(-9223372036854775808),
			want:    big.NewInt(-9223372036854775808),
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(-42),
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "string positive",
			input:   "42",
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name:    "string negative",
			input:   "-42",
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToI128(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToI128() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Cmp(tt.want) != 0 {
				t.Errorf("ConvertToI128() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToI256(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    *big.Int
		wantErr bool
	}{
		{
			name:    "int positive",
			input:   42,
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name:    "int negative",
			input:   -42,
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name:    "big.Int",
			input:   big.NewInt(-42),
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name:    "nil big.Int",
			input:   (*big.Int)(nil),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "string positive",
			input:   "42",
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name:    "string negative",
			input:   "-42",
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   float64(42),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertToI256(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToI256() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Cmp(tt.want) != 0 {
				t.Errorf("ConvertToI256() = %v, want %v", got, tt.want)
			}
		})
	}
}
