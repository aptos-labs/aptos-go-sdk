package aptos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPTTransferTransaction_Mock(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			_ = json.NewEncoder(w).Encode(NodeInfo{ChainId: 4})
		case "/accounts/" + AccountOne.String():
			_ = json.NewEncoder(w).Encode(AccountInfo{
				SequenceNumberStr:    "10",
				AuthenticationKeyHex: "0x" + AccountOne.String(),
			})
		case "/estimate_gas_price":
			_ = json.NewEncoder(w).Encode(EstimateGasInfo{
				GasEstimate: 100,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(NetworkConfig{
		Name:    "mocknet",
		NodeUrl: server.URL,
		ChainId: 4,
	})
	require.NoError(t, err)

	sender, err := NewEd25519Account()
	require.NoError(t, err)

	rawTxn, err := APTTransferTransaction(client, sender, AccountTwo, 1000,
		SequenceNumber(5), ChainIdOption(4), GasUnitPrice(100))
	require.NoError(t, err)

	assert.Equal(t, sender.Address, rawTxn.Sender)
	assert.Equal(t, uint64(5), rawTxn.SequenceNumber)
	assert.Equal(t, uint8(4), rawTxn.ChainId)
}
