package confidentialasset_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// .../v2/confidentialasset/compile_constraint_test.go -> v2
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func goBuild(t *testing.T, cgoEnabled string, pkg string) (stdout, stderr []byte, err error) {
	t.Helper()
	root := moduleRoot(t)
	cmd := exec.Command("go", "build", "-o", os.DevNull, pkg)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED="+cgoEnabled)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

func TestNativeImportFailsWithoutCGO(t *testing.T) {
	_, stderr, err := goBuild(t, "0", "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/native")
	if err == nil {
		t.Fatal("expected go build of native to fail when CGO_ENABLED=0")
	}
	combined := string(stderr)
	if !strings.Contains(combined, "build constraints exclude all Go files") {
		t.Fatalf("stderr should mention excluded build constraints; got:\n%s", combined)
	}
}

func TestRootPackageBuildsWithoutCGO(t *testing.T) {
	_, stderr, err := goBuild(t, "0", "github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset")
	if err != nil {
		t.Fatalf("root confidentialasset should build without CGO: %v\n%s", err, stderr)
	}
}
