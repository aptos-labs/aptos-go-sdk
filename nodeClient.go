package aptos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"time"
)

// For Content-Type header when POST-ing a Transaction

// ContentTypeAptosSignedTxnBcs header for sending BCS transaction payloads
const ContentTypeAptosSignedTxnBcs = "application/x.aptos.signed_transaction+bcs"

// ContentTypeAptosViewFunctionBcs header for sending BCS view function payloads
const ContentTypeAptosViewFunctionBcs = "application/x.aptos.view_function+bcs"

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
		return nil, fmt.Errorf("failed to parse RPC url '%s': %w", rpcUrl, err)
	}
	return &NodeClient{
		client:  client,
		baseUrl: baseUrl,
		chainId: chainId,
	}, nil
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
		rc.chainId = info.ChainId
	}
	return
}

func (rc *NodeClient) Account(address AccountAddress, ledgerVersion ...uint64) (info AccountInfo, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String())
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
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
		_, _ = fmt.Fprintf(os.Stderr, "account api err: %v\n%s\n", err, string(blob))
	}
	return
}

func (rc *NodeClient) AccountResource(address AccountAddress, resourceType string, ledgerVersion ...uint64) (data map[string]any, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resource", resourceType)
	// TODO: offer a list of known-good resourceType string constants
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
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
func (rc *NodeClient) AccountResources(address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceInfo, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
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
	req.Header.Set(ClientHeader, ClientHeaderValue)
	return rc.client.Do(req)
}

func (rc *NodeClient) GetBCS(getUrl string) (*http.Response, error) {
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-bcs")
	req.Header.Set(ClientHeader, ClientHeaderValue)
	return rc.client.Do(req)
}

func (rc *NodeClient) Post(postUrl string, contentType string, body io.Reader) (resp *http.Response, err error) {
	if body == nil {
		body = http.NoBody
	}
	req, err := http.NewRequest("POST", postUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(ClientHeader, ClientHeaderValue)
	return rc.client.Do(req)
}

// AccountResourcesBCS fetches account resources as raw Move struct BCS blobs in AccountResourceRecord.Data []byte
func (rc *NodeClient) AccountResourcesBCS(address AccountAddress, ledgerVersion ...uint64) (resources []AccountResourceRecord, err error) {
	au := rc.baseUrl.JoinPath("accounts", address.String(), "resources")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
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
	_ = response.Body.Close()
	deserializer := bcs.NewDeserializer(blob)
	// See resource_test.go TestMoveResourceBCS
	resources = bcs.DeserializeSequence[AccountResourceRecord](deserializer)
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
func (rc *NodeClient) TransactionByHash(txnHash string) (data *api.Transaction, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_hash", txnHash)
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) TransactionByVersion(version uint64) (data *api.Transaction, err error) {
	restUrl := rc.baseUrl.JoinPath("transactions/by_version", strconv.FormatUint(version, 10))
	return rc.getTransactionCommon(restUrl)
}

func (rc *NodeClient) getTransactionCommon(restUrl *url.URL) (data *api.Transaction, err error) {
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

	// Read body to JSON
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close() // We don't care about the error about closing the body
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *NodeClient) BlockByVersion(ledgerVersion uint64, withTransactions bool) (data *api.Block, err error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_version", strconv.FormatUint(ledgerVersion, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

func (rc *NodeClient) BlockByHeight(blockHeight uint64, withTransactions bool) (data *api.Block, err error) {
	restUrl := rc.baseUrl.JoinPath("blocks/by_height", strconv.FormatUint(blockHeight, 10))
	return rc.getBlockCommon(restUrl, withTransactions)
}

func (rc *NodeClient) getBlockCommon(restUrl *url.URL, withTransactions bool) (data *api.Block, err error) {
	params := url.Values{}
	params.Set("with_transactions", strconv.FormatBool(withTransactions))
	restUrl.RawQuery = params.Encode()

	// Fetch block
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

	// Read body to JSON
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	_ = response.Body.Close() // We don't care about the error about closing the body
	data = &api.Block{}
	err = json.Unmarshal(blob, data)
	return
}

// WaitForTransaction does a long-GET for one transaction and wait for it to complete.
// Initially poll at 10 Hz for up to 1 second if node replies with 404 (wait for txn to propagate).
// Accept option arguments PollPeriod and PollTimeout like PollForTransactions.
func (rc *NodeClient) WaitForTransaction(txnHash string, options ...any) (data *api.UserTransaction, err error) {
	return rc.PollForTransaction(txnHash, options...)
}

// PollPeriod is an option to PollForTransactions
type PollPeriod time.Duration

// PollTimeout is an option to PollForTransactions
type PollTimeout time.Duration

func getTransactionPollOptions(defaultPeriod, defaultTimeout time.Duration, options ...any) (period time.Duration, timeout time.Duration, err error) {
	period = defaultPeriod
	timeout = defaultTimeout
	for i, arg := range options {
		switch value := arg.(type) {
		case PollPeriod:
			period = time.Duration(value)
		case PollTimeout:
			timeout = time.Duration(value)
		default:
			err = fmt.Errorf("PollForTransactions arg %d bad type %T", i+1, arg)
			return
		}
	}
	return
}

func (rc *NodeClient) PollForTransaction(hash string, options ...any) (*api.UserTransaction, error) {
	period, timeout, err := getTransactionPollOptions(100*time.Millisecond, 10*time.Second, options...)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	deadline := start.Add(timeout)
	for {
		if time.Now().After(deadline) {
			return nil, errors.New("timeout waiting for faucet transactions")
		}
		time.Sleep(period)
		txn, err := rc.TransactionByHash(hash)
		if err == nil {
			if txn.Type == api.TransactionVariantPendingTransaction {
				// not done yet!
			} else if txn.Type == api.TransactionVariantUserTransaction {
				// done!
				slog.Debug("txn done", "hash", hash)
				return txn.UserTransaction()
			}
		}
	}
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
			txn, err := rc.TransactionByHash(hash)
			if err == nil {
				if txn.Type == api.TransactionVariantPendingTransaction {
					// not done yet!
				} else if txn.Type == api.TransactionVariantUserTransaction {
					// done!
					delete(hashSet, hash)
					slog.Debug("txn done", "hash", hash)
				}
			}
		}
	}
	return nil
}

// Transactions Get recent transactions.
// Start is a version number. Nil for most recent transactions.
// Limit is a number of transactions to return. 'about a hundred' by default.
func (rc *NodeClient) Transactions(start *uint64, limit *uint64) (data []*api.Transaction, err error) {
	au := rc.baseUrl.JoinPath("transactions")
	params := url.Values{}
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
		return nil, err
	}
	if response.StatusCode >= 400 {
		err = NewHttpError(response)
		return nil, err
	}
	blob, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return nil, err
	}
	_ = response.Body.Close()
	err = json.Unmarshal(blob, &data)
	if err != nil {
		return nil, err
	}
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
	response, err := rc.Post(au.String(), "application/api", bodyReader)
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

func (rc *NodeClient) SubmitTransaction(signedTxn *SignedTransaction) (data *api.SubmitTransactionResponse, err error) {
	sblob, err := bcs.Serialize(signedTxn)
	if err != nil {
		return
	}
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("transactions")
	response, err := rc.Post(au.String(), ContentTypeAptosSignedTxnBcs, bodyReader)
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

type MaxGasAmount uint64

type GasUnitPrice uint64

type ExpirationSeconds int64

type FeePayer *AccountAddress

type AdditionalSigners []AccountAddress

type SequenceNumber uint64

type ChainIdOption uint8

// BuildTransaction builds a raw transaction for signing
// Accepts options: MaxGasAmount, GasUnitPrice, ExpirationSeconds, SequenceNumber, ChainIdOption, FeePayer, AdditionalSigners
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

	// TODO: fetch gas price on-chain
	// TODO: optionally simulate for max gas

	expirationTimestampSeconds := uint64(time.Now().Unix() + expirationSeconds)

	// Base raw transaction used for all requests
	return &RawTransaction{
		Sender:                     sender,
		SequenceNumber:             sequenceNumber,
		Payload:                    payload,
		MaxGasAmount:               maxGasAmount,
		GasUnitPrice:               gasUnitPrice,
		ExpirationTimestampSeconds: expirationTimestampSeconds,
		ChainId:                    chainId,
	}, nil
}

// BuildTransactionMultiAgent builds a raw transaction for signing with fee payer or multi-agent
// Accepts options: MaxGasAmount, GasUnitPrice, ExpirationSeconds, SequenceNumber, ChainIdOption, FeePayer, AdditionalSigners
func (rc *NodeClient) BuildTransactionMultiAgent(sender AccountAddress, payload TransactionPayload, options ...any) (rawTxnImpl *RawTransactionWithData, err error) {

	maxGasAmount := uint64(100_000) // Default to 0.001 APT max gas amount
	gasUnitPrice := uint64(100)     // Default to min gas price
	expirationSeconds := int64(300) // Default to 5 minutes
	sequenceNumber := uint64(0)
	haveSequenceNumber := false
	chainId := uint8(0)
	haveChainId := false
	var feePayer *AccountAddress
	var additionalSigners []AccountAddress

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
		case FeePayer:
			feePayer = ovalue
		case AdditionalSigners:
			additionalSigners = ovalue
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

	// TODO: fetch gas price on-chain
	// TODO: optionally simulate for max gas

	expirationTimestampSeconds := uint64(time.Now().Unix() + expirationSeconds)

	// Base raw transaction used for all requests
	rawTxn := &RawTransaction{
		Sender:                     sender,
		SequenceNumber:             sequenceNumber,
		Payload:                    payload,
		MaxGasAmount:               maxGasAmount,
		GasUnitPrice:               gasUnitPrice,
		ExpirationTimestampSeconds: expirationTimestampSeconds,
		ChainId:                    chainId,
	}

	// Based on the options, choose which to use
	if feePayer != nil {
		return &RawTransactionWithData{
			Variant: MultiAgentWithFeePayerRawTransactionWithDataVariant,
			Inner: &MultiAgentWithFeePayerRawTransactionWithData{
				RawTxn:           rawTxn,
				FeePayer:         feePayer,
				SecondarySigners: additionalSigners,
			},
		}, nil
	} else {
		return &RawTransactionWithData{
			Variant: MultiAgentRawTransactionWithDataVariant,
			Inner: &MultiAgentRawTransactionWithData{
				RawTxn:           rawTxn,
				SecondarySigners: additionalSigners,
			},
		}, nil
	}
}

type ViewPayload struct {
	Module   ModuleId
	Function string
	ArgTypes []TypeTag
	Args     [][]byte
}

func (vp *ViewPayload) MarshalBCS(ser *bcs.Serializer) {
	vp.Module.MarshalBCS(ser)
	ser.WriteString(vp.Function)
	bcs.SerializeSequence(vp.ArgTypes, ser)
	ser.Uleb128(uint32(len(vp.Args)))
	for _, a := range vp.Args {
		ser.WriteBytes(a)
	}
}

func (rc *NodeClient) View(payload *ViewPayload, ledgerVersion ...uint64) (data []any, err error) {
	serializer := bcs.Serializer{}
	payload.MarshalBCS(&serializer)
	err = serializer.Error()
	if err != nil {
		return
	}
	sblob := serializer.ToBytes()
	bodyReader := bytes.NewReader(sblob)
	au := rc.baseUrl.JoinPath("view")
	if len(ledgerVersion) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.FormatUint(ledgerVersion[0], 10))
		au.RawQuery = params.Encode()
	}
	response, err := rc.Post(au.String(), ContentTypeAptosViewFunctionBcs, bodyReader)
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

func (rc *NodeClient) EstimateGasPrice() (info EstimateGasInfo, err error) {
	au := rc.baseUrl.JoinPath("estimate_gas_price")
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
		err = fmt.Errorf("failed to deserialize estimate gas price response: %w", err)
		return
	}
	return
}

func (rc *NodeClient) AccountAPTBalance(account AccountAddress) (balance uint64, err error) {
	accountBytes, err := bcs.Serialize(&account)
	if err != nil {
		return 0, err
	}
	values, err := rc.View(&ViewPayload{Module: ModuleId{
		Address: AccountOne,
		Name:    "coin",
	},
		Function: "balance",
		ArgTypes: []TypeTag{AptosCoinTypeTag},
		Args:     [][]byte{accountBytes},
	})
	if err != nil {
		return 0, err
	}
	return util.StrToUint64(values[0].(string))
}

// deprecated, will need to rework accordingly, as this is only built for a single account, very useful for testing though
func (rc *NodeClient) BuildSignAndSubmitTransaction(sender *Account, payload TransactionPayload, options ...any) (data *api.SubmitTransactionResponse, err error) {
	rawTxn, err := rc.BuildTransaction(sender.Address, payload, options...)
	if err != nil {
		return nil, err
	}
	signedTxn, err := rawTxn.SignedTransaction(sender)
	if err != nil {
		return nil, err
	}
	return rc.SubmitTransaction(signedTxn)
}
