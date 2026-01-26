package transaction

import (
	"testing"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Validate(t *testing.T) {
	t.Run("empty builder fails", func(t *testing.T) {
		b := New()
		err := b.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sender is required")
	})

	t.Run("missing payload fails", func(t *testing.T) {
		b := New().Sender(aptos.AccountOne)
		err := b.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payload is required")
	})

	t.Run("valid builder passes", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil)
		err := b.Validate()
		assert.NoError(t, err)
	})
}

func TestBuilder_EntryFunction(t *testing.T) {
	t.Run("parses full address", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil)
		require.NoError(t, b.Validate())
		assert.IsType(t, &aptos.EntryFunctionPayload{}, b.payload)

		payload, ok := b.payload.(*aptos.EntryFunctionPayload)
		require.True(t, ok, "expected EntryFunctionPayload")
		assert.Equal(t, aptos.AccountOne, payload.Module.Address)
		assert.Equal(t, "coin", payload.Module.Name)
		assert.Equal(t, "transfer", payload.Function)
	})

	t.Run("invalid format", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("invalid", nil)
		err := b.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("with type args and args", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", []aptos.TypeTag{}, "arg1", uint64(100))
		require.NoError(t, b.Validate())

		payload, ok := b.payload.(*aptos.EntryFunctionPayload)
		require.True(t, ok, "expected EntryFunctionPayload")
		assert.Len(t, payload.Args, 2)
	})
}

func TestBuilder_Options(t *testing.T) {
	t.Run("max gas", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			MaxGas(50000)
		assert.NotNil(t, b.maxGasAmount)
		assert.Equal(t, uint64(50000), *b.maxGasAmount)
	})

	t.Run("gas price", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			GasPrice(100)
		assert.NotNil(t, b.gasUnitPrice)
		assert.Equal(t, uint64(100), *b.gasUnitPrice)
	})

	t.Run("sequence number", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			SequenceNumber(42)
		assert.NotNil(t, b.sequenceNum)
		assert.Equal(t, uint64(42), *b.sequenceNum)
	})

	t.Run("expiration", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			Expiration(60 * time.Second)
		assert.NotNil(t, b.expiration)
		assert.Equal(t, 60*time.Second, *b.expiration)
	})

	t.Run("estimate gas", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			EstimateGas()
		assert.True(t, b.estimateGas)
		assert.False(t, b.prioritizedGas)
	})

	t.Run("prioritized gas", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			PrioritizedGas()
		assert.True(t, b.estimateGas)
		assert.True(t, b.prioritizedGas)
	})

	t.Run("simulate first", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			SimulateFirst()
		assert.True(t, b.simulateBeforeSubmit)
	})
}

func TestBuilder_MultiAgent(t *testing.T) {
	t.Run("adds secondary signers", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			SecondarySigners(aptos.AccountTwo, aptos.AccountThree).
			EntryFunction("0x1::multisig::execute", nil)

		assert.True(t, b.IsMultiAgent())
		assert.Len(t, b.secondarySigners, 2)
		assert.Equal(t, aptos.AccountTwo, b.secondarySigners[0])
		assert.Equal(t, aptos.AccountThree, b.secondarySigners[1])
	})

	t.Run("chained secondary signers", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			SecondarySigners(aptos.AccountTwo).
			SecondarySigners(aptos.AccountThree).
			EntryFunction("0x1::multisig::execute", nil)

		assert.Len(t, b.secondarySigners, 2)
	})
}

func TestBuilder_FeePayer(t *testing.T) {
	t.Run("sets fee payer", func(t *testing.T) {
		b := New().
			Sender(aptos.AccountOne).
			FeePayer(aptos.AccountTwo).
			EntryFunction("0x1::coin::transfer", nil)

		assert.True(t, b.IsFeePayer())
		assert.Equal(t, aptos.AccountTwo, *b.feePayer)
	})
}

func TestBuilder_Script(t *testing.T) {
	t.Run("sets script payload", func(t *testing.T) {
		bytecode := []byte{0x01, 0x02, 0x03}
		b := New().
			Sender(aptos.AccountOne).
			Script(bytecode, nil, "arg1")

		require.NoError(t, b.Validate())
		assert.IsType(t, &aptos.ScriptPayload{}, b.payload)

		payload, ok := b.payload.(*aptos.ScriptPayload)
		require.True(t, ok, "expected ScriptPayload")
		assert.Equal(t, bytecode, payload.Code)
		assert.Len(t, payload.Args, 1)
	})
}

func TestBuilder_ExpirationTimestamp(t *testing.T) {
	t.Run("future timestamp", func(t *testing.T) {
		futureTime := time.Now().Add(1 * time.Hour)
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			ExpirationTimestamp(futureTime)

		require.NoError(t, b.Validate())
		assert.NotNil(t, b.expiration)
	})

	t.Run("past timestamp fails", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Hour)
		b := New().
			Sender(aptos.AccountOne).
			EntryFunction("0x1::coin::transfer", nil).
			ExpirationTimestamp(pastTime)

		err := b.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "past")
	})
}

func TestBuilder_Options_Output(t *testing.T) {
	b := New().
		Sender(aptos.AccountOne).
		EntryFunction("0x1::coin::transfer", nil).
		MaxGas(50000).
		GasPrice(100).
		SequenceNumber(42).
		Expiration(60 * time.Second).
		EstimateGas().
		SimulateFirst()

	opts := b.Options()
	// We can't easily inspect the options, but we can verify count
	assert.GreaterOrEqual(t, len(opts), 5)
}

func TestParseModuleFunction(t *testing.T) {
	tests := []struct {
		input    string
		wantAddr aptos.AccountAddress
		wantMod  string
		wantFunc string
		wantErr  bool
	}{
		{"0x1::coin::transfer", aptos.AccountOne, "coin", "transfer", false},
		{"0x1::aptos_account::transfer", aptos.AccountOne, "aptos_account", "transfer", false},
		{"invalid", aptos.AccountAddress{}, "", "", true},
		{"0x1::coin", aptos.AccountAddress{}, "", "", true},
		{"bad_addr::coin::transfer", aptos.AccountAddress{}, "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			mod, fn, err := parseModuleFunction(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantAddr, mod.Address)
			assert.Equal(t, tc.wantMod, mod.Name)
			assert.Equal(t, tc.wantFunc, fn)
		})
	}
}

func TestTransferAPT(t *testing.T) {
	from := aptos.AccountOne
	to := aptos.AccountTwo
	amount := uint64(1_000_000)

	b := TransferAPT(from, to, amount)

	assert.Equal(t, &from, b.sender)
	assert.NotNil(t, b.payload)
	require.NoError(t, b.Validate())

	payload, ok := b.payload.(*aptos.EntryFunctionPayload)
	require.True(t, ok, "expected EntryFunctionPayload")
	assert.Equal(t, "aptos_account", payload.Module.Name)
	assert.Equal(t, "transfer", payload.Function)
}

func TestTransferCoins(t *testing.T) {
	from := aptos.AccountOne
	to := aptos.AccountTwo
	amount := uint64(1_000_000)

	// Use the actual AptosCoinTypeTag
	coinType := aptos.AptosCoinTypeTag

	b := TransferCoins(from, to, coinType, amount)

	assert.Equal(t, &from, b.sender)
	assert.NotNil(t, b.payload)
	require.NoError(t, b.Validate())

	payload, ok := b.payload.(*aptos.EntryFunctionPayload)
	require.True(t, ok, "expected EntryFunctionPayload")
	assert.Equal(t, "coin", payload.Module.Name)
	assert.Equal(t, "transfer", payload.Function)
	assert.Len(t, payload.TypeArgs, 1)
	assert.Equal(t, "0x1::aptos_coin::AptosCoin", payload.TypeArgs[0].String())
}
