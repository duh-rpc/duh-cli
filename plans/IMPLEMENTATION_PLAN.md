# DUH-RPC Linter Implementation Plan

## Overview

This plan guides the implementation of `duhrpc-lint`, a command-line tool that validates OpenAPI 3.0 YAML specifications for compliance with DUH-RPC conventions. The tool will parse OpenAPI files using `pb33f.io/libopenapi`, validate against 8 DUH-RPC rules, and report violations with actionable suggestions.

**Source:** This plan is based on the comprehensive technical specification in `docs/TECHNICAL_SPEC.md`.

**Approach:** Test-focused development with incremental feature delivery across 5 distinct phases. Each phase delivers a complete, working increment with clear validation criteria.

## Current State Analysis

### What Exists
- **Technical Specification:** Complete 2,442-line specification defining all requirements, architecture, and validation rules
- **Project Guidelines:** `CLAUDE.md` with testing patterns and code conventions
- **Empty Project Directory:** Ready for implementation

### What's Missing
- Go module initialization
- Directory structure (`cmd/`, `internal/`, `testdata/`)
- All source code and tests
- All 8 validation rules
- CLI infrastructure
- Documentation (README, usage examples)

### Key Discoveries
From the technical specification:
- **Library Choice:** `pb33f.io/libopenapi` v0.18+ for OpenAPI parsing (proven, high-performance)
- **Architecture Pattern:** Rule-based validation with clear separation of concerns
- **CLI Pattern:** Use `run(stdin, stdout, stderr, args)` pattern for testability (reference: rebranch.go pattern)
- **Exit Codes:** 0 (success), 1 (violations), 2 (errors)
- **Test Strategy:** Tests in `package XXX_test`, table-driven, using require/assert from testify

## Desired End State

A production-ready CLI tool that:
- Accepts OpenAPI 3.0 YAML file as positional argument
- Validates against all 8 DUH-RPC rules
- Reports all violations in human-readable format with suggestions
- Returns appropriate exit codes for CI/CD integration
- Includes comprehensive test coverage
- Provides clear documentation and examples

### Validation Criteria

**Functional Validation:**
All tests call `lint.RunCmd()` directly with io.Reader/Writer parameters. Examples below show manual CLI usage for reference:

```bash
# Valid spec returns success
./duhrpc-lint testdata/valid-spec.yaml
# Output: ✓ valid-spec.yaml is DUH-RPC compliant
# Exit code: 0

# Invalid spec reports violations
./duhrpc-lint testdata/invalid-spec.yaml
# Output: Lists all violations with suggestions
# Exit code: 1

# Missing file returns error
./duhrpc-lint missing.yaml
# Output: Error: File not found: missing.yaml
# Exit code: 2
```

**Test Validation:**
```bash
# All tests pass
go test ./...
# Exit code: 0

# Build succeeds
go build ./cmd/duhrpc-lint
# Exit code: 0
```

## What We're NOT Doing

Explicitly out of scope for this implementation:
- JSON input support (YAML only)
- OpenAPI 3.1 or 2.0 support (3.0 only)
- Machine-readable JSON output format
- Configuration files or rule customization
- Auto-fix mode
- Quiet mode flag
- IDE integration
- Multiple file validation
- Watch mode

## Implementation Approach

### High-Level Strategy

**5-Phase Delivery:**
1. **Foundation** - Build CLI infrastructure and file loading (no rules)
2. **Simple Rules Batch 1** - Implement 3 straightforward validation rules
3. **Simple Rules Batch 2** - Implement 3 more validation rules
4. **Complex Rules** - Implement 2 complex rules requiring schema resolution
5. **Integration & Documentation** - End-to-end tests, docs, polish

**Key Principles:**
- Each phase delivers working, testable functionality
- Tests validate behavior, not implementation
- Incremental test data creation (add as rules are implemented)
- CLI designed for testability using `run()` pattern
- Follow CLAUDE.md guidelines strictly

### Technology Decisions

- **Language:** Go 1.21+
- **Dependencies:**
  - `github.com/pb33f/libopenapi` v0.18+ (OpenAPI parsing)
  - `github.com/stretchr/testify` (testing assertions)
  - Go standard library for CLI (`flag`, `os`, `fmt`)
- **Module Path:** `github.com/duh-rpc/duhrpc-lint`

---

## Phase 1: Foundation

### Overview
Establish the core CLI infrastructure, file loading, output formatting, and validation framework. This phase creates a working CLI that can load OpenAPI files and report results, but performs no validation yet.

**Goal:** A testable CLI skeleton that loads files and outputs formatted results.

**Validation Criteria:**
- `go test ./...` passes all foundation tests
- `go build ./cmd/duhrpc-lint` succeeds
- CLI can load valid OpenAPI file without crashing
- CLI reports file not found errors correctly
- CLI reports parse errors correctly
- Help and version flags work

---

### Changes Required

#### 1. Project Initialization

**Files to Create:**
- `go.mod`
- `go.sum` (generated)
- `.gitignore`

**Commands:**
```bash
go mod init github.com/duh-rpc/duhrpc-lint
go get github.com/pb33f/libopenapi@latest
go get github.com/stretchr/testify@latest
```

**Validation:**
```bash
go mod tidy
go mod verify
```

---

#### 2. Core Types (`internal/types.go`)

**File:** `internal/types.go`

**Changes:** Define core data structures for violations and validation results.

```go
package internal

import "github.com/pb33f/libopenapi/datamodel/high/v3"

// Violation represents a single DUH-RPC compliance violation
type Violation struct {
    RuleName   string
    Location   string
    Message    string
    Suggestion string
}

// String formats violation for display
func (v Violation) String() string

// ValidationResult contains all violations found in a document
type ValidationResult struct {
    Violations []Violation
    FilePath   string
}

// Valid returns true if no violations found
func (vr ValidationResult) Valid() bool

// Rule interface that all validation rules must implement
type Rule interface {
    Name() string
    Validate(doc *v3.Document) []Violation
}
```

**Function Responsibilities:**
- `Violation.String()`: Format violation as multi-line string with `[rule-name] location\n  message\n  suggestion`
- `ValidationResult.Valid()`: Return `len(vr.Violations) == 0`

**Testing Requirements:**

**File:** `internal/types_test.go`

```go
package internal_test

func TestViolationString(t *testing.T)
func TestValidationResultValid(t *testing.T)
```

**Test Objectives:**
- Verify `Violation.String()` formats correctly with all fields
- Verify `ValidationResult.Valid()` returns true when no violations
- Verify `ValidationResult.Valid()` returns false when violations exist

**Context for Implementation:**
- Follow CLAUDE.md struct field ordering (visual tapering)
- Tests in `internal_test` package, not `internal` package
- Use table-driven tests for multiple violation scenarios

---

#### 3. File Loader (`internal/loader.go`)

**File:** `internal/loader.go`

**Changes:** Implement OpenAPI file loading and parsing.

```go
package internal

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Load reads and parses an OpenAPI 3.0 YAML file
func Load(filePath string) (*v3.Document, error)
```

**Function Responsibilities:**
- Check file exists using `os.Stat()`
- Read file contents with `os.ReadFile()`
- Parse YAML with `libopenapi.NewDocument()`
- Build v3 model with `doc.BuildV3Model()`
- Return parsed document or descriptive error
- Error messages: `"File not found: %s"`, `"Failed to parse OpenAPI spec: %w"`

**Testing Requirements:**

**File:** `internal/loader_test.go`

```go
package internal_test

func TestLoadValidFile(t *testing.T)
func TestLoadFileNotFound(t *testing.T)
func TestLoadInvalidYAML(t *testing.T)
```

**Test Objectives:**
- Verify loading valid OpenAPI 3.0 YAML returns document without error
- Verify loading non-existent file returns "File not found" error
- Verify loading invalid YAML syntax returns parse error
- Create minimal test fixture: `testdata/minimal-valid.yaml` (single path, POST, valid structure)
- Create invalid fixture: `testdata/invalid-syntax.yaml` (malformed YAML)

**Context for Implementation:**
- Use `require.NoError()` for critical assertions
- Use `require.ErrorContains()` for error message validation
- Reference: Technical spec section "loader/loader.go" (lines 1174-1211)
- **Edge Case:** Empty paths object (`paths: {}`) is valid OpenAPI - allow it (report compliant)

---

#### 4. Reporter (`internal/reporter.go`)

**File:** `internal/reporter.go`

**Changes:** Implement output formatting for validation results.

```go
package internal

import (
    "io"
)

// Print formats and outputs validation results
func Print(w io.Writer, result ValidationResult)
```

**Function Responsibilities:**
- If `result.Valid()`: print success message `✓ {filename} is DUH-RPC compliant`
- If violations exist:
  - Print header: `Validating {filename}...`
  - Print `ERRORS FOUND:\n`
  - Print each violation using `violation.String()` in discovery order (no sorting/grouping)
  - Print summary: `Summary: N violations found in {filename}`
- Use `filepath.Base()` to extract filename from path
- Write to provided `io.Writer` (for testability)

**Testing Requirements:**

**File:** `internal/reporter_test.go`

```go
package internal_test

func TestPrintValidResult(t *testing.T)
func TestPrintWithViolations(t *testing.T)
func TestPrintMultipleViolations(t *testing.T)
```

**Test Objectives:**
- Verify success output contains checkmark and "compliant" message
- Verify violation output contains header, violations, and summary
- Verify output format matches specification exactly
- Test with `bytes.Buffer` as writer to capture output

**Context for Implementation:**
- Accept `io.Writer` parameter for testability
- Reference: Technical spec section "reporter/reporter.go" (lines 1297-1325)
- Follow REQ-010 output format specification (lines 869-929)

---

#### 5. Validator Framework (`internal/validator.go`)

**File:** `internal/validator.go`

**Changes:** Create validator that orchestrates rule execution (no rules yet).

```go
package internal

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Validate runs all registered rules against the document
func Validate(doc *v3.Document, filePath string) ValidationResult
```

**Function Responsibilities:**
- Create empty slice of rules (will be populated in later phases)
- Iterate through rules, calling `rule.Validate(doc)`
- Collect all violations from all rules
- Return `ValidationResult` with violations and file path

**Testing Requirements:**

**File:** `internal/validator_test.go`

```go
package internal_test

func TestValidateEmptyRules(t *testing.T)
```

**Test Objectives:**
- Verify validator with no rules returns empty violations
- Verify `ValidationResult.FilePath` is set correctly
- Create mock document for testing (minimal valid OpenAPI structure)

**Context for Implementation:**
- In Phase 1, rules slice is empty: `var rules []Rule`
- Later phases will add rule instantiation here
- Reference: Technical spec section "validator/validator.go" (lines 1215-1250)

---

#### 6. CLI Command (`run_cmd.go`)

**File:** `run_cmd.go` (root package)

**Changes:** Create testable `RunCmd()` function following rebranch pattern.

```go
package lint

import (
    "flag"
    "io"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

const Version = "1.0.0"

// RunCmd executes the CLI logic and returns exit code
func RunCmd(stdin io.Reader, stdout, stderr io.Writer, args []string) int
```

**Function Responsibilities:**

`RunCmd()` function:
- Parse flags: `--help`, `--version`
- If `--help`: print usage to stdout, return 0
- If `--version`: print version to stdout, return 0
- Validate exactly one positional argument (file path)
- Call `internal.Load(filePath)` - if error, print to stderr, return 2
- Call `internal.Validate(doc, filePath)` - returns result
- Call `internal.Print(stdout, result)` - outputs results
- Return 0 if valid, 1 if violations

**Help Text Format:**
```
duhrpc-lint - Validate OpenAPI specs for DUH-RPC compliance

Usage:
  duhrpc-lint <openapi-file>
  duhrpc-lint --help
  duhrpc-lint --version

Arguments:
  <openapi-file>    Path to OpenAPI 3.0 YAML file

Options:
  --help            Show this help message
  --version         Show version information

Exit Codes:
  0    Validation passed (spec is DUH-RPC compliant)
  1    Validation failed (violations found)
  2    Error (file not found, parse error, etc.)
```

**Testing Requirements:**

**File:** `run_cmd_test.go`

```go
package lint_test

func TestRunCmdHelp(t *testing.T)
func TestRunCmdVersion(t *testing.T)
func TestRunCmdValidFile(t *testing.T)
func TestRunCmdFileNotFound(t *testing.T)
func TestRunCmdInvalidYAML(t *testing.T)
```

**Test Objectives:**
- Verify `--help` prints usage and returns 0
- Verify `--version` prints version and returns 0
- Verify valid file loads without error and returns 0 (no violations yet)
- Verify missing file prints error to stderr and returns 2
- Verify invalid YAML prints error to stderr and returns 2
- Use `bytes.Buffer` for stdin/stdout/stderr capture
- All tests call `lint.RunCmd()` directly

**Context for Implementation:**
- Pattern reference: https://github.com/derrick-wippler-anchor/rebranch/blob/2c96e794222d76bdc583e735bac599a70dfe768f/rebranch.go#L38
- Use standard library `flag` package, not cobra
- All I/O through parameters for testability
- Package name is `lint`, tests are in `lint_test`

---

#### 7. CLI Entry Point (`cmd/duhrpc-lint/main.go`)

**File:** `cmd/duhrpc-lint/main.go`

**Changes:** Minimal main that delegates to RunCmd.

```go
package main

import (
    "os"
    "github.com/duh-rpc/duhrpc-lint"
)

func main() {
    os.Exit(lint.RunCmd(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}
```

**Function Responsibilities:**
- Import the `lint` package
- Call `lint.RunCmd()` with standard streams and arguments
- Exit with returned code

**Context for Implementation:**
- This file has no tests (logic is in RunCmd)
- Follows rebranch pattern for testability
- Main package is just a thin wrapper

---

#### 8. Test Data

**Files to Create:**
- `testdata/minimal-valid.yaml` - Minimal valid OpenAPI 3.0 spec (used by loader tests)
- `testdata/invalid-syntax.yaml` - Malformed YAML (used by loader tests)

**Content:** `testdata/minimal-valid.yaml`
```yaml
openapi: 3.0.0
info:
  title: Minimal Test API
  version: 1.0.0
paths:
  /v1/test.ping:
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
                type: object
```

**Content:** `testdata/invalid-syntax.yaml`
```yaml
openapi: 3.0.0
info:
  title: Invalid
  invalid yaml syntax here: [unclosed bracket
```

---

### Phase 1 Validation Commands

**Build:**
```bash
go build ./cmd/duhrpc-lint
```

**Unit Tests:**
```bash
go test ./internal
go test .  # Tests run_cmd_test.go in root package
```

**All Tests:**
```bash
go test ./...
```

**Integration Test:**
```bash
./duhrpc-lint --help
./duhrpc-lint --version
./duhrpc-lint testdata/minimal-valid.yaml  # Should succeed with "compliant"
./duhrpc-lint nonexistent.yaml             # Should fail with "File not found"
```

---

## Phase 2: Simple Rules Batch 1

### Overview
Implement the first batch of validation rules: path format, HTTP method, and query parameters. These rules check structural compliance without complex schema traversal.

**Goal:** CLI validates 3 fundamental DUH-RPC rules and reports violations.

**Validation Criteria:**
- `go test ./internal/rules` passes all rule tests
- CLI correctly identifies and reports violations for:
  - Invalid path formats
  - Non-POST HTTP methods
  - Query parameters
- End-to-end test with spec containing these violations succeeds

---

### Changes Required

#### 1. Rule Interface Utilities (`internal/rules/rule.go`)

**File:** `internal/rules/rule.go`

**Changes:** Create shared rule utilities (the interface is already in `types`).

```go
package rules

// This file can contain shared utilities for rules if needed
// The Rule interface itself is defined in internal/types.go
```

**Context for Implementation:**
- This file may be minimal or empty in Phase 2
- Future phases might add shared helpers (e.g., $ref resolution utilities)
- Keep it simple for now

**Rule Error Handling Strategy:**
- Rules should be defensive: check for nil pointers before dereferencing
- If a required field is unexpectedly nil, skip that path/operation (don't report violation)
- Assumption: libopenapi has already validated basic OpenAPI structure
- Rules report violations for DUH-RPC compliance, not malformed OpenAPI

---

#### 2. Path Format Rule (`internal/rules/path_format.go`)

**File:** `internal/rules/path_format.go`

**Changes:** Implement REQ-002 - Path format validation.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// PathFormatRule validates DUH-RPC path format
type PathFormatRule struct{}

// NewPathFormatRule creates a new path format rule
func NewPathFormatRule() *PathFormatRule

func (r *PathFormatRule) Name() string

func (r *PathFormatRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"path-format"`
- `Validate()`:
  - Iterate through all paths in `doc.Paths.PathItems`
  - Check each path against regex: `^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$`
  - For violations, determine specific reason (missing version, wrong format, etc.)
  - Generate helpful suggestions
  - Return slice of violations

**Validation Rules (from REQ-002):**
- Path must start with `/v{N}/` where N is non-negative integer
- Subject and method segments must be lowercase alphanumeric, hyphens, underscores
- Must start with letter, 1-50 chars each
- Separated by exactly one dot
- No path parameters allowed:
  - Detect by regex: `\{[^}]+\}` in path string
  - Also check PathItem.Parameters for `in: "path"` entries (defensive validation)
  - Report violation if either found

**Testing Requirements:**

**File:** `internal/rules/path_format_test.go`

```go
package rules_test

func TestPathFormatRule(t *testing.T)
```

**Test Objectives:**
- Verify valid paths pass: `/v1/users.create`, `/v0/beta.test`, `/v10/user-accounts.get-by-id`
- Verify violations detected:
  - Missing version: `/users.create`
  - Invalid version: `/v1.2/users.create`, `/vbeta/users.create`
  - Uppercase: `/v1/Users.create`
  - Missing dot: `/v1/users`
  - Path parameters: `/v1/users/{id}.get`
  - Invalid characters: `/v1/user$accounts.create`
- Verify violation messages are descriptive
- Use table-driven tests with 10+ test cases

**Context for Implementation:**
- Reference: Technical spec REQ-002 (lines 316-374)
- Reference: Test example (lines 1620-1689)
- Compile regex once, reuse for all paths
- Provide specific error messages based on which part fails

---

#### 3. HTTP Method Rule (`internal/rules/http_method.go`)

**File:** `internal/rules/http_method.go`

**Changes:** Implement REQ-003 - POST-only enforcement.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// HTTPMethodRule validates only POST is used
type HTTPMethodRule struct{}

func NewHTTPMethodRule() *HTTPMethodRule

func (r *HTTPMethodRule) Name() string

func (r *HTTPMethodRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"http-method"`
- `Validate()`:
  - Iterate through all paths and their operations
  - Check which HTTP methods are defined (Get, Put, Post, Delete, Options, Head, Patch, Trace)
  - Report violation for any method that is not Post
  - Format location as `"{METHOD} {path}"`

**Testing Requirements:**

**File:** `internal/rules/http_method_test.go`

```go
package rules_test

func TestHTTPMethodRule(t *testing.T)
```

**Test Objectives:**
- Verify POST operations pass validation
- Verify GET, PUT, DELETE, PATCH operations are reported
- Verify multiple non-POST methods on same path are all reported
- Verify violation message suggests changing to POST

**Context for Implementation:**
- Reference: Technical spec REQ-003 (lines 377-415)
- libopenapi PathItem has fields: Get, Put, Post, Delete, Options, Head, Patch, Trace
- Check each field; if non-nil and not Post, it's a violation

---

#### 4. Query Parameters Rule (`internal/rules/query_params.go`)

**File:** `internal/rules/query_params.go`

**Changes:** Implement REQ-004 - No query parameters.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// QueryParamsRule validates no query parameters are used
type QueryParamsRule struct{}

func NewQueryParamsRule() *QueryParamsRule

func (r *QueryParamsRule) Name() string

func (r *QueryParamsRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"query-parameters"`
- `Validate()`:
  - Iterate through all paths and operations
  - Check operation `Parameters` array
  - Find parameters with `In == "query"`
  - Report each query parameter found
  - Suggestion: `"Move \"{paramName}\" to request body"`

**Testing Requirements:**

**File:** `internal/rules/query_params_test.go`

```go
package rules_test

func TestQueryParamsRule(t *testing.T)
```

**Test Objectives:**
- Verify operations with no parameters pass
- Verify header parameters are allowed (not violations)
- Verify cookie parameters are allowed
- Verify query parameters are reported with parameter name
- Verify multiple query parameters on same operation are all reported

**Context for Implementation:**
- Reference: Technical spec REQ-004 (lines 418-457)
- Check `parameter.In` field for value `"query"`
- Header and cookie parameters are allowed

---

#### 5. Wire Rules into Validator

**File:** `internal/validator.go`

**Changes:** Register the 3 new rules.

```go
func Validate(doc *v3.Document, filePath string) ValidationResult {
    allRules := []Rule{
        rules.NewPathFormatRule(),
        rules.NewHTTPMethodRule(),
        rules.NewQueryParamsRule(),
    }

    var violations []Violation
    for _, rule := range allRules {
        ruleViolations := rule.Validate(doc)
        violations = append(violations, ruleViolations...)
    }

    return ValidationResult{
        Violations: violations,
        FilePath:   filePath,
    }
}
```

**Testing Requirements:**

Update `internal/validator_test.go`:

```go
func TestValidateWithRules(t *testing.T)
```

**Test Objectives:**
- Verify validator calls all registered rules
- Verify violations from multiple rules are combined
- Use mock OpenAPI document with known violations

---

#### 6. Test Data

**Files to Create:**
- `testdata/invalid-specs/bad-path-format.yaml` - Violations of REQ-002
- `testdata/invalid-specs/wrong-http-method.yaml` - Violations of REQ-003
- `testdata/invalid-specs/has-query-params.yaml` - Violations of REQ-004

**Content:** `testdata/invalid-specs/bad-path-format.yaml`
```yaml
openapi: 3.0.0
info:
  title: Bad Path Format
  version: 1.0.0
paths:
  /users.create:  # Missing version
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
                type: object
  /v1/Users.get:  # Uppercase
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
                type: object
```

**Content:** `testdata/invalid-specs/wrong-http-method.yaml`
```yaml
openapi: 3.0.0
info:
  title: Wrong HTTP Methods
  version: 1.0.0
paths:
  /v1/users.list:
    get:  # Should be POST
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: array
  /v1/users.update:
    put:  # Should be POST
      requestBody:
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
                type: object
```

**Content:** `testdata/invalid-specs/has-query-params.yaml`
```yaml
openapi: 3.0.0
info:
  title: Query Parameters
  version: 1.0.0
paths:
  /v1/users.search:
    post:
      parameters:
        - name: query
          in: query  # Not allowed
          schema:
            type: string
        - name: limit
          in: query  # Not allowed
          schema:
            type: integer
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
                type: object
```

---

### Phase 2 Validation Commands

**Unit Tests:**
```bash
go test ./internal/rules
go test ./internal
```

**All Tests:**
```bash
go test ./...
```

**Integration Tests:**
```bash
./duhrpc-lint testdata/invalid-specs/bad-path-format.yaml
# Should report path format violations

./duhrpc-lint testdata/invalid-specs/wrong-http-method.yaml
# Should report http-method violations

./duhrpc-lint testdata/invalid-specs/has-query-params.yaml
# Should report query-parameters violations
```

**Verify Exit Codes:**
```bash
./duhrpc-lint testdata/minimal-valid.yaml
echo $?  # Should be 0 (but may have violations from other rules not yet implemented)

./duhrpc-lint testdata/invalid-specs/bad-path-format.yaml
echo $?  # Should be 1
```

---

## Phase 3: Simple Rules Batch 2

### Overview
Implement the second batch of validation rules: request body requirements, status code restrictions, and success response requirements. These rules validate operation structure and response definitions.

**Goal:** CLI validates 6 of 8 DUH-RPC rules (3 from Phase 2 + 3 new).

**Validation Criteria:**
- `go test ./internal/rules` passes all tests
- CLI correctly identifies and reports violations for:
  - Missing or non-required request bodies
  - Invalid status codes
  - Missing or incomplete success responses
- End-to-end test with spec containing these violations succeeds

---

### Changes Required

#### 1. Request Body Rule (`internal/rules/request_body.go`)

**File:** `internal/rules/request_body.go`

**Changes:** Implement REQ-005 - Required request body validation.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// RequestBodyRule validates all operations have required request body
type RequestBodyRule struct{}

func NewRequestBodyRule() *RequestBodyRule

func (r *RequestBodyRule) Name() string

func (r *RequestBodyRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"request-body-required"`
- `Validate()`:
  - Iterate through all paths and operations
  - Check if `RequestBody` is nil - violation: "No request body defined"
  - Check if `RequestBody.Required` is not true - violation: "required: false"
  - Both checks necessary for complete validation

**Testing Requirements:**

**File:** `internal/rules/request_body_test.go`

```go
package rules_test

func TestRequestBodyRule(t *testing.T)
```

**Test Objectives:**
- Verify operations with `requestBody.required: true` pass
- Verify operations missing requestBody are reported
- Verify operations with `requestBody.required: false` are reported
- Verify violation messages distinguish between missing vs not required

**Context for Implementation:**
- Reference: Technical spec REQ-005 (lines 462-504)
- Check both nil and Required field
- libopenapi RequestBody has `Required *bool` field

---

#### 2. Status Code Rule (`internal/rules/status_code.go`)

**File:** `internal/rules/status_code.go`

**Changes:** Implement REQ-008 - Allowed status codes only.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// StatusCodeRule validates only allowed HTTP status codes are used
type StatusCodeRule struct{}

func NewStatusCodeRule() *StatusCodeRule

func (r *StatusCodeRule) Name() string

func (r *StatusCodeRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"status-code"`
- `Validate()`:
  - Define allowed codes: `200, 400, 401, 403, 404, 429, 452, 453, 454, 455, 500`
  - Iterate through all paths, operations, and response status codes
  - Check if status code is in allowed list
  - Report violations with code and path

**Testing Requirements:**

**File:** `internal/rules/status_code_test.go`

```go
package rules_test

func TestStatusCodeRule(t *testing.T)
```

**Test Objectives:**
- Verify allowed status codes pass: 200, 400, 401, 403, 404, 429, 452-455, 500
- Verify disallowed codes are reported: 201, 202, 204, 405, 503
- Verify violation message includes status code and allowed list
- Use table-driven test for multiple status codes

**Context for Implementation:**
- Reference: Technical spec REQ-008 (lines 729-781)
- Create const slice of allowed codes
- libopenapi Responses is map[string]*Response where key is status code string

---

#### 3. Success Response Rule (`internal/rules/success_response.go`)

**File:** `internal/rules/success_response.go`

**Changes:** Implement REQ-009 - 200 response required.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// SuccessResponseRule validates 200 response exists with content
type SuccessResponseRule struct{}

func NewSuccessResponseRule() *SuccessResponseRule

func (r *SuccessResponseRule) Name() string

func (r *SuccessResponseRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"success-response"`
- `Validate()`:
  - Iterate through all paths and operations
  - Check if `responses["200"]` exists - if not, violation
  - Check if 200 response has `Content` - if nil or empty, violation
  - Check if content has at least one media type with schema - if not, violation
  - No restrictions on schema structure (any schema valid)

**Testing Requirements:**

**File:** `internal/rules/success_response_test.go`

```go
package rules_test

func TestSuccessResponseRule(t *testing.T)
```

**Test Objectives:**
- Verify operations with 200 response and content pass
- Verify missing 200 response is reported
- Verify 200 response without content is reported
- Verify 200 response with content but no schema is reported
- Verify various schema types are allowed (object, array, string)

**Context for Implementation:**
- Reference: Technical spec REQ-009 (lines 786-865)
- 200 response must exist and have content
- No validation of schema structure (any valid schema OK)

---

#### 4. Wire Rules into Validator

**File:** `internal/validator.go`

**Changes:** Add 3 new rules to the list.

```go
func Validate(doc *v3.Document, filePath string) ValidationResult {
    allRules := []Rule{
        rules.NewPathFormatRule(),
        rules.NewHTTPMethodRule(),
        rules.NewQueryParamsRule(),
        rules.NewRequestBodyRule(),       // NEW
        rules.NewStatusCodeRule(),        // NEW
        rules.NewSuccessResponseRule(),   // NEW
    }

    // ... rest unchanged
}
```

---

#### 5. Test Data

**Files to Create:**
- `testdata/invalid-specs/missing-request-body.yaml` - REQ-005 violations
- `testdata/invalid-specs/invalid-status-code.yaml` - REQ-008 violations
- `testdata/invalid-specs/missing-success-response.yaml` - REQ-009 violations

**Content:** `testdata/invalid-specs/missing-request-body.yaml`
```yaml
openapi: 3.0.0
info:
  title: Missing Request Body
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      # Missing requestBody entirely
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
  /v1/users.update:
    post:
      requestBody:
        required: false  # Should be true
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
                type: object
```

**Content:** `testdata/invalid-specs/invalid-status-code.yaml`
```yaml
openapi: 3.0.0
info:
  title: Invalid Status Codes
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
        201:  # Not allowed, should be 200
          description: Created
          content:
            application/json:
              schema:
                type: object
        503:  # Not allowed, should be 500
          description: Service Unavailable
          content:
            application/json:
              schema:
                type: object
```

**Content:** `testdata/invalid-specs/missing-success-response.yaml`
```yaml
openapi: 3.0.0
info:
  title: Missing Success Response
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
        201:  # Has 201 but not 200
          description: Created
          content:
            application/json:
              schema:
                type: object
  /v1/users.list:
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
          # Missing content
```

---

### Phase 3 Validation Commands

**Unit Tests:**
```bash
go test ./internal/rules
go test ./internal
```

**All Tests:**
```bash
go test ./...
```

**Integration Tests:**
```bash
./duhrpc-lint testdata/invalid-specs/missing-request-body.yaml
# Should report request-body-required violations

./duhrpc-lint testdata/invalid-specs/invalid-status-code.yaml
# Should report status-code violations

./duhrpc-lint testdata/invalid-specs/missing-success-response.yaml
# Should report success-response violations
```

---

## Phase 4: Complex Rules

### Overview
Implement the final 2 validation rules that require schema traversal and reference resolution: content type validation and error response schema validation. These are the most complex rules requiring navigation of OpenAPI schema structures.

**Goal:** CLI validates all 8 DUH-RPC rules completely.

**Validation Criteria:**
- `go test ./internal/rules` passes all tests
- CLI correctly identifies and reports violations for:
  - Invalid content types or missing application/json
  - Error response schemas missing required fields
- Complex schema features work: $ref resolution, allOf/oneOf/anyOf
- Complete valid spec passes with no violations

---

### Changes Required

#### 1. Content Type Rule (`internal/rules/content_type.go`)

**File:** `internal/rules/content_type.go`

**Changes:** Implement REQ-006 - Content type validation.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// ContentTypeRule validates only allowed content types are used
type ContentTypeRule struct{}

func NewContentTypeRule() *ContentTypeRule

func (r *ContentTypeRule) Name() string

func (r *ContentTypeRule) Validate(doc *v3.Document) []internal.Violation
```

**Function Responsibilities:**
- `Name()`: Return `"content-type"`
- `Validate()`:
  - Define allowed types: `application/json`, `application/protobuf`, `application/octet-stream`
  - Iterate through all paths and operations
  - Check request body content types (if requestBody exists)
  - Check all response content types
  - For each content type key:
    - Convert to lowercase for comparison (case-insensitive per RFC 2045)
    - If contains semicolon (`;`) - violation: "MIME parameters not allowed"
    - If not in allowed list - violation: "Invalid content type"
  - Special check: Ensure `application/json` is present in request body content types

**Testing Requirements:**

**File:** `internal/rules/content_type_test.go`

```go
package rules_test

func TestContentTypeRule(t *testing.T)
```

**Test Objectives:**
- Verify allowed content types pass: `application/json`, `application/protobuf`, `application/octet-stream`
- Verify multiple content types are allowed (JSON + protobuf)
- Verify disallowed types are reported: `application/xml`, `text/html`, `text/plain`
- Verify MIME parameters are rejected: `application/json; charset=utf-8`
- Verify application/json is required (error if only protobuf present)
- Test both request body and response content types

**Context for Implementation:**
- Reference: Technical spec REQ-006 (lines 508-576)
- Check Content map keys (not HTTP headers)
- Use `strings.Contains(contentType, ";")` to detect MIME parameters
- Request and response can have different content types

---

#### 2. Error Response Schema Rule (`internal/rules/error_response.go`)

**File:** `internal/rules/error_response.go`

**Changes:** Implement REQ-007 - Error response schema validation.

```go
package rules

import (
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/pb33f/libopenapi/datamodel/high/base"
    "github.com/duh-rpc/duhrpc-lint/internal"
)

// ErrorResponseRule validates error response schemas have required structure
type ErrorResponseRule struct{}

func NewErrorResponseRule() *ErrorResponseRule

func (r *ErrorResponseRule) Name() string

func (r *ErrorResponseRule) Validate(doc *v3.Document) []internal.Violation

// Helper function to validate schema structure (recursive for $ref, allOf, etc.)
func validateErrorSchema(schema *base.Schema) error
```

**Import Note:**
- `base.Schema` is from `github.com/pb33f/libopenapi/datamodel/high/base`
- This package contains the base schema types used by libopenapi

**Function Responsibilities:**
- `Name()`: Return `"error-response-schema"`
- `Validate()`:
  - Define error status codes: 400, 401, 403, 404, 429, 452, 453, 454, 455, 500
  - Iterate through all paths, operations, and responses
  - For error status codes, validate schema structure:
    - Must have literal `Type: "object"` field (strict mode - not inferred)
    - Must have `required: [code, message]`
    - `code` field must be type integer
    - `message` field must be type string
    - `details` field (if present) must be type object
  - Handle $ref resolution using libopenapi's reference resolution
  - Handle allOf/oneOf/anyOf combinators
  - Report specific validation failures

`validateErrorSchema()`:
- Check if schema uses $ref - resolve and validate resolved schema
- Check if schema uses allOf - ensure combined schema satisfies requirements
- Check type field is explicitly "object" (not inferred from properties)
- Check required fields, and field types
- Return descriptive error if validation fails

**$ref Resolution Strategy:**
- Access document's model for schema lookups
- For MediaType, access `Schema` field which may be a SchemaProxy
- Use SchemaProxy's `Schema` field to get resolved schema
- If SchemaProxy.Schema is nil, this indicates unresolved reference (skip validation)
- For allOf/oneOf/anyOf, recursively validate each sub-schema
- Handle circular references by tracking visited schemas (defensive)

**Testing Requirements:**

**File:** `internal/rules/error_response_test.go`

```go
package rules_test

func TestErrorResponseRule(t *testing.T)
func TestErrorResponseRuleWithRef(t *testing.T)
func TestErrorResponseRuleWithAllOf(t *testing.T)
```

**Test Objectives:**
- Verify valid inline error schema passes
- Verify $ref to valid error schema passes
- Verify allOf with Error schema passes
- Verify missing required fields is reported
- Verify wrong field types are reported (code as string, message as integer)
- Verify missing "object" type is reported
- Verify details field with wrong type is reported
- Test complex schema scenarios (nested $ref, allOf combinations)

**Context for Implementation:**
- Reference: Technical spec REQ-007 (lines 580-725)
- This is the most complex rule requiring schema traversal
- Use libopenapi's Index for reference resolution
- May need to handle Schema vs SchemaProxy
- Reference: Technical spec "Challenge 2" (lines 1924-1929)
- Reference: Technical spec "Challenge 3" (lines 1931-1936)

---

#### 3. Wire Rules into Validator

**File:** `internal/validator.go`

**Changes:** Add final 2 rules to complete the set.

```go
func Validate(doc *v3.Document, filePath string) ValidationResult {
    allRules := []Rule{
        rules.NewPathFormatRule(),
        rules.NewHTTPMethodRule(),
        rules.NewQueryParamsRule(),
        rules.NewRequestBodyRule(),
        rules.NewStatusCodeRule(),
        rules.NewSuccessResponseRule(),
        rules.NewContentTypeRule(),        // NEW
        rules.NewErrorResponseRule(),      // NEW
    }

    // ... rest unchanged
}
```

---

#### 4. Test Data

**Files to Create:**
- `testdata/invalid-specs/invalid-content-type.yaml` - REQ-006 violations
- `testdata/invalid-specs/bad-error-schema.yaml` - REQ-007 violations
- `testdata/valid-spec.yaml` - Fully compliant spec (passes all 8 rules)

**Content:** `testdata/invalid-specs/invalid-content-type.yaml`
```yaml
openapi: 3.0.0
info:
  title: Invalid Content Types
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: true
        content:
          application/xml:  # Not allowed
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
  /v1/users.update:
    post:
      requestBody:
        required: true
        content:
          application/json; charset=utf-8:  # MIME parameters not allowed
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            text/html:  # Not allowed
              schema:
                type: string
```

**Content:** `testdata/invalid-specs/bad-error-schema.yaml`
```yaml
openapi: 3.0.0
info:
  title: Bad Error Schema
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
                type: object
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                # Missing required: [code, message]
                properties:
                  error:
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: string  # Should be integer
                  message:
                    type: integer  # Should be string
```

**Content:** `testdata/valid-spec.yaml` (Complete valid DUH-RPC spec)
```yaml
openapi: 3.0.0
info:
  title: Valid DUH-RPC API
  version: 1.0.0
  description: Fully compliant DUH-RPC specification

paths:
  /v1/users.create:
    post:
      operationId: createUser
      summary: Create a new user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, email]
              properties:
                name:
                  type: string
                email:
                  type: string
                  format: email
      responses:
        200:
          description: User created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  userId:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        500:
          description: Internal Server Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /v1/users.get-by-id:
    post:
      operationId: getUserById
      summary: Get user by ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [userId]
              properties:
                userId:
                  type: string
          application/protobuf:
            schema:
              type: string
              format: binary
      responses:
        200:
          description: User found
          content:
            application/json:
              schema:
                type: object
                properties:
                  userId:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
        404:
          description: User not found
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                    enum: [404]
                  message:
                    type: string
                  details:
                    type: object
        500:
          description: Internal Server Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
        details:
          type: object
```

---

### Phase 4 Validation Commands

**Unit Tests:**
```bash
go test ./internal/rules
```

**All Tests:**
```bash
go test ./...
```

**Integration Tests:**
```bash
./duhrpc-lint testdata/invalid-specs/invalid-content-type.yaml
# Should report content-type violations

./duhrpc-lint testdata/invalid-specs/bad-error-schema.yaml
# Should report error-response-schema violations

./duhrpc-lint testdata/valid-spec.yaml
# Should report: ✓ valid-spec.yaml is DUH-RPC compliant
# Exit code: 0
```

**Critical Success Test:**
```bash
./duhrpc-lint testdata/valid-spec.yaml
echo $?
# Must output: ✓ valid-spec.yaml is DUH-RPC compliant
# Must exit with code 0
```

---

// REVIEW: Each phase should include functional tests for each rule, thus no longer needing to specificly add integration tests in this phase, let me know if you have questions about functional testing requirements
## Phase 5: Integration & Documentation

### Overview
Finalize the tool with comprehensive end-to-end testing, documentation, build automation, and polish. This phase ensures the tool is production-ready and easy to use.

**Goal:** Production-ready CLI with complete documentation and build automation.

**Validation Criteria:**
- All end-to-end tests pass
- README provides clear installation and usage instructions
- Makefile automates common tasks
- All test data files exist and work correctly
- Tool can be installed and used by external developers

---

### Changes Required

#### 1. Comprehensive Test Data

**Files to Create:**
- `testdata/invalid-specs/multiple-violations.yaml` - Mix of all violation types

**Content:** `testdata/invalid-specs/multiple-violations.yaml`
```yaml
openapi: 3.0.0
info:
  title: Multiple Violations
  version: 1.0.0
  description: Spec with multiple DUH-RPC violations for testing

paths:
  /api/users/create:  # VIOLATION: bad path format
    get:  # VIOLATION: wrong HTTP method
      operationId: createUser
      parameters:
        - name: userId  # VIOLATION: query parameter
          in: query
          schema:
            type: string
      # VIOLATION: missing request body
      responses:
        201:  # VIOLATION: invalid status code
          description: Created
          content:
            application/xml:  # VIOLATION: invalid content type
              schema:
                type: object
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                # VIOLATION: bad error schema (missing required fields)
                properties:
                  error:
                    type: string

  /v1/products.list:
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
          # VIOLATION: missing content in 200 response
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
```

---

#### 2. End-to-End Integration Tests

**File:** `cmd/duhrpc-lint/integration_test.go`

**Changes:** Create comprehensive integration tests.

```go
package main_test

func TestIntegrationValidSpec(t *testing.T)
func TestIntegrationAllRuleViolations(t *testing.T)
func TestIntegrationMultipleViolations(t *testing.T)
```

**Test Objectives:**
- Verify complete valid spec passes with exit code 0 and success message
- Verify each invalid-specs/*.yaml file:
  - Returns exit code 1
  - Reports expected violation types
  - Includes suggestions in output
- Verify multiple-violations.yaml reports all expected violations
- Verify output format matches specification
- Test using actual filesystem files (not mocks)

**Context for Implementation:**
- Use `lint.RunCmd()` function for testability
- Capture stdout/stderr with bytes.Buffer
- Parse output to verify violation presence
- Count violations in output vs expected

**Output Verification Strategy:**
- Capture stdout using bytes.Buffer
- Check for expected strings:
  - Success: `strings.Contains(output, "✓")` and `strings.Contains(output, "compliant")`
  - Violations: `strings.Contains(output, "[rule-name]")` for each expected violation
  - Error: `strings.Contains(output, "Error:")`
- Count violations: `strings.Count(output, "[")` matches expected count
- Don't parse exact formatting (brittle), check for presence of key indicators

---

#### 3. Documentation

**File:** `README.md`

**Changes:** Create comprehensive README.

```markdown
# duhrpc-lint

Validate OpenAPI 3.0 specifications for DUH-RPC compliance.

## Overview

`duhrpc-lint` is a command-line tool that validates OpenAPI YAML specifications
against DUH-RPC conventions...

## Installation

### Using go install
...

### From source
...

## Usage

### Basic Usage
...

### Examples
...

### Exit Codes
...

## Validation Rules

Brief description of 8 rules...

## Development

### Running Tests
...

### Building
...

## License
...
```

**Content Sections:**
- Overview and purpose
- Installation instructions (go install, from source)
- Usage examples with real commands
- Exit code documentation
- Brief rule descriptions (link to technical spec for details)
- Development instructions
- Contributing guidelines (if applicable)
- License information

---

#### 4. Build Automation

**File:** `Makefile`

**Changes:** Create Makefile for common tasks.

```makefile
.PHONY: build test install lint clean coverage integration-test

# Build binary
build:
	go build -o duhrpc-lint ./cmd/duhrpc-lint

# Run all tests
test:
	go test -v -cover ./...

# Install locally
install:
	go install ./cmd/duhrpc-lint

# Lint code
lint:
	go vet ./...
	go fmt ./...

# Clean build artifacts
clean:
	rm -f duhrpc-lint
	go clean

# Coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Integration tests
integration-test: build
	@echo "Testing valid spec..."
	./duhrpc-lint testdata/valid-spec.yaml
	@echo "\nTesting invalid specs..."
	! ./duhrpc-lint testdata/invalid-specs/bad-path-format.yaml
	! ./duhrpc-lint testdata/invalid-specs/wrong-http-method.yaml
	! ./duhrpc-lint testdata/invalid-specs/has-query-params.yaml
	@echo "\nIntegration tests passed!"
```

**Targets:**
- `build`: Build binary
- `test`: Run all tests with coverage
- `install`: Install to GOPATH/bin
- `lint`: Run go vet and go fmt
- `clean`: Remove build artifacts
- `coverage`: Generate HTML coverage report
- `integration-test`: Run end-to-end tests

---

#### 5. Project Files

**File:** `.gitignore`

**Changes:** Create .gitignore for Go projects.

```
# Binaries
duhrpc-lint
*.exe
*.dll
*.so
*.dylib

# Test coverage
*.out
coverage.html

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
```

---

#### 6. Version Management

**File:** `cmd/duhrpc-lint/main.go`

**Changes:** Ensure version constant is easy to update.

```go
const version = "1.0.0"  // Update for releases
```

**Context for Implementation:**
- Keep version in main.go for simplicity
- Future: Could use ldflags for version injection during build

---

### Phase 5 Validation Commands

**All Tests:**
```bash
make test
# All unit tests pass
```

**Integration Tests:**
```bash
make integration-test
# All end-to-end tests pass
```

**Build:**
```bash
make build
./duhrpc-lint --version
# Outputs: duhrpc version 1.0.0
```

**Install and Use:**
```bash
make install
duhrpc-lint testdata/valid-spec.yaml
# Works from anywhere
```

**Coverage:**
```bash
make coverage
# Generates coverage.html
# Open in browser to verify >90% coverage
```

**Lint:**
```bash
make lint
# No issues reported
```

**Complete Validation Suite:**
```bash
# Test all invalid specs report violations
for file in testdata/invalid-specs/*.yaml; do
  echo "Testing $file..."
  ./duhrpc-lint "$file" || echo "  ✓ Violations detected"
done

# Test valid spec passes
./duhrpc-lint testdata/valid-spec.yaml
# Output: ✓ valid-spec.yaml is DUH-RPC compliant
```

---

## Summary of Deliverables

### Phase 1: Foundation
- ✅ Go module initialized
- ✅ Core types (Violation, ValidationResult, Rule interface)
- ✅ File loader with error handling
- ✅ Output formatter
- ✅ Validator framework
- ✅ CLI with testable run() pattern
- ✅ --help and --version flags
- ✅ Basic test data

### Phase 2: Simple Rules Batch 1
- ✅ Path format validation (REQ-002)
- ✅ HTTP method validation (REQ-003)
- ✅ Query parameter validation (REQ-004)
- ✅ Test data for 3 rules

### Phase 3: Simple Rules Batch 2
- [x] Request body validation (REQ-005)
- [x] Status code validation (REQ-008)
- [x] Success response validation (REQ-009)
- [x] Test data for 3 rules

### Phase 4: Complex Rules
- [x] Content type validation (REQ-006)
- [x] Error response schema validation (REQ-007)
- [x] Complete valid spec test data
- [x] All 8 rules implemented

### Phase 5: Integration & Documentation
- [x] Comprehensive end-to-end tests
- [x] README with usage examples
- [x] Makefile for automation
- [x] Multiple violations test data
- [x] Complete test coverage
- [x] Production-ready tool

---

## Final Acceptance Criteria

**Functional Requirements:**
- [x] CLI accepts OpenAPI YAML file path as argument
- [x] All 8 DUH-RPC rules are validated
- [x] Violations reported with clear messages and suggestions
- [x] Exit codes: 0 (valid), 1 (violations), 2 (errors)
- [x] --help and --version flags work
- [x] File not found and parse errors handled gracefully

**Testing Requirements:**
- [x] `go test ./...` passes all tests
- [x] Coverage >90% for all packages
- [x] Integration tests verify end-to-end behavior
- [x] All testdata files work correctly

**Documentation Requirements:**
- [x] README explains installation and usage
- [x] Code includes clear comments
- [x] Examples demonstrate common use cases

**Build Requirements:**
- [x] `make build` produces working binary
- [x] `make install` installs to GOPATH/bin
- [x] `make test` runs all tests successfully
- [x] `make integration-test` validates end-to-end
- [x] `go build ./cmd/duhrpc-lint` produces working binary (Phase 1)

**Code Quality:**
- [x] Follows CLAUDE.md guidelines (tests in _test package, table-driven, etc.)
- [x] No linting errors (`go vet`, `go fmt`)
- [x] Clear separation of concerns (loader, validator, rules, reporter)
- [x] Testable design (run() pattern, io.Writer parameters)

---

## Implementation Notes

### Key Design Decisions

1. **Testability First:** CLI uses `run(stdin, stdout, stderr, args)` pattern for easy testing without subprocess execution
2. **Rule Independence:** Each rule is self-contained, no dependencies between rules
3. **Progressive Enhancement:** Each phase builds on previous, always maintaining working state
4. **Test Data Strategy:** Incremental test data creation, add files as rules implemented
5. **Error Handling:** Clear distinction between tool errors (exit 2) and validation failures (exit 1)

### Common Pitfalls to Avoid

1. **CLAUDE.md Compliance:**
   - ❌ Don't put tests in same package as code
   - ✅ Use `package XXX_test` for all tests
   - ❌ Don't add descriptive messages to require/assert
   - ✅ Use `require.NoError(t, err)` not `require.NoError(t, err, "should not error")`

2. **Schema Traversal:**
   - ❌ Don't forget to resolve $ref references
   - ✅ Use libopenapi's Index for reference resolution
   - ❌ Don't assume schemas are inline
   - ✅ Handle allOf/oneOf/anyOf combinators

3. **Testing:**
   - ❌ Don't test implementation details
   - ✅ Test behavior and functionality
   - ❌ Don't use mocks for integration tests
   - ✅ Use real files from testdata/

4. **Output:**
   - ❌ Don't write directly to os.Stdout in libraries
   - ✅ Accept io.Writer parameter for testability

### Performance Considerations

- Compile regex patterns once at rule creation, not per-path
- Single document traversal per rule (don't re-parse)
- Lazy $ref resolution (only when needed for error responses)
- Expected performance: <1 second for 100 operations, <2 seconds for 500 operations

### Extension Points for Future

- Add new rules: Implement Rule interface, add to validator
- JSON output: Add --format flag and new reporter
- Configuration: Add config file parsing before validator
- Auto-fix: Add writer methods to rules that return fixed spec

---

END OF IMPLEMENTATION PLAN
