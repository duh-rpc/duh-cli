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

**Testing:** Each phase fully tested before proceeding to next phase (TDD approach).

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

**File**: `internal/generate/duh/generator_test.go`

```go
func TestRenderServer(t *testing.T)
func TestFormatCode(t *testing.T)
func TestGeneratedServerCompiles(t *testing.T)
```

**Test Objectives:**
- Render server template with mock TemplateData (2 operations, one list)
- Verify output contains expected constants, interface methods, handlers
- Verify code formats successfully with gofmt
- Verify generated code contains package declaration, imports, RPC constants
- Write to temp file, verify syntax with go/parser

**File**: `internal/generate/duh/server_test.go` (functional)

```go
func TestGenerateServerCode(t *testing.T)
```

**Test Objectives:**
- Create complete generation pipeline: parse config, parse spec, render server
- Write server.go to temp directory with go.mod
- Verify file exists and compiles (go build)
- Check generated code structure matches reference

**Context:**
- Use inline spec with 2-3 operations
- Mock TemplateData if spec parsing not ready yet
- Verify no build tags in output

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestRenderServer
go test ./internal/generate/duh -v -run TestGeneratedServerCompiles
```

---

## Phase 3: Iterator Template Generation

### Overview
Create iterator.go template. This is independent of client/server. Only generated when HasListOps is true.

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

**File**: `internal/generate/duh/generator_test.go` (add tests)

```go
func TestRenderIterator(t *testing.T)
func TestIteratorOnlyWhenListOps(t *testing.T)
func TestGeneratedIteratorCompiles(t *testing.T)
```

**Test Objectives:**
- Render iterator template with TemplateData where HasListOps=true
- Verify output contains Page, Iterator, PageFetcher, GenericIterator
- Verify iterator NOT generated when HasListOps=false
- Verify generated code compiles independently

**File**: `internal/generate/duh/iterator_test.go` (functional)

```go
func TestGenerateIteratorCode(t *testing.T)
func TestNoIteratorWithoutListOps(t *testing.T)
```

**Test Objectives:**
- Generate iterator.go for spec with list operation
- Verify file created and compiles
- Verify iterator.go NOT created for spec without list operations
- Check generated code matches reference structure

**Context:**
- Use two inline specs: one with list op, one without
- Verify conditional generation logic

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestRenderIterator
go test ./internal/generate/duh -v -run TestGeneratedIteratorCompiles
go test ./internal/generate/duh -v -run TestNoIteratorWithoutListOps
```

---

## Phase 4: Client Template Generation

### Overview
Create client.go template. This depends on iterator types, so iterator.go must be generated first for list operations.

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

**File**: `internal/generate/duh/generator_test.go` (add tests)

```go
func TestRenderClient(t *testing.T)
func TestClientWithIterators(t *testing.T)
func TestClientWithoutIterators(t *testing.T)
func TestGeneratedClientCompiles(t *testing.T)
```

**Test Objectives:**
- Render client template with TemplateData (multiple operations, some list)
- Verify ClientInterface has all expected methods
- Verify iterator methods only generated for list operations
- Verify WithTLS/WithNoTLS helpers present
- Verify generated code compiles with iterator.go

**File**: `internal/generate/duh/client_test.go` (functional)

```go
func TestGenerateClientCode(t *testing.T)
func TestClientIteratorIntegration(t *testing.T)
```

**Test Objectives:**
- Generate client.go for spec with list operations
- Verify client + iterator compile together (go build on directory)
- Check PageFetcher and Iterator method generated correctly
- Verify client without list ops compiles (no iterator dependency)

**Context:**
- Use inline spec with mixed operations (create, get, list, update)
- Generate both client.go and iterator.go, verify they compile together

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestRenderClient
go test ./internal/generate/duh -v -run TestGeneratedClientCompiles
go test ./internal/generate/duh -v -run TestClientIteratorIntegration
```

---

## Phase 5: Proto Generation and Orchestration

### Overview
Implement ProtoConverter mock and main orchestration function that ties all pieces together.

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

**File**: `internal/generate/duh/converter_test.go`

```go
func TestMockProtoConverter(t *testing.T)
func TestExtractSchemaNames(t *testing.T)
func TestGenerateProtoFile(t *testing.T)
```

**Test Objectives:**
- Extract schema names from inline OpenAPI spec
- Generate proto with correct syntax, package, empty messages
- Verify schema names appear as message declarations

**File**: `internal/generate/duh/duh_test.go`

```go
func TestRunFullPipeline(t *testing.T)
func TestRunWithListOperations(t *testing.T)
func TestRunWithoutListOperations(t *testing.T)
func TestRunValidationFailure(t *testing.T)
func TestRunMissingGoMod(t *testing.T)
```

**Test Objectives:**
- Run full generation pipeline with valid spec
- Verify all files created (client, server, iterator, proto)
- Verify files compile together (go build)
- Verify iterator not created when no list ops
- Verify validation errors stop generation (no files created)
- Verify error when go.mod missing

**Context:**
- Use temp directory with go.mod
- Copy/create test OpenAPI spec
- Run full pipeline, check all outputs

### Validation Commands

```bash
go test ./internal/generate/duh -v
go test ./internal/generate/duh -run TestRunFullPipeline -v
```

---

## Phase 6: Command Integration

### Overview
Wire up the generator to the CLI command structure in run_cmd.go.

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

### Testing Requirements

**File**: `internal/generate/duh/command_test.go`

```go
func TestDuhCommandWithDefaults(t *testing.T)
func TestDuhCommandWithCustomFlags(t *testing.T)
func TestDuhCommandValidation(t *testing.T)
func TestDuhCommandMissingGoMod(t *testing.T)
```

**Test Objectives:**
- Run command via duh.RunCmd() with minimal args
- Verify default values applied
- Test custom package name, output dir, proto flags
- Verify validation errors propagate correctly
- Verify helpful error when go.mod missing

**Context:**
- Functional tests using duh.RunCmd(&stdout, []string{"generate", "duh", ...})
- Create temp directories, go.mod, OpenAPI specs
- Verify exit codes and stdout messages

### Validation Commands

```bash
go test ./internal/generate/duh -v -run TestDuhCommand
go build ./...
```

---

## Phase 7: End-to-End Integration Testing

### Overview
Comprehensive functional tests that exercise the entire system through the CLI.

### Testing Requirements

**File**: `internal/generate/duh/integration_test.go`

```go
func TestEndToEndGeneration(t *testing.T)
func TestGeneratedCodeCompiles(t *testing.T)
func TestListOperationIterator(t *testing.T)
func TestMultipleOperations(t *testing.T)
func TestProtoImportPaths(t *testing.T)
func TestTimestampInHeaders(t *testing.T)
func TestNonAtomicGeneration(t *testing.T)
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

### Unit Testing
- **Config**: Validation, module path detection, proto import construction
- **Naming**: Operation name generation, camel case conversion
- **Parser**: Operation extraction, list detection, schema type resolution
- **Generator**: Template rendering, code formatting, conditional generation
- **Converter**: Schema extraction, proto file generation

### Functional Testing
- **Each Phase**: Standalone functional test verifying generated code compiles
- **Integration**: End-to-end tests through CLI interface
- **Error Cases**: Validation failures, missing dependencies, invalid configs

### Testing Pattern
- Use `duh.RunCmd()` for functional tests
- Inline specs as const strings (unless >30 lines)
- Temp directories via `t.TempDir()`
- require for critical assertions, assert for non-critical
- No explanatory messages in assertions

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

### Phase 1 (Foundation)
- [ ] Config loads from flags with defaults
- [ ] Module path extracted from go.mod
- [ ] Operation names generated correctly (all test cases pass)
- [ ] Operations extracted from OpenAPI spec
- [ ] List operations detected with 3-criteria test
- [ ] All unit tests pass

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
- [ ] Command registered under generate
- [ ] Flags parsed correctly
- [ ] Default values applied
- [ ] Error messages helpful
- [ ] Exit codes correct

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
