//go:build amd64 || arm64 || ppc64 || ppc64le || mips64 || mips64le || loong64 || riscv64

package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntToU32_maxUint32(t *testing.T) {
	t.Parallel()

	got, err := IntToU32(int(math.MaxUint32))
	require.NoError(t, err)
	assert.Equal(t, uint32(math.MaxUint32), got)

	_, err = IntToU32(int(math.MaxUint32) + 1)
	assert.Error(t, err)
}
