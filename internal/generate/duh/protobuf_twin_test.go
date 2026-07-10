package duh_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// twinJSONFirst co-declares, on both request and response, the canonical DUH dual
// content types: application/json carrying the message $ref and application/protobuf
// carrying its wire twin {type:string,format:binary}. This is duh-cli's documented
// "Multiple Content Types" pattern (Rule 5), used by every demo/slip-stream endpoint.
// JSON is declared first.
const twinJSONFirst = `openapi: 3.0.3
info:
  title: Polls API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /polls.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
          application/protobuf:
            schema:
              type: string
              format: binary
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
            application/protobuf:
              schema:
                type: string
                format: binary
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        question:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// twinProtobufFirst is the same dual-content endpoint with application/protobuf (the
// binary wire twin) declared BEFORE application/json on both request and response.
// Type derivation must be independent of content-type declaration order.
const twinProtobufFirst = `openapi: 3.0.3
info:
  title: Polls API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /polls.create:
    post:
      requestBody:
        required: true
        content:
          application/protobuf:
            schema:
              type: string
              format: binary
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/protobuf:
              schema:
                type: string
                format: binary
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        question:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// twinProtobufRef declares application/protobuf as a $ref to the same message rather
// than the binary wire twin. A $ref on application/protobuf must keep working.
const twinProtobufRef = `openapi: 3.0.3
info:
  title: Polls API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /polls.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
          application/protobuf:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
            application/protobuf:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        question:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// propertylessInlineJSON declares a propertyless inline object ({type:object}, no
// properties) under application/json. The SCHEMA_NO_INLINE_OBJECTS lint rule only
// flags inline objects with properties, so this passes lint and reaches the parser —
// which must still reject it. application/json is the required structural content
// type, so the propertyless-binary exemption is application/protobuf-only and does
// not apply here.
const propertylessInlineJSON = `openapi: 3.0.3
info:
  title: Polls API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /polls.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        question:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// inlineStructuredResponse declares a genuine inline structured object (type:object
// WITH properties) under application/json — the generator cannot name it. The
// SCHEMA_NO_INLINE_OBJECTS lint rule rejects it before generation proceeds.
const inlineStructuredResponse = `openapi: 3.0.3
info:
  title: Polls API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /polls.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        question:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// The documented protobuf wire twin generates a unary endpoint: the binary
// application/protobuf entry carries no message and is skipped, so the signatures are
// derived from the application/json $ref (req *pb.CreateRequest / resp *pb.CreateResponse).
func TestGenerateProtobufTwinJSONFirst(t *testing.T) {
	dir := writeProject(t, twinJSONFirst)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")
}

// Declaring application/protobuf before application/json yields the identical unary
// signatures: content-type order does not change type derivation.
func TestGenerateProtobufTwinProtobufFirst(t *testing.T) {
	dir := writeProject(t, twinProtobufFirst)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")
}

// An application/protobuf entry that is itself a $ref (not the binary twin) still
// generates the same unary endpoint, confirming the fix did not regress $ref handling.
func TestGenerateProtobufRefStillWorks(t *testing.T) {
	dir := writeProject(t, twinProtobufRef)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")
}

// A genuinely inline structured object (an inline type:object WITH properties) under
// application/json is still rejected — the fix only exempts the propertyless binary
// twin, not nameable inline messages. The SCHEMA_NO_INLINE_OBJECTS lint rule is the
// gate here, failing generation before the parser runs. The gate prints the
// violations themselves, not just a bare failure, so the user need not re-run
// 'duh lint' to learn what to fix.
func TestGenerateInlineStructuredStillRejected(t *testing.T) {
	writeProject(t, inlineStructuredResponse)

	exitCode, out := generate(t, "--output-dir", "out")
	assert.Equal(t, 2, exitCode, out)
	assert.Contains(t, out, "validation failed")
	assert.Contains(t, out, "SCHEMA_NO_INLINE_OBJECTS")
	assert.Contains(t, out, "POST /polls.create")
}

// A propertyless inline object under application/json passes lint but is rejected by
// the parser with the inline-schema error. This proves the propertyless-binary
// exemption is conditioned on content type: application/json never accepts an inline
// schema, only application/protobuf's binary wire twin is skipped.
func TestGenerateInlinePropertylessJSONRejected(t *testing.T) {
	writeProject(t, propertylessInlineJSON)

	exitCode, out := generate(t, "--output-dir", "out")
	assert.Equal(t, 2, exitCode, out)
	assert.Contains(t, out, "inline schema not supported for response body")
}
