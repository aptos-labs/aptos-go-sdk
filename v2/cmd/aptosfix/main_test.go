package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestFixImports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "basic import",
			input: `package main

import "github.com/aptos-labs/aptos-go-sdk"

func main() {}
`,
			expected: `github.com/aptos-labs/aptos-go-sdk/v2`,
		},
		{
			name: "crypto import",
			input: `package main

import "github.com/aptos-labs/aptos-go-sdk/crypto"

func main() {}
`,
			expected: `github.com/aptos-labs/aptos-go-sdk/v2/internal/crypto`,
		},
		{
			name: "multiple imports",
			input: `package main

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

func main() {}
`,
			expected: `github.com/aptos-labs/aptos-go-sdk/v2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.input, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			changed := fixImports(fset, file)
			if !changed {
				t.Error("expected imports to change")
			}

			// Check that v2 import path is now in the file
			found := false
			for _, imp := range file.Imports {
				if strings.Contains(imp.Path.Value, tt.expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected import containing %q", tt.expected)
			}
		})
	}
}

func TestApplyTextFixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty method call",
			input:    `client.Info()`,
			expected: `client.Info(ctx)`,
		},
		{
			name:     "method with args",
			input:    `client.Account(address)`,
			expected: `client.Account(ctx, address)`,
		},
		{
			name:     "already has ctx",
			input:    `client.Info(ctx)`,
			expected: `client.Info(ctx)`,
		},
		{
			name:     "already has ctx with args",
			input:    `client.Account(ctx, address)`,
			expected: `client.Account(ctx, address)`,
		},
		{
			name:     "multiple calls",
			input:    `client.Info(); client.Account(address)`,
			expected: `client.Info(ctx); client.Account(ctx, address)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := string(applyTextFixes([]byte(tt.input)))
			if !strings.Contains(got, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestMethodsNeedingContext(t *testing.T) {
	t.Parallel()
	// Verify key methods are in the list
	required := []string{
		"Info", "Account", "AccountResource", "AccountResources",
		"TransactionByHash", "TransactionByVersion",
		"SubmitTransaction", "SimulateTransaction",
		"View", "Fund",
	}

	for _, method := range required {
		if !methodsNeedingContext[method] {
			t.Errorf("method %s should be in methodsNeedingContext", method)
		}
	}
}
