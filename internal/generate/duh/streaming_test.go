package duh_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
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

// duhStreamRuntime is the duh.go/v2 release the generated streaming code compiles
// against. The structured-stream server dispatch calls the four-argument
// duh.HandleStream(w, r, handler, conf), which landed in v2.3.0; earlier releases
// expose only the three-argument form and would fail this compile check.
const duhStreamRuntime = "v2.3.0"

// eventsProtoStub satisfies the pb.EventsWatchRequest reference in the generated
// structured-stream client and server. A structured stream carries its payload
// over duh.StreamWriter/StreamReader, so the response message is never named by
// the generated code and is not stubbed here.
const eventsProtoStub = `package v1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type EventsWatchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *EventsWatchRequest) Reset()         {}
func (x *EventsWatchRequest) String() string { return "EventsWatchRequest{}" }
func (x *EventsWatchRequest) ProtoMessage()  {}
func (x *EventsWatchRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}
`

// jobsProtoStub satisfies the pb.JobsRunRequest reference in the generated
// octet-stream client and server.
const jobsProtoStub = `package v1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type JobsRunRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *JobsRunRequest) Reset()         {}
func (x *JobsRunRequest) String() string { return "JobsRunRequest{}" }
func (x *JobsRunRequest) ProtoMessage()  {}
func (x *JobsRunRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}
`

// buildGenerated drops the proto stub next to the generated client/server, pins
// the duh.go runtime that defines the streaming symbols, and runs `go build` in
// dir. It compiles the generated code against the real runtime, so a wrong or
// renamed runtime symbol (e.g. duh.HandleStream's arity, duh.StreamWriter,
// duh.ContentStreamJSON) fails here where a source-text assertion cannot.
func buildGenerated(t *testing.T, dir, protoStub string) {
	t.Helper()

	protoDir := filepath.Join(dir, "proto", "v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(protoStub), 0644))

	goMod := `module github.com/example/test

go 1.24

require github.com/duh-rpc/duh.go/v2 ` + duhStreamRuntime + `
require github.com/kapetan-io/tackle v0.0.0
require google.golang.org/protobuf v0.0.0
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))

	run := func(args ...string) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("%s\n%s", strings.Join(args, " "), out)
		}
		require.NoError(t, err)
	}

	run("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go/v2=github.com/duh-rpc/duh.go/v2@"+duhStreamRuntime)
	run("go", "mod", "edit", "-replace", "github.com/kapetan-io/tackle=github.com/kapetan-io/tackle@v0.13.0")
	run("go", "mod", "tidy")
	run("go", "build", ".")
}

// A generated structured JSON stream compiles against the real duh.go runtime,
// proving the StreamReader/StreamWriter client and server and the four-argument
// duh.HandleStream dispatch resolve to actual runtime symbols.
func TestGeneratedStructuredJSONStreamCompiles(t *testing.T) {
	specPath, stdout := setupTest(t, specStructuredJSONStream)

	require.Equal(t, 0, duh.RunCmd(context.Background(), stdout, []string{"generate", specPath}))
	buildGenerated(t, filepath.Dir(specPath), eventsProtoStub)
}

// A generated structured protobuf stream differs only in the Accept content type
// constant the client sends; it must compile against the runtime too.
func TestGeneratedStructuredProtobufStreamCompiles(t *testing.T) {
	specPath, stdout := setupTest(t, specStructuredProtobufStream)

	require.Equal(t, 0, duh.RunCmd(context.Background(), stdout, []string{"generate", specPath}))
	buildGenerated(t, filepath.Dir(specPath), eventsProtoStub)
}

// The generated octet-stream code references duh.HandleBytes and duh.BytesWriter,
// which no published duh.go/v2 release defines yet (pending duh.go#15). Skip until
// that runtime ships; remove the Skip to turn this into a live compile check.
func TestGeneratedOctetStreamCompiles(t *testing.T) {
	t.Skip("pending duh.go#15: duh.HandleBytes/duh.BytesWriter are not in a published duh.go/v2 release")

	specPath, stdout := setupTest(t, specOctetStream)

	require.Equal(t, 0, duh.RunCmd(context.Background(), stdout, []string{"generate", specPath}))
	buildGenerated(t, filepath.Dir(specPath), jobsProtoStub)
}
