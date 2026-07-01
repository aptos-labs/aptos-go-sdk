package confidentialasset

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk/v2"
)

// EffectiveAuditorHint mirrors TS getEffectiveAuditorHint return shape.
type EffectiveAuditorHint struct {
	IsGlobal bool
	Epoch    uint64
}

// viewToJSONArray normalizes client.View output to a top-level []any (TS client returns nested shapes).
func viewToJSONArray(out []any) ([]any, error) {
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(raw, &obj); err2 != nil {
			return nil, fmt.Errorf("view JSON: %w", err2)
		}
		arr = []any{obj}
	}
	return arr, nil
}

// HasUserRegistered wraps has_confidential_store (TS hasUserRegistered).
func (c *Client) HasUserRegistered(ctx context.Context, account, token aptos.AccountAddress) (bool, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "has_confidential_store",
		TypeArgs: nil,
		Args:     []any{account.String(), token.String()},
	})
	if err != nil {
		return false, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return false, fmt.Errorf("has_confidential_store: empty")
	}
	return viewBool(arr[0])
}

// IsBalanceNormalized wraps is_normalized.
func (c *Client) IsBalanceNormalized(ctx context.Context, account, token aptos.AccountAddress) (bool, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "is_normalized",
		TypeArgs: nil,
		Args:     []any{account.String(), token.String()},
	})
	if err != nil {
		return false, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return false, fmt.Errorf("is_normalized: empty")
	}
	return viewBool(arr[0])
}

// IncomingTransfersPaused wraps incoming_transfers_paused.
func (c *Client) IncomingTransfersPaused(ctx context.Context, account, token aptos.AccountAddress) (bool, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "incoming_transfers_paused",
		TypeArgs: nil,
		Args:     []any{account.String(), token.String()},
	})
	if err != nil {
		return false, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return false, fmt.Errorf("incoming_transfers_paused: empty")
	}
	return viewBool(arr[0])
}

// IsEmergencyPaused wraps is_emergency_paused.
func (c *Client) IsEmergencyPaused(ctx context.Context) (bool, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "is_emergency_paused",
		TypeArgs: nil,
		Args:     []any{},
	})
	if err != nil {
		return false, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return false, fmt.Errorf("is_emergency_paused: empty")
	}
	return viewBool(arr[0])
}

// GetEncryptionKeyHex returns get_encryption_key compressed point hex (0x + 64 hex); parse to bytes for crypto.
func (c *Client) GetEncryptionKeyHex(ctx context.Context, account, token aptos.AccountAddress) (string, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "get_encryption_key",
		TypeArgs: nil,
		Args:     []any{account.String(), token.String()},
	})
	if err != nil {
		return "", err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return "", fmt.Errorf("get_encryption_key: empty")
	}
	m, ok := arr[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("get_encryption_key: want object, got %T", arr[0])
	}
	data, _ := m["data"].(string)
	if data == "" {
		return "", fmt.Errorf("get_encryption_key: missing data")
	}
	return data, nil
}

// GetEffectiveAuditorHint wraps get_effective_auditor_hint.
func (c *Client) GetEffectiveAuditorHint(ctx context.Context, account, token aptos.AccountAddress) (*EffectiveAuditorHint, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "get_effective_auditor_hint",
		TypeArgs: nil,
		Args:     []any{account.String(), token.String()},
	})
	if err != nil {
		return nil, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, nil
	}
	root, ok := arr[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("get_effective_auditor_hint: unexpected response type %T", arr[0])
	}
	vec, _ := root["vec"].([]any)
	if len(vec) == 0 {
		return nil, nil
	}
	v0, _ := vec[0].(map[string]any)
	isGlobal, _ := v0["is_global"].(bool)
	epochStr, _ := v0["epoch"].(string)
	var epoch uint64
	if epochStr != "" {
		e, ok := new(big.Int).SetString(epochStr, 10)
		if ok && e.Sign() >= 0 && e.IsUint64() {
			epoch = e.Uint64()
		}
	}
	return &EffectiveAuditorHint{IsGlobal: isGlobal, Epoch: epoch}, nil
}

// GetEffectiveAuditorEncryptionKeyHex returns compressed EK hex or "" if none (TS getAssetAuditorEncryptionKey).
func (c *Client) GetEffectiveAuditorEncryptionKeyHex(ctx context.Context, token aptos.AccountAddress) (string, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "get_effective_auditor_config",
		TypeArgs: nil,
		Args:     []any{token.String()},
	})
	if err != nil {
		return "", err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return "", fmt.Errorf("get_effective_auditor_config: empty")
	}
	wrap, ok := arr[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("get_effective_auditor_config: want object")
	}
	cfg, _ := wrap["config"].(map[string]any)
	if cfg == nil {
		return "", fmt.Errorf("get_effective_auditor_config: missing config")
	}
	ek, _ := cfg["ek"].(map[string]any)
	if ek == nil {
		return "", nil
	}
	vec, _ := ek["vec"].([]any)
	if len(vec) == 0 {
		return "", nil
	}
	pt, _ := vec[0].(map[string]any)
	data, _ := pt["data"].(string)
	return data, nil
}

// GetMaxMemoBytes wraps get_max_memo_bytes.
func (c *Client) GetMaxMemoBytes(ctx context.Context) (uint64, error) {
	out, err := c.Aptos.View(ctx, &aptos.ViewPayload{
		Module:   c.ViewModule(),
		Function: "get_max_memo_bytes",
		TypeArgs: nil,
		Args:     []any{},
	})
	if err != nil {
		return 0, err
	}
	arr, err := viewToJSONArray(out)
	if err != nil || len(arr) == 0 {
		return 0, fmt.Errorf("get_max_memo_bytes: empty")
	}
	switch v := arr[0].(type) {
	case string:
		n := new(big.Int)
		if _, ok := n.SetString(v, 10); ok && n.Sign() >= 0 && n.IsUint64() {
			return n.Uint64(), nil
		}
	case float64:
		return uint64(v), nil
	}
	return 0, fmt.Errorf("get_max_memo_bytes: unexpected %T", arr[0])
}

// ChainID returns the connected chain id (TS getChainId for sigma domain).
func (c *Client) ChainID(ctx context.Context) (uint8, error) {
	return c.Aptos.ChainID(ctx)
}

func viewBool(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	case string:
		return x == "true", nil
	default:
		return false, fmt.Errorf("unexpected bool view type %T", v)
	}
}
