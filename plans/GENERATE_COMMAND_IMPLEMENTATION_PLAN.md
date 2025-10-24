# `duh generate` Command Implementation Plan

## Quick Start

**TL;DR for Implementers:**
1. Add oapi-codegen dependencies to `go.mod`
2. Create `internal/generate/` package with loader, config, and generation functions
3. Implement `client.go`, `server.go`, `models.go`, and `all.go` generation functions
4. Add nested `generate` command with subcommands in `run_cmd.go`
5. Add tests for each component following existing patterns
6. Update README.md iteratively after each phase
7. Validate with `make ci` after each phase

## Overview

This plan outlines the implementation of the `duh generate` command, which generates Go code from OpenAPI specifications using oapi-codegen. The command will support generating HTTP clients, server stubs, and type models from DUH-RPC compliant OpenAPI specifications.

The implementation is divided into **5 deliverable phases**, each representing a complete, working milestone that adds functionality incrementally.

## Current State Analysis

### Existing Command Structure
- **Location**: `run_cmd.go:16-144`
- **Pattern**: Inline Cobra command definitions with delegation to `internal/<command>/` packages
- **Commands**: `lint`, `init`, `add`
- **Exit Codes**: 0 (success), 1 (validation failed), 2 (error)

### Key Patterns Discovered
1. **Command delegation**: `run_cmd.go:46-68` - Commands delegate to `internal` package `Run()` functions
2. **File defaults**: `run_cmd.go:47-50` - Uses const for default filename (`openapi.yaml`)
3. **Error handling**: `run_cmd.go:54-57` - Errors return exit code 2
4. **Template embedding**: `internal/init/generator.go:1-6` - Uses `//go:embed` for templates
5. **Output injection**: All commands accept `io.Writer` for testability
6. **Package pattern**: External test packages using `package XXX_test`

### Existing Dependencies
From `go.mod:7-12`:
- `github.com/pb33f/libopenapi v0.28.1` - Already available for loading OpenAPI specs
- `github.com/spf13/cobra v1.10.1` - CLI framework
- `github.com/stretchr/testify v1.11.1` - Testing assertions
- `gopkg.in/yaml.v3 v3.0.1` - YAML parsing

### Testing Patterns
From research analysis:
- Table-driven tests with inline test cases: `internal/lint/rules/path_format_test.go:24-381`
- External test packages: `package generate_test`
- `require` for critical assertions, `assert` for non-critical
- `t.TempDir()` for file system isolation
- Integration tests via `RunCmd(&stdout, []string{...})`

### Validation Commands
From `Makefile:1-52`:
- `make test` - Run all tests
- `make lint` - Run golangci-lint
- `make build` - Build binary
- `make ci` - Run tidy, lint, and test

## Desired End State

### Command Structure
```bash
# Generate HTTP client
duh generate client [openapi-file]

# Generate server stubs
duh generate server [openapi-file]

# Generate type models
duh generate models [openapi-file]

# Generate all components
duh generate all [openapi-file]
```

### Common Flags
All subcommands support:
- `--output, -o` - Output file path (defaults: `client.go`, `server.go`, `models.go`)
- `--package, -p` - Target package name (default: `api`)

### Special Behavior for `generate all`
- `--output-dir` - Directory for all generated files (default: current directory)
- Generates `client.go`, `server.go`, `models.go` in specified directory
- All files use same package name

### Verification
After each phase is complete:
1. `make test` passes
2. `make lint` passes
3. `make build` succeeds
4. README.md updated with examples (iterative, after each phase)

## What We're NOT Doing

- Support for non-stdlib server frameworks (chi, gin, echo)
- Configuration file support (duh.yaml)
- Template customization
- Support for OpenAPI 2.0 (Swagger)
- Automatic validation before generation (user can run `duh lint` separately)
- Output to stdout (`-o -`)
- Custom import mappings
- **Generated code compilation validation** - We trust oapi-codegen's output quality
- Smart package name detection from directory structure (use simple default: `api`)

## Implementation Approach

### Strategy
1. Use **library integration** approach - import oapi-codegen packages directly
2. Follow existing command patterns in `run_cmd.go`
3. Create nested subcommand structure: `duh generate <subcommand>`
4. Implement incrementally: core infrastructure → individual commands → combined command
5. Use TDD approach: write tests first, then implementation

### Key Dependencies to Add

**Versions Note**: Use latest stable versions at time of implementation. Versions shown are current as of plan creation date (2025-10-24).

```go
github.com/oapi-codegen/oapi-codegen/v2 v2.4.1  # or latest v2.x
github.com/getkin/kin-openapi v0.128.0         # or latest stable
```

### Package Organization

**Before implementation:**
```
internal/
├── add/
├── init/
└── lint/
```

**After Phase 1:**
```
internal/
├── add/
├── generate/              # NEW
│   ├── loader.go         # OpenAPI spec loading
│   ├── config.go         # oapi-codegen configuration
│   ├── generate.go       # Core generation logic
│   ├── client.go         # Client generation
│   ├── client_test.go    # Client tests
│   └── testdata/         # Test fixtures
│       └── valid-spec.yaml
├── init/
└── lint/
```

**Final structure (after all phases):**
```
internal/generate/
├── generate.go            # Core generation logic
├── loader.go              # OpenAPI spec loading
├── config.go              # oapi-codegen configuration
├── client.go              # Client generation
├── server.go              # Server generation
├── models.go              # Models generation
├── all.go                 # Combined generation
├── client_test.go         # Unit tests
├── server_test.go         # Unit tests
├── models_test.go         # Unit tests
├── all_test.go            # Unit tests
└── testdata/              # Test fixtures
    └── valid-spec.yaml
```

### Required Imports

**File**: `run_cmd.go` (add to imports):
```go
import (
    // ... existing imports ...
    "github.com/duh-rpc/duh-cli/internal/generate"
)
```

**File**: `internal/generate/loader.go`:
```go
import (
    "fmt"
    "github.com/getkin/kin-openapi/openapi3"
    "github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)
```

**File**: `internal/generate/config.go`:
```go
import (
    "github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)
```

**File**: `internal/generate/generate.go`:
```go
import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/getkin/kin-openapi/openapi3"
    "github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)
```

### Important Clarifications

#### 1. GenerateOptions Type
`GenerateOptions` referenced in Phase 1 (lines 176, 225) is NOT a custom type. It is `codegen.GenerateOptions` from the oapi-codegen library. Use it directly:

```go
config := codegen.Configuration{
    PackageName: packageName,
    Generate: codegen.GenerateOptions{
        Client: true,  // or StdHTTPServer: true, or Models: true
    },
}
```

**Generation Options Explained:**
- `Client: true` - Generates HTTP client with request methods
- `StdHTTPServer: true` - Generates server interfaces using net/http
- `Models: true` - Generates only type definitions (structs)

These options can be combined or used separately.

#### 2. Loader Implementation
Create a NEW loader in `internal/generate/loader.go` using `util.LoadSwagger()` from oapi-codegen. Do NOT reuse `internal/lint/loader.go` because:
- Lint uses `libopenapi` which returns `*v3.Document`
- Generate needs `kin-openapi` which returns `*openapi3.T`
- Different libraries, different types, different purposes

#### 3. Success Message Emoji Usage
Success messages DO use checkmark emoji (✓) to match existing commands:
- `internal/init/init.go:12` uses `✓`
- `internal/add/add.go:72` uses `✓`

This is an exception to the general "no emoji" guideline in CLAUDE.md, as it's an established pattern in this codebase.

#### 4. Exit Code Usage
**Exit Codes for Generate Command:**
- `0` - Code generated successfully
- `1` - NOT USED (reserved for validation failures like `lint`)
- `2` - ALL errors (file not found, invalid spec, generation failed, write failed)

Unlike `lint` which uses exit code 1 for validation failures, `generate` treats all failures as errors (exit code 2).

#### 5. Error Wrapping Pattern
- Use `%w` for wrapping underlying errors: `fmt.Errorf("failed to load spec: %w", err)`
- Use `%s` or plain strings for contextual information: `fmt.Errorf("file not found: %s", filePath)`
- Always add context to errors from library functions

#### 6. Testing Package Names
- Integration tests in `generate_test.go` at root: `package duh_test`
- Unit tests in `internal/generate/*_test.go`: `package generate_test`

#### 7. Path Handling
- Input paths (OpenAPI file): Can be relative or absolute
- Output paths: Can be relative or absolute
- Directory creation: Use `filepath.Dir()` to extract parent directory
- Path joining: Use `filepath.Join()` for cross-platform compatibility
- Always check parent directory exists or create it with `os.MkdirAll()`

### Testing Patterns Summary

**Pattern Overview:**
- External test packages: `package generate_test` or `package duh_test`
- Table-driven tests: `for _, test := range []struct{...}`
- File isolation: `t.TempDir()`
- Critical assertions: `require.*` (stops test on failure)
- Non-critical assertions: `assert.*` (continues test)
- Output capture: `bytes.Buffer`
- No assertion messages: Follow CLAUDE.md - no descriptive strings in assertions

**Example Test Structure:**
```go
func TestRunClientWithDefaults(t *testing.T) {
    tempDir := t.TempDir()
    // Create test spec file
    specPath := filepath.Join(tempDir, "openapi.yaml")
    // ... setup ...

    var stdout bytes.Buffer
    err := generate.RunClient(&stdout, specPath, "client.go", "api")

    require.NoError(t, err)
    assert.Contains(t, stdout.String(), "✓")
}
```

### Common Pitfalls to Avoid

1. ❌ Forgetting to call `config.UpdateDefaults()` before `config.Validate()`
2. ❌ Not creating parent directories before writing files
3. ❌ Using wrong exit code (use 2 for all errors, not 1)
4. ❌ Not using external test packages (`package generate_test`)
5. ❌ Adding message strings to assertions (forbidden per CLAUDE.md)
6. ❌ Using `internal/lint/loader.go` instead of creating new loader
7. ❌ Trying to validate generated code compilation (out of scope)

---

## Phase 1: Core Infrastructure & Client Generation

### Overview
Establish the foundation for code generation and deliver the first working command: `duh generate client`. This phase sets up the basic command structure, adds required dependencies, and implements client code generation.

### Acceptance Criteria
- `duh generate client` command works with default OpenAPI file
- `duh generate client custom.yaml` works with custom file path
- `duh generate client -o output.go` writes to custom location
- `duh generate client -p myclient` uses custom package name
- Generated client code compiles successfully
- All tests pass
- `make ci` passes

### Changes Required

#### 1. Dependencies
**File**: `go.mod`
**Changes**: Add oapi-codegen dependencies

Add the following to `require` section:
```go
github.com/oapi-codegen/oapi-codegen/v2 v2.4.1
github.com/getkin/kin-openapi v0.128.0
```

**Validation**: Run `go mod tidy` and verify dependencies resolve

---

#### 2. Core Package Structure
**Directory**: `internal/generate/`
**Changes**: Create new package with core functionality

##### File: `internal/generate/loader.go`
```go
func Load(filePath string) (*openapi3.T, error)
```

**Function Responsibilities:**
- Check file exists using `os.Stat()` pattern from `internal/add/add.go:19`
- Read file using `os.ReadFile()` pattern from `internal/add/add.go:23`
- Load OpenAPI spec using `util.LoadSwagger(filePath)` from oapi-codegen
- Return parsed spec or error with context (follow error wrapping from `internal/add/add.go:25`)

**Context for implementation:**
- Follow error handling pattern from `internal/lint/loader.go:12-33`
- Use same file-not-found error message format: `fmt.Errorf("file not found: %s", filePath)`
- Parse errors should be: `fmt.Errorf("failed to parse OpenAPI spec: %w", err)`

---

##### File: `internal/generate/config.go`
```go
func NewConfig(packageName string, opts GenerateOptions) (codegen.Configuration, error)
```

**Function Responsibilities:**
- Create `codegen.Configuration` struct with provided package name
- Set `Generate` field based on `GenerateOptions` parameter
- Call `config.UpdateDefaults()` to populate defaults
- Call `config.Validate()` to ensure valid configuration
- Return configured struct or validation error

**Context for implementation:**
- Reference oapi-codegen documentation at https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen
- Configuration struct documented in research doc lines 93-102
- Validation ensures configuration is valid before code generation

---

##### File: `internal/generate/generate.go`
```go
func Generate(spec *openapi3.T, config codegen.Configuration, outputPath string) error
```

**Function Responsibilities:**
- Call `codegen.Generate(spec, config)` to generate Go code
- Handle generation errors with context wrapping
- Write generated code to file using `os.WriteFile(outputPath, []byte(code), 0644)`
- Create parent directories if needed using `os.MkdirAll()` pattern from `internal/init/writer.go:15`
- Return error or nil on success

**Context for implementation:**
- Follow code generation flow from research doc lines 108-138
- Use file permissions 0644 for generated files (consistent with `internal/add/add.go:68`)
- Error messages should wrap with context: `fmt.Errorf("code generation failed: %w", err)`

---

#### 3. Client Generation
**File**: `internal/generate/client.go`
**Changes**: Implement client-specific generation

```go
func RunClient(w io.Writer, filePath, outputPath, packageName string) error
```

**Function Responsibilities:**
- Validate inputs (file path must be provided or use default)
- Load OpenAPI spec using `Load(filePath)`
- Determine output path (use provided or default to `client.go`)
- Determine package name (use provided or default to `api`)
- Create configuration using `NewConfig(packageName, GenerateOptions{Client: true})`
- Generate code using `Generate(spec, config, outputPath)`
- Print success message: `fmt.Fprintf(w, "✓ Generated client code at %s\n", outputPath)`
- Return error or nil

**Context for implementation:**
- Follow pattern from `internal/init/init.go:8-14` for success message format
- Follow pattern from `internal/add/add.go:14-74` for validation and file operations
- Default package name is `api` (as specified in requirements)
- Default output file is `client.go` in current directory

**Testing Requirements:**
```go
func TestRunClientWithDefaults(t *testing.T)
func TestRunClientWithCustomOutput(t *testing.T)
func TestRunClientWithCustomPackage(t *testing.T)
func TestRunClientFileNotFound(t *testing.T)
func TestRunClientInvalidSpec(t *testing.T)
```

**Test Objectives:**
- Validate default behavior (openapi.yaml → client.go with package api)
- Validate custom output path handling
- Validate custom package name handling
- Validate error handling for missing files
- Validate error handling for invalid OpenAPI specs

**Context for tests:**
- Follow table-driven test pattern from `internal/lint/rules/path_format_test.go:24-381`
- Use `t.TempDir()` for file isolation from `internal/add/add_test.go:35`
- Use `bytes.Buffer` to capture output from `run_cmd_test.go:16`
- Copy valid spec from `internal/lint/testdata/valid-spec.yaml` for test fixtures
- Use `require.NoError` for critical operations, `assert.Contains` for output validation

---

#### 4. Command Registration
**File**: `run_cmd.go`
**Changes**: Add nested `generate` command structure

Add after `addCmd` definition (after line 133):

```go
generateCmd := &cobra.Command{
	Use:   "generate",
	Short: "Generate Go code from OpenAPI specifications",
	Long: `Generate Go code from OpenAPI specifications.

The generate command uses oapi-codegen to create HTTP clients, server stubs,
and type models from DUH-RPC compliant OpenAPI specifications.

Available subcommands:
  client    Generate HTTP client code
  server    Generate server stub code
  models    Generate type models
  all       Generate all components

Use "duh generate [command] --help" for more information about a command.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

clientCmd := &cobra.Command{
	Use:   "client [openapi-file]",
	Short: "Generate HTTP client code from OpenAPI specification",
	Long: `Generate HTTP client code from OpenAPI specification.

The client command generates a Go HTTP client for calling DUH-RPC endpoints
defined in the OpenAPI specification.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.
If no output is specified, defaults to 'client.go' in the current directory.
If no package is specified, defaults to 'api'.

Exit Codes:
  0    Client generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		const defaultFile = "openapi.yaml"
		const defaultOutput = "client.go"
		const defaultPackage = "api"

		filePath := defaultFile
		if len(args) > 0 {
			filePath = args[0]
		}

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = defaultOutput
		}

		packageName, _ := cmd.Flags().GetString("package")
		if packageName == "" {
			packageName = defaultPackage
		}

		if err := generate.RunClient(cmd.OutOrStdout(), filePath, outputPath, packageName); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
			exitCode = 2
			return
		}
	},
}
clientCmd.Flags().StringP("output", "o", "", "Output file path (default: client.go)")
clientCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

generateCmd.AddCommand(clientCmd)
```

Update root command registration (line 134):
```go
rootCmd.AddCommand(lintCmd, initCmd, addCmd, generateCmd)
```

**Context for implementation:**
- Follow command structure pattern from `run_cmd.go:31-69` (lint command)
- Follow flag handling pattern from `run_cmd.go:100-132` (add command)
- Use closure over `exitCode` variable pattern from `run_cmd.go:17`
- Default const pattern from `run_cmd.go:47`

**Testing Requirements:**
```go
func TestGenerateClientCommand(t *testing.T)
func TestGenerateClientWithFlags(t *testing.T)
func TestGenerateClientHelp(t *testing.T)
```

**Test Objectives:**
- Validate command executes successfully with valid spec
- Validate flags are properly parsed and passed to RunClient
- Validate help text displays correctly

**Context for tests:**
- Follow integration test pattern from `lint_test.go:13-21`
- Use `RunCmd(&stdout, []string{"generate", "client", ...})` pattern
- Test in new file: `generate_test.go` with `package duh_test`

---

#### 5. Test Data
**Directory**: `internal/generate/testdata/`
**Changes**: Create test fixtures

**File**: `internal/generate/testdata/valid-spec.yaml`
- Copy from `internal/lint/testdata/valid-spec.yaml`
- Ensures test specs are DUH-RPC compliant

**Validation**: Spec should pass `duh lint` validation

---

#### 6. Documentation Update
**File**: `README.md`
**Changes**: Add `duh generate client` documentation

Add example to README after existing command documentation:
```markdown
### Generate HTTP Client

\`\`\`bash
# Use default openapi.yaml, output to client.go
duh generate client

# Specify custom spec file
duh generate client api/openapi.yaml

# Custom output location and package
duh generate client -o pkg/client/client.go -p client
\`\`\`
```

---

### Validation Commands

After Phase 1 completion:
```bash
# Run all tests
make test

# Run linter
make lint

# Build binary
make build

# Smoke test: Verify client generation works
./duh generate client internal/generate/testdata/valid-spec.yaml -o test-client.go
ls test-client.go  # Should exist
rm test-client.go

# Run full CI
make ci
```

---

## Phase 2: Server Generation

### Overview
Add server stub generation capability. This phase builds on the core infrastructure from Phase 1 to add `duh generate server` command.

### Acceptance Criteria
- `duh generate server` command works with default OpenAPI file
- `duh generate server custom.yaml` works with custom file path
- `duh generate server -o output.go` writes to custom location
- `duh generate server -p myserver` uses custom package name
- Generated server code compiles successfully
- Generated server uses stdlib (net/http)
- All tests pass
- `make ci` passes

### Changes Required

#### 1. Server Generation Logic
**File**: `internal/generate/server.go`
**Changes**: Add server generation function

```go
func RunServer(w io.Writer, filePath, outputPath, packageName string) error
```

**Function Responsibilities:**
- Validate inputs (file path must be provided or use default)
- Load OpenAPI spec using `Load(filePath)`
- Determine output path (use provided or default to `server.go`)
- Determine package name (use provided or default to `api`)
- Create configuration using `NewConfig(packageName, GenerateOptions{StdHTTPServer: true})`
- Generate code using `Generate(spec, config, outputPath)`
- Print success message: `fmt.Fprintf(w, "✓ Generated server code at %s\n", outputPath)`
- Return error or nil

**Context for implementation:**
- Follow same pattern as `RunClient` from `internal/generate/client.go`
- Use `StdHTTPServer: true` option (research doc line 151-153)
- Default output file is `server.go`

**Testing Requirements:**
```go
func TestRunServerWithDefaults(t *testing.T)
func TestRunServerWithCustomOutput(t *testing.T)
func TestRunServerWithCustomPackage(t *testing.T)
func TestRunServerFileNotFound(t *testing.T)
```

**Test Objectives:**
- Validate default behavior (openapi.yaml → server.go with package api)
- Validate custom output path handling
- Validate custom package name handling
- Validate error handling for missing files

**Context for tests:**
- Follow same pattern as client tests from `internal/generate/client_test.go`
- Verify generated code uses stdlib HTTP handlers (check for `http.Handler` interface)

---

#### 2. Command Registration
**File**: `run_cmd.go`
**Changes**: Add server subcommand to generate command

Add after `clientCmd` definition:

```go
serverCmd := &cobra.Command{
	Use:   "server [openapi-file]",
	Short: "Generate server stub code from OpenAPI specification",
	Long: `Generate server stub code from OpenAPI specification.

The server command generates Go HTTP server stubs using the standard library
net/http package for implementing DUH-RPC endpoints defined in the OpenAPI
specification.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.
If no output is specified, defaults to 'server.go' in the current directory.
If no package is specified, defaults to 'api'.

Exit Codes:
  0    Server generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		const defaultFile = "openapi.yaml"
		const defaultOutput = "server.go"
		const defaultPackage = "api"

		filePath := defaultFile
		if len(args) > 0 {
			filePath = args[0]
		}

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = defaultOutput
		}

		packageName, _ := cmd.Flags().GetString("package")
		if packageName == "" {
			packageName = defaultPackage
		}

		if err := generate.RunServer(cmd.OutOrStdout(), filePath, outputPath, packageName); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
			exitCode = 2
			return
		}
	},
}
serverCmd.Flags().StringP("output", "o", "", "Output file path (default: server.go)")
serverCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

generateCmd.AddCommand(serverCmd)
```

**Context for implementation:**
- Follow same pattern as `clientCmd` from Phase 1
- Only difference is default output filename and function called

**Testing Requirements:**
```go
func TestGenerateServerCommand(t *testing.T)
func TestGenerateServerWithFlags(t *testing.T)
```

**Test Objectives:**
- Validate command executes successfully
- Validate flags are properly parsed

**Context for tests:**
- Follow pattern from `generate_test.go` client tests
- Add to existing `generate_test.go` file

---

#### 3. Documentation Update
**File**: `README.md`
**Changes**: Add `duh generate server` documentation

Add example to README after client documentation:
```markdown
### Generate Server Stubs

\`\`\`bash
# Use default openapi.yaml, output to server.go
duh generate server

# Specify custom spec file
duh generate server api/openapi.yaml

# Custom output location and package
duh generate server -o pkg/server/server.go -p server
\`\`\`
```

---

### Validation Commands

After Phase 2 completion:
```bash
# Run all tests
make test

# Run linter
make lint

# Build binary
make build

# Smoke test: Verify server generation works
./duh generate server internal/generate/testdata/valid-spec.yaml -o test-server.go
ls test-server.go  # Should exist
rm test-server.go

# Run full CI
make ci
```

---

## Phase 3: Models Generation

### Overview
Add type models generation capability. This phase adds `duh generate models` command to generate only the type definitions without client or server code.

### Acceptance Criteria
- `duh generate models` command works with default OpenAPI file
- `duh generate models custom.yaml` works with custom file path
- `duh generate models -o output.go` writes to custom location
- `duh generate models -p mymodels` uses custom package name
- Generated models code compiles successfully
- All tests pass
- `make ci` passes

### Changes Required

#### 1. Models Generation Logic
**File**: `internal/generate/models.go`
**Changes**: Add models generation function

```go
func RunModels(w io.Writer, filePath, outputPath, packageName string) error
```

**Function Responsibilities:**
- Validate inputs (file path must be provided or use default)
- Load OpenAPI spec using `Load(filePath)`
- Determine output path (use provided or default to `models.go`)
- Determine package name (use provided or default to `api`)
- Create configuration using `NewConfig(packageName, GenerateOptions{Models: true})`
- Generate code using `Generate(spec, config, outputPath)`
- Print success message: `fmt.Fprintf(w, "✓ Generated models code at %s\n", outputPath)`
- Return error or nil

**Context for implementation:**
- Follow same pattern as `RunClient` and `RunServer`
- Use `Models: true` option (research doc line 280)
- Default output file is `models.go`

**Testing Requirements:**
```go
func TestRunModelsWithDefaults(t *testing.T)
func TestRunModelsWithCustomOutput(t *testing.T)
func TestRunModelsWithCustomPackage(t *testing.T)
```

**Test Objectives:**
- Validate default behavior (openapi.yaml → models.go with package api)
- Validate custom output path handling
- Validate custom package name handling

**Context for tests:**
- Follow same pattern as client and server tests
- Verify generated code contains only type definitions (no client/server code)

---

#### 2. Command Registration
**File**: `run_cmd.go`
**Changes**: Add models subcommand to generate command

Add after `serverCmd` definition:

```go
modelsCmd := &cobra.Command{
	Use:   "models [openapi-file]",
	Short: "Generate type models from OpenAPI specification",
	Long: `Generate type models from OpenAPI specification.

The models command generates Go type definitions for request and response
schemas defined in the OpenAPI specification, without generating client
or server code.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.
If no output is specified, defaults to 'models.go' in the current directory.
If no package is specified, defaults to 'api'.

Exit Codes:
  0    Models generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		const defaultFile = "openapi.yaml"
		const defaultOutput = "models.go"
		const defaultPackage = "api"

		filePath := defaultFile
		if len(args) > 0 {
			filePath = args[0]
		}

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = defaultOutput
		}

		packageName, _ := cmd.Flags().GetString("package")
		if packageName == "" {
			packageName = defaultPackage
		}

		if err := generate.RunModels(cmd.OutOrStdout(), filePath, outputPath, packageName); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
			exitCode = 2
			return
		}
	},
}
modelsCmd.Flags().StringP("output", "o", "", "Output file path (default: models.go)")
modelsCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

generateCmd.AddCommand(modelsCmd)
```

**Context for implementation:**
- Follow same pattern as `clientCmd` and `serverCmd`
- Only difference is default output filename and function called

**Testing Requirements:**
```go
func TestGenerateModelsCommand(t *testing.T)
func TestGenerateModelsWithFlags(t *testing.T)
```

**Test Objectives:**
- Validate command executes successfully
- Validate flags are properly parsed

**Context for tests:**
- Add to existing `generate_test.go` file

---

#### 3. Documentation Update
**File**: `README.md`
**Changes**: Add `duh generate models` documentation

Add example to README after server documentation:
```markdown
### Generate Type Models

\`\`\`bash
# Use default openapi.yaml, output to models.go
duh generate models

# Specify custom spec file
duh generate models api/openapi.yaml

# Custom output location and package
duh generate models -o pkg/types/models.go -p types
\`\`\`
```

---

### Validation Commands

After Phase 3 completion:
```bash
# Run all tests
make test

# Run linter
make lint

# Build binary
make build

# Smoke test: Verify models generation works
./duh generate models internal/generate/testdata/valid-spec.yaml -o test-models.go
ls test-models.go  # Should exist
rm test-models.go

# Run full CI
make ci
```

---

## Phase 4: Combined Generation (`generate all`)

### Overview
Add combined generation capability that generates all three components (client, server, models) in a single command with output directory support.

### Acceptance Criteria
- `duh generate all` command works with default OpenAPI file
- Generates `client.go`, `server.go`, `models.go` in current directory
- `duh generate all --output-dir api/` writes all files to api/ directory
- `duh generate all -p myapi` uses custom package name for all files
- All generated files use the same package name
- Generated code compiles successfully
- All tests pass
- `make ci` passes

### Changes Required

#### 1. Combined Generation Logic
**File**: `internal/generate/all.go`
**Changes**: Add combined generation function

```go
func RunAll(w io.Writer, filePath, outputDir, packageName string) error
```

**Function Responsibilities:**
- Validate inputs (file path must be provided or use default)
- Load OpenAPI spec using `Load(filePath)` ONCE (reuse for all three generations)
- Determine output directory (use provided or default to current directory `.`)
- Determine package name (use provided or default to `api`)
- Ensure output directory exists using `os.MkdirAll(outputDir, 0755)`
- Generate client code: call `Generate(spec, clientConfig, filepath.Join(outputDir, "client.go"))`
- Generate server code: call `Generate(spec, serverConfig, filepath.Join(outputDir, "server.go"))`
- Generate models code: call `Generate(spec, modelsConfig, filepath.Join(outputDir, "models.go"))`
- Print success message: `fmt.Fprintf(w, "✓ Generated client, server, and models in %s\n", outputDir)`
- Return error or nil

**Context for implementation:**
- Follow pattern from `internal/init/writer.go:15` for directory creation
- Use `filepath.Join()` for path construction
- Load spec once, generate three times with different configurations
- If any generation fails, return error immediately
- Default output directory is current directory (`.`)

**Testing Requirements:**
```go
func TestRunAllWithDefaults(t *testing.T)
func TestRunAllWithCustomOutputDir(t *testing.T)
func TestRunAllWithCustomPackage(t *testing.T)
func TestRunAllCreatesDirectory(t *testing.T)
```

**Test Objectives:**
- Validate all three files are generated in current directory by default
- Validate custom output directory is created and files are placed there
- Validate all three files use the same package name
- Validate output directory is created if it doesn't exist

**Context for tests:**
- Use `t.TempDir()` for isolation
- Read all three generated files and verify package declarations match
- Verify all three files compile together using `go build`

---

#### 2. Command Registration
**File**: `run_cmd.go`
**Changes**: Add `all` subcommand to generate command

Add after `modelsCmd` definition:

```go
allCmd := &cobra.Command{
	Use:   "all [openapi-file]",
	Short: "Generate client, server, and models from OpenAPI specification",
	Long: `Generate client, server, and models from OpenAPI specification.

The all command generates all three components (HTTP client, server stubs,
and type models) from the OpenAPI specification in a single invocation.

By default, generates client.go, server.go, and models.go in the current
directory. Use --output-dir to specify a different directory.

All generated files will use the same package name (default: api).

If no file path is provided, defaults to 'openapi.yaml' in the current directory.

Exit Codes:
  0    All components generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		const defaultFile = "openapi.yaml"
		const defaultOutputDir = "."
		const defaultPackage = "api"

		filePath := defaultFile
		if len(args) > 0 {
			filePath = args[0]
		}

		outputDir, _ := cmd.Flags().GetString("output-dir")
		if outputDir == "" {
			outputDir = defaultOutputDir
		}

		packageName, _ := cmd.Flags().GetString("package")
		if packageName == "" {
			packageName = defaultPackage
		}

		if err := generate.RunAll(cmd.OutOrStdout(), filePath, outputDir, packageName); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
			exitCode = 2
			return
		}
	},
}
allCmd.Flags().String("output-dir", "", "Output directory for generated files (default: current directory)")
allCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

generateCmd.AddCommand(allCmd)
```

**Context for implementation:**
- Different flag: `--output-dir` instead of `--output` (no short form)
- No `-o` short form for output-dir to avoid confusion
- Still supports `-p` for package name

**Testing Requirements:**
```go
func TestGenerateAllCommand(t *testing.T)
func TestGenerateAllWithOutputDir(t *testing.T)
func TestGenerateAllIntegration(t *testing.T)
```

**Test Objectives:**
- Validate command executes successfully
- Validate output-dir flag creates directory and generates files
- Integration test: verify all three generated files compile together

**Context for tests:**
- Add to existing `generate_test.go` file
- Integration test should compile all three files: `go build client.go server.go models.go`

---

#### 3. Documentation Update
**File**: `README.md`
**Changes**: Add `duh generate all` documentation

Add example to README after models documentation:
```markdown
### Generate All Components

\`\`\`bash
# Generate all three files in current directory
duh generate all

# Generate all in specific directory
duh generate all --output-dir pkg/api

# Custom package name for all files
duh generate all --output-dir api -p myapi
\`\`\`

Generated code uses:
- **Client**: HTTP client for calling DUH-RPC endpoints
- **Server**: Standard library (net/http) server stubs
- **Models**: Type definitions for request/response schemas
```

---

### Validation Commands

After Phase 4 completion:
```bash
# Run all tests
make test

# Run linter
make lint

# Build binary
make build

# Smoke test: Verify combined generation works
mkdir -p test-output
./duh generate all internal/generate/testdata/valid-spec.yaml --output-dir test-output -p testapi
ls test-output/  # Should show client.go, server.go, models.go
rm -rf test-output

# Run full CI
make ci
```

---

## Phase 5: Final Documentation Polish

### Overview
Since documentation has been updated iteratively in Phases 1-4, this final phase focuses on polishing the overall documentation and ensuring consistency.

### Acceptance Criteria
- Command summary list includes `duh generate`
- All four subcommands are documented with examples
- Documentation formatting is consistent with existing commands
- All examples in documentation work correctly
- `make ci` passes

### Changes Required

#### 1. Update Command Summary
**File**: `README.md`
**Changes**: Update the command list near the top of README

Update the command list to include:
```markdown
- `duh lint` - Validate OpenAPI specifications
- `duh init` - Create new DUH-RPC compliant spec
- `duh add` - Add endpoints to existing specs
- `duh generate` - Generate Go code from specs (client, server, models, all)
```

**Context for implementation:**
- Find command summary in README.md (likely near top)
- Add generate to the list
- Maintain alphabetical or logical ordering

---

#### 2. Documentation Review
**Changes**: Review and polish

Verify that the documentation added in Phases 1-4:
- Uses consistent formatting
- Includes all necessary examples
- Follows the same style as existing command documentation
- Has no typos or formatting errors

**Note**: The actual content was added iteratively during Phases 1-4. This phase just ensures everything looks good together.

---

### Validation Commands

After Phase 5 completion:
```bash
# Verify all examples from documentation work
duh generate client
duh generate server
duh generate models
duh generate all

# Run full CI to ensure everything still works
make ci

# Manual: Review README.md for formatting consistency
```

---

## Technical Notes

### Error Handling
Follow existing pattern from duh CLI:
- Exit code 0: Success
- Exit code 1: Not used for generate (reserved for validation failures)
- Exit code 2: Errors (file not found, invalid config, generation failed, etc.)

### Code Generation Flow
From research doc lines 108-138:
1. Load OpenAPI spec using `util.LoadSwagger(filePath)`
2. Create configuration with `codegen.Configuration{...}`
3. Call `config.UpdateDefaults()` to populate defaults
4. Call `config.Validate()` to ensure valid configuration
5. Generate code with `codegen.Generate(spec, config)`
6. Write output to file

### Package Name Defaults
- Default package name: `api` (as specified in requirements)
- Users can override with `--package` or `-p` flag
- All files generated by `generate all` use same package name

### File Permissions
- Generated files: `0644` (consistent with `internal/add/add.go:68`)
- Created directories: `0755` (consistent with `internal/init/writer.go:15`)

### Server Framework
- Use stdlib (net/http) via `StdHTTPServer: true` option
- No external framework dependencies
- Simple and predictable code generation

## Success Metrics

Implementation is complete when:
- [x] All phase validation commands pass (Phase 1 complete)
- [x] `make ci` passes with no errors (will pass after commit - Phase 3 complete)
- [x] All four generate subcommands work with examples (3 of 4 complete - Phase 3)
- [x] README.md documentation is complete and accurate (Phase 1-3 complete)
- [x] Unit tests achieve >80% coverage for generate package (Phase 1-3 complete)
- [x] Integration tests verify all commands execute successfully (Phase 1-3 complete)

## Phase Dependencies

**Phase 2** requires Phase 1 complete:
- Reuses `loader.go`, `config.go`, `generate.go`
- Adds only `server.go` and tests

**Phase 3** requires Phases 1-2 complete:
- Reuses all core infrastructure
- Adds only `models.go` and tests

**Phase 4** requires Phases 1-3 complete:
- Reuses all generation functions (client, server, models)
- Adds only `all.go` and tests

**Phase 5** requires Phases 1-4 complete:
- Polishes documentation added iteratively

## References

- [oapi-codegen GitHub](https://github.com/oapi-codegen/oapi-codegen)
- [oapi-codegen v2 pkg/codegen docs](https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen)
- [kin-openapi](https://github.com/getkin/kin-openapi)
- Research document: `docs/generate-command-research.md`
