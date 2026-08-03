package cipherparse_test

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/cipherparse"
)

func TestParseCipherChunks_singleChunk(t *testing.T) {
	const p = "e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76"
	const r = "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"
	raw := []byte(`[{"P":[{"data":"0x` + p + `"}],"R":[{"data":"0x` + r + `"}]}]`)
	var top []any
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	c, d, err := cipherparse.ParseCipherChunks(top)
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 1 || len(d) != 1 {
		t.Fatalf("chunks: %d %d", len(c), len(d))
	}
}

func TestViewRowToSlice(t *testing.T) {
	t.Parallel()
	const p = "e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76"
	const r = "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"
	row := map[string]any{"P": []any{map[string]any{"data": "0x" + p}}, "R": []any{map[string]any{"data": "0x" + r}}}
	arr, err := cipherparse.ViewRowToSlice([]any{row})
	if err != nil {
		t.Fatal(err)
	}
	c, d, err := cipherparse.ParseCipherChunks(arr)
	if err != nil || len(c) != 1 {
		t.Fatalf("c=%d err=%v", len(c), err)
	}
	_ = d
}

func TestParseCipherChunks_badInput(t *testing.T) {
	t.Parallel()
	_, _, err := cipherparse.ParseCipherChunks([]any{map[string]any{"P": "not-a-point"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCipherChunks_nestedPR(t *testing.T) {
	t.Parallel()
	const p = "e2f2ae0a6abc4e71a884a961c500515f58e30b6aa582dd8db6a65945e08d2d76"
	const r = "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"
	wrapped := []any{[]any{map[string]any{
		"P": []any{map[string]any{"data": "0x" + p}},
		"R": []any{map[string]any{"data": "0x" + r}},
	}}}
	c, d, err := cipherparse.ParseCipherChunks(wrapped)
	if err != nil || len(c) != 1 {
		t.Fatalf("err=%v len=%d", err, len(c))
	}
	_ = d
}

func TestParseCipherChunks_badHex(t *testing.T) {
	t.Parallel()
	top := []any{map[string]any{
		"P": []any{map[string]any{"data": "0xzz"}},
		"R": []any{map[string]any{"data": "0x00"}},
	}}
	_, _, err := cipherparse.ParseCipherChunks(top)
	if err == nil {
		t.Fatal("expected hex error")
	}
}
