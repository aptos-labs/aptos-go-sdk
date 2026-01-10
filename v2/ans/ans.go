package ans

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

// Common errors
var (
	// ErrNameNotFound indicates the requested name is not registered.
	ErrNameNotFound = errors.New("ans: name not found")

	// ErrNameExpired indicates the name registration has expired.
	ErrNameExpired = errors.New("ans: name has expired")

	// ErrInvalidName indicates the name format is invalid.
	ErrInvalidName = errors.New("ans: invalid name format")

	// ErrNameTaken indicates the name is already registered.
	ErrNameTaken = errors.New("ans: name is already registered")

	// ErrNoPermission indicates insufficient permissions for the operation.
	ErrNoPermission = errors.New("ans: insufficient permissions")
)

// Default ANS contract addresses
var (
	// RouterAddress is the ANS router contract address on mainnet.
	RouterAddress = aptos.MustParseAddress("0x867ed1f6bf916171b1de3ee92849b8978b7d1b9e0a8cc982a3d19d535dfd9c0c")

	// TestnetRouterAddress is the ANS router contract address on testnet.
	TestnetRouterAddress = aptos.MustParseAddress("0x5f8fd2347449685cf41d4db97926ec3a096eaf381332be4f1318ad4d16a8497c")
)

// TLD is the top-level domain for Aptos names.
const TLD = ".apt"

// Name validation regex
var nameRegex = regexp.MustCompile(`^[a-z0-9-]{3,63}$`)

// Client provides ANS resolution and management functionality.
type Client struct {
	client        aptos.Client
	routerAddress aptos.AccountAddress
}

// NewClient creates a new ANS client.
func NewClient(client aptos.Client) *Client {
	return &Client{
		client:        client,
		routerAddress: RouterAddress,
	}
}

// NewTestnetClient creates a new ANS client configured for testnet.
func NewTestnetClient(client aptos.Client) *Client {
	return &Client{
		client:        client,
		routerAddress: TestnetRouterAddress,
	}
}

// WithRouterAddress sets a custom router address.
func (c *Client) WithRouterAddress(addr aptos.AccountAddress) *Client {
	c.routerAddress = addr
	return c
}

// Name represents a parsed ANS name.
type Name struct {
	// Domain is the primary domain (e.g., "alice" from "alice.apt").
	Domain string

	// Subdomain is the optional subdomain (e.g., "wallet" from "wallet.alice.apt").
	Subdomain string
}

// String returns the full name with TLD.
func (n Name) String() string {
	if n.Subdomain != "" {
		return n.Subdomain + "." + n.Domain + TLD
	}
	return n.Domain + TLD
}

// ParseName parses an ANS name string.
func ParseName(name string) (*Name, error) {
	// Normalize: lowercase and remove TLD if present
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, TLD)

	parts := strings.Split(name, ".")

	switch len(parts) {
	case 1:
		// Primary domain: "alice"
		if !isValidLabel(parts[0]) {
			return nil, fmt.Errorf("%w: invalid domain '%s'", ErrInvalidName, parts[0])
		}
		return &Name{Domain: parts[0]}, nil

	case 2:
		// Subdomain: "wallet.alice"
		if !isValidLabel(parts[0]) {
			return nil, fmt.Errorf("%w: invalid subdomain '%s'", ErrInvalidName, parts[0])
		}
		if !isValidLabel(parts[1]) {
			return nil, fmt.Errorf("%w: invalid domain '%s'", ErrInvalidName, parts[1])
		}
		return &Name{Domain: parts[1], Subdomain: parts[0]}, nil

	default:
		return nil, fmt.Errorf("%w: too many parts in name", ErrInvalidName)
	}
}

func isValidLabel(label string) bool {
	return nameRegex.MatchString(label)
}

// NameInfo contains information about a registered name.
type NameInfo struct {
	// Name is the parsed name.
	Name Name

	// Owner is the address that owns this name.
	Owner aptos.AccountAddress

	// Target is the address this name resolves to (may differ from owner).
	Target aptos.AccountAddress

	// ExpiresAt is when the name registration expires.
	ExpiresAt time.Time

	// IsPrimary indicates if this is the owner's primary name.
	IsPrimary bool
}

// IsExpired returns true if the name has expired.
func (n *NameInfo) IsExpired() bool {
	return time.Now().After(n.ExpiresAt)
}

// Resolve resolves an ANS name to an address.
func (c *Client) Resolve(ctx context.Context, name string) (aptos.AccountAddress, error) {
	parsed, err := ParseName(name)
	if err != nil {
		return aptos.AccountAddress{}, err
	}

	info, err := c.GetNameInfo(ctx, *parsed)
	if err != nil {
		return aptos.AccountAddress{}, err
	}

	if info.IsExpired() {
		return aptos.AccountAddress{}, ErrNameExpired
	}

	return info.Target, nil
}

// GetNameInfo returns detailed information about a name.
func (c *Client) GetNameInfo(ctx context.Context, name Name) (*NameInfo, error) {
	// Call the router view function to get name info
	payload := &aptos.ViewPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "get_target_addr",
		TypeArgs: nil,
		Args:     []any{name.Domain, name.Subdomain},
	}

	result, err := c.client.View(ctx, payload)
	if err != nil {
		if aptos.IsNotFound(err) {
			return nil, ErrNameNotFound
		}
		return nil, fmt.Errorf("failed to query name: %w", err)
	}

	if len(result) == 0 {
		return nil, ErrNameNotFound
	}

	// Parse the result
	// The view function returns Option<address>
	optResult, ok := result[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	vecData, ok := optResult["vec"].([]interface{})
	if !ok || len(vecData) == 0 {
		return nil, ErrNameNotFound
	}

	addrStr, ok := vecData[0].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected address format")
	}

	target, err := aptos.ParseAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid target address: %w", err)
	}

	// Get expiration info
	expiresAt, err := c.getExpiration(ctx, name)
	if err != nil {
		// If we can't get expiration, use a far future date
		expiresAt = time.Now().Add(100 * 365 * 24 * time.Hour)
	}

	return &NameInfo{
		Name:      name,
		Target:    target,
		ExpiresAt: expiresAt,
	}, nil
}

func (c *Client) getExpiration(ctx context.Context, name Name) (time.Time, error) {
	payload := &aptos.ViewPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "get_expiration",
		TypeArgs: nil,
		Args:     []any{name.Domain, name.Subdomain},
	}

	result, err := c.client.View(ctx, payload)
	if err != nil {
		return time.Time{}, err
	}

	if len(result) == 0 {
		return time.Time{}, fmt.Errorf("no expiration returned")
	}

	// Parse expiration timestamp
	expStr, ok := result[0].(string)
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected expiration format")
	}

	var expSecs int64
	if _, err := fmt.Sscanf(expStr, "%d", &expSecs); err != nil {
		return time.Time{}, fmt.Errorf("invalid expiration: %w", err)
	}

	return time.Unix(expSecs, 0), nil
}

// GetPrimaryName returns the primary ANS name for an address.
func (c *Client) GetPrimaryName(ctx context.Context, address aptos.AccountAddress) (string, error) {
	payload := &aptos.ViewPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "get_primary_name",
		TypeArgs: nil,
		Args:     []any{address.String()},
	}

	result, err := c.client.View(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("failed to query primary name: %w", err)
	}

	if len(result) < 2 {
		return "", ErrNameNotFound
	}

	// Result is [domain_option, subdomain_option]
	domainOpt, ok := result[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected domain format")
	}

	domainVec, ok := domainOpt["vec"].([]interface{})
	if !ok || len(domainVec) == 0 {
		return "", ErrNameNotFound
	}

	domain, ok := domainVec[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected domain string format")
	}

	// Check for subdomain
	subdomainOpt, ok := result[1].(map[string]interface{})
	if ok {
		subdomainVec, ok := subdomainOpt["vec"].([]interface{})
		if ok && len(subdomainVec) > 0 {
			subdomain, ok := subdomainVec[0].(string)
			if ok && subdomain != "" {
				return subdomain + "." + domain + TLD, nil
			}
		}
	}

	return domain + TLD, nil
}

// IsAvailable checks if a name is available for registration.
func (c *Client) IsAvailable(ctx context.Context, name string) (bool, error) {
	parsed, err := ParseName(name)
	if err != nil {
		return false, err
	}

	_, err = c.GetNameInfo(ctx, *parsed)
	if err != nil {
		if errors.Is(err, ErrNameNotFound) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

// RegisterOptions contains options for name registration.
type RegisterOptions struct {
	// Years is the number of years to register the name for.
	Years int

	// Target is the address the name should resolve to.
	// If empty, defaults to the registrant's address.
	Target *aptos.AccountAddress
}

// RegisterPayload returns the payload for registering a name.
// The transaction must be signed and submitted separately.
func (c *Client) RegisterPayload(name string, opts RegisterOptions) (*aptos.EntryFunctionPayload, error) {
	parsed, err := ParseName(name)
	if err != nil {
		return nil, err
	}

	if parsed.Subdomain != "" {
		return nil, fmt.Errorf("%w: cannot register subdomains directly", ErrInvalidName)
	}

	if opts.Years <= 0 {
		opts.Years = 1
	}

	return &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "register_domain",
		TypeArgs: nil,
		Args:     []any{parsed.Domain, opts.Years},
	}, nil
}

// SetPrimaryNamePayload returns the payload for setting a primary name.
func (c *Client) SetPrimaryNamePayload(name string) (*aptos.EntryFunctionPayload, error) {
	parsed, err := ParseName(name)
	if err != nil {
		return nil, err
	}

	subdomain := ""
	if parsed.Subdomain != "" {
		subdomain = parsed.Subdomain
	}

	return &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "set_primary_name",
		TypeArgs: nil,
		Args:     []any{parsed.Domain, subdomain},
	}, nil
}

// SetTargetAddressPayload returns the payload for setting the target address.
func (c *Client) SetTargetAddressPayload(name string, target aptos.AccountAddress) (*aptos.EntryFunctionPayload, error) {
	parsed, err := ParseName(name)
	if err != nil {
		return nil, err
	}

	subdomain := ""
	if parsed.Subdomain != "" {
		subdomain = parsed.Subdomain
	}

	return &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "set_target_addr",
		TypeArgs: nil,
		Args:     []any{parsed.Domain, subdomain, target.String()},
	}, nil
}

// RenewPayload returns the payload for renewing a name.
func (c *Client) RenewPayload(name string, years int) (*aptos.EntryFunctionPayload, error) {
	parsed, err := ParseName(name)
	if err != nil {
		return nil, err
	}

	if parsed.Subdomain != "" {
		return nil, fmt.Errorf("%w: cannot renew subdomains directly", ErrInvalidName)
	}

	if years <= 0 {
		years = 1
	}

	return &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "renew_domain",
		TypeArgs: nil,
		Args:     []any{parsed.Domain, years},
	}, nil
}

// AddSubdomainPayload returns the payload for adding a subdomain.
func (c *Client) AddSubdomainPayload(domain, subdomain string, target aptos.AccountAddress) (*aptos.EntryFunctionPayload, error) {
	parsed, err := ParseName(domain)
	if err != nil {
		return nil, err
	}

	if !isValidLabel(subdomain) {
		return nil, fmt.Errorf("%w: invalid subdomain '%s'", ErrInvalidName, subdomain)
	}

	return &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: c.routerAddress, Name: "router"},
		Function: "register_subdomain",
		TypeArgs: nil,
		Args:     []any{parsed.Domain, subdomain, target.String(), 0, false},
	}, nil
}
