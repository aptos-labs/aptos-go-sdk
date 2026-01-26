package keyless

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultProverURL is the default Aptos prover service URL.
const DefaultProverURL = "https://prover.aptoslabs.com"

// ProverClient is used to fetch ZK proofs from the prover service.
type ProverClient struct {
	url        string
	httpClient *http.Client
}

// NewProverClient creates a new prover client.
func NewProverClient(url string) *ProverClient {
	if url == "" {
		url = DefaultProverURL
	}
	return &ProverClient{
		url: url,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // ZK proof generation can take time
		},
	}
}

// WithHTTPClient sets a custom HTTP client for the prover.
func (p *ProverClient) WithHTTPClient(client *http.Client) *ProverClient {
	p.httpClient = client
	return p
}

// ProveRequest represents a request to the prover service.
type ProveRequest struct {
	// JWT is the OIDC JWT token.
	JWT string `json:"jwt"`

	// EphemeralPublicKey is the ephemeral public key bytes.
	EphemeralPublicKey []byte `json:"ephemeral_public_key"`

	// ExpiryDateSecs is when the ephemeral key expires (Unix timestamp).
	ExpiryDateSecs uint64 `json:"expiry_date_secs"`

	// Blinder is the random blinder used in the nonce.
	Blinder []byte `json:"blinder"`

	// UIDKey is the claim key used for identity (e.g., "sub" or "email").
	UIDKey string `json:"uid_key"`

	// Pepper is the pepper used for address derivation.
	Pepper []byte `json:"pepper"`

	// KeylessConfiguration is the on-chain configuration hash.
	KeylessConfiguration *KeylessConfiguration `json:"keyless_configuration,omitempty"`
}

// KeylessConfiguration represents the on-chain keyless configuration.
type KeylessConfiguration struct {
	// OverrideAudVal allows overriding the audience value.
	OverrideAudVal string `json:"override_aud_val,omitempty"`
}

// ProveResponse represents a response from the prover service.
type ProveResponse struct {
	// Proof is the zero-knowledge proof.
	Proof *ZKProof `json:"proof"`

	// TrainingWheelsSignature is an optional fallback signature.
	TrainingWheelsSignature []byte `json:"training_wheels_signature,omitempty"`

	// Error is set if the prover encountered an error.
	Error string `json:"error,omitempty"`
}

// GetProof requests a ZK proof from the prover service.
func (p *ProverClient) GetProof(ctx context.Context, req *ProveRequest) (*ProveResponse, error) {
	// Validate request
	if req.JWT == "" {
		return nil, errors.New("JWT is required")
	}
	if len(req.EphemeralPublicKey) == 0 {
		return nil, errors.New("ephemeral public key is required")
	}
	if req.UIDKey == "" {
		req.UIDKey = "sub"
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url+"/v0/prove", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%w: %s", ErrProverFailed, errResp.Error)
		}
		return nil, fmt.Errorf("%w: status %d: %s", ErrProverFailed, resp.StatusCode, string(respBody))
	}

	// Parse response
	var proveResp ProveResponse
	if err := json.Unmarshal(respBody, &proveResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if proveResp.Error != "" {
		return nil, fmt.Errorf("%w: %s", ErrProverFailed, proveResp.Error)
	}

	return &proveResp, nil
}

// PepperClient is used to fetch peppers from the pepper service.
type PepperClient struct {
	url        string
	httpClient *http.Client
}

// DefaultPepperURL is the default Aptos pepper service URL.
const DefaultPepperURL = "https://pepper.aptoslabs.com"

// NewPepperClient creates a new pepper client.
func NewPepperClient(url string) *PepperClient {
	if url == "" {
		url = DefaultPepperURL
	}
	return &PepperClient{
		url: url,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PepperRequest represents a request to the pepper service.
type PepperRequest struct {
	// JWT is the OIDC JWT token.
	JWT string `json:"jwt"`

	// EphemeralPublicKey is the ephemeral public key bytes.
	EphemeralPublicKey []byte `json:"ephemeral_public_key"`

	// ExpiryDateSecs is when the ephemeral key expires (Unix timestamp).
	ExpiryDateSecs uint64 `json:"expiry_date_secs"`

	// UIDKey is the claim key used for identity.
	UIDKey string `json:"uid_key"`

	// Blinder is the random blinder used in the nonce.
	Blinder []byte `json:"blinder"`
}

// PepperResponse represents a response from the pepper service.
type PepperResponse struct {
	// Pepper is the derived pepper for address computation.
	Pepper []byte `json:"pepper"`

	// Error is set if the service encountered an error.
	Error string `json:"error,omitempty"`
}

// GetPepper fetches a pepper from the pepper service.
// The pepper service derives a deterministic pepper from the JWT claims,
// ensuring the same identity always gets the same pepper (and thus address).
func (p *PepperClient) GetPepper(ctx context.Context, req *PepperRequest) (*PepperResponse, error) {
	// Validate request
	if req.JWT == "" {
		return nil, errors.New("JWT is required")
	}
	if len(req.EphemeralPublicKey) == 0 {
		return nil, errors.New("ephemeral public key is required")
	}
	if req.UIDKey == "" {
		req.UIDKey = "sub"
	}

	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url+"/v0/fetch", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("pepper service error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("pepper service error: status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var pepperResp PepperResponse
	if err := json.Unmarshal(respBody, &pepperResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if pepperResp.Error != "" {
		return nil, fmt.Errorf("pepper service error: %s", pepperResp.Error)
	}

	return &pepperResp, nil
}
