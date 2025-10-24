# duh

Command-line tools for working with DUH-RPC specifications.

## Overview

`duh` is a command-line tool for working with DUH-RPC specifications. It provides commands for validating OpenAPI YAML specifications against DUH-RPC conventions, ensuring your API specifications follow the Document-Unified HTTP RPC pattern with clear error messages and actionable suggestions when violations are found.

## Installation

### Using go install

```bash
go install github.com/duh-rpc/duh-cli/cmd/duh@latest
```

### From source

```bash
git clone https://github.com/duh-rpc/duh-cli.git
cd duh-cli
make install
```

### Building locally

```bash
make build
# Binary will be created as ./duh
```

## Usage

### Basic Usage

```bash
duh lint <openapi-file>
```

### Examples

**Validate a compliant specification:**
```bash
duh lint api-spec.yaml
```
Output:
```
✓ api-spec.yaml is DUH-RPC compliant
```

**Validate a specification with violations:**
```bash
duh lint api-spec.yaml
```
Output:
```
Validating api-spec.yaml...

ERRORS FOUND:

[path-format] /users.create
  Path must start with version prefix (e.g., /v1/)
  Suggestion: Change path to /v1/users.create

[http-method] GET /v1/users.list
  DUH-RPC only allows POST method
  Suggestion: Change GET to POST

Summary: 2 violations found in api-spec.yaml
```

### Generate HTTP Client

```bash
# Use default openapi.yaml, output to client.go
duh generate client

# Specify custom spec file
duh generate client api/openapi.yaml

# Custom output location and package
duh generate client -o pkg/client/client.go -p client
```

### Generate Server Stubs

```bash
# Use default openapi.yaml, output to server.go
duh generate server

# Specify custom spec file
duh generate server api/openapi.yaml

# Custom output location and package
duh generate server -o pkg/server/server.go -p server
```

### Generate Type Models

```bash
# Use default openapi.yaml, output to models.go
duh generate models

# Specify custom spec file
duh generate models api/openapi.yaml

# Custom output location and package
duh generate models -o pkg/types/models.go -p types
```

### Command-line Options

```bash
# Show help
duh lint --help

# Show version
duh --version
```

### Exit Codes

- **0**: Validation passed (spec is DUH-RPC compliant)
- **1**: Validation failed (violations found)
- **2**: Error occurred (file not found, parse error, etc.)

## Validation Rules

`duh lint` validates against 8 DUH-RPC requirements:

1. **Path Format (REQ-002)**: Paths must follow `/v{N}/{subject}.{method}` format
   - Must start with version prefix like `/v1/`
   - Subject and method must be lowercase alphanumeric with hyphens/underscores
   - No path parameters allowed

2. **HTTP Method (REQ-003)**: Only POST method is allowed
   - All operations must use POST, not GET, PUT, DELETE, etc.

3. **Query Parameters (REQ-004)**: Query parameters are not allowed
   - All input must be in the request body

4. **Request Body Required (REQ-005)**: All operations must have a required request body
   - `requestBody.required` must be `true`

5. **Content Type (REQ-006)**: Only specific content types allowed
   - Allowed: `application/json`, `application/protobuf`, `application/octet-stream`
   - MIME parameters (like charset) not allowed
   - `application/json` must be present in request body

6. **Error Response Schema (REQ-007)**: Error responses must have specific structure
   - Must be type `object` with required fields: `code` (integer), `message` (string)
   - Optional `details` field must be type `object`
   - Applies to status codes: 400, 401, 403, 404, 429, 452-455, 500

7. **Status Code (REQ-008)**: Only specific status codes allowed
   - Allowed: 200, 400, 401, 403, 404, 429, 452, 453, 454, 455, 500

8. **Success Response (REQ-009)**: 200 response required with content
   - All operations must define a 200 response
   - Response must have content with a schema

For detailed specifications, see the [technical specification document](docs/TECHNICAL_SPEC.md).

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run integration tests
make integration-test
```

### Building

```bash
# Build binary
make build

# Install to GOPATH/bin
make install

# Clean build artifacts
make clean
```

### Code Coverage

```bash
# Generate coverage report
make coverage

# Opens coverage.html in your browser
```

### Linting

```bash
# Run code linters
make lint
```