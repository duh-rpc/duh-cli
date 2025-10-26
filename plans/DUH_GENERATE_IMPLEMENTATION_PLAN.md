# DUH-RPC Code Generator Implementation Plan

## Overview

This plan details the implementation of `duh generate duh` - a template-based code generator that transforms DUH-RPC compliant OpenAPI specifications into production-ready Go client and server code with built-in pagination support.

**Business Value:** Enables developers to maintain a single source of truth (OpenAPI spec) and automatically generate consistent, type-safe client/server implementations, eliminating manual code duplication.

## Current State Analysis

### Existing Patterns Discovered

**Command Structure** (run_cmd.go:135-201):
- Commands follow cobra pattern with subcommands under main command groups
- Exit codes: 0 = success, 2 = error (including validation failures)
- Flags use cobra's StringP/String methods with defaults
- Success messages use `✓` checkmark pattern

**Generate Pattern** (internal/generate/oapi/):
- Package structure: `oapi.go` exports main `RunOapi()` function
- Delegates to helper functions: Load(), Generate(), NewConfig()
- Creates output directories if they don't exist
- Writes success message to io.Writer

**Lint Integration** (internal/lint/):
- `lint.Load(filePath)` returns `*v3.Document` from libopenapi
- `lint.Validate(doc, filePath)` returns ValidationResult
- `lint.Print(w, result)` outputs formatted validation results
- ValidationResult has `Valid()` method returning bool

**Template Embedding** (internal/init/generator.go:3-6):
```go
import _ "embed"
//go:embed template/openapi.yaml
var Template []byte
```

**Testing Pattern** (lint_test.go, oapi_test.go):
- Functional tests use `duh.RunCmd(&stdout, []string{...})`
- Test package suffix: `package XXX_test`
- Inline specs as const strings for short specs
- Temp directories via `t.TempDir()`
- File verification with `os.Stat()` and content checks

### Reference Files from duh-poc

**server.go** (137 lines):
- RPC path constants at top (e.g., `RPCCreateUser = "/v1/users.create"`)
- `ServiceInterface` with method signatures + Shutdown method
- `MiddlewareFunc` type and `HandlerConfig` struct
- `NewHandler()` applies middlewares in order
- `Handler.ServeHTTP()` validates POST method, switch/case routing
- Individual handler methods: read request, call service, reply

**iterator.go** (83 lines):
- Generic `Page[T]` struct with Items, Total, Page, PerPage, TotalPages
- `Iterator[T]` interface with Next() and Err()
- `PageFetcher[T]` interface with FetchPage()
- `GenericIterator[T]` implementation with context cancellation
- Not generated with build tags

**client.go** (218 lines):
- `ClientInterface` with RPC methods + Close()
- `ClientConfig` struct with Client and Endpoint fields
- `NewClient()` constructor with validation
- RPC methods: marshal request, create HTTP request, set Content-Type, call client.Do()
- Iterator support: operation-specific PageFetcher struct (e.g., `UserPageFetcher`)
- Iterator method (e.g., `ListUsersIter()`) creates fetcher and returns GenericIterator
- `WithTLS()` and `WithNoTLS()` helper functions with hardcoded connection pool sizes

### Key Constraints and Clarifications

- **No atomic generation**: If error occurs, abort without cleanup (leave partial files on disk for debugging)
- **No build tags**: Reference files have `//go:build go1.24` but spec requirement overrides - don't include
- **List detection**: Check method portion of path (after dot) for "list" - e.g., `/v1/users.list` and `/v1/users.list-active` both match
- **Module path**: Single-line `module <path>`, ignore comments, validate contains "/" character
- **Proto mock**: Empty message declarations with package and syntax - used in PRODUCTION until real converter ready
- **Test specs**: Inline (const strings) unless >30 lines
- **Exit codes**: 0 = success, 2 = all errors (including validation failures)
- **Timestamp format**: Human-readable "2006-01-02 15:04:05 UTC" (not RFC3339)
- **String capitalization**: Use `strings.ToUpper(s[:1]) + s[1:]` for ASCII (strings.Title is deprecated)
- **Array field selection**: First array in YAML definition order via libopenapi's ordered property map
- **Schema handling**: Only $ref schemas (error on inline, skip operations without $ref)

## Desired End State

After implementation, users can:
1. Run `duh generate duh openapi.yaml` to generate client.go, server.go, iterator.go (conditional), and proto file
2. Generated code compiles with `go build` and passes `go vet`
3. Code follows DUH-RPC patterns from reference files
4. Pagination iterators automatically generated for list operations
5. Proto import paths correctly detected from go.mod

**Verification:**
- `go build` succeeds in output directory
- Generated files match structure of reference files
- RPC constants match OpenAPI paths
- Iterator only generated when list operations exist

## What We're NOT Doing

- Custom template overrides by users
- Configuration file support (duh.yaml)
- Incremental generation (only changed operations)
- Source maps (template line → output line)
- Real ProtoConverter implementation (using mock)
- TypeScript or Python generators
- Atomic file generation with rollback
- Build tag directives in generated code

## Implementation Approach

**Strategy:** Build incrementally in testable phases, starting with foundation (config, naming, parser), then server template (simplest), iterator template (independent), client template (uses iterator), and finally integration.

**Template Development:** Model templates after reference files, using embedded text/template with go:embed, render with structured TemplateData.

**Testing:** Each phase fully tested before proceeding to next phase using functional e2e tests through CLI.

### Key Implementation Patterns (Established in Phase 1 & 6)

**Barebones Orchestration:**
- Create `duh.Run()` early that only uses implemented components
- Expand `duh.Run()` incrementally as each phase completes
- Phase 1 & 6: `duh.Run()` validates, parses, and outputs summary
- Phase 2-5: Add file generation to `duh.Run()` incrementally

**Functional Testing Only:**
- **NO unit tests** - all tests call `duh.RunCmd()` through CLI
- Tests verify file creation, compilation, and structure via CLI
- Use `setupTest(t, spec)` helper for common test setup
- Specs must be valid DUH-RPC compliant with proper error schemas
- Tests located in `internal/generate/duh/generate_test.go`

**Example Test Pattern:**
```go
func TestGenerateDuhCreatesServerFile(t *testing.T) {
    specPath, stdout := setupTest(t, validSpec)
    exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

    require.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "✓")

    _, err := os.Stat(filepath.Join(tempDir, "server.go"))
    require.NoError(t, err)
}
```

**Benefits of This Approach:**
- Tests verify actual CLI behavior users will experience
- Tests catch integration issues early
- No mocking needed - tests use real components
- Each phase adds tests as features are implemented
- Command works end-to-end from Phase 6 onward

---

## Phase 1: Foundation (Config, Naming, Parser)

### Overview
Establish core infrastructure for configuration management, operation name generation, and OpenAPI parsing. This phase has no code generation - only data extraction and validation.

### Changes Required

#### 1. Configuration Management
**File**: `internal/generate/duh/config.go`

```go
type Config struct {
    PackageName  string
    OutputDir    string
    ProtoPath    string
    ProtoImport  string
    ProtoPackage string
}

func NewConfig(flags ConfigFlags) (*Config, error)
func (c *Config) Validate() error
func (c *Config) DetectModulePath() (string, error)
func (c *Config) ConstructProtoImport(modulePath string) string
func (c *Config) DeriveProtoPackage(spec *v3.Document) string
```

**Function Responsibilities:**
- `NewConfig`: Load config from flags, apply defaults (package="api", outputDir=".", protoPath="proto/v1/api.proto")
- `Validate`: Check package name is valid Go identifier, not "main", output dir exists
- `DetectModulePath`: Read go.mod, extract module path from first line matching `^module (.+)$`, validate contains "/"
- `ConstructProtoImport`: If ProtoImport set, return it; else join modulePath + dirname(ProtoPath)
- `DeriveProtoPackage`: If ProtoPackage set, return it; else extract version from first path, return "duh.api.{version}"

**Context:**
- Follow pattern from internal/generate/oapi/config.go
- Use `os.ReadFile("go.mod")` for module detection
- Use `filepath.Dir()` for path manipulation

#### 2. Operation Naming Algorithm
**File**: `internal/generate/duh/naming.go`

```go
func GenerateOperationName(path string) (string, error)
func GenerateConstName(operationName string) string
func parseSubjectMethod(path string) (subject, method string, err error)
func toCamelCase(s string) string
```

**Function Responsibilities:**
- `GenerateOperationName`: Extract {subject}.{method} from path, split on '.', convert each to camel case, concatenate
- `GenerateConstName`: Prepend "RPC" to operation name
- `parseSubjectMethod`: Extract portion after version (e.g., "/v1/users.create" → "users.create")
- `toCamelCase`: Split on '-' or '_', capitalize first letter of each word using simple ASCII conversion (not strings.Title which is deprecated), join (e.g., "user-profiles" → "UserProfiles")

**Context:**
- Standard camel case: Id, Http, Url (not ID, HTTP, URL)
- No special acronym handling
- Path format: /v{N}/{subject}.{method}
- Use simple ASCII capitalization: `strings.ToUpper(s[:1]) + s[1:]` instead of deprecated strings.Title

#### 3. OpenAPI Parser
**File**: `internal/generate/duh/parser.go`

```go
type Parser struct {
    spec   *v3.Document
    config Config
}

func NewParser(spec *v3.Document, config Config) *Parser
func (p *Parser) Parse() (*TemplateData, error)
func (p *Parser) extractOperations() ([]Operation, error)
func (p *Parser) detectListOperations(ops []Operation) ([]ListOperation, error)
func (p *Parser) isListOperation(path string, requestSchema, responseSchema *base.SchemaProxy) bool
func (p *Parser) findFirstArrayField(schema *base.SchemaProxy) (fieldName, itemType string, found bool)
```

**Function Responsibilities:**
- `Parse`: Orchestrate parsing, populate TemplateData with all fields
- `extractOperations`: Iterate paths, extract method, path, summary, request/response types (only $ref, prefix with "pb."). Error if schemas are inline (not $ref)
- `detectListOperations`: Filter operations, test 3-criteria, build ListOperation structs
- `isListOperation`: Check (1) path method contains "list", (2) request has "offset", (3) response has array field
- `findFirstArrayField`: Iterate response schema properties in YAML definition order (as preserved by libopenapi's ordered map), return first array field name + item type

**Context:**
- Use libopenapi's v3.Document structure (same as internal/lint/loader.go)
- Only extract top-level schema references via $ref (error on inline schemas, skip operations without $ref)
- List detection: path portion after dot (e.g., "/v1/users.list" and "/v1/users.list-active" both match)
- Array field selection: first array field in YAML definition order (deterministic via libopenapi's ordered property iteration)

#### 4. Template Data Model
**File**: `internal/generate/duh/types.go`

```go
type TemplateData struct {
    Package      string
    ModulePath   string
    ProtoImport  string
    ProtoPackage string
    Operations   []Operation
    ListOps      []ListOperation
    HasListOps   bool
    Timestamp    string
}

type Operation struct {
    MethodName   string
    Path         string
    ConstName    string
    Summary      string
    RequestType  string
    ResponseType string
}

type ListOperation struct {
    Operation
    IteratorName  string
    FetcherName   string
    ItemType      string
    ResponseField string
}
```

### Testing Requirements

**File**: `internal/generate/duh/config_test.go`

```go
func TestNewConfigWithDefaults(t *testing.T)
func TestValidatePackageName(t *testing.T)
func TestDetectModulePath(t *testing.T)
func TestConstructProtoImport(t *testing.T)
```

**Test Objectives:**
- Verify default values applied correctly
- Reject "main" package name
- Extract module path from go.mod (single line, with/without comments)
- Construct proto import from module path + proto path dir
- Override with --proto-import flag

**File**: `internal/generate/duh/naming_test.go`

```go
func TestGenerateOperationName(t *testing.T)
func TestGenerateConstName(t *testing.T)
func TestToCamelCase(t *testing.T)
```

**Test Objectives:**
- Test all examples from spec: users.create → UsersCreate, user-profiles.get-by-id → UserProfilesGetById
- Verify RPC prefix added correctly
- Handle hyphens, underscores, mixed separators

**File**: `internal/generate/duh/parser_test.go`

```go
func TestParseOperations(t *testing.T)
func TestDetectListOperations(t *testing.T)
func TestIsListOperation(t *testing.T)
func TestFindFirstArrayField(t *testing.T)
```

**Test Objectives:**
- Parse valid spec with multiple operations
- Extract request/response types as pb-prefixed
- Detect list operations using 3-criteria test
- Find first array field in YAML definition order
- Verify "list" in method portion of path (not subject)
- Test error handling for inline schemas (should error)
- Test handling of missing $ref (should skip operation or error)

**Context:**
- Use inline OpenAPI specs (const strings) for tests <30 lines
- Follow functional testing pattern: create temp dir, write go.mod, parse spec
- Use require/assert from testify

### Validation Commands

```bash
go test ./internal/generate/duh -v
```

---

## Phase 2: Server Template Generation

### Overview
Create server.go template and generator infrastructure. This is the simplest template (no dependencies on iterator). Establishes template embedding pattern and rendering pipeline.

**Implementation Note:** Update `duh.Run()` to generate server.go. Tests will drive implementation by calling CLI and verifying file creation/compilation.

### Changes Required

#### 1. Template File
**File**: `internal/generate/duh/templates/server.go.tmpl`

**Template Structure** (based on server.go reference):
- File header: generation comment (no build tag)
- Package declaration
- Imports (context, fmt, net/http, duh-go, proto)
- RPC path constants
- ServiceInterface definition
- MiddlewareFunc type
- HandlerConfig struct
- NewHandler function
- Handler struct
- ServeHTTP method with POST validation + switch/case routing
- Individual handler methods (handleCreateUser, etc.)

**Template Data Access:**
- `.Package` - package name
- `.ProtoImport` - proto import path
- `.Timestamp` - generation timestamp
- `.Operations` - slice of Operation structs
- Template loops over Operations to generate constants, interface methods, cases, handlers

**Context:**
- Model after /Users/thrawn/Development/duh-poc/server.go
- Use `{{range .Operations}}` for iteration
- Include operation summary as comment
- Hard-code request size limit: `5*duh.MegaByte`

#### 2. Generator Infrastructure
**File**: `internal/generate/duh/generator.go`

```go
type Generator struct {
    templates *template.Template
    timestamp string
}

func NewGenerator() (*Generator, error)
func (g *Generator) RenderServer(data *TemplateData) ([]byte, error)
func (g *Generator) FormatCode(code []byte) ([]byte, error)
func (g *Generator) generateTimestamp() string
```

**Function Responsibilities:**
- `NewGenerator`: Parse embedded templates, store timestamp (UTC, format "2006-01-02 15:04:05 UTC")
- `RenderServer`: Execute server.go.tmpl with TemplateData, return bytes
- `FormatCode`: Call go/format.Source() to format generated code
- `generateTimestamp`: Return current time in UTC formatted as "2006-01-02 15:04:05 UTC" (human-readable, not RFC3339)

**Context:**
- Embed templates with `//go:embed templates/*.tmpl` into `templateFS embed.FS`
- Use text/template.ParseFS() to parse templates
- Template helper functions: lower, upper, hasPrefix, hasSuffix, trimPrefix, trimSuffix, join
- Timestamp format: "2006-01-02 15:04:05 UTC" (not RFC3339 format with 'T' and 'Z')

#### 3. Template Embedding
**File**: `internal/generate/duh/embed.go`

```go
import "embed"

//go:embed templates/*.tmpl
var templateFS embed.FS
```

### Testing Requirements

**File**: `internal/generate/duh/server_test.go` (new file)

```go
func TestGenerateDuhCreatesServerFile(t *testing.T)
func TestGeneratedServerCompiles(t *testing.T)
func TestGeneratedServerStructure(t *testing.T)
func TestServerWithMultipleOperations(t *testing.T)
```

**Test Approach:**
- ALL tests are functional e2e tests calling `duh.RunCmd()`
- Tests verify file creation, compilation, and structure through CLI
- Use valid DUH-RPC specs with proper error schemas

**Test Objectives:**
1. **TestGenerateDuhCreatesServerFile**:
   - Run `duh generate duh openapi.yaml`
   - Verify exit code 0
   - Verify server.go file exists in output directory
   - Verify success message contains "server.go"

2. **TestGeneratedServerCompiles**:
   - Generate server.go via CLI
   - Run `go build` in output directory with go.mod and proto stubs
   - Verify compilation succeeds

3. **TestGeneratedServerStructure**:
   - Generate server.go via CLI
   - Read file content
   - Verify contains: package declaration, RPC constants, ServiceInterface, Handler struct, ServeHTTP method
   - Verify no build tags present

4. **TestServerWithMultipleOperations**:
   - Use spec with 3+ operations
   - Verify all operation constants generated
   - Verify all methods in ServiceInterface
   - Verify all cases in ServeHTTP switch

**Context:**
- Update `duh.Run()` to call generator and write server.go
- Tests drive the implementation through CLI interface
- Each test verifies a different aspect of the generated code

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestGenerateDuhCreatesServerFile
go test ./internal/generate/duh -v -run TestGeneratedServerCompiles
go test ./internal/generate/duh -v -run Server
```

---

## Phase 3: Iterator Template Generation

### Overview
Create iterator.go template. This is independent of client/server. Only generated when HasListOps is true.

**Implementation Note:** Update `duh.Run()` to conditionally generate iterator.go. Tests will verify file is created only when list operations exist.

### Changes Required

#### 1. Template File
**File**: `internal/generate/duh/templates/iterator.go.tmpl`

**Template Structure** (based on iterator.go reference):
- File header: generation comment
- Package declaration
- Imports (context only)
- Page[T] struct
- Iterator[T] interface
- PageFetcher[T] interface
- GenericIterator[T] struct
- NewGenericIterator[T] constructor
- Next() method with context cancellation
- Err() method

**Template Data Access:**
- `.Package` - package name
- `.Timestamp` - generation timestamp
- No iteration needed (generic code, not operation-specific)

**Context:**
- Model exactly after /Users/thrawn/Development/duh-poc/iterator.go
- No build tags
- Pure generic implementation

#### 2. Generator Method
**File**: `internal/generate/duh/generator.go` (add method)

```go
func (g *Generator) RenderIterator(data *TemplateData) ([]byte, error)
```

**Function Responsibilities:**
- Execute iterator.go.tmpl with TemplateData
- Return formatted bytes
- Only called when data.HasListOps is true

### Testing Requirements

**File**: `internal/generate/duh/iterator_test.go` (new file)

```go
func TestGenerateDuhCreatesIteratorWithListOps(t *testing.T)
func TestGenerateDuhSkipsIteratorWithoutListOps(t *testing.T)
func TestGeneratedIteratorCompiles(t *testing.T)
func TestIteratorStructure(t *testing.T)
```

**Test Approach:**
- ALL tests are functional e2e tests calling `duh.RunCmd()`
- Tests verify conditional iterator generation through CLI
- Use two spec variants: one with list ops, one without

**Test Objectives:**
1. **TestGenerateDuhCreatesIteratorWithListOps**:
   - Run CLI with spec containing list operation
   - Verify iterator.go file exists
   - Verify success message mentions iterator.go

2. **TestGenerateDuhSkipsIteratorWithoutListOps**:
   - Run CLI with spec without list operations
   - Verify iterator.go file does NOT exist
   - Verify success message does not mention iterator.go

3. **TestGeneratedIteratorCompiles**:
   - Generate all files via CLI (spec with list ops)
   - Verify iterator.go compiles independently
   - Verify contains: Page, Iterator, PageFetcher, GenericIterator types

4. **TestIteratorStructure**:
   - Generate iterator.go via CLI
   - Verify no build tags present
   - Verify generic implementation (not operation-specific)
   - Verify context cancellation in Next() method

**Context:**
- Update `duh.Run()` to conditionally generate iterator.go
- Tests drive conditional logic through CLI interface
- Reuse existing test helper and spec constants from config_test.go

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestGenerateDuhCreatesIteratorWithListOps
go test ./internal/generate/duh -v -run TestGenerateDuhSkipsIteratorWithoutListOps
go test ./internal/generate/duh -v -run Iterator
```

---

## Phase 4: Client Template Generation

### Overview
Create client.go template. This depends on iterator types, so iterator.go must be generated first for list operations.

**Implementation Note:** Update `duh.Run()` to generate client.go. Tests will verify client compiles with iterator.go when list operations exist.

### Changes Required

#### 1. Template File
**File**: `internal/generate/duh/templates/client.go.tmpl`

**Template Structure** (based on client.go reference):
- File header: generation comment
- Package declaration
- Imports (bytes, context, crypto/tls, errors, fmt, net/http, duh-go, proto, clock, set)
- ClientInterface definition
- ClientConfig struct
- Client struct
- NewClient constructor
- RPC method implementations (CreateUser, etc.)
- Close method
- For each list operation:
  - PageFetcher struct (e.g., UserPageFetcher)
  - FetchPage method
  - Iterator method (e.g., ListUsersIter)
- WithTLS helper
- WithNoTLS helper

**Template Data Access:**
- `.Package`, `.ProtoImport`, `.Timestamp`
- `.Operations` - for ClientInterface methods and RPC implementations
- `.ListOps` - for PageFetcher structs and iterator methods

**Context:**
- Model after /Users/thrawn/Development/duh-poc/client.go
- Hard-coded connection pool sizes: 2000 for all (MaxConnsPerHost, MaxIdleConns, MaxIdleConnsPerHost)
- IdleConnTimeout: 60 * clock.Second
- NewClient default: 5000 for connection pool sizes
- Iterator methods use limit parameter, default to 10 if <= 0

#### 2. Generator Method
**File**: `internal/generate/duh/generator.go` (add method)

```go
func (g *Generator) RenderClient(data *TemplateData) ([]byte, error)
```

**Function Responsibilities:**
- Execute client.go.tmpl with TemplateData
- Return formatted bytes
- Template handles conditional iterator generation based on ListOps

### Testing Requirements

**File**: `internal/generate/duh/client_test.go` (new file)

```go
func TestGenerateDuhCreatesClientFile(t *testing.T)
func TestGeneratedClientCompiles(t *testing.T)
func TestClientIteratorIntegration(t *testing.T)
func TestClientWithoutIterator(t *testing.T)
func TestClientStructure(t *testing.T)
```

**Test Approach:**
- ALL tests are functional e2e tests calling `duh.RunCmd()`
- Tests verify client generation and iterator integration through CLI
- Use specs with and without list operations

**Test Objectives:**
1. **TestGenerateDuhCreatesClientFile**:
   - Run CLI with valid spec
   - Verify client.go file exists
   - Verify success message mentions client.go
   - Verify file contains ClientInterface, NewClient, WithTLS/WithNoTLS

2. **TestGeneratedClientCompiles**:
   - Generate all files via CLI
   - Create stub proto files for compilation
   - Run `go build` in output directory
   - Verify compilation succeeds

3. **TestClientIteratorIntegration**:
   - Generate with spec containing list operations
   - Verify client.go contains iterator methods (e.g., ListUsersIter)
   - Verify client.go contains PageFetcher structs
   - Verify client + iterator compile together

4. **TestClientWithoutIterator**:
   - Generate with spec without list operations
   - Verify client compiles without iterator.go
   - Verify no iterator methods in client

5. **TestClientStructure**:
   - Verify ClientInterface has all RPC methods
   - Verify hard-coded connection pool sizes
   - Verify WithTLS/WithNoTLS helper functions present

**Context:**
- Update `duh.Run()` to generate client.go
- Tests verify integration with iterator.go when present
- Client compiles with or without iterator depending on list ops

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestGenerateDuhCreatesClientFile
go test ./internal/generate/duh -v -run TestGeneratedClientCompiles
go test ./internal/generate/duh -v -run Client
```

---

## Phase 5: Proto Generation and Orchestration

### Overview
Implement ProtoConverter mock and finalize orchestration. This completes `duh.Run()` to generate all files and output final success message.

**Implementation Note:** Update `duh.Run()` to generate proto file and output complete success message listing all generated files.

### Changes Required

#### 1. ProtoConverter Interface and Mock Implementation
**File**: `internal/generate/duh/converter.go`

```go
type ProtoConverter interface {
    Convert(openapi []byte, packageName string) ([]byte, error)
}

type MockProtoConverter struct{}

func NewMockProtoConverter() *MockProtoConverter
func (m *MockProtoConverter) Convert(openapi []byte, packageName string) ([]byte, error)
func extractSchemaNames(openapi []byte) ([]string, error)
func generateProtoFile(packageName string, schemas []string) []byte
```

**Function Responsibilities:**
- `NewMockProtoConverter`: Create mock converter instance (used in production until real implementation available)
- `Convert`: Parse OpenAPI YAML, extract schema names, generate proto file with empty messages
- `extractSchemaNames`: Use yaml parser to find components.schemas keys
- `generateProtoFile`: Generate proto file with syntax, package, and empty message declarations

**Important Note:**
- MockProtoConverter is used in PRODUCTION (not just tests) until real ProtoConverter implementation is provided
- This allows the generator to work end-to-end while proto conversion is developed separately
- Generated proto files will have correct structure but empty message bodies

**Generated Proto Format:**
```protobuf
syntax = "proto3";

package duh.api.v1;

message CreateUserRequest {}
message CreateUserResponse {}
message UserResponse {}
```

**Context:**
- Use gopkg.in/yaml.v3 to parse OpenAPI (already in dependencies)
- Extract only root-level schema names (keys under components.schemas)
- Convert schema names to camel case (already done if following naming convention)

#### 2. Main Orchestration
**File**: `internal/generate/duh/duh.go`

```go
func Run(w io.Writer, specPath string, config Config, converter ProtoConverter) error
func writeFile(path string, content []byte) error
func ensureDir(path string) error
```

**Function Responsibilities:**
- `Run`: Orchestrate full generation pipeline
  1. Load OpenAPI spec using lint.Load()
  2. Validate spec using lint.Validate() - exit on failure
  3. Detect module path from go.mod
  4. Parse spec into TemplateData
  5. Generate server.go (write immediately)
  6. Generate iterator.go if HasListOps (write immediately)
  7. Generate client.go (write immediately)
  8. Generate proto file via converter (write immediately)
  9. Print success message with file list
- `writeFile`: Write bytes to file path, create parent dirs if needed
- `ensureDir`: Create directory and parents with os.MkdirAll

**Context:**
- Non-atomic: write each file as generated, abort on first error
- Don't clean up partial files on error
- Use fmt.Fprintf for success message with ✓

#### 3. Success Message Format
```
✓ Generated 4 files in <output-dir>
  - client.go
  - server.go
  - iterator.go
  - proto/v1/api.proto
```

Or without list operations:
```
✓ Generated 3 files in <output-dir>
  - client.go
  - server.go
  - proto/v1/api.proto
```

### Testing Requirements

**File**: `internal/generate/duh/proto_test.go` (new file)

```go
func TestGenerateDuhCreatesProtoFile(t *testing.T)
func TestProtoFileStructure(t *testing.T)
func TestProtoWithCustomPath(t *testing.T)
func TestProtoSchemaExtraction(t *testing.T)
```

**Test Approach:**
- ALL tests are functional e2e tests calling `duh.RunCmd()`
- Tests verify proto generation through CLI
- Tests verify proto file structure and content

**Test Objectives:**
1. **TestGenerateDuhCreatesProtoFile**:
   - Run CLI with valid spec
   - Verify proto file exists at configured path
   - Verify proto file contains: syntax, package, message declarations
   - Verify success message mentions proto file

2. **TestProtoFileStructure**:
   - Generate proto file via CLI
   - Verify syntax declaration (proto3)
   - Verify package declaration matches config
   - Verify all schema names appear as messages
   - Verify messages are empty (mock implementation)

3. **TestProtoWithCustomPath**:
   - Use --proto-path flag with custom path
   - Verify proto file created at custom location
   - Verify proto import path updated correctly

4. **TestProtoSchemaExtraction**:
   - Use spec with multiple schemas
   - Verify all schema names extracted
   - Verify no duplicates in proto file

**Context:**
- Update `duh.Run()` to generate proto file via MockProtoConverter
- Tests verify proto generation and configuration
- MockProtoConverter creates empty messages (production use until real converter ready)

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestGenerateDuhCreatesProtoFile
go test ./internal/generate/duh -v -run Proto
```

---

## Phase 6: Command Integration (COMPLETED)

### Overview
Wire up the generator to the CLI command structure in run_cmd.go. Create barebones `duh.Run()` that uses Phase 1 components and establish functional testing pattern.

### Changes Required

#### 1. Add Command to CLI
**File**: `run_cmd.go`

```go
import "github.com/duh-rpc/duh-cli/internal/generate/duh"
```

Add duhCmd under generateCmd (after line 201):

```go
duhCmd := &cobra.Command{
    Use:   "duh [openapi-file]",
    Short: "Generate DUH-RPC client, server, and proto from OpenAPI specification",
    Long: `Generate DUH-RPC client, server, and proto from OpenAPI specification.

The duh command generates DUH-RPC specific code including HTTP client with
pagination iterators, server with routing, and protobuf definitions.

By default, generates client.go, server.go, iterator.go (if list operations),
and proto file. Use flags to customize output.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.

Exit Codes:
  0    All components generated successfully
  2    Error (file not found, validation failed, generation failed, etc.)`,
    Args: cobra.MaximumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation here
    },
}
```

**Function Responsibilities:**
- Parse flags into Config struct
- Detect module path or use override
- Create MockProtoConverter instance with NewMockProtoConverter()
- Call duh.Run() with config and converter
- Set exitCode based on error (0 = success, 2 = any error including validation)

**Flags to register:**
- `--package` / `-p`: Package name (default: "api")
- `--output-dir`: Output directory (default: ".")
- `--proto-path`: Proto file path (default: "proto/v1/api.proto")
- `--proto-import`: Proto import override (optional)
- `--proto-package`: Proto package override (optional)

**Context:**
- Follow pattern from oapiCmd (lines 152-200)
- Add duhCmd to generateCmd: `generateCmd.AddCommand(oapiCmd, duhCmd)`

### Testing Requirements (COMPLETED)

**File**: `internal/generate/duh/config_test.go` (renamed from generate_test.go)

**Implemented Tests:**
```go
func TestGenerateDuhParsesSimpleSpec(t *testing.T)           // ✓ PASSING
func TestGenerateDuhParsesListOperation(t *testing.T)        // ✓ PASSING
func TestGenerateDuhWithCustomPackage(t *testing.T)          // ✓ PASSING
func TestGenerateDuhRejectsMainPackage(t *testing.T)         // ✓ PASSING
func TestGenerateDuhRejectsInvalidPackage(t *testing.T)      // ✓ PASSING
func TestGenerateDuhDetectsModulePath(t *testing.T)          // ✓ PASSING
func TestGenerateDuhMissingGoMod(t *testing.T)               // ✓ PASSING
```

**Test Coverage:**
- ✓ Command registered and accessible via CLI
- ✓ Default flag values applied correctly
- ✓ Custom flags (package, output-dir, proto-*) parsed correctly
- ✓ Validation errors return exit code 2
- ✓ Module path detection from go.mod
- ✓ Error messages are clear and helpful

**Pattern Established:**
- Helper function `setupTest(t, spec)` for common setup
- Use valid DUH-RPC compliant specs with proper error schemas
- Verify exit codes and stdout output
- All tests functional e2e through CLI

### Validation Commands

```bash
go test ./internal/generate/duh -v        # ✓ All 7 tests passing
go build ./...                            # ✓ Builds successfully
go vet ./...                              # ✓ No issues
```

---

## Phase 7: End-to-End Integration Testing

### Overview
Comprehensive functional tests that exercise the entire system through the CLI. These tests verify the complete pipeline with realistic specs and full compilation.

### Testing Requirements

**File**: `internal/generate/duh/integration_test.go` (new file)

```go
func TestEndToEndGeneration(t *testing.T)
func TestGeneratedCodeCompiles(t *testing.T)
func TestGeneratedCodeStructure(t *testing.T)
func TestListOperationIterator(t *testing.T)
func TestMultipleOperations(t *testing.T)
func TestProtoImportPaths(t *testing.T)
func TestTimestampInHeaders(t *testing.T)
func TestNonAtomicGeneration(t *testing.T)
func TestFullPipelineWithDependencies(t *testing.T)
```

**Test Objectives:**
- Generate all files from reference OpenAPI spec (duh-poc/openapi.yaml)
- Verify all files compile together with go build
- Verify generated code structure matches reference files
- Test list operation detection and iterator generation
- Test multiple operation types (create, get, list, update)
- Verify proto import paths constructed correctly
- Verify all files have consistent timestamp in header
- Verify partial file creation on error (non-atomic behavior)

**Test Scenarios:**

1. **Complete Generation**: Valid spec with 4 operations (1 list)
   - Verify 4 files created
   - Verify all compile together
   - Check RPC constants match paths

2. **No List Operations**: Valid spec with only create/get/update
   - Verify 3 files created (no iterator.go)
   - Verify client compiles without iterator

3. **Custom Configuration**: Flags for package, output-dir, proto-path
   - Verify custom values used throughout
   - Verify proto import constructed correctly

4. **Validation Failure**: Invalid spec
   - Verify exit code 2
   - Verify no files created

5. **Module Path Detection**: Various go.mod formats
   - Single-line module declaration
   - Module with comment
   - Nested module path

6. **Non-Atomic Behavior**: Trigger error mid-generation
   - Verify partial files left on disk
   - Verify error message clear

**Context:**
- Use realistic OpenAPI specs (can reuse duh-poc/openapi.yaml content)
- Create complete Go module environment with go.mod
- Run go build to verify compilation
- Check file contents for expected patterns

### Validation Commands

```bash
# Run all integration tests
go test ./internal/generate/duh -v -run TestEndToEnd

# Verify generated code compiles
go test ./internal/generate/duh -v -run TestGeneratedCodeCompiles

# Full test suite
go test ./internal/generate/duh -v

# Manual test
go run ./cmd/duh generate duh testdata/openapi.yaml
cd output && go build .
```

---

## Testing Strategy Summary

### Functional Testing Only
**ALL TESTS MUST BE FUNCTIONAL E2E TESTS** - No unit tests allowed. All tests call `duh.RunCmd()` and test through the CLI interface.

### Testing Pattern (As Implemented in Phase 1 & 6)
- **Location**: Tests in `internal/generate/duh/` alongside implementation
- **Style**: Call `duh.RunCmd(&stdout, []string{"generate", "duh", ...})`
- **Verification**: Check exit codes (0=success, 2=error) and stdout output
- **Specs**: Use valid DUH-RPC compliant specs with proper error schemas
- **Setup**: Use helper function `setupTest(t, spec)` for common setup
- **Assertions**: Use require for critical (exit code), assert for non-critical (output content)

### Test File Organization
Tests are organized by feature/phase to keep files manageable:
- `config_test.go` - Configuration, validation, module detection (Phase 1 & 6)
- `server_test.go` - Server generation and compilation (Phase 2)
- `iterator_test.go` - Iterator generation (conditional) (Phase 3)
- `client_test.go` - Client generation and iterator integration (Phase 4)
- `proto_test.go` - Proto file generation (Phase 5)
- `integration_test.go` - End-to-end full pipeline tests (Phase 7)

Each file contains functional e2e tests calling `duh.RunCmd()` through CLI

### Example Test Pattern
```go
func TestGenerateDuhCreatesServerFile(t *testing.T) {
    specPath, stdout := setupTest(t, validSpec)

    exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

    require.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "✓")

    // Verify file exists
    _, err := os.Stat(filepath.Join(tempDir, "server.go"))
    require.NoError(t, err)

    // Verify it compiles
    output, err := exec.Command("go", "build", "./...").CombinedOutput()
    require.NoError(t, err, string(output))
}
```

### What Each Phase Tests
- **Phase 1**: Config validation, module detection, operation parsing
- **Phase 2**: Server file generation and compilation
- **Phase 3**: Iterator file generation (conditional on list ops)
- **Phase 4**: Client file generation and compilation
- **Phase 5**: Proto file generation, full pipeline integration
- **Phase 6**: Command flags, error handling (COMPLETED)
- **Phase 7**: End-to-end with real specs, structure validation

---

## Error Handling Strategy

### Validation Errors (exit code 2)
```
Error: OpenAPI validation failed

/v1/api/users
  Path must follow format: /v{version}/{subject}.{method}
  Found: /v1/api/users

Run 'duh lint openapi.yaml' for details
```

### Missing go.mod (exit code 2)
```
Error: go.mod not found in current directory

The 'duh generate duh' command requires a go.mod file to determine
the module import path for generated code.

To initialize a Go module, run:
    go mod init github.com/yourorg/yourproject
```

### Template Errors (exit code 2)
```
Error: Failed to render client.go template

template: client.go.tmpl:45:12: executing "client.go.tmpl" at <.InvalidField>:
can't evaluate field InvalidField in type *duh.TemplateData

This is a bug in the template. Please report at:
https://github.com/duh-rpc/duh-cli/issues
```

### File Write Errors (exit code 2)
```
Error: Failed to write client.go

open /path/to/client.go: permission denied

Ensure you have write permissions to the output directory
```

---

## Code Style and Conventions

### Generated Code
- Match reference files from duh-poc
- No build tags (spec requirement overrides reference)
- Header format: `// Code generated by 'duh generate' on 2025-10-25 20:15:32 UTC. DO NOT EDIT.`
- Use visual tapering for struct fields where applicable
- Standard camel case for names (Id, Http, Url)
- Import order: stdlib, external, local

### Template Code
- Use `{{- ... -}}` sparingly (only where whitespace matters)
- Comment template logic with `{{/* comment */}}`
- Consistent indentation
- Clear variable names in loops

### Go Code
- Follow CLAUDE.md guidelines
- Prefer one/two word variables
- No inline variables used once
- Use const for unchanging values
- Visual tapering for struct literals

---

## Dependencies

### External (already in go.mod)
- github.com/pb33f/libopenapi v0.28.1 - OpenAPI parsing
- github.com/spf13/cobra v1.10.1 - CLI framework
- gopkg.in/yaml.v3 v3.0.1 - YAML parsing for proto converter
- text/template (stdlib) - Template rendering
- go/format (stdlib) - Code formatting
- embed (stdlib) - Template embedding

### Internal
- internal/lint - Spec validation (lint.Load, lint.Validate, lint.Print)
- Requires go.mod in target directory

---

## Success Criteria

### Phase 1 (Foundation) - COMPLETED
- [x] Config loads from flags with defaults
- [x] Module path extracted from go.mod
- [x] Operation names generated correctly (all test cases pass)
- [x] Operations extracted from OpenAPI spec
- [x] List operations detected with 3-criteria test
- [x] All functional tests pass (26 tests)

### Phase 2 (Server)
- [ ] Server template renders without errors
- [ ] Generated server.go compiles independently
- [ ] RPC constants match spec paths
- [ ] ServiceInterface has all operations
- [ ] Switch/case routing includes all paths
- [ ] Individual handlers generated correctly

### Phase 3 (Iterator)
- [ ] Iterator template renders without errors
- [ ] Generated iterator.go compiles independently
- [ ] Iterator only generated when HasListOps=true
- [ ] Generic types work correctly

### Phase 4 (Client)
- [ ] Client template renders without errors
- [ ] Generated client.go compiles with iterator.go
- [ ] ClientInterface has all operations
- [ ] Iterator methods generated for list operations
- [ ] PageFetcher structs created correctly
- [ ] WithTLS/WithNoTLS helpers present

### Phase 5 (Proto & Orchestration)
- [ ] Mock proto converter generates valid proto
- [ ] Schema names extracted correctly
- [ ] Full pipeline runs end-to-end
- [ ] All files generated in correct locations
- [ ] Success message accurate

### Phase 6 (Command)
- [x] Command registered under generate
- [x] Flags parsed correctly
- [x] Default values applied
- [x] Error messages helpful
- [x] Exit codes correct

### Phase 7 (Integration)
- [ ] Generated code from real spec compiles
- [ ] Generated structure matches reference files
- [ ] Multiple operation types handled
- [ ] Proto imports correct
- [ ] Non-atomic behavior verified

### Overall
- [ ] All tests pass: `go test ./internal/generate/duh -v`
- [ ] Code passes vet: `go vet ./...`
- [ ] Binary builds: `go build -o duh .`
- [ ] Manual test succeeds: `./duh generate duh openapi.yaml && cd output && go build .`
- [ ] CLAUDE.md guidelines followed

---

## Implementation Notes

### Template Development Tips
1. Start with static sections (headers, imports)
2. Add dynamic sections incrementally
3. Test after each addition
4. Use edge case data (0 ops, 1 op, many ops)

### Common Pitfalls
- Forgetting to check HasListOps before generating iterator
- Not formatting code before writing (go/format.Source)
- Using wrong path separator (use filepath.Join)
- Not handling missing operationId in OpenAPI
- Incorrect list operation detection (checking path vs operationId)

### Debugging
- Print TemplateData struct before rendering
- Check template syntax with minimal data first
- Verify go.mod parsing with test files
- Test proto converter with sample schemas

---

## Rollback Strategy

**Non-Atomic Generation (as specified):**
- Abort on first error
- Leave partial files on disk
- User can inspect generated files to debug
- User re-runs command after fixing issues

**No cleanup required** - partial state helps debugging

---

## Timeline Estimate

- **Phase 1**: 2 days (Foundation)
- **Phase 2**: 2 days (Server Template)
- **Phase 3**: 1 day (Iterator Template)
- **Phase 4**: 2 days (Client Template)
- **Phase 5**: 1 day (Proto & Orchestration)
- **Phase 6**: 1 day (Command Integration)
- **Phase 7**: 1 day (Integration Testing)

**Total**: 10 days

---

## Next Steps

1. Begin Phase 1: Implement config.go, naming.go, parser.go, types.go
2. Write comprehensive unit tests for Phase 1
3. Verify all Phase 1 tests pass before proceeding
4. Move to Phase 2 only after Phase 1 complete and tested
5. Follow phase sequence strictly (each fully tested before next)

---

## Review History and Corrections

### Review 1 Fixes Applied

**Critical Issues Resolved:**
1. ✅ **Atomic vs Non-Atomic**: Confirmed non-atomic generation (leave partial files on error) - this is the requirement
2. ✅ **ProtoConverter Usage**: Clarified MockProtoConverter used in PRODUCTION until real implementation ready
3. ✅ **Array Field Selection**: Specified "YAML definition order via libopenapi's ordered property iteration"
4. ✅ **Exit Codes**: Fixed - all errors return exit code 2 (including validation), removed incorrect "exit code 1" mention
5. ✅ **Timestamp Format**: Confirmed human-readable format "2006-01-02 15:04:05 UTC" (not RFC3339), updated all references
6. ✅ **strings.Title Deprecation**: Replaced with `strings.ToUpper(s[:1]) + s[1:]` for ASCII capitalization
7. ✅ **Parser Edge Cases**: Added requirement to error on inline schemas, skip operations without $ref

**Documentation Improvements:**
- Added "Key Constraints and Clarifications" section consolidating all critical decisions
- Clarified that list detection checks method portion of path (after dot)
- Documented module path validation (must contain "/")
- Clarified MockProtoConverter is production code, not test-only
- Added test cases for inline schema errors and missing $ref handling

All review issues addressed. Plan is now ready for implementation.
