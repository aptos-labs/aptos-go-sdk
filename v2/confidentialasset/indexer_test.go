package confidentialasset

import (
	"context"
	"errors"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
)

func TestGetActivities_notImplemented(t *testing.T) {
	t.Parallel()
	cc := NewClient(testutil.NewFakeClient())
	_, err := cc.GetActivities(context.Background(), nil)
	if !errors.Is(err, ErrIndexerNotImplemented) {
		t.Fatalf("got %v", err)
	}
}
