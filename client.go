package aptos

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// TODO: rename 'NodeClient' (vs IndexerClient) ?
// what looks best for `import aptos "github.com/aptoslabs/aptos-go-sdk"` then aptos.NewClient() ?
type RestClient struct {
	ChainId int

	client  http.Client
	baseUrl url.URL
}

func NewClient(baseUrl string) (rc *RestClient, err error) {
	rc = new(RestClient)
	tu, err := url.Parse(baseUrl)
	if err != nil {
		rc = nil
		return
	}
	rc.baseUrl = *tu
	rc.client.Timeout = 60 * time.Second
	return
}

// TODO: set HTTP header "x-aptos-client: aptos-go-sdk/{version}"

// AccountInfo is returned from calls to #Account()
type AccountInfo struct {
	SequenceNumberStr    string `json:"sequence_number"`
	AuthenticationKeyHex string `json:"authentication_key"`
}

// Hex decode of AuthenticationKeyHex
func (ai AccountInfo) AuthenticationKey() ([]byte, error) {
	ak := ai.AuthenticationKeyHex
	if strings.HasPrefix(ak, "0x") {
		ak = ak[2:]
	}
	return hex.DecodeString(ak)
}

// ParseUint of SequenceNumberStr
func (ai AccountInfo) SequenceNumber() (uint64, error) {
	return strconv.ParseUint(ai.SequenceNumberStr, 10, 64)
}

func (rc *RestClient) Account(address AccountAddress, ledger_version ...int) (info AccountInfo, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String())
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if response.StatusCode != 200 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "account json err: %v\n%s\n", err, string(blob))
	}
	return
}

// AccountResourceInfo is returned by #AccountResource() and #AccountResources()
type AccountResourceInfo struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"` // TODO: what are these? Build a struct.
}

func (rc *RestClient) AccountResource(address AccountAddress, resourceType string, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl
	// TODO: offer a list of known-good resourceType string constants
	au.Path = path.Join(au.Path, "accounts", address.String(), "resource", resourceType)
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) AccountResources(address AccountAddress, ledger_version ...int) (resources []AccountResourceInfo, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String(), "resources")
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &resources)
	return
}

func (rc *RestClient) Info() (data map[string]any, err error) {
	au := rc.baseUrl
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) TransactionByHash(txnHash string) (data map[string]any, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "transactions/by_hash", txnHash)
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}
