package ans

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// optionSome builds the decoded JSON shape the node returns for a populated
// Move Option<T>: {"vec": [value]}.
func optionSome(value any) map[string]any {
	return map[string]any{"vec": []any{value}}
}

// optionNone builds the decoded JSON shape for an empty Move Option<T>.
func optionNone() map[string]any {
	return map[string]any{"vec": []any{}}
}

func futureUnixString() string {
	return strconv.FormatInt(time.Now().Add(365*24*time.Hour).Unix(), 10)
}

func TestClient_Resolve_Success(t *testing.T) {
	t.Parallel()
	addr := aptos.AccountFour
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionSome(addr.String())}).
		WithViewResult("get_expiration", []any{futureUnixString()})

	c := NewClient(fake)
	got, err := c.Resolve(context.Background(), "alice.apt")
	require.NoError(t, err)
	assert.Equal(t, addr, got)
}

func TestClient_Resolve_Expired(t *testing.T) {
	t.Parallel()
	past := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionSome(aptos.AccountFour.String())}).
		WithViewResult("get_expiration", []any{past})

	c := NewClient(fake)
	_, err := c.Resolve(context.Background(), "alice.apt")
	require.ErrorIs(t, err, ErrNameExpired)
}

func TestClient_Resolve_InvalidName(t *testing.T) {
	t.Parallel()
	c := NewClient(testutil.NewFakeClient())
	_, err := c.Resolve(context.Background(), "")
	require.Error(t, err)
}

func TestClient_GetNameInfo_NotFound_EmptyOption(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionNone()})

	c := NewClient(fake)
	_, err := c.GetNameInfo(context.Background(), Name{Domain: "bob"})
	require.ErrorIs(t, err, ErrNameNotFound)
}

func TestClient_GetNameInfo_NotFound_EmptyResult(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{})

	c := NewClient(fake)
	_, err := c.GetNameInfo(context.Background(), Name{Domain: "bob"})
	require.ErrorIs(t, err, ErrNameNotFound)
}

func TestClient_GetNameInfo_ViewNotFoundError(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().WithError("View", aptos.ErrNotFound)

	c := NewClient(fake)
	_, err := c.GetNameInfo(context.Background(), Name{Domain: "bob"})
	require.ErrorIs(t, err, ErrNameNotFound)
}

func TestClient_GetNameInfo_ViewGenericError(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().WithError("View", errors.New("network boom"))

	c := NewClient(fake)
	_, err := c.GetNameInfo(context.Background(), Name{Domain: "bob"})
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNameNotFound)
}

func TestClient_GetNameInfo_UnexpectedFormat(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{"not-an-option-object"})

	c := NewClient(fake)
	_, err := c.GetNameInfo(context.Background(), Name{Domain: "bob"})
	require.Error(t, err)
}

func TestClient_GetNameInfo_ExpirationFallback(t *testing.T) {
	t.Parallel()
	// get_target_addr resolves, but get_expiration returns an empty result.
	// GetNameInfo must fall back to a far-future expiration rather than fail.
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionSome(aptos.AccountFour.String())}).
		WithViewResult("get_expiration", []any{})

	c := NewClient(fake)
	info, err := c.GetNameInfo(context.Background(), Name{Domain: "alice"})
	require.NoError(t, err)
	assert.Equal(t, aptos.AccountFour, info.Target)
	assert.False(t, info.IsExpired())
}

func TestClient_GetNameInfo_InvalidTargetAddress(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionSome("not-a-hex-address")})

	c := NewClient(fake)
	_, err := c.GetNameInfo(context.Background(), Name{Domain: "alice"})
	require.Error(t, err)
}

func TestClient_GetPrimaryName_DomainOnly(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_primary_name", []any{
			optionSome("alice"),
			optionNone(),
		})

	c := NewClient(fake)
	name, err := c.GetPrimaryName(context.Background(), aptos.AccountFour)
	require.NoError(t, err)
	assert.Equal(t, "alice.apt", name)
}

func TestClient_GetPrimaryName_WithSubdomain(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_primary_name", []any{
			optionSome("alice"),
			optionSome("wallet"),
		})

	c := NewClient(fake)
	name, err := c.GetPrimaryName(context.Background(), aptos.AccountFour)
	require.NoError(t, err)
	assert.Equal(t, "wallet.alice.apt", name)
}

func TestClient_GetPrimaryName_NotFound(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_primary_name", []any{
			optionNone(),
			optionNone(),
		})

	c := NewClient(fake)
	_, err := c.GetPrimaryName(context.Background(), aptos.AccountFour)
	require.ErrorIs(t, err, ErrNameNotFound)
}

func TestClient_GetPrimaryName_ShortResult(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_primary_name", []any{optionSome("alice")})

	c := NewClient(fake)
	_, err := c.GetPrimaryName(context.Background(), aptos.AccountFour)
	require.ErrorIs(t, err, ErrNameNotFound)
}

func TestClient_GetPrimaryName_ViewError(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().WithError("View", errors.New("boom"))

	c := NewClient(fake)
	_, err := c.GetPrimaryName(context.Background(), aptos.AccountFour)
	require.Error(t, err)
}

func TestClient_IsAvailable_True(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionNone()})

	c := NewClient(fake)
	available, err := c.IsAvailable(context.Background(), "free.apt")
	require.NoError(t, err)
	assert.True(t, available)
}

func TestClient_IsAvailable_False(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().
		WithViewResult("get_target_addr", []any{optionSome(aptos.AccountFour.String())}).
		WithViewResult("get_expiration", []any{futureUnixString()})

	c := NewClient(fake)
	available, err := c.IsAvailable(context.Background(), "taken.apt")
	require.NoError(t, err)
	assert.False(t, available)
}

func TestClient_IsAvailable_InvalidName(t *testing.T) {
	t.Parallel()
	c := NewClient(testutil.NewFakeClient())
	_, err := c.IsAvailable(context.Background(), "")
	require.Error(t, err)
}

func TestClient_IsAvailable_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := testutil.NewFakeClient().WithError("View", errors.New("boom"))

	c := NewClient(fake)
	_, err := c.IsAvailable(context.Background(), "x.apt")
	require.Error(t, err)
}

func TestClient_GetNameInfo_ViewFuncRouting(t *testing.T) {
	t.Parallel()
	// Exercise WithViewFunc: route by function name dynamically.
	fake := testutil.NewFakeClient().WithViewFunc(func(_ context.Context, p *aptos.ViewPayload, _ ...aptos.ViewOption) ([]any, error) {
		switch p.Function {
		case "get_target_addr":
			return []any{optionSome(aptos.AccountFour.String())}, nil
		case "get_expiration":
			return []any{futureUnixString()}, nil
		default:
			return nil, errors.New("unexpected function: " + p.Function)
		}
	})

	c := NewClient(fake)
	info, err := c.GetNameInfo(context.Background(), Name{Domain: "alice"})
	require.NoError(t, err)
	assert.Equal(t, aptos.AccountFour, info.Target)
}
