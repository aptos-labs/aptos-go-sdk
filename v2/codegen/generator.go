package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/aptos-labs/aptos-go-sdk/api"
)

// Options configures code generation behavior.
type Options struct {
	// PackageName is the Go package name for the generated code.
	// If empty, defaults to the module name.
	PackageName string

	// ModuleAddress overrides the module address in the generated code.
	// If empty, uses the address from the ABI.
	ModuleAddress string

	// GenerateStructs controls whether to generate Go structs from Move structs.
	// Default: true
	GenerateStructs bool

	// GenerateEntryFunctions controls whether to generate entry function wrappers.
	// Default: true
	GenerateEntryFunctions bool

	// GenerateViewFunctions controls whether to generate view function helpers.
	// Default: true
	GenerateViewFunctions bool

	// IncludePrivate includes private functions in generation (for testing).
	// Default: false
	IncludePrivate bool
}

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		GenerateStructs:        true,
		GenerateEntryFunctions: true,
		GenerateViewFunctions:  true,
		IncludePrivate:         false,
	}
}

// GeneratedCode represents the output of code generation.
type GeneratedCode struct {
	// PackageName is the Go package name.
	PackageName string

	// ModuleName is the Move module name.
	ModuleName string

	// Code is the formatted Go source code.
	Code []byte
}

// GenerateModule generates Go code from a Move module ABI.
func GenerateModule(module *api.MoveModule, opts Options) (*GeneratedCode, error) {
	if module == nil {
		return nil, errors.New("module ABI is nil")
	}

	// Apply defaults
	if opts.PackageName == "" {
		opts.PackageName = sanitizePackageName(module.Name)
	}
	if !opts.GenerateStructs && !opts.GenerateEntryFunctions && !opts.GenerateViewFunctions {
		opts = DefaultOptions()
		opts.PackageName = sanitizePackageName(module.Name)
	}

	// Build template data
	data := templateData{
		PackageName:   opts.PackageName,
		ModuleName:    module.Name,
		ModuleAddress: formatModuleAddress(module, opts),
		Structs:       make([]structData, 0),
		EntryFuncs:    make([]funcData, 0),
		ViewFuncs:     make([]funcData, 0),
	}

	// Generate structs
	if opts.GenerateStructs {
		for _, s := range module.Structs {
			if s.IsNative {
				continue // Skip native types
			}
			sd := convertStruct(s)
			data.Structs = append(data.Structs, sd)
		}
	}

	// Generate functions
	for _, f := range module.ExposedFunctions {
		if f.Visibility == api.MoveVisibilityPrivate && !opts.IncludePrivate {
			continue
		}

		fd := convertFunction(f, module.Name)

		if f.IsEntry && opts.GenerateEntryFunctions {
			data.EntryFuncs = append(data.EntryFuncs, fd)
		}
		if f.IsView && opts.GenerateViewFunctions {
			data.ViewFuncs = append(data.ViewFuncs, fd)
		}
	}

	// Determine required imports
	data.Imports = determineImports(data)

	// Execute template
	var buf bytes.Buffer
	tmpl, err := template.New("module").Funcs(templateFuncs).Parse(moduleTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code for debugging
		return &GeneratedCode{
			PackageName: opts.PackageName,
			ModuleName:  module.Name,
			Code:        buf.Bytes(),
		}, fmt.Errorf("formatting code: %w (unformatted code returned)", err)
	}

	return &GeneratedCode{
		PackageName: opts.PackageName,
		ModuleName:  module.Name,
		Code:        formatted,
	}, nil
}

// templateData holds all data needed for code generation.
type templateData struct {
	PackageName   string
	ModuleName    string
	ModuleAddress string
	Imports       []string
	Structs       []structData
	EntryFuncs    []funcData
	ViewFuncs     []funcData
}

type structData struct {
	Name          string
	GoName        string
	Fields        []fieldData
	GenericParams []string
	Abilities     []string
	Comment       string
}

type fieldData struct {
	Name    string
	GoName  string
	Type    string
	GoType  string
	Comment string
}

type funcData struct {
	Name           string
	GoName         string
	Params         []paramData
	Returns        []returnData
	GenericParams  []string
	IsEntry        bool
	IsView         bool
	Comment        string
	MoveSignature  string
	HasSigner      bool
	NonSignerCount int
}

type paramData struct {
	Name     string
	GoName   string
	Type     string
	GoType   string
	IsSigner bool
}

type returnData struct {
	Type   string
	GoType string
}

// sanitizePackageName converts a Move module name to a valid Go package name.
func sanitizePackageName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace invalid characters with underscores
	var result strings.Builder
	foundLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			foundLetter = true
			result.WriteRune(r)
		} else if foundLetter && (unicode.IsDigit(r) || r == '_') {
			result.WriteRune(r)
		} else if foundLetter {
			result.WriteRune('_')
		}
		// Skip leading non-letter characters
	}
	return result.String()
}

// formatModuleAddress formats the module address for code generation.
func formatModuleAddress(module *api.MoveModule, opts Options) string {
	if opts.ModuleAddress != "" {
		return opts.ModuleAddress
	}
	if module.Address != nil {
		return module.Address.String()
	}
	return "0x1"
}

// convertStruct converts a Move struct to Go struct data.
func convertStruct(s *api.MoveStruct) structData {
	sd := structData{
		Name:      s.Name,
		GoName:    toGoPublicName(s.Name),
		Abilities: make([]string, len(s.Abilities)),
		Comment:   fmt.Sprintf("%s represents the Move struct %s", toGoPublicName(s.Name), s.Name),
	}

	for i, a := range s.Abilities {
		sd.Abilities[i] = string(a)
	}

	for _, g := range s.GenericTypeParams {
		_ = g // Generic params are tracked but not fully implemented yet
		sd.GenericParams = append(sd.GenericParams, "T")
	}

	for _, f := range s.Fields {
		fd := fieldData{
			Name:   f.Name,
			GoName: toGoPublicName(f.Name),
			Type:   f.Type,
			GoType: moveTypeToGoType(f.Type),
		}
		sd.Fields = append(sd.Fields, fd)
	}

	return sd
}

// convertFunction converts a Move function to Go function data.
func convertFunction(f *api.MoveFunction, moduleName string) funcData {
	fd := funcData{
		Name:    f.Name,
		GoName:  toGoPublicName(f.Name),
		IsEntry: f.IsEntry,
		IsView:  f.IsView,
	}

	// Build Move signature for comment
	fd.MoveSignature = fmt.Sprintf("%s::%s(%s)", moduleName, f.Name, strings.Join(f.Params, ", "))

	// Convert generic type params
	for range f.GenericTypeParams {
		fd.GenericParams = append(fd.GenericParams, "T")
	}

	// Convert parameters
	paramIdx := 0
	for _, p := range f.Params {
		isSigner := p == "signer" || p == "&signer"
		if isSigner {
			fd.HasSigner = true
			continue // Signer is handled separately
		}

		pd := paramData{
			Name:     fmt.Sprintf("arg%d", paramIdx),
			GoName:   fmt.Sprintf("arg%d", paramIdx),
			Type:     p,
			GoType:   moveTypeToGoType(p),
			IsSigner: false,
		}
		fd.Params = append(fd.Params, pd)
		paramIdx++
	}
	fd.NonSignerCount = len(fd.Params)

	// Convert return types
	for _, r := range f.Return {
		rd := returnData{
			Type:   r,
			GoType: moveTypeToGoType(r),
		}
		fd.Returns = append(fd.Returns, rd)
	}

	// Generate comment
	if f.IsEntry {
		fd.Comment = fmt.Sprintf("%s creates an entry function payload for %s", fd.GoName, fd.MoveSignature)
	} else if f.IsView {
		fd.Comment = fmt.Sprintf("%s calls the view function %s", fd.GoName, fd.MoveSignature)
	}

	return fd
}

// moveTypeToGoType converts a Move type string to a Go type.
func moveTypeToGoType(moveType string) string {
	moveType = strings.TrimSpace(moveType)

	// Handle reference types
	if strings.HasPrefix(moveType, "&") {
		inner := strings.TrimPrefix(moveType, "&")
		inner = strings.TrimPrefix(inner, "mut ")
		return moveTypeToGoType(inner)
	}

	// Handle vector types
	if strings.HasPrefix(moveType, "vector<") && strings.HasSuffix(moveType, ">") {
		inner := moveType[7 : len(moveType)-1]
		if inner == "u8" {
			return "[]byte"
		}
		return "[]" + moveTypeToGoType(inner)
	}

	// Handle primitive types
	switch moveType {
	case "bool":
		return "bool"
	case "u8":
		return "uint8"
	case "u16":
		return "uint16"
	case "u32":
		return "uint32"
	case "u64":
		return "uint64"
	case "u128":
		return "*big.Int"
	case "u256":
		return "*big.Int"
	case "address", "signer", "&signer":
		return "aptos.AccountAddress"
	}

	// Handle generic type parameters (T0, T1, etc.)
	if matched, _ := regexp.MatchString(`^T\d+$`, moveType); matched {
		return "any"
	}

	// Handle struct types (0x1::module::Struct or 0x1::module::Struct<...>)
	// Need to parse carefully to handle nested generics
	if strings.Contains(moveType, "::") {
		// Find the struct address::module::name portion (everything before first <)
		structPath := moveType
		genericPart := ""
		if bracketIdx := strings.Index(moveType, "<"); bracketIdx != -1 {
			structPath = moveType[:bracketIdx]
			genericPart = moveType[bracketIdx:]
		}

		// Split the path
		parts := strings.Split(structPath, "::")
		if len(parts) >= 3 {
			moduleName := parts[1]
			structName := parts[len(parts)-1]

			// Special cases for common types
			switch {
			case moduleName == "string" && structName == "String":
				return "string"
			case moduleName == "option" && structName == "Option":
				// Extract the inner type from Option<...>
				if genericPart != "" && strings.HasPrefix(genericPart, "<") && strings.HasSuffix(genericPart, ">") {
					innerType := genericPart[1 : len(genericPart)-1]
					return "*" + moveTypeToGoType(innerType)
				}
				return "*any"
			case moduleName == "object" && structName == "Object":
				// Objects are always addresses
				return "aptos.AccountAddress"
			case moduleName == "table" && structName == "Table":
				// Tables map to a generic map type
				return "any" // Tables need special handling
			case moduleName == "aggregator" && structName == "Aggregator":
				return "*big.Int"
			}

			// For structs from other modules, use 'any' since they won't be defined
			// in the generated package. The generated code should still compile.
			return "any"
		}
	}

	// Default to any for unknown types
	return "any"
}

// toGoPublicName converts a snake_case name to PascalCase.
func toGoPublicName(name string) string {
	parts := strings.Split(name, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}
	return result.String()
}

// determineImports determines which imports are needed based on generated code.
func determineImports(data templateData) []string {
	imports := make(map[string]bool)

	// Always need aptos SDK and BCS
	imports[`"github.com/aptos-labs/aptos-go-sdk/v2"`] = true
	imports[`"github.com/aptos-labs/aptos-go-sdk/bcs"`] = true

	// Check for big.Int usage
	needsBigInt := false
	for _, s := range data.Structs {
		for _, f := range s.Fields {
			if f.GoType == "*big.Int" {
				needsBigInt = true
			}
		}
	}
	for _, f := range data.EntryFuncs {
		for _, p := range f.Params {
			if p.GoType == "*big.Int" {
				needsBigInt = true
			}
		}
	}
	for _, f := range data.ViewFuncs {
		for _, r := range f.Returns {
			if r.GoType == "*big.Int" {
				needsBigInt = true
			}
		}
	}

	if needsBigInt {
		imports[`"math/big"`] = true
	}

	// Convert to sorted slice
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	return result
}

var templateFuncs = template.FuncMap{
	"lower": strings.ToLower,
	"join":  strings.Join,
	"serializeCall": func(goType, varName string) string {
		switch goType {
		case "bool":
			return fmt.Sprintf("bcs.SerializeBool(%s)", varName)
		case "uint8":
			return fmt.Sprintf("bcs.SerializeU8(%s)", varName)
		case "uint16":
			return fmt.Sprintf("bcs.SerializeU16(%s)", varName)
		case "uint32":
			return fmt.Sprintf("bcs.SerializeU32(%s)", varName)
		case "uint64":
			return fmt.Sprintf("bcs.SerializeU64(%s)", varName)
		case "*big.Int":
			return fmt.Sprintf("bcs.SerializeU128(*%s)", varName)
		case "aptos.AccountAddress":
			return fmt.Sprintf("bcs.Serialize(&%s)", varName)
		case "[]byte":
			return fmt.Sprintf("bcs.SerializeBytes(%s)", varName)
		case "string":
			return fmt.Sprintf("bcs.SerializeBytes([]byte(%s))", varName)
		default:
			// For slices of known types
			if strings.HasPrefix(goType, "[]") {
				return fmt.Sprintf("bcs.SerializeSequenceOnly(%s)", varName)
			}
			// For any/interface types, just serialize directly (may fail at runtime)
			return fmt.Sprintf("bcs.Serialize(%s.(bcs.Marshaler))", varName)
		}
	},
}

const moduleTemplate = `// Code generated by aptosgen. DO NOT EDIT.
// Source: {{.ModuleName}} module at {{.ModuleAddress}}

package {{.PackageName}}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)

// ModuleAddress is the address where the {{.ModuleName}} module is deployed.
var ModuleAddress = mustParseAddress("{{.ModuleAddress}}")

// ModuleName is the name of the Move module.
const ModuleName = "{{.ModuleName}}"

func mustParseAddress(s string) aptos.AccountAddress {
	addr, err := aptos.ParseAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}

{{- range $s := .Structs}}

// {{$s.Comment}}
// Abilities: {{join $s.Abilities ", "}}
type {{$s.GoName}} struct {
{{- range $s.Fields}}
	{{.GoName}} {{.GoType}} ` + "`" + `json:"{{.Name}}"` + "`" + ` // {{.Type}}
{{- end}}
}
{{- end}}

{{- range $f := .EntryFuncs}}

// {{$f.Comment}}
func {{$f.GoName}}({{if $f.HasSigner}}sender aptos.TransactionSigner{{end}}{{if and $f.HasSigner $f.Params}}, {{end}}{{range $i, $p := $f.Params}}{{if $i}}, {{end}}{{$p.GoName}} {{$p.GoType}}{{end}}{{if $f.GenericParams}}{{if or $f.HasSigner $f.Params}}, {{end}}typeArgs ...aptos.TypeTag{{end}}) (*aptos.EntryFunction, error) {
	args := make([][]byte, 0, {{len $f.Params}})
	{{- range $f.Params}}
	{{.GoName}}Bytes, err := {{serializeCall .GoType .GoName}}
	if err != nil {
		return nil, err
	}
	args = append(args, {{.GoName}}Bytes)
	{{- end}}

	{{if $f.GenericParams -}}
	return &aptos.EntryFunction{
		Module: aptos.ModuleID{
			Address: ModuleAddress,
			Name:    ModuleName,
		},
		Function: "{{$f.Name}}",
		ArgTypes: typeArgs,
		Args:     args,
	}, nil
	{{- else -}}
	return &aptos.EntryFunction{
		Module: aptos.ModuleID{
			Address: ModuleAddress,
			Name:    ModuleName,
		},
		Function: "{{$f.Name}}",
		ArgTypes: []aptos.TypeTag{},
		Args:     args,
	}, nil
	{{- end}}
}
{{- end}}

{{- range $f := .ViewFuncs}}

// {{$f.Comment}}
func {{$f.GoName}}(client aptos.Client{{- range $i, $p := $f.Params}}, {{$p.GoName}} {{$p.GoType}}{{end}}{{if $f.GenericParams}}, typeArgs ...aptos.TypeTag{{end}}) ({{range $i, $r := $f.Returns}}{{if $i}}, {{end}}{{$r.GoType}}{{end}}{{if $f.Returns}}, {{end}}error) {
	args := make([][]byte, 0, {{len $f.Params}})
	{{- range $f.Params}}
	{{.GoName}}Bytes, err := {{serializeCall .GoType .GoName}}
	if err != nil {
		return {{range $f.Returns}}{{zeroValue .GoType}}, {{end}}err
	}
	args = append(args, {{.GoName}}Bytes)
	{{- end}}

	payload := &aptos.ViewPayload{
		Module: aptos.ModuleID{
			Address: ModuleAddress,
			Name:    ModuleName,
		},
		Function: "{{$f.Name}}",
		ArgTypes: {{if $f.GenericParams}}typeArgs{{else}}[]aptos.TypeTag{}{{end}},
		Args:     args,
	}

	result, err := client.View(context.Background(), payload)
	if err != nil {
		return {{range $f.Returns}}{{zeroValue .GoType}}, {{end}}err
	}

	_ = result // TODO: Parse result based on return types
	{{if eq (len $f.Returns) 0 -}}
	return nil
	{{- else if eq (len $f.Returns) 1 -}}
	return result[0].({{(index $f.Returns 0).GoType}}), nil
	{{- else -}}
	// Multiple return values need manual parsing
	return {{range $i, $r := $f.Returns}}{{if $i}}, {{end}}result[{{$i}}].({{$r.GoType}}){{end}}, nil
	{{- end}}
}
{{- end}}
`

func init() {
	templateFuncs["zeroValue"] = func(goType string) string {
		switch goType {
		case "bool":
			return "false"
		case "string":
			return `""`
		case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
			return "0"
		case "*big.Int":
			return "nil"
		case "aptos.AccountAddress":
			return "aptos.AccountAddress{}"
		case "[]byte":
			return "nil"
		default:
			if strings.HasPrefix(goType, "[]") {
				return "nil"
			}
			if strings.HasPrefix(goType, "*") {
				return "nil"
			}
			return goType + "{}"
		}
	}
}
