# duh

Command-line tools for working with DUH-RPC specifications.

[![Go Version](https://img.shields.io/github/go-mod/go-version/duh-rpc/duh-cli)](https://golang.org/dl/)
[![CI Status](https://github.com/duh-rpc/duh-cli/workflows/CI/badge.svg)](https://github.com/duh-rpc/duh-cli/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/duh-rpc/duh-cli)](https://goreportcard.com/report/github.com/duh-rpc/duh-cli)

## Overview

`duh` is a command-line tool for working with OpenAPI specifications. It provides commands for:
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

## Quick Start

Get a DUH-RPC service up and running in minutes:

```bash
# 1. Create a new directory for your service
mkdir my-service && cd my-service

# 2. Initialize a new DUH-RPC OpenAPI specification
duh init

# 3. Initialize Go module
go mod init github.com/my-org/my-service

# 4. Generate complete service scaffolding (client, server, daemon, tests, Makefile)
duh generate --full

# 5. Install dependencies
go mod tidy

# 6. Generate Go code from protobuf definitions
buf generate

# 7. Run tests to verify everything works
make test
```

**Success!** You now have a fully functional DUH-RPC service with:
- HTTP client with pagination support
- Server with routing and handlers
- Service implementation with example endpoints
- Integration tests
- Protobuf definitions
- Build automation via Makefile

### Making Changes

After your initial setup, iterate on your API by modifying the OpenAPI spec:

```bash
# 1. Add a new endpoint (optional - you can also edit openapi.yaml directly)
duh add /v1/products.create CreateProduct

# 2. Validate your changes follow DUH-RPC conventions
duh lint openapi.yaml

# 3. Regenerate code to incorporate changes
duh generate

# 4. Regenerate protobuf Go code
buf generate

# 5. Run tests
make test
```

Your service stays in sync with your OpenAPI specification, ensuring consistency between your API contract and implementation.

## Command Reference

### `duh init` - Initialize a New Specification

Creates a new DUH-RPC compliant OpenAPI specification template with example endpoints.

**Basic usage:**
```bash
# Creates openapi.yaml in current directory
duh init

# Create with custom filename
duh init my-api.yaml

# Create in a specific directory
duh init api/openapi.yaml
```

**What's included:**
The generated template includes a complete working example with four endpoints demonstrating all DUH-RPC requirements:
- `users.create` - Creating resources
- `users.get` - Retrieving a single resource
- `users.list` - List operations with pagination
- `users.update` - Updating resources

The `openapi.yaml` is ready to use immediately or can be modified.

### `duh lint` - Validate DUH-RPC Compliance

Lints OpenAPI file against all 8 DUH-RPC requirements, providing clear error messages and actionable suggestions for violations.

**Basic usage:**
```bash
# Validate openapi.yaml (default)
duh lint

# Validate specific file
duh lint api/openapi.yaml

# Validate multiple files
duh lint api-v1.yaml api-v2.yaml
```
**Example output for compliant spec:**
```
✓ api-spec.yaml is DUH-RPC compliant
```

**Example output with violations:**
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

See the [Validation Rules](#validation-rules) section for details on all requirements.

### `duh add` - Add New Endpoints

Adds a new DUH-RPC compliant endpoint to an existing OpenAPI specification with placeholder schemas.

**Basic usage:**
```bash
# Add endpoint to openapi.yaml
duh add /v1/products.create CreateProduct

# Add to custom spec file
duh add /v1/orders.cancel CancelOrder -f api/openapi.yaml

# Add to file in different directory
duh add /v1/payments.process ProcessPayment -f api/v2/openapi.yaml
```

**What's created:**
- New POST operation at the specified path
- Request schema with placeholder fields
- Response schema (200 status) with placeholder structure
- Error response schemas for common error codes
- Proper operationId for code generation

**Common endpoint patterns:**
```bash
# Create operations
duh add /v1/users.create CreateUser

# Get operations (single resource)
duh add /v1/users.get GetUser

# List operations (with pagination)
duh add /v1/users.list ListUsers

# Update operations
duh add /v1/users.update UpdateUser

# Delete operations
duh add /v1/users.delete DeleteUser

# Custom actions
duh add /v1/orders.cancel CancelOrder
duh add /v1/payments.refund RefundPayment
```

After adding an endpoint, edit the generated schemas to match your needs, then run `duh lint` to verify compliance.

### `duh generate` - Generate Code

Generates production-ready Go code from OpenAPI specifications, including HTTP clients, servers, protobuf definitions, and optional full service scaffolding.

**Basic usage:**
```bash
# Generate from openapi.yaml (default)
duh generate

# Specify custom spec file
duh generate api/openapi.yaml

# Generate with full service scaffolding
duh generate --full
```

**Common options:**
```bash
# Custom output directory
duh generate --output-dir pkg/api

# Custom package name
duh generate -p myapi

# Custom protobuf path and package
duh generate --proto-path proto/v1/api.proto --proto-package myapi.v1

# Combine multiple options
duh generate --full --output-dir internal/api -p api
```

**Basic generation (default):**
Creates core API components:
- `client.go` - HTTP client with typed methods for each endpoint
- `server.go` - HTTP server with routing and handler registration
- `iterator.go` - Pagination iterators for list operations (if applicable)
- `proto/v1/api.proto` - Protobuf message definitions
- `buf.yaml` - Buf configuration for protobuf compilation
- `buf.gen.yaml` - Buf code generation configuration

**Full scaffolding (--full flag):**
Generates a complete service with everything from basic generation plus:
- `daemon.go` - Service orchestration with TLS/HTTP support and graceful shutdown
- `service.go` - Service implementation (complete example or stub interface)
- `api_test.go` - Integration test suite or minimal test example
- `Makefile` - Build automation with targets for test, lint, build, and proto generation

**Generated client features:**
- Type-safe method calls for all endpoints
- Automatic pagination for list operations
- Context support for timeouts and cancellation
- Configurable base URL and HTTP client
- Built-in error handling

**Generated server features:**
- Automatic routing based on OpenAPI paths
- Request validation
- Response serialization
- Error response formatting
- Middleware support

**Customization options:**

| Flag | Description | Default |
|------|-------------|---------|
| `--output-dir` | Directory for generated files | Current directory |
| `-p, --package` | Go package name | Inferred from module |
| `--proto-path` | Path for protobuf file | `proto/v1/api.proto` |
| `--proto-package` | Protobuf package name | `api.v1` |
| `--full` | Generate complete service scaffold | `false` |

## Lint Rules

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

