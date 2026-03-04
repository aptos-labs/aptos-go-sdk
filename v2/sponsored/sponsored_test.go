package sponsored

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/account"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestRawTransaction() *aptos.RawTransaction {
	return &aptos.RawTransaction{
		Sender:                     aptos.AccountAddress{},
		SequenceNumber:             1,
		Payload:                    &aptos.EntryFunctionPayload{Module: aptos.ModuleID{Address: aptos.AccountAddress{}, Name: "coin"}, Function: "transfer"},
		MaxGasAmount:               2000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 9999999999,
		ChainID:                    1,
	}
}

func TestNewFeePayerRawTransaction(t *testing.T) {
	t.Parallel()

	rawTxn := makeTestRawTransaction()
	feePayer := aptos.AccountAddress{}
	feePayer[31] = 0x01

	fpTxn := NewFeePayerRawTransaction(rawTxn, feePayer)
	require.NotNil(t, fpTxn)
	assert.Equal(t, rawTxn, fpTxn.RawTxn)
	assert.Equal(t, feePayer, fpTxn.FeePayer)
	assert.Nil(t, fpTxn.SecondarySigners)
}

func TestWithSecondarySigners(t *testing.T) {
	t.Parallel()

	rawTxn := makeTestRawTransaction()
	fpTxn := NewFeePayerRawTransaction(rawTxn, aptos.AccountAddress{})

	signer1 := aptos.AccountAddress{}
	signer1[31] = 0x01
	signer2 := aptos.AccountAddress{}
	signer2[31] = 0x02

	fpTxn.WithSecondarySigners(signer1).WithSecondarySigners(signer2)
	assert.Len(t, fpTxn.SecondarySigners, 2)
	assert.Equal(t, signer1, fpTxn.SecondarySigners[0])
	assert.Equal(t, signer2, fpTxn.SecondarySigners[1])
}

func TestFeePayerRawTransaction_SigningMessage(t *testing.T) {
	t.Parallel()

	rawTxn := makeTestRawTransaction()
	fpTxn := NewFeePayerRawTransaction(rawTxn, aptos.AccountAddress{})

	msg, err := fpTxn.SigningMessage()
	require.NoError(t, err)
	assert.NotEmpty(t, msg)

	// Should start with the prehash
	prehash := aptos.RawTransactionWithDataPrehash()
	assert.Equal(t, prehash, msg[:len(prehash)])
}

func TestNewGasStation(t *testing.T) {
	t.Parallel()

	gs := NewGasStation("http://example.com", "api-key")
	assert.NotNil(t, gs)
}

func TestNewGasStation_WithHTTPClient(t *testing.T) {
	t.Parallel()

	customClient := &http.Client{}
	gs := NewGasStation("http://example.com", "api-key", WithHTTPClient(customClient))
	assert.NotNil(t, gs)
}

func TestGasStation_SponsorTransaction_Success(t *testing.T) {
	t.Parallel()

	// Create a proper authenticator
	privKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	auth := &aptos.AccountAuthenticator{}
	err = auth.FromKeyAndSignature(privKey.PubKey(), &crypto.Ed25519Signature{})
	require.NoError(t, err)

	authBCS, err := bcs.Serialize(auth)
	require.NoError(t, err)

	sponsorAddr := aptos.AccountAddress{}
	sponsorAddr[31] = 0xaa

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sponsor", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		resp := SponsorResponse{
			SponsorAddress: sponsorAddr.String(),
			SponsorAuthBCS: authBCS,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	gs := NewGasStation(server.URL, "test-key")
	rawTxn := makeTestRawTransaction()

	sponsorAuth, addr, err := gs.SponsorTransaction(context.Background(), rawTxn, auth, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, sponsorAuth)
	assert.Equal(t, sponsorAddr, addr)
}

func TestGasStation_SponsorTransaction_RateLimited(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	gs := NewGasStation(server.URL, "test-key")
	rawTxn := makeTestRawTransaction()

	privKey, _ := crypto.GenerateEd25519PrivateKey()
	auth := &aptos.AccountAuthenticator{}
	_ = auth.FromKeyAndSignature(privKey.PubKey(), &crypto.Ed25519Signature{})

	_, _, err := gs.SponsorTransaction(context.Background(), rawTxn, auth, nil, nil)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestGasStation_SponsorTransaction_Rejected(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SponsorResponse{
			ErrorCode:    "REJECTED",
			ErrorMessage: "not allowed",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	gs := NewGasStation(server.URL, "test-key")
	rawTxn := makeTestRawTransaction()

	privKey, _ := crypto.GenerateEd25519PrivateKey()
	auth := &aptos.AccountAuthenticator{}
	_ = auth.FromKeyAndSignature(privKey.PubKey(), &crypto.Ed25519Signature{})

	_, _, err := gs.SponsorTransaction(context.Background(), rawTxn, auth, nil, nil)
	assert.ErrorIs(t, err, ErrSponsorRejected)
}

func TestGasStation_SponsorTransaction_InsufficientFunds(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SponsorResponse{
			ErrorCode:    "INSUFFICIENT_FUNDS",
			ErrorMessage: "not enough balance",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	gs := NewGasStation(server.URL, "test-key")
	rawTxn := makeTestRawTransaction()

	privKey, _ := crypto.GenerateEd25519PrivateKey()
	auth := &aptos.AccountAuthenticator{}
	_ = auth.FromKeyAndSignature(privKey.PubKey(), &crypto.Ed25519Signature{})

	_, _, err := gs.SponsorTransaction(context.Background(), rawTxn, auth, nil, nil)
	assert.ErrorIs(t, err, ErrInsufficientFunds)
}

func TestGasStation_SponsorTransaction_MalformedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	gs := NewGasStation(server.URL, "test-key")
	rawTxn := makeTestRawTransaction()

	privKey, _ := crypto.GenerateEd25519PrivateKey()
	auth := &aptos.AccountAuthenticator{}
	_ = auth.FromKeyAndSignature(privKey.PubKey(), &crypto.Ed25519Signature{})

	_, _, err := gs.SponsorTransaction(context.Background(), rawTxn, auth, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestFeePayerRawTransaction_Sign(t *testing.T) {
	t.Parallel()

	rawTxn := makeTestRawTransaction()
	feePayer := aptos.AccountAddress{}
	feePayer[31] = 0x01

	fpTxn := NewFeePayerRawTransaction(rawTxn, feePayer)

	privKey1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	privKey2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	signingMsg, err := fpTxn.SigningMessage()
	require.NoError(t, err)

	senderAuth, err := privKey1.Sign(signingMsg)
	require.NoError(t, err)
	feePayerAuth, err := privKey2.Sign(signingMsg)
	require.NoError(t, err)

	signed, err := fpTxn.Sign(senderAuth, nil, feePayerAuth)
	require.NoError(t, err)
	assert.NotNil(t, signed)
	assert.NotNil(t, signed.Transaction)
	assert.NotNil(t, signed.Authenticator)
}

func TestBuildFeePayerTransaction(t *testing.T) {
	t.Parallel()

	sender := aptos.AccountAddress{}
	sender[31] = 0x01
	feePayer := aptos.AccountAddress{}
	feePayer[31] = 0x02

	fakeClient := testutil.NewFakeClient().
		WithAccount(sender, &aptos.AccountInfo{SequenceNumber: 5})

	payload := &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
		Function: "transfer",
	}

	fpTxn, err := BuildFeePayerTransaction(context.Background(), fakeClient, sender, feePayer, payload)
	require.NoError(t, err)
	assert.NotNil(t, fpTxn)
	assert.Equal(t, sender, fpTxn.RawTxn.Sender)
	assert.Equal(t, feePayer, fpTxn.FeePayer)
	assert.Equal(t, uint64(5), fpTxn.RawTxn.SequenceNumber)
}

func TestSignAndSubmitSponsoredTransaction(t *testing.T) {
	t.Parallel()

	senderAcct, err := account.NewEd25519()
	require.NoError(t, err)
	sponsorAcct, err := account.NewEd25519()
	require.NoError(t, err)

	fakeClient := testutil.NewFakeClient().
		WithAccount(senderAcct.Address(), &aptos.AccountInfo{SequenceNumber: 0})

	payload := &aptos.EntryFunctionPayload{
		Module:   aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
		Function: "transfer",
	}

	result, err := SignAndSubmitSponsoredTransaction(
		context.Background(),
		fakeClient,
		senderAcct,
		sponsorAcct,
		payload,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Hash)
}
