package aptos_go_sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/types"
	"github.com/aptos-labs/aptos-go-sdk/util"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/core"
)

// For Content-Type header when POST-ing a Transaction
const APTOS_SIGNED_BCS = "application/x.aptos.signed_transaction+bcs"
const APTOS_VIEW_BCS = "application/x.aptos.view_function+bcs"

type NodeClient struct {
	client  *http.Client
	baseUrl *url.URL
	chainId uint8
}

func NewNodeClient(rpcUrl string, chainId uint8) (*NodeClient, error) {
	// Set cookie jar so cookie stickiness applies to connections
	// TODO Add appropriate suffix list
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	defaultClient := &http.Client{
		Jar:     jar,
		Timeout: 60 * time.Second,
	}

	return NewNodeClientWithHttpClient(rpcUrl, chainId, defaultClient)
}

func NewNodeClientWithHttpClient(rpcUrl string, chainId uint8, client *http.Client) (*NodeClient, error) {
	baseUrl, err := url.Parse(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RPC url '%s': %+w", rpcUrl, err)
	}
	return &NodeClient{
		client:  client,
		baseUrl: baseUrl,
		chainId: chainId,
	}, nil
}

func (rc *NodeClient) Info() (info types.NodeInfo, err error) {
	response, err := rc.Get(rc.baseUrl.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", rc.baseUrl.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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
		rc.chainId = info.ChainId
	}
	return
}

func (rc *NodeClient) Account(address core.AccountAddress, ledger_version ...int) (info types.AccountInfo, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String())
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.Get(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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

func (rc *NodeClient) AccountResource(address core.AccountAddress, resourceType string, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resource", resourceType)
	// TODO: offer a list of known-good resourceType string constants
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.Get(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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
func (rc *NodeClient) AccountResources(address core.AccountAddress, ledger_version ...int) (resources []types.AccountResourceInfo, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.Get(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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

func (rc *NodeClient) Get(getUrl string) (*http.Response, error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(APTOS_CLIENT_HEADER, AptosClientHeaderValue)
	return rc.client.Do(req)
}

func (rc *NodeClient) GetBCS(getUrl string) (*http.Response, error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-bcs")
	req.Header.Set(APTOS_CLIENT_HEADER, AptosClientHeaderValue)
	return rc.client.Do(req)
}

func (rc *NodeClient) Post(postUrl string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", postUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(APTOS_CLIENT_HEADER, AptosClientHeaderValue)
	if body == nil {
		req.Body = &nilBodySingleton
	} else {
		readCloser, ok := body.(io.ReadCloser)
		if ok {
			req.Body = readCloser
		} else {
			req.Body = io.NopCloser(body)
		}
	}
	return rc.client.Do(req)
}

// empty io.ReadCloser
type NilBody struct {
}

var nilBodySingleton NilBody

func (nb *NilBody) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
func (nb *NilBody) Close() error {
	return nil
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (rc *NodeClient) AccountResourcesBCS(address core.AccountAddress, ledger_version ...int) (resources []types.AccountResourceRecord, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
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
		err = util.NewHttpError(response)
		return
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	deserializer := bcs.NewDeserializer(blob)
	// See resource_test.go TestMoveResourceBCS
	resources = bcs.DeserializeSequence[types.AccountResourceRecord](deserializer)
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
func (rc *NodeClient) TransactionByHash(txnHash string) (data map[string]any, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_hash", txnHash)
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) TransactionByVersion(version uint64) (data map[string]any, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_version", strconv.FormatUint(version, 10))
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) getTransactionCommon(restUrl *url.URL) (data map[string]any, err error) {
	// Fetch transaction
	response, err := rc.Get(restUrl.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", restUrl.String(), err)
		return
	}

	// Handle Errors TODO: Handle ratelimits, etc.
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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

// WaitForTransaction does a long-GET for one transaction and wait for it to complete.
// Initially poll at 10 Hz for up to 1 second if node replies with 404 (wait for txn to propagate).
// Accept option arguments PollPeriod and PollTimeout like PollForTransactions.
func (rc *NodeClient) WaitForTransaction(txnHash string, options ...any) (data map[string]any, err error) {
	period, timeout, err := getTransactionPollOptions(100*time.Millisecond, 1*time.Second, options...)
	if err != nil {
		return nil, err
	}
	restUrl := rc.baseUrl.JoinPath("transactions/wait_by_hash", txnHash)
	start := time.Now()
	deadline := start.Add(timeout)
	for {
		data, err = rc.getTransactionCommon(restUrl)
		if err == nil {
			return
		}
		if httpErr, ok := err.(*util.HttpError); ok {
			if httpErr.StatusCode == 404 {
				if time.Now().Before(deadline) {
					time.Sleep(period)
				} else {
					return
				}
			}
		} else {
			return
		}
	}

}

// PollPeriod is an option to PollForTransactions
type PollPeriod time.Duration

// PollTimeout is an option to PollForTransactions
type PollTimeout time.Duration

func getTransactionPollOptions(defaultPeriod, defaultTimeout time.Duration, options ...any) (period time.Duration, timeout time.Duration, err error) {
	period = defaultPeriod
	timeout = defaultTimeout
	for argi, arg := range options {
		switch value := arg.(type) {
		case PollPeriod:
			period = time.Duration(value)
		case PollTimeout:
			timeout = time.Duration(value)
		default:
			err = fmt.Errorf("PollForTransactions arg %d bad type %T", argi+1, arg)
			return
		}
	}
	return
}

// PollForTransactions waits up to 10 seconds for transactions to be done, polling at 10Hz
// Accepts options PollPeriod and PollTimeout which should wrap time.Duration values.
func (rc *NodeClient) PollForTransactions(txnHashes []string, options ...any) error {
	period, timeout, err := getTransactionPollOptions(100*time.Millisecond, 10*time.Second, options...)
	if err != nil {
		return err
	}
	hashSet := make(map[string]bool, len(txnHashes))
	for _, hash := range txnHashes {
		hashSet[hash] = true
	}
	start := time.Now()
	deadline := start.Add(timeout)
	for len(hashSet) > 0 {
		if time.Now().After(deadline) {
			return errors.New("timeout waiting for faucet transactions")
		}
		time.Sleep(period)
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
func (rc *NodeClient) Transactions(start *uint64, limit *uint64) (data []map[string]any, err error) {
	au := rc.baseUrl.JoinPath("transactions")
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
	response, err := rc.Get(au.String())
	if err != nil {
		err = fmt.Errorf("GET %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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
func (rc *NodeClient) transactionEncode(request map[string]any) (data []byte, err error) {
	rblob, err := json.Marshal(request)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(rblob)
	au := rc.baseUrl.JoinPath("transactions/encode_submission")
	response, err := rc.Post(au.String(), "application/json", bodyReader)
	if err != nil {
		err = fmt.Errorf("POST %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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

func (rc *NodeClient) SubmitTransaction(stxn *types.SignedTransaction) (data map[string]any, err error) {
	serializer := bcs.Serializer{}
	stxn.MarshalBCS(&serializer)
	err = serializer.Error()
	if err != nil {
		return
	}
	sblob := serializer.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions")
	response, err := rc.Post(au.String(), APTOS_SIGNED_BCS, bodyReader)
	if err != nil {
		err = fmt.Errorf("POST %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
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

func (rc *NodeClient) GetChainId() (chainId uint8, err error) {
	if rc.chainId == 0 {
		info, err := rc.Info()
		if err != nil {
			return 0, err
		}
		// Cache the ChainId for later calls, because performance
		rc.chainId = info.ChainId
	}
	return rc.chainId, nil
}

// MaxGasAmount is an option to APTTransferTransaction
type MaxGasAmount uint64

// GasUnitPrice is an option to APTTransferTransaction
type GasUnitPrice uint64

// ExpirationSeconds is an option to APTTransferTransaction
type ExpirationSeconds int64

// SequenceNumber is an option to APTTransferTransaction
type SequenceNumber uint64

// ChainIdOption is an option to APTTransferTransaction
type ChainIdOption uint8

// BuildTransaction builds a raw transaction for signing
// Accepts options: MaxGasAmount, GasUnitPrice, ExpirationSeconds, SequenceNumber, ChainIdOption
func (rc *NodeClient) BuildTransaction(sender core.AccountAddress, payload types.TransactionPayload, options ...any) (rawTxn *types.RawTransaction, err error) {

	maxGasAmount := uint64(100_000) // Default to 0.001 APT max gas amount
	gasUnitPrice := uint64(100)     // Default to min gas price
	expirationSeconds := int64(300) // Default to 5 minutes
	sequenceNumber := uint64(0)
	haveSequenceNumber := false
	chainId := uint8(0)
	haveChainId := false

	for opti, option := range options {
		switch ovalue := option.(type) {
		case MaxGasAmount:
			maxGasAmount = uint64(ovalue)
		case GasUnitPrice:
			gasUnitPrice = uint64(ovalue)
		case ExpirationSeconds:
			expirationSeconds = int64(ovalue)
			if expirationSeconds < 0 {
				err = errors.New("ExpirationSeconds cannot be less than 0")
				return nil, err
			}
		case SequenceNumber:
			sequenceNumber = uint64(ovalue)
			haveSequenceNumber = true
		case ChainIdOption:
			chainId = uint8(ovalue)
			haveChainId = true
		default:
			err = fmt.Errorf("APTTransferTransaction arg [%d] unknown option type %T", opti+4, option)
			return nil, err
		}
	}

	// Fetch ChainId which may be cached
	if !haveChainId {
		chainId, err = rc.GetChainId()
		if err != nil {
			return nil, err
		}
	}

	// Fetch sequence number unless provided
	if !haveSequenceNumber {
		info, err := rc.Account(sender)
		if err != nil {
			return nil, err
		}
		sequenceNumber, err = info.SequenceNumber()
		if err != nil {
			return nil, err
		}
	}

	// TODO: fetch gas price onchain
	// TODO: optionally simulate for max gas

	expirationTimestampSeconds := uint64(time.Now().Unix() + expirationSeconds)
	rawTxn = &types.RawTransaction{
		Sender:                    sender,
		SequenceNumber:            sequenceNumber,
		Payload:                   payload,
		MaxGasAmount:              maxGasAmount,
		GasUnitPrice:              gasUnitPrice,
		ExpirationTimetampSeconds: expirationTimestampSeconds,
		ChainId:                   chainId,
	}

	return
}

// BuildSignAndSubmitTransaction right now, this is "easy mode", all in one, no configuration.  More configuration comes
// from splitting into multiple calls
func (rc *NodeClient) BuildSignAndSubmitTransaction(sender *core.Account, payload types.TransactionPayload, options ...any) (hash string, err error) {
	rawTxn, err := rc.BuildTransaction(sender.Address, payload, options...)
	if err != nil {
		return
	}
	// TODO: This shows we should be taking the account, and let it handle the sign part rather than the private key
	signedTxn, err := rawTxn.Sign(sender)
	if err != nil {
		return
	}

	response, err := rc.SubmitTransaction(signedTxn)
	if err != nil {
		return
	}
	return response["hash"].(string), nil
}

type ViewPayload struct {
	Module   types.ModuleId
	Function string
	ArgTypes []types.TypeTag
	Args     [][]byte
}

func (vp *ViewPayload) MarshalBCS(serializer *bcs.Serializer) {
	vp.Module.MarshalBCS(serializer)
	serializer.WriteString(vp.Function)
	bcs.SerializeSequence(vp.ArgTypes, serializer)
	serializer.Uleb128(uint32(len(vp.Args)))
	for _, a := range vp.Args {
		serializer.WriteBytes(a)
	}
}

func (rc *NodeClient) View(payload *ViewPayload) (data []any, err error) {
	serializer := bcs.Serializer{}
	payload.MarshalBCS(&serializer)
	err = serializer.Error()
	if err != nil {
		return
	}
	sblob := serializer.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("view")
	response, err := rc.Post(au.String(), APTOS_VIEW_BCS, bodyReader)
	if err != nil {
		err = fmt.Errorf("POST %s, %w", au.String(), err)
		return
	}
	if response.StatusCode >= 400 {
		err = util.NewHttpError(response)
		return nil, err
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
