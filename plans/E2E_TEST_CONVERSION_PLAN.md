# E2E Test Conversion Implementation Plan

## Overview

This plan outlines the simplification of the CLI interface followed by the conversion of 8 unit test files in the duhrpc-lint project from non-functional unit tests to functional end-to-end (e2e) tests.

**Phase 1** simplifies the `RunCmd()` signature by removing unused `stdin` and `stderr` parameters.

**Phases 2-9** convert test files from inline YAML with direct `rule.Validate()` calls to functional tests that write YAML to temp files and call `lint.RunCmd()`.

**Total scope**: 1 code simplification + 8 test file conversions, 2,104 lines of test code, approximately 40+ individual test cases

## Current State Analysis

### What Exists
- **`RunCmd()` signature**: `func RunCmd(stdin io.Reader, stdout, stderr io.Writer, args []string) int`
  - `stdin` parameter is never used
  - `stderr` is used for errors, but all output can go to `stdout`
  - Overly complex for testing

- **8 unit test files** in `internal/rules/`:
  1. `path_format_test.go` (359 lines, 12 test cases)
  2. `http_method_test.go` (224 lines, 7 test cases)
  3. `query_params_test.go` (232 lines, 6 test cases)
  4. `request_body_test.go` (115 lines, 4 test cases)
  5. `status_code_test.go` (158 lines, 7 test cases)
  6. `success_response_test.go` (163 lines, 6 test cases)
  7. `content_type_test.go` (308 lines, 10 test cases)
  8. `error_response_test.go` (545 lines, 10+ test cases)

### Current Test Pattern
```go
func TestPathFormatRule(t *testing.T) {
    for _, test := range []struct {
        name            string
        spec            string
        expectViolation bool
        violationCount  int
    }{
        // test cases...
    } {
        t.Run(test.name, func(t *testing.T) {
            // Parse YAML inline
            doc, err := libopenapi.NewDocument([]byte(test.spec))
            require.NoError(t, err)

            model, errs := doc.BuildV3Model()
            require.Empty(t, errs)

            // Call rule directly
            rule := rules.NewPathFormatRule()
            violations := rule.Validate(&model.Model)

            // Check violations
            if test.expectViolation {
                require.NotEmpty(t, violations)
                require.Len(t, violations, test.violationCount)
            } else {
                require.Empty(t, violations)
            }
        })
    }
}
```

### What's Changing

**Phase 1 - RunCmd Simplification**:
1. Remove `stdin io.Reader` parameter (never used)
2. Remove `stderr io.Writer` parameter
3. All output (success, violations, errors) goes to `stdout`
4. Update `run_cmd.go`, `main.go`, and all existing tests

**Phases 2-9 - Test Conversion**:
1. **Test execution**: From `rule.Validate()` to `lint.RunCmd()`
2. **Test assertions**: From checking violation structs to checking exit codes and output strings
3. **File handling**: From inline YAML strings to temporary files using `t.TempDir()`
4. **Import dependencies**: Remove libopenapi imports, add lint package import
5. **Test structure**: Maintain table-driven tests but simplify assertion logic

### Exit Code Semantics
- **0**: Spec is valid, no violations found
- **1**: Violations found (spec is invalid)
- **2**: Error occurred (file not found, parse error, etc.)

---

## Phase 1: Simplify RunCmd Signature

### Overview
Simplify the CLI interface by removing unused parameters and consolidating all output to stdout. This makes the code simpler and testing easier.

**Goal**: Single output stream, simpler function signature, cleaner test code.

### Changes Required

#### 1. Update `run_cmd.go`

**File**: `run_cmd.go`

**Current signature**:
```go
func RunCmd(stdin io.Reader, stdout, stderr io.Writer, args []string) int
```

**New signature**:
```go
func RunCmd(stdout io.Writer, args []string) int
```

**Changes**:
1. Remove `stdin io.Reader` parameter
2. Remove `stderr io.Writer` parameter
3. Update flag set error output from `stderr` to `stdout`:
   ```go
   fs := flag.NewFlagSet("duhrpc-lint", flag.ContinueOnError)
   fs.SetOutput(stdout)  // Changed from stderr
   ```
4. Update all error messages to write to `stdout` instead of `stderr`:
   ```go
   // Before
   _, _ = fmt.Fprintln(stderr, "Error: Exactly one OpenAPI file path is required")
   _, _ = fmt.Fprintln(stderr, "Use --help for usage information")

   // After
   _, _ = fmt.Fprintln(stdout, "Error: Exactly one OpenAPI file path is required")
   _, _ = fmt.Fprintln(stdout, "Use --help for usage information")
   ```
5. Update all error messages (lines 58-59, 67):
   ```go
   // Line 58-59
   fmt.Fprintln(stdout, "Error: Exactly one OpenAPI file path is required")
   fmt.Fprintln(stdout, "Use --help for usage information")

   // Line 67
   fmt.Fprintf(stdout, "Error: %v\n", err)
   ```

**Full updated function**:
```go
func RunCmd(stdout io.Writer, args []string) int {
    fs := flag.NewFlagSet("duhrpc-lint", flag.ContinueOnError)
    fs.SetOutput(stdout)

    var showHelp bool
    var showVersion bool
    fs.BoolVar(&showHelp, "help", false, "Show help message")
    fs.BoolVar(&showVersion, "version", false, "Show version information")

    if err := fs.Parse(args); err != nil {
        return 2
    }

    if showHelp {
        _, _ = fmt.Fprint(stdout, helpText)
        return 0
    }

    if showVersion {
        _, _ = fmt.Fprintf(stdout, "duhrpc-lint version %s\n", Version)
        return 0
    }

    if fs.NArg() != 1 {
        _, _ = fmt.Fprintln(stdout, "Error: Exactly one OpenAPI file path is required")
        _, _ = fmt.Fprintln(stdout, "Use --help for usage information")
        return 2
    }

    filePath := fs.Arg(0)

    doc, err := internal.Load(filePath)
    if err != nil {
        _, _ = fmt.Fprintf(stdout, "Error: %v\n", err)
        return 2
    }

    result := internal.Validate(doc, filePath)
    internal.Print(stdout, result)

    if result.Valid() {
        return 0
    }
    return 1
}
```

#### 2. Update `cmd/duhrpc-lint/main.go`

**File**: `cmd/duhrpc-lint/main.go`

**Current**:
```go
func main() {
    os.Exit(lint.RunCmd(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}
```

**Updated**:
```go
func main() {
    os.Exit(lint.RunCmd(os.Stdout, os.Args[1:]))
}
```

#### 3. Update `run_cmd_test.go`

**File**: `run_cmd_test.go`

**Changes**: Update all test functions to use new signature.

**Before**:
```go
func TestRunCmdHelp(t *testing.T) {
    var stdout, stderr bytes.Buffer
    exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"--help"})

    assert.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "duhrpc-lint - Validate OpenAPI specs")
    assert.Contains(t, stdout.String(), "Usage:")
    assert.Contains(t, stdout.String(), "Exit Codes:")
    assert.Empty(t, stderr.String())
}
```

**After**:
```go
func TestRunCmdHelp(t *testing.T) {
    var stdout bytes.Buffer
    exitCode := lint.RunCmd(&stdout, []string{"--help"})

    assert.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "duhrpc-lint - Validate OpenAPI specs")
    assert.Contains(t, stdout.String(), "Usage:")
    assert.Contains(t, stdout.String(), "Exit Codes:")
}
```

**Update all 7 test functions**:
- `TestRunCmdHelp`
- `TestRunCmdVersion`
- `TestRunCmdValidFile`
- `TestRunCmdFileNotFound` - errors now in stdout
- `TestRunCmdInvalidYAML` - errors now in stdout
- `TestRunCmdNoArguments` - errors now in stdout
- `TestRunCmdMultipleArguments` - errors now in stdout

**Key changes for error tests**:
```go
// Before
assert.Contains(t, stderr.String(), "Error:")
assert.Empty(t, stdout.String())

// After
assert.Contains(t, stdout.String(), "Error:")
```

#### 4. Update `cmd/duhrpc-lint/integration_test.go`

**File**: `cmd/duhrpc-lint/integration_test.go`

**Changes**: Update all test functions (5 total).

**Before**:
```go
func TestIntegrationValidSpec(t *testing.T) {
    var stdout, stderr bytes.Buffer

    exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"../../testdata/valid-spec.yaml"})

    require.Equal(t, 0, exitCode)
    assert.Empty(t, stderr.String())
    assert.Contains(t, stdout.String(), "✓")
    assert.Contains(t, stdout.String(), "compliant")
}
```

**After**:
```go
func TestIntegrationValidSpec(t *testing.T) {
    var stdout bytes.Buffer

    exitCode := lint.RunCmd(&stdout, []string{"../../testdata/valid-spec.yaml"})

    require.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "✓")
    assert.Contains(t, stdout.String(), "compliant")
}
```

**Update all 5 test functions**:
- `TestIntegrationValidSpec`
- `TestIntegrationAllRuleViolations` (7 sub-tests)
- `TestIntegrationMultipleViolations`
- `TestIntegrationFileNotFound` - errors now in stdout
- `TestIntegrationInvalidYAML` - errors now in stdout

### Validation Commands

**Build**:
```bash
go build ./cmd/duhrpc-lint
```

**Run existing tests**:
```bash
go test ./...
```

**Manual verification**:
```bash
./duhrpc-lint --help
./duhrpc-lint --version
./duhrpc-lint testdata/valid-spec.yaml
./duhrpc-lint nonexistent.yaml  # Error should appear in stdout
```

### Success Criteria
- [x] `RunCmd` signature simplified to `func RunCmd(stdout io.Writer, args []string) int`
- [x] `main.go` updated
- [x] All 7 tests in `run_cmd_test.go` pass
- [x] All 5 tests in `integration_test.go` pass
- [x] `go test ./...` passes 100%
- [x] Build succeeds: `go build ./cmd/duhrpc-lint`
- [x] Manual testing shows errors in stdout

---

## Conversion Pattern (Phases 2-9)

### Helper Functions Pattern

After Phase 1 completes, create a single helper function in Phase 2:

```go
package rules_test

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/duh-rpc/duhrpc-lint"
    "github.com/stretchr/testify/assert"
)

// writeYAML writes YAML content to a temporary file and returns the path
func writeYAML(t *testing.T, yaml string) string {
    t.Helper()

    dir := t.TempDir()
    filePath := filepath.Join(dir, "spec.yaml")

    err := os.WriteFile(filePath, []byte(yaml), 0644)
    if err != nil {
        t.Fatalf("Failed to write test YAML: %v", err)
    }

    return filePath
}
```

### Test Conversion Template

**Self-Documenting Table-driven test structure**:
```go
for _, test := range []struct {
    name           string
    spec           string
    expectedExit   int
    expectedOutput string
}{
    {
        name: "ValidSpec",
        spec: `openapi: 3.0.0...`,
        expectedExit: 0,
        expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
    },
    {
        name: "InvalidSpec",
        spec: `openapi: 3.0.0...`,
        expectedExit: 1,
        expectedOutput: `[rule-name] /path/to/violation
  Error message explaining the problem
  Suggestion on how to fix it`,
    },
} {
    t.Run(test.name, func(t *testing.T) {
        filePath := writeYAML(t, test.spec)

        var stdout bytes.Buffer
        exitCode := lint.RunCmd(&stdout, []string{filePath})

        assert.Equal(t, test.expectedExit, exitCode)
        assert.Contains(t, stdout.String(), test.expectedOutput)
    })
}
```

**Key Principles**:
- **Self-documenting**: Full expected output visible in test case - no need to run linter to see format
- **Simple assertions**: Just compare exit code and check output contains expected text
- **Complete error messages**: Include the full violation output so tests document user experience

### Code Example: Before and After

**Before (Unit Test)**:
```go
package rules_test

import (
    "testing"

    "github.com/duh-rpc/duhrpc-lint/internal/rules"
    "github.com/pb33f/libopenapi"
    "github.com/stretchr/testify/require"
)

func TestPathFormatRule(t *testing.T) {
    for _, test := range []struct {
        name            string
        spec            string
        expectViolation bool
        violationCount  int
    }{
        {
            name: "ValidPathV1",
            spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
            expectViolation: false,
        },
        {
            name: "MissingVersion",
            spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
            expectViolation: true,
            violationCount:  1,
        },
    } {
        t.Run(test.name, func(t *testing.T) {
            doc, err := libopenapi.NewDocument([]byte(test.spec))
            require.NoError(t, err)

            model, errs := doc.BuildV3Model()
            require.Empty(t, errs)

            rule := rules.NewPathFormatRule()
            violations := rule.Validate(&model.Model)

            if test.expectViolation {
                require.NotEmpty(t, violations)
                require.Len(t, violations, test.violationCount)
                require.Equal(t, "path-format", violations[0].RuleName)
            } else {
                require.Empty(t, violations)
            }
        })
    }
}
```

**After (E2E Test - Self-Documenting)**:
```go
package rules_test

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/duh-rpc/duhrpc-lint"
    "github.com/stretchr/testify/assert"
)

func writeYAML(t *testing.T, yaml string) string {
    t.Helper()
    dir := t.TempDir()
    filePath := filepath.Join(dir, "spec.yaml")
    err := os.WriteFile(filePath, []byte(yaml), 0644)
    if err != nil {
        t.Fatalf("Failed to write test YAML: %v", err)
    }
    return filePath
}

func TestPathFormatRule(t *testing.T) {
    for _, test := range []struct {
        name           string
        spec           string
        expectedExit   int
        expectedOutput string
    }{
        {
            name: "ValidPathV1",
            spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
            expectedExit:   0,
            expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
        },
        {
            name: "MissingVersion",
            spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
            expectedExit: 1,
            expectedOutput: `[path-format] /users.create
  Path must start with version prefix (e.g., /v1/)
  Add a version prefix like /v1/`,
        },
    } {
        t.Run(test.name, func(t *testing.T) {
            filePath := writeYAML(t, test.spec)

            var stdout bytes.Buffer
            exitCode := lint.RunCmd(&stdout, []string{filePath})

            assert.Equal(t, test.expectedExit, exitCode)
            assert.Contains(t, stdout.String(), test.expectedOutput)
        })
    }
}
```

### Key Differences Highlighted

1. **Imports**: Removed libopenapi/rules, added bytes, os, filepath, lint
2. **Single helper**: Only `writeYAML()` needed - no assertion helpers
3. **Test struct**: Changed to `expectedExit` and `expectedOutput` fields
4. **YAML specs**: Remain complete, valid OpenAPI documents
5. **Expected output**: Full violation messages visible in test cases - self-documenting!
6. **Test execution**: File write + RunCmd instead of parsing + rule.Validate()
7. **Simple assertions**: Just exit code and output contains - no complex logic
8. **No abstraction**: Direct, readable test body - easier to understand

---

## Phase 2: path_format_test.go ✅ COMPLETE

**File**: `internal/rules/path_format_test.go`
**Lines**: 382 (after conversion)
**Test function**: `TestPathFormatRule`
**Number of test cases**: 12
**Validation command**: `go test ./internal/rules -run TestPathFormatRule -v`

### Changes Completed

1. **Updated imports** ✅:
   - Removed: `github.com/duh-rpc/duhrpc-lint/internal/rules`, `github.com/pb33f/libopenapi`, `require`
   - Added: `bytes`, `os`, `path/filepath`, `github.com/duh-rpc/duhrpc-lint`, `assert`

2. **Added writeYAML helper function** ✅ - Single helper for writing YAML to temp files

3. **Updated struct definition** ✅:
   ```go
   for _, test := range []struct {
       name           string
       spec           string
       expectedExit   int
       expectedOutput string
   }{
   ```

4. **Converted all 12 test cases** ✅ - Each includes full expected output visible in test

5. **Replaced test body** ✅:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)

       var stdout bytes.Buffer
       exitCode := lint.RunCmd(&stdout, []string{filePath})

       assert.Equal(t, test.expectedExit, exitCode)
       assert.Contains(t, stdout.String(), test.expectedOutput)
   })
   ```

### Results
- ✅ All 12 tests pass
- ✅ Tests are self-documenting - full error messages visible in test cases
- ✅ No need to run linter to understand output format

---

## Phase 3: http_method_test.go ✅ COMPLETE

**File**: `internal/rules/http_method_test.go`
**Lines**: 228 (after conversion)
**Test function**: `TestHTTPMethodRule`
**Number of test cases**: 7
**Validation command**: `go test ./internal/rules -run TestHTTPMethodRule -v`

### Changes Completed

1. **Removed duplicate writeYAML helper** ✅ - Reuses helper from Phase 2 (same package)

2. **Updated imports** ✅:
   - Removed: `github.com/duh-rpc/duhrpc-lint/internal/rules`, `github.com/pb33f/libopenapi`, `require`, `os`, `path/filepath`
   - Added: `bytes`, `github.com/duh-rpc/duhrpc-lint`, `assert`

3. **Updated struct definition** ✅:
   ```go
   for _, test := range []struct {
       name           string
       spec           string
       expectedExit   int
       expectedOutput string
   }{
   ```

4. **Converted all 7 test cases** ✅ - Each includes full expected output visible in test

5. **Replaced test body** ✅:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)

       var stdout bytes.Buffer
       exitCode := lint.RunCmd(&stdout, []string{filePath})

       assert.Equal(t, test.expectedExit, exitCode)
       assert.Contains(t, stdout.String(), test.expectedOutput)
   })
   ```

### Results
- ✅ All 7 tests pass
- ✅ Tests are self-documenting - full error messages visible in test cases
- ✅ Shares writeYAML helper with other tests in same package

---

## Phase 4: query_params_test.go ✅ COMPLETE

**File**: `internal/rules/query_params_test.go`
**Lines**: 230 (after conversion)
**Test function**: `TestQueryParamsRule`
**Number of test cases**: 6
**Validation command**: `go test ./internal/rules -run TestQueryParamsRule -v`

### Changes Completed

1. **Reused writeYAML helper** ✅ - Shares helper from Phase 2 (same package)

2. **Updated imports** ✅:
   - Removed: `github.com/duh-rpc/duhrpc-lint/internal/rules`, `github.com/pb33f/libopenapi`, `require`
   - Added: `bytes`, `github.com/duh-rpc/duhrpc-lint`, `assert`

3. **Updated struct definition** ✅:
   ```go
   for _, test := range []struct {
       name           string
       spec           string
       expectedExit   int
       expectedOutput string
   }{
   ```

4. **Converted all 6 test cases** ✅ - Each includes full expected output visible in test

5. **Replaced test body** ✅:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)

       var stdout bytes.Buffer
       exitCode := lint.RunCmd(&stdout, []string{filePath})

       assert.Equal(t, test.expectedExit, exitCode)
       assert.Contains(t, stdout.String(), test.expectedOutput)
   })
   ```

### Results
- ✅ All 6 tests pass
- ✅ Tests are self-documenting - full error messages visible in test cases
- ✅ Shares writeYAML helper with other tests in same package

---

## Phase 5: request_body_test.go ✅ COMPLETE

**File**: `internal/rules/request_body_test.go`
**Lines**: 129 (after conversion)
**Test function**: `TestRequestBodyRule`
**Number of test cases**: 4
**Validation command**: `go test ./internal/rules -run TestRequestBodyRule -v`

### Changes Completed

1. **Reused writeYAML helper** ✅ - Shares helper from Phase 2 (same package)

2. **Updated imports** ✅:
   - Removed: `github.com/duh-rpc/duhrpc-lint/internal/rules`, `github.com/pb33f/libopenapi`, `require`
   - Added: `bytes`, `github.com/duh-rpc/duhrpc-lint`, `assert`

3. **Updated struct definition** ✅:
   ```go
   for _, test := range []struct {
       name           string
       spec           string
       expectedExit   int
       expectedOutput string
   }{
   ```

4. **Converted all 4 test cases** ✅ - Each includes full expected output visible in test

5. **Replaced test body** ✅:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)

       var stdout bytes.Buffer
       exitCode := lint.RunCmd(&stdout, []string{filePath})

       assert.Equal(t, test.expectedExit, exitCode)
       assert.Contains(t, stdout.String(), test.expectedOutput)
   })
   ```

### Results
- ✅ All 4 tests pass
- ✅ Tests are self-documenting - full error messages visible in test cases
- ✅ Shares writeYAML helper with other tests in same package

---

## Phase 6: status_code_test.go ✅ COMPLETE

**File**: `internal/rules/status_code_test.go`
**Lines**: 302 (after conversion)
**Test function**: `TestStatusCodeRule`
**Number of test cases**: 8
**Validation command**: `go test ./internal/rules -run TestStatusCodeRule -v`

### Changes Completed

1. **Reused writeYAML helper** ✅ - Shares helper from Phase 2 (same package)

2. **Updated imports** ✅:
   - Removed: `github.com/duh-rpc/duhrpc-lint/internal/rules`, `github.com/pb33f/libopenapi`, `require`
   - Added: `bytes`, `github.com/duh-rpc/duhrpc-lint`, `assert`

3. **Updated struct definition** ✅:
   ```go
   for _, test := range []struct {
       name           string
       spec           string
       expectedExit   int
       expectedOutput string
   }{
   ```

4. **Converted all 8 test cases** ✅ - Each includes full expected output visible in test

5. **Replaced test body** ✅:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)

       var stdout bytes.Buffer
       exitCode := lint.RunCmd(&stdout, []string{filePath})

       assert.Equal(t, test.expectedExit, exitCode)
       assert.Contains(t, stdout.String(), test.expectedOutput)
   })
   ```

### Results
- ✅ All 8 tests pass (note: original plan said 7, but there are actually 8 test cases)
- ✅ Tests are self-documenting - full error messages visible in test cases
- ✅ Shares writeYAML helper with other tests in same package
- ✅ Fixed error response schemas to use integer type for code field

---

## Phase 7: success_response_test.go ✅ COMPLETE

**File**: `internal/rules/success_response_test.go`
**Lines**: 163
**Test function**: `TestSuccessResponseRule`
**Number of test cases**: 6
**Validation command**: `go test ./internal/rules -run TestSuccessResponseRule -v`

### Changes Required

1. **Copy helper functions** from Phase 2

2. **Update imports**: Same as Phase 2

3. **Convert struct definition**:
   ```go
   for _, test := range []struct {
       name          string
       spec          string
       wantViolation bool
   }{
   ```

4. **Replace test body**:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)
       exitCode, stdout := runLint(t, filePath)

       if test.wantViolation {
           assertViolation(t, exitCode, stdout, "success-response")
       } else {
           assertNoViolations(t, exitCode, stdout)
       }
   })
   ```

---

## Phase 8: content_type_test.go ✅ COMPLETE

**File**: `internal/rules/content_type_test.go`
**Lines**: 308
**Test function**: `TestContentTypeRule`
**Number of test cases**: 10
**Validation command**: `go test ./internal/rules -run TestContentTypeRule -v`

### Changes Required

1. **Copy helper functions** from Phase 2

2. **Update imports**: Same as Phase 2

3. **Convert struct definition**:
   ```go
   for _, test := range []struct {
       name          string
       spec          string
       wantViolation bool
   }{
   ```

4. **Replace test body**:
   ```go
   t.Run(test.name, func(t *testing.T) {
       filePath := writeYAML(t, test.spec)
       exitCode, stdout := runLint(t, filePath)

       if test.wantViolation {
           assertViolation(t, exitCode, stdout, "content-type")
       } else {
           assertNoViolations(t, exitCode, stdout)
       }
   })
   ```

---

## Phase 9: error_response_test.go ✅ COMPLETE

**File**: `internal/rules/error_response_test.go`
**Lines**: 545
**Test functions**: `TestErrorResponseRule`, `TestErrorResponseRuleWithRef`, `TestErrorResponseRuleWithAllOf`
**Number of test cases**: 10+ (across 3 test functions)
**Validation command**: `go test ./internal/rules -run TestErrorResponse -v`

### Changes Required

1. **Copy helper functions** from Phase 2

2. **Update imports**: Same as Phase 2

3. **Convert all 3 test functions**:
   - `TestErrorResponseRule` - main table-driven test
   - `TestErrorResponseRuleWithRef` - single test with $ref
   - `TestErrorResponseRuleWithAllOf` - single test with allOf

4. **For table-driven test**:
   ```go
   for _, test := range []struct {
       name          string
       spec          string
       wantViolation bool
   }{
   ```

5. **For single tests** (WithRef, WithAllOf):
   ```go
   func TestErrorResponseRuleWithRef(t *testing.T) {
       spec := `...`  // Complete YAML

       filePath := writeYAML(t, spec)
       exitCode, stdout := runLint(t, filePath)

       assertNoViolations(t, exitCode, stdout)
   }
   ```

---

## Validation Strategy

### Per-Phase Validation
After completing each phase:

1. **Run the specific test**:
   ```bash
   go test ./internal/rules -run [TestName] -v
   ```

2. **Verify test output**:
   - All tests should PASS
   - No compilation errors
   - No import errors

### Intermediate Checkpoints

**After Phase 1** (RunCmd simplification):
```bash
go test ./...
```

**After Phase 4** (3 test files converted):
```bash
go test ./internal/rules -v
```

**After Phase 7** (6 test files converted):
```bash
go test ./internal/rules -v
```

### Final Validation

After all phases complete:

1. **Run all rule tests**:
   ```bash
   go test ./internal/rules -v
   ```

2. **Run entire test suite**:
   ```bash
   go test ./... -v
   ```

3. **Build binary**:
   ```bash
   go build ./cmd/duhrpc-lint
   ```

---

## Success Criteria

### Completion Checklist

- [x] **Phase 1**: RunCmd signature simplified
- [x] **Phase 1**: main.go updated
- [x] **Phase 1**: run_cmd_test.go updated (7 tests)
- [x] **Phase 1**: integration_test.go updated (5 tests)
- [x] **Phases 2-9**: All 8 test files converted
- [x] All helper functions implemented and reused
- [x] All imports updated correctly
- [x] All test cases preserved (no test scenarios lost)
- [x] `go test ./internal/rules` passes 100%
- [x] `go test ./...` passes 100%
- [x] No direct calls to `rule.Validate()` remain
- [x] No inline YAML parsing with libopenapi remains
- [x] All tests use `t.TempDir()` for file creation
- [x] All tests use `lint.RunCmd()` for execution
- [x] Exit codes verified (0 for pass, 1 for violations)
- [x] Violation markers verified (e.g., `[path-format]`)
- [x] No regression in test coverage

### Quality Criteria

1. **Code Quality**:
   - Helper functions are DRY and reusable
   - Test names are descriptive
   - Test cases are well-organized

2. **Test Reliability**:
   - Tests are deterministic (no flaky tests)
   - Tests clean up after themselves (t.TempDir() auto-cleanup)
   - Tests are isolated (no shared state)

3. **Performance**:
   - Tests run quickly (no unnecessary delays)
   - File I/O is minimized
   - Parallel test execution works correctly

---

## Implementation Notes

### Best Practices

1. **Keep YAML specs complete**: Ensure test YAML includes all required fields for a valid OpenAPI spec

2. **Use t.Helper()**: Mark helper functions with `t.Helper()` so test failures point to the actual test case

3. **Consistent naming**: Use consistent field names across test structs:
   - `name` - test case name
   - `spec` - YAML specification
   - `wantViolation` - boolean indicating if violation is expected

4. **Reuse helpers**: Copy the 4 helper functions to each test file for independence

5. **Test isolation**: Each test case should be completely independent

### Common Pitfalls to Avoid

1. **Incomplete YAML specs**: Don't forget required OpenAPI fields (info, paths, responses, etc.)

2. **Hardcoded paths**: Use `t.TempDir()` and `filepath.Join()` for cross-platform compatibility

3. **Forgetting t.Helper()**: Without it, test failures point to helper functions instead of actual test cases

4. **Not checking all output**: Verify both exit code AND output content

5. **Overly specific assertions**: Don't check exact output format - just verify key markers are present

---

## Summary

This plan provides a comprehensive roadmap for:

1. **Phase 1**: Simplifying the CLI interface (RunCmd signature)
2. **Phases 2-9**: Converting 8 unit test files to functional e2e tests

Each phase is independently implementable and testable. The helper functions established in Phase 2 are reused across all subsequent phases, ensuring consistency and reducing code duplication.

**Total estimated effort**: 9 phases

**Final deliverable**: Clean, functional e2e tests that verify the entire linter pipeline end-to-end

---

END OF IMPLEMENTATION PLAN
