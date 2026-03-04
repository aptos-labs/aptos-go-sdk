package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountInfo_AuthenticationKey(t *testing.T) {
	t.Parallel()

	t.Run("with 0x prefix", func(t *testing.T) {
		t.Parallel()
		ai := AccountInfo{AuthenticationKeyHex: "0x0102030405060708"}
		key, err := ai.AuthenticationKey()
		require.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, key)
	})

	t.Run("without 0x prefix", func(t *testing.T) {
		t.Parallel()
		ai := AccountInfo{AuthenticationKeyHex: "0102030405060708"}
		key, err := ai.AuthenticationKey()
		require.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, key)
	})

	t.Run("invalid hex", func(t *testing.T) {
		t.Parallel()
		ai := AccountInfo{AuthenticationKeyHex: "0xZZZZ"}
		_, err := ai.AuthenticationKey()
		require.Error(t, err)
	})
}

func TestAccountInfo_SequenceNumber(t *testing.T) {
	t.Parallel()

	t.Run("valid uint64", func(t *testing.T) {
		t.Parallel()
		ai := AccountInfo{SequenceNumberStr: "42"}
		seq, err := ai.SequenceNumber()
		require.NoError(t, err)
		assert.Equal(t, uint64(42), seq)
	})

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		ai := AccountInfo{SequenceNumberStr: "0"}
		seq, err := ai.SequenceNumber()
		require.NoError(t, err)
		assert.Equal(t, uint64(0), seq)
	})

	t.Run("invalid string", func(t *testing.T) {
		t.Parallel()
		ai := AccountInfo{SequenceNumberStr: "not_a_number"}
		_, err := ai.SequenceNumber()
		require.Error(t, err)
	})
}
