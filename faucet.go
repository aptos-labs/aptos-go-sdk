package aptos

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

func truthy(x any) bool {
	switch v := x.(type) {
	case nil:
		return false
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case string:
		v = strings.ToLower(v)
		return (v == "t") || (v == "true")
	default:
		return false
	}
}

type FaucetClient struct {
	restClient *NodeClient
	url        *url.URL
}

// Fund account with the given amount of AptosCoin
func (faucetClient *FaucetClient) Fund(address AccountAddress, amount uint64) error {
	mintUrl := faucetClient.url
	mintUrl.Path = path.Join(mintUrl.Path, "mint")
	params := url.Values{}
	params.Set("amount", strconv.FormatUint(amount, 10))
	params.Set("address", address.String())
	mintUrl.RawQuery = params.Encode()
	response, err := http.Post(mintUrl.String(), "text/plain", nil)
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return NewHttpError(response)
	}
	decoder := json.NewDecoder(response.Body)
	var txnHashes []string
	err = decoder.Decode(&txnHashes)
	if err != nil {
		return fmt.Errorf("response json decode error, %w", err)
	}
	if faucetClient.restClient == nil {
		slog.Debug("FundAccount no txns to wait for")
		// no Aptos client to wait on txn completion
		return nil
	}
	slog.Debug("FundAccount wait for txns", "ntxn", len(txnHashes))
	return faucetClient.restClient.WaitForTransactions(txnHashes)
}
