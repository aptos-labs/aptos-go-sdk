package cipherparse

import (
	"strings"
	"testing"
)

func Test_hexFromPointJSON_mapNoData(t *testing.T) {
	_, err := hexFromPointJSON(map[string]any{"other": "value"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_hexFromPointJSON_unknownType(t *testing.T) {
	_, err := hexFromPointJSON(42)
	if err == nil {
		t.Fatal("expected error for int")
	}
}

func TestViewRowToSlice_marshalError(t *testing.T) {
	ch := make(chan int)
	_, err := ViewRowToSlice([]any{ch})
	if err == nil {
		t.Fatal("expected marshal error for channel")
	}
}

func TestParseChunks_errors(t *testing.T) {
	// Empty view
	if _, _, err := ParseCipherChunks(nil); err == nil {
		t.Fatal("expected error for nil")
	}
	// P not an array
	if _, _, err := ParseCipherChunks([]any{map[string]any{"P": "not-array", "R": []any{}}}); err == nil {
		t.Fatal("expected error P not array")
	}
	// R not an array
	if _, _, err := ParseCipherChunks([]any{map[string]any{"P": []any{}, "R": "not-array"}}); err == nil {
		t.Fatal("expected error R not array")
	}
	// P/R length mismatch
	if _, _, err := ParseCipherChunks([]any{map[string]any{
		"P": []any{"0x" + strings.Repeat("aa", 32)},
		"R": []any{},
	}}); err == nil {
		t.Fatal("expected error length mismatch")
	}
}
