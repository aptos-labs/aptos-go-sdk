package sponsored

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// Common errors
var (
	// ErrNoFeePayer indicates the transaction doesn't have a fee payer set.
	ErrNoFeePayer = errors.New("sponsored: no fee payer set on transaction")

	// ErrSponsorRejected indicates the sponsor rejected the transaction.
	ErrSponsorRejected = errors.New("sponsored: sponsor rejected transaction")

	// ErrInsufficientFunds indicates the sponsor has insufficient funds.
	ErrInsufficientFunds = errors.New("sponsored: sponsor has insufficient funds")

	// ErrRateLimited indicates the sponsor service rate limited the request.
	ErrRateLimited = errors.New("sponsored: rate limited")
)

// FeePayerRawTransaction wraps a RawTransaction with fee payer information.
type FeePayerRawTransaction struct {
	// RawTxn is the underlying transaction.
	RawTxn *aptos.RawTransaction

	// SecondarySigners are addresses of additional signers (for multi-agent).
	SecondarySigners []aptos.AccountAddress

	// FeePayer is the address of the fee payer.
	FeePayer aptos.AccountAddress
}

// NewFeePayerRawTransaction creates a new fee payer raw transaction.
func NewFeePayerRawTransaction(rawTxn *aptos.RawTransaction, feePayer aptos.AccountAddress) *FeePayerRawTransaction {
	return &FeePayerRawTransaction{
		RawTxn:   rawTxn,
		FeePayer: feePayer,
	}
}

// WithSecondarySigners adds secondary signers for multi-agent transactions.
func (f *FeePayerRawTransaction) WithSecondarySigners(addrs ...aptos.AccountAddress) *FeePayerRawTransaction {
	f.SecondarySigners = append(f.SecondarySigners, addrs...)
	return f
}

// SigningMessage creates the message that needs to be signed by all parties.
func (f *FeePayerRawTransaction) SigningMessage() ([]byte, error) {
	// Use the RawTransactionWithData prehash
	prehash := aptos.RawTransactionWithDataPrehash()

	// Serialize the fee payer transaction data
	ser := bcs.NewSerializer()

	// Variant: 0 = MultiAgent, 1 = FeePayerWithData
	ser.Uleb128(1) // FeePayerWithData variant

	// Serialize raw transaction
	f.RawTxn.MarshalBCS(ser)

	// Secondary signers
	bcs.SerializeSequenceFunc(ser, f.SecondarySigners, func(s *bcs.Serializer, addr aptos.AccountAddress) {
		addr.MarshalBCS(s)
	})

	// Fee payer address
	f.FeePayer.MarshalBCS(ser)

	if ser.Error() != nil {
		return nil, fmt.Errorf("failed to serialize fee payer transaction: %w", ser.Error())
	}

	return append(prehash, ser.ToBytes()...), nil
}

// Sign creates a signed transaction with all the provided authenticators.
func (f *FeePayerRawTransaction) Sign(
	sender *aptos.AccountAuthenticator,
	secondarySigners []*aptos.AccountAuthenticator,
	feePayerAuth *aptos.AccountAuthenticator,
) (*aptos.SignedTransaction, error) {
	return aptos.NewFeePayerSignedTransaction(
		f.RawTxn,
		sender,
		f.SecondarySigners,
		secondarySigners,
		f.FeePayer,
		feePayerAuth,
	)
}

// GasStation provides integration with a gas station sponsorship service.
type GasStation struct {
	url        string
	apiKey     string
	httpClient *http.Client
}

// GasStationOption configures a GasStation.
type GasStationOption func(*GasStation)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) GasStationOption {
	return func(g *GasStation) {
		g.httpClient = client
	}
}

// NewGasStation creates a new gas station client.
func NewGasStation(url, apiKey string, opts ...GasStationOption) *GasStation {
	gs := &GasStation{
		url:    url,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(gs)
	}
	return gs
}

// SponsorRequest contains the data needed for sponsorship.
type SponsorRequest struct {
	// RawTxnBCS is the BCS-serialized raw transaction.
	RawTxnBCS []byte `json:"raw_txn_bcs"`

	// SenderAuthBCS is the BCS-serialized sender authenticator.
	SenderAuthBCS []byte `json:"sender_auth_bcs"`

	// SecondarySigners are addresses of secondary signers.
	SecondarySigners []string `json:"secondary_signers,omitempty"`

	// SecondaryAuthsBCS are BCS-serialized secondary authenticators.
	SecondaryAuthsBCS [][]byte `json:"secondary_auths_bcs,omitempty"`
}

// SponsorResponse contains the sponsor's response.
type SponsorResponse struct {
	// SponsorAddress is the sponsor's address.
	SponsorAddress string `json:"sponsor_address"`

	// SponsorAuthBCS is the BCS-serialized sponsor authenticator.
	SponsorAuthBCS []byte `json:"sponsor_auth_bcs"`

	// ErrorCode is set if sponsorship failed.
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage provides details about failures.
	ErrorMessage string `json:"error_message,omitempty"`
}

// SponsorTransaction requests sponsorship for a transaction.
func (g *GasStation) SponsorTransaction(
	ctx context.Context,
	rawTxn *aptos.RawTransaction,
	senderAuth *aptos.AccountAuthenticator,
	secondarySigners []aptos.AccountAddress,
	secondaryAuths []*aptos.AccountAuthenticator,
) (*aptos.AccountAuthenticator, aptos.AccountAddress, error) {
	// Serialize raw transaction
	rawTxnBCS, err := bcs.Serialize(rawTxn)
	if err != nil {
		return nil, aptos.AccountAddress{}, fmt.Errorf("failed to serialize raw transaction: %w", err)
	}

	// Serialize sender auth
	senderAuthBCS, err := bcs.Serialize(senderAuth)
	if err != nil {
		return nil, aptos.AccountAddress{}, fmt.Errorf("failed to serialize sender auth: %w", err)
	}

	// Build request
	req := SponsorRequest{
		RawTxnBCS:     rawTxnBCS,
		SenderAuthBCS: senderAuthBCS,
	}

	// Add secondary signers if present
	if len(secondarySigners) > 0 {
		req.SecondarySigners = make([]string, len(secondarySigners))
		for i, addr := range secondarySigners {
			req.SecondarySigners[i] = addr.String()
		}

		req.SecondaryAuthsBCS = make([][]byte, len(secondaryAuths))
		for i, auth := range secondaryAuths {
			authBCS, err := bcs.Serialize(auth)
			if err != nil {
				return nil, aptos.AccountAddress{}, fmt.Errorf("failed to serialize secondary auth %d: %w", i, err)
			}
			req.SecondaryAuthsBCS[i] = authBCS
		}
	}

	// Send request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, aptos.AccountAddress{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url+"/sponsor", bytes.NewReader(reqBody))
	if err != nil {
		return nil, aptos.AccountAddress{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, aptos.AccountAddress{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, aptos.AccountAddress{}, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, aptos.AccountAddress{}, ErrRateLimited
	}

	var sponsorResp SponsorResponse
	if err := json.Unmarshal(body, &sponsorResp); err != nil {
		return nil, aptos.AccountAddress{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if sponsorResp.ErrorCode != "" {
		switch sponsorResp.ErrorCode {
		case "REJECTED":
			return nil, aptos.AccountAddress{}, fmt.Errorf("%w: %s", ErrSponsorRejected, sponsorResp.ErrorMessage)
		case "INSUFFICIENT_FUNDS":
			return nil, aptos.AccountAddress{}, fmt.Errorf("%w: %s", ErrInsufficientFunds, sponsorResp.ErrorMessage)
		default:
			return nil, aptos.AccountAddress{}, fmt.Errorf("sponsor error: %s - %s", sponsorResp.ErrorCode, sponsorResp.ErrorMessage)
		}
	}

	// Parse sponsor address
	sponsorAddr, err := aptos.ParseAddress(sponsorResp.SponsorAddress)
	if err != nil {
		return nil, aptos.AccountAddress{}, fmt.Errorf("invalid sponsor address: %w", err)
	}

	// Deserialize sponsor auth
	sponsorAuth := &aptos.AccountAuthenticator{}
	if err := bcs.Deserialize(sponsorAuth, sponsorResp.SponsorAuthBCS); err != nil {
		return nil, aptos.AccountAddress{}, fmt.Errorf("failed to deserialize sponsor auth: %w", err)
	}

	return sponsorAuth, sponsorAddr, nil
}

// Helper functions for building fee payer transactions

// BuildFeePayerTransaction builds a transaction with fee payer.
func BuildFeePayerTransaction(
	ctx context.Context,
	client aptos.Client,
	sender aptos.AccountAddress,
	feePayer aptos.AccountAddress,
	payload aptos.Payload,
	opts ...aptos.TransactionOption,
) (*FeePayerRawTransaction, error) {
	// Add fee payer option
	opts = append(opts, aptos.WithFeePayer(feePayer))

	// Build raw transaction
	rawTxn, err := client.BuildTransaction(ctx, sender, payload, opts...)
	if err != nil {
		return nil, err
	}

	return NewFeePayerRawTransaction(rawTxn, feePayer), nil
}

// SignAndSubmitSponsoredTransaction is a convenience function for simple sponsored transactions.
func SignAndSubmitSponsoredTransaction(
	ctx context.Context,
	client aptos.Client,
	sender aptos.TransactionSigner,
	sponsor aptos.TransactionSigner,
	payload aptos.Payload,
	opts ...aptos.TransactionOption,
) (*aptos.SubmitResult, error) {
	// Build fee payer transaction
	fpTxn, err := BuildFeePayerTransaction(ctx, client, sender.Address(), sponsor.Address(), payload, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Get signing message
	signingMsg, err := fpTxn.SigningMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to get signing message: %w", err)
	}

	// Sign with sender
	senderAuth, err := sender.Sign(signingMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign as sender: %w", err)
	}

	// Sign with sponsor
	sponsorAuth, err := sponsor.Sign(signingMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign as sponsor: %w", err)
	}

	// Create signed transaction
	signedTxn, err := fpTxn.Sign(senderAuth, nil, sponsorAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed transaction: %w", err)
	}

	// Submit
	return client.SubmitTransaction(ctx, signedTxn)
}
