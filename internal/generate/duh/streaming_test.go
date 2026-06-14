package duh_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fixed, lint-clean scaffolding shared by the streaming specs: an error schema
// for the 400 response and a tiny request schema. Each spec below varies only the
// path, the 200 response content, and the response payload schema.
const streamingComponents = `components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    JobsRunRequest:
      type: object
      description: Request payload to run a job
      properties:
        command:
          type: string
          description: The command to run
`

const specOctetStream = `openapi: 3.0.3
info:
  title: Jobs API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /jobs.run:
    post:
      summary: Run a job and stream its raw output
      description: Starts a job and streams the raw output bytes
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/JobsRunRequest'
      responses:
        '200':
          description: Raw job output stream
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
` + streamingComponents

const specStructuredJSONStream = `openapi: 3.0.3
info:
  title: Events API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /events.watch:
    post:
      summary: Stream events to the client
      description: Streams events as a structured JSON stream
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/EventsWatchRequest'
      responses:
        '200':
          description: A structured stream of events
          content:
            application/duh-stream+json:
              schema:
                $ref: '#/components/schemas/EventsWatchResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    EventsWatchRequest:
      type: object
      description: Request payload to watch events
      properties:
        since:
          type: string
          description: Only events after this cursor
    EventsWatchResponse:
      type: object
      description: A single event frame
      properties:
        name:
          type: string
          description: The event name
`

// specStructuredProtobufStream is the JSON-stream spec with the response content
// type swapped to the protobuf streaming type.
var specStructuredProtobufStream = strings.ReplaceAll(specStructuredJSONStream,
	"application/duh-stream+json", "application/duh-stream+protobuf")

const specContentEndpoint = `openapi: 3.0.3
info:
  title: Pages API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pages.render:
    post:
      summary: Render a page
      description: Returns rendered HTML content
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PagesRenderRequest'
      responses:
        '200':
          description: Rendered HTML
          content:
            text/html:
              schema:
                type: string
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    PagesRenderRequest:
      type: object
      description: Request payload to render a page
      properties:
        path:
          type: string
          description: The page path
`

// specMixed pairs an unstructured octet-stream endpoint with an ordinary typed
// endpoint so a single spec exercises both generated shapes.
const specMixed = `openapi: 3.0.3
info:
  title: Jobs API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /jobs.run:
    post:
      summary: Run a job and stream its raw output
      description: Starts a job and streams the raw output bytes
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/JobsRunRequest'
      responses:
        '200':
          description: Raw job output stream
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /jobs.status:
    post:
      summary: Get the status of a job
      description: Returns the structured status of a job
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/JobsStatusRequest'
      responses:
        '200':
          description: The job status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/JobsStatusResponse'
        '400':
          description: Bad request error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      description: Standard error details payload
      required:
        - message
      properties:
        message:
          type: string
          description: Human readable error message
    JobsRunRequest:
      type: object
      description: Request payload to run a job
      properties:
        command:
          type: string
          description: The command to run
    JobsStatusRequest:
      type: object
      description: Request payload to query a job's status
      properties:
        job_id:
          type: string
          description: The job identifier
    JobsStatusResponse:
      type: object
      description: The status of a job
      properties:
        state:
          type: string
          description: The current job state
`

// An unstructured octet-stream response generates a raw-bytes client method and a
// BytesResponse server method; the JSON request message still flows into the lock.
func TestGenerateOctetStreamResponse(t *testing.T) {
	dir := writeProject(t, specOctetStream)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "JobsRun(ctx context.Context, req *pb.JobsRunRequest) (io.ReadCloser, error)")
	assert.Contains(t, client, "c.client.DoBytes(ctx, r)")
	assert.Contains(t, client, `r.Header.Set("Accept", duh.ContentOctetStream)`)

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "JobsRun(ctx context.Context, req *pb.JobsRunRequest, w duh.BytesWriter) error")
	assert.Contains(t, server, "duh.HandleBytes(w, r, h.handleJobsRun)")
	assert.Contains(t, server, "func (h *Handler) handleJobsRun(r *http.Request, w duh.BytesWriter) error")

	// The octet-stream body has no proto message, but the JSON request message is
	// still locked, so the lock is populated and FIELDMAP_LOCK is satisfiable.
	lock := readFile(t, filepath.Join(dir, "fieldmap.lock"))
	assert.Contains(t, lock, "JobsRunRequest:")
	assert.Contains(t, lock, "command: {number: 1}")
}

// A structured JSON stream generates a StreamReader client and a StreamWriter
// server dispatched through duh.HandleStream.
func TestGenerateStructuredJSONStream(t *testing.T) {
	dir := writeProject(t, specStructuredJSONStream)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "EventsWatch(ctx context.Context, req *pb.EventsWatchRequest) (duh.StreamReader, error)")
	assert.Contains(t, client, "c.client.DoStream(ctx, r)")
	assert.Contains(t, client, `r.Header.Set("Accept", duh.ContentStreamJSON)`)

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "EventsWatch(ctx context.Context, req *pb.EventsWatchRequest, stream duh.StreamWriter) error")
	assert.Contains(t, server, "duh.HandleStream(w, r, h.handleEventsWatch, nil)")
}

// A structured protobuf stream differs from the JSON stream only in the Accept
// header content type the client sends.
func TestGenerateStructuredProtobufStream(t *testing.T) {
	dir := writeProject(t, specStructuredProtobufStream)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "EventsWatch(ctx context.Context, req *pb.EventsWatchRequest) (duh.StreamReader, error)")
	assert.Contains(t, client, "c.client.DoStream(ctx, r)")
	assert.Contains(t, client, `r.Header.Set("Accept", duh.ContentStreamProtoBuf)`)
}

// A single spec mixing an octet-stream endpoint and an ordinary typed endpoint
// generates both shapes correctly.
func TestGenerateMixedStreamAndTyped(t *testing.T) {
	dir := writeProject(t, specMixed)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "JobsRun(ctx context.Context, req *pb.JobsRunRequest) (io.ReadCloser, error)")
	assert.Contains(t, client, "JobsStatus(ctx context.Context, req *pb.JobsStatusRequest, resp *pb.JobsStatusResponse) error")
}

// A content endpoint (opaque body in a native MIME type) is not yet supported by
// the generator; it fails with a clear, actionable error pointing at ENG-100
// rather than the cryptic "inline schema not supported".
func TestGenerateContentEndpointNotSupported(t *testing.T) {
	writeProject(t, specContentEndpoint)

	exitCode, out := generate(t, "--output-dir", "out")
	assert.Equal(t, 2, exitCode, out)
	assert.Contains(t, out, "content endpoints not yet supported")
	assert.Contains(t, out, "text/html")
	assert.Contains(t, out, "ENG-100")
}

// The steve flow end to end: generate populates the lock for the JSON messages
// while skipping the octet-stream endpoint, and a subsequent lint passes clean,
// proving FIELDMAP_LOCK is satisfiable for a hand-authored streaming library.
func TestGenerateThenLintOctetStream(t *testing.T) {
	writeProject(t, specOctetStream)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	exitCode, out = lint(t)
	require.Equal(t, 0, exitCode, out)
	assert.Contains(t, out, "compliant")
}
