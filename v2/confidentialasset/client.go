package confidentialasset

import (
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/v2"
)

// Client wraps aptos.Client with the confidential_asset module address (same role as TS Aptos + module string).
type Client struct {
	Aptos         aptos.Client
	ModuleAddress aptos.AccountAddress
	WithFeePayer  bool
	// RESTBaseURL is the fullnode REST base with /v1 suffix, no trailing slash (e.g. https://fullnode.testnet.aptoslabs.com/v1).
	// Used to read public FA balance for gas budgeting in SubmitWithSimulatedGas.
	RESTBaseURL string
}

// NewClient uses framework address 0x1 for confidential_asset (override with opts).
func NewClient(aptosClient aptos.Client, opts ...ClientOption) *Client {
	c := &Client{Aptos: aptosClient, ModuleAddress: aptos.AccountOne}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ClientOption configures Client.
type ClientOption func(*Client)

// WithModuleAddress sets the account hosting confidential_asset (default 0x1).
func WithModuleAddress(addr aptos.AccountAddress) ClientOption {
	return func(c *Client) {
		c.ModuleAddress = addr
	}
}

// WithFeePayer matches TS ConfidentialAsset.withFeePayer (submit path must set fee payer on txn when true).
func WithFeePayer(v bool) ClientOption {
	return func(c *Client) {
		c.WithFeePayer = v
	}
}

// WithRESTBaseURL sets the REST /v1 base URL for FA balance reads used in gas estimation.
func WithRESTBaseURL(base string) ClientOption {
	return func(c *Client) {
		c.RESTBaseURL = strings.TrimSuffix(strings.TrimSpace(base), "/")
	}
}

// ViewModule returns the Move module ID for confidential_asset views and entry functions.
func (c *Client) ViewModule() aptos.ModuleID {
	return aptos.ModuleID{Address: c.ModuleAddress, Name: ModuleName}
}
