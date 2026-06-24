//go:build 386 || arm || mips || mipsle || wasm

package bcs

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadBytesExceedsIntMax(t *testing.T) {
	t.Parallel()

	ser := NewSerializer()
	ser.Uleb128(uint32(math.MaxInt32) + 1)
	require.NoError(t, ser.Error())

	des := NewDeserializer(ser.ToBytes())
	result := des.ReadBytes()
	assert.Nil(t, result)
	assert.ErrorIs(t, des.Error(), ErrOverflow)
}
