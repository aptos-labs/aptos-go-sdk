package confidentialasset

import (
	"context"
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
			if err != nil || !b {
				return err
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
			if err != nil || h == "" {
				return err
			}
			return nil
		}},
		{"GetEffectiveAuditorHint", func() error {
			h, err := cc.GetEffectiveAuditorHint(ctx, acct, token)
			if err != nil || h == nil || !h.IsGlobal || h.Epoch != 42 {
				return err
			}
			return nil
		}},
		{"GetEffectiveAuditorEncryptionKeyHex", func() error {
			h, err := cc.GetEffectiveAuditorEncryptionKeyHex(ctx, token)
			if err != nil || h != "" {
				return err
			}
			return nil
		}},
		{"GetMaxMemoBytes", func() error {
			n, err := cc.GetMaxMemoBytes(ctx)
			if err != nil || n != 256 {
				return err
			}
			return nil
		}},
		{"ChainID", func() error {
			ch, err := cc.ChainID(ctx)
			if err != nil || ch != 4 {
				return err
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
