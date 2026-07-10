package duh_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exampleBesideRef carries a media-type `example` beside the $ref schema on both the
// request and response, with the protobuf wire twin co-declared — the exact shape of
// the git-server graph-API contract that ENG-135 reported as rejected. An example is
// documentation, not structure; it must never affect type derivation.
const exampleBesideRef = `openapi: 3.0.3
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
            example:
              question: What is DUH?
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
              example:
                id: poll-42
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

// examplesPlural uses the OpenAPI plural `examples` map instead of the singular
// `example`. Lint warns (PROHIBITED_MULTIPLE_EXAMPLES) but does not error, so
// generation must still succeed.
const examplesPlural = `openapi: 3.0.3
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
            examples:
              basic:
                summary: A poll
                value:
                  question: What is DUH?
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
              examples:
                basic:
                  summary: A created poll
                  value:
                    id: poll-42
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

// A media-type example beside a $ref schema generates the same unary endpoint as the
// spec without it: the example is a sibling of `schema`, not part of it, and must not
// be mistaken for an inline schema (ENG-135).
func TestGenerateMediaTypeExampleBesideRef(t *testing.T) {
	dir := writeProject(t, exampleBesideRef)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")
}

// The plural `examples` map is lint-discouraged (warning) but must not block
// generation: whatever lint accepts, generate accepts.
func TestGenerateMediaTypeExamplesPlural(t *testing.T) {
	dir := writeProject(t, examplesPlural)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "PollsCreate(ctx context.Context, req *pb.CreateRequest, resp *pb.CreateResponse) error")
}
