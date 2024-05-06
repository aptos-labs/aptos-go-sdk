package aptos

import (
	"bytes"
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
	"time"
)

// For Content-Type header when POST-ing a Transaction
const APTOS_SIGNED_BCS = "application/x.aptos.signed_transaction+bcs"
const APTOS_VIEW_BCS = "application/x.aptos.view_function+bcs"

type NodeClient struct {
	ChainId uint8

	client  http.Client
	baseUrl url.URL
}

func (rc *NodeClient) Info() (info NodeInfo, err error) {
	response, err := rc.Get(rc.baseUrl.String())
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

func (rc *NodeClient) Account(address AccountAddress, ledger_version ...int) (info AccountInfo, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String())
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

// TODO: set HTTP header "x-aptos-client: aptos-go-sdk/{version}"

func (rc *NodeClient) AccountResource(address AccountAddress, resourceType string, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl
	// TODO: offer a list of known-good resourceType string constants
	// TODO: set "Accept: application/x-bcs" and parse BCS objects for lossless (and faster) transmission
	au.Path = path.Join(au.Path, "accounts", address.String(), "resource", resourceType)
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
func (rc *NodeClient) AccountResources(address AccountAddress, ledger_version ...int) (resources []AccountResourceInfo, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String(), "resources")
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

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (rc *NodeClient) AccountResourcesBCS(address AccountAddress, ledger_version ...int) (resources []AccountResourceRecord, err error) {
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
func (rc *NodeClient) TransactionByHash(txnHash string) (data map[string]any, err error) {
	restUrl := rc.baseUrl
	restUrl.Path = path.Join(restUrl.Path, "transactions/by_hash", txnHash)
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) TransactionByVersion(version uint64) (data map[string]any, err error) {
	restUrl := rc.baseUrl
	restUrl.Path = path.Join(restUrl.Path, "transactions/by_version", strconv.FormatUint(version, 10))
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) getTransactionCommon(restUrl url.URL) (data map[string]any, err error) {
	// Fetch transaction
	response, err := rc.Get(restUrl.String())
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
// TODO: use new hanging-GET endpoint
func (rc *NodeClient) WaitForTransactions(txnHashes []string) error {
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
func (rc *NodeClient) Transactions(start *uint64, limit *uint64) (data []map[string]any, err error) {
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
	response, err := rc.Get(au.String())
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
func (rc *NodeClient) transactionEncode(request map[string]any) (data []byte, err error) {
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

func (rc *NodeClient) SubmitTransaction(stxn *SignedTransaction) (data map[string]any, err error) {
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

func (rc *NodeClient) GetChainId() (chainId uint8, err error) {
	if rc.ChainId != 0 {
		return rc.ChainId, nil
	}
	info, err := rc.Info()
	if err != nil {
		return 0, err
	}

	// Cache the ChainId for later calls, because performance
	rc.ChainId = info.ChainId

	return rc.ChainId, nil
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
func (rc *NodeClient) BuildTransaction(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxn *RawTransaction, err error) {

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
	rawTxn = &RawTransaction{
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
func (rc *NodeClient) BuildSignAndSubmitTransaction(sender *Account, payload TransactionPayload, options ...any) (hash string, err error) {
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
	Module   ModuleId
	Function string
	ArgTypes []TypeTag
	Args     [][]byte
}

func (vp *ViewPayload) MarshalBCS(bcs *Serializer) {
	vp.Module.MarshalBCS(bcs)
	bcs.WriteString(vp.Function)
	SerializeSequence(vp.ArgTypes, bcs)
	bcs.Uleb128(uint32(len(vp.Args)))
	for _, a := range vp.Args {
		bcs.WriteBytes(a)
	}
}

func (rc *NodeClient) View(payload *ViewPayload) (data []any, err error) {
	bcs := Serializer{}
	payload.MarshalBCS(&bcs)
	err = bcs.Error()
	if err != nil {
		return
	}
	sblob := bcs.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "view")
	response, err := rc.client.Post(au.String(), APTOS_VIEW_BCS, bodyReader)
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
	err = json.Unmarshal(blob, &data)
	return
}
