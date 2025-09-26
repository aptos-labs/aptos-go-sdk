package aptos

import (
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/qimeila/aptos-go-sdk/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeB64(x string) ([]byte, error) {
	reader := strings.NewReader(x)
	dec := base64.NewDecoder(base64.StdEncoding, reader)
	return io.ReadAll(dec)
}

func TestMoveResourceBCS(t *testing.T) {
	t.Parallel()
	// fetched from local aptos-node 20240501_152556
	// curl -o /tmp/ar_bcs --header "Accept: application/x-bcs" http://127.0.0.1:8080/v1/accounts/{addr}/resources
	// base64 < /tmp/ar_bcs
	b64text := "AgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABBGNvaW4JQ29pblN0b3JlAQcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQphcHRvc19jb2luCUFwdG9zQ29pbgBpKsLrCwAAAAAAAgAAAAAAAAACAAAAAAAAANGdA6RyqwjAFP2cXRokfP3YJqHHNb55lM2GQFYwd6a7AAAAAAAAAAADAAAAAAAAANGdA6RyqwjAFP2cXRokfP3YJqHHNb55lM2GQFYwd6a7AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEHYWNjb3VudAdBY2NvdW50AJMBINGdA6RyqwjAFP2cXRokfP3YJqHHNb55lM2GQFYwd6a7AAAAAAAAAAAEAAAAAAAAAAEAAAAAAAAAAAAAAAAAAADRnQOkcqsIwBT9nF0aJHz92CahxzW+eZTNhkBWMHemuwAAAAAAAAAAAQAAAAAAAADRnQOkcqsIwBT9nF0aJHz92CahxzW+eZTNhkBWMHemuwAA"
	blob, err := decodeB64(b64text)
	require.NoError(t, err)
	assert.NotNil(t, blob)

	deserializer := bcs.NewDeserializer(blob)
	resources := bcs.DeserializeSequence[AccountResourceRecord](deserializer)
	require.NoError(t, deserializer.Error())
	assert.Len(t, resources, 2)
	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", resources[0].Tag.String())
	assert.Equal(t, "0x1::account::Account", resources[1].Tag.String())
}
