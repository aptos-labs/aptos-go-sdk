package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test_HealthCheckResponse(t *testing.T) {
	testJson := `{
		"message": "aptos-node:ok"
	}`
	data := &HealthCheckResponse{}
	err := json.Unmarshal([]byte(testJson), &data)
	require.NoError(t, err)
	assert.Equal(t, "aptos-node:ok", data.Message)
}
