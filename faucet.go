package aptos_go_sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"github.com/aptos-labs/aptos-go-sdk/util"
	"log/slog"
	"net/url"
	"strconv"
)

type FaucetClient struct {
	nodeClient *NodeClient
	url        *url.URL
}

func NewFaucetClient(nodeClient *NodeClient, faucetUrl string) (*FaucetClient, error) {
	url, err := url.Parse(faucetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse faucet url '%s': %+w", faucetUrl, err)
	}
	return &FaucetClient{
		nodeClient,
		url,
	}, nil
}

// Fund account with the given amount of AptosCoin
func (faucetClient *FaucetClient) Fund(address core.AccountAddress, amount uint64) error {
	if faucetClient.nodeClient == nil {
		return errors.New("faucet's node-client not initialized")
	}
	mintUrl := faucetClient.url.JoinPath("mint")
	params := url.Values{}
	params.Set("amount", strconv.FormatUint(amount, 10))
	params.Set("address", address.String())
	mintUrl.RawQuery = params.Encode()
	response, err := faucetClient.nodeClient.Post(mintUrl.String(), "text/plain", nil)
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return util.NewHttpError(response)
	}
	decoder := json.NewDecoder(response.Body)
	var txnHashes []string
	err = decoder.Decode(&txnHashes)
	if err != nil {
		return fmt.Errorf("response json decode error, %w", err)
	}
	if faucetClient.nodeClient == nil {
		slog.Debug("FundAccount no txns to wait for")
		// no Aptos client to wait on txn completion
		return nil
	}
	slog.Debug("FundAccount wait for txns", "ntxn", len(txnHashes))
	if len(txnHashes) == 1 {
		_, err = faucetClient.nodeClient.WaitForTransaction(txnHashes[0])
		return err
	} else {
		return faucetClient.nodeClient.PollForTransactions(txnHashes)
	}
}
