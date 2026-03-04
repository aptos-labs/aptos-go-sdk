package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishPackagePayloadFromJsonFile(t *testing.T) {
	t.Parallel()
	metadata := []byte{0x01, 0x02, 0x03}
	bytecode := [][]byte{{0x10, 0x20, 0x30}}

	result, err := PublishPackagePayloadFromJsonFile(metadata, bytecode)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Payload)

	ef, ok := result.Payload.(*EntryFunction)
	require.True(t, ok)

	assert.Equal(t, AccountOne, ef.Module.Address)
	assert.Equal(t, "code", ef.Module.Name)
	assert.Equal(t, "publish_package_txn", ef.Function)
	assert.Empty(t, ef.ArgTypes)
	assert.Len(t, ef.Args, 2)
}

func TestPublishPackagePayloadFromJsonFile_MultipleBytecodes(t *testing.T) {
	t.Parallel()
	metadata := []byte{0x01, 0x02, 0x03, 0x04}
	bytecode := [][]byte{
		{0x10, 0x20, 0x30},
		{0x40, 0x50, 0x60},
		{0x70, 0x80, 0x90},
	}

	result, err := PublishPackagePayloadFromJsonFile(metadata, bytecode)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Payload)

	ef, ok := result.Payload.(*EntryFunction)
	require.True(t, ok)

	assert.Equal(t, AccountOne, ef.Module.Address)
	assert.Equal(t, "code", ef.Module.Name)
	assert.Equal(t, "publish_package_txn", ef.Function)
	assert.Empty(t, ef.ArgTypes)
	assert.Len(t, ef.Args, 2)
}
