package confidentialasset

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aptos-labs/aptos-go-sdk/v2"
)

const (
	minGasUnits         = uint64(2000)
	gasHeadroomNum      = uint64(13)
	gasHeadroomDen      = uint64(10)
	balanceFeeBudgetNum = uint64(90)
	balanceFeeBudgetDen = uint64(100)
)

// FetchPublicFABalanceOctas returns primary fungible store balance for metadata address (hex with 0x).
func (c *Client) FetchPublicFABalanceOctas(ctx context.Context, who aptos.AccountAddress, faMetadataHex string) (uint64, error) {
	if strings.TrimSpace(c.RESTBaseURL) == "" {
		return 0, fmt.Errorf("confidentialasset: set Client.RESTBaseURL or use WithRESTBaseURL for FA gas balance lookup")
	}
	meta := strings.TrimSpace(faMetadataHex)
	if !strings.HasPrefix(meta, "0x") || len(meta) < 3 {
		return 0, fmt.Errorf("confidentialasset: faMetadataHex must be hex with 0x prefix, got %q", faMetadataHex)
	}
	url := fmt.Sprintf("%s/accounts/%s/balance/%s", c.RESTBaseURL, who.String(), meta)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GET balance: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	s := strings.TrimSpace(string(body))
	if n, err := strconv.ParseUint(s, 10, 64); err == nil {
		return n, nil
	}
	var n uint64
	if err := json.Unmarshal(body, &n); err == nil {
		return n, nil
	}
	var js string
	if err := json.Unmarshal(body, &js); err != nil {
		return 0, fmt.Errorf("parse balance body %q: %w", s, err)
	}
	return strconv.ParseUint(js, 10, 64)
}

// SubmitWithSimulatedGas builds, simulates, caps max gas from FA balance, signs, submits (aligned with TS examples + native chain_flow).
func (c *Client) SubmitWithSimulatedGas(ctx context.Context, signer aptos.TransactionSigner, label string, payload aptos.Payload, faMetadataHex string) (*aptos.Transaction, error) {
	if c.WithFeePayer {
		return nil, fmt.Errorf("confidentialasset: fee payer submit not implemented yet")
	}
	bal, err := c.FetchPublicFABalanceOctas(ctx, signer.Address(), faMetadataHex)
	if err != nil {
		return nil, fmt.Errorf("%s: fa balance: %w", label, err)
	}
	draft, err := c.Aptos.BuildTransaction(ctx, signer.Address(), payload, aptos.WithGasEstimation())
	if err != nil {
		return nil, fmt.Errorf("%s: build draft: %w", label, err)
	}
	sim, err := c.Aptos.SimulateTransaction(ctx, draft, signer)
	if err != nil {
		return nil, fmt.Errorf("%s: simulate: %w", label, err)
	}
	if !sim.Success {
		return nil, fmt.Errorf("%s: simulation failed: %s", label, sim.VMStatus)
	}
	gasUsed := sim.GasUsed
	gasPrice := sim.GasUnitPrice
	if gasPrice == 0 {
		return nil, fmt.Errorf("%s: gas_unit_price=0", label)
	}
	maxGas := (gasUsed*gasHeadroomNum + gasHeadroomDen - 1) / gasHeadroomDen
	if maxGas < minGasUnits {
		maxGas = minGasUnits
	}
	maxByBalanceBig := new(big.Int).Mul(
		new(big.Int).SetUint64(bal),
		new(big.Int).SetUint64(balanceFeeBudgetNum),
	)
	maxByBalanceBig.Div(maxByBalanceBig, new(big.Int).Mul(
		new(big.Int).SetUint64(balanceFeeBudgetDen),
		new(big.Int).SetUint64(gasPrice),
	))
	maxByBalance := ^uint64(0) // if big.Int overflows uint64, balance is more than sufficient
	if maxByBalanceBig.IsUint64() {
		maxByBalance = maxByBalanceBig.Uint64()
	}
	if maxGas > maxByBalance {
		maxGas = maxByBalance
	}
	if maxGas < minGasUnits {
		return nil, fmt.Errorf("%s: balance too low for fees (bal=%d gas_price=%d affordable_max_gas≈%d)", label, bal, gasPrice, maxByBalance)
	}
	if maxGas < gasUsed {
		return nil, fmt.Errorf("%s: affordable max_gas %d < gas_used %d", label, maxGas, gasUsed)
	}
	final, err := c.Aptos.BuildTransaction(
		ctx, signer.Address(), payload,
		aptos.WithMaxGas(maxGas),
		aptos.WithGasPrice(gasPrice),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: build final: %w", label, err)
	}
	signed, err := aptos.SignTransaction(signer, final)
	if err != nil {
		return nil, fmt.Errorf("%s: sign: %w", label, err)
	}
	sub, err := c.Aptos.SubmitTransaction(ctx, signed)
	if err != nil {
		return nil, fmt.Errorf("%s: submit: %w", label, err)
	}
	tx, err := c.Aptos.WaitForTransaction(ctx, sub.Hash, aptos.WithPollTimeout(3*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("%s: wait: %w", label, err)
	}
	if !tx.Success {
		return nil, fmt.Errorf("%s: committed but failed: %s", label, tx.VMStatus)
	}
	return tx, nil
}

// BuildDepositPayload builds the confidential_asset::deposit entry-function payload without
// signing or submitting it, so callers can run their own build/simulate/sign/submit pipeline
// (mirrors the TS ConfidentialAssetTransactionBuilder.deposit split).
func (c *Client) BuildDepositPayload(token aptos.AccountAddress, amountOctas uint64) (*aptos.EntryFunctionPayload, error) {
	// Args must match native chain_flow + Move: address (metadata object id) + u64 octas (not BCS string).
	return &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: "deposit",
		TypeArgs: nil,
		Args:     []any{token, amountOctas},
	}, nil
}

// Deposit submits confidential_asset::deposit (TS ConfidentialAssetTransactionBuilder.deposit).
func (c *Client) Deposit(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, amountOctas uint64, faMetadataHex string) (*aptos.Transaction, error) {
	payload, err := c.BuildDepositPayload(token, amountOctas)
	if err != nil {
		return nil, err
	}
	return c.SubmitWithSimulatedGas(ctx, signer, "deposit", payload, faMetadataHex)
}

// BuildRolloverPendingBalancePayload builds the rollover_pending_balance(_and_pause) payload
// without signing or submitting it.
func (c *Client) BuildRolloverPendingBalancePayload(token aptos.AccountAddress, withPauseIncoming bool) (*aptos.EntryFunctionPayload, error) {
	fn := "rollover_pending_balance"
	if withPauseIncoming {
		fn = "rollover_pending_balance_and_pause"
	}
	return &aptos.EntryFunctionPayload{
		Module:   c.ViewModule(),
		Function: fn,
		TypeArgs: nil,
		Args:     []any{token},
	}, nil
}

// RolloverPendingBalance submits rollover_pending_balance or rollover_pending_balance_and_pause.
func (c *Client) RolloverPendingBalance(ctx context.Context, signer aptos.TransactionSigner, token aptos.AccountAddress, withPauseIncoming bool, faMetadataHex string) (*aptos.Transaction, error) {
	payload, err := c.BuildRolloverPendingBalancePayload(token, withPauseIncoming)
	if err != nil {
		return nil, err
	}
	fn := "rollover_pending_balance"
	if withPauseIncoming {
		fn = "rollover_pending_balance_and_pause"
	}
	return c.SubmitWithSimulatedGas(ctx, signer, fn, payload, faMetadataHex)
}
