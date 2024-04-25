package aptos

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	if err != nil {
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
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
		err = NewHttpError(response)
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
		err = NewHttpError(response)
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
		err = NewHttpError(response)
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

// TransactionByHash gets info on a transaction
// The transaction may be pending or recently committed.
//
//	data, err := c.TransactionByHash("0xabcd")
//	if err != nil {
//		if httpErr, ok := err.(aptos.HttpError) {
//			if httpErr.StatusCode == 404 {
//				// if we're sure this has been submitted, assume it is still pending elsewhere in the mempool
//			}
//		}
//	} else {
//		if data["type"] == "pending_transaction" {
//			// known to local mempool, but not committed yet
//		}
//	}
func (rc *RestClient) TransactionByHash(txnHash string) (data map[string]any, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "transactions/by_hash", txnHash)
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
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

// Waits up to 10 seconds for transactions to be done, polling at 10Hz
// TODO: options for polling period and timeout
func (rc *RestClient) WaitForTransactions(txnHashes []string) error {
	hashSet := make(map[string]bool, len(txnHashes))
	for _, hash := range txnHashes {
		hashSet[hash] = true
	}
	start := time.Now()
	deadline := start.Add(10 * time.Second)
	for len(hashSet) > 0 {
		if time.Now().After(deadline) {
			return errors.New("timeout waiting for faucet transactions")
		}
		time.Sleep(100 * time.Millisecond)
		for _, hash := range txnHashes {
			if !hashSet[hash] {
				// already done
				continue
			}
			status, err := rc.TransactionByHash(hash)
			if err == nil {
				if status["type"] == "pending_transaction" {
					// not done yet!
				} else if truthy(status["success"]) {
					// done!
					delete(hashSet, hash)
				}
			}
		}
	}
	return nil
}

// Get recent transactions.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
func (rc *RestClient) Transactions(start *uint64, limit *uint64) (data map[string]any, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "transactions")
	var params url.Values
	if start != nil {
		params.Set("start", strconv.FormatUint(*start, 10))
	}
	if limit != nil {
		params.Set("limit", strconv.FormatUint(*limit, 10))
	}
	if len(params) != 0 {
		au.RawQuery = params.Encode()
	}
	// TODO: ?limit=N&start=V
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
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

func (rc *RestClient) TransactionEncode(request map[string]any) (data []byte, err error) {
	rblob, err := json.Marshal(request)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(rblob)
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "transactions/encode_submission")
	response, err := rc.client.Post(au.String(), "application/json", bodyReader)
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
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

type HttpError struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Header     http.Header
	Body       []byte
}

func NewHttpError(response *http.Response) *HttpError {
	body, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	return &HttpError{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Header:     response.Header,
		Body:       body,
	}
}

// implement error interface
func (he *HttpError) Error() string {
	return fmt.Sprintf("HttpError %#v", he.Status)
}
