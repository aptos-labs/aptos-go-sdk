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

// Ask the faucet to send some money to a test account
func FundAccount(rc *RestClient, faucetUrl string, address AccountAddress, amount uint64) error {
	au, err := url.Parse(faucetUrl)
	if err != nil {
		return err
	}
	au.Path = path.Join(au.Path, "mint")
	params := url.Values{}
	params.Set("amount", strconv.FormatUint(amount, 10))
	params.Set("address", address.String())
	au.RawQuery = params.Encode()
	response, err := http.Post(au.String(), "text/plain", nil)
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return NewHttpError(response)
	}
	dec := json.NewDecoder(response.Body)
	var txnHashes []string
	err = dec.Decode(&txnHashes)
	if err != nil {
		return fmt.Errorf("response json decode error, %w", err)
	}
	if rc == nil {
		slog.Debug("FundAccount no txns to wait for")
		// no Aptos client to wait on txn completion
		return nil
	}
	slog.Debug("FundAccount wait for txns", "ntxn", len(txnHashes))
	return rc.WaitForTransactions(txnHashes)
}
