package aptos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeClient_GetTableItem(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/tables/0xhandle/item", r.URL.Path)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "address", body["key_type"])
		assert.Equal(t, "u64", body["value_type"])
		assert.Equal(t, "0x1", body["key"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode("42")
	}))

	var out string
	err := client.GetTableItem(context.Background(), "0xhandle", TableItemRequest{
		KeyType:   "address",
		ValueType: "u64",
		Key:       "0x1",
	}, &out)
	require.NoError(t, err)
	assert.Equal(t, "42", out)
}

func TestNodeClient_GetTableItem_LedgerVersion(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ledger_version=99", r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode("0")
	}))

	var out string
	err := client.GetTableItem(context.Background(), "0xhandle", TableItemRequest{
		KeyType:   "u64",
		ValueType: "u64",
		Key:       "1",
	}, &out, AtResourceLedgerVersion(99))
	require.NoError(t, err)
}

func TestNodeClient_AccountBalanceOf(t *testing.T) {
	t.Parallel()
	const fa = "0xa"
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/accounts/"+AccountOne.String()+"/balance/"+fa, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		// The node serializes u64 balances as JSON strings.
		_, _ = w.Write([]byte(`"123456789"`))
	}))

	bal, err := client.AccountBalanceOf(context.Background(), AccountOne, fa)
	require.NoError(t, err)
	assert.Equal(t, uint64(123456789), bal)
}

func TestNodeClient_AccountBalanceOf_CoinType(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A fully-qualified coin struct id must survive path-escaping intact.
		assert.Contains(t, r.URL.Path, "/balance/0x1::aptos_coin::AptosCoin")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"7"`))
	}))

	bal, err := client.AccountBalanceOf(context.Background(), AccountOne, "0x1::aptos_coin::AptosCoin")
	require.NoError(t, err)
	assert.Equal(t, uint64(7), bal)
}

func TestNodeClient_AccountModules_SinglePage(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/accounts/"+AccountOne.String()+"/modules", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]ModuleBytecode{
			{Bytecode: "0x01", ABI: &ModuleABI{Name: "a"}},
			{Bytecode: "0x02", ABI: &ModuleABI{Name: "b"}},
		})
	}))

	modules, err := client.AccountModules(context.Background(), AccountOne)
	require.NoError(t, err)
	require.Len(t, modules, 2)
	assert.Equal(t, "a", modules[0].ABI.Name)
	assert.Equal(t, "b", modules[1].ABI.Name)
}

func TestNodeClient_AccountModules_Paginates(t *testing.T) {
	t.Parallel()
	var calls int
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("start") {
		case "": // first page: return a module and a cursor pointing at the next
			w.Header().Set("X-Aptos-Cursor", "cursor1")
			_ = json.NewEncoder(w).Encode([]ModuleBytecode{{Bytecode: "0x01", ABI: &ModuleABI{Name: "a"}}})
		case "cursor1": // last page: no cursor header, iteration stops
			_ = json.NewEncoder(w).Encode([]ModuleBytecode{{Bytecode: "0x02", ABI: &ModuleABI{Name: "b"}}})
		default:
			t.Fatalf("unexpected start cursor %q", r.URL.Query().Get("start"))
		}
	}))

	modules, err := client.AccountModules(context.Background(), AccountOne)
	require.NoError(t, err)
	require.Len(t, modules, 2)
	assert.Equal(t, "a", modules[0].ABI.Name)
	assert.Equal(t, "b", modules[1].ABI.Name)
	assert.Equal(t, 2, calls, "should follow the cursor for exactly one extra page")
}
