package aptos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// modulesPageSize is the number of modules requested per page when listing an
// account's modules. The node caps page sizes, so this is only an upper bound.
const modulesPageSize = 100

// TableItemRequest identifies a single item within an on-chain Move table.
//
// KeyType and ValueType are the fully-qualified Move types of the table's key
// and value (for example "address", "u64", or "0x1::string::String"). Key is
// the key value encoded the way the node's JSON API expects it — a string for
// integers and addresses, or a nested object for struct keys.
type TableItemRequest struct {
	KeyType   string `json:"key_type"`
	ValueType string `json:"value_type"`
	Key       any    `json:"key"`
}

// GetTableItem reads a single item from an on-chain Move table by its handle
// and key, decoding the JSON-encoded value into result.
//
// Reading table items directly is discouraged for most applications: table
// handles are an implementation detail of the Move modules that own them and
// can change, and the indexer usually exposes the same data in a more stable
// form. It is provided for the cases where a direct read is unavoidable.
func (c *nodeClient) GetTableItem(ctx context.Context, handle string, req TableItemRequest, result any, opts ...ResourceOption) error {
	config := &ResourceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	path := fmt.Sprintf("tables/%s/item", url.PathEscape(handle))
	if config.LedgerVersion != nil {
		path += fmt.Sprintf("?ledger_version=%d", *config.LedgerVersion)
	}

	return c.post(ctx, path, req, result)
}

// AccountBalanceOf returns an account's balance of an arbitrary asset.
//
// asset may be either a coin type's fully-qualified struct id (for example
// "0x1::aptos_coin::AptosCoin") or the address of a fungible asset's metadata
// object (for example "0xa" for APT). The node transparently handles both the
// legacy coin representation and the fungible-asset representation, so this
// works for coins that have migrated to the fungible-asset standard.
//
// For the plain APT balance, AccountBalance is a convenient shorthand.
func (c *nodeClient) AccountBalanceOf(ctx context.Context, address AccountAddress, asset string, opts ...ResourceOption) (uint64, error) {
	config := &ResourceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	path := fmt.Sprintf("accounts/%s/balance/%s", address.String(), url.PathEscape(asset))
	if config.LedgerVersion != nil {
		path += fmt.Sprintf("?ledger_version=%d", *config.LedgerVersion)
	}

	var value any
	if err := c.get(ctx, path, &value); err != nil {
		return 0, err
	}
	return parseU64Balance(value)
}

// AccountModules returns every Move module published under an address.
//
// The node returns modules one page at a time and reports the next page's
// cursor in the X-Aptos-Cursor response header; this method follows that
// cursor until the account is exhausted and returns the concatenated result.
func (c *nodeClient) AccountModules(ctx context.Context, address AccountAddress, opts ...ResourceOption) ([]*ModuleBytecode, error) {
	config := &ResourceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	var modules []*ModuleBytecode
	cursor := ""
	for {
		params := url.Values{}
		params.Set("limit", strconv.Itoa(modulesPageSize))
		if cursor != "" {
			params.Set("start", cursor)
		}
		if config.LedgerVersion != nil {
			params.Set("ledger_version", strconv.FormatUint(*config.LedgerVersion, 10))
		}
		path := fmt.Sprintf("accounts/%s/modules?%s", address.String(), params.Encode())

		var page []*ModuleBytecode
		headers, err := c.getReturningHeaders(ctx, path, &page)
		if err != nil {
			return nil, err
		}
		modules = append(modules, page...)

		cursor = headers.Get("X-Aptos-Cursor")
		if cursor == "" {
			return modules, nil
		}
	}
}

// getReturningHeaders issues a GET and returns the response headers alongside
// the decoded body, for endpoints paginated via a cursor header.
func (c *nodeClient) getReturningHeaders(ctx context.Context, path string, result any) (http.Header, error) {
	reqURL := c.buildURL(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return c.doRequestReturningHeaders(ctx, req, result)
}
