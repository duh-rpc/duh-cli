# DUH Linter Rules

This document is the authoritative reference for all rules enforced by the `duh lint` command. It is
intended as a reference for developers writing OpenAPI specs that target the DUH-RPC standard, and as
the source of truth for linter implementors.

---

## Overview

### Severity Levels

Every rule carries one of two severity levels:

| Severity | Effect |
|---|---|
| `ERROR` | Violation causes `duh lint` to exit with code 1 |
| `WARNING` | Violation is reported but exit code remains 0 |

### Rule Names

All rules are identified by `SCREAMING_SNAKE_CASE` names (e.g. `PATH_FORMAT`, `NO_NULLABLE`).

### Configuration

Rules can be disabled project-wide via a `.duh.yaml` file in the project root:

```yaml
lint:
  disable:
    - DESCRIPTION_REQUIRED
    - PROPERTY_CAMELCASE
```

Rules can also be disabled for a single run using the `--disable` flag with a comma-separated list:

```bash
duh lint --disable TIMESTAMP_FORMAT,DATE_FORMAT
```

Both mechanisms only suppress rules — there is no way to re-enable a disabled rule for a specific
location. Disabled rules are skipped entirely and produce no output.

### Opting Out of a Rule

Operations and schemas can suppress specific rules using the `x-duh-lint-ignore` extension field. This
is the official escape hatch for intentional violations — for example, a third-party schema you do not
control, or an endpoint that deliberately deviates from a convention.

```yaml
# Operation level — suppress a rule for one endpoint
paths:
  /items.export:
    post:
      x-duh-lint-ignore: [PAGINATION_PARAMETERS]

# Schema level — suppress a rule for one schema and all its properties
components:
  schemas:
    ThirdPartyError:
      x-duh-lint-ignore: [NO_NULLABLE]
```

`x-duh-lint-ignore` accepts a list of rule names. It is supported at the **operation** and **schema**
levels only — not at the individual property level. All rules must check for this field before
reporting a violation.

### Field Name Normalization

Rules that match field names (pagination parameters, schema structure fields, etc.) MUST normalize
names before comparison. Normalization: strip all `-` and `_` separators, then lowercase. This means
`endCursor`, `end_cursor`, and `end-cursor` are all treated as equivalent.

A single `normalize(name string) string` utility function MUST be used consistently across all rules
that perform field name matching.

---

## Document Rules

### `OPENAPI_VERSION` — ERROR

The `openapi` field MUST specify a `3.x` version (e.g. `3.0.3`, `3.1.0`).

```yaml
# ✅ valid
openapi: "3.0.3"

# ❌ invalid
openapi: "2.0"
```

---

### `DESCRIPTION_REQUIRED` — WARNING

Operations, parameters, and schema properties SHOULD have a `description` field. This rule encourages
documentation quality but does not block compliance.

---

## HTTP Rules

### `HTTP_METHOD_ALLOWED` — ERROR

All operations MUST use the `POST` HTTP method. No other HTTP verbs are permitted.

```yaml
# ✅ valid
paths:
  /users.create:
    post: ...

# ❌ invalid
paths:
  /users.list:
    get: ...
```

---

### `REQUEST_BODY_REQUIRED` — ERROR

Every `POST` operation MUST define a request body.

---

### `POST_NO_QUERY_PARAMS` — ERROR

`POST` operations MUST NOT define query parameters. All request data belongs in the request body.

---

### `STATUS_CODE_ALLOWED` — ERROR

Operations MUST only use the following HTTP status codes in their responses:

| Code | Name | Notes |
|---|---|---|
| `200` | OK | Standard success |
| `201` | Created | Resource successfully created |
| `202` | Accepted | Request accepted for async processing |
| `400` | Bad Request | Missing or invalid parameter |
| `401` | Unauthorized | Not authenticated |
| `403` | Forbidden | Authenticated but not authorized |
| `404` | Not Found | Resource not found |
| `409` | Conflict | Request conflicts with existing state |
| `429` | Too Many Requests | Rate limited |
| `452` | Client Error | Client error before request reached server |
| `453` | Request Failed | Valid request, but operation failed |
| `454` | Retry Request | Valid request, server asks client to retry |
| `455` | Client Content Error | Content does not follow the DUH spec |
| `500` | Internal Error | Unexpected server error |
| `501` | Not Implemented | Method not implemented on this server |

Any other status code is a violation.

---

### `SUCCESS_RESPONSE` — ERROR

Every operation MUST define a `200` response.

---

### `NO_PLAIN_TEXT_RESPONSE` — ERROR

Response content types MUST NOT include `text/plain`. A `text/plain` response indicates the reply
came from infrastructure (proxy, load balancer, etc.) rather than the service itself.

---

## Path Rules

### `PATH_FORMAT` — ERROR

Paths MUST follow one of these two forms:

```
/{resource}.{method}
/{domain}/{resource}.{method}
```

Where:
- `{domain}`, `{resource}`, and `{method}` are lowercase
- `{resource}` and `{method}` are separated by a `.`
- An optional `{domain}` segment may precede the `{resource}.{method}` segment
- No version prefix is permitted in the path (see `PATH_NO_VERSION_PREFIX`)

```yaml
# ✅ valid
/users.create
/dogs.feed
/billing/invoices.create

# ❌ invalid
/v1/users.create        # version prefix not allowed in path
/users/create           # missing dot separator
/Users.Create           # not lowercase
```

---

### `PATH_NO_VERSION_PREFIX` — ERROR

Paths MUST NOT include a version prefix such as `/v1/`. Versioning belongs in `servers[].url`, not
in the path itself.

```yaml
# ✅ valid
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
    post: ...

# ❌ invalid
paths:
  /v1/users.create:
    post: ...
```

---

### `SERVER_URL_VERSIONING` — ERROR

Every entry in `servers[].url` MUST end with a version segment in the form `/v{N}`, where `v` is
lowercase and `{N}` is a positive integer (e.g. `/v1`, `/v2`).

```yaml
# ✅ valid
servers:
  - url: https://api.example.com/v1

# ❌ invalid
servers:
  - url: https://api.example.com          # no version
  - url: https://api.example.com/V1       # uppercase V
  - url: https://api.example.com/v1.2.1   # not vN format
```

---

### `PATH_HYPHEN_SEPARATOR` — ERROR

Multi-word path segments MUST use hyphens as word separators. Underscores and camelCase are not
permitted in path segments.

```yaml
# ✅ valid
/user-profiles.get

# ❌ invalid
/user_profiles.get
/userProfiles.get
```

---

### `PATH_PLURAL_RESOURCES` — WARNING

Collection resource names in paths SHOULD use plural nouns.

```yaml
# ✅ recommended
/users.list
/invoices.create

# ⚠️ warned
/user.list
/invoice.create
```

---

### `PATH_MULTIPLE_PARAMETERS` — ERROR

A path MUST NOT define more than one path parameter (e.g. `{id}`).

```yaml
# ✅ valid
/users/{id}.get

# ❌ invalid
/users/{userId}/orders/{orderId}.get
```

---

## Content Type Rules

### `CONTENT_TYPE` — ERROR

This rule enforces three related constraints on operation content types:

**1. JSON must be supported**

Every operation MUST support `application/json` as both a request and response content type. JSON
is the baseline content type and MUST always be available regardless of what other types are
supported.

**2. Protobuf is permitted**

`application/protobuf` is an allowed content type for both requests and responses.

**3. Multipart and form-encoded types are prohibited**

`multipart/form-data` and `application/x-www-form-urlencoded` MUST NOT be used.

**Note on `application/octet-stream`:** This content type is neither required nor prohibited by the
linter. The streaming section of the DUH spec is not yet complete. Tooling will not flag or validate
`octet-stream` until that section is finalized.

```yaml
# ✅ valid
requestBody:
  content:
    application/json:
      schema: ...
    application/protobuf:
      schema: ...

# ❌ invalid — missing application/json
requestBody:
  content:
    application/protobuf:
      schema: ...

# ❌ invalid — multipart not allowed
requestBody:
  content:
    multipart/form-data:
      schema: ...
```

---

## Schema Rules

### `ERROR_SCHEMA` — ERROR

All error responses MUST reference a schema that conforms to the DUH Reply structure. The schema
must satisfy:

- `message` — **required**, type `string`. A human-readable description of the error.
- `code` — **optional**, type `string`. May be a numeric string matching the HTTP status code
  (e.g. `"400"`) or a semantic string (e.g. `"CARD_DECLINED"`). Not an integer. Not an enum.
- `details` — **optional**. When present, MUST be defined as `additionalProperties: { type: string }`.
  Represents a flat map of string key/value pairs.

```yaml
# ✅ valid ErrorSchema
components:
  schemas:
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string
        code:
          type: string
        details:
          type: object
          additionalProperties:
            type: string
```

---

### `SCHEMA_NO_INLINE_OBJECTS` — ERROR

All object schemas MUST be defined in `components/schemas` and referenced via `$ref`. Inline object
definitions within operation request bodies, responses, or other schema properties are not permitted.

```yaml
# ✅ valid
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        address:
          $ref: '#/components/schemas/Address'

# ❌ invalid — inline object
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        address:
          type: object
          properties:
            street:
              type: string
```

---

### `SCHEMA_ADDITIONAL_PROPERTIES_RESPONSE` — ERROR

Response schemas MUST NOT use `additionalProperties: false`. This constraint prevents forward
compatibility as new fields may be added to responses over time.

```yaml
# ✅ valid
CreateUserResponse:
  type: object
  properties:
    id:
      type: string

# ❌ invalid
CreateUserResponse:
  type: object
  additionalProperties: false
  properties:
    id:
      type: string
```

---

### `SCHEMA_EXAMPLE_VALIDATION` — ERROR

When a schema includes an `example` field, the example value MUST validate against the schema's own
type and constraints.

---

### `PROHIBITED_READONLY_WRITEONLY` — ERROR

`readOnly` and `writeOnly` MUST NOT be used on schema properties. These annotations imply a single
schema is shared between request and response contexts, which violates the dedicated schema rule.
If a field only appears in a response, it belongs only in the response schema. If a field only
appears in a request, it belongs only in the request schema.

```yaml
# ❌ invalid
properties:
  createdAt:
    type: string
    readOnly: true
```

---

### `NO_NULLABLE` — ERROR

Nullable fields are not permitted in any form. Protobuf has no native concept of a null value, and
this blanket rule avoids the need for complex per-field nullable analysis. Use optional fields
(absence of value) to represent the lack of a value rather than an explicit null.

Both of the following forms are violations:

```yaml
# ❌ invalid — nullable: true
properties:
  middleName:
    type: string
    nullable: true

# ❌ invalid — array type syntax
properties:
  middleName:
    type: ["string", "null"]
```

---

### `NULLABLE_OPTIONAL_RESPONSE` — ERROR

When `NO_NULLABLE` is disabled, this rule provides a fallback guardrail. Response properties that are
marked `nullable: true` MUST also be listed in the `required` array. A property that is both nullable
and optional creates ambiguity — the client cannot distinguish "field absent" from "field is null."

This rule only inspects success (2xx) response schemas.

```yaml
# ❌ invalid — nullable but not required
CreateUserResponse:
  type: object
  properties:
    middleName:
      type: string
      nullable: true

# ✅ valid — nullable and required
CreateUserResponse:
  type: object
  required: [middleName]
  properties:
    middleName:
      type: string
      nullable: true
```

---

## Protobuf Compatibility Rules

DUH-RPC natively supports `application/protobuf`. The following rules ensure schemas can be
represented as protobuf messages without loss of fidelity.

---

### `INTEGER_FORMAT_REQUIRED` — ERROR

All `integer` and `number` fields MUST specify an explicit `format`. Protobuf requires knowing the
exact numeric type at compile time.

Allowed formats:
- `int32`
- `int64`
- `float`
- `double`

```yaml
# ✅ valid
count:
  type: integer
  format: int32

# ❌ invalid
count:
  type: integer   # no format
```

---

### `NO_NESTED_ARRAYS` — ERROR

Schemas MUST NOT define arrays whose `items` are themselves arrays. Protobuf does not support
`repeated repeated` fields. Wrap the inner array in a message type instead.

```yaml
# ❌ invalid
matrix:
  type: array
  items:
    type: array
    items:
      type: integer
      format: int32
```

---

### `TYPED_ADDITIONAL_PROPERTIES` — ERROR

When `additionalProperties` is used, the value type MUST be explicitly specified. Untyped
`additionalProperties` cannot be mapped to a protobuf map field.

```yaml
# ✅ valid
metadata:
  type: object
  additionalProperties:
    type: string

# ❌ invalid
metadata:
  type: object
  additionalProperties: true
```

---

### `PROHIBITED_ALLOF` — ERROR

`allOf` has no equivalent in protobuf and MUST NOT be used.

```yaml
# ❌ invalid
CreateUserRequest:
  allOf:
    - $ref: '#/components/schemas/BaseRequest'
    - type: object
      properties:
        name:
          type: string
```

---

### `PROHIBITED_ANYOF` — ERROR

`anyOf` introduces ambiguous typing that cannot be represented in protobuf and MUST NOT be used.

---

### `PROHIBITED_ONEOF` — ERROR

`oneOf` MUST NOT be used. While OpenAPI supports discriminated unions via `oneOf` with a
`discriminator`, the JSON wire format for a discriminated union is structurally incompatible with
protobuf's `oneof` JSON serialization:

- **OpenAPI `oneOf`** flattens variant fields and uses a discriminator property:
  `{"eventType": "cat", "name": "Whiskers"}`
- **Protobuf `oneof`** nests the variant under its field name:
  `{"cat": {"name": "Whiskers"}}`

Since DUH-RPC supports both `application/json` and `application/protobuf`, a single schema cannot
accurately describe both wire formats. Use flat objects with optional properties instead.

```yaml
# ❌ invalid — even with discriminator
oneOf:
  - $ref: '#/components/schemas/CatEvent'
  - $ref: '#/components/schemas/DogEvent'
discriminator:
  propertyName: eventType

# ✅ valid — flat optional properties
Event:
  type: object
  properties:
    eventType:
      type: string
    cat:
      $ref: '#/components/schemas/CatEventData'
    dog:
      $ref: '#/components/schemas/DogEventData'
```

**Note:** If this rule is disabled via configuration, the `DISCRIMINATOR_REQUIRED`,
`DISCRIMINATOR_MAPPING`, and `DISCRIMINATOR_VARIANT_FIELD` rules serve as fallback guardrails
ensuring that any `oneOf` usage is at least properly discriminated.

---

### `DISCRIMINATOR_REQUIRED` — ERROR

`oneOf` is only permitted as a discriminated union. Every `oneOf` MUST include a `discriminator`
object. Bare `oneOf` without a discriminator is prohibited.

This rule is a fallback for when `PROHIBITED_ONEOF` is disabled.

```yaml
# ✅ valid (when PROHIBITED_ONEOF is disabled)
oneOf:
  - $ref: '#/components/schemas/CatEvent'
  - $ref: '#/components/schemas/DogEvent'
discriminator:
  propertyName: eventType
  mapping:
    cat: '#/components/schemas/CatEvent'
    dog: '#/components/schemas/DogEvent'

# ❌ invalid
oneOf:
  - $ref: '#/components/schemas/CatEvent'
  - $ref: '#/components/schemas/DogEvent'
```

---

### `DISCRIMINATOR_MAPPING` — ERROR

The `discriminator` object on a `oneOf` MUST include an explicit `mapping`. A discriminator without
a mapping is not permitted.

This rule is a fallback for when `PROHIBITED_ONEOF` is disabled.

---

### `DISCRIMINATOR_VARIANT_FIELD` — ERROR

Each variant schema referenced in a `oneOf` MUST include the discriminator property as a field in
its own schema definition.

This rule is a fallback for when `PROHIBITED_ONEOF` is disabled.

---

## Naming Rules

### `PROPERTY_CAMELCASE` — ERROR

Schema property names MUST use camelCase. Property names using snake_case, kebab-case, or other
formats are violations.

```yaml
# ✅ valid
properties:
  firstName:
    type: string
  createdAt:
    type: string

# ❌ invalid
properties:
  first_name:
    type: string
  created-at:
    type: string
```

---

### `DISCRIMINATOR_PROPERTY_NAME` — ERROR

When a `oneOf` schema includes a `discriminator`, the discriminator `propertyName` MUST be `type`.

```yaml
# ✅ valid
discriminator:
  propertyName: type

# ❌ invalid
discriminator:
  propertyName: eventType
```

---

### `REQUEST_STANDARD_NAME` — ERROR

Request schemas MUST follow one of these naming patterns (separator style is normalized — any
consistent use of camelCase, snake_case, or kebab-case is accepted):

- `{Method}Request` — e.g. `CreateRequest`
- `{Service}{Method}Request` — e.g. `UsersCreateRequest`
- `{Domain}{Service}{Method}Request` — e.g. `BillingInvoicesCreateRequest`

---

### `RESPONSE_STANDARD_NAME` — ERROR

Response schemas MUST follow one of these naming patterns (separator style is normalized):

- `{Method}Response` — e.g. `CreateResponse`
- `{Service}{Method}Response` — e.g. `UsersCreateResponse`
- `{Domain}{Service}{Method}Response` — e.g. `BillingInvoicesCreateResponse`

---

### `REQUEST_RESPONSE_UNIQUE` — ERROR

Each operation MUST use its own unique request and response schemas. Schemas MUST NOT be shared
across operations.

```yaml
# ❌ invalid — SharedRequest used by two operations
paths:
  /users.create:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SharedRequest'
  /users.update:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SharedRequest'
```

---

## Pagination Rules

DUH-RPC uses cursor-based forward pagination for collection endpoints. Offset and limit style
pagination is not supported.

### Detecting Paginated Endpoints

An endpoint is considered paginated if its response schema contains both an `items` array field and
a `pagination` object.

---

### `PAGINATION_PARAMETERS` — ERROR

Paginated requests MUST include pagination parameters nested under a `pagination` sub-object in the
request body:

| Field | Type | Required | Constraints |
|---|---|---|---|
| `pagination.first` | integer | Yes | Minimum: 1, Maximum: 100 |
| `pagination.after` | string | No | Cursor from previous response |

Field names are matched using normalization (strip separators, lowercase).

```yaml
# ✅ valid request body
{
  "pagination": {
    "first": 25,
    "after": "cursor_abc123"
  }
}

# first page — after is optional
{
  "pagination": {
    "first": 10
  }
}
```

---

### `PAGINATION_NO_LIMIT_OFFSET` — ERROR

The following parameter names are prohibited as they imply offset-style pagination:

- `limit`
- `offset`
- `page` as a standalone top-level integer parameter

The `pagination` sub-object described above is the only permitted pagination construct.

---

### `PAGINATED_REQUEST_STRUCTURE` — ERROR

Pagination parameters `first` and `after` MUST be nested under the `pagination` sub-object. These
parameters MUST NOT appear as top-level properties in the request body.

```yaml
# ✅ valid — nested under pagination
requestBody:
  content:
    application/json:
      schema:
        type: object
        properties:
          pagination:
            type: object
            properties:
              first:
                type: integer
              after:
                type: string

# ❌ invalid — first/after at root level
requestBody:
  content:
    application/json:
      schema:
        type: object
        properties:
          first:
            type: integer
          after:
            type: string
```

---

### `RESPONSE_PAGINATED_STRUCTURE` — ERROR

Paginated responses MUST include the following structure:

| Field | Type | Required | Description |
|---|---|---|---|
| `items` | array | Yes | The page of results |
| `pagination.endCursor` | string | Yes | Cursor to pass as `pagination.after` on the next request |
| `pagination.hasNextPage` | boolean | Yes | Whether additional results exist beyond this page |

Field names are matched using normalization (strip separators, lowercase).

```yaml
# ✅ valid paginated response
{
  "items": [...],
  "pagination": {
    "endCursor": "cursor_xyz789",
    "hasNextPage": true
  }
}
```

When `hasNextPage` is `false`, the client SHOULD NOT make a further request. When `hasNextPage` is
`true` and `endCursor` is present, the client MAY request the next page by passing `endCursor` as
`pagination.after`.

---

## Prohibited Feature Rules

These rules prohibit OpenAPI features that are irrelevant or harmful in DUH-RPC specifications.

---

### `PROHIBITED_XML` — ERROR

The `xml` property MUST NOT be used on schemas or schema properties. DUH-RPC uses JSON and protobuf
only.

---

### `PROHIBITED_COOKIES` — ERROR

Cookie parameters and cookie-based security schemes MUST NOT be used. All request data belongs in
the request body or authorization headers.

```yaml
# ❌ invalid — cookie parameter
parameters:
  - name: session
    in: cookie

# ❌ invalid — cookie security scheme
components:
  securitySchemes:
    cookieAuth:
      type: apiKey
      in: cookie
```

---

### `PROHIBITED_HATEOAS` — ERROR

Response `links` (HATEOAS) MUST NOT be used. DUH-RPC uses explicit API endpoints rather than
hypermedia-driven navigation.

---

### `PROHIBITED_PARAMETER_STYLES` — ERROR

Parameters MUST NOT use `style` or `explode` properties. Use default serialization only.

---

### `PROHIBITED_MULTIPLE_EXAMPLES` — WARNING

Parameters and media types SHOULD use the singular `example` field instead of the plural `examples`
field. A single canonical example is preferred for clarity.

---

## Format Convention Rules

These rules enforce consistent representation of common data types.

---

### `TIMESTAMP_FORMAT` — ERROR

Properties whose names end in `At` or `Timestamp` (e.g. `createdAt`, `lastModifiedTimestamp`) MUST
be defined as `type: string` with `format: date-time`.

```yaml
# ✅ valid
createdAt:
  type: string
  format: date-time

# ❌ invalid
createdAt:
  type: integer
  format: int64
```

---

### `DATE_FORMAT` — ERROR

Properties whose names end in `Date` (e.g. `birthDate`, `expirationDate`) MUST be defined as
`type: string` with `format: date`.

```yaml
# ✅ valid
birthDate:
  type: string
  format: date

# ❌ invalid
birthDate:
  type: string
```

---

### `AMOUNT_DECIMAL_STRING` — ERROR

Properties named `amount` MUST be `type: string`. Representing monetary amounts as strings avoids
floating-point precision issues that arise with `number` or `integer` types.

```yaml
# ✅ valid
amount:
  type: string

# ❌ invalid
amount:
  type: number
  format: double
```

---

### `AMOUNT_SCHEMA_PATTERN` — WARNING

Schemas that contain an `amount` property SHOULD also include an `assetType` property to clarify
what currency or asset the amount represents.

---

## Idempotency Rules

### `IDEMPOTENCY_KEY_DEFINITION` — ERROR

When a schema includes an `idempotencyKey` property, it MUST be defined as `type: string` with
`maxLength: 128`.

```yaml
# ✅ valid
idempotencyKey:
  type: string
  maxLength: 128

# ❌ invalid
idempotencyKey:
  type: string
```

---

## Rule Reference

| Rule | Severity | Category |
|---|---|---|
| `OPENAPI_VERSION` | ERROR | Document |
| `DESCRIPTION_REQUIRED` | WARNING | Document |
| `HTTP_METHOD_ALLOWED` | ERROR | HTTP |
| `REQUEST_BODY_REQUIRED` | ERROR | HTTP |
| `POST_NO_QUERY_PARAMS` | ERROR | HTTP |
| `STATUS_CODE_ALLOWED` | ERROR | HTTP |
| `SUCCESS_RESPONSE` | ERROR | HTTP |
| `NO_PLAIN_TEXT_RESPONSE` | ERROR | HTTP |
| `PATH_FORMAT` | ERROR | Path |
| `PATH_NO_VERSION_PREFIX` | ERROR | Path |
| `SERVER_URL_VERSIONING` | ERROR | Path |
| `PATH_HYPHEN_SEPARATOR` | ERROR | Path |
| `PATH_PLURAL_RESOURCES` | WARNING | Path |
| `PATH_MULTIPLE_PARAMETERS` | ERROR | Path |
| `CONTENT_TYPE` | ERROR | Content Type |
| `ERROR_SCHEMA` | ERROR | Schema |
| `SCHEMA_NO_INLINE_OBJECTS` | ERROR | Schema |
| `SCHEMA_ADDITIONAL_PROPERTIES_RESPONSE` | ERROR | Schema |
| `SCHEMA_EXAMPLE_VALIDATION` | ERROR | Schema |
| `PROHIBITED_READONLY_WRITEONLY` | ERROR | Schema |
| `NO_NULLABLE` | ERROR | Schema |
| `NULLABLE_OPTIONAL_RESPONSE` | ERROR | Schema |
| `INTEGER_FORMAT_REQUIRED` | ERROR | Protobuf |
| `NO_NESTED_ARRAYS` | ERROR | Protobuf |
| `TYPED_ADDITIONAL_PROPERTIES` | ERROR | Protobuf |
| `PROHIBITED_ALLOF` | ERROR | Protobuf |
| `PROHIBITED_ANYOF` | ERROR | Protobuf |
| `PROHIBITED_ONEOF` | ERROR | Protobuf |
| `DISCRIMINATOR_REQUIRED` | ERROR | Protobuf |
| `DISCRIMINATOR_MAPPING` | ERROR | Protobuf |
| `DISCRIMINATOR_VARIANT_FIELD` | ERROR | Protobuf |
| `PROPERTY_CAMELCASE` | ERROR | Naming |
| `DISCRIMINATOR_PROPERTY_NAME` | ERROR | Naming |
| `REQUEST_STANDARD_NAME` | ERROR | Naming |
| `RESPONSE_STANDARD_NAME` | ERROR | Naming |
| `REQUEST_RESPONSE_UNIQUE` | ERROR | Naming |
| `PAGINATION_PARAMETERS` | ERROR | Pagination |
| `PAGINATION_NO_LIMIT_OFFSET` | ERROR | Pagination |
| `PAGINATED_REQUEST_STRUCTURE` | ERROR | Pagination |
| `RESPONSE_PAGINATED_STRUCTURE` | ERROR | Pagination |
| `PROHIBITED_XML` | ERROR | Prohibited Feature |
| `PROHIBITED_COOKIES` | ERROR | Prohibited Feature |
| `PROHIBITED_HATEOAS` | ERROR | Prohibited Feature |
| `PROHIBITED_PARAMETER_STYLES` | ERROR | Prohibited Feature |
| `PROHIBITED_MULTIPLE_EXAMPLES` | WARNING | Prohibited Feature |
| `TIMESTAMP_FORMAT` | ERROR | Format Convention |
| `DATE_FORMAT` | ERROR | Format Convention |
| `AMOUNT_DECIMAL_STRING` | ERROR | Format Convention |
| `AMOUNT_SCHEMA_PATTERN` | WARNING | Format Convention |
| `IDEMPOTENCY_KEY_DEFINITION` | ERROR | Idempotency |

---

## Rules Merged Into Other Rules

| Former Rule | Merged Into | Reason |
|---|---|---|
| `NULLABLE_SYNTAX` | `NO_NULLABLE` | `NO_NULLABLE` now checks both `nullable: true` and `type: ["string", "null"]` forms |
| `NULLABLE_REQUIRED_ONLY` | `NULLABLE_OPTIONAL_RESPONSE` | Identical check with weaker severity; consolidated into single rule |
| `RPC_PAGINATION_PARAMETERS` | `PAGINATION_NO_LIMIT_OFFSET` | Both prohibit offset-style parameters; consolidated |

---

## Rules Intentionally Not Implemented

| Rule | Reason |
|---|---|
| `PATH_LOWERCASE` | Covered by `PATH_FORMAT` regex; no separate named rule needed. |
| `PATH_STYLE_CONSISTENCY` | All rules apply uniformly; no separate consistency check needed. |
| `CONTENT_TYPE_JSON_REQUIRED` | Implemented as a check within `CONTENT_TYPE`. |
| `CONTENT_TYPE_NO_MULTIPART` | Implemented as a check within `CONTENT_TYPE`. |
| `ERROR_DETAILS_SCHEMA` | Implemented as part of `ERROR_SCHEMA`. |
| `PROHIBITED_FILE_UPLOAD` | Deferred. Streaming spec is incomplete. `CONTENT_TYPE` already excludes `octet-stream` enforcement. |
