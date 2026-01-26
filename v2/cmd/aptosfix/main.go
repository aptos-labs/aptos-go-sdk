// Command aptosfix migrates Go code from Aptos Go SDK v1 to v2.
//
// Usage:
//
//	aptosfix [flags] [path ...]
//
// Flags:
//
//	-w        Write result to (source) file instead of stdout
//	-d        Display diffs instead of rewriting files
//	-l        List files whose formatting differs from aptosfix's
//	-v        Verbose mode: print files being processed
//	-imports  Only update import paths (skip other transformations)
//	-dry-run  Show what would be changed without modifying files
//
// Examples:
//
//	# Preview changes to a single file
//	aptosfix -d myfile.go
//
//	# Fix all files in current directory
//	aptosfix -w .
//
//	# Only update imports (safer first step)
//	aptosfix -w -imports .
//
//	# List files that need updating
//	aptosfix -l ./...
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	v1ModulePath = "github.com/aptos-labs/aptos-go-sdk"
	v2ModulePath = "github.com/aptos-labs/aptos-go-sdk/v2"
)

var (
	writeFile   = flag.Bool("w", false, "write result to (source) file instead of stdout")
	showDiff    = flag.Bool("d", false, "display diffs instead of rewriting files")
	listFiles   = flag.Bool("l", false, "list files whose formatting differs")
	verbose     = flag.Bool("v", false, "verbose: print files being processed")
	importsOnly = flag.Bool("imports", false, "only update import paths")
	dryRun      = flag.Bool("dry-run", false, "show what would be changed without modifying")
)

// importMapping maps v1 import paths to v2 import paths.
var importMapping = map[string]string{
	v1ModulePath:                     v2ModulePath,
	v1ModulePath + "/bcs":            v2ModulePath + "/internal/bcs",
	v1ModulePath + "/crypto":         v2ModulePath + "/internal/crypto",
	v1ModulePath + "/api":            "", // No direct equivalent, inline changes needed
	v1ModulePath + "/internal/types": v2ModulePath + "/internal/types",
}

// typeRenames maps v1 type names to v2 type names (within same package).
var typeRenames = map[string]string{
	// No major type renames currently
}

// funcRenames maps v1 function names to v2 function names.
var funcRenames = map[string]string{
	"BCSSerialize":     "BCSMarshal",
	"BCSDeserialize":   "BCSUnmarshal",
	"Serialize":        "Marshal",
	"Deserialize":      "Unmarshal",
	"SerializeU64":     "SerializeU64",
	"SerializeBytes":   "SerializeBytes",
	"NewClient":        "NewClient",
	"ParseTypeTag":     "ParseTypeTag",
	"MustParseTypeTag": "MustParseTypeTag",
}

// methodsNeedingContext lists Client methods that now require context.Context.
var methodsNeedingContext = map[string]bool{
	"Info":                          true,
	"Account":                       true,
	"AccountResource":               true,
	"AccountResources":              true,
	"AccountResourcesBCS":           true,
	"AccountModule":                 true,
	"BlockByHeight":                 true,
	"BlockByVersion":                true,
	"TransactionByHash":             true,
	"TransactionByVersion":          true,
	"Transactions":                  true,
	"AccountTransactions":           true,
	"SubmitTransaction":             true,
	"SimulateTransaction":           true,
	"BuildTransaction":              true,
	"BuildTransactionMultiAgent":    true,
	"BuildSignAndSubmitTransaction": true,
	"View":                          true,
	"EstimateGasPrice":              true,
	"AccountAPTBalance":             true,
	"GetChainId":                    true,
	"Fund":                          true,
	"WaitForTransaction":            true,
	"PollForTransaction":            true,
	"PollForTransactions":           true,
	"EventsByHandle":                true,
	"EventsByCreationNumber":        true,
	"NodeAPIHealthCheck":            true,
	"EntryFunctionWithArgs":         true,
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: aptosfix [flags] [path ...]\n\n")
		fmt.Fprintf(os.Stderr, "aptosfix migrates Go code from Aptos Go SDK v1 to v2.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  aptosfix -d myfile.go        Preview changes to a single file\n")
		fmt.Fprintf(os.Stderr, "  aptosfix -w .                Fix all files in current directory\n")
		fmt.Fprintf(os.Stderr, "  aptosfix -w -imports .       Only update imports\n")
		fmt.Fprintf(os.Stderr, "  aptosfix -l ./...            List files that need updating\n")
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	var exitCode int
	for _, arg := range args {
		if err := processPath(arg); err != nil {
			fmt.Fprintf(os.Stderr, "aptosfix: %v\n", err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func processPath(path string) error {
	// Handle ./... pattern
	if strings.HasSuffix(path, "/...") {
		root := strings.TrimSuffix(path, "/...")
		if root == "." {
			root = ""
		}
		return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Skip vendor and .git directories
				name := d.Name()
				if name == "vendor" || name == ".git" || name == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
				return processFile(p)
			}
			return nil
		})
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
				if err := processFile(filepath.Join(path, entry.Name())); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return processFile(path)
}

func processFile(filename string) error {
	if *verbose {
		fmt.Fprintf(os.Stderr, "processing: %s\n", filename)
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Check if file uses v1 SDK
	if !bytes.Contains(src, []byte(v1ModulePath)) {
		return nil // No v1 imports, skip
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filename, err)
	}

	// Fix imports
	changed := fixImports(fset, file)

	// Apply other transformations unless -imports flag is set
	if !*importsOnly && fixAST(file) {
		changed = true
	}

	if !changed {
		return nil
	}

	// Format the result
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return fmt.Errorf("formatting %s: %w", filename, err)
	}

	result := buf.Bytes()

	// Apply text-based fixes that can't be done with AST
	result = applyTextFixes(result)

	if *listFiles {
		if !bytes.Equal(src, result) {
			fmt.Println(filename)
		}
		return nil
	}

	if *dryRun {
		fmt.Printf("Would modify: %s\n", filename)
		if *showDiff {
			showDiffOutput(src, result, filename)
		}
		return nil
	}

	if *showDiff {
		showDiffOutput(src, result, filename)
		return nil
	}

	if *writeFile {
		// Write back to file
		perm := os.FileMode(0o644)
		if info, err := os.Stat(filename); err == nil {
			perm = info.Mode().Perm()
		}
		return os.WriteFile(filename, result, perm)
	}

	// Print to stdout
	_, err = os.Stdout.Write(result)
	return err
}

func fixImports(fset *token.FileSet, file *ast.File) bool {
	changed := false

	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		// Check for exact match first
		if newPath, ok := importMapping[path]; ok {
			if newPath == "" {
				// Mark for removal (will need manual intervention)
				continue
			}
			imp.Path.Value = strconv.Quote(newPath)
			changed = true
			continue
		}

		// Check for prefix match (subpackages)
		if strings.HasPrefix(path, v1ModulePath+"/") {
			// Map v1 subpackage to v2
			suffix := strings.TrimPrefix(path, v1ModulePath)
			newPath := v2ModulePath + suffix
			imp.Path.Value = strconv.Quote(newPath)
			changed = true
		}
	}

	// Sort imports using the same fset that was used to parse
	ast.SortImports(fset, file)

	return changed
}

func fixAST(file *ast.File) bool {
	changed := false

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// Check for function renames
			if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
				if newName, ok := funcRenames[sel.Sel.Name]; ok && newName != sel.Sel.Name {
					sel.Sel.Name = newName
					changed = true
				}
			}

		case *ast.SelectorExpr:
			// Check for type renames
			if newName, ok := typeRenames[node.Sel.Name]; ok {
				node.Sel.Name = newName
				changed = true
			}
		}
		return true
	})

	return changed
}

// applyTextFixes applies text-based fixes that are difficult to do with AST manipulation.
func applyTextFixes(src []byte) []byte {
	result := string(src)

	// Add common context import hint as a comment
	if strings.Contains(result, v2ModulePath) && !strings.Contains(result, "\"context\"") {
		// Check if file uses client methods
		for method := range methodsNeedingContext {
			if strings.Contains(result, "."+method+"(") {
				// Add a hint comment at the top
				result = addContextHint(result)
				break
			}
		}
	}

	// Fix common patterns
	for method := range methodsNeedingContext {
		// Match method calls with no arguments: .Method() -> .Method(ctx)
		patternEmpty := regexp.MustCompile(`\.` + regexp.QuoteMeta(method) + `\(\s*\)`)
		result = patternEmpty.ReplaceAllString(result, "."+method+"(ctx)")

		// Match method calls with arguments that don't start with ctx:
		// .Method(arg) -> .Method(ctx, arg)
		// This pattern matches .Method( followed by something that's not "ctx"
		patternWithArgs := regexp.MustCompile(`\.` + regexp.QuoteMeta(method) + `\(\s*([^c\s)]|c[^t]|ct[^x]|ctx[^,\s)])`)
		result = patternWithArgs.ReplaceAllStringFunc(result, func(match string) string {
			// Find the position of the opening paren
			idx := strings.Index(match, "(")
			if idx == -1 {
				return match
			}
			return match[:idx+1] + "ctx, " + match[idx+1:]
		})
	}

	return []byte(result)
}

func addContextHint(src string) string {
	// Find the import block and add a hint
	hint := "// TODO(aptosfix): Add \"context\" import and pass context.Context to client methods\n"

	// Find first import statement
	idx := strings.Index(src, "import")
	if idx == -1 {
		return src
	}

	// Find line start
	lineStart := strings.LastIndex(src[:idx], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	return src[:lineStart] + hint + src[lineStart:]
}

func showDiffOutput(original, modified []byte, filename string) {
	// Try to use external diff tool
	diff, err := computeDiff(original, modified, filename)
	if err != nil {
		// Fall back to simple comparison
		fmt.Printf("--- %s (original)\n", filename)
		fmt.Printf("+++ %s (modified)\n", filename)
		fmt.Println(string(modified))
		return
	}
	fmt.Print(string(diff))
}

func computeDiff(original, modified []byte, filename string) ([]byte, error) {
	// Write to temp files
	origFile, err := os.CreateTemp("", "aptosfix-orig-*.go")
	if err != nil {
		return nil, err
	}
	defer os.Remove(origFile.Name())
	defer origFile.Close()

	modFile, err := os.CreateTemp("", "aptosfix-mod-*.go")
	if err != nil {
		return nil, err
	}
	defer os.Remove(modFile.Name())
	defer modFile.Close()

	if _, err := origFile.Write(original); err != nil {
		return nil, err
	}
	if _, err := modFile.Write(modified); err != nil {
		return nil, err
	}

	// Close files before running diff
	origFile.Close()
	modFile.Close()

	// Run diff - G204: filename comes from development tool, not user input
	cmd := exec.Command("diff", "-u", "--label", filename+" (original)", origFile.Name(), "--label", filename+" (modified)", modFile.Name()) //nolint:gosec
	output, _ := cmd.Output()                                                                                                                // diff returns non-zero exit code when files differ
	return output, nil
}

// Report generates a migration report for a directory.
type Report struct {
	Files           []FileReport
	TotalFiles      int
	FilesWithV1     int
	ImportChanges   int
	ContextRequired int
	ManualReview    int
}

// FileReport contains migration info for a single file.
type FileReport struct {
	Path              string
	HasV1Import       bool
	ImportChanges     []string
	ContextMethods    []string
	ManualReviewItems []string
}

// GenerateReport analyzes a path and generates a migration report.
func GenerateReport(w io.Writer, path string) error {
	report := &Report{}

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" || name == "v2" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}

		report.TotalFiles++

		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		if !bytes.Contains(src, []byte(v1ModulePath)) {
			return nil
		}

		fileReport := analyzeFile(p, src)
		if fileReport.HasV1Import {
			report.FilesWithV1++
			report.ImportChanges += len(fileReport.ImportChanges)
			report.ContextRequired += len(fileReport.ContextMethods)
			report.ManualReview += len(fileReport.ManualReviewItems)
			report.Files = append(report.Files, fileReport)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Print report
	fmt.Fprintf(w, "Aptos Go SDK v1 to v2 Migration Report\n")
	fmt.Fprintf(w, "======================================\n\n")
	fmt.Fprintf(w, "Summary:\n")
	fmt.Fprintf(w, "  Total Go files scanned: %d\n", report.TotalFiles)
	fmt.Fprintf(w, "  Files using v1 SDK:     %d\n", report.FilesWithV1)
	fmt.Fprintf(w, "  Import changes needed:  %d\n", report.ImportChanges)
	fmt.Fprintf(w, "  Context additions:      %d\n", report.ContextRequired)
	fmt.Fprintf(w, "  Manual review items:    %d\n", report.ManualReview)
	fmt.Fprintf(w, "\n")

	if len(report.Files) > 0 {
		fmt.Fprintf(w, "Files requiring changes:\n")
		for _, f := range report.Files {
			fmt.Fprintf(w, "\n  %s\n", f.Path)
			if len(f.ImportChanges) > 0 {
				fmt.Fprintf(w, "    Import changes:\n")
				for _, ic := range f.ImportChanges {
					fmt.Fprintf(w, "      - %s\n", ic)
				}
			}
			if len(f.ContextMethods) > 0 {
				fmt.Fprintf(w, "    Methods needing context:\n")
				for _, m := range f.ContextMethods {
					fmt.Fprintf(w, "      - %s\n", m)
				}
			}
			if len(f.ManualReviewItems) > 0 {
				fmt.Fprintf(w, "    Manual review needed:\n")
				for _, item := range f.ManualReviewItems {
					fmt.Fprintf(w, "      - %s\n", item)
				}
			}
		}
	}

	return nil
}

func analyzeFile(path string, src []byte) FileReport {
	report := FileReport{Path: path}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		report.ManualReviewItems = append(report.ManualReviewItems, "Parse error: "+err.Error())
		return report
	}

	// Check imports
	for _, imp := range file.Imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		if strings.HasPrefix(importPath, v1ModulePath) {
			report.HasV1Import = true
			if newPath, ok := importMapping[importPath]; ok {
				if newPath == "" {
					report.ManualReviewItems = append(report.ManualReviewItems,
						fmt.Sprintf("Import %q has no v2 equivalent", importPath))
				} else {
					report.ImportChanges = append(report.ImportChanges,
						fmt.Sprintf("%q -> %q", importPath, newPath))
				}
			} else if strings.HasPrefix(importPath, v1ModulePath+"/") {
				suffix := strings.TrimPrefix(importPath, v1ModulePath)
				newPath := v2ModulePath + suffix
				report.ImportChanges = append(report.ImportChanges,
					fmt.Sprintf("%q -> %q", importPath, newPath))
			}
		}
	}

	// Find method calls that need context
	methodsUsed := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				methodName := sel.Sel.Name
				if methodsNeedingContext[methodName] {
					methodsUsed[methodName] = true
				}
			}
		}
		return true
	})

	// Sort method names for consistent output
	var methods []string
	for m := range methodsUsed {
		methods = append(methods, m)
	}
	sort.Strings(methods)
	report.ContextMethods = methods

	return report
}
