package cipherparse

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseCipherChunks parses get_available_balance / get_pending_balance view JSON into C,D chunks (32-byte compressed Ristretto).
func ParseCipherChunks(viewTop []any) (c [][]byte, d [][]byte, err error) {
	if len(viewTop) == 0 {
		return nil, nil, errors.New("empty view")
	}
	root, err := findPRRoot(viewTop[0])
	if err != nil {
		return nil, nil, err
	}
	pArr, ok := root["P"].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("P not an array, got %T", root["P"])
	}
	rArr, ok := root["R"].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("R not an array, got %T", root["R"])
	}
	if len(pArr) != len(rArr) {
		return nil, nil, fmt.Errorf("P len %d != R len %d", len(pArr), len(rArr))
	}
	for i := range pArr {
		hs, err := hexFromPointJSON(pArr[i])
		if err != nil {
			return nil, nil, fmt.Errorf("P[%d]: %w", i, err)
		}
		ci, err := decodeHex32(hs)
		if err != nil {
			return nil, nil, fmt.Errorf("P[%d]: %w", i, err)
		}
		hr, err := hexFromPointJSON(rArr[i])
		if err != nil {
			return nil, nil, fmt.Errorf("R[%d]: %w", i, err)
		}
		di, err := decodeHex32(hr)
		if err != nil {
			return nil, nil, fmt.Errorf("R[%d]: %w", i, err)
		}
		c = append(c, ci)
		d = append(d, di)
	}
	return c, d, nil
}

func decodeHex32(s string) ([]byte, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	return b, nil
}

func hexFromPointJSON(el any) (string, error) {
	switch v := el.(type) {
	case string:
		return v, nil
	case map[string]any:
		if s, ok := v["data"].(string); ok {
			return s, nil
		}
	}
	return "", fmt.Errorf("unexpected point JSON type %T", el)
}

func findPRRoot(v any) (map[string]any, error) {
	switch t := v.(type) {
	case []any:
		for _, el := range t {
			if m, err := findPRRoot(el); err == nil {
				return m, nil
			}
		}
	case map[string]any:
		if _, ok := t["P"]; ok {
			if _, ok2 := t["R"]; ok2 {
				return t, nil
			}
		}
		for _, el := range t {
			if m, err := findPRRoot(el); err == nil {
				return m, nil
			}
		}
	}
	return nil, errors.New("no object with P and R keys")
}

// ViewRowToSlice wraps a single view return row to []any for ParseCipherChunks.
func ViewRowToSlice(out []any) ([]any, error) {
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(raw, &obj); err2 != nil {
			return nil, err2
		}
		arr = []any{obj}
	}
	return arr, nil
}
