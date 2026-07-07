package aptos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func objectCoreHandler(t *testing.T, owner string, allowUngated bool) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/resource/")
		assert.Contains(t, r.URL.Path, "0x1::object::ObjectCore")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Resource{
			Type: ObjectCoreResourceType,
			Data: map[string]any{
				"owner":                  owner,
				"allow_ungated_transfer": allowUngated,
				"guid_creation_num":      "1125899906842625",
			},
		})
	}
}

func TestGetObjectCore(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, objectCoreHandler(t, AccountOne.String(), true))

	core, err := GetObjectCore(context.Background(), client, AccountTwo)
	require.NoError(t, err)
	assert.Equal(t, AccountOne, core.Owner)
	assert.True(t, core.AllowUngatedTransfer)
	assert.Equal(t, uint64(1125899906842625), core.GuidCreationNum)
}

func TestIsObjectOwner(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, objectCoreHandler(t, AccountOne.String(), true))

	yes, err := IsObjectOwner(context.Background(), client, AccountTwo, AccountOne)
	require.NoError(t, err)
	assert.True(t, yes)

	no, err := IsObjectOwner(context.Background(), client, AccountTwo, AccountThree)
	require.NoError(t, err)
	assert.False(t, no)
}

func TestObjectTransferPayload(t *testing.T) {
	t.Parallel()
	object := AccountTwo
	recipient := AccountThree

	payload := ObjectTransferPayload(object, recipient)

	assert.Equal(t, AccountOne, payload.Module.Address)
	assert.Equal(t, "object", payload.Module.Name)
	assert.Equal(t, "transfer", payload.Function)
	require.Len(t, payload.TypeArgs, 1)
	assert.Equal(t, "0x1::object::ObjectCore", payload.TypeArgs[0].String())
	require.Len(t, payload.Args, 2)
	assert.Equal(t, object, payload.Args[0])
	assert.Equal(t, recipient, payload.Args[1])
}
