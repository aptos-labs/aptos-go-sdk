//go:build 386 || arm || mips || mipsle

package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntToU32_maxInt32(t *testing.T) {
	t.Parallel()

	got, err := IntToU32(math.MaxInt32)
	require.NoError(t, err)
	assert.Equal(t, uint32(math.MaxInt32), got)
}
