package codegen

import (
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// accountOne is 0x1 for tests
var accountOne = types.AccountOne

func TestGenerateModule_Basic(t *testing.T) {
	t.Parallel()

	addr := accountOne
	module := &api.MoveModule{
		Address: &addr,
		Name:    "test_module",
		Friends: []string{},
		ExposedFunctions: []*api.MoveFunction{
			{
				Name:       "transfer",
				Visibility: api.MoveVisibilityPublic,
				IsEntry:    true,
				IsView:     false,
				Params:     []string{"&signer", "address", "u64"},
				Return:     []string{},
			},
		},
		Structs: []*api.MoveStruct{
			{
				Name:     "Coin",
				IsNative: false,
				Abilities: []api.MoveAbility{
					api.MoveAbilityStore,
				},
				Fields: []*api.MoveStructField{
					{Name: "value", Type: "u64"},
				},
			},
		},
	}

	result, err := GenerateModule(module, DefaultOptions())
	require.NoError(t, err)
	require.NotNil(t, result)

	code := string(result.Code)

	// Check package name
	assert.Contains(t, code, "package test_module")

	// Check struct generation
	assert.Contains(t, code, "type Coin struct")
	assert.Contains(t, code, "Value uint64")

	// Check entry function generation
	assert.Contains(t, code, "func Transfer(")
	assert.Contains(t, code, "*aptos.EntryFunction")

	// Check module address
	assert.Contains(t, code, `ModuleAddress = mustParseAddress("0x1")`)
}

func TestGenerateModule_ViewFunction(t *testing.T) {
	t.Parallel()

	addr := accountOne
	module := &api.MoveModule{
		Address: &addr,
		Name:    "coin",
		ExposedFunctions: []*api.MoveFunction{
			{
				Name:       "balance",
				Visibility: api.MoveVisibilityPublic,
				IsEntry:    false,
				IsView:     true,
				Params:     []string{"address"},
				Return:     []string{"u64"},
				GenericTypeParams: []*api.GenericTypeParam{
					{Constraints: []api.MoveAbility{}},
				},
			},
		},
	}

	result, err := GenerateModule(module, DefaultOptions())
	require.NoError(t, err)

	code := string(result.Code)

	// Check view function generation
	assert.Contains(t, code, "func Balance(")
	assert.Contains(t, code, "client aptos.Client")
	assert.Contains(t, code, "typeArgs ...aptos.TypeTag")
}

func TestGenerateModule_ComplexTypes(t *testing.T) {
	t.Parallel()

	addr := accountOne
	module := &api.MoveModule{
		Address: &addr,
		Name:    "complex_module",
		Structs: []*api.MoveStruct{
			{
				Name:      "ComplexStruct",
				IsNative:  false,
				Abilities: []api.MoveAbility{api.MoveAbilityCopy, api.MoveAbilityDrop},
				Fields: []*api.MoveStructField{
					{Name: "balance", Type: "u128"},
					{Name: "data", Type: "vector<u8>"},
					{Name: "name", Type: "0x1::string::String"},
					{Name: "maybe_value", Type: "0x1::option::Option<u64>"},
				},
			},
		},
	}

	result, err := GenerateModule(module, DefaultOptions())
	require.NoError(t, err)

	code := string(result.Code)

	// Check complex type conversions (using Contains to handle formatting)
	assert.Contains(t, code, "Balance")
	assert.Contains(t, code, "*big.Int")
	assert.Contains(t, code, "Data")
	assert.Contains(t, code, "[]byte")
	assert.Contains(t, code, "Name")
	assert.Contains(t, code, "MaybeValue")
	assert.Contains(t, code, "*uint64")
}

func TestGenerateModule_Options(t *testing.T) {
	t.Parallel()

	addr := accountOne
	module := &api.MoveModule{
		Address: &addr,
		Name:    "test",
		ExposedFunctions: []*api.MoveFunction{
			{
				Name:       "entry_fn",
				Visibility: api.MoveVisibilityPublic,
				IsEntry:    true,
			},
			{
				Name:       "view_fn",
				Visibility: api.MoveVisibilityPublic,
				IsView:     true,
				Return:     []string{"bool"},
			},
		},
		Structs: []*api.MoveStruct{
			{Name: "TestStruct", Fields: []*api.MoveStructField{{Name: "x", Type: "u8"}}},
		},
	}

	t.Run("custom package name", func(t *testing.T) {
		t.Parallel()
		opts := DefaultOptions()
		opts.PackageName = "mypackage"
		result, err := GenerateModule(module, opts)
		require.NoError(t, err)
		assert.Contains(t, string(result.Code), "package mypackage")
	})

	t.Run("no structs", func(t *testing.T) {
		t.Parallel()
		opts := DefaultOptions()
		opts.GenerateStructs = false
		result, err := GenerateModule(module, opts)
		require.NoError(t, err)
		assert.NotContains(t, string(result.Code), "type TestStruct struct")
	})

	t.Run("no entry functions", func(t *testing.T) {
		t.Parallel()
		opts := DefaultOptions()
		opts.GenerateEntryFunctions = false
		result, err := GenerateModule(module, opts)
		require.NoError(t, err)
		assert.NotContains(t, string(result.Code), "func EntryFn(")
	})

	t.Run("no view functions", func(t *testing.T) {
		t.Parallel()
		opts := DefaultOptions()
		opts.GenerateViewFunctions = false
		result, err := GenerateModule(module, opts)
		require.NoError(t, err)
		assert.NotContains(t, string(result.Code), "func ViewFn(")
	})
}

func TestMoveTypeToGoType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		moveType string
		expected string
	}{
		{"bool", "bool"},
		{"u8", "uint8"},
		{"u16", "uint16"},
		{"u32", "uint32"},
		{"u64", "uint64"},
		{"u128", "*big.Int"},
		{"u256", "*big.Int"},
		{"address", "aptos.AccountAddress"},
		{"signer", "aptos.AccountAddress"},
		{"&signer", "aptos.AccountAddress"},
		{"vector<u8>", "[]byte"},
		{"vector<u64>", "[]uint64"},
		{"vector<address>", "[]aptos.AccountAddress"},
		{"0x1::string::String", "string"},
		{"0x1::option::Option<u64>", "*uint64"},
		{"0x1::object::Object<T>", "aptos.AccountAddress"},
		{"&u64", "uint64"},
		{"&mut u64", "uint64"},
		{"T0", "any"},
		{"T1", "any"},
	}

	for _, tt := range tests {
		t.Run(tt.moveType, func(t *testing.T) {
			t.Parallel()
			got := moveTypeToGoType(tt.moveType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestToGoPublicName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"transfer", "Transfer"},
		{"transfer_coins", "TransferCoins"},
		{"get_balance", "GetBalance"},
		{"APT", "APT"},
		{"a_b_c", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := toGoPublicName(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSanitizePackageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"coin", "coin"},
		{"aptos_coin", "aptos_coin"},
		{"AptosCoin", "aptoscoin"},
		{"my-module", "my_module"},
		{"123module", "module"}, // Can't start with number
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := sanitizePackageName(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGenerateModule_NilModule(t *testing.T) {
	t.Parallel()
	_, err := GenerateModule(nil, DefaultOptions())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGenerateModule_NativeStructsSkipped(t *testing.T) {
	t.Parallel()

	addr := accountOne
	module := &api.MoveModule{
		Address: &addr,
		Name:    "test",
		Structs: []*api.MoveStruct{
			{Name: "NativeType", IsNative: true},
			{Name: "RegularType", IsNative: false, Fields: []*api.MoveStructField{{Name: "x", Type: "u8"}}},
		},
	}

	result, err := GenerateModule(module, DefaultOptions())
	require.NoError(t, err)

	code := string(result.Code)
	assert.NotContains(t, code, "NativeType")
	assert.Contains(t, code, "RegularType")
}

func TestGenerateModule_FormattedOutput(t *testing.T) {
	t.Parallel()

	addr := accountOne
	module := &api.MoveModule{
		Address: &addr,
		Name:    "test",
		Structs: []*api.MoveStruct{
			{Name: "Simple", Fields: []*api.MoveStructField{{Name: "x", Type: "u8"}}},
		},
	}

	result, err := GenerateModule(module, DefaultOptions())
	require.NoError(t, err)

	// Check code is properly formatted (no double spaces, proper indentation)
	code := string(result.Code)
	assert.NotContains(t, code, "  \n") // No trailing whitespace
	assert.True(t, strings.HasPrefix(code, "// Code generated"))
}
