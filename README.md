# duh

Command-line tools for working with DUH-RPC specifications.

[![Go Version](https://img.shields.io/github/go-mod/go-version/duh-rpc/duh-cli)](https://golang.org/dl/)
[![CI Status](https://github.com/duh-rpc/duh-cli/workflows/CI/badge.svg)](https://github.com/duh-rpc/duh-cli/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/duh-rpc/duh-cli)](https://goreportcard.com/report/github.com/duh-rpc/duh-cli)

## Overview

`duh` is a command-line tool for working with DUH-RPC specifications. It provides commands for:
- **Creating** OpenAPI specifications that follow DUH-RPC conventions
- **Validating** OpenAPI YAML specifications against DUH-RPC requirements
- **Generating** Go client, server, and type code from OpenAPI specifications
- **Adding** new endpoints to existing OpenAPI specifications

The tool ensures your API specifications follow the Document-Unified HTTP RPC pattern with clear error messages and actionable suggestions when violations are found.

For detailed DUH-RPC specification requirements and examples, see the [DUH-RPC OpenAPI Reference](docs/duh-openapi-reference.md).

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

### Initialize a New Specification

Create a new DUH-RPC compliant OpenAPI specification template:

```bash
# Creates openapi.yaml in current directory
duh init

# Create with custom filename
duh init my-api.yaml
```

The template includes a complete example with users.create, users.get, users.list, and users.update endpoints demonstrating all DUH-RPC requirements.

### Validate a Specification

```bash
duh lint <openapi-file>
```

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

### Add a New Endpoint

Add a new DUH-RPC endpoint to an existing specification:

```bash
# Add endpoint to openapi.yaml (default)
duh add /v1/products.create CreateProduct

# Add to custom spec file
duh add /v1/orders.cancel CancelOrder -f api/openapi.yaml
```

This creates a new POST endpoint with placeholder request and response schemas.

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

### Exit Codes

- **0**: Success (validation passed, code generated, endpoint added)
- **1**: Validation failed (violations found in lint command)
- **2**: Error occurred (file not found, parse error, invalid arguments, etc.)

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