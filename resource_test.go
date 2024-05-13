package aptos

import (
	"encoding/base64"
	bcs2 "github.com/aptos-labs/aptos-go-sdk/bcs"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeB64(x string) ([]byte, error) {
	reader := strings.NewReader(x)
	dec := base64.NewDecoder(base64.StdEncoding, reader)
	return io.ReadAll(dec)
}

func TestMoveResourceBCS(t *testing.T) {
	// fetched from local aptos-node 20240501_152556
	// curl -o /tmp/ar_bcs --header "Accept: application/x-bcs" http://127.0.0.1:8080/v1/accounts/{addr}/resources
	// base64 < /tmp/ar_bcs
	b64text := "AgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABBGNvaW4JQ29pblN0b3JlAQcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQphcHRvc19jb2luCUFwdG9zQ29pbgBpKsLrCwAAAAAAAgAAAAAAAAACAAAAAAAAANGdA6RyqwjAFP2cXRokfP3YJqHHNb55lM2GQFYwd6a7AAAAAAAAAAADAAAAAAAAANGdA6RyqwjAFP2cXRokfP3YJqHHNb55lM2GQFYwd6a7AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEHYWNjb3VudAdBY2NvdW50AJMBINGdA6RyqwjAFP2cXRokfP3YJqHHNb55lM2GQFYwd6a7AAAAAAAAAAAEAAAAAAAAAAEAAAAAAAAAAAAAAAAAAADRnQOkcqsIwBT9nF0aJHz92CahxzW+eZTNhkBWMHemuwAAAAAAAAAAAQAAAAAAAADRnQOkcqsIwBT9nF0aJHz92CahxzW+eZTNhkBWMHemuwAA"
	blob, err := decodeB64(b64text)
	assert.NoError(t, err)
	assert.NotNil(t, blob)

	bcs := bcs2.NewDeserializer(blob)
	//resources := DeserializeSequence[MoveResource](bcs)
	//resourceKeys, resourceValues := DeserializeMap[StructTag, []byte](bcs)
	// like client.go AccountResourcesBCS
	resources := bcs2.DeserializeSequence[AccountResourceRecord](bcs)
	assert.NoError(t, bcs.Error())
	assert.Equal(t, 2, len(resources))
	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", resources[0].Tag.String())
	assert.Equal(t, "0x1::account::Account", resources[1].Tag.String())
}
