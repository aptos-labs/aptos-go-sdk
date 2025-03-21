package aptos

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
)

// FaucetClient uses the underlying NodeClient to request for APT for gas on a network.
// This can only be used in a test network (e.g. Localnet, Devnet, Testnet)
type FaucetClient struct {
	nodeClient *NodeClient // NodeClient to use for requesting funds
	url        *url.URL    // URL of the faucet e.g. https://testnet.faucet.aptoslabs.com
}

// NewFaucetClient creates a new client specifically for requesting faucet funds
func NewFaucetClient(nodeClient *NodeClient, faucetUrl string) (*FaucetClient, error) {
	parsedUrl, err := url.Parse(faucetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse faucet url '%s': %w", faucetUrl, err)
	}
	return &FaucetClient{
		nodeClient,
		parsedUrl,
	}, nil
}

// Fund account with the given amount of AptosCoin
func (faucetClient *FaucetClient) Fund(address AccountAddress, amount uint64) error {
	if faucetClient.nodeClient == nil {
		return errors.New("faucet's node-client not initialized")
	}

	// Build URL
	mintUrl := faucetClient.url.JoinPath("mint")
	params := url.Values{}
	params.Set("amount", strconv.FormatUint(amount, 10))
	params.Set("address", address.String())
	mintUrl.RawQuery = params.Encode()

	// Make request for funds
	txnHashes, err := Post[[]string](faucetClient.nodeClient, mintUrl.String(), "text/plain", nil)
	if err != nil {
		return fmt.Errorf("response api decode error, %w", err)
	}

	// Wait for fund transactions to go through
	slog.Debug("FundAccount wait for transactions", "number of transactions", len(txnHashes))
	if len(txnHashes) == 1 {
		_, err = faucetClient.nodeClient.WaitForTransaction(txnHashes[0])
		return err
	}

	return faucetClient.nodeClient.PollForTransactions(txnHashes)
}
