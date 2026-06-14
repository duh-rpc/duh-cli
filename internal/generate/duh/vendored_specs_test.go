package duh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// steve's real spec is vendored under testdata/ from the mono-repo as the single
// integration fixture exercising all three ENG-102 defects at once (two-segment
// /admin/jobs.* paths, octet-stream streaming, and additionalProperties maps). The
// per-defect mechanisms are covered by the focused specs in protobuf_twin_test.go,
// multisegment_test.go, and openapi-schema.go's map tests. Re-vendor if the contract
// changes: platform/libs/go/steve/v1/api/v1/openapi.yaml.

// writeVendoredProject reads a vendored spec from testdata/ (via the captured package
// directory, since other tests chdir into a t.TempDir() without restoring) and lays
// it down as a generatable project.
func writeVendoredProject(t *testing.T, name string) string {
	t.Helper()
	// Read via the captured package directory: other tests chdir into a t.TempDir()
	// without restoring, so cwd is not reliably the package directory here.
	data, err := os.ReadFile(filepath.Join(pkgDir, "testdata", name))
	require.NoError(t, err)
	return writeProject(t, string(data))
}

// steve is the end-to-end acceptance for Defects 2+3 together: two-segment
// /admin/jobs.* paths, octet-stream streaming responses, and additionalProperties
// maps. Generation must succeed, the proto must carry the map fields, the lock must be
// populated, and lint must pass clean — the exact ENG-101 outcome.
func TestGenerateVendoredSteveSpec(t *testing.T) {
	dir := writeVendoredProject(t, "steve_openapi.yaml")

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	assertValidGo(t, filepath.Join(dir, "out", "server.go"))
	assertValidGo(t, filepath.Join(dir, "out", "client.go"))

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "RPCAdminJobsTypes")
	assert.NotContains(t, server, "Admin/jobs")

	proto := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.Contains(t, proto, "map<string, string> details")
	assert.Contains(t, proto, "map<string, string> params")

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "Reply:")
	assert.Contains(t, lock, "details: {number:")

	exitCode, out = lint(t)
	require.Equal(t, 0, exitCode, out)
	assert.Contains(t, out, "compliant")
}
