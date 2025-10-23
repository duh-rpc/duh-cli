# Technical Specification: DUH-RPC OpenAPI Linter

**Document Version:** 1.0
**Date:** 2025-10-22
**Status:** APPROVED - Ready for Implementation Planning
**Review Cycles:** 1 (Complete)

---

## Document Purpose

This technical specification was created during the research and planning phase for building `duhrpc-lint`, a CLI tool that validates OpenAPI 3.0 specifications for compliance with the DUH-RPC HTTP specification. This document contains all research findings, decisions, requirements, and context needed for implementation planning and development.

**How to use this document:**
- Pass this to implementation planning phase to create detailed implementation plan
- Reference during development for requirements and validation rules
- Use examples as test cases and documentation

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Research Findings](#research-findings)
3. [Decisions Made](#decisions-made)
4. [Requirements Specification](#requirements-specification)
5. [Technical Architecture](#technical-architecture)
6. [Validation Rules Reference](#validation-rules-reference)
7. [Examples](#examples)
8. [Testing Strategy](#testing-strategy)
9. [Implementation Guidance](#implementation-guidance)

---

## Executive Summary

### What We're Building

`duhrpc-lint` is a command-line tool that validates OpenAPI 3.0 YAML specifications for compliance with DUH-RPC conventions. It will:
- Parse OpenAPI YAML files using `pb33f.io/libopenapi`
- Check against 8 DUH-RPC validation rules
- Report all violations with actionable suggestions
- Provide clear exit codes for CI/CD integration

### Why This Approach

**Decision: Build standalone CLI (not Vacuum extension)**
- **Rationale:** Simpler architecture for focused use case, complete control over UX, no plugin complexity
- **Trade-off:** Less feature-rich than Vacuum, but perfectly suited for DUH-RPC validation

**Key Design Principles:**
1. **Strict enforcement only** - No configuration, no rule toggling, all violations are errors
2. **Fail fast, report all** - Collect all violations in single run, don't stop at first error
3. **Developer-friendly** - Clear messages with specific suggestions
4. **Simple distribution** - Single binary, no dependencies

### Estimated Effort

**Total:** 7-11 days of development
- Phase 1 (Foundation): 1-2 days
- Phase 2 (Core Rules): 3-4 days
- Phase 3 (Advanced Rules): 2-3 days
- Phase 4 (Documentation): 1-2 days

---

## Research Findings

### DUH-RPC Specification Analysis

**Source:** https://github.com/duh-rpc/duh-go

**Core DUH-RPC Conventions:**

1. **Path Format:** `/v{version}/{subject}.{method}`
   - Version is mandatory for all endpoints
   - Subject-action naming (Yoda-style: subject before method)
   - Examples: `/v1/users.create`, `/v1/messages.send`

2. **HTTP Method:** POST only for all RPC operations
   - All other methods should be rejected
   - Data passed in request body, not query parameters

3. **Content Negotiation:**
   - Required: `application/json` (UTF-8)
   - Optional: `application/protobuf` (ASCII), `application/octet-stream`
   - No MIME type parameters allowed

4. **Error Response Structure:**
   - Must contain: `code` (integer), `message` (string)
   - Optional: `details` (object)
   - HTTP status code should match `code` field

5. **Status Codes:**
   - Success: 200 only
   - Client errors: 400, 401, 403, 404, 429
   - Custom: 452-455 (service-specific)
   - Server error: 500

### Library Research: pb33f.io/libopenapi

**Why chosen:**
- Industry-standard Go library for OpenAPI parsing
- Mature, actively maintained by pb33f team
- High-performance with built-in indexing
- Strong reference resolution for `$ref` handling
- Supports OpenAPI 3.0 and 3.1
- Powers Vacuum (proven in production)

**Key Capabilities:**
- `NewDocument()` - Parse YAML/JSON to document model
- `BuildV3Model()` - Build high-level OpenAPI 3.x model
- Full schema traversal and reference resolution
- `$ref`, `allOf`, `oneOf`, `anyOf` support

**Version:** v0.18.0 or higher required

### Alternative Considered: Vacuum

**What is Vacuum:**
- Ultra-fast OpenAPI linter built in Go by pb33f team
- Compatible with Spectral rulesets
- Supports custom rules via Go plugins or JavaScript

**Why NOT chosen:**
- Added complexity for focused use case
- Plugin system overhead not needed
- Less control over error message format
- Overkill for DUH-RPC-specific validation

**Trade-off accepted:** Build simpler tool with less features but perfect fit for DUH-RPC

---

## Decisions Made

### Session Q&A Resolutions

**Q: How will this linter be used?**
✅ **Decision:** CLI tool that validates OpenAPI files on disk

**Q: Input format?**
✅ **Decision:** YAML only, OpenAPI 3.0 only (MVP)
- JSON support deferred to future version
- OpenAPI 3.1 support deferred to future version
- OpenAPI 2.0 (Swagger) not supported

**Q: Which DUH-RPC rules are highest priority?**
✅ **Decision:** Subset of rules focused on structural validation:
- Path format enforcement
- HTTP method enforcement
- Query parameter prohibition
- Request body requirements
- Content type validation
- Error response schema validation
- Status code restrictions
- Success response requirements

**Q: Error response structure validation?**
✅ **Decision:** Validate the response schema structure (not runtime message format)
- Check schema has required fields: `code`, `message`
- Optional field: `details`
- Resolve `$ref` references to validate structure

**Q: What output format?**
✅ **Decision:** Human-readable text for CLI
- Grouped by rule type
- Shows location, violation, and suggestion
- Summary count at end
- Machine-readable JSON deferred to future

**Q: Error severity levels?**
✅ **Decision:** All violations are errors (no warnings)
- Fail-fast approach
- Either compliant or not compliant
- No configurable severity

**Q: Configuration file?**
✅ **Decision:** No configuration, no rule disabling
- Opinionated by design
- Reduces complexity
- Clear expectations

**Q: Strictness?**
✅ **Decision:** All paths must be DUH-RPC compliant
- No path exclusions
- No ignore patterns
- Strict enforcement only

**Q: Beyond validation?**
✅ **Decision:** Just linting, no auto-fixing
- Read-only tool
- No file modification
- Suggestions only

**Q: Purpose of reference document?**
✅ **Decision:** Reference for developers writing OpenAPI specs
- Target audience: API designers
- Use case: Style guide for DUH-RPC compliance

**Q: Format preferences?**
✅ **Decision:** Rule-by-rule reference format
- Each rule documented separately
- Examples of valid/invalid
- Clear explanations

### Critical Specifications from Review

**Path Format Details:**

✅ **Version Format:**
- Major version only: `v0`, `v1`, `v2`, `v10`, `v100`
- NOT semantic versioning: `v1.2.3` is invalid
- `v0` allowed for preview/beta APIs
- Regex: `^/v(0|[1-9][0-9]*)/.+$`

✅ **Subject/Method Characters:**
- Lowercase alphanumeric, hyphens, underscores only
- Must start with letter
- Length: 1-50 characters per segment
- Multi-word allowed: `user-accounts`, `send_notification`
- Regex per segment: `^[a-z][a-z0-9_-]*$`

✅ **Path Parameters:**
- NOT allowed: `/v1/users/{userId}.get` is INVALID
- All paths must be static
- Data passed in request body

✅ **Complete Path Regex:**
```regex
^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]*\.[a-z][a-z0-9_-]*$
```

**Content Type Validation:**

✅ **Multiple Content Types:**
- Operations CAN list multiple allowed content types
- Example: Both `application/json` AND `application/protobuf` allowed
- Minimum requirement: `application/json` must be present

✅ **Request vs Response:**
- Content types CAN differ between request and response
- Example: Request JSON, response protobuf is valid

✅ **MIME Parameters:**
- Reject if semicolon found in content-type string
- `application/json; charset=utf-8` is INVALID
- Check the OpenAPI content type keys (not HTTP headers)

**Error Response Schema:**

✅ **Code Field:**
- Type: integer (not string)
- Optional: Check for `enum` constraint matching status code
- If no enum present, just validate integer type

✅ **Schema Complexity:**
- MUST resolve `$ref` references and validate structure
- MUST handle `allOf`, `oneOf`, `anyOf` combinators
- Check that combined/resolved schema satisfies requirements

✅ **Feasibility:**
- Checking enum value matches status code is optional (best effort)
- Primary validation: schema structure has required fields

**Request Body:**

✅ **Required Validation:**
- Check `requestBody.required: true` on the requestBody object
- Not the schema required fields, but the requestBody itself

**Success Response:**

✅ **Schema Requirements:**
- 200 response can have ANY valid JSON schema
- No DUH-RPC restrictions on success schema structure
- Can be object, array, or primitive

**CLI Flags:**

✅ **Argument Format:**
- Positional argument: `duhrpc-lint <file>`
- Not a flag-based file argument

✅ **Flags Included:**
- `--version`: Show tool version ✅
- `--help`: Show usage help ✅
- `--quiet`: NOT in MVP ❌

---

## Requirements Specification

### Functional Requirements

#### REQ-001: Parse OpenAPI 3.0 YAML Files

**Description:** Load and parse OpenAPI YAML files using libopenapi

**Acceptance Criteria:**
- Accept file path as CLI positional argument
- Use `pb33f.io/libopenapi` v0.18+ for parsing
- Support OpenAPI 3.0.x only (not 2.0 or 3.1 in MVP)
- Validate file exists before parsing
- Handle YAML syntax errors gracefully
- Report parsing errors with clear messages
- Exit code 2 for parse failures

**Error Messages:**
- File not found: `Error: File not found: {path}`
- Parse error: `Error: Failed to parse OpenAPI spec: {error}`
- Not OpenAPI 3.0: `Error: Only OpenAPI 3.0 is supported (found: {version})`

---

#### REQ-002: Validate DUH-RPC Path Format

**Description:** All paths must follow DUH-RPC naming convention

**Format:** `/v{version}/{subject}.{method}`

**Detailed Specification:**

**Version Segment:**
- Format: `/v{N}` where N is non-negative integer
- Valid: `v0`, `v1`, `v2`, `v10`, `v100`
- Invalid: `v1.2`, `V1`, `version1`, `v-1`, `vbeta`
- Regex: `^/v(0|[1-9][0-9]*)/.+$`
- Must start with lowercase 'v'
- No semantic versioning in MVP (major only)

**Subject Segment:**
- Characters: lowercase letters, digits, hyphens, underscores
- Must start with letter (not digit or special char)
- Length: 1-50 characters
- Multi-word allowed: `user-accounts`, `order_history`
- Valid: `users`, `api-keys`, `message_queue`
- Invalid: `Users`, `api.keys`, `123users`, `user$accounts`
- Regex: `^[a-z][a-z0-9_-]*$`

**Method Segment:**
- Same rules as subject segment
- Valid: `create`, `get-by-id`, `send_notification`
- Invalid: `Create`, `get.by.id`, `_update`

**Dot Separator:**
- Required between subject and method
- Exactly one dot: `users.create` ✅, `users..create` ❌

**Path Parameters:**
- NOT allowed in DUH-RPC paths
- `/v1/users/{userId}.get` is INVALID
- All paths must be static
- Dynamic data should be in request body

**Complete Validation Regex:**
```regex
^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$
```

**Acceptance Criteria:**
- Parse each path in OpenAPI spec
- Match against DUH-RPC format regex
- Report specific violation reason
- Suggest corrected path format

**Violation Examples:**
- `/users.create` → Missing version: `Path must start with /v{version}/`
- `/v1.2/users.create` → Invalid version: `Version must be integer (v0, v1, v2, ...)`
- `/v1/Users.create` → Invalid subject: `Subject must be lowercase`
- `/v1/users/create` → Missing dot: `Subject and method must be separated by dot`
- `/v1/users` → Missing method: `Path must include method after dot`
- `/v1/users/{id}.get` → Path param: `Path parameters not allowed in DUH-RPC`

---

#### REQ-003: Enforce POST-Only HTTP Methods

**Description:** All operations must use POST method exclusively

**Acceptance Criteria:**
- Check each path for defined HTTP methods
- Reject any method other than POST
- Report path and violating method
- If multiple methods on same path, report all non-POST methods

**Valid:**
```yaml
paths:
  /v1/users.create:
    post:  # Only POST allowed
      ...
```

**Invalid:**
```yaml
paths:
  /v1/users.list:
    get:  # Violation
      ...
  /v1/users.update:
    put:  # Violation
      ...
    post:  # Valid, but PUT also present
      ...
```

**Violation Message:**
```
[http-method] {METHOD} {PATH}
  Only POST method is allowed in DUH-RPC
  Found: {METHOD}
  Suggestion: Change {METHOD} to POST and move parameters to request body
```

---

#### REQ-004: Prohibit Query Parameters

**Description:** Query parameters are not allowed in DUH-RPC

**Acceptance Criteria:**
- Check `parameters` array in each operation
- Find parameters with `in: query`
- Report each query parameter found
- Path parameters also not allowed (covered by REQ-002)
- Header and cookie parameters ARE allowed (not restricted)

**Valid:**
```yaml
parameters:
  - name: X-Request-ID
    in: header  # Allowed
    schema:
      type: string
```

**Invalid:**
```yaml
parameters:
  - name: userId
    in: query  # Violation
    schema:
      type: string
  - name: page
    in: query  # Violation
    schema:
      type: integer
```

**Violation Message:**
```
[query-parameters] {PATH}
  Query parameters are not allowed in DUH-RPC
  Found: query parameter "{name}"
  Suggestion: Move "{name}" to request body
```

---

#### REQ-005: Require Request Bodies

**Description:** All operations must have required request body

**Acceptance Criteria:**
- Check each operation has `requestBody` object
- Check `requestBody.required: true` is set
- Report operations missing requestBody
- Report operations with `required: false`

**Valid:**
```yaml
requestBody:
  required: true  # Must be true
  content:
    application/json:
      schema:
        type: object
```

**Invalid:**
```yaml
# Missing requestBody entirely

# OR

requestBody:
  required: false  # Must be true
  content:
    ...
```

**Violation Messages:**
```
[request-body-required] {PATH}
  Request body is required for all DUH-RPC operations
  Found: No request body defined
  Suggestion: Add requestBody with required: true

[request-body-required] {PATH}
  Request body must be required
  Found: required: false
  Suggestion: Set requestBody.required to true
```

---

#### REQ-006: Validate Content Types

**Description:** Only specific content types are allowed

**Allowed Content Types:**
- `application/json` (REQUIRED - must be present)
- `application/protobuf` (OPTIONAL)
- `application/octet-stream` (OPTIONAL)

**Rules:**
1. Operations MAY list multiple allowed content types
2. At minimum, `application/json` MUST be present
3. No other content types allowed
4. No MIME type parameters (e.g., `; charset=utf-8`)
5. Request and response content types CAN differ

**Validation Locations:**
- `requestBody.content` keys
- `responses[status].content` keys

**Valid:**
```yaml
requestBody:
  content:
    application/json:  # Required
      schema: {...}
    application/protobuf:  # Optional
      schema: {...}

responses:
  200:
    content:
      application/json:  # Valid
        schema: {...}
```

**Invalid:**
```yaml
requestBody:
  content:
    application/xml:  # Not allowed
      schema: {...}
    text/html:  # Not allowed
      schema: {...}
    application/json; charset=utf-8:  # Parameters not allowed
      schema: {...}
```

**Content Type Detection:**
- Check exact string match of content type keys
- If semicolon present in key, reject (MIME parameters)

**Violation Messages:**
```
[content-type] {PATH} request body
  Invalid content type: {type}
  Allowed: application/json, application/protobuf, application/octet-stream
  Suggestion: Change to application/json

[content-type] {PATH} response {status}
  Invalid content type: {type}
  Allowed: application/json, application/protobuf, application/octet-stream
  Suggestion: Change to application/json

[content-type] {PATH} request body
  application/json content type is required
  Found: Only application/protobuf defined
  Suggestion: Add application/json as required content type
```

---

#### REQ-007: Validate Error Response Schemas

**Description:** Error responses (4xx, 5xx) must have required schema structure

**Required Schema Structure:**
```yaml
type: object
required: [code, message]
properties:
  code:
    type: integer
  message:
    type: string
  details:
    type: object  # Optional field
```

**Validation Rules:**

1. **Schema must be object type**
   - `type: object` required at root level

2. **Required fields must include `code` and `message`**
   - `required: [code, message]` array must be present

3. **Code field validation:**
   - `properties.code.type: integer` required
   - If `enum` constraint present, check it includes the HTTP status code (optional)
   - Example: For 400 response, `enum: [400]` is valid

4. **Message field validation:**
   - `properties.message.type: string` required

5. **Details field validation (if present):**
   - `properties.details.type: object` required
   - Field is optional (not in required array)

**Schema Reference Resolution:**
- If schema uses `$ref`, resolve the reference and validate
- Example: `$ref: '#/components/schemas/Error'` - resolve and check structure
- Must handle `allOf`, `oneOf`, `anyOf` combinators
- Check that combined/resolved schema satisfies requirements

**Applies to Status Codes:**
- 4xx: 400, 401, 403, 404, 429
- 4xx custom: 452, 453, 454, 455
- 5xx: 500

**Valid Example:**
```yaml
responses:
  400:
    description: Bad Request
    content:
      application/json:
        schema:
          type: object
          required: [code, message]
          properties:
            code:
              type: integer
              enum: [400]  # Optional but recommended
            message:
              type: string
            details:
              type: object

  404:
    description: Not Found
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'  # Must resolve this

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

**Invalid Examples:**
```yaml
# Missing required fields
schema:
  type: object
  properties:
    code:
      type: integer
    # Missing message field

# Wrong types
schema:
  type: object
  required: [code, message]
  properties:
    code:
      type: string  # Should be integer
    message:
      type: string

# Missing required array
schema:
  type: object
  properties:
    code:
      type: integer
    message:
      type: string
    # required array not specified
```

**Violation Messages:**
```
[error-response-schema] {PATH} response {status}
  Error response schema must be an object
  Found: {type}
  Suggestion: Change schema type to object

[error-response-schema] {PATH} response {status}
  Error response must include 'code' and 'message' in required fields
  Found required: {fields}
  Suggestion: Add required: [code, message]

[error-response-schema] {PATH} response {status}
  'code' field must be integer type
  Found: {type}
  Suggestion: Change code type to integer

[error-response-schema] {PATH} response {status}
  'message' field must be string type
  Found: {type}
  Suggestion: Change message type to string

[error-response-schema] {PATH} response {status}
  'details' field must be object type (if present)
  Found: {type}
  Suggestion: Change details type to object
```

---

#### REQ-008: Enforce Allowed HTTP Status Codes

**Description:** Only specific status codes are allowed in DUH-RPC

**Allowed Status Codes:**
- **2xx:** 200 (only)
- **4xx:** 400, 401, 403, 404, 429
- **4xx Custom:** 452, 453, 454, 455 (service-specific errors, no restrictions)
- **5xx:** 500 (only)

**Not Allowed:**
- 201, 202, 203, 204 (use 200)
- 405, 406, 408, 410, 415, etc. (use 400 or defined 4xx codes)
- 501, 502, 503, 504 (use 500)

**Acceptance Criteria:**
- Check all defined response status codes in operation
- Report any codes not in allowed list
- Provide list of allowed codes in error message

**Valid:**
```yaml
responses:
  200:
    description: Success
  400:
    description: Bad Request
  404:
    description: Not Found
  452:
    description: Custom Service Error
  500:
    description: Internal Server Error
```

**Invalid:**
```yaml
responses:
  201:  # Not allowed, use 200
    description: Created
  202:  # Not allowed, use 200
    description: Accepted
  503:  # Not allowed, use 500
    description: Service Unavailable
```

**Violation Message:**
```
[status-code] {PATH} response {status}
  Invalid status code: {status}
  Allowed: 200, 400, 401, 403, 404, 429, 452, 453, 454, 455, 500
  Suggestion: Use 200 for success, 400/4xx for client errors, 500 for server errors
```

---

#### REQ-009: Require Success Responses

**Description:** All operations must define a 200 response with content

**Acceptance Criteria:**
- Check each operation defines `responses.200`
- Check 200 response has `content` object
- Check content has at least one media type with schema
- No restrictions on schema structure (any valid JSON schema)

**Valid:**
```yaml
responses:
  200:
    description: Success
    content:
      application/json:
        schema:
          type: object  # Can be any valid schema
          properties:
            userId:
              type: string

  # OR array
  200:
    content:
      application/json:
        schema:
          type: array
          items:
            type: string

  # OR primitive
  200:
    content:
      application/json:
        schema:
          type: string
```

**Invalid:**
```yaml
responses:
  201:  # 200 missing
    description: Created
    content:
      application/json:
        schema: {...}

  # OR

  200:
    description: Success
    # Missing content

  # OR

  200:
    description: Success
    content:
      application/json:
        # Missing schema
```

**Violation Messages:**
```
[success-response] {PATH}
  200 response is required for all operations
  Found: No 200 response defined
  Suggestion: Add 200 response with content and schema

[success-response] {PATH} response 200
  200 response must have content defined
  Found: No content in 200 response
  Suggestion: Add content with at least application/json

[success-response] {PATH} response 200
  200 response content must have schema defined
  Found: Content without schema
  Suggestion: Add schema to content type
```

---

#### REQ-010: Human-Readable CLI Output

**Description:** Format validation results for terminal display

**Success Output:**
```
✓ openapi.yaml is DUH-RPC compliant
```

**Failure Output Format:**
```
Validating openapi.yaml...

ERRORS FOUND:

[path-format] /api/users/create
  Path must follow format: /v{version}/{subject}.{method}
  Found: /api/users/create
  Suggestion: Change to /v1/users.create

[http-method] GET /v1/users.list
  Only POST method is allowed in DUH-RPC
  Found: GET
  Suggestion: Change to POST and move query parameters to request body

[query-parameters] /v1/search.query
  Query parameters are not allowed in DUH-RPC
  Found: query parameter "q"
  Suggestion: Move all parameters to request body

Summary: 3 violations found in openapi.yaml
```

**Output Requirements:**
- Start with "Validating {filename}..."
- Group violations by rule type (optional, can be path order)
- Each violation shows:
  - `[rule-name]` identifier
  - Location (path + optional context)
  - Message explaining what's wrong
  - Suggestion for how to fix
- End with summary count
- Use symbols: `✓` for success, `✗` or none for failure

**Exit Codes:**
- `0`: Validation passed (no violations)
- `1`: Validation failed (violations found)
- `2`: Tool error (file not found, parse error, internal error)

**Tool Error Output:**
```
Error: File not found: missing.yaml

Error: Failed to parse OpenAPI spec: yaml: line 5: could not find expected ':'
```

**Acceptance Criteria:**
- Clear, scannable output
- Suitable for terminal and CI/CD logs
- Exit codes work correctly for scripting
- Color optional (use if terminal supports, no color in CI)

---

### Non-Functional Requirements

#### NFR-001: Performance

**Requirement:** Parse and validate specs with 100-500 operations in under 2 seconds

**Rationale:**
- Interactive CLI usage requires responsive feedback
- CI/CD pipelines benefit from fast validation
- libopenapi is highly optimized for this scale

**Acceptance Criteria:**
- 100-operation spec: < 1 second
- 500-operation spec: < 2 seconds
- Measured on standard developer hardware (M1/M2 Mac or equivalent)

**Implementation Notes:**
- Rules run sequentially (no concurrency needed for MVP)
- Leverage libopenapi's indexing
- Avoid redundant traversals

---

#### NFR-002: Usability

**Requirement:** Simple, zero-configuration CLI tool

**Acceptance Criteria:**
- Single command: `duhrpc-lint openapi.yaml`
- No configuration files required
- No environment variables required
- Works out of the box after installation
- Includes `--help` for usage guidance
- Includes `--version` for version display

**Installation Methods:**
- `go install github.com/duh-rpc/duhrpc-lint/cmd/duhrpc-lint@latest`
- Direct binary download (future)
- Homebrew (future)

---

#### NFR-003: Maintainability

**Requirement:** Modular, testable, well-documented codebase

**Acceptance Criteria:**
- One rule per file in `internal/rules/`
- Rules are independent (no dependencies between rules)
- 100% test coverage for rules
- Each rule has clear documentation
- Code follows CLAUDE.md conventions
- Easy to add new rules (copy pattern)

---

## Technical Architecture

### Technology Stack

**Language:** Go 1.21+

**Dependencies:**
- `github.com/pb33f/libopenapi` v0.18.0+ (OpenAPI parsing)
- `github.com/stretchr/testify` (testing - require/assert)
- Go standard library (flag, fmt, os, etc.)

**No Dependencies On:**
- CLI frameworks (cobra, etc.) - use stdlib `flag`
- Logging frameworks - use stdlib `log`
- Config libraries - no config needed

---

### Module Structure

```
duhrpc-lint/
├── cmd/
│   └── duhrpc-lint/
│       └── main.go              # CLI entry, arg parsing, orchestration
│
├── internal/
│   ├── loader/
│   │   ├── loader.go            # OpenAPI file loading
│   │   └── loader_test.go       # File and parse error tests
│   │
│   ├── validator/
│   │   ├── validator.go         # Core validation engine
│   │   └── validator_test.go    # Integration tests
│   │
│   ├── rules/
│   │   ├── rule.go              # Rule interface definition
│   │   │
│   │   ├── path_format.go       # REQ-002 implementation
│   │   ├── path_format_test.go
│   │   │
│   │   ├── http_method.go       # REQ-003 implementation
│   │   ├── http_method_test.go
│   │   │
│   │   ├── query_params.go      # REQ-004 implementation
│   │   ├── query_params_test.go
│   │   │
│   │   ├── request_body.go      # REQ-005 implementation
│   │   ├── request_body_test.go
│   │   │
│   │   ├── content_type.go      # REQ-006 implementation
│   │   ├── content_type_test.go
│   │   │
│   │   ├── error_response.go    # REQ-007 implementation
│   │   ├── error_response_test.go
│   │   │
│   │   ├── status_code.go       # REQ-008 implementation
│   │   ├── status_code_test.go
│   │   │
│   │   ├── success_response.go  # REQ-009 implementation
│   │   └── success_response_test.go
│   │
│   ├── reporter/
│   │   ├── reporter.go          # Output formatting (REQ-010)
│   │   └── reporter_test.go     # Format tests
│   │
│   └── types/
│       ├── types.go             # Violation, ValidationResult
│       └── types_test.go        # Type tests
│
├── docs/
│   ├── TECHNICAL_SPEC.md        # This file
│   └── duhrpc-openapi-rules.md  # Rule reference for API designers
│
├── testdata/
│   ├── valid-spec.yaml                          # Fully compliant
│   └── invalid-specs/
│       ├── bad-path-format.yaml                 # REQ-002 violations
│       ├── wrong-http-method.yaml               # REQ-003 violations
│       ├── has-query-params.yaml                # REQ-004 violations
│       ├── missing-request-body.yaml            # REQ-005 violations
│       ├── invalid-content-type.yaml            # REQ-006 violations
│       ├── bad-error-schema.yaml                # REQ-007 violations
│       ├── invalid-status-code.yaml             # REQ-008 violations
│       ├── missing-success-response.yaml        # REQ-009 violations
│       └── multiple-violations.yaml             # Mixed violations
│
├── go.mod
├── go.sum
├── README.md                    # Usage, installation, examples
├── Makefile                     # Build, test, install targets
└── .gitignore
```

---

### Core Type Definitions

#### types/types.go

```go
package types

import (
    "fmt"
    "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Violation represents a single DUH-RPC compliance violation
type Violation struct {
    RuleName   string  // Rule identifier (e.g., "path-format")
    Location   string  // Path or operation (e.g., "/v1/users.create", "GET /api/users")
    Message    string  // What's wrong (e.g., "Path must follow format...")
    Suggestion string  // How to fix (e.g., "Change to /v1/users.create")
}

// String formats violation for display
func (v Violation) String() string {
    return fmt.Sprintf("[%s] %s\n  %s\n  %s", v.RuleName, v.Location, v.Message, v.Suggestion)
}

// ValidationResult contains all violations found in a document
type ValidationResult struct {
    Violations []Violation
    FilePath   string
}

// Valid returns true if no violations found
func (vr ValidationResult) Valid() bool {
    return len(vr.Violations) == 0
}

// Rule interface that all validation rules must implement
type Rule interface {
    // Name returns the rule identifier (e.g., "path-format")
    Name() string

    // Validate checks the document and returns violations
    Validate(doc *v3.Document) []Violation
}
```

---

### Component Responsibilities

#### 1. main.go (cmd/duhrpc-lint/main.go)

**Responsibilities:**
- Parse CLI arguments (positional file path, --help, --version)
- Call loader to load OpenAPI file
- Call validator to run validation
- Call reporter to format output
- Handle exit codes

**Pseudocode:**
```go
func main() {
    // Parse args
    if --help { showHelp(); return }
    if --version { showVersion(); return }
    filePath := args[0]

    // Load document
    doc, err := loader.Load(filePath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(2)
    }

    // Validate
    result := validator.Validate(doc, filePath)

    // Report
    reporter.Print(result)

    // Exit
    if !result.Valid() {
        os.Exit(1)
    }
}
```

---

#### 2. loader/loader.go

**Responsibilities:**
- Check file exists
- Read file contents
- Parse with libopenapi
- Build v3.Document model
- Return parsed document or error

**Pseudocode:**
```go
func Load(filePath string) (*v3.Document, error) {
    // Check exists
    if !fileExists(filePath) {
        return nil, fmt.Errorf("File not found: %s", filePath)
    }

    // Read file
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("Failed to read file: %w", err)
    }

    // Parse OpenAPI
    doc, err := libopenapi.NewDocument(content)
    if err != nil {
        return nil, fmt.Errorf("Failed to parse OpenAPI spec: %w", err)
    }

    // Build v3 model
    model, errs := doc.BuildV3Model()
    if len(errs) > 0 {
        return nil, fmt.Errorf("Invalid OpenAPI 3.0 spec: %v", errs[0])
    }

    return model.Model, nil
}
```

---

#### 3. validator/validator.go

**Responsibilities:**
- Register all validation rules
- Execute each rule against document
- Collect violations from all rules
- Return ValidationResult

**Pseudocode:**
```go
func Validate(doc *v3.Document, filePath string) types.ValidationResult {
    // Create rule instances
    allRules := []types.Rule{
        rules.NewPathFormatRule(),
        rules.NewHTTPMethodRule(),
        rules.NewQueryParamsRule(),
        rules.NewRequestBodyRule(),
        rules.NewContentTypeRule(),
        rules.NewErrorResponseRule(),
        rules.NewStatusCodeRule(),
        rules.NewSuccessResponseRule(),
    }

    // Collect violations
    var violations []types.Violation
    for _, rule := range allRules {
        ruleViolations := rule.Validate(doc)
        violations = append(violations, ruleViolations...)
    }

    return types.ValidationResult{
        Violations: violations,
        FilePath:   filePath,
    }
}
```

---

#### 4. rules/*.go (8 rule files)

**Responsibilities:**
- Implement Rule interface
- Traverse OpenAPI document structure
- Check for violations of specific rule
- Return list of violations with clear messages

**Example Pattern (path_format.go):**
```go
type PathFormatRule struct{}

func NewPathFormatRule() *PathFormatRule {
    return &PathFormatRule{}
}

func (r *PathFormatRule) Name() string {
    return "path-format"
}

func (r *PathFormatRule) Validate(doc *v3.Document) []types.Violation {
    const pathRegex = `^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]*\.[a-z][a-z0-9_-]*$`
    re := regexp.MustCompile(pathRegex)

    var violations []types.Violation

    for path := range doc.Paths.PathItems {
        if !re.MatchString(path) {
            violations = append(violations, types.Violation{
                RuleName:   r.Name(),
                Location:   path,
                Message:    "Path must follow format: /v{version}/{subject}.{method}",
                Suggestion: fmt.Sprintf("Change to %s", suggestCorrection(path)),
            })
        }
    }

    return violations
}
```

---

#### 5. reporter/reporter.go

**Responsibilities:**
- Format ValidationResult for output
- Print violations in human-readable format
- Show summary count
- Handle success vs failure display

**Pseudocode:**
```go
func Print(result types.ValidationResult) {
    if result.Valid() {
        fmt.Printf("✓ %s is DUH-RPC compliant\n", filepath.Base(result.FilePath))
        return
    }

    fmt.Printf("Validating %s...\n\n", filepath.Base(result.FilePath))
    fmt.Println("ERRORS FOUND:\n")

    for _, v := range result.Violations {
        fmt.Println(v.String())
        fmt.Println()
    }

    fmt.Printf("Summary: %d violations found in %s\n",
        len(result.Violations),
        filepath.Base(result.FilePath))
}
```

---

## Validation Rules Reference

### Summary Table

| Rule ID | Name | Severity | Requirement |
|---------|------|----------|-------------|
| REQ-002 | path-format | Error | Path must match `/v{version}/{subject}.{method}` |
| REQ-003 | http-method-post-only | Error | Only POST method allowed |
| REQ-004 | no-query-parameters | Error | No query parameters allowed |
| REQ-005 | request-body-required | Error | Request body must be required |
| REQ-006 | content-type-allowed | Error | Only JSON/protobuf/octet-stream |
| REQ-007 | error-response-schema | Error | Error schema must have code/message |
| REQ-008 | status-code-allowed | Error | Only allowed status codes |
| REQ-009 | success-response-required | Error | 200 response required |

### Implementation Complexity

**Simple Rules (1-2 hours each):**
- REQ-003: http-method-post-only
- REQ-004: no-query-parameters
- REQ-005: request-body-required
- REQ-008: status-code-allowed
- REQ-009: success-response-required

**Medium Rules (2-4 hours each):**
- REQ-006: content-type-allowed

**Complex Rules (4-6 hours each):**
- REQ-002: path-format (regex, multiple checks, good suggestions)
- REQ-007: error-response-schema ($ref resolution, combinators)

---

## Examples

### Example 1: Fully Valid DUH-RPC Spec

**File:** `testdata/valid-spec.yaml`

```yaml
openapi: 3.0.0
info:
  title: Valid DUH-RPC API
  version: 1.0.0
  description: Example of fully compliant DUH-RPC OpenAPI specification

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
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                    enum: [400]
                  message:
                    type: string
                  details:
                    type: object
        500:
          description: Internal Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                    enum: [500]
                  message:
                    type: string
                  details:
                    type: object

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

**Expected Output:**
```
✓ valid-spec.yaml is DUH-RPC compliant
```

---

### Example 2: Multiple Violations

**File:** `testdata/invalid-specs/multiple-violations.yaml`

```yaml
openapi: 3.0.0
info:
  title: Invalid DUH-RPC API
  version: 1.0.0

paths:
  /api/users/create:  # Violation: Wrong path format
    get:  # Violation: Wrong HTTP method
      operationId: createUser
      parameters:
        - name: userId  # Violation: Query parameter
          in: query
          schema:
            type: string
      responses:
        201:  # Violation: Invalid status code
          description: Created
          content:
            application/xml:  # Violation: Invalid content type
              schema:
                type: object

  /v1/products.list:
    post:
      operationId: listProducts
      # Violation: Missing request body
      responses:
        200:
          description: Success
          # Violation: Missing content in 200
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                # Violation: Missing required fields in error schema
                properties:
                  error:
                    type: string
```

**Expected Output:**
```
Validating multiple-violations.yaml...

ERRORS FOUND:

[path-format] /api/users/create
  Path must follow format: /v{version}/{subject}.{method}
  Found: /api/users/create
  Suggestion: Change to /v1/users.create

[http-method] GET /api/users/create
  Only POST method is allowed in DUH-RPC
  Found: GET
  Suggestion: Change to POST and move parameters to request body

[query-parameters] /api/users/create
  Query parameters are not allowed in DUH-RPC
  Found: query parameter "userId"
  Suggestion: Move "userId" to request body

[status-code] /api/users/create response 201
  Invalid status code: 201
  Allowed: 200, 400, 401, 403, 404, 429, 452, 453, 454, 455, 500
  Suggestion: Use 200 for success responses

[content-type] /api/users/create response 201
  Invalid content type: application/xml
  Allowed: application/json, application/protobuf, application/octet-stream
  Suggestion: Change to application/json

[request-body-required] /v1/products.list
  Request body is required for all DUH-RPC operations
  Found: No request body defined
  Suggestion: Add requestBody with required: true

[success-response] /v1/products.list response 200
  200 response must have content defined
  Found: No content in 200 response
  Suggestion: Add content with at least application/json

[error-response-schema] /v1/products.list response 400
  Error response must include 'code' and 'message' in required fields
  Found required: []
  Suggestion: Add required: [code, message]

Summary: 8 violations found in multiple-violations.yaml
```

---

## Testing Strategy

### Unit Testing

**Test Package Convention:**
- Tests in `package XXX_test` (not `package XXX`)
- Per CLAUDE.md guidelines

**Table-Driven Tests:**
```go
func TestPathFormatRule(t *testing.T) {
    for _, test := range []struct {
        name          string
        path          string
        wantViolation bool
        wantMessage   string
    }{
        {
            name:          "valid v1 path",
            path:          "/v1/users.create",
            wantViolation: false,
        },
        {
            name:          "valid v0 path",
            path:          "/v0/beta.test",
            wantViolation: false,
        },
        {
            name:          "valid multi-word",
            path:          "/v1/user-accounts.get-by-id",
            wantViolation: false,
        },
        {
            name:          "missing version",
            path:          "/users.create",
            wantViolation: true,
            wantMessage:   "Path must start with /v{version}/",
        },
        {
            name:          "non-numeric version",
            path:          "/vbeta/users.create",
            wantViolation: true,
            wantMessage:   "Version must be numeric",
        },
        {
            name:          "uppercase subject",
            path:          "/v1/Users.create",
            wantViolation: true,
            wantMessage:   "must be lowercase",
        },
        {
            name:          "missing dot separator",
            path:          "/v1/users",
            wantViolation: true,
            wantMessage:   "must include method after dot",
        },
        {
            name:          "path parameter",
            path:          "/v1/users/{id}.get",
            wantViolation: true,
            wantMessage:   "Path parameters not allowed",
        },
    } {
        t.Run(test.name, func(t *testing.T) {
            doc := createMockDoc(test.path)
            rule := rules.NewPathFormatRule()

            violations := rule.Validate(doc)

            if test.wantViolation {
                require.Len(t, violations, 1)
                require.Contains(t, violations[0].Message, test.wantMessage)
                require.Equal(t, "path-format", violations[0].RuleName)
            } else {
                require.Empty(t, violations)
            }
        })
    }
}
```

**Coverage Requirements:**
- All rules: 100% coverage
- Core validator: 100% coverage
- Loader: 90%+ coverage
- Reporter: 90%+ coverage

---

### Integration Testing

**End-to-End CLI Tests:**

```go
func TestCLIValidSpec(t *testing.T) {
    const validSpec = "../../testdata/valid-spec.yaml"

    cmd := exec.Command("./duhrpc", validSpec)
    output, err := cmd.CombinedOutput()

    require.NoError(t, err)
    require.Contains(t, string(output), "DUH-RPC compliant")
    require.Contains(t, string(output), "✓")
}

func TestCLIInvalidSpec(t *testing.T) {
    const invalidSpec = "../../testdata/invalid-specs/bad-path-format.yaml"

    cmd := exec.Command("./duhrpc", invalidSpec)
    output, err := cmd.CombinedOutput()

    require.Error(t, err)
    require.Contains(t, string(output), "[path-format]")
    require.Contains(t, string(output), "violations found")

    exitErr := err.(*exec.ExitError)
    require.Equal(t, 1, exitErr.ExitCode())
}

func TestCLIFileNotFound(t *testing.T) {
    cmd := exec.Command("./duhrpc", "nonexistent.yaml")
    output, err := cmd.CombinedOutput()

    require.Error(t, err)
    require.Contains(t, string(output), "File not found")

    exitErr := err.(*exec.ExitError)
    require.Equal(t, 2, exitErr.ExitCode())
}
```

---

### Test Data Requirements

**Valid Specs (1 file):**
- `valid-spec.yaml` - Passes all 8 rules

**Invalid Specs (9 files):**
- `bad-path-format.yaml` - Multiple path format violations
- `wrong-http-method.yaml` - GET, PUT, DELETE methods
- `has-query-params.yaml` - Query parameters defined
- `missing-request-body.yaml` - No requestBody or required: false
- `invalid-content-type.yaml` - application/xml, text/html
- `bad-error-schema.yaml` - Missing code/message, wrong types
- `invalid-status-code.yaml` - 201, 202, 503 responses
- `missing-success-response.yaml` - No 200 response
- `multiple-violations.yaml` - Mixed violations (comprehensive)

---

## Implementation Guidance

### Phase 1: Foundation (1-2 days)

**Day 1: Project Setup**
1. Initialize Go module: `go mod init github.com/duh-rpc/duhrpc-lint`
2. Create directory structure (cmd/, internal/, docs/, testdata/)
3. Add go.mod dependencies:
   - `go get github.com/pb33f/libopenapi@latest`
   - `go get github.com/stretchr/testify@latest`
4. Create types/types.go (Violation, ValidationResult, Rule interface)
5. Write types tests

**Day 2: Core Infrastructure**
6. Implement loader/loader.go (file loading, OpenAPI parsing)
7. Write loader tests (valid YAML, file not found, parse errors)
8. Implement reporter/reporter.go (output formatting)
9. Write reporter tests (success output, error output)
10. Create main.go skeleton (arg parsing, help/version, orchestration)
11. Test end-to-end with empty validator

---

### Phase 2: Core Validation Rules (3-4 days)

**Day 3: Simple Rules**
12. Implement http_method.go (REQ-003)
13. Write http_method tests (POST valid, GET/PUT/DELETE invalid)
14. Implement query_params.go (REQ-004)
15. Write query_params tests
16. Implement request_body.go (REQ-005)
17. Write request_body tests

**Day 4: Path Format Rule**
18. Implement path_format.go (REQ-002) - MOST COMPLEX
    - Version regex validation
    - Subject/method regex validation
    - Dot separator check
    - Path parameter detection
    - Suggestion generation
19. Write comprehensive path_format tests (10+ cases)
20. Create first integration test (valid-spec.yaml passes)

**Day 5: Status Code Rules**
21. Implement status_code.go (REQ-008)
22. Write status_code tests
23. Implement success_response.go (REQ-009)
24. Write success_response tests
25. Create integration tests for simple rule violations

---

### Phase 3: Advanced Rules (2-3 days)

**Day 6: Content Type Validation**
26. Implement content_type.go (REQ-006)
    - Check allowed types
    - Detect MIME parameters (semicolon)
    - Validate both request and response content
27. Write content_type tests

**Day 7-8: Error Response Schema**
28. Implement error_response.go (REQ-007) - MOST COMPLEX
    - Schema structure validation
    - $ref resolution using libopenapi
    - allOf/oneOf/anyOf handling
    - Required fields check
    - Type validation (code: integer, message: string)
    - Optional enum check for code
29. Write comprehensive error_response tests
    - Inline schemas
    - $ref schemas
    - allOf combinators
    - Missing required fields
    - Wrong types

---

### Phase 4: Integration & Documentation (1-2 days)

**Day 9: Integration**
30. Wire all 8 rules into validator.go
31. Create all 9 test specs in testdata/
32. Write comprehensive CLI integration tests
33. Test exit codes (0, 1, 2)
34. Test output formatting
35. Performance testing (validate 100-operation spec)

**Day 10: Documentation**
36. Write docs/duhrpc-openapi-rules.md (rule reference)
37. Write README.md (installation, usage, examples)
38. Add inline code documentation
39. Create Makefile (build, test, install, lint)
40. Final testing across all specs

---

### Code Style Checklist (CLAUDE.md)

When implementing, follow these conventions:

**Tests:**
- [ ] Package name: `package XXX_test` (not `package XXX`)
- [ ] Test names: camelCase starting with capital (e.g., `TestPathFormatRule`)
- [ ] Table-driven tests with `for _, test := range []struct {`
- [ ] Use `require` for critical assertions (stops test)
- [ ] Use `assert` for non-critical assertions (continues test)
- [ ] No explanatory messages in require/assert (e.g., no `require.NoError(t, err, "should parse")`)

**Code:**
- [ ] Use `const` for unchanging values used multiple times
- [ ] Inline struct literals if used once
- [ ] Short variable names (1-2 words)
- [ ] No abbreviations (use full words)
- [ ] Visual tapering for struct fields (longest lines first)
- [ ] No single-use variables (inline into function calls)

**Example Visual Tapering:**
```go
violation := types.Violation{
    RuleName:   "error-response-schema",  // Longest
    Location:   fmt.Sprintf("%s response %d", path, status),
    Message:    "Error response missing required fields",  // Medium
    Suggestion: "Add required: [code, message]",  // Shortest
}
```

---

### Implementation Priority

**Critical Path (Must Have for MVP):**
1. Loader (can't validate without parsing)
2. Types (all components depend on)
3. Validator (orchestration)
4. Reporter (need output)
5. Main (CLI entry point)

**Rule Priority (Implement in Order):**
1. path-format (most important DUH-RPC rule)
2. http-method-post-only (fundamental RPC requirement)
3. request-body-required (DUH-RPC semantics)
4. success-response-required (basic validation)
5. status-code-allowed (DUH-RPC constraint)
6. content-type-allowed (DUH-RPC standard)
7. no-query-parameters (DUH-RPC convention)
8. error-response-schema (nice-to-have, complex)

**Rationale:** Implement high-value, simple rules first. Save complex rules (error-response-schema) for last when infrastructure is solid.

---

### Technical Challenges & Solutions

**Challenge 1: Path Format Regex Complexity**
- **Issue:** Multiple validation aspects (version, subject, method, separator)
- **Solution:** Break into multiple regex checks with specific error messages
  - Check version: `^/v(0|[1-9][0-9]*)/`
  - Check subject: `[a-z][a-z0-9_-]*`
  - Check method: `\.[a-z][a-z0-9_-]*$`
  - Provide specific error for each failure

**Challenge 2: $ref Resolution in Error Schemas**
- **Issue:** Schemas may reference components/schemas
- **Solution:** Use libopenapi's built-in reference resolution
  - libopenapi provides `Index` with resolved references
  - Traverse schema tree to resolve `$ref` pointers
  - Check resolved schema structure

**Challenge 3: allOf/oneOf/anyOf Combinators**
- **Issue:** Complex schema composition
- **Solution:** Recursive schema validation
  - For `allOf`: All schemas must satisfy requirements
  - For `oneOf`/`anyOf`: At least one must satisfy
  - Flatten combined schema and check structure

**Challenge 4: Suggestion Generation for Path Format**
- **Issue:** Hard to suggest correct path from invalid path
- **Solution:** Pattern matching and heuristics
  - If missing version, prepend `/v1`
  - If has path params, show static example
  - If wrong separator, replace with dot
  - Keep simple - suggest most common correction

---

### Performance Considerations

**Expected Performance:**
- 100 operations: < 1 second
- 500 operations: < 2 seconds

**Optimization Strategies:**
1. **Single document traversal** - Rules share same iteration where possible
2. **Lazy evaluation** - Don't resolve $ref unless needed
3. **No redundant parsing** - Parse once, validate many times
4. **Compiled regexes** - Compile path regex once, reuse for all paths

**Measurement:**
```go
func BenchmarkValidate100Ops(b *testing.B) {
    doc := loadTestDoc("testdata/large-spec-100-ops.yaml")
    validator := validator.New()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = validator.Validate(doc, "test.yaml")
    }
}
```

---

### Error Handling Strategy

**File Errors (Exit Code 2):**
- File not found: Check with `os.Stat()` before reading
- Permission denied: Catch and report clearly
- Message format: `Error: {issue}: {path}`

**Parse Errors (Exit Code 2):**
- YAML syntax errors: Include line/column from parser
- Not OpenAPI 3.0: Check version and report
- Invalid structure: Report specific issue from libopenapi
- Message format: `Error: Failed to parse OpenAPI spec: {details}`

**Validation Errors (Exit Code 1):**
- Collect ALL violations before reporting
- Don't stop at first error (give complete picture)
- Group by rule or path (either works)
- Format each violation consistently

**Internal Errors (Exit Code 2):**
- Panic recovery: `defer recover()` in main
- Unexpected errors: Log and exit gracefully
- Message format: `Error: Internal error: {details}`

---

## Future Enhancements (Out of Scope for MVP)

These features are explicitly NOT in MVP, but documented for future consideration:

### Phase 2 Features
- JSON input support (in addition to YAML)
- OpenAPI 3.1 support
- OpenAPI 2.0 (Swagger) support
- `--format json` flag for machine-readable output
- `--quiet` flag to suppress non-error output

### Phase 3 Features
- Auto-fix mode: Generate corrected OpenAPI spec
- IDE integration via Language Server Protocol (LSP)
- GitHub Action for CI/CD
- Pre-commit hook integration
- Configuration file support (`.duhrpc-lint.yaml`)
- Rule severity customization
- Rule exclusion/inclusion

### Phase 4 Features
- OpenAPI document generation from templates
- Code generator integration
- Batch validation (multiple files)
- Watch mode for development
- HTML report output
- Detailed statistics (compliance score)

---

## Appendix A: Complete Example OpenAPI Specs

### A.1: Minimal Valid Spec

**Use Case:** Smallest possible DUH-RPC compliant spec

```yaml
openapi: 3.0.0
info:
  title: Minimal DUH-RPC API
  version: 1.0.0

paths:
  /v1/ping.test:
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
        500:
          description: Error
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
                  details:
                    type: object
```

### A.2: Comprehensive Valid Spec

**Use Case:** Demonstrates all features and edge cases

```yaml
openapi: 3.0.0
info:
  title: Comprehensive DUH-RPC API
  version: 1.0.0
  description: Shows all DUH-RPC features

paths:
  /v1/users.create:
    post:
      summary: Create user (JSON only)
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
          $ref: '#/components/responses/BadRequest'
        500:
          $ref: '#/components/responses/ServerError'

  /v2/messages.send:
    post:
      summary: Send message (multiple content types)
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
          application/protobuf:
            schema:
              type: string
              format: binary
          application/octet-stream:
            schema:
              type: string
              format: binary
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
        401:
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        429:
          description: Rate Limited
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/Error'
                  - type: object
                    properties:
                      retryAfter:
                        type: integer
        500:
          $ref: '#/components/responses/ServerError'

  /v0/beta.feature-test:
    post:
      summary: Beta feature (v0 allowed)
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
                type: string
        452:
          description: Custom service error (452-455 allowed)
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                    enum: [452]
                  message:
                    type: string
                  details:
                    type: object
        500:
          $ref: '#/components/responses/ServerError'

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

  responses:
    BadRequest:
      description: Bad Request
      content:
        application/json:
          schema:
            type: object
            required: [code, message]
            properties:
              code:
                type: integer
                enum: [400]
              message:
                type: string
              details:
                type: object

    ServerError:
      description: Internal Server Error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

---

## Appendix B: CLI Usage Examples

### B.1: Basic Usage

```bash
# Validate a spec
$ duhrpc-lint openapi.yaml
✓ openapi.yaml is DUH-RPC compliant

# Validate with violations
$ duhrpc bad-spec.yaml
Validating bad-spec.yaml...

ERRORS FOUND:

[path-format] /api/users
  Path must follow format: /v{version}/{subject}.{method}
  Found: /api/users
  Suggestion: Change to /v1/users.create

Summary: 1 violation found in bad-spec.yaml

# Check exit code
$ echo $?
1

# File not found
$ duhrpc missing.yaml
Error: File not found: missing.yaml

$ echo $?
2
```

### B.2: Help and Version

```bash
# Show help
$ duhrpc-lint --help
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

Examples:
  duhrpc-lint openapi.yaml
  duhrpc-lint api/spec.yaml

Exit Codes:
  0    Validation passed (spec is DUH-RPC compliant)
  1    Validation failed (violations found)
  2    Error (file not found, parse error, etc.)

Documentation:
  https://github.com/duh-rpc/duhrpc-lint

# Show version
$ duhrpc-lint --version
duhrpc-lint version 1.0.0
```

### B.3: CI/CD Usage

```bash
# In GitHub Actions / CI pipeline
- name: Validate OpenAPI Spec
  run: |
    duhrpc-lint api/openapi.yaml
    if [ $? -eq 0 ]; then
      echo "✓ API spec is DUH-RPC compliant"
    else
      echo "✗ API spec has DUH-RPC violations"
      exit 1
    fi

# In Makefile
validate-api:
	@duhrpc-lint api/openapi.yaml

# In pre-commit hook
#!/bin/bash
if [ -f api/openapi.yaml ]; then
  duhrpc-lint api/openapi.yaml || exit 1
fi
```

---

## Appendix C: Development Commands

### C.1: Makefile

```makefile
.PHONY: build test install lint clean

# Build binary
build:
	go build -o duhrpc-lint ./cmd/duhrpc-lint

# Run tests
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

# Run integration tests
integration-test: build
	./duhrpc-lint testdata/valid-spec.yaml
	! ./duhrpc-lint testdata/invalid-specs/bad-path-format.yaml

# Coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Benchmark
bench:
	go test -bench=. -benchmem ./...
```

### C.2: Common Development Commands

```bash
# Initialize project
go mod init github.com/duh-rpc/duhrpc-lint
go get github.com/pb33f/libopenapi@latest
go get github.com/stretchr/testify@latest

# Run tests
go test ./...
go test -v ./internal/rules
go test -run TestPathFormatRule ./internal/rules

# Build and run
go build ./cmd/duhrpc-lint
./duhrpc-lint testdata/valid-spec.yaml

# Install locally
go install ./cmd/duhrpc-lint
duhrpc-lint testdata/valid-spec.yaml

# Coverage
go test -cover ./...
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out

# Format code
go fmt ./...
gofmt -s -w .

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

---

## Document Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-22 | Initial specification approved for implementation |

---

## Next Steps

This technical specification is complete and ready for implementation planning phase.

**To begin implementation:**

1. **Create new session** with implementation planning agent
2. **Provide this document** as context
3. **Request:** "Create detailed implementation plan for duhrpc-lint based on TECHNICAL_SPEC.md"

**Implementation planning phase should produce:**
- Day-by-day implementation tasks
- File-by-file creation order
- Stub code for each component
- Test data creation plan
- Integration test scenarios

**Estimated timeline:**
- Implementation planning: 1 day
- Implementation: 7-11 days
- Testing & documentation: 1-2 days
- **Total: 9-14 days from spec to completion**

---

END OF TECHNICAL SPECIFICATION
