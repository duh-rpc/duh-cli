# OpenAPI-Proto Library Integration Plan

## Overview

Replace the mock stub `MockProtoConverter` implementation with the real `github.com/duh-rpc/openapi-proto.go` library to generate proper protobuf definitions from OpenAPI specifications.

## Current State Analysis

### Existing Implementation

The current codebase uses a mock implementation for proto generation:

**File**: `internal/generate/duh/converter.go`
- Defines `ProtoConverter` interface with `Convert(openapi []byte, packageName string) ([]byte, error)`
- `MockProtoConverter` generates stub proto3 files with empty message declarations
- Example output: `message CreateUserRequest {}` (no fields)

**Usage Location**: `run_cmd.go:232`
- Command handler instantiates `duh.NewMockProtoConverter()` for production use
- Passed to `duh.Run()` to generate proto files

**Test Coverage**: `internal/generate/duh/proto_test.go`
- Tests verify empty message structure only
- Check for syntax, package declaration, and message names
- Do not validate actual field generation

### openapi-proto.go Library

**Package**: `github.com/duh-rpc/openapi-proto.go`

**API**:
```go
func Convert(openapi []byte, packageName string) ([]byte, error)
```

**Capabilities**:
- Generates complete proto3 message definitions with fields
- Maps OpenAPI types to proto3 equivalents (string → string, integer → int32, etc.)
- Handles nested objects, enums, arrays (repeated fields)
- Adds `json_name` annotations for field mapping
- Produces compilable proto files

**Installation**:
```bash
go get github.com/duh-rpc/openapi-proto.go
```

### Key Discoveries

1. **Interface Compatibility**: The library's `Convert` function signature exactly matches the `ProtoConverter` interface
2. **Direct Replacement**: We can create a simple adapter that wraps the library function
3. **Single Usage Point**: Only one location instantiates the converter (`run_cmd.go:232`)
4. **Test Updates Needed**: Tests must verify message creation (not field details)

## Desired End State

After implementation:
1. ✅ `MockProtoConverter` completely removed from codebase
2. ✅ Real proto converter using `openapi-proto.go` library
3. ✅ Updated tests verifying message creation
4. ✅ Generated proto files contain actual field definitions
5. ✅ All existing tests pass with real proto generation

**Verification**:
```bash
make test          # All tests pass
make lint          # No lint errors
```

## What We're NOT Doing

1. ❌ Validating protobuf field generation details (library's responsibility)
2. ❌ Keeping `MockProtoConverter` for testing purposes
3. ❌ Adding custom error handling or validation before library calls
4. ❌ Modifying the `ProtoConverter` interface signature
5. ❌ Changing how proto files are written to disk (only changing content generation)

## Implementation Approach

**Strategy**: Direct replacement with thin adapter pattern

1. Add `openapi-proto.go` dependency to `go.mod`
2. Replace `MockProtoConverter` with `RealProtoConverter` that wraps library
3. Update single instantiation point in `run_cmd.go`
4. Update tests to verify message structure (not field details)
5. Remove all mock-related code

**Reasoning**: Since the library API matches our interface exactly, we need minimal adapter code. The implementation is straightforward with clear success criteria.

---

## Phase 1: Dependency Addition and Converter Replacement

### Overview
Add the openapi-proto.go library dependency and replace the mock converter implementation with a real implementation that delegates to the library.

### Changes Required

#### 1. Add Dependency
**File**: `go.mod`
**Changes**: Add openapi-proto.go library

**Manual Step** (run via Bash):
```bash
go get github.com/duh-rpc/openapi-proto.go@latest
```

**Note**: Library is published but not widely indexed. Use latest available release.

**Testing Requirements**:
```bash
# Verify dependency added
go mod tidy
```

**Test Objectives**:
- Dependency successfully added to go.mod
- go.sum updated with library checksums
- No conflicts with existing dependencies

---

#### 2. Replace Converter Implementation
**File**: `internal/generate/duh/converter.go`
**Changes**: Replace entire file with real converter implementation

```go
// ProtoConverter interface (unchanged)
type ProtoConverter interface {
    Convert(openapi []byte, packageName string) ([]byte, error)
}

// NewProtoConverter creates a new proto converter instance
func NewProtoConverter() ProtoConverter

// realProtoConverter wraps the openapi-proto.go library
type realProtoConverter struct{}

// Convert delegates to the openapi-proto.go library
func (r *realProtoConverter) Convert(openapi []byte, packageName string) ([]byte, error)
```

**Function Responsibilities**:
- `NewProtoConverter`: Create and return new converter instance
- `realProtoConverter.Convert`: Call `conv.Convert` from openapi-proto.go library and return result directly

**Testing Requirements**:
No unit tests for this file (follows project guidelines). Error cases tested via functional tests in `proto_test.go`.

**Context for implementation**:
- Import `conv "github.com/duh-rpc/openapi-proto.go"`
- Keep `ProtoConverter` interface unchanged
- Remove all `MockProtoConverter` code completely
- Remove `extractSchemaNames` and `generateProtoFile` helper functions
- `realProtoConverter` is unexported (lowercase)
- `NewProtoConverter` is exported and returns the interface type
- Library may return errors for invalid protobuf field names (see Phase 2 error tests)

---

#### 3. Update Command Handler
**File**: `run_cmd.go`
**Changes**: Update converter instantiation

**Old Code** (line 232):
```go
converter := duh.NewMockProtoConverter()
```

**New Code**:
```go
converter := duh.NewProtoConverter()
```

**Testing Requirements**:
Verified through functional tests in Phase 2.

**Context for implementation**:
- Single line change
- No other modifications to `run_cmd.go` needed
- Converter still passed to `duh.Run()` the same way

---

### Validation Commands

```bash
go mod tidy              # Verify dependencies clean
go build ./cmd/duh       # Verify code compiles
make test                # Run all tests (will fail until Phase 2 complete)
```

---

## Phase 2: Test Updates for Message Verification

### Overview
Update proto generation tests to verify that real proto files contain actual message definitions with fields, not just empty messages.

### Changes Required

#### 1. Update Proto File Structure Test
**File**: `internal/generate/duh/proto_test.go`
**Changes**: Update `TestProtoFileStructure` to verify messages contain fields

**Current Test Behavior** (lines 116-146):
- Checks for empty messages: `message CreateUserRequest {}`
- Counts exactly 3 messages

**Updated Test Behavior**:
- Verify messages exist (by name)
- Verify messages contain field declarations (not empty)
- Do NOT validate specific field names or types (library's responsibility)

**Testing Requirements**:
```go
func TestProtoFileStructure(t *testing.T)
```

**Test Objectives**:
- Proto file starts with `syntax = "proto3"`
- Contains correct package declaration
- Contains message declarations for all schemas
- Messages are NOT empty (contain at least one field)
- Each message appears exactly once

**Context for implementation**:
- Use existing test setup pattern from `TestGenerateDuhCreatesProtoFile`
- Use simple string contains check: verify absence of `{}` pattern immediately after message name
- Example check: `assert.NotContains(t, content, "message CreateUserRequest {}")`
- Check that message declarations exist: `assert.Contains(t, content, "message CreateUserRequest")`
- Follow pattern from existing assertion style in file

---

#### 2. Update Proto Creation Test
**File**: `internal/generate/duh/proto_test.go`
**Changes**: Update `TestGenerateDuhCreatesProtoFile` assertions

**Current Assertions** (lines 108-113):
- Checks for empty message strings

**Updated Assertions**:
- Verify message names exist in output
- Remove checks for empty message syntax `{}`
- Keep existing syntax and package checks

**Testing Requirements**:
```go
func TestGenerateDuhCreatesProtoFile(t *testing.T)
```

**Test Objectives**:
- Proto file created at expected path
- Contains syntax declaration
- Contains package declaration
- Contains message names from OpenAPI schemas
- File is valid proto3 format

**Context for implementation**:
- Keep file path and existence checks
- Update message assertions to not expect empty bodies
- Verify message names appear in content

---

#### 3. Update Schema Extraction Test
**File**: `internal/generate/duh/proto_test.go`
**Changes**: Update `TestProtoSchemaExtraction` to handle real messages

**Current Behavior** (lines 169-201):
- Checks for empty message syntax
- Verifies unique schema names

**Updated Behavior**:
- Verify all schema names appear as message declarations
- Ensure each schema generates exactly one message
- Do not check for empty message bodies

**Testing Requirements**:
```go
func TestProtoSchemaExtraction(t *testing.T)
```

**Test Objectives**:
- All OpenAPI schemas become proto messages
- Message names match schema names
- Each schema appears exactly once in proto file
- No duplicate message definitions

**Context for implementation**:
- Update message detection logic to not require `{}`
- Look for `message SchemaName {` pattern
- Keep uniqueness validation
- Remove empty message assertions

---

#### 4. Add Error Case Tests
**File**: `internal/generate/duh/proto_test.go`
**Changes**: Add new functional test for library error cases

**New Test**:
```go
func TestProtoGenerationErrorCases(t *testing.T)
```

**Test Objectives**:
- Test OpenAPI spec with invalid protobuf field names (e.g., fields starting with digits)
- Verify `duh generate duh` returns non-zero exit code
- Verify error message is shown in stdout
- Based on library documentation: field names must start with ASCII letter, not digit/underscore

**Context for implementation**:
- Create OpenAPI spec with schema containing field like `"1invalid": {type: "string"}`
- Use `duh.RunCmd()` functional test pattern from other tests
- Verify exit code is 2 (error condition)
- Verify stderr/stdout contains error indication
- Follow test pattern from existing tests in file

---

### Validation Commands

```bash
make test                           # All tests should pass
go test -v ./internal/generate/duh  # Verify proto tests specifically
```

---

## Phase 3: Cleanup and Documentation

### Overview
Remove references to mock implementation from documentation and verify complete integration.

### Changes Required

#### 1. Verify Plan Documentation References
**Files**: `plans/*.md`
**Changes**: Informational only - note that old plans referenced MockProtoConverter

**Manual Review**:
- Plans in `plans/` directory document the historical mock implementation
- These are historical records and should NOT be modified
- New plan (this file) supersedes old proto converter documentation

**Testing Requirements**:
None - documentation review only.

**Context for implementation**:
- No file changes needed
- Historical plans remain as-is for reference

---

#### 2. Final Integration Test
**Manual Testing**: End-to-end proto generation verification

**Test Steps**:
1. Create test OpenAPI spec with multiple schemas
2. Run `duh generate duh` command
3. Verify generated proto file contains:
   - Proper syntax declaration
   - Package name
   - Message definitions with fields
   - Valid proto3 syntax

**Testing Requirements**:
```bash
# Create test spec
cat > /tmp/test-spec.yaml << 'EOF'
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          description: Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
    UserResponse:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
        message:
          type: string
EOF

# Generate proto
./duh generate duh /tmp/test-spec.yaml --output-dir /tmp/test-output

# Verify proto file
cat /tmp/test-output/proto/v1/api.proto
```

**Test Objectives**:
- Command executes successfully
- Proto file generated at expected location
- File contains message definitions with actual fields
- Messages include CreateUserRequest, UserResponse, Error
- Fields have types and field numbers

**Context for implementation**:
- Use built binary: `./duh` or `make build` first
- Verify output matches expected proto3 format
- Check that fields from OpenAPI schemas appear in proto messages

---

### Validation Commands

```bash
make test          # All tests pass
make lint          # No lint errors
```

---

## Success Criteria

### Phase 1 Complete When:
- [ ] `github.com/duh-rpc/openapi-proto.go` added to go.mod
- [ ] `MockProtoConverter` removed from converter.go
- [ ] `NewProtoConverter()` creates real converter
- [ ] `run_cmd.go` uses `NewProtoConverter()` instead of `NewMockProtoConverter()`
- [ ] Code compiles without errors

### Phase 2 Complete When:
- [ ] All proto tests updated to verify non-empty messages
- [ ] Tests use simple string contains checks (not regex parsing)
- [ ] Error case test added for invalid protobuf field names
- [ ] All tests in `internal/generate/duh/proto_test.go` pass
- [ ] No test checks for empty message syntax `{}`

### Phase 3 Complete When:
- [ ] Manual integration test succeeds
- [ ] Generated proto files contain actual field definitions
- [ ] No references to `MockProtoConverter` in code (only in historical plans)

### Overall Success:
```bash
go run ./cmd/duh generate duh testdata/*.yaml  # Generates valid proto files
```

## Context References

**Existing Patterns to Follow**:
- Test setup: `internal/generate/duh/proto_test.go:93-95` (setupTest helper)
- File writing: `internal/generate/duh/duh.go:102-108` (writeFile function)
- Error handling: `internal/generate/duh/duh.go:82-85` (direct error passthrough)
- Test assertions: `internal/generate/duh/proto_test.go:109-113` (content verification)

**Integration Points**:
- Command handler: `run_cmd.go:202-246` (duh command definition)
- Main orchestrator: `internal/generate/duh/duh.go:12-100` (Run function)
- Proto interface: `internal/generate/duh/converter.go:11-13` (ProtoConverter interface)

**Dependencies**:
- OpenAPI loading: `internal/lint/lint.go` (Load function)
- OpenAPI validation: `internal/lint/lint.go` (Validate function)
