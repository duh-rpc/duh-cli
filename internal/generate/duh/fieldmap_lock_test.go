package duh_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// widgetSpec returns a lint-clean DUH spec whose request message carries the given
// properties block (each line indented 8 spaces). The rest of the spec is fixed and
// compliant so tests only vary the fields under test.
func widgetSpec(requestProperties string) string {
	return `openapi: 3.0.3
info:
  title: Widget API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    post:
      summary: Create a widget
      description: Creates a new widget resource
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WidgetsCreateRequest'
      responses:
        '200':
          description: Widget created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WidgetsCreateResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    WidgetsCreateRequest:
      type: object
      description: Request payload to create a widget
      properties:
` + requestProperties + `
    WidgetsCreateResponse:
      type: object
      description: Response payload after creating a widget
      properties:
        widget_id:
          type: string
          description: The created widget identifier
`
}

const (
	propName = `        name:
          type: string
          description: The widget name`
	propColor = `        color:
          type: string
          description: The widget color`
	propSize = `        size:
          type: integer
          format: int32
          description: The widget size`
)

// pkgDir is the package directory, captured once at package initialization while
// the working directory is still valid. Other tests in this package chdir into a
// t.TempDir() without restoring, so by the time these tests run the prior cwd may be
// a deleted directory — calling os.Getwd() then fails on Linux (ENOENT). Restoring
// to this captured directory avoids that.
var pkgDir = func() string {
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return d
}()

// writeProject lays down a go.mod (required for code generation) and the spec, and
// chdir's into the temp dir so the default lock path resolves next to the spec.
func writeProject(t *testing.T, spec string) string {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() { _ = os.Chdir(pkgDir) })

	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/example/widget\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(spec), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "out"), 0755))
	require.NoError(t, os.Chdir(dir))
	return dir
}

func generate(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, append([]string{"generate", "openapi.yaml"}, args...))
	return exitCode, stdout.String()
}

func lint(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, append([]string{"lint", "openapi.yaml"}, args...))
	return exitCode, stdout.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

// Acceptance #1: first-run seeding creates fieldmap.lock adjacent to the spec,
// mapping each field by declaration order; the lock is never written to --output-dir.
func TestFieldmapLockFirstRunSeeding(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lockPath := filepath.Join(dir, "fieldmap.lock")
	_, err := os.Stat(lockPath)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "out", "fieldmap.lock"))
	assert.True(t, os.IsNotExist(err))

	lock := readFile(t, lockPath)
	assert.Contains(t, lock, "name: {number: 1}")
	assert.Contains(t, lock, "color: {number: 2}")
	assert.Contains(t, lock, "widget_id: {number: 1}")
}

// Acceptance #1: a --lock-path flag overrides the default location.
func TestFieldmapLockPathOverride(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName))

	custom := filepath.Join(dir, "contract", "fieldmap.lock")
	require.NoError(t, os.MkdirAll(filepath.Dir(custom), 0755))

	exitCode, out := generate(t, "--output-dir", "out", "--lock-path", custom)
	require.Equal(t, 0, exitCode, out)

	_, err := os.Stat(custom)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "fieldmap.lock"))
	assert.True(t, os.IsNotExist(err))
}

// Acceptance #2: inserting a field mid-schema preserves existing numbers and assigns
// the new field the next available number; the proto reflects locked numbers.
func TestFieldmapLockMidSchemaInsertionPreservesNumbers(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Insert 'size' between name and color.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propName+"\n"+propSize+"\n"+propColor)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "name: {number: 1}")
	assert.Contains(t, lock, "color: {number: 2}")
	assert.Contains(t, lock, "size: {number: 3}")

	proto := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.Contains(t, proto, "string name = 1")
	assert.Contains(t, proto, "string color = 2")
	assert.Contains(t, proto, "int32 size = 3")
}

// Acceptance #3: reordering fields (same names) yields a byte-identical lock and
// proto, and lint passes clean.
func TestFieldmapLockReorderIsNoOp(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lockBefore, err := os.ReadFile(filepath.Join(dir, "fieldmap.lock"))
	require.NoError(t, err)
	protoBefore, err := os.ReadFile(filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	require.NoError(t, err)

	// Reorder: color before name.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propColor+"\n"+propName)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lockAfter, err := os.ReadFile(filepath.Join(dir, "fieldmap.lock"))
	require.NoError(t, err)
	protoAfter, err := os.ReadFile(filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	require.NoError(t, err)

	assert.Equal(t, string(lockBefore), string(lockAfter))
	assert.Equal(t, string(protoBefore), string(protoAfter))

	exitCode, out = lint(t)
	assert.Equal(t, 0, exitCode, out)
}

// Acceptance #4: removing a field retains its number as reserved, and a later-added
// field never reuses a reserved number.
func TestFieldmapLockRemovalReserves(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Remove 'name' (number 1).
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propColor)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "name: {number: 1, reserved: true}")
	assert.Contains(t, lock, "color: {number: 2}")

	// Add 'size'; it must take 3, never the reserved 1.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propColor+"\n"+propSize)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock = readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "name: {number: 1, reserved: true}")
	assert.Contains(t, lock, "size: {number: 3}")
	assert.NotContains(t, lock, "size: {number: 1}")

	proto := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.Contains(t, proto, "reserved 1;")
}

// Acceptance #5: renaming a field reserves the old name's number and assigns the new
// name the next available number, without error.
func TestFieldmapLockRenameIsRemoveAndAdd(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Rename 'name' → 'title'.
	renamed := `        title:
          type: string
          description: The widget title`
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(renamed+"\n"+propColor)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "name: {number: 1, reserved: true}")
	assert.Contains(t, lock, "color: {number: 2}")
	assert.Contains(t, lock, "title: {number: 3}")
}

// Acceptance #4 (re-add): re-adding a previously-removed field gets the next
// available number and never reuses the removed number, which stays reserved. The
// reservation must also be stable across a further regeneration.
func TestFieldmapLockReaddFieldKeepsOldNumberReserved(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Remove 'name' (number 1) → tombstone.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propColor)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Re-add 'name': it must take a fresh number (3), never the reserved 1.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propColor+"\n"+propName)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "name: {number: 3}")
	assert.NotContains(t, lock, "name: {number: 1}")
	assert.Contains(t, lock, "reserved: [1]")

	proto := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.Contains(t, proto, "string name = 3")
	assert.Contains(t, proto, "reserved 1;")

	// Regenerating leaves the reservation stable and lint clean.
	before := lock
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)
	assert.Equal(t, before, readFile(t, filepath.Join(dir, "fieldmap.lock")))

	exitCode, out = lint(t)
	assert.Equal(t, 0, exitCode, out)
}

// Acceptance #7: two generate runs against the same spec + lock produce identical
// field numbering (byte-identical lock).
func TestFieldmapLockIndependentRegenerationsAgree(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)
	first, err := os.ReadFile(filepath.Join(dir, "fieldmap.lock"))
	require.NoError(t, err)

	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)
	second, err := os.ReadFile(filepath.Join(dir, "fieldmap.lock"))
	require.NoError(t, err)

	assert.Equal(t, string(first), string(second))
}

// Acceptance #8: lint fails when the lock is missing a mapping the spec requires.
func TestFieldmapLockLintCatchesStaleLock(t *testing.T) {
	writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Spec gains 'size' but generate is not re-run: the committed lock is stale.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propName+"\n"+propColor+"\n"+propSize)), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "size")
}

// Acceptance #8: lint fails on a reused number within a message.
func TestFieldmapLockLintCatchesReusedNumber(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
            color: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
`), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "number 1 is used by both")
}

// Acceptance #8: lint fails when a live spec field is assigned a reserved number.
func TestFieldmapLockLintCatchesLiveFieldOnReservedNumber(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1, reserved: true}
            color: {number: 2}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
`), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "name")
}

// Acceptance #8: lint fails on a structurally invalid (merge-mangled) lock.
func TestFieldmapLockLintCatchesStructurallyInvalid(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"),
		[]byte("version: 1\nmessages: [this is not\n  a mapping: {{{\n"), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
}

// Acceptance #9: when generate would be forced to change an existing mapping (a
// corrupt lock with a duplicated number), it exits non-zero, writes no files, and
// names the offending message/field/number.
func TestFieldmapLockGenerateFailsLoudOnCorruptLock(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
            color: {number: 1}
`), 0644))
	lockBefore := readFile(t, filepath.Join(dir, "fieldmap.lock"))

	exitCode, out := generate(t, "--output-dir", "out")
	assert.Equal(t, 2, exitCode)
	assert.Contains(t, out, "WidgetsCreateRequest")
	assert.Contains(t, out, "1")
	assert.Contains(t, out, "no files written")

	// No artifacts were written.
	_, err := os.Stat(filepath.Join(dir, "out", "server.go"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.True(t, os.IsNotExist(err))

	// The corrupt lock is left untouched.
	assert.Equal(t, lockBefore, readFile(t, filepath.Join(dir, "fieldmap.lock")))
}

// Acceptance #10: deleting fieldmap.lock and regenerating reseeds from current
// declaration order (a valid recovery path while nothing has shipped).
func TestFieldmapLockPrePublishCleanup(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Remove 'name' so it becomes reserved, accumulating a tombstone.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propColor)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)
	assert.Contains(t, readFile(t, filepath.Join(dir, "fieldmap.lock")), "reserved: true")

	// Delete the lock and regenerate: it reseeds with no tombstones, color at 1.
	require.NoError(t, os.Remove(filepath.Join(dir, "fieldmap.lock")))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.NotContains(t, lock, "reserved: true")
	assert.Contains(t, lock, "color: {number: 1}")
}

// Acceptance #11: running lint against a present lock leaves it byte-identical, and
// running lint when no lock exists creates none.
func TestFieldmapLockLintNeverWritesLock(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))

	// No lock yet: lint passes and creates none.
	exitCode, out := lint(t)
	assert.Equal(t, 0, exitCode, out)
	_, err := os.Stat(filepath.Join(dir, "fieldmap.lock"))
	assert.True(t, os.IsNotExist(err))

	// With a present lock: lint leaves it byte-identical.
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)
	before := readFile(t, filepath.Join(dir, "fieldmap.lock"))

	exitCode, out = lint(t)
	assert.Equal(t, 0, exitCode, out)
	assert.Equal(t, before, readFile(t, filepath.Join(dir, "fieldmap.lock")))
}

// enumSpec returns a lint-clean-except-for-the-enum-rule DUH spec with a top-level
// integer enum (a proto enum) referenced by a response field, plus the variants
// block under test (each line indented 8 spaces).
func enumSpec(variants string) string {
	return `openapi: 3.0.3
info:
  title: Widget API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    post:
      summary: Create a widget
      description: Creates a new widget resource
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WidgetsCreateRequest'
      responses:
        '200':
          description: Widget created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WidgetsCreateResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    Code:
      type: integer
      format: int32
      description: A status code enum
      enum:
` + variants + `
    WidgetsCreateRequest:
      type: object
      description: Request payload to create a widget
      properties:
        name:
          type: string
          description: The widget name
    WidgetsCreateResponse:
      type: object
      description: Response payload after creating a widget
      properties:
        widget_id:
          type: string
          description: The created widget identifier
        code:
          $ref: '#/components/schemas/Code'
          description: The status code
`
}

// enumLock is a committed lock for enumSpec with the Code variants keyed by their
// literal value (proto numbers 0,1,2).
const enumLock = `version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
            code: {number: 2}
enums:
    Code:
        variants:
            "200": {number: 0}
            "404": {number: 1}
            "500": {number: 2}
`

// Acceptance #6 (validation surface): an enum's variants are keyed by literal value,
// so reordering them in the spec does not disturb the locked proto integers and lint
// still passes. The ENUM_UNSPECIFIED_VARIANT rule is disabled because a bare integer
// enum (a proto enum) legitimately has no *_UNSPECIFIED sentinel.
func TestFieldmapLockEnumReorderIsSafe(t *testing.T) {
	dir := writeProject(t, enumSpec("        - 200\n        - 404\n        - 500"))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(enumLock), 0644))

	exitCode, out := lint(t, "--disable", "ENUM_UNSPECIFIED_VARIANT")
	require.Equal(t, 0, exitCode, out)

	// Reorder the enum variants; lint must still pass because numbers are keyed by
	// literal value, not declaration order.
	require.NoError(t, os.WriteFile("openapi.yaml",
		[]byte(enumSpec("        - 500\n        - 200\n        - 404")), 0644))

	exitCode, out = lint(t, "--disable", "ENUM_UNSPECIFIED_VARIANT")
	assert.Equal(t, 0, exitCode, out)
}

// Acceptance #6 / #8 (validation surface): lint fails when the lock is missing a
// variant the spec's enum requires.
func TestFieldmapLockEnumCompleteness(t *testing.T) {
	dir := writeProject(t, enumSpec("        - 200\n        - 404\n        - 500"))

	// Lock omits the 500 variant.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
            code: {number: 2}
enums:
    Code:
        variants:
            "200": {number: 0}
            "404": {number: 1}
`), 0644))

	exitCode, out := lint(t, "--disable", "ENUM_UNSPECIFIED_VARIANT")
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "500")
}

// Acceptance #6 (re-add, validation surface): an enum whose variant was removed and
// re-added keeps the old proto number reserved while the re-added variant carries a
// fresh number. A lock recording exactly that passes lint; a lock that reuses the
// reserved number for the live variant is rejected.
func TestFieldmapLockEnumReaddVariantReserved(t *testing.T) {
	dir := writeProject(t, enumSpec("        - 404\n        - 500\n        - 200"))

	// '200' was removed (reserving proto 0) and re-added at the next number (3).
	validLock := `version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
            code: {number: 2}
enums:
    Code:
        variants:
            "404": {number: 1}
            "500": {number: 2}
            "200": {number: 3}
        reserved: [0]
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(validLock), 0644))
	exitCode, out := lint(t, "--disable", "ENUM_UNSPECIFIED_VARIANT")
	assert.Equal(t, 0, exitCode, out)

	// A lock that hands the live '200' variant the reserved number 0 is rejected.
	badLock := `version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
            code: {number: 2}
enums:
    Code:
        variants:
            "404": {number: 1}
            "500": {number: 2}
            "200": {number: 0}
        reserved: [0]
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(badLock), 0644))
	exitCode, out = lint(t, "--disable", "ENUM_UNSPECIFIED_VARIANT")
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
}

// Absent lock is a first-class state: lint passes silently with no lock-related
// output even when other rules are evaluated.
func TestFieldmapLockAbsentLockPassesSilently(t *testing.T) {
	writeProject(t, widgetSpec(propName))

	exitCode, out := lint(t)
	assert.Equal(t, 0, exitCode, out)
	assert.NotContains(t, out, "FIELDMAP_LOCK")
}

// Acceptance #8: lint fails when a lock entry is live but the spec no longer
// declares it (a field was removed from the spec without regenerating). This is the
// inverse of the stale-lock case, where the spec gains a field the lock lacks.
func TestFieldmapLockLintCatchesLiveFieldRemovedFromSpec(t *testing.T) {
	writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Remove 'color' from the spec without regenerating; the lock still has it live.
	require.NoError(t, os.WriteFile("openapi.yaml", []byte(widgetSpec(propName)), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "color")
	assert.Contains(t, out, "live")
}

// Acceptance #8: lint fails when the lock has no section at all for a spec message,
// not merely a missing field within a present section.
func TestFieldmapLockLintCatchesMissingMessageSection(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName+"\n"+propColor))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Drop the entire WidgetsCreateResponse section from the lock.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
            color: {number: 2}
`), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "WidgetsCreateResponse")
}

// Tech-spec "Number assignment" / "Determinism" (PRD Open Question #4): when
// multiple new fields appear in a single generate run they are numbered in OpenAPI
// declaration order among the new ones, and a second run is byte-identical.
func TestFieldmapLockMultiNewFieldDeterminism(t *testing.T) {
	dir := writeProject(t, widgetSpec(propName))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Add two new fields at once: size is declared before color, so size must take
	// the lower number regardless of high-water assignment order.
	require.NoError(t, os.WriteFile("openapi.yaml",
		[]byte(widgetSpec(propName+"\n"+propSize+"\n"+propColor)), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "name: {number: 1}")
	assert.Contains(t, lock, "size: {number: 2}")
	assert.Contains(t, lock, "color: {number: 3}")

	first := lock
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)
	assert.Equal(t, first, readFile(t, filepath.Join(dir, "fieldmap.lock")))
}

// protoEnumSpec returns a DUH spec with a top-level integer enum (a real proto enum)
// referenced by a response field. The enum carries x-duh-lint-ignore so it passes the
// ENUM_UNSPECIFIED_VARIANT rule, which duh generate runs un-disabled; this is the only
// way an integer enum reaches generate's lock/proto enum path. variants is the enum
// block (each line indented 6 spaces, e.g. "      - 0").
func protoEnumSpec(variants string) string {
	return `openapi: 3.0.3
info:
  title: Widget API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    post:
      summary: Create a widget
      description: Creates a new widget resource
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WidgetsCreateRequest'
      responses:
        '200':
          description: Widget created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WidgetsCreateResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    Priority:
      type: integer
      format: int32
      description: A proto enum priority
      x-duh-lint-ignore:
        - ENUM_UNSPECIFIED_VARIANT
      enum:
` + variants + `
    WidgetsCreateRequest:
      type: object
      description: Request payload to create a widget
      properties:
        name:
          type: string
          description: The widget name
    WidgetsCreateResponse:
      type: object
      description: Response payload after creating a widget
      properties:
        widget_id:
          type: string
          description: The created widget identifier
        priority:
          $ref: '#/components/schemas/Priority'
          description: The priority
`
}

// Acceptance #6 (generate surface): first-run generate on a spec with a proto
// (integer) enum seeds the lock's enums section keyed by literal value and emits a
// proto enum whose numbers come from the lock. This exercises the enum write path
// (Reconcile's enum loop, priorEnum, toEnumNumbers, enumsNode) end-to-end, which the
// lint-only enum tests never reach.
func TestFieldmapLockEnumGenerateSeedsLock(t *testing.T) {
	dir := writeProject(t, protoEnumSpec("        - 0\n        - 1\n        - 2"))

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "Priority:")
	assert.Contains(t, lock, `"0": {number: 0}`)
	assert.Contains(t, lock, `"1": {number: 1}`)
	assert.Contains(t, lock, `"2": {number: 2}`)

	proto := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.Contains(t, proto, "enum Priority {")
	assert.Contains(t, proto, "PRIORITY_0 = 0;")
	assert.Contains(t, proto, "PRIORITY_1 = 1;")
	assert.Contains(t, proto, "PRIORITY_2 = 2;")
}

// Acceptance #6 (generate surface): reordering a proto enum's variants yields a
// byte-identical lock and proto, because numbers are keyed by literal value, not
// declaration order. This is the generate-path counterpart to the lint-only
// TestFieldmapLockEnumReorderIsSafe.
func TestFieldmapLockEnumGenerateReorderIsNoOp(t *testing.T) {
	dir := writeProject(t, protoEnumSpec("        - 0\n        - 1\n        - 2"))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lockBefore := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	protoBefore := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))

	// Reorder the variants; numbers are pinned by value, so nothing changes.
	require.NoError(t, os.WriteFile("openapi.yaml",
		[]byte(protoEnumSpec("        - 0\n        - 2\n        - 1")), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	assert.Equal(t, lockBefore, readFile(t, filepath.Join(dir, "fieldmap.lock")))
	assert.Equal(t, protoBefore, readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto")))
}

// sentinelEnumSpec returns a DUH spec with a top-level integer enum whose variant
// values are sentinel-named strings (e.g. CODE_UNSPECIFIED, CODE_OK, CODE_ERR).
// Because type: integer with string enum values is legal OpenAPI, and CODE_UNSPECIFIED
// is declared first so ENUM_UNSPECIFIED_VARIANT passes, generate succeeds with no
// x-duh-lint-ignore. The lock keys are the literal variant values, so they satisfy
// isUnspecifiedSentinel — this is the only legitimate path that produces a
// sentinel-keyed lock entry. variants is the enum block (each line indented 6 spaces).
func sentinelEnumSpec(variants string) string {
	return `openapi: 3.0.3
info:
  title: Widget API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    post:
      summary: Create a widget
      description: Creates a new widget resource
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WidgetsCreateRequest'
      responses:
        '200':
          description: Widget created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WidgetsCreateResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    Code:
      type: integer
      format: int32
      description: A status code enum
      enum:
` + variants + `
    WidgetsCreateRequest:
      type: object
      description: Request payload to create a widget
      properties:
        name:
          type: string
          description: The widget name
    WidgetsCreateResponse:
      type: object
      description: Response payload after creating a widget
      properties:
        widget_id:
          type: string
          description: The created widget identifier
        code:
          $ref: '#/components/schemas/Code'
          description: The status code
`
}

// Acceptance #8 (checkEnumInvariant finding a): lint catches a hand-mangled lock
// where the *_UNSPECIFIED sentinel is assigned a non-zero proto number. A
// generate-produced lock always assigns the sentinel 0 (enumBase), so this state
// is only reachable via direct lock editing or a merge conflict.
func TestFieldmapLockLintCatchesSentinelAtNonZero(t *testing.T) {
	dir := writeProject(t, sentinelEnumSpec("        - CODE_UNSPECIFIED\n        - CODE_OK\n        - CODE_ERR"))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Hand-mangle: sentinel mapped to 1 instead of 0.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
            code: {number: 2}
enums:
    Code:
        variants:
            CODE_UNSPECIFIED: {number: 1}
            CODE_OK: {number: 2}
            CODE_ERR: {number: 3}
`), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "must be 0")
}

// Acceptance #8 (checkEnumInvariant finding b): lint catches a hand-mangled lock
// where a non-sentinel variant occupies proto 0 while an *_UNSPECIFIED sentinel
// exists at a non-zero number. Proto 0 is the wire default and must be reserved
// exclusively for the sentinel.
func TestFieldmapLockLintCatchesNonSentinelAtProtoZero(t *testing.T) {
	dir := writeProject(t, sentinelEnumSpec("        - CODE_UNSPECIFIED\n        - CODE_OK\n        - CODE_ERR"))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Hand-mangle: non-sentinel CODE_OK at 0, sentinel CODE_UNSPECIFIED displaced to 2.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fieldmap.lock"), []byte(
		`version: 1
messages:
    ErrorDetails:
        fields:
            message: {number: 1}
    WidgetsCreateRequest:
        fields:
            name: {number: 1}
    WidgetsCreateResponse:
        fields:
            widget_id: {number: 1}
            code: {number: 2}
enums:
    Code:
        variants:
            CODE_UNSPECIFIED: {number: 2}
            CODE_OK: {number: 0}
            CODE_ERR: {number: 3}
`), 0644))

	exitCode, out = lint(t)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, out, "FIELDMAP_LOCK")
	assert.Contains(t, out, "occupies proto 0")
}

// Acceptance #9 (assertUnspecifiedFirst): generate refuses to seed an enum whose
// *_UNSPECIFIED sentinel exists but is not declared first in the spec. The only
// way to bypass the ENUM_UNSPECIFIED_VARIANT lint rule (which generate runs
// without --disable) is per-schema x-duh-lint-ignore. With the sentinel out of
// position, assertUnspecifiedFirst fires before any file is written.
func TestFieldmapLockGenerateFailsWhenSentinelNotFirst(t *testing.T) {
	// sentinel-not-first spec: x-duh-lint-ignore suppresses ENUM_UNSPECIFIED_VARIANT
	// so the full lint passes, but CODE_OK is declared before CODE_UNSPECIFIED.
	spec := `openapi: 3.0.3
info:
  title: Widget API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    post:
      summary: Create a widget
      description: Creates a new widget resource
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WidgetsCreateRequest'
      responses:
        '200':
          description: Widget created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WidgetsCreateResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    Code:
      type: integer
      format: int32
      description: A status code enum
      x-duh-lint-ignore:
        - ENUM_UNSPECIFIED_VARIANT
      enum:
        - CODE_OK
        - CODE_UNSPECIFIED
        - CODE_ERR
    WidgetsCreateRequest:
      type: object
      description: Request payload to create a widget
      properties:
        name:
          type: string
          description: The widget name
    WidgetsCreateResponse:
      type: object
      description: Response payload after creating a widget
      properties:
        widget_id:
          type: string
          description: The created widget identifier
        code:
          $ref: '#/components/schemas/Code'
          description: The status code
`
	dir := writeProject(t, spec)

	exitCode, out := generate(t, "--output-dir", "out")
	assert.Equal(t, 2, exitCode)
	// Error names the first (mis-ordered) variant and the missing *_UNSPECIFIED requirement.
	assert.Contains(t, out, "CODE_OK")
	assert.Contains(t, out, "no files written")

	// No artifacts written.
	_, err := os.Stat(filepath.Join(dir, "out", "server.go"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(dir, "fieldmap.lock"))
	assert.True(t, os.IsNotExist(err))
}

// Acceptance #6 (generate surface): removing a proto enum variant retains its number
// as a reserved tombstone in the lock and emits a proto reserved statement; the
// surviving variants keep their numbers.
func TestFieldmapLockEnumGenerateRemovalReserves(t *testing.T) {
	dir := writeProject(t, protoEnumSpec("        - 0\n        - 1\n        - 2"))
	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	// Remove the middle variant (1); its number must be retained as reserved.
	require.NoError(t, os.WriteFile("openapi.yaml",
		[]byte(protoEnumSpec("        - 0\n        - 2")), 0644))
	exitCode, out = generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, `"0": {number: 0}`)
	assert.Contains(t, lock, `"1": {number: 1, reserved: true}`)
	assert.Contains(t, lock, `"2": {number: 2}`)

	proto := readFile(t, filepath.Join(dir, "out", "proto", "v1", "api.proto"))
	assert.Contains(t, proto, "reserved 1;")
}
