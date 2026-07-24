package confidentialasset

import (
	"context"
	"fmt"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2"
)

func TestClient_views(t *testing.T) {
	t.Parallel()
	cc, _ := newTestConfidentialClient()
	ctx := context.Background()
	token := aptos.AccountOne
	acct := aptos.AccountOne

	for _, fn := range []struct {
		name string
		run  func() error
	}{
		{"HasUserRegistered", func() error {
			b, err := cc.HasUserRegistered(ctx, acct, token)
			if err != nil {
				return err
			}
			if !b {
				return fmt.Errorf("HasUserRegistered: want true, got false")
			}
			return nil
		}},
		{"IsBalanceNormalized", func() error {
			_, err := cc.IsBalanceNormalized(ctx, acct, token)
			return err
		}},
		{"IncomingTransfersPaused", func() error {
			_, err := cc.IncomingTransfersPaused(ctx, acct, token)
			return err
		}},
		{"IsEmergencyPaused", func() error {
			_, err := cc.IsEmergencyPaused(ctx)
			return err
		}},
		{"GetEncryptionKeyHex", func() error {
			h, err := cc.GetEncryptionKeyHex(ctx, acct, token)
			if err != nil {
				return err
			}
			if h == "" {
				return fmt.Errorf("GetEncryptionKeyHex: want non-empty hex, got empty")
			}
			return nil
		}},
		{"GetEffectiveAuditorHint", func() error {
			h, err := cc.GetEffectiveAuditorHint(ctx, acct, token)
			if err != nil {
				return err
			}
			if h == nil || !h.IsGlobal || h.Epoch != 42 {
				return fmt.Errorf("GetEffectiveAuditorHint: want {IsGlobal:true Epoch:42}, got %+v", h)
			}
			return nil
		}},
		{"GetEffectiveAuditorEncryptionKeyHex", func() error {
			h, err := cc.GetEffectiveAuditorEncryptionKeyHex(ctx, token)
			if err != nil {
				return err
			}
			if h != "" {
				return fmt.Errorf("GetEffectiveAuditorEncryptionKeyHex: want empty, got %q", h)
			}
			return nil
		}},
		{"GetMaxMemoBytes", func() error {
			n, err := cc.GetMaxMemoBytes(ctx)
			if err != nil {
				return err
			}
			if n != 256 {
				return fmt.Errorf("GetMaxMemoBytes: want 256, got %d", n)
			}
			return nil
		}},
		{"ChainID", func() error {
			ch, err := cc.ChainID(ctx)
			if err != nil {
				return err
			}
			if ch != 4 {
				return fmt.Errorf("ChainID: want 4, got %d", ch)
			}
			return nil
		}},
	} {
		fn := fn
		t.Run(fn.name, func(t *testing.T) {
			t.Parallel()
			if err := fn.run(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGetEffectiveAuditorEncryptionKeyHex_withEK(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(ctx context.Context, payload *aptos.ViewPayload, opts ...aptos.ViewOption) ([]any, error) {
		if payload.Function == "get_effective_auditor_config" {
			return []any{map[string]any{
				"config": map[string]any{
					"ek": map[string]any{
						"vec": []any{map[string]any{"data": "0x" + testPointP}},
					},
				},
			}}, nil
		}
		return testViewFunc(ctx, payload, opts...)
	})
	h, err := cc.GetEffectiveAuditorEncryptionKeyHex(context.Background(), aptos.AccountOne)
	if err != nil || h == "" {
		t.Fatalf("h=%q err=%v", h, err)
	}
}

func TestGetEffectiveAuditorEncryptionKeyHex_missingConfig(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(ctx context.Context, payload *aptos.ViewPayload, opts ...aptos.ViewOption) ([]any, error) {
		if payload.Function == "get_effective_auditor_config" {
			return []any{map[string]any{}}, nil
		}
		return testViewFunc(ctx, payload, opts...)
	})
	_, err := cc.GetEffectiveAuditorEncryptionKeyHex(context.Background(), aptos.AccountOne)
	if err == nil {
		t.Fatal("expected missing config error")
	}
}

func TestClient_views_viewError(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(_ context.Context, _ *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		return nil, fmt.Errorf("injected view error")
	})
	ctx := context.Background()
	acct := aptos.AccountOne
	token := aptos.AccountOne

	if _, err := cc.HasUserRegistered(ctx, acct, token); err == nil {
		t.Fatal("HasUserRegistered: expected error")
	}
	if _, err := cc.IsBalanceNormalized(ctx, acct, token); err == nil {
		t.Fatal("IsBalanceNormalized: expected error")
	}
	if _, err := cc.IncomingTransfersPaused(ctx, acct, token); err == nil {
		t.Fatal("IncomingTransfersPaused: expected error")
	}
	if _, err := cc.IsEmergencyPaused(ctx); err == nil {
		t.Fatal("IsEmergencyPaused: expected error")
	}
	if _, err := cc.GetEncryptionKeyHex(ctx, acct, token); err == nil {
		t.Fatal("GetEncryptionKeyHex: expected error")
	}
	if _, err := cc.GetMaxMemoBytes(ctx); err == nil {
		t.Fatal("GetMaxMemoBytes: expected error")
	}
	if _, err := cc.GetEffectiveAuditorHint(ctx, acct, token); err == nil {
		t.Fatal("GetEffectiveAuditorHint: expected error")
	}
	if _, err := cc.GetEffectiveAuditorEncryptionKeyHex(ctx, token); err == nil {
		t.Fatal("GetEffectiveAuditorEncryptionKeyHex: expected error")
	}
}

func TestClient_views_emptyResult(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(_ context.Context, _ *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		return []any{}, nil
	})
	ctx := context.Background()
	acct := aptos.AccountOne
	token := aptos.AccountOne

	if _, err := cc.HasUserRegistered(ctx, acct, token); err == nil {
		t.Fatal("HasUserRegistered: expected empty error")
	}
	if _, err := cc.IsBalanceNormalized(ctx, acct, token); err == nil {
		t.Fatal("IsBalanceNormalized: expected empty error")
	}
	if _, err := cc.IncomingTransfersPaused(ctx, acct, token); err == nil {
		t.Fatal("IncomingTransfersPaused: expected empty error")
	}
	if _, err := cc.IsEmergencyPaused(ctx); err == nil {
		t.Fatal("IsEmergencyPaused: expected empty error")
	}
	if _, err := cc.GetEncryptionKeyHex(ctx, acct, token); err == nil {
		t.Fatal("GetEncryptionKeyHex: expected empty error")
	}
	if _, err := cc.GetMaxMemoBytes(ctx); err == nil {
		t.Fatal("GetMaxMemoBytes: expected empty error")
	}
}

func TestGetEncryptionKeyHex_missingData(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(_ context.Context, payload *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		if payload.Function == "get_encryption_key" {
			return []any{map[string]any{"other": "value"}}, nil
		}
		return testViewFunc(context.Background(), payload)
	})
	_, err := cc.GetEncryptionKeyHex(context.Background(), aptos.AccountOne, aptos.AccountOne)
	if err == nil {
		t.Fatal("expected error for missing data field")
	}
}

func TestGetEncryptionKeyHex_wrongType(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(_ context.Context, payload *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		if payload.Function == "get_encryption_key" {
			return []any{"not-a-map"}, nil
		}
		return testViewFunc(context.Background(), payload)
	})
	_, err := cc.GetEncryptionKeyHex(context.Background(), aptos.AccountOne, aptos.AccountOne)
	if err == nil {
		t.Fatal("expected error for non-map result")
	}
}

func TestGetMaxMemoBytes_float64(t *testing.T) {
	t.Parallel()
	cc, fc := newTestConfidentialClient()
	fc.WithViewFunc(func(_ context.Context, payload *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		if payload.Function == "get_max_memo_bytes" {
			return []any{float64(128)}, nil
		}
		return testViewFunc(context.Background(), payload)
	})
	n, err := cc.GetMaxMemoBytes(context.Background())
	if err != nil || n != 128 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}
