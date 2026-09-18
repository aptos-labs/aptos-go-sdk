package compatibility

import (
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const v1ModulePath = "github.com/aptos-labs/aptos-go-sdk"

// TestV1DependencyIsPublishedRelease ensures a v2 release cannot ship against
// an older or untagged v1. Consumers inherit this require from v2/go.mod, so
// tagging v2.2.0 before bumping it would publish v2 still on v1.13.0.
func TestV1DependencyIsPublishedRelease(t *testing.T) {
	t.Parallel()

	info, ok := debug.ReadBuildInfo()
	require.True(t, ok, "ReadBuildInfo")

	var version string
	for _, d := range info.Deps {
		if d.Path == v1ModulePath {
			version = d.Version
			break
		}
	}
	require.NotEmpty(t, version, "v2 must require %s", v1ModulePath)
	require.False(t, strings.Contains(version, "-"), "v1 must be a tagged release, not a pseudo-version; got %s", version)
	require.Equal(t, "v1.14.0", version)
}
