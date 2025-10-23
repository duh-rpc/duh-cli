# DUH-RPC OpenAPI Specification Rules

**Version:** 1.0
**Last Updated:** 2025-10-22

## Introduction

This document provides a reference guide for writing OpenAPI 3.0 specifications that comply with DUH-RPC conventions. DUH-RPC is a simple, opinionated RPC-over-HTTP specification that uses POST-only endpoints with a structured path format.

**Target Audience:** API designers and developers writing OpenAPI specifications for DUH-RPC services.

**How to Use This Guide:**
- Review the Quick Reference table for an overview of all rules
- Refer to individual rule sections for detailed requirements and examples
- Use the Complete Example as a template for your specifications
- Validate your spec with `duhrpc-lint` to ensure compliance

---

## Quick Reference

| Rule | Requirement | Example |
|------|-------------|---------|
| **Path Format** | `/v{version}/{subject}.{method}` | `/v1/users.create` |
| **HTTP Method** | POST only | `post:` |
| **Query Parameters** | Not allowed | Use request body instead |
| **Request Body** | Required (`required: true`) | All operations must have request body |
| **Content Types** | JSON (required), protobuf/octet-stream (optional) | `application/json` |
| **Error Schema** | Must have `code` (integer) and `message` (string) | See Error Response rule |
| **Status Codes** | 200, 400, 401, 403, 404, 429, 452-455, 500 only | Success: 200, Errors: 4xx/5xx |
| **Success Response** | 200 with content required | Must define 200 response with schema |

---

## Rule 1: Path Format

**Requirement:** All paths must follow the format `/v{version}/{subject}.{method}`

### Path Components

**Version:**
- Format: `/v{N}` where N is a non-negative integer (major version only)
- Valid: `v0`, `v1`, `v2`, `v10`, `v100`
- Must start with lowercase 'v'
- `v0` is allowed for preview/beta APIs

**Subject:**
- Lowercase letters, digits, hyphens, underscores only
- Must start with a letter
- Length: 1-50 characters
- Multi-word subjects allowed: `user-accounts`, `message_queue`

**Method:**
- Same character rules as subject
- Lowercase letters, digits, hyphens, underscores
- Must start with a letter
- Multi-word methods allowed: `get-by-id`, `send_notification`

**Separator:**
- Exactly one dot (`.`) between subject and method

**Path Parameters:**
- Not allowed in DUH-RPC paths
- All paths must be static
- Dynamic data should be passed in request body

### Valid Examples

```yaml
paths:
  /v1/users.create:
    post: ...

  /v1/users.get-by-id:
    post: ...

  /v2/message-queue.send:
    post: ...

  /v0/beta-features.test:
    post: ...
```

### Path Regex

```regex
^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$
```

---

## Rule 2: POST-Only HTTP Methods

**Requirement:** All operations must use the POST method exclusively.

DUH-RPC is RPC-over-HTTP, not REST. All operations use POST, regardless of whether they are reads or writes. Data is passed in the request body, not in the URL or query parameters.

### Valid Example

```yaml
paths:
  /v1/users.create:
    post:
      operationId: createUser
      summary: Create a new user
      requestBody: ...
      responses: ...

  /v1/users.get-by-id:
    post:
      operationId: getUserById
      summary: Get user by ID
      requestBody: ...
      responses: ...
```

---

## Rule 3: No Query Parameters

**Requirement:** Query parameters are not allowed in DUH-RPC operations.

All input data must be passed in the request body. Header and cookie parameters are allowed (not restricted by DUH-RPC).

### Valid Example

```yaml
paths:
  /v1/users.search:
    post:
      parameters:
        - name: X-Request-ID
          in: header
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                searchTerm:
                  type: string
                limit:
                  type: integer
```

---

## Rule 4: Required Request Bodies

**Requirement:** All operations must have a request body with `required: true`.

Every DUH-RPC operation must define a `requestBody` object with `required: true`. This ensures consistent semantics where all input data is in the request body.

### Valid Example

```yaml
paths:
  /v1/users.create:
    post:
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
      responses: ...
```

---

## Rule 5: Content Type Restrictions

**Requirement:** Only specific content types are allowed.

**Allowed Content Types:**
- `application/json` (REQUIRED - must be present)
- `application/protobuf` (OPTIONAL)
- `application/octet-stream` (OPTIONAL)

**Rules:**
- At minimum, `application/json` must be present in both request and response
- Operations may list multiple content types
- Request and response content types can differ
- No MIME type parameters allowed (e.g., no `; charset=utf-8`)

### Valid Examples

**JSON Only:**
```yaml
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
```

**Multiple Content Types:**
```yaml
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

responses:
  200:
    description: Success
    content:
      application/json:
        schema:
          type: object
      application/protobuf:
        schema:
          type: string
          format: binary
```

---

## Rule 6: Error Response Schema

**Requirement:** Error responses (4xx, 5xx) must have a specific schema structure.

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

**Field Requirements:**
- **code:** Must be integer type, may optionally have `enum` constraint matching HTTP status
- **message:** Must be string type
- **details:** Optional object field for additional error information

**Applies to Status Codes:**
- 4xx: 400, 401, 403, 404, 429, 452, 453, 454, 455
- 5xx: 500

### Valid Examples

**Inline Schema:**
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
              enum: [400]
            message:
              type: string
            details:
              type: object
```

**Using $ref:**
```yaml
responses:
  404:
    description: Not Found
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

**Reusable Response:**
```yaml
responses:
  500:
    $ref: '#/components/responses/ServerError'

components:
  responses:
    ServerError:
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
```

---

## Rule 7: Allowed Status Codes

**Requirement:** Only specific HTTP status codes are allowed in DUH-RPC.

**Allowed Status Codes:**
- **2xx:** 200 (only)
- **4xx Client Errors:** 400, 401, 403, 404, 429
- **4xx Custom Service Errors:** 452, 453, 454, 455
- **5xx Server Errors:** 500 (only)

**Usage Guidelines:**
- **200:** All successful responses
- **400:** Invalid request (validation errors, malformed input)
- **401:** Authentication required or failed
- **403:** Forbidden (authenticated but not authorized)
- **404:** Resource not found
- **429:** Rate limit exceeded
- **452-455:** Custom service-specific errors (no restrictions on meaning)
- **500:** Internal server error

### Valid Example

```yaml
responses:
  200:
    description: Success
    content:
      application/json:
        schema:
          type: object
  400:
    description: Bad Request
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
  401:
    description: Unauthorized
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
  404:
    description: Not Found
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
  452:
    description: Custom Service Error
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
  500:
    description: Internal Server Error
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
```

---

## Rule 8: Required Success Response

**Requirement:** All operations must define a 200 response with content.

Every operation must have a 200 response defined with a `content` object containing at least one media type with a schema. The schema can be any valid JSON schema (object, array, primitive, etc.).

### Valid Examples

**Object Response:**
```yaml
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
```

**Array Response:**
```yaml
responses:
  200:
    description: List of users
    content:
      application/json:
        schema:
          type: array
          items:
            type: object
            properties:
              userId:
                type: string
              name:
                type: string
```

**Primitive Response:**
```yaml
responses:
  200:
    description: Operation result
    content:
      application/json:
        schema:
          type: boolean
```

---

## Complete Minimal Example

Here is a minimal DUH-RPC compliant OpenAPI specification that passes all rules:

```yaml
openapi: 3.0.0
info:
  title: Minimal DUH-RPC API
  version: 1.0.0

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
                  message:
                    type: string
                  details:
                    type: object
```

---

## Validation Tool

Use `duhrpc-lint` to automatically validate your OpenAPI specifications against all DUH-RPC rules:

```bash
# Validate a spec
duhrpc-lint openapi.yaml

# If compliant
✓ openapi.yaml is DUH-RPC compliant

# If violations found
Validating openapi.yaml...

ERRORS FOUND:

[path-format] /api/users
  Path must follow format: /v{version}/{subject}.{method}
  Found: /api/users
  Suggestion: Change to /v1/users.create

Summary: 1 violation found in openapi.yaml
```

**Exit Codes:**
- `0` - Validation passed (compliant)
- `1` - Validation failed (violations found)
- `2` - Tool error (file not found, parse error)

---

## Additional Resources

- **DUH-RPC Specification:** [github.com/duh-rpc/duh-go](https://github.com/duh-rpc/duh-go)
- **OpenAPI 3.0 Specification:** [spec.openapis.org/oas/v3.0.3](https://spec.openapis.org/oas/v3.0.3.html)
