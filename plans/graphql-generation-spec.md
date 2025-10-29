# Technical Specification: GraphQL Query Generation from DUH-RPC OpenAPI Specifications

## Review Status
- Review Cycles Completed: 4 (Final)
- Last Updated: 2025-10-29
- Changes in v4 (Final Polish):
  - **FIX:** Updated generated file header to match duh-cli template pattern (line 326)
  - **FIX:** Corrected Query return types to nullable (removed `!`) for error handling consistency (lines 612, 615)
  - **VERIFIED:** All previous critical/major fixes properly integrated
  - **VERIFIED:** Internal consistency across all sections
  - **VERIFIED:** Code examples match stated requirements
- Changes in v3 (Critical/Major Issues Resolved):
  - **CRITICAL FIX:** Clarified resolver implementation - complete gqlgen boilerplate generation (REQ-007)
  - **MAJOR FIX 1:** Completed extension validation rules with collision detection and error handling (Section 4.3)
  - **MAJOR FIX 2:** Specified list operation detection algorithm with 3-step process (REQ-010)
  - **MAJOR FIX 3:** Detailed error schema generation using GraphQL errors extension (REQ-009)
  - **MAJOR FIX 4:** Added complete gqlgen.yml template with scalars and autobind (Section 4.4)
  - **MAJOR FIX 5:** Clarified Service interface uses gqlgen-generated types from graph/model (REQ-007)
  - Added pre-validation requirement (lint.Validate() runs first)
  - Added file generation behavior and overwrite policy
  - Documented edge cases for valid DUH-RPC specs
- Changes from v2:
  - Updated naming strategy from `subject_method` to camelCase `subjectMethod`
  - Added rationale for camelCase vs. idiomatic naming (simpler, predictable)
  - Clarified hybrid approach: defaults with extension overrides
  - Updated all examples to use camelCase (usersGet, usersList, productInventorySearch)
- Final Approval: **READY FOR IMPLEMENTATION** ✅✅

## 1. Overview

This specification describes the addition of GraphQL schema and code generation capabilities to the duh-cli ecosystem. The feature enables developers to generate GraphQL query APIs from DUH-RPC compliant OpenAPI specifications, creating a unified approach to exposing RPC operations via GraphQL.

**Business Value:**
- Enables GraphQL clients to consume DUH-RPC services without requiring separate GraphQL schema maintenance
- Leverages existing OpenAPI specs as single source of truth for API contracts
- Provides automatic type-safe GraphQL client and server code generation
- Facilitates gradual migration or dual-protocol support (HTTP + GraphQL) for existing DUH-RPC services

**Key Design Decisions:**

*Naming Strategy:* This spec uses **camelCase naming** for GraphQL fields derived from DUH-RPC paths:
- `/v1/users.get` → `usersGet` (not `user`)
- `/v1/users.list` → `usersList` (not `users`)
- `/v1/product-inventory.search` → `productInventorySearch`

**Rationale:** Predictable, simple to implement, no pluralization complexity. Users can override to more idiomatic names (e.g., `user`, `users`) using the `x-graphql-field-name` extension.

**Future:** A `--naming-strategy=idiomatic` flag could be added later for automatic intelligent naming with pluralization.

## 2. Current State Analysis

**Affected modules:**
- `cmd/duh/main.go:1-11` - CLI entry point (will add new `graphql` subcommand under `generate`)
- `internal/generate/` - Generation package directory (will add new `graphql` subdirectory)
- `run_cmd.go:202-270` - Command routing (will add GraphQL command handler similar to `duhCmd`)
- `internal/generate/duh/parser.go` - Parsing patterns to reference/reuse

**Current behavior:**
- duh-cli generates Go HTTP clients/servers and protobuf definitions from DUH-RPC OpenAPI specs
- DUH-RPC uses POST for ALL operations (read and write) - HTTP method is always POST
- Method names in paths indicate semantic operation type: `users.get`, `users.list` are reads despite POST
- Uses parser → validator → template generator architecture
- Current generators: `duh` (Go client/server/proto) and `oapi` (oapi-codegen wrapper)
- Parser (parser.go:60-148) extracts operations with clear request/response types from `components/schemas`
- Service interface pattern exists: `duh generate duh --full` creates service.go with interface methods

**Relevant ADRs reviewed:** None found (no ADR directory exists in project)

**Technical debt identified:**
- No existing GraphQL generation capabilities
- No pattern for calling external code generators from within duh-cli (will establish with gqlgen)
- Query operation classification logic does not exist

## 3. Architectural Context

### Relevant ADRs
No ADRs exist for this project currently.

### Architectural Principles
Based on codebase analysis (docs/duh-openapi-reference.md, internal/generate/duh/):
- **Single source of truth:** OpenAPI spec drives all code generation
- **Validation first:** Lint/validate OpenAPI before generation (lint.Validate in duh.go:12-20)
- **Template-based generation:** Use Go templates for code generation where possible
- **Modular generators:** Separate generators for different output formats (duh, oapi, proto)
- **Cobra CLI:** Use cobra for command structure and flag parsing
- **DUH-RPC semantics:** POST-only protocol, method names carry semantic meaning

## 4. Requirements

### Functional Requirements

**REQ-001: Separate OpenAPI-to-GraphQL Library**
- Create standalone Go library named `openapi-to-gql`
- Library accepts DUH-RPC compliant OpenAPI specification (must pass `lint.Validate`)
- Library outputs GraphQL Schema Definition Language (SDL) string
- Library validates DUH-RPC compliance before conversion
- Library is importable and usable independently of duh-cli
- Acceptance: Can import library, call `Convert(spec) (string, error)`, receive valid GraphQL SDL

**REQ-002: Query Operation Detection**
- Detect query operations using method name heuristics on path `/v{version}/{subject}.{method}`:
  - Heuristic patterns: `get`, `list`, `search`, `find`, `fetch`, `query`, `retrieve`
  - Match if method name **contains** any pattern (e.g., `get-by-id` matches `get`)
- Support OpenAPI extension `x-graphql-query: true` to force include operation
- Support OpenAPI extension `x-graphql-query: false` to force exclude operation
- Extension takes precedence over heuristics
- Non-query operations are excluded from GraphQL schema
- Remember: All DUH-RPC operations use POST method, classification is by path method name only
- Acceptance:
  - `/v1/users.get` → included (matches `get`)
  - `/v1/users.create` → excluded (no match)
  - `/v1/products.search-inventory` → included (matches `search`)
  - `/v1/orders.delete` with `x-graphql-query: true` → included (extension override)

**REQ-003: GraphQL Type Mapping**
- **Pre-Validation:** OpenAPI spec MUST pass `lint.Validate()` before GraphQL generation
  - Ensures DUH-RPC compliance (valid request/response schemas, proper structure, etc.)
  - Generation assumes spec is valid - no need to handle malformed schemas
- Map OpenAPI `components/schemas` request types to GraphQL `input` types
- Map OpenAPI `components/schemas` response types to GraphQL `type` types (output types)
- Use hybrid type mapping strategy (see Type Mapping Table in Section 4.1)
- Sanitize names for GraphQL compatibility:
  - Remove unsupported characters: `-`, `.`, `,`, `:`, `;`, `@`, `#`, `$`, `%`, `^`, `&`, `*`
  - Replace with underscore `_`
  - If first character becomes invalid, prefix with `Type_`
  - Handle collisions by appending `_2`, `_3`, etc.
- All sanitized names are reversible (store original in GraphQL description/comment)
- **Edge Cases (Valid DUH-RPC, Special GraphQL Handling):**
  - **Multiple Content Types:** If operation has multiple response content types (e.g., `application/json` and `application/xml`):
    - Use `application/json` schema (GraphQL standard)
    - If `application/json` not present, fail with error suggesting use of duh generator instead
  - **Unsupported Schema Constructs:** If `oneOf`, `anyOf`, or `allOf` encountered (valid OpenAPI, not yet supported for GraphQL):
    ```
    Error: Unsupported schema construct: oneOf
    Reason: GraphQL union/interface generation not yet implemented (Phase 2)
    ```
- Acceptance: All OpenAPI types have corresponding GraphQL type definitions with valid names

**REQ-004: Field Naming Strategy**
- Default: Use camelCase `{subject}{Method}` from DUH-RPC path `/v{version}/{subject}.{method}`
  - Example: `/v1/users.get` → `usersGet`
  - Example: `/v2/product-inventory.search` → `productInventorySearch` (sanitized, camelCased)
  - Remove hyphens/underscores and capitalize following word: `product-inventory` → `productInventory`
- Support OpenAPI extension `x-graphql-field-name: "customName"` to override default
- Field names must be valid GraphQL identifiers: `[a-zA-Z_][a-zA-Z0-9_]*`
- Acceptance: Field names follow camelCase convention and respect custom overrides

**REQ-005: duh-cli GraphQL Command Integration**
- Add `duh generate graphql [flags] [openapi-file]` subcommand
- Integrate openapi-to-gql library for schema generation
- Import gqlgen as Go library to generate resolver scaffolding and Go types
- Follow gqlgen directory conventions: `graph/schema.graphqls`, `graph/generated/`, `graph/model/`
- Command flags (see Section 4.2 for full interface):
  - `--output-dir` or `-o`: Output directory (default: `graph`)
  - `--package` or `-p`: Package name (default: `api`)
  - `--schema-only`: Generate schema file only, skip Go code generation
- Acceptance: `duh generate graphql openapi.yaml` produces:
  - `graph/schema.graphqls` - GraphQL schema
  - `graph/gqlgen.yml` - gqlgen configuration
  - `graph/generated/generated.go` - gqlgen runtime
  - `graph/resolver.go` - Resolver interface
  - `graph/service.go` - Service interface (business logic layer)
  - `graph/client.go` - Type-safe GraphQL client

**REQ-006: gqlgen Configuration Generation**
- Generate `graph/gqlgen.yml` configuration file
- Configure autobind to use generated GraphQL models in `graph/model/`
- Map GraphQL scalar types to Go types:
  - `DateTime` → `time.Time`
  - `UUID` → `string` (with validation in resolvers)
  - Standard scalars (Int, String, Boolean, Float) → Go primitives
- Specify resolver paths and package structure
- Acceptance: Generated gqlgen.yml enables successful execution of gqlgen library functions

**REQ-007: Service Interface Resolver Pattern**
- Generate resolver struct with Service interface field (similar to duh server.go pattern)
- Generate Service interface with one method per query operation
- Generate COMPLETE resolver implementation that delegates to Service interface
- **Resolver Implementation Details:**
  - Generate `Resolver` struct implementing gqlgen's `ResolverRoot` interface
  - Generate `Query()` method returning `QueryResolver` implementation
  - Generate `queryResolver` struct that wraps Service interface
  - Generate resolver methods on `queryResolver` that delegate to Service
  - ALL gqlgen boilerplate is generated - no manual wiring required
- **File Structure:**
  - `graph/resolver.go` - Resolver struct, Query() method, and queryResolver implementation
  - `graph/service.go` - Service interface definition (user implements this)
- Service methods have signature: `MethodName(ctx context.Context, input *InputType) (*OutputType, error)`
  - Service method names match GraphQL field names (PascalCase)
  - Input/output types use gqlgen-generated types from `graph/model/` package
- Acceptance:
  - Resolver fully implements gqlgen's `ResolverRoot` and `QueryResolver` interfaces
  - Each query operation has corresponding Service method
  - Developers ONLY implement Service interface for business logic (can call HTTP, gRPC, database, etc.)
  - No manual resolver wiring or gqlgen interface implementation required

**REQ-008: Custom GraphQL Client Generation**
- Generate type-safe GraphQL client in `graph/client.go`
- Client struct with methods for each query operation
- Client methods build GraphQL query strings and execute via HTTP POST to GraphQL endpoint
- Method signature: `MethodName(ctx context.Context, input InputType) (*OutputType, error)`
- Use Go's `net/http` and `encoding/json` for GraphQL requests
- Acceptance:
  - Generated client compiles without errors
  - Client provides one function per query operation
  - Type-safe: uses generated input/output types from `graph/model/`

**REQ-009: Error Schema Generation**
- **Phase 1 Approach: GraphQL Errors Extension (Recommended)**
  - Resolvers return `(result, error)` following Go conventions
  - GraphQL runtime adds errors to response `errors` array automatically
  - Query fields return nullable types (null on error)
  - No schema changes needed for errors
  - **Error Response Format:**
    ```json
    {
      "data": { "user": null },
      "errors": [{
        "message": "User not found",
        "path": ["user"],
        "extensions": {
          "code": "NOT_FOUND",
          "httpStatus": 404
        }
      }]
    }
    ```

- **Error Handling in Resolvers:**
  - Service interface returns Go errors
  - Resolvers pass errors to gqlgen
  - gqlgen adds errors to GraphQL response automatically
  - DUH-RPC error codes/messages preserved in error extensions

- **OpenAPI Error Schema Detection:**
  - Parse 4xx and 5xx response schemas from OpenAPI
  - Extract error structure (code, message, details fields)
  - Document expected error format in GraphQL schema comments
  - No explicit error types generated in Phase 1

- **Future (Phase 2): Explicit Error Unions:**
  - Could generate explicit error types in schema:
    ```graphql
    type UserError {
      code: String!
      message: String!
      details: JSON
    }

    union UserResult = User | UserError

    type Query {
      user(input: GetUserRequest!): UserResult!
    }
    ```
  - Requires more complex client handling
  - Deferred to Phase 2 based on user feedback

- **Acceptance:**
  - Resolvers can return Go errors
  - GraphQL errors include DUH-RPC error code/message in extensions
  - Error responses follow GraphQL error specification
  - **Query field return types MUST be nullable (no `!` suffix) to handle errors**
    - Example: `user(input: GetUserRequest!): GetUserResponse` (nullable)
    - NOT: `user(input: GetUserRequest!): GetUserResponse!` (non-nullable)

**REQ-010: Pagination Support for List Operations**
- **List Operation Detection Algorithm:**
  1. **Method Name Check:** Operation method name matches "list" heuristic (per REQ-002), OR
  2. **Response Schema Check:** 200 response schema has array field at root level, OR
  3. **Extension Override:** Operation has OpenAPI extension `x-duh-list-operation: true`

- **Detection Examples:**
  - `/v1/users.list` → List (method name)
  - `/v1/users.search` with response `{ users: [...] }` → List (array field)
  - `/v1/users.query` with response `[...]` → List (root array)
  - `/v1/users.get-multiple` with `x-duh-list-operation: true` → List (extension)

- **Pagination Field Generation:**
  - If operation is detected as list AND request schema doesn't already have `offset`/`limit`:
    - Add `offset: Int` (optional, default 0) to input type
    - Add `limit: Int` (optional, default 100) to input type
  - If request schema already has pagination fields, preserve them as-is

- **Response Type Handling:**
  - Generate wrapper type for paginated responses if not already wrapped:
    ```graphql
    type UsersListResponse {
      users: [User!]!
      total: Int
    }
    ```
  - If response is root array `[User]`, wrap in generated response type
  - If response is already object with array field, use as-is

- **Acceptance:**
  - List queries support offset/limit parameters in GraphQL
  - Non-list operations don't get pagination fields
  - Pagination fields match OpenAPI request schema when present

### Non-Functional Requirements

**Performance:**
- Schema generation completes in <2 seconds for specs with 10-50 operations
- Schema generation completes in <10 seconds for specs with 200+ operations
- Memory usage remains under 500MB for large specs (200+ operations)

**Security:**
- No code execution from OpenAPI spec content
- Sanitize all user-provided names to prevent GraphQL injection
- Validate generated GraphQL schema parses correctly before writing files
- Validate generated Go code compiles (optional check via `go/parser`)

**Scalability:**
- Support OpenAPI specs with 200+ operations
- Handle deeply nested type definitions (10+ levels)
- Handle large schemas (100+ type definitions)

**Maintainability:**
- Clear separation between openapi-to-gql library (conversion) and duh-cli integration (orchestration)
- Comprehensive error messages following duh-cli style (see lint package for examples)
- Extensive unit test coverage (>80%) for both library and CLI integration
- All generated code follows duh-cli formatting conventions (gofmt, golangci-lint)

**File Generation Behavior:**
- If output directory doesn't exist, create it
- If output directory exists with generated files:
  - **Overwrite without prompt:** `schema.graphqls`, `gqlgen.yml`, `generated/`, `model/`
  - **Overwrite without prompt:** `resolver.go`, `service.go`, `client.go`
  - Rationale: All files are code-generated, users should not edit them manually
  - User implementations go in separate files (e.g., `service_impl.go` implementing Service interface)
- Generated files include comment header:
  ```go
  // Code generated by 'duh generate graphql' on {{.Timestamp}}. DO NOT EDIT.
  ```
  - Format matches existing duh-cli template pattern (see `internal/generate/duh/templates/*.tmpl`)
  - `{{.Timestamp}}` is populated at generation time with format: `2025-10-29 12:34:56 UTC`

### 4.1 Type Mapping Table

This table defines exact OpenAPI → GraphQL type mappings:

| OpenAPI Type | OpenAPI Format | GraphQL Type | Go Type | Notes |
|--------------|----------------|--------------|---------|-------|
| `string` | (none) | `String` | `string` | Standard mapping |
| `string` | `date-time` | `DateTime` | `time.Time` | Custom scalar (RFC3339) |
| `string` | `uuid` | `UUID` | `string` | Custom scalar with validation |
| `string` | `email` | `String` | `string` | Validation in resolver |
| `string` | `uri` | `String` | `string` | Validation in resolver |
| `string` | `byte` | `String` | `string` | Base64 encoded |
| `integer` | (none) or `int32` | `Int` | `int32` | GraphQL Int is 32-bit |
| `integer` | `int64` | `Int64` | `int64` | Custom scalar for 64-bit |
| `number` | (none) or `float` | `Float` | `float64` | Standard mapping |
| `number` | `double` | `Float` | `float64` | GraphQL Float is 64-bit |
| `boolean` | (none) | `Boolean` | `bool` | Standard mapping |
| `array` | (any) | `[Type]` | `[]Type` | List type |
| `object` | (none) | Custom `type` | `struct` | Generate from schema |
| `object` (request) | (none) | Custom `input` | `struct` | Input type for mutations |
| `enum` | (none) | GraphQL `enum` | Go `string` type | Enum values preserved |

**Custom Scalar Definitions Required:**
```graphql
scalar DateTime  # RFC3339 format: 2024-01-15T10:30:00Z
scalar UUID      # RFC4122 format: 123e4567-e89b-12d3-a456-426614174000
scalar Int64     # 64-bit integer (JSON number)
```

**Unsupported/Future Work:**
- `oneOf` / `anyOf` / `allOf` → Future: GraphQL unions/interfaces
- `$ref` circular references → Handled by gqlgen, but validate during generation
- `additionalProperties` → Map to JSON scalar or reject with error

### 4.2 Command Line Interface Specification

```bash
duh generate graphql [flags] [openapi-file]

Generate GraphQL schema and Go code from DUH-RPC OpenAPI specifications.

The graphql command generates a complete GraphQL server including schema,
resolvers, service interface, and type-safe client from a DUH-RPC compliant
OpenAPI specification.

Only query operations (detected by method name heuristics: get, list, search,
find, fetch, query, retrieve) are included in the GraphQL schema. Override
detection with x-graphql-query extension.

If no file path is provided, defaults to 'openapi.yaml' in current directory.

Examples:
  # Generate in default directory (graph/)
  duh generate graphql openapi.yaml

  # Generate in custom directory
  duh generate graphql openapi.yaml --output-dir api/graphql

  # Generate schema only (skip Go code generation)
  duh generate graphql openapi.yaml --schema-only

  # Custom package name
  duh generate graphql openapi.yaml --package graphqlapi

Flags:
  -o, --output-dir string   Output directory for generated files (default "graph")
  -p, --package string      Package name for generated code (default "api")
      --schema-only         Generate schema file only, skip Go code generation
  -h, --help                Help for graphql command

Exit Codes:
  0    Generation successful
  2    Error (file not found, validation failed, generation error)

Generated Files:
  <output-dir>/schema.graphqls          GraphQL schema
  <output-dir>/gqlgen.yml               gqlgen configuration
  <output-dir>/generated/generated.go   gqlgen runtime
  <output-dir>/model/models_gen.go      Generated Go types
  <output-dir>/resolver.go              Resolver implementation
  <output-dir>/service.go               Service interface
  <output-dir>/client.go                Type-safe GraphQL client
```

### 4.3 OpenAPI Extension Specification

**Extension: `x-graphql-query`**
- **Location:** Operation level (under `post:` in path item)
- **Type:** `boolean`
- **Purpose:** Override heuristic-based query detection
- **Values:**
  - `true`: Force include operation in GraphQL schema as query
  - `false`: Force exclude operation from GraphQL schema
  - (absent): Use heuristic detection
- **Example:**
  ```yaml
  paths:
    /v1/orders.delete:
      post:
        x-graphql-query: true  # Override: include despite 'delete' name
        operationId: deleteOrder
        # ...
  ```

**Extension: `x-graphql-field-name`**
- **Location:** Operation level (under `post:` in path item)
- **Type:** `string`
- **Purpose:** Override default GraphQL field name
- **Validation:** Must match regex `^[a-zA-Z_][a-zA-Z0-9_]*$` (valid GraphQL identifier)
- **Default:** `{subject}{Method}` in camelCase from path (sanitized)
- **Example:**
  ```yaml
  paths:
    /v1/users.get-by-id:
      post:
        x-graphql-field-name: "user"  # Override default: usersGetById
        operationId: getUserById
        # ...
  ```

**Validation Rules:**
1. **Invalid Identifier:**
   - If `x-graphql-field-name` is invalid identifier (doesn't match `^[a-zA-Z_][a-zA-Z0-9_]*$`), fail with error:
     ```
     Error: Invalid x-graphql-field-name "user name" (contains space)
     Suggestion: Use valid GraphQL identifier (e.g., "userName")
     ```

2. **Duplicate Field Names:**
   - If `x-graphql-field-name` creates duplicate field name (collision with another operation), fail with error:
     ```
     Error: Duplicate GraphQL field name "user"
     Conflicts:
       - /v1/users.get has x-graphql-field-name: "user"
       - /v1/user.get generates default field name "userGet"
     Suggestion: Use different x-graphql-field-name value or rename operations
     ```

3. **Invalid Operation:**
   - If `x-graphql-query: true` on operation without valid request/response schema, fail with error:
     ```
     Error: x-graphql-query: true on operation without response schema
     Operation: /v1/users.get
     Suggestion: Add 200 response schema or remove x-graphql-query extension
     ```

4. **Non-POST Operation:**
   - If `x-graphql-query` is set on non-POST operation, fail with error:
     ```
     Error: x-graphql-query extension only valid on POST operations (DUH-RPC requirement)
     Operation: GET /v1/users.get
     Suggestion: Change HTTP method to POST or remove extension
     ```

5. **Ignored Extensions:**
   - If `x-graphql-field-name` is set on excluded operation (`x-graphql-query: false`), log warning but continue:
     ```
     Warning: x-graphql-field-name ignored on excluded operation
     Operation: /v1/users.create
     Reason: x-graphql-query: false
     ```

6. **Optional Extensions:**
   - Extensions are optional; generation works without any extensions using heuristics
   - Default behavior: heuristic-based query detection + camelCase field naming

### 4.4 Concrete Example: OpenAPI → GraphQL Transformation

**Input: OpenAPI Specification (excerpt)**
```yaml
openapi: 3.0.0
info:
  title: Users API
  version: 1.0.0

paths:
  /v1/users.get:
    post:
      operationId: getUser
      summary: Get user by ID
      x-graphql-field-name: "user"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetUserResponse'

  /v1/users.list:
    post:
      operationId: listUsers
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListUsersRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'

  /v1/users.create:
    post:
      operationId: createUser
      summary: Create new user (excluded - not a query)
      # ...

components:
  schemas:
    GetUserRequest:
      type: object
      required: [userId]
      properties:
        userId:
          type: string
          format: uuid

    GetUserResponse:
      type: object
      properties:
        user:
          $ref: '#/components/schemas/User'

    ListUsersRequest:
      type: object
      properties:
        offset:
          type: integer
          format: int32
        limit:
          type: integer
          format: int32

    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/User'
        total:
          type: integer

    User:
      type: object
      required: [id, email, createdAt]
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
        createdAt:
          type: string
          format: date-time
```

**Output: GraphQL Schema (`graph/schema.graphqls`)**
```graphql
# Custom scalars
scalar DateTime
scalar UUID
scalar Int64

# Query operations (read-only)
type Query {
  "Get user by ID (from /v1/users.get)"
  user(input: GetUserRequest!): GetUserResponse

  "List users (from /v1/users.list)"
  usersList(input: ListUsersRequest!): ListUsersResponse
}

# Input types (from request schemas)
input GetUserRequest {
  userId: UUID!
}

input ListUsersRequest {
  offset: Int
  limit: Int
}

# Output types (from response schemas)
type GetUserResponse {
  user: User
}

type ListUsersResponse {
  users: [User!]!
  total: Int!
}

type User {
  id: UUID!
  email: String!
  name: String
  createdAt: DateTime!
}
```

**Output: gqlgen Configuration (`graph/gqlgen.yml`)**
```yaml
# Generated gqlgen configuration
schema:
  - schema.graphqls

exec:
  filename: generated/generated.go
  package: generated

model:
  filename: model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: .
  package: api
  filename_template: "{name}.resolvers.go"

# Autobind to generated models
autobind:
  - github.com/yourmodule/graph/model

# Custom scalar type mappings
models:
  DateTime:
    model: time.Time
  UUID:
    model: github.com/99designs/gqlgen/graphql.String
  Int64:
    model: github.com/99designs/gqlgen/graphql.Int64

# Omit specific fields from generation (resolvers will be custom-generated)
omit_resolvers:
  - Query
```

**Output: Service Interface (`graph/service.go`)**
```go
package api

import (
    "context"
    "time"
)

// Service interface defines business logic operations
// Implement this interface to provide GraphQL resolver logic
// Note: Uses gqlgen-generated types from graph/model package
type Service interface {
    // User gets a user by ID (from /v1/users.get)
    // Note: Method name matches GraphQL field "user" (uses x-graphql-field-name extension)
    User(ctx context.Context, input *model.GetUserRequest) (*model.GetUserResponse, error)

    // UsersList lists users with pagination (from /v1/users.list)
    // Note: Method name matches GraphQL field "usersList" (default camelCase naming)
    UsersList(ctx context.Context, input *model.ListUsersRequest) (*model.ListUsersResponse, error)
}

// Note: Input/Output types are generated by gqlgen in graph/model/models_gen.go
// Example types (actual generation done by gqlgen):
//
// type GetUserRequest struct {
//     UserID string `json:"userId"`
// }
//
// type GetUserResponse struct {
//     User *User `json:"user"`
// }
//
// type ListUsersRequest struct {
//     Offset *int32 `json:"offset"`
//     Limit  *int32 `json:"limit"`
// }
//
// type ListUsersResponse struct {
//     Users []*User `json:"users"`
//     Total int32   `json:"total"`
// }
//
// type User struct {
//     ID        string    `json:"id"`
//     Email     string    `json:"email"`
//     Name      *string   `json:"name"`
//     CreatedAt time.Time `json:"createdAt"`
// }
```

**Output: Resolver (`graph/resolver.go`)**
```go
package api

import (
    "context"
    "github.com/yourmodule/graph/model"
)

// Resolver implements gqlgen's ResolverRoot interface
type Resolver struct {
    Service Service
}

// Query returns the QueryResolver implementation
func (r *Resolver) Query() QueryResolver {
    return &queryResolver{r}
}

// queryResolver wraps the Resolver and delegates to Service
type queryResolver struct {
    *Resolver
}

// User resolver delegates to Service interface
// Note: GraphQL field is "user" (lowercase), Go method is User (capitalized per Go conventions)
func (r *queryResolver) User(ctx context.Context, input model.GetUserRequest) (*model.GetUserResponse, error) {
    return r.Service.User(ctx, &input)
}

// UsersList resolver delegates to Service interface
func (r *queryResolver) UsersList(ctx context.Context, input model.ListUsersRequest) (*model.ListUsersResponse, error) {
    return r.Service.UsersList(ctx, &input)
}
```

**Output: Client (`graph/client.go`)**
```go
package api

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    endpoint string
    client   *http.Client
}

func NewClient(endpoint string) *Client {
    return &Client{
        endpoint: endpoint,
        client:   &http.Client{},
    }
}

// User executes the user query
func (c *Client) User(ctx context.Context, input GetUserRequest) (*GetUserResponse, error) {
    query := `query User($input: GetUserRequest!) {
        user(input: $input) {
            user {
                id
                email
                name
                createdAt
            }
        }
    }`

    var response struct {
        Data struct {
            User GetUserResponse `json:"user"`
        } `json:"data"`
    }

    if err := c.execute(ctx, query, map[string]interface{}{"input": input}, &response); err != nil {
        return nil, err
    }

    return &response.Data.User, nil
}

// Similar implementation for UsersList...

func (c *Client) execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
    // GraphQL request execution logic
    // ...
}
```

## 5. Technical Approach

### Chosen Solution

**Two-Component Architecture:**

**Component 1: openapi-to-gql Library** (New Repository)
- Separate Go module/repository: `github.com/duh-rpc/openapi-to-gql`
- Input: DUH-RPC compliant OpenAPI v3 spec (libopenapi Document)
- Output: GraphQL SDL string (.graphql file content)
- Architecture: Parser → Validator → Mapper → Generator
  - **Parser:** Extract operations and schemas from OpenAPI using pb33f/libopenapi
  - **Validator:** Ensure DUH-RPC compliance (reuse duh-cli's lint logic)
  - **Mapper:** Transform OpenAPI types to GraphQL types per Type Mapping Table
  - **Generator:** Render GraphQL SDL using Go text/template
- Public API:
  ```go
  type Converter struct {
      HeuristicPatterns []string // Default: get, list, search, find, fetch, query, retrieve
  }
  func (c *Converter) Convert(spec *v3.Document) (string, error)
  ```

**Component 2: duh-cli GraphQL Command** (Modify Existing Repository)
- New subcommand under `duh generate graphql`
- Located in `internal/generate/graphql/` package
- Orchestration workflow:
  1. **Load & Validate**: Load OpenAPI spec using `lint.Load()`, validate with `lint.Validate()`
  2. **Generate Schema**: Call `openapi-to-gql.Convert()` to generate GraphQL SDL string
  3. **Write Schema**: Write SDL to `graph/schema.graphqls`
  4. **Generate gqlgen.yml**: Create gqlgen configuration file
  5. **Call gqlgen**: Import and call gqlgen library functions for code generation
  6. **Generate Service**: Create service.go with Service interface using templates
  7. **Generate Client**: Create client.go with type-safe client using templates
  8. **Output**: Print success message with file list

### Rationale

**Why separate library?**
- Enables reuse outside duh-cli ecosystem (other tools can convert DUH-RPC specs to GraphQL)
- Cleaner separation of concerns (conversion logic vs. orchestration/CLI)
- Easier to test conversion logic independently
- Could be used as standalone CLI tool or imported as library

**Why camelCase naming (usersGet) instead of idiomatic (user)?**
- Simpler implementation: No pluralization/singularization logic needed
- Predictable: Users can easily derive field names from DUH-RPC paths
- Consistent: Same pattern for all operations (no special cases for get/list/etc.)
- Flexible: Users can override to idiomatic names via `x-graphql-field-name` extension
- Reduces bugs: Pluralization is complex (person→people, data→data) and English-centric
- Future-proof: Can add `--naming-strategy=idiomatic` flag later if demand exists

**Why queries only (no mutations)?**
- Simplifies initial implementation scope (Phase 1 deliverable)
- DUH-RPC uses POST for all operations, but only reads make sense as GraphQL queries
- Write operations (create, update, delete) can be added as mutations in Phase 2
- Query operations are more straightforward to map (no side effects to manage)

**Why custom client vs. genqlient?**
- Matches existing duh-cli patterns (client.go in `duh generate duh`)
- No external tool dependency or workflow change for users
- Can generate optimized queries specific to schema
- Full control over client API and error handling
- Consistent with "duh-cli generates everything" philosophy

**Why service interface pattern?**
- Matches existing `duh generate duh --full` pattern (users familiar with this)
- Separates GraphQL layer (resolvers) from business logic (service)
- Allows service implementation to call HTTP, gRPC, database, or any backend
- Testable: mock Service interface for resolver tests
- Flexible: service can be implemented multiple ways (HTTP proxy, direct DB, etc.)

**Why import gqlgen as library vs. shell out?**
- Better error handling and control flow (no subprocess parsing)
- Avoids subprocess management complexity
- Enables programmatic configuration of gqlgen (dynamic gqlgen.yml)
- Faster execution (no process spawning overhead)
- Consistent with Go ecosystem best practices

**Why hybrid type mapping?**
- Common formats (date-time, uuid, int64) benefit from type safety via custom scalars
- Less common formats (email, uri) add complexity; validate in resolvers instead
- Balance between type safety and schema complexity
- GraphQL clients can handle custom scalars with proper codecs

### Architectural Alignment

This approach aligns with duh-cli's existing patterns:
- Similar to protobuf generation: separate converter (`internal/generate/duh/converter.go`), orchestrated by duh-cli
- Follows modular generator pattern: `oapi` package, `duh` package, now `graphql` package
- Uses same validation-first approach (`lint.Validate` before generation)
- Extends cobra CLI structure consistently (subcommand under `generate`)
- Reuses parsing utilities from `internal/generate/duh/parser.go` where applicable
- Service interface pattern mirrors `duh generate duh --full` service.go

### Component Changes

**New Repository: openapi-to-gql**
- Repository: `github.com/duh-rpc/openapi-to-gql`
- Package structure:
  ```
  openapi-to-gql/
  ├── pkg/
  │   └── converter/
  │       ├── converter.go      # Public API: Convert() function
  │       ├── parser.go          # Extract operations/schemas from OpenAPI
  │       ├── mapper.go          # Type mapping (OpenAPI → GraphQL)
  │       ├── generator.go       # SDL generation using templates
  │       └── validator.go       # DUH-RPC validation
  ├── internal/
  │   ├── templates/
  │   │   └── schema.tmpl        # GraphQL schema template
  │   └── sanitize/
  │       └── names.go           # Name sanitization utilities
  ├── cmd/
  │   └── openapi-to-gql/
  │       └── main.go            # Optional standalone CLI
  ├── go.mod
  ├── go.sum
  └── README.md
  ```
- Dependencies:
  - `github.com/pb33f/libopenapi` - OpenAPI parsing
  - `github.com/stretchr/testify` - Testing

**Modified: duh-cli**
- Repository: `github.com/duh-rpc/duh-cli`
- New package: `internal/generate/graphql/`
  ```
  internal/generate/graphql/
  ├── graphql.go         # Main entry point: RunGraphQL(config) error
  ├── config.go          # gqlgen.yml generation
  ├── gqlgen.go          # gqlgen library integration
  ├── service.go         # Service interface template generation
  ├── client.go          # Client code generation
  └── templates/
      ├── service.tmpl   # Service interface template
      ├── client.tmpl    # Client template
      └── gqlgen.tmpl    # gqlgen.yml template
  ```
- Modified files:
  - `run_cmd.go` - Add `graphqlCmd` cobra command (lines 270+)
  - `go.mod` - Add dependencies:
    ```
    require (
        github.com/99designs/gqlgen v0.17.45
        github.com/duh-rpc/openapi-to-gql v0.1.0
    )
    ```

**File Generation Workflow (in duh-cli):**
```go
// Pseudocode for internal/generate/graphql/graphql.go
func RunGraphQL(config RunConfig) error {
    // 1. Load and validate OpenAPI
    spec, err := lint.Load(config.SpecPath)
    if err != nil { return err }

    result := lint.Validate(spec, config.SpecPath)
    if !result.Valid() { return fmt.Errorf("validation failed") }

    // 2. Generate GraphQL schema using openapi-to-gql library
    converter := openapiogql.NewConverter()
    schema, err := converter.Convert(spec)
    if err != nil { return err }

    // 3. Write schema file
    schemaPath := filepath.Join(config.OutputDir, "schema.graphqls")
    if err := writeFile(schemaPath, []byte(schema)); err != nil { return err }

    // 4. Generate gqlgen.yml configuration
    gqlgenConfig, err := generateGqlgenConfig(config)
    if err != nil { return err }

    gqlgenPath := filepath.Join(config.OutputDir, "gqlgen.yml")
    if err := writeFile(gqlgenPath, gqlgenConfig); err != nil { return err }

    // 5. Call gqlgen library to generate Go code
    if err := runGqlgen(config.OutputDir); err != nil { return err }

    // 6. Generate service.go interface
    serviceCode, err := generateServiceInterface(spec, config)
    if err != nil { return err }

    servicePath := filepath.Join(config.OutputDir, "service.go")
    if err := writeFile(servicePath, serviceCode); err != nil { return err }

    // 7. Generate client.go
    clientCode, err := generateClient(spec, schema, config)
    if err != nil { return err }

    clientPath := filepath.Join(config.OutputDir, "client.go")
    if err := writeFile(clientPath, clientCode); err != nil { return err }

    // 8. Print success
    fmt.Fprintf(config.Writer, "✓ Generated GraphQL code in %s\n", config.OutputDir)
    return nil
}
```

## 6. Dependencies and Impacts

**External dependencies:**
- `github.com/99designs/gqlgen` v0.17.45 (or latest v0.17.x) - GraphQL Go server generation
  - Used as library, not CLI tool
  - Import paths: `github.com/99designs/gqlgen/api`, `github.com/99designs/gqlgen/codegen`, `github.com/99designs/gqlgen/graphql`
  - License: MIT
- `github.com/pb33f/libopenapi` v0.16.x (already used in duh-cli) - OpenAPI parsing
- New: `github.com/duh-rpc/openapi-to-gql` v0.1.0 - OpenAPI to GraphQL schema conversion
  - This is the new library being created as part of this project

**Internal dependencies:**
- `internal/lint` package - Reuse `Load()` and `Validate()` for OpenAPI validation
- `internal/generate/duh/parser.go` - May share operation extraction patterns (paths, schemas)
- `internal/generate/duh/types.go` - Reference for TemplateData structure pattern
- `internal/generate/duh/naming.go` - May reuse name generation utilities

**Database impacts:** None

**Build/deployment impacts:**
- Two new Go modules to maintain:
  1. `openapi-to-gql` library (new repository)
  2. `duh-cli` updates (existing repository)
- CI/CD must build and test both repositories
- Version coordination: duh-cli go.mod specifies openapi-to-gql version
- Release process: openapi-to-gql releases first, then duh-cli imports new version

**Documentation impacts:**
- Update duh-cli README with `generate graphql` command
- Add GraphQL generation guide to docs/
- Document OpenAPI extensions (x-graphql-query, x-graphql-field-name)
- Add examples showing generated GraphQL code
- Update comparison table (duh vs oapi vs graphql generators)

## 7. Backward Compatibility

### Is this project in production?
- **Yes** - duh-cli is actively used for code generation

### Breaking Changes Allowed
- **Yes** - This is a new feature addition, not a modification of existing functionality

### Breaking Changes
None. This is purely additive functionality:
- New subcommand: `duh generate graphql`
- No changes to existing `duh generate duh` or `duh generate oapi`
- No changes to `lint`, `init`, or `add` commands
- No changes to OpenAPI spec requirements (extensions are optional)
- New output directory (`graph/`) separate from existing generators

### Compatibility Constraints
- Must maintain existing duh-cli behavior for all non-graphql commands
- Generated GraphQL code must not conflict with existing `duh` or `oapi` generated code
- Different output directories prevent file conflicts: `graph/` vs `.` (default for duh/oapi)
- OpenAPI extensions (`x-graphql-*`) are optional; specs without them still work

### Migration Path
For existing duh-cli users:
1. No migration required - existing commands work unchanged
2. To adopt GraphQL: run `duh generate graphql openapi.yaml` in addition to existing generation
3. GraphQL generation can coexist with `duh generate duh` (different output dirs)
4. Gradual adoption: add `x-graphql-*` extensions incrementally to customize behavior

## 8. Testing Strategy

### Unit Testing Approach

**openapi-to-gql Library Tests (`pkg/converter/*_test.go`):**

**Query Detection (parser_test.go):**
- Test each heuristic pattern (get, list, search, find, fetch, query, retrieve)
- Test method name variants: `get`, `get-by-id`, `get_user`, `getUserById`
- Test `x-graphql-query: true` forces include
- Test `x-graphql-query: false` forces exclude
- Test extension precedence over heuristics
- Test mixed scenarios (some operations included, some excluded)

**Type Mapping (mapper_test.go):**
- Test all primitive types: string, integer, number, boolean
- Test all format specifiers: date-time, uuid, int32, int64, email, uri
- Test object mapping to GraphQL type
- Test array mapping to GraphQL list
- Test enum mapping to GraphQL enum
- Test nested objects (3+ levels deep)
- Test required vs optional fields (GraphQL `!` suffix)
- Test circular references (ensure no infinite loops)

**Name Sanitization (sanitize/names_test.go):**
- Test removal of invalid characters and camelCase conversion: `user-name` → `userName`
- Test multiple invalid characters: `api.v2:users.get` → `apiV2UsersGet`
- Test collision handling: `userName` + `user-name` → `userName` + `userName2`
- Test edge cases: empty string, numbers only, starting with digit
- Test reversibility: can retrieve original name from sanitized name

**Schema Generation (generator_test.go):**
- Test Query root type generation with multiple fields
- Test input type generation from request schemas
- Test output type generation from response schemas
- Test custom scalar declarations
- Test GraphQL comments/descriptions from OpenAPI summary/description
- Validate generated SDL parses correctly (use graphql-go parser)

**Extension Validation (validator_test.go):**
- Test valid `x-graphql-field-name` values
- Test invalid `x-graphql-field-name` (spaces, special chars, starting with number)
- Test `x-graphql-query` with valid/invalid boolean values
- Test operation with extension but missing request/response schemas

**duh-cli Integration Tests (`internal/generate/graphql/*_test.go`):**

**Command Tests (graphql_test.go):**
- Test flag parsing: `--output-dir`, `--package`, `--schema-only`
- Test default values: output-dir=graph, package=api
- Test missing OpenAPI file error
- Test OpenAPI validation failure propagation
- Test successful generation with minimal spec
- Mock openapi-to-gql library calls

**gqlgen Config Tests (config_test.go):**
- Test gqlgen.yml generation with custom output dir
- Test gqlgen.yml generation with custom package name
- Test autobind configuration matches generated models
- Test scalar type mapping configuration
- Validate generated YAML parses correctly

**Service Interface Tests (service_test.go):**
- Test Service interface generation for single operation
- Test Service interface generation for multiple operations
- Test method signature correctness (params, returns)
- Verify generated Go code compiles

**Client Tests (client_test.go):**
- Test client generation for single query operation
- Test client generation for multiple query operations
- Test client method signatures match schema
- Test GraphQL query string construction
- Verify generated client code compiles

### Integration Testing Needs

**End-to-End Tests:**

**Test 1: Minimal Spec → Complete GraphQL Server**
- Input: OpenAPI spec with single query operation (`users.get`)
- Execute: `duh generate graphql openapi.yaml`
- Verify:
  - `graph/schema.graphqls` exists and parses as valid GraphQL
  - `graph/gqlgen.yml` exists and is valid YAML
  - `graph/generated/generated.go` exists
  - `graph/service.go` has Service interface with correct method
  - `graph/resolver.go` has resolver delegating to Service
  - `graph/client.go` has client method for query
  - All generated Go code compiles: `go build ./graph/...`

**Test 2: DUH-RPC Init Template Spec**
- Input: Output from `duh init` (users.create, users.get, users.list, users.update)
- Execute: `duh generate graphql openapi.yaml`
- Verify:
  - Only `users.get` and `users.list` appear in schema (query operations)
  - `users.create` and `users.update` excluded (not query operations)
  - List operation includes pagination (offset, limit)
  - Generated client has `UsersGet()` and `UsersList()` methods

**Test 3: Custom Spec with Extensions**
- Input: OpenAPI spec with `x-graphql-query` and `x-graphql-field-name` extensions
- Execute: `duh generate graphql openapi.yaml`
- Verify:
  - Operations with `x-graphql-query: true` are included
  - Operations with `x-graphql-query: false` are excluded
  - Field names use `x-graphql-field-name` values
  - Heuristics ignored for operations with extensions

**Test 4: Large Spec (Performance)**
- Input: OpenAPI spec with 100 operations, 50 schemas, deep nesting (5+ levels)
- Execute: `duh generate graphql openapi.yaml`
- Verify:
  - Completes in <10 seconds
  - Memory usage <500MB
  - All operations correctly classified (query vs non-query)
  - All types generated without errors
  - Generated Go code compiles

**Test 5: GraphQL Server Runtime**
- Input: Generated GraphQL server from test 1
- Implement: Stub Service interface with mock data
- Execute: Start GraphQL server, send query via HTTP POST
- Verify:
  - Server responds with valid GraphQL response
  - Query execution calls Service method
  - Response matches schema structure

**Test 6: GraphQL Client Execution**
- Input: Generated GraphQL client from test 1
- Execute: Mock GraphQL server, call client method
- Verify:
  - Client sends correct GraphQL query string
  - Client deserializes response correctly
  - Type safety enforced (compile-time errors for invalid usage)

### Test Data

**Test OpenAPI Specs:**
- `testdata/minimal.yaml` - Single query operation
- `testdata/init-template.yaml` - DUH-RPC init template
- `testdata/extensions.yaml` - Spec with x-graphql-* extensions
- `testdata/complex-types.yaml` - Nested objects, arrays, enums
- `testdata/large.yaml` - 100+ operations for performance testing
- `testdata/invalid-extension.yaml` - Invalid extension values (error case)

### User Acceptance Criteria

Criteria for feature acceptance by end users:

1. **Generation Success:**
   - Developer runs `duh generate graphql openapi.yaml`
   - Command completes without errors
   - All expected files created in `graph/` directory

2. **Schema Validity:**
   - Generated `graph/schema.graphqls` is valid GraphQL SDL
   - Can be parsed by GraphQL tools (GraphiQL, Playground, etc.)
   - Only query operations appear in schema

3. **Code Compilation:**
   - Generated Go code compiles: `go build ./graph/...`
   - No syntax errors, no missing imports
   - gofmt and golangci-lint pass

4. **Type Safety:**
   - Generated client provides typed methods
   - Compile-time errors for incorrect usage
   - IDE autocomplete works for client methods

5. **Naming Conventions:**
   - Field names follow `subjectMethod` camelCase pattern (e.g., `usersGet`, `usersList`)
   - Custom `x-graphql-field-name` values respected (e.g., override to `user` for more idiomatic naming)
   - Names are valid GraphQL identifiers

6. **Service Interface Pattern:**
   - Service interface exists with all query operation methods
   - Resolver delegates to Service interface
   - Developers can implement Service without touching resolver code

7. **Client Functionality:**
   - Generated client has method per query operation
   - Client methods build correct GraphQL queries
   - Type-safe: input/output types match schema

8. **Documentation:**
   - Command `--help` explains usage
   - Error messages are clear and actionable
   - README includes GraphQL generation examples

## 9. Implementation Notes

**Estimated complexity:** High
- New library creation: Medium-High complexity (new repo, API design, testing)
- Type mapping logic: Medium complexity (well-defined table, edge cases)
- gqlgen integration: Medium complexity (library API learning curve, config generation)
- Service/client generation: Medium complexity (templates, similar to existing duh generator)
- Testing both components: High effort (unit + integration + E2E tests)

**Suggested implementation order:**

### Phase 1: openapi-to-gql Library Foundation (Week 1-2)
1. **Repository setup:**
   - Create `github.com/duh-rpc/openapi-to-gql` repository
   - Initialize Go module, set up CI/CD (GitHub Actions)
   - Add README, LICENSE (MIT), .gitignore

2. **Query operation detection:**
   - Implement parser to extract operations from OpenAPI
   - Implement heuristic matching (get, list, search, etc.)
   - Implement `x-graphql-query` extension parsing
   - Write unit tests for detection logic (20+ test cases)

3. **Basic validation:**
   - Validate OpenAPI is DUH-RPC compliant (reuse lint logic)
   - Validate operations have request/response schemas
   - Validate extension values (boolean for x-graphql-query)

**Deliverable:** Library can parse OpenAPI and identify query operations

### Phase 2: Type Mapping Engine (Week 2-3)
1. **Type mapper implementation:**
   - Implement mapper per Type Mapping Table (Section 4.1)
   - Handle primitives: string, integer, number, boolean
   - Handle format specifiers: date-time, uuid, int64
   - Handle objects, arrays, enums
   - Handle required vs optional fields

2. **Name sanitization:**
   - Implement sanitize function (remove invalid chars, replace with `_`)
   - Implement collision detection and numbering (`_2`, `_3`)
   - Store original names in GraphQL descriptions

3. **`x-graphql-field-name` support:**
   - Parse extension value (string)
   - Validate value is valid GraphQL identifier
   - Override default naming with extension value

4. **Comprehensive testing:**
   - Test all type mappings (30+ test cases)
   - Test name sanitization edge cases (15+ test cases)
   - Test nested objects (5+ levels deep)
   - Test circular reference detection

**Deliverable:** Library can map all OpenAPI types to GraphQL types

### Phase 3: GraphQL Schema Generation (Week 3-4)
1. **SDL generator:**
   - Create GraphQL schema template (`internal/templates/schema.tmpl`)
   - Implement template rendering (text/template)
   - Generate Query root type with fields
   - Generate input types (from request schemas)
   - Generate output types (from response schemas)
   - Generate custom scalar declarations

2. **Schema validation:**
   - Parse generated SDL with graphql-go parser
   - Validate all types referenced exist
   - Validate field types are valid
   - Return errors for invalid schema

3. **Integration tests:**
   - Test with minimal OpenAPI spec (1 operation)
   - Test with complex spec (10+ operations, nested types)
   - Test with DUH-RPC init template
   - Validate generated SDL with GraphQL validator

**Deliverable:** Library generates valid GraphQL SDL from OpenAPI

### Phase 4: duh-cli Integration (Week 4-5)
1. **Add graphql command:**
   - Create `internal/generate/graphql/` package
   - Add `graphqlCmd` to `run_cmd.go` (similar to `duhCmd`)
   - Implement flag parsing: `--output-dir`, `--package`, `--schema-only`
   - Integrate `openapi-to-gql` library

2. **gqlgen configuration generation:**
   - Create `gqlgen.yml` template
   - Generate config with correct paths, package names
   - Configure autobind to generated models
   - Configure scalar type mappings

3. **gqlgen integration:**
   - Import gqlgen as library: `github.com/99designs/gqlgen/api`
   - Call `api.Generate()` with generated config
   - Handle gqlgen errors and map to duh-cli error format
   - Verify generated files exist after gqlgen runs

4. **Testing:**
   - Test command flag parsing
   - Test file creation workflow
   - Mock openapi-to-gql and gqlgen calls
   - Test error handling and propagation

**Deliverable:** `duh generate graphql` creates schema and runs gqlgen

### Phase 5: Service Interface & Client Generation (Week 5-6)
1. **Service interface generation:**
   - Create service.go template (`templates/service.tmpl`)
   - Extract query operations from OpenAPI
   - Generate Service interface with one method per operation
   - Generate method signatures with correct types
   - Write service.go to output directory

2. **Resolver generation:**
   - Create resolver.go template (or modify gqlgen output)
   - Generate resolver struct with Service field
   - Generate resolver methods delegating to Service
   - Ensure resolver satisfies gqlgen's interface

3. **Client generation:**
   - Create client.go template (`templates/client.tmpl`)
   - Generate Client struct with HTTP client
   - Generate one method per query operation
   - Build GraphQL query strings
   - Implement execute() helper for HTTP POST
   - Handle response deserialization

4. **Testing:**
   - Test service interface generation
   - Test resolver delegation pattern
   - Test client method generation
   - Test client query string construction
   - Verify all generated code compiles

**Deliverable:** Complete GraphQL server + client generation

### Phase 6: Error Handling & Polish (Week 6)
1. **Error schema mapping:**
   - Detect DUH-RPC error schemas in OpenAPI (4xx, 5xx)
   - Generate GraphQL error type
   - Map error code/message to GraphQL errors

2. **Pagination support:**
   - Detect list operations (REQ-010)
   - Generate pagination fields (offset, limit)
   - Generate paginated response types

3. **Error messages:**
   - Improve error messages throughout
   - Add suggestions for common errors
   - Format errors like lint command (see lint/print.go)

4. **Documentation:**
   - Update duh-cli README
   - Add examples/ directory with sample OpenAPI specs
   - Document OpenAPI extensions in docs/
   - Add CONTRIBUTING.md to openapi-to-gql

**Deliverable:** Production-ready feature with docs

### Phase 7: Testing & Performance (Week 7)
1. **End-to-end tests:**
   - Implement all integration tests from Section 8
   - Test with real OpenAPI specs
   - Test GraphQL server runtime
   - Test GraphQL client execution

2. **Performance testing:**
   - Test with large spec (100+ operations)
   - Measure generation time (<10 seconds)
   - Measure memory usage (<500MB)
   - Profile and optimize if needed

3. **Edge case testing:**
   - Test invalid OpenAPI specs
   - Test invalid extension values
   - Test empty specs, specs with no queries
   - Test name collision scenarios

**Deliverable:** Fully tested, performant feature

### Phase 8: Documentation & Release (Week 8)
1. **User documentation:**
   - Write comprehensive guide in docs/graphql.md
   - Add quick start example
   - Document all flags and options
   - Add troubleshooting section

2. **Release preparation:**
   - Tag openapi-to-gql v0.1.0
   - Update duh-cli go.mod to use v0.1.0
   - Write release notes
   - Update CHANGELOG.md

3. **Examples:**
   - Add example OpenAPI specs to examples/
   - Generate sample output for each example
   - Add README in examples/ explaining samples

**Deliverable:** Released feature with complete documentation

**Code style considerations:**
- Follow existing duh-cli patterns for error handling (see internal/lint/print.go)
- Use same linting/formatting standards (gofmt, golangci-lint)
- Parser should use pb33f/libopenapi like existing code (see internal/lint/load.go)
- Template-based generation where possible (see internal/generate/duh/generator.go)
- Comprehensive error messages following duh-cli style:
  ```
  Error: Failed to generate GraphQL schema

  Cause: Operation /v1/users.get has invalid x-graphql-field-name
  Value: "user name" (contains space)
  Suggestion: Use valid GraphQL identifier (e.g., "userName")
  ```

**Rollback strategy:**
- Feature flag approach: graphql command is separate from existing commands
- If issues arise, remove graphql command in patch release (revert changes to run_cmd.go)
- No impact on existing functionality since completely isolated
- Library version pinning in go.mod enables rollback to previous working version
- Can disable feature with build tags if needed: `//go:build !nographql`

## 10. ADR Recommendation

**Recommended:** Create ADR for this feature

**Suggested ADR Title:** "ADR-001: GraphQL Query Generation from DUH-RPC Specifications"

**Key decisions to document:**

1. **Separate Library Architecture:**
   - Decision: Create standalone `openapi-to-gql` library
   - Rationale: Reusability, separation of concerns, independent testing
   - Alternatives considered: Integrated solution, using existing tools (IBM openapi-to-graphql)
   - Trade-offs: Additional maintenance burden vs. cleaner architecture

2. **Queries Only (Phase 1):**
   - Decision: Support only GraphQL queries, not mutations
   - Rationale: Simplifies scope, read operations are lower risk, natural fit for DUH-RPC
   - Future path: Add mutations in Phase 2 using similar patterns

3. **Custom Client Generation:**
   - Decision: Generate custom GraphQL client, not use genqlient
   - Rationale: Matches duh-cli patterns, no workflow change, consistent developer experience
   - Alternatives considered: genqlient, graphql-go, no client generation
   - Trade-offs: Maintenance burden vs. control and consistency

4. **Service Interface Resolver Pattern:**
   - Decision: Generate resolvers that delegate to Service interface
   - Rationale: Separation of layers, testability, matches existing duh patterns
   - Alternatives considered: HTTP proxy resolvers, direct implementation stubs
   - Trade-offs: Extra indirection vs. flexibility and testability

5. **gqlgen Library Integration:**
   - Decision: Import gqlgen as Go library
   - Rationale: Better control, error handling, no subprocess management
   - Alternatives considered: Shell out to gqlgen CLI
   - Trade-offs: Tighter coupling to gqlgen version vs. better integration

6. **Type Mapping Strategy:**
   - Decision: Hybrid approach with custom scalars for common formats
   - Rationale: Balance type safety and schema complexity
   - Alternatives considered: All standard types, all custom scalars, configurable
   - Trade-offs: Less flexibility vs. simpler implementation

7. **DUH-RPC Only Scope:**
   - Decision: Support only DUH-RPC compliant OpenAPI specs
   - Rationale: Simpler implementation, leverages DUH-RPC constraints
   - Alternatives considered: General OpenAPI support
   - Future path: Could expand to general OpenAPI if demand exists

8. **Heuristic-Based Query Detection:**
   - Decision: Use method name heuristics with extension override
   - Rationale: Convention over configuration, easy to understand
   - Alternatives considered: Extension-only, HTTP method based (doesn't work for POST-only)
   - Trade-offs: Heuristics may need tuning vs. explicit annotation overhead

9. **CamelCase Field Naming:**
   - Decision: Use simple camelCase conversion `{subject}{Method}` (e.g., `usersGet`, `usersList`)
   - Rationale: Predictable, no pluralization complexity, easy to implement correctly
   - Alternatives considered: Idiomatic naming (user/users), snake_case, keeping path structure
   - Trade-offs: Less "GraphQL-like" initially vs. simplicity and predictability
   - Mitigation: `x-graphql-field-name` extension allows per-operation idiomatic overrides

**Rationale for ADR:**

This is a significant architectural addition that introduces:
- New external library with separate repository
- New code generation workflow and orchestration pattern
- Integration with third-party tool (gqlgen) as library
- Design decisions affecting future GraphQL features (mutations, subscriptions)
- New patterns for resolvers and clients

An ADR will help future contributors understand:
- Why these decisions were made
- What alternatives were considered
- Context for extending GraphQL capabilities
- History of architectural evolution

The ADR will be referenced when:
- Adding mutation support (should follow similar patterns)
- Considering subscription support
- Evaluating other protocol additions (e.g., gRPC)
- Making breaking changes to GraphQL generation

## 11. Open Questions

### For Implementation Phase:

**Low Priority (Can be decided during implementation):**
- Should openapi-to-gql provide standalone CLI in addition to library API?
  - Lean towards: Yes, for debugging and standalone usage
- Should we generate `.graphqlconfig` for IDE support?
  - Lean towards: Yes, improves developer experience
- Should we support custom template overrides for service/client generation?
  - Lean towards: Not in v1, add if users request it
- What log level should generation use? (silent, normal, verbose)
  - Lean towards: Normal by default, add `--verbose` flag if needed

**Research Needed During Implementation:**
- Investigate gqlgen's autobind with protobuf types - does it work out of the box?
- Research GraphQL union type generation from OpenAPI `oneOf` (future feature)
- Evaluate if graphql-go parser is sufficient for SDL validation
- Determine optimal template structure for client.go (one template vs. multiple)

**Future Features (Not blocking implementation):**
- GraphQL mutation support (POST operations: create, update, delete)
- GraphQL subscription support (real-time updates)
- Connection pattern support for cursor-based pagination
- OpenAPI webhooks → GraphQL subscriptions mapping
- Custom directive support in generated schema
- Schema stitching/federation support

### Resolved Questions (From Technical Reviews):

**From Initial Review:**
✅ **Resolver implementation:** Service interface pattern (Section 4, REQ-007)
✅ **Client generation:** Custom generated client (Section 4, REQ-008)
✅ **POST vs Query semantics:** All operations use POST, classify by method name (Section 4, REQ-002)
✅ **Extension format:** Simple separate extensions (Section 4.3)
✅ **Type mapping:** Hybrid approach with custom scalars (Section 4.1)
✅ **gqlgen integration:** Import as library (Section 5)
✅ **Output location:** Follow gqlgen conventions (graph/) (Section 4, REQ-005)
✅ **Heuristics configuration:** Heuristics + annotation override (Section 4, REQ-002)

**From Comprehensive Review (v3):**
✅ **Resolver wiring:** duh-cli generates COMPLETE gqlgen resolver boilerplate - no manual wiring required (REQ-007)
✅ **Type usage:** Service interface uses gqlgen-generated types from `graph/model/` package (REQ-007, Section 4.4)
✅ **Error handling:** GraphQL errors extension approach - nullable results with errors array (REQ-009)
✅ **Multiple content types:** Use `application/json` schema, fail if not present (REQ-003)
✅ **File overwrite:** Overwrite all generated files without prompt - users implement in separate files (Section 4, File Generation Behavior)
✅ **Empty/malformed schemas:** Pre-validation with `lint.Validate()` ensures DUH-RPC compliance (REQ-003)
✅ **List operation detection:** 3-step algorithm - method name, array field, or extension (REQ-010)
✅ **Extension validation:** Comprehensive rules with collision detection and error messages (Section 4.3)
✅ **Generation order:** Schema → gqlgen (generates model/) → Service interface → Client (REQ-007, Section 5)

---

**END OF SPECIFICATION**
