# Remove `duh generate oapi` and Promote `duh generate duh` Implementation Plan

## Overview

This plan restructures the CLI command hierarchy by removing the `duh generate oapi` command entirely and promoting `duh generate duh` to become the primary `duh generate` command. This simplifies the user interface by having a single code generation path.

## Current State Analysis

### Command Structure
The CLI currently has a two-level generation command hierarchy:
```
duh generate (parent command)
├── duh generate oapi    (generates via oapi-codegen)
└── duh generate duh     (generates DUH-RPC specific code)
```

**Implementation Details:**
- `run_cmd.go:136-151` - `generateCmd` parent command with help text
- `run_cmd.go:153-200` - `oapiCmd` command definition
- `run_cmd.go:202-261` - `duhCmd` command definition
- `run_cmd.go:270` - Subcommand registration: `generateCmd.AddCommand(oapiCmd, duhCmd)`
- `run_cmd.go:9` - Import: `"github.com/duh-rpc/duh-cli/internal/generate/oapi"`

### Implementation Code
- `/internal/generate/oapi/` - Complete oapi-codegen wrapper implementation
  - `oapi.go` - Main entry point
  - `client.go`, `server.go`, `models.go` - Component generators
  - `generate.go`, `config.go`, `loader.go` - Core functionality
  - `oapi_test.go` - Functional tests
- `/internal/generate/duh/` - DUH-RPC specific generation (remains unchanged)

### Documentation
- `README.md:113-152` - Documents `duh generate duh` command
- `README.md:153-172` - Documents `duh generate oapi` command
- `README.md:179` - References `duh generate duh --help`
- `internal/generate/duh/templates/Makefile.tmpl:1` - Comment references `duh generate duh --full`

### Historical Plan Documents (No Changes Required)
Files in `/plans/` directory contain references to both commands but will remain unchanged as historical records:
- `buf_files_generation_plan.md`
- `openapi_proto_integration_plan.md`
- `graphql-generation-spec.md`
- `duh_generate_technical_spec.md`
- `full_flag_implementation_plan.md`
- `duh_generate_implementation_plan.md`
- `fix-service-stub-generation.md`

### Key Discoveries
- Framework: Uses Cobra CLI framework (github.com/spf13/cobra v1.10.1)
- Testing Style: All tests use functional testing via `duh.RunCmd()`
- No existing tests directly test the `generate` commands in `run_cmd_test.go`
- The `duh` command generates more comprehensive output than `oapi` (includes proto files, buf config, pagination iterators)

## Desired End State

### Command Structure
```
duh generate (executable command, not just parent)
```

Users will run:
```bash
duh generate              # instead of: duh generate duh
duh generate --full       # instead of: duh generate duh --full
duh generate api.yaml     # instead of: duh generate duh api.yaml
```

### Verification
After implementation:
1. `duh generate --help` shows the DUH-RPC generation help (current `duhCmd` help)
2. `duh generate` successfully generates client, server, proto files, etc.
3. `duh generate oapi` returns "unknown command" error
4. All existing tests in `/internal/generate/duh/` continue to pass
5. README documentation accurately reflects new command structure

## What We're NOT Doing

- Not maintaining backward compatibility or deprecation warnings for `duh generate oapi`
- Not keeping the oapi-codegen code in the codebase
- Not updating historical plan documents in `/plans/` directory
- Not changing the internal functionality of the DUH-RPC generation
- Not modifying the underlying `internal/generate/duh/` package implementation

## Implementation Approach

This is a straightforward refactoring with file deletions and command structure changes. The core generation logic remains untouched. The primary risk is ensuring all documentation references are updated consistently.

## Phase 1: Remove OAPI Command and Promote DUH Command

### Overview
Transform the CLI command structure by replacing the parent `generateCmd` with the functionality of `duhCmd`, removing all oapi-related code, and updating documentation.

### Changes Required

#### 1. Update CLI Command Structure
**File**: `run_cmd.go`

**Changes**:
1. Remove the oapi import (line 9)
2. Remove `oapiCmd` definition (lines 153-200)
3. Replace `generateCmd` definition (lines 136-151) with `duhCmd` functionality
4. Remove `duhCmd` variable definition (lines 202-262)
5. **Delete subcommand registration** (line 270: `generateCmd.AddCommand(oapiCmd, duhCmd)`)

**Modified Command Definition:**

Replace lines 136-270 with a single `generateCmd` that has the current `duhCmd` functionality:

```go
generateCmd := &cobra.Command{
	Use:   "generate [openapi-file]",
	Short: "Generate DUH-RPC client, server, and proto from OpenAPI specification",
	Long: `Generate DUH-RPC client, server, and proto from OpenAPI specification.

The generate command generates DUH-RPC specific code including HTTP client with
pagination iterators, server with routing, and protobuf definitions.

By default, generates client.go, server.go, iterator.go (if list operations),
proto file, buf.yaml, and buf.gen.yaml. Use flags to customize output.

After generation, run 'buf generate' to generate Go code from proto files,
then run 'go mod tidy' to update dependencies.

With --full flag, additionally generates editable scaffolding files:
  - daemon.go: Service orchestration with TLS/HTTP support
  - service.go: Service implementation (full or stub based on spec)
  - api_test.go: Integration tests (full suite or minimal example)
  - Makefile: Build automation with test, lint, and proto targets

If the OpenAPI spec matches 'duh init' template (users.create, users.get,
users.list, users.update), full implementations are generated. Otherwise,
stub implementations with TODO comments are generated for you to fill in.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.

Exit Codes:
  0    All components generated successfully
  2    Error (file not found, validation failed, generation failed, etc.)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		const defaultFile = "openapi.yaml"
		filePath := defaultFile
		if len(args) > 0 {
			filePath = args[0]
		}

		packageName, _ := cmd.Flags().GetString("package")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		protoPath, _ := cmd.Flags().GetString("proto-path")
		protoImport, _ := cmd.Flags().GetString("proto-import")
		protoPackage, _ := cmd.Flags().GetString("proto-package")
		fullFlag, _ := cmd.Flags().GetBool("full")

		if err := duh.Run(duh.RunConfig{
			Writer:       cmd.OutOrStdout(),
			SpecPath:     filePath,
			PackageName:  packageName,
			OutputDir:    outputDir,
			ProtoPath:    protoPath,
			ProtoImport:  protoImport,
			ProtoPackage: protoPackage,
			FullFlag:     fullFlag,
			Converter:    duh.NewProtoConverter(),
		}); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
			exitCode = 2
			return
		}
	},
}
generateCmd.Flags().StringP("package", "p", "api", "Package name for generated code")
generateCmd.Flags().String("output-dir", ".", "Output directory for generated files")
generateCmd.Flags().String("proto-path", "proto/v1/api.proto", "Proto file path")
generateCmd.Flags().String("proto-import", "", "Proto import override (optional)")
generateCmd.Flags().String("proto-package", "", "Proto package override (optional)")
generateCmd.Flags().Bool("full", false, "Generate additional editable scaffolding files")
```

**Import Changes:**
Remove line 9:
```go
"github.com/duh-rpc/duh-cli/internal/generate/oapi"
```

**Subcommand Registration:**
Delete line 270 entirely:
```go
generateCmd.AddCommand(oapiCmd, duhCmd)  // DELETE THIS LINE
```

**Root Command Registration:**
Line 272 remains unchanged:
```go
rootCmd.AddCommand(lintCmd, initCmd, addCmd, generateCmd)
```

**Function Responsibilities:**
- Parse command arguments and flags following existing pattern from `lintCmd`
- Call `duh.Run()` with configuration struct
- Handle errors and set exit codes appropriately
- Default to `openapi.yaml` if no file specified

**Testing Requirements:**

No new tests required. Existing tests will validate:

**Existing tests that will continue to pass:**
```go
func TestRunCmdHelp(t *testing.T)
func TestRunCmdVersion(t *testing.T)
```

All tests in `/internal/generate/duh/*_test.go` will continue to pass as the underlying implementation is unchanged.

**Test Objectives:**
- Verify `duh generate --help` displays correct help text
- Verify `duh generate` command executes successfully
- Verify existing DUH generation tests continue to pass
- Verify `duh generate oapi` returns appropriate error (unknown command)

**Context for Implementation:**
- Follow existing command pattern from `lintCmd` (run_cmd.go:33-71)
- Use same error handling approach: set `exitCode = 2` on error
- Maintain consistent flag naming with existing codebase conventions
- Keep help text format consistent with other commands

#### 2. Update README Documentation
**File**: `README.md`

**Changes**:
1. Update "Generate Code" section header (approximately line 111)
2. Replace all `duh generate duh` with `duh generate` (approximately lines 119, 122, 125, 128, 131, 179)
3. Remove "Generate OAPI Client, Server, and Models" section entirely (approximately lines 153-172, which is ~20 lines)

**Note**: Line numbers are approximate and will shift after removing the OAPI section. Use search functionality to locate the exact sections.

**Modified Sections:**

**Line 111-152 - Update to:**
```markdown
### Generate Code

Generate DUH-RPC specific code including HTTP client with pagination iterators, server with routing, and protobuf definitions:

```bash
# Generate from openapi.yaml (default) to current directory
duh generate

# Specify custom spec file
duh generate api/openapi.yaml

# Custom output directory and package
duh generate --output-dir pkg/api -p myapi

# Generate with full scaffolding (daemon, service implementation, tests, Makefile)
duh generate --full

# Custom proto settings
duh generate --proto-path proto/v1/api.proto --proto-package myapi.v1
```

**Generated files:**
- `client.go` - HTTP client with method calls for each endpoint
- `server.go` - HTTP server with routing
- `iterator.go` - Pagination iterators (if list operations exist)
- `proto/v1/api.proto` - Protobuf definitions
- `buf.yaml` and `buf.gen.yaml` - Buf configuration

**With --full flag, also generates:**
- `daemon.go` - Service orchestration with TLS/HTTP support
- `service.go` - Service implementation (full or stub based on spec)
- `api_test.go` - Integration tests (full suite or minimal example)
- `Makefile` - Build automation with test, lint, and proto targets

After generation, run:
```bash
buf generate      # Generate Go code from proto files
go mod tidy       # Update dependencies
```
```

**Line 173-179 - Update to:**
```markdown
### Command-line Options

```bash
# Show help for any command
duh help
duh lint --help
duh generate --help

# Show version
duh --version

# Generate shell completion script
duh completion bash   # For bash
duh completion zsh    # For zsh
duh completion fish   # For fish
```
```

**Testing Requirements:**

**Existing tests that may require review:**
None - documentation changes don't affect tests

**Test Objectives:**
- Manually verify README renders correctly in markdown viewer
- Verify all command examples are accurate
- Verify code blocks have proper syntax highlighting

**Context for Implementation:**
- Maintain existing markdown formatting style
- Keep consistent indentation in code blocks
- Preserve heading hierarchy
- Ensure all bash examples are executable

#### 3. Update Template File Comments
**Files**:
- `internal/generate/duh/templates/daemon.go.tmpl`
- `internal/generate/duh/templates/service.go.tmpl`
- `internal/generate/duh/templates/api_test.go.tmpl`
- `internal/generate/duh/templates/Makefile.tmpl`

**Changes**: Update the generation command reference in header comments from `'duh generate duh --full'` to `'duh generate --full'`

**Modified Content for each file:**

**daemon.go.tmpl line 1:**
```go
// Code generated by 'duh generate --full' on {{.Timestamp}}. YOU CAN EDIT.
```

**service.go.tmpl line 1:**
```go
// Code generated by 'duh generate --full' on {{.Timestamp}}. YOU CAN EDIT.
```

**api_test.go.tmpl line 1:**
```go
// Code generated by 'duh generate --full' on {{.Timestamp}}. YOU CAN EDIT.
```

**Makefile.tmpl line 1:**
```makefile
# Code generated by 'duh generate --full' on {{.Timestamp}}. YOU CAN EDIT.
```

**Testing Requirements:**

**Existing tests that will continue to pass:**
```go
func TestGenerateDuhWithFullFlagAndInitSpec(t *testing.T)
func TestGenerateDuhWithFullFlagAndCustomSpec(t *testing.T)
func TestFullGeneratedCodeFormat(t *testing.T)
```

These tests in `/internal/generate/duh/full_test.go` verify the "YOU CAN EDIT" marker is present but do not check the exact command text, so they will continue to pass without modification.

**Test Objectives:**
- Verify all generated files have correct header comments
- Verify timestamp is properly interpolated in all templates
- Verify "YOU CAN EDIT" notice is preserved

**Context for Implementation:**
- Templates use Go template syntax with `{{.Timestamp}}`
- Comment format must remain valid for each file type (Go vs Makefile syntax)
- Keep capitalization of "YOU CAN EDIT" for consistency
- Note: client.go.tmpl, server.go.tmpl, and iterator.go.tmpl already use `'duh generate'` without the subcommand

#### 4. Update Test Invocations
**Files** (12 test files total):
- `internal/generate/duh/full_test.go`
- `internal/generate/duh/generator_test.go`
- `internal/generate/duh/matcher_test.go`
- `internal/generate/duh/message_test.go`
- `internal/generate/duh/naming_test.go`
- `internal/generate/duh/parser_test.go`
- `internal/generate/duh/config_test.go`
- `internal/generate/duh/client_test.go`
- `internal/generate/duh/server_test.go`
- `internal/generate/duh/proto_test.go`
- `internal/generate/duh/iterator_test.go`
- `internal/generate/duh/integration_test.go`

**Changes**: Update all test invocations from `[]string{"generate", "duh", ...}` to `[]string{"generate", ...}`

**Patterns to find and replace:**

Find all these variations:
```go
[]string{"generate", "duh", "openapi.yaml"
[]string{"generate", "duh", filepath
[]string{"generate", "duh", specPath
args := []string{"generate", "duh",
RunCmd(&stdout, []string{"generate", "duh",
duh.RunCmd(&stdout, []string{"generate", "duh",
```

Replace pattern `"generate", "duh",` with `"generate",` in all cases

**Specific examples:**

**Example 1 - Variable assignment:**
```go
// Before:
args := []string{"generate", "duh", "openapi.yaml", "--full"}

// After:
args := []string{"generate", "openapi.yaml", "--full"}
```

**Example 2 - Inline invocation:**
```go
// Before:
exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})

// After:
exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})
```

**Example 3 - With filepath:**
```go
// Before:
[]string{"generate", "duh", filepath.Join(tempDir, "openapi.yaml")}

// After:
[]string{"generate", filepath.Join(tempDir, "openapi.yaml")}
```

**Implementation Approach:**
Use global search and replace in the `/internal/generate/duh/` directory:
- Search for: `"generate", "duh",`
- Replace with: `"generate",`
- This will catch all variations across all 12 test files

**Testing Requirements:**

After these changes, all existing tests should pass:
- All tests in `/internal/generate/duh/*_test.go` will continue to validate the same functionality
- The command being tested changes from `duh generate duh` to `duh generate`
- Test logic and assertions remain unchanged

**Test Objectives:**
- Verify all tests pass after updating command invocation
- Verify no regression in generation functionality
- Verify test coverage remains complete

**Context for Implementation:**
- Use search and replace to update all occurrences consistently
- Be careful to only update "generate", "duh" patterns, not other uses of "duh"
- Verify each file compiles after changes
- Run `go test ./internal/generate/duh/...` to verify all tests pass

#### 5. Add Test for Removed Command
**File**: `run_cmd_test.go`

**Changes**: Add a new test function to verify the `duh generate oapi` command has been removed

**Test to add:**

```go
func TestGenerateOapiCommandRemoved(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "oapi"})

	require.Equal(t, 2, exitCode)
	output := strings.ToLower(stdout.String())
	require.Contains(t, output, "unknown command")
}
```

**Testing Requirements:**

This is a new test that validates:
- The `oapi` subcommand no longer exists under `generate`
- Attempting to use it results in an error exit code (2)
- The error message indicates an unknown command

**Test Objectives:**
- Verify backward incompatibility is intentional
- Provide clear regression test for the removal
- Document expected behavior in test suite

**Context for Implementation:**
- Add this test function to the end of `run_cmd_test.go`
- Follow the existing test pattern in the file (use `duh.RunCmd` with `bytes.Buffer`)
- Use `require` for assertions (consistent with project guidelines)
- Use `strings.ToLower` for case-insensitive matching (consistent with existing tests like `TestRunCmdMultipleArguments`)

#### 6. Delete OAPI Implementation
**Directory**: `/internal/generate/oapi/`

**Changes**: Delete entire directory and all contents

**Files to be deleted:**
- `oapi.go` - Main entry point
- `client.go` - Client generation
- `server.go` - Server generation
- `models.go` - Models generation
- `generate.go` - Core generation logic
- `config.go` - Configuration setup
- `loader.go` - OpenAPI spec loader
- `oapi_test.go` - All tests

**Testing Requirements:**

No tests exist that depend on this package outside of the deleted `oapi_test.go`.

**Test Objectives:**
- Verify `go test ./...` passes after deletion
- Verify `go build ./...` succeeds after deletion
- Verify no import errors from deleted package

**Context for Implementation:**
- Use `rm -rf internal/generate/oapi` to delete directory
- Verify no other files import from this package (already confirmed only `run_cmd.go` imports it)

### Validation Commands

After completing all changes in Phase 1:

```bash
# Build the CLI
go build -o duh ./cmd/duh

# Verify the codebase builds without errors
go build ./...

# Verify no import errors or unused imports
go vet ./...

# Run all tests (critical - must pass before manual testing)
go test ./...

# Verify the CLI help works
./duh generate --help

# Test basic generation (requires valid openapi.yaml)
./duh generate openapi.yaml

# Test full generation
./duh generate openapi.yaml --full

# Verify old command is removed (should return error)
./duh generate oapi 2>&1 | grep -qi "unknown command"

# Verify duh subcommand is removed (should return error)
./duh generate duh 2>&1 | grep -qi "unknown command"

# Check for unused dependencies in go.mod
go mod tidy
git diff go.mod go.sum

# Verify oapi-codegen dependency was removed
grep -q "oapi-codegen" go.mod && echo "WARNING: oapi-codegen still in go.mod" || echo "✓ oapi-codegen removed"
```

**Expected Outcomes:**
- All builds succeed without errors (`go build ./...`)
- No import errors or unused imports (`go vet ./...`)
- All tests pass including the new `TestGenerateOapiCommandRemoved` (`go test ./...`)
- `duh generate --help` shows DUH-RPC generation help text
- `duh generate` executes successfully and generates files
- `duh generate oapi` returns error with "unknown command" message
- `duh generate duh` returns error with "unknown command" message
- No references to oapi command in help output
- `go mod tidy` removes unused dependencies:
  - `github.com/oapi-codegen/oapi-codegen/v2` (line 8 in current go.mod)
  - Any transitive dependencies only used by oapi package
- `git diff go.mod go.sum` shows the dependency removals

## Implementation Notes

### Breaking Changes
This is a **hard breaking change** with no backward compatibility:
- Users running `duh generate oapi` will receive an error
- Users running `duh generate duh` will receive an error (command no longer exists)
- Users must update to `duh generate`

### Risk Mitigation
- The core generation logic in `/internal/generate/duh/` is not modified
- All existing tests for DUH generation remain valid
- Changes are primarily structural (command routing) not functional

### Testing Strategy
- Rely on existing comprehensive test suite in `/internal/generate/duh/`
- Manually verify command help text and execution
- Validate README examples are accurate
