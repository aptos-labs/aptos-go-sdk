package aptos

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// For Content-Type header
const APTOS_SIGNED_BCS = "application/x.aptos.signed_transaction+bcs"

const (
	localnet = "localnet"
	devnet   = "devnet"
	testnet  = "testnet"
	mainnet  = "mainnet"
)

const (
	localnet_api = "http://localhost:8080/v1"
	devnet_api   = "https://api.devnet.aptoslabs.com/v1"
	testnet_api  = "https://api.testnet.aptoslabs.com/v1"
	mainnet_api  = "https://api.mainnet.aptoslabs.com/v1"
)

const (
	localnet_faucet = "http://localhost:8081/v1"
	devnet_faucet   = "https://faucet.devnet.aptoslabs.com/"
	testnet_faucet  = "https://faucet.testnet.aptoslabs.com/"
)

const (
	localnet_chain_id = 4
	testnet_chain_id  = 2
	mainnet_chain_id  = 1
)

type NetworkConfig struct {
	network *string
	api     *string
	faucet  *string
	indexer *string
}

type Client struct {
	restClient   RestClient
	faucetClient FaucetClient
	// TODO: Add indexer client
}

// TODO: rename 'NodeClient' (vs IndexerClient) ?
// what looks best for `import aptos "github.com/aptoslabs/aptos-go-sdk"` then aptos.NewClient() ?
type RestClient struct {
	ChainId uint8

	client  http.Client
	baseUrl url.URL
}

func NewNetworkClient(config NetworkConfig) (client *Client, err error) {
	var apiUrl *url.URL = nil

	switch {
	case config.api == nil && config.network == nil:
		err = errors.New("aptos api url or network is required")
		return
	case config.api != nil:
		apiUrl, err = url.Parse(*config.api)
		if err != nil {
			return
		}
	case *config.network == localnet:
		apiUrl, err = url.Parse(localnet_api)
		if err != nil {
			return
		}
	case *config.network == devnet:
		apiUrl, err = url.Parse(devnet_api)
		if err != nil {
			return
		}
	case *config.network == testnet:
		apiUrl, err = url.Parse(testnet_api)
		if err != nil {
			return
		}
	case *config.network == mainnet:
		apiUrl, err = url.Parse(mainnet_api)
		if err != nil {
			return
		}
	default:
		err = errors.New("network name is unknown, please put localnet, devnet, testnet, or mainnet")
		return
	}
	var faucetUrl *url.URL = nil

	switch {
	case config.faucet == nil && config.network == nil:
		err = errors.New("aptos faucet url or network is required")
		return
	case config.faucet != nil:
		faucetUrl, err = url.Parse(*config.faucet)
		if err != nil {
			return
		}
	case *config.network == localnet:
		faucetUrl, err = url.Parse(localnet_faucet)
		if err != nil {
			return
		}
	case *config.network == devnet:
		faucetUrl, err = url.Parse(devnet_faucet)
		if err != nil {
			return
		}
	case *config.network == testnet:
		faucetUrl, err = url.Parse(testnet_faucet)
		if err != nil {
			return
		}
	case *config.network == mainnet:
		faucetUrl = nil
	default:
		err = errors.New("network name is unknown, please put localnet, devnet, testnet, or mainnet")
		return
	}

	// TODO: add indexer

	restClient := new(RestClient)
	restClient.baseUrl = *apiUrl
	restClient.client.Timeout = 60 * time.Second // TODO: Make configurable
	faucetClient := &FaucetClient{
		restClient,
		faucetUrl,
	}
	client = &Client{
		*restClient,
		*faucetClient,
	}
	return
}

// TODO: Deprecate
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

type NodeInfo struct {
	ChainId                uint8  `json:"chain_id"`
	EpochStr               string `json:"epoch"`
	LedgerVersionStr       string `json:"ledger_version"`
	OldestLedgerVersionStr string `json:"oldest_ledger_version"`
	NodeRole               string `json:"node_role"`
	BlockHeightStr         string `json:"block_height"`
	OldestBlockHeightStr   string `json:"oldest_block_height"`
	GitHash                string `json:"git_hash"`
}

// TODO: write NodeInfo accessors to ParseUint on *Str which work around 53 bit float64 limit in JavaScript

func (rc *RestClient) Info() (info NodeInfo, err error) {
	response, err := rc.client.Get(rc.baseUrl.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", rc.baseUrl.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &info)
	if err == nil {
		rc.ChainId = info.ChainId
	}
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
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "account json err: %v\n%s\n", err, string(blob))
	}
	return
}

// AccountResourceInfo is returned by #AccountResource() and #AccountResources()
type AccountResourceInfo struct {
	// e.g. "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>"
	Type string `json:"type"`

	// Decoded from Move contract data, could really be anything
	Data map[string]any `json:"data"`
}

func (rc *RestClient) AccountResource(address AccountAddress, resourceType string, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl
	// TODO: offer a list of known-good resourceType string constants
	// TODO: set "Accept: application/x-bcs" and parse BCS objects for lossless (and faster) transmission
	au.Path = path.Join(au.Path, "accounts", address.String(), "resource", resourceType)
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

// AccountResources fetches resources for an account into a JSON-like map[string]any in AccountResourceInfo.Data
// For fetching raw Move structs as BCS, See #AccountResourcesBCS
func (rc *RestClient) AccountResources(address AccountAddress, ledger_version ...int) (resources []AccountResourceInfo, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String(), "resources")
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &resources)
	return
}

// DeserializeSequence[AccountResourceRecord](bcs) approximates the Rust side BTreeMap<StructTag,Vec<u8>>
// They should BCS the same with a prefix Uleb128 length followed by (StructTag,[]byte) pairs.
type AccountResourceRecord struct {
	// Account::Module::Name
	Tag StructTag

	// BCS data as stored by Move contract
	Data []byte
}

func (aar *AccountResourceRecord) MarshalBCS(bcs *Serializer) {
	aar.Tag.MarshalBCS(bcs)
	bcs.WriteBytes(aar.Data)
}
func (aar *AccountResourceRecord) UnmarshalBCS(bcs *Deserializer) {
	aar.Tag.UnmarshalBCS(bcs)
	aar.Data = bcs.ReadBytes()
}

func (rc *RestClient) GetBCS(getUrl string) (*http.Response, error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-bcs")
	return rc.client.Do(req)
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (rc *RestClient) AccountResourcesBCS(address AccountAddress, ledger_version ...int) (resources []AccountResourceRecord, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String(), "resources")
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.GetBCS(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	bcs := NewDeserializer(blob)
	// See resource_test.go TestMoveResourceBCS
	resources = DeserializeSequence[AccountResourceRecord](bcs)
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
	restUrl := rc.baseUrl
	restUrl.Path = path.Join(restUrl.Path, "transactions/by_hash", txnHash)
	return rc.getTransactionCommon(restUrl)
}

func (rc *RestClient) TransactionByVersion(version uint64) (data map[string]any, err error) {
	restUrl := rc.baseUrl
	restUrl.Path = path.Join(restUrl.Path, "transactions/by_version", strconv.FormatUint(version, 10))
	return rc.getTransactionCommon(restUrl)
}

func (rc *RestClient) getTransactionCommon(restUrl url.URL) (data map[string]any, err error) {
	// Fetch transaction
	response, err := rc.client.Get(restUrl.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", restUrl.String(), err)
		return
	}

	// Handle Errors TODO: Handle ratelimits, etc.
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}

	// Read body to JSON TODO: BCS
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close() // We don't care about the error about closing the body
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
					slog.Debug("txn done", "hash", hash, "status", status["success"])
				}
			}
		}
	}
	return nil
}

// Get recent transactions.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
func (rc *RestClient) Transactions(start *uint64, limit *uint64) (data []map[string]any, err error) {
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
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

// testing only
// There exists an aptos-node API for submitting JSON and having the node Rust code encode it to BCS, we should only use this for testing to validate our local BCS. Actual GO-SDK usage should use BCS encoding locally in Go code.
func (rc *RestClient) transactionEncode(request map[string]any) (data []byte, err error) {
	rblob, err := json.Marshal(request)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(rblob)
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "transactions/encode_submission")
	response, err := rc.client.Post(au.String(), "application/json", bodyReader)
	if err != nil {
		err = fmt.Errorf("POST %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) SubmitTransaction(stxn *SignedTransaction) (data map[string]any, err error) {
	bcs := Serializer{}
	stxn.MarshalBCS(&bcs)
	err = bcs.Error()
	if err != nil {
		return
	}
	sblob := bcs.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "transactions")
	response, err := rc.client.Post(au.String(), APTOS_SIGNED_BCS, bodyReader)
	if err != nil {
		err = fmt.Errorf("POST %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return nil, err
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close()
	//return blob, nil
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) GetChainId() (chainId uint8, err error) {
	if rc.ChainId != 0 {
		return rc.ChainId, nil
	}
	info, err := rc.Info()
	if err != nil {
		return 0, err
	}
	return info.ChainId, nil
}

type HttpError struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Header     http.Header
	Body       []byte
}

func NewHttpError(response *http.Response) *HttpError {
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	return &HttpError{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Header:     response.Header,
		Body:       body,
	}
}

// implement error interface
func (he *HttpError) Error() string {
	return fmt.Sprintf("HttpError %#v (%d bytes %s)", he.Status, len(he.Body), he.Header.Get("Content-Type"))
}
