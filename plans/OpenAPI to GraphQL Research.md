Perfect! This clarifies everything. Let me provide a focused recommendation for your **OpenAPI → GraphQL Schema converter**.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    Your Ecosystem                             │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  OpenAPI Spec (DUH-RPC)                                      │
│         ↓                           ↓                         │
│  openapi-proto.go          [NEW] openapi2graphql            │
│  (existing)                (what we're building)             │
│         ↓                           ↓                         │
│  Protobuf → Go Structs      GraphQL Schema (.graphql)       │
│         ↓                           ↓                         │
│  Business Methods           gqlgen (generates code)          │
│  Service.UserGet()                  ↓                         │
│                            Go types + Resolver stubs         │
│                                     ↓                         │
│                            [Manual] Wire resolvers           │
│                            to Service.UserGet()              │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

## GraphQL Naming Strategy (Idiomatic)

For DUH-RPC paths like `/v1/users.get`, `/v1/users.list`, here's the idiomatic GraphQL approach:

### **Recommended Naming Rules:**

|DUH-RPC Path|Action Type|GraphQL Field|Return Type|Rationale|
|---|---|---|---|---|
|`/v1/users.get`|get|`user`|`User`|Singular - fetches one item|
|`/v1/users.list`|list|`users`|`[User!]!`|Plural - fetches many|
|`/v1/users.find`|find|`findUser`|`User`|Keep action for clarity|
|`/v1/users.search`|search|`searchUsers`|`[User!]!`|Keep action for clarity|
|`/v1/users.query`|query|`queryUsers`|`[User!]!`|Keep action for clarity|
|`/v1/say.hello`|*|`sayHello`|`SayHelloResponse`|CamelCase both parts|

### **Algorithm:**

1. **Parse path**: `/v1/subject.action` → extract `subject` and `action`
2. **Determine cardinality** from action:
    - `get` → singular (one item)
    - `list`, `search`, `query` → plural (array)
    - `find` → depends on response schema (could be either)
3. **Generate field name**:
    
    ```
    if action == "get":    fieldName = singularize(subject)  # "users" → "user"elif action == "list":    fieldName = pluralize(subject)    # "user" → "users"  elif action in ["find", "search", "query"]:    fieldName = action + capitalize(subject)  # "findUsers", "searchUsers"else:    fieldName = camelCase(subject + action)  # "sayHello"
    ```
    

### **Example Outputs:**

```graphql
type Query {
  # From /v1/users.get
  user(input: UserGetInput!): User
  
  # From /v1/users.list
  users(input: UserListInput): [User!]!
  
  # From /v1/users.find
  findUser(input: UserFindInput!): User
  
  # From /v1/products.search
  searchProducts(input: ProductSearchInput!): [Product!]!
  
  # From /v1/say.hello
  sayHello(input: SayHelloInput!): SayHelloResponse
}
```

## Converter Tool Design

### **Tool Name:** `openapi2graphql` or `duh2graphql`

### **Core Components:**

```
pkg/
  parser/
    openapi.go          # Parse OpenAPI using libopenapi
  filter/
    operations.go       # Filter query operations
  mapper/
    types.go           # OpenAPI schema → GraphQL types
    queries.go         # Generate Query fields
  naming/
    converter.go       # subject.action → GraphQL field name
    inflector.go       # Pluralize/singularize
  writer/
    schema.go          # Write .graphql SDL files
```

### **Phase 1: Parse & Filter**

```go
type Operation struct {
    Path      string  // /v1/users.get
    Subject   string  // users
    Action    string  // get
    Request   Schema  // OpenAPI request body schema
    Response  Schema  // OpenAPI response schema (200)
    Error     Schema  // OpenAPI error schema (4xx/5xx)
}

// Parse OpenAPI and extract DUH-RPC operations
func ParseDuhOperations(spec *openapi.Document) []Operation {
    var ops []Operation
    
    for path, pathItem := range spec.Paths {
        // Only POST operations
        if pathItem.Post == nil {
            continue
        }
        
        // Parse /v1/subject.action
        subject, action := parseDuhPath(path)
        
        // Filter by query actions
        if !isQueryAction(action) {
            continue
        }
        
        ops = append(ops, Operation{
            Path:     path,
            Subject:  subject,
            Action:   action,
            Request:  extractRequestSchema(pathItem.Post),
            Response: extractResponseSchema(pathItem.Post, "200"),
            Error:    extractResponseSchema(pathItem.Post, "4XX"),
        })
    }
    
    return ops
}

func isQueryAction(action string) bool {
    queryActions := []string{"get", "list", "query", "find", "search"}
    for _, qa := range queryActions {
        if action == qa {
            return true
        }
    }
    return false
}
```

### **Phase 2: Type Mapping**

```go
type TypeMapper struct {
    types map[string]*GraphQLType  // De-duplication
}

func (tm *TypeMapper) MapSchema(schema Schema) *GraphQLType {
    // OpenAPI → GraphQL type mappings
    // string → String
    // integer → Int
    // number → Float
    // boolean → Boolean
    // array → [T]
    // object → custom type
    
    // Handle required fields → non-null (!)
    // Handle nested objects
    // Handle $ref resolution
}

// Example output
type User {
  id: ID!
  name: String!
  email: String
}

input UserGetInput {
  id: ID!
}
```

### **Phase 3: Query Generation**

```go
func GenerateQueryField(op Operation, typeMapper *TypeMapper) QueryField {
    // 1. Generate field name
    fieldName := convertToGraphQLName(op.Subject, op.Action)
    
    // 2. Map input type
    inputType := typeMapper.MapSchema(op.Request)
    
    // 3. Map output type
    outputType := typeMapper.MapSchema(op.Response)
    
    // 4. Determine cardinality
    isList := isListAction(op.Action) || isArrayResponse(op.Response)
    
    return QueryField{
        Name:       fieldName,
        InputType:  inputType,
        OutputType: outputType,
        IsList:     isList,
    }
}

func convertToGraphQLName(subject, action string) string {
    switch action {
    case "get":
        return singularize(subject)  // "users" → "user"
    case "list":
        return pluralize(subject)    // "user" → "users"
    case "find", "search", "query":
        return action + capitalize(subject)  // "findUsers"
    default:
        return camelCase(subject + "." + action)  // "sayHello"
    }
}
```

### **Phase 4: Schema Writing**

```go
func WriteGraphQLSchema(queries []QueryField, types []*GraphQLType, output string) error {
    // Generate SDL
    schemaBuilder := &strings.Builder{}
    
    // Write types
    for _, typ := range types {
        schemaBuilder.WriteString(typ.ToSDL())
        schemaBuilder.WriteString("\n\n")
    }
    
    // Write Query root
    schemaBuilder.WriteString("type Query {\n")
    for _, query := range queries {
        schemaBuilder.WriteString(fmt.Sprintf("  %s\n", query.ToSDL()))
    }
    schemaBuilder.WriteString("}\n")
    
    return os.WriteFile(output, []byte(schemaBuilder.String()), 0644)
}
```

## CLI Interface

```bash
# Convert OpenAPI to GraphQL
duh2graphql \
  --input api.openapi.yaml \
  --output schema.graphql \
  --naming-strategy idiomatic  # or "preserve" to keep subject.action

# Output
✓ Parsed 24 operations
✓ Filtered to 12 query operations
✓ Generated 12 GraphQL query fields
✓ Wrote schema to schema.graphql
```

## Full Example

### Input: `api.openapi.yaml` (DUH-RPC)

```yaml
openapi: 3.0.0
info:
  title: Example API
paths:
  /v1/users.get:
    post:
      operationId: v1_users_get
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                id:
                  type: string
              required: [id]
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
                  email:
                    type: string
  /v1/users.list:
    post:
      operationId: v1_users_list
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                limit:
                  type: integer
                offset:
                  type: integer
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  users:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
```

### Output: `schema.graphql`

```graphql
type User {
  id: ID!
  name: String!
  email: String
}

input UserGetInput {
  id: ID!
}

input UserListInput {
  limit: Int
  offset: Int
}

type UserListResponse {
  users: [User!]!
}

type Query {
  user(input: UserGetInput!): User
  users(input: UserListInput): UserListResponse
}
```

## Error Handling Strategy

For DUH-RPC errors, you mentioned error schemas are returned. GraphQL convention:

**Option A: Nullable responses (errors in data)**

```graphql
type Query {
  user(input: UserGetInput!): User  # null on error
}
```

**Option B: Union types (explicit errors)**

```graphql
type Query {
  user(input: UserGetInput!): UserResult!
}

union UserResult = User | Error

type Error {
  code: String!
  message: String!
}
```

**Recommendation**: Start with Option A (nullable) for simplicity. The GraphQL resolver layer will handle errors and return null with error details in the GraphQL errors array.

## Implementation Priorities

### **MVP (Week 1)**

1. Parse DUH-RPC OpenAPI with libopenapi
2. Filter operations by path pattern (subject.action where action is query-like)
3. Basic type mapping (primitives, objects, arrays)
4. Generate Query fields with inputs
5. Write SDL file

### **Phase 2 (Week 2)**

6. Proper naming strategy (pluralization, camelCase)
7. Handle $ref resolution and type deduplication
8. Handle required vs optional fields (non-null)
9. CLI with options

### **Phase 3 (Polish)**

10. Error type generation
11. Custom scalar support (dates, etc.)
12. Validation and warnings
13. Documentation comments in SDL

## Key Decisions Needed

1. **Input wrapping**: Should all inputs be wrapped?
    
    - `user(input: UserGetInput!)` ← Recommended (consistent)
    - `user(id: ID!)` ← Simpler (but less consistent)
2. **Response wrapping**: List responses often wrapped?
    
    - `users: [User!]!` ← Direct (simpler)
    - `users: UserListResponse` where `type UserListResponse { users: [User!]! }` ← Explicit (better for metadata)
3. **Naming collisions**: If `users.get` and `user.get` both exist?
    
    - Add suffix: `user1`, `user2`
    - Use full path: `usersGet`, `userGet`
    - Error and require manual resolution?

Let me know your preferences and I can help with the implementation!