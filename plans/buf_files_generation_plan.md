# buf.yaml and buf.gen.yaml Generation Implementation Plan

## Overview

This plan adds buf.gen.yaml generation to `duh generate duh` command and ensures both buf configuration files are always generated (not just with --full flag). It also adds a reminder message to run `buf generate` and `go mod tidy` after generation completes, and fixes the Makefile output location.

## Current State Analysis

### What Exists Now

**buf.yaml Generation** (`internal/generate/duh/duh.go:95-108`):
- Currently ONLY generated when `--full` flag is used
- Template exists at `internal/generate/duh/templates/buf.yaml.tmpl`
- Written to `config.OutputDir` (correct location)
- Checks if file exists before generating (line 97)

**buf.gen.yaml Generation**:
- Does NOT exist yet
- Template needs to be created
- Should be generated EVERY time, not just with --full

**Makefile Generation** (`internal/generate/duh/duh.go:146-156`):
- Currently written to `"Makefile"` (root directory) - line 151
- Should be written to `config.OutputDir` instead
- Only generated with --full flag (correct behavior)

**Success Message** (`internal/generate/duh/duh.go:159-162`):
- Prints "✓ Generated N file(s) in {outputDir}"
- Lists all generated files
- No reminder message currently

### Key Discoveries

- Reference implementation at `/Users/thrawn/Development/duh-poc/buf.gen.yaml`
- Generator has `RenderBufYaml` method (generator.go:104-111)
- Templates loaded via embed.FS (embed.go:3-6)
- Makefile template references `buf generate` command (Makefile.tmpl:6)
- Tests expect specific file counts (full_integration_test.go:37)

## Desired End State

Running `duh generate duh openapi.yaml` will:

1. **Always generate buf files** (regardless of --full flag):
   - buf.yaml in `config.OutputDir`
   - buf.gen.yaml in `config.OutputDir`

2. **Fix Makefile location**:
   - Write to `filepath.Join(config.OutputDir, "Makefile")` instead of root

3. **Show reminder message**:
   ```
   ✓ Generated N file(s) in {outputDir}
     - server.go
     - client.go
     - ...

   Next steps:
     1. Run 'buf generate' to generate Go code from proto files
     2. Run 'go mod tidy' to update dependencies
   ```

### Verification

```bash
# Test without --full flag
mkdir test-buf && cd test-buf
go mod init github.com/test/example
duh init openapi.yaml
duh generate duh openapi.yaml
ls -la | grep buf  # Should see buf.yaml and buf.gen.yaml
cat buf.gen.yaml   # Should contain valid buf.gen.yaml config

# Test with --full flag
duh generate duh openapi.yaml --full
ls -la Makefile    # Should see Makefile in current directory (not root)
cat Makefile       # Should contain 'buf generate' command

# Verify output message
# Should see reminder about running buf generate and go mod tidy

# Run the commands
buf generate
go mod tidy
go test ./...
```

## What We're NOT Doing

- NOT adding customization options for buf.yaml or buf.gen.yaml content
- NOT executing `buf generate` automatically (user must run manually)
- NOT validating that buf CLI is installed
- NOT checking if buf.yaml/buf.gen.yaml already exist before overwriting
- NOT generating buf.lock or other buf-related files

## Implementation Approach

1. **Create buf.gen.yaml Template**: Copy content from duh-poc reference implementation
2. **Add Generator Method**: Follow existing `RenderBufYaml` pattern
3. **Reorganize Generation Flow**: Move buf file generation outside --full block
4. **Fix File Paths**: Ensure all files go to correct locations
5. **Add Reminder Message**: Simple fmt.Fprintf after file list
6. **Update Tests**: Adjust file count expectations

## Phase 1: Create buf.gen.yaml Template and Generator Method

### Overview
Create the buf.gen.yaml template file and add the render method to the Generator, following the existing pattern for buf.yaml.

### Changes Required

#### 1. buf.gen.yaml Template

**File**: `internal/generate/duh/templates/buf.gen.yaml.tmpl` (new file)

**Template Structure:**
```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: .
    opt:
      - paths=source_relative
  - remote: buf.build/grpc/go
    out: .
    opt:
      - paths=source_relative
```

**Context for Implementation:**
- Copy exact content from `/Users/thrawn/Development/duh-poc/buf.gen.yaml`
- No template variables needed (static content)
- No conditional logic required
- This is a YAML file, not Go code

#### 2. Add Render Method to Generator

**File**: `internal/generate/duh/generator.go`

**Changes**: Add render method for buf.gen.yaml

```go
func (g *Generator) RenderBufGenYaml(data *TemplateData) ([]byte, error)
```

**Function Responsibilities:**
- Execute template: `g.templates.ExecuteTemplate(&buf, "buf.gen.yaml.tmpl", data)`
- Return raw bytes (no formatting needed - it's YAML, not Go)
- Return any template execution errors
- Follow exact pattern from `RenderBufYaml` (generator.go:104-111)

**Implementation Pattern** (copy from RenderBufYaml):
```go
func (g *Generator) RenderBufYaml(data *TemplateData) ([]byte, error) {
    var buf bytes.Buffer
    if err := g.templates.ExecuteTemplate(&buf, "buf.yaml.tmpl", data); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
```

**Note**: Although the buf.gen.yaml template is static, we still pass `data *TemplateData` to maintain a consistent interface across all Render methods

**Context for Implementation:**
- Place method after `RenderBufYaml` method (around line 112)
- Templates automatically loaded via embed.FS in `NewGenerator()` (line 16)
- No need to set timestamp (buf.gen.yaml doesn't use it)
- Verify embed directive: Check that `/Users/thrawn/Development/duh-cli/internal/generate/duh/embed.go` includes pattern matching `*.tmpl` files

### Testing Requirements

**File**: `internal/generate/duh/generator_test.go` (new file or add to existing)

**New Tests:**

```go
func TestRenderBufGenYaml(t *testing.T)
```

**Test Objectives:**
- Verify template renders without errors
- Verify output contains expected YAML structure
- Verify plugins section is present
- Verify version: v2 is present

**Context for Implementation:**
- Check if `generator_test.go` already exists in `/Users/thrawn/Development/duh-cli/internal/generate/duh/`
- If it exists, add test to existing file rather than creating new file
- If it doesn't exist, create new file
- Create minimal TemplateData struct
- Call `generator.RenderBufGenYaml(data)`
- Use `strings.Contains()` to verify content
- Check for "buf.build/protocolbuffers/go" and "buf.build/grpc/go"

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestRenderBufGenYaml
```

## Phase 2: Reorganize Generation Flow and Fix File Paths

### Overview
Move buf.yaml and buf.gen.yaml generation outside the --full flag block so they're always generated, and fix the Makefile path to write to config.OutputDir.

### Changes Required

#### 1. Reorganize buf File Generation

**File**: `internal/generate/duh/duh.go`

**Changes**: Move buf file generation outside --full block

```go
func Run(config RunConfig) error
```

**Function Responsibilities:**
- After proto file generation (line 93), ALWAYS generate buf files:
  - Generate buf.yaml: `bufYamlCode, err := generator.RenderBufYaml(data)`
  - Write to: `filepath.Join(config.OutputDir, "buf.yaml")`
  - Generate buf.gen.yaml: `bufGenYamlCode, err := generator.RenderBufGenYaml(data)`
  - Write to: `filepath.Join(config.OutputDir, "buf.gen.yaml")`
  - Add both to `filesGenerated` list
- Remove buf.yaml generation from --full block (delete lines 96-108)
- **Delete the existence check** (lines 97-98: `if _, err := os.Stat(bufYamlPath); os.IsNotExist(err)`) so buf.yaml is always overwritten, matching the behavior of other generated files
- Follow existing error handling pattern
- No new imports needed - all necessary packages (os, filepath, fmt) are already imported

**Context for Implementation:**
- Current proto generation at lines 78-93
- Current buf.yaml generation at lines 96-108 (will be moved)
- Use existing `writeFile()` helper (line 167-173)
- Maintain file generation order: proto → buf files → conditional --full files

#### 2. Fix Makefile Path

**File**: `internal/generate/duh/duh.go`

**Changes**: Update Makefile path to use config.OutputDir

**Current code** (line 151):
```go
makefilePath := "Makefile"
```

**Updated code**:
```go
makefilePath := filepath.Join(config.OutputDir, "Makefile")
```

**Function Responsibilities:**
- Change Makefile path from root directory to config.OutputDir
- Maintain all other Makefile generation logic
- Keep Makefile generation inside --full block

**Context for Implementation:**
- Change occurs at line 151 in duh.go
- Makefile generation is lines 146-156
- This ensures all generated files go to the same directory

### Testing Requirements

**File**: `internal/generate/duh/buf_generation_test.go` (new file)

**New Tests:**

```go
func TestBufFilesGeneratedWithoutFullFlag(t *testing.T)
func TestBufFilesGeneratedWithFullFlag(t *testing.T)
func TestBufFilesAlwaysOverwrite(t *testing.T)
func TestMakefileWrittenToOutputDir(t *testing.T)
```

**Test Objectives:**
- Verify buf.yaml and buf.gen.yaml generated without --full flag
- Verify buf files still generated with --full flag
- Verify buf files overwrite existing files (no existence check)
- Verify Makefile written to config.OutputDir, not root
- Verify all files in same directory when using custom output-dir

**Context for Implementation:**
- Follow functional testing pattern from `full_integration_test.go`
- Use `duh.RunCmd(&stdout, []string{"generate", "duh", specPath})`
- Check file existence with `os.Stat()`
- Verify file location with `filepath.Join(tempDir, "buf.yaml")`
- Create pre-existing buf.yaml to test overwrite behavior

**Existing Tests to Update:**

```go
// File: internal/generate/duh/full_integration_test.go
func TestGenerateDuhWithFullFlagAndInitSpec(t *testing.T)  // Update: expect 10 files (was 9)
func TestGenerateDuhWithFullFlagAndCustomSpec(t *testing.T)  // Update: expect 8 files (was 7)
func TestGenerateDuhWithoutFullFlag(t *testing.T)  // Update: expect 5 files (was 4)
```

**Test Update Requirements:**

**File Count Reference** (initTemplateSpec has list operations, so iterator.go is generated):

| Scenario | Current Count | New Count | Files |
|----------|--------------|-----------|-------|
| Without --full | 4 files | 6 files | server.go, client.go, iterator.go, proto/v1/api.proto, buf.yaml, buf.gen.yaml |
| With --full | 9 files | 10 files | (above 6) + daemon.go, service.go, api_test.go, Makefile |

**Tests to Update:**
- `TestGenerateDuhWithFullFlagAndInitSpec`: Change from "Generated 9 file(s)" to "Generated 10 file(s)"
- `TestGenerateDuhWithoutFullFlag`: Change from "Generated 4 file(s)" to "Generated 6 file(s)"
- Any other tests using `initTemplateSpec`: Add 1 to expected file count (for buf.gen.yaml)

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestBufFilesGenerated
go test ./internal/generate/duh -v -run TestMakefileWrittenToOutputDir
go test ./internal/generate/duh -v -run TestGenerateDuhWith  # Run all integration tests
```

## Phase 3: Add Reminder Message

### Overview
Add a simple informational message after the file list reminding users to run `buf generate` and `go mod tidy`.

### Changes Required

#### 1. Add Reminder Message

**File**: `internal/generate/duh/duh.go`

**Changes**: Add message after file list

```go
func Run(config RunConfig) error
```

**Function Responsibilities:**
- After file list (currently line 162), add:
  ```go
  _, _ = fmt.Fprintf(config.Writer, "\nNext steps:\n")
  _, _ = fmt.Fprintf(config.Writer, "  1. Run 'buf generate' to generate Go code from proto files\n")
  _, _ = fmt.Fprintf(config.Writer, "  2. Run 'go mod tidy' to update dependencies\n")
  ```
- Use same `config.Writer` as existing output
- Add blank line before "Next steps:" for visual separation
- Use numbered list for clarity
- Follow existing output pattern with fmt.Fprintf

**Context for Implementation:**
- Current success message at lines 159-162
- Add reminder after the file list loop
- Before the `return nil` statement (line 164)
- Use `_, _` to ignore fprintf return values (existing pattern)

### Testing Requirements

**File**: `internal/generate/duh/message_test.go` (new file)

**New Tests:**

```go
func TestReminderMessageDisplayed(t *testing.T)
func TestReminderMessageWithFullFlag(t *testing.T)
func TestReminderMessageFormat(t *testing.T)
```

**Test Objectives:**
- Verify reminder message appears in output
- Verify message contains exact strings to avoid false positives:
  ```go
  assert.Contains(t, stdout.String(), "Next steps:")
  assert.Contains(t, stdout.String(), "buf generate")
  assert.Contains(t, stdout.String(), "go mod tidy")
  ```
- Verify message appears with and without --full flag
- Verify message formatting (numbered list, blank line before "Next steps:")

**Context for Implementation:**
- Capture stdout in bytes.Buffer
- Use `strings.Contains()` to check for message parts
- Check for "Next steps:" header
- Check for both step 1 and step 2
- Run test with and without --full flag

**Existing Tests to Update:**

```go
// File: internal/generate/duh/full_integration_test.go
// All tests that check stdout output may need updates:
func TestGenerateDuhWithFullFlagAndInitSpec(t *testing.T)  // May need: verify reminder message present
func TestGenerateDuhWithoutFullFlag(t *testing.T)  // May need: verify reminder message present
```

**Test Update Requirements:**
- Tests should continue to pass with new message
- Optionally add assertions to verify reminder message presence
- Ensure existing "Generated N file(s)" check still works

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestReminderMessage
go test ./internal/generate/duh -v  # Run all tests to ensure no regressions
```

## Phase 4: End-to-End Testing and Documentation

### Overview
Comprehensive manual testing and CLI help text updates to document the new behavior.

### Changes Required

#### 1. CLI Help Text Update

**File**: `/Users/thrawn/Development/duh-cli/run_cmd.go`

**Changes**: Update duhCmd.Long description

**Current description** (lines 205-227):
```
By default, generates client.go, server.go, iterator.go (if list operations),
and proto file. Use flags to customize output.
```

**Updated description**:
```
By default, generates client.go, server.go, iterator.go (if list operations),
proto file, buf.yaml, and buf.gen.yaml. Use flags to customize output.

After generation, run 'buf generate' to generate Go code from proto files,
then run 'go mod tidy' to update dependencies.
```

**Function Responsibilities:**
- Add buf.yaml and buf.gen.yaml to default file list
- Add reminder about post-generation steps
- Keep all other help text unchanged
- Maintain existing structure and formatting

**Context for Implementation:**
- Update occurs at lines 210-211 (approximately)
- Keep --full flag documentation (lines 213-221)
- Maintain exit codes section

### Testing Requirements

**File**: Manual testing checklist

**Test Scenarios:**

1. **Basic generation without --full:**
```bash
mkdir test-basic && cd test-basic
go mod init github.com/test/basic
duh init openapi.yaml
duh generate duh openapi.yaml

# Verify output
# Should see: "✓ Generated 6 file(s) in ."
# Should list: server.go, client.go, iterator.go, proto/v1/api.proto, buf.yaml, buf.gen.yaml
# Should see reminder: "Next steps:" with buf generate and go mod tidy

# Verify files exist
ls -la | grep -E "(buf\.yaml|buf\.gen\.yaml)"
test -f buf.yaml && echo "✓ buf.yaml exists"
test -f buf.gen.yaml && echo "✓ buf.gen.yaml exists"

# Verify content
grep "version: v2" buf.yaml
grep "buf.build/protocolbuffers/go" buf.gen.yaml

# Run suggested commands
buf generate
go mod tidy
go test ./...
```

2. **Generation with --full flag:**
```bash
mkdir test-full && cd test-full
go mod init github.com/test/full
duh init openapi.yaml
duh generate duh openapi.yaml --full

# Verify output
# Should see: "✓ Generated 10 file(s) in ."
# Should include all files: server, client, iterator, proto, buf.yaml, buf.gen.yaml, daemon, service, api_test, Makefile

# Verify Makefile location
test -f Makefile && echo "✓ Makefile in current directory"
grep "buf generate" Makefile

# Run full build
make proto
make test
```

3. **Custom output directory:**
```bash
mkdir test-custom && cd test-custom
go mod init github.com/test/custom
duh init openapi.yaml
mkdir api
duh generate duh openapi.yaml -o api --full

# Verify all files in api/ directory
test -f api/buf.yaml && echo "✓ buf.yaml in api/"
test -f api/buf.gen.yaml && echo "✓ buf.gen.yaml in api/"
test -f api/Makefile && echo "✓ Makefile in api/"
test -f api/daemon.go && echo "✓ daemon.go in api/"
```

4. **Overwrite behavior:**
```bash
mkdir test-overwrite && cd test-overwrite
go mod init github.com/test/overwrite
duh init openapi.yaml

# First generation
duh generate duh openapi.yaml
echo "# CUSTOM CONTENT" >> buf.yaml

# Second generation
duh generate duh openapi.yaml

# Verify overwrite
grep "CUSTOM CONTENT" buf.yaml && echo "✗ File NOT overwritten" || echo "✓ File overwritten"
```

### Validation Commands

```bash
# Run all duh generate tests
go test ./internal/generate/duh -v

# Build entire project
go build ./...

# Run all project tests
go test ./...

# Test CLI help
duh generate duh --help | grep "buf.yaml"
duh generate duh --help | grep "buf generate"
```

## Success Criteria

### Functional Requirements
- ✅ buf.gen.yaml template created
- ✅ buf.yaml and buf.gen.yaml generated EVERY time (not just with --full)
- ✅ Both buf files written to config.OutputDir
- ✅ Makefile written to config.OutputDir (not root)
- ✅ Reminder message displayed after file list
- ✅ Message includes both "buf generate" and "go mod tidy"
- ✅ All generated files compile successfully
- ✅ buf generate command works with generated files

### Code Quality
- ✅ All tests pass: `go test ./internal/generate/duh -v`
- ✅ File count tests updated correctly
- ✅ No linting errors: `make lint`
- ✅ Follows existing code patterns
- ✅ CLAUDE.md guidelines followed

### Documentation
- ✅ CLI help text updated
- ✅ Implementation plan completed

## Testing Strategy

### Test Organization

**Location**: `internal/generate/duh/` alongside implementation

**Test Files:**
- `generator_test.go` - Template rendering for buf.gen.yaml
- `buf_generation_test.go` - Buf file generation tests
- `message_test.go` - Reminder message tests
- `full_integration_test.go` (updates) - Update file count expectations

**Testing Style:**
- ALL tests MUST be functional (per CLAUDE.md)
- Call `duh.RunCmd()` to test through CLI interface
- Verify exit codes and stdout output
- Check generated file contents with file reads
- Each test function stands on its own

**Example Test Pattern:**
```go
func TestBufFilesGeneratedWithoutFullFlag(t *testing.T) {
    tempDir := t.TempDir()
    specPath := filepath.Join(tempDir, "openapi.yaml")

    require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
    require.NoError(t, os.WriteFile(
        filepath.Join(tempDir, "go.mod"),
        []byte("module github.com/test/example\n\ngo 1.24\n"),
        0644,
    ))

    originalDir, err := os.Getwd()
    require.NoError(t, err)
    defer func() { _ = os.Chdir(originalDir) }()
    require.NoError(t, os.Chdir(tempDir))

    var stdout bytes.Buffer
    exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", "openapi.yaml"})

    require.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "Generated 6 file(s)")  // 6 files because initTemplateSpec has list ops

    // Verify buf.yaml exists
    _, err = os.Stat("buf.yaml")
    require.NoError(t, err)

    // Verify buf.gen.yaml exists
    _, err = os.Stat("buf.gen.yaml")
    require.NoError(t, err)

    // Verify content
    bufGenContent, err := os.ReadFile("buf.gen.yaml")
    require.NoError(t, err)
    assert.Contains(t, string(bufGenContent), "buf.build/protocolbuffers/go")
}

func TestBufFilesAlwaysOverwrite(t *testing.T) {
    tempDir := t.TempDir()
    specPath := filepath.Join(tempDir, "openapi.yaml")

    require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
    require.NoError(t, os.WriteFile(
        filepath.Join(tempDir, "go.mod"),
        []byte("module github.com/test/example\n\ngo 1.24\n"),
        0644,
    ))

    originalDir, err := os.Getwd()
    require.NoError(t, err)
    defer func() { _ = os.Chdir(originalDir) }()
    require.NoError(t, os.Chdir(tempDir))

    // Create buf.yaml with custom content
    require.NoError(t, os.WriteFile("buf.yaml", []byte("# CUSTOM CONTENT\n"), 0644))

    var stdout bytes.Buffer
    exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", "openapi.yaml"})
    require.Equal(t, 0, exitCode)

    // Verify buf.yaml was overwritten (custom content should be gone)
    content, err := os.ReadFile("buf.yaml")
    require.NoError(t, err)
    assert.NotContains(t, string(content), "# CUSTOM CONTENT")
    assert.Contains(t, string(content), "version: v2")
}
```

## References

### Key Files to Reference

**Current Implementation:**
- `internal/generate/duh/duh.go:95-108` - Current buf.yaml generation (in --full block)
- `internal/generate/duh/duh.go:146-156` - Current Makefile generation
- `internal/generate/duh/duh.go:159-162` - Current success message
- `internal/generate/duh/generator.go:104-111` - RenderBufYaml method
- `internal/generate/duh/templates/buf.yaml.tmpl` - Existing buf.yaml template

**Reference Files:**
- `/Users/thrawn/Development/duh-poc/buf.yaml` - Source for buf.yaml
- `/Users/thrawn/Development/duh-poc/buf.gen.yaml` - Source for buf.gen.yaml

**Testing Patterns:**
- `internal/generate/duh/full_integration_test.go` - Integration test examples
- `CLAUDE.md` - Testing guidelines

### Important Patterns

**File Generation Pattern:**
```go
// Generate file
fileCode, err := generator.RenderSomething(data)
if err != nil {
    return fmt.Errorf("failed to render something: %w", err)
}

// Write file
filePath := filepath.Join(config.OutputDir, "filename")
if err := writeFile(filePath, fileCode); err != nil {
    return fmt.Errorf("failed to write file: %w", err)
}

// Track generated file
filesGenerated = append(filesGenerated, "filename")
```

**Output Message Pattern:**
```go
_, _ = fmt.Fprintf(config.Writer, "✓ Generated %d file(s) in %s\n", len(filesGenerated), config.OutputDir)
for _, file := range filesGenerated {
    _, _ = fmt.Fprintf(config.Writer, "  - %s\n", file)
}
```
