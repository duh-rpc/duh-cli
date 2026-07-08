package duh_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// didYouMeanSpec declares two wrong-guess paths for the /commits.diff operation.
// A server mounted at /v1 must teach both guesses with a 404 naming the canonical
// route /v1/commits.diff.
const didYouMeanSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /commits.diff:
    x-duh-did-you-mean:
      - /diffs.get
      - /revisions.compare
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DiffRequest:
      type: object
      properties:
        id:
          type: string
    DiffResponse:
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

// TestDidYouMeanTeaching404 drives the generated handler through its ServeHTTP
// surface to prove the teaching arms behave exactly as the blueprint specifies:
// a declared wrong guess replies 404 with the canonical hint on any method, an
// unmatched path falls through writing nothing, and the canonical route keeps its
// method check.
func TestDidYouMeanTeaching404(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod",
		[]byte("module github.com/example/test\n\ngo 1.24\n\nrequire github.com/duh-rpc/duh.go/v2 v2.0.0\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(didYouMeanSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	protoDir := filepath.Join(tempDir, "proto/v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(diffProtoStub), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "server_run_test.go"), []byte(didYouMeanInnerTest), 0644))

	run := func(args ...string) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))
	}
	run("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go/v2=github.com/duh-rpc/duh.go/v2@v2.0.0")
	run("go", "mod", "tidy")
	run("go", "test", ".")
}

// diffProtoStub is the hand-written proto twin the generated server compiles
// against, providing the request/response messages for /commits.diff.
const diffProtoStub = `package v1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type DiffRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DiffRequest) Reset()         {}
func (x *DiffRequest) String() string { return "DiffRequest{}" }
func (x *DiffRequest) ProtoMessage()  {}
func (x *DiffRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type DiffResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DiffResponse) Reset()         {}
func (x *DiffResponse) String() string { return "DiffResponse{}" }
func (x *DiffResponse) ProtoMessage()  {}
func (x *DiffResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}
`

// didYouMeanInnerTest is compiled and run against the GENERATED handler by the
// parent test. It exercises the exported ServeHTTP surface only. JSON is decoded
// (not string-matched) so assertions read the wire reply any client would see.
const didYouMeanInnerTest = `package api

import (
	"context"
	stdjson "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/example/test/proto/v1"
)

type stubService struct{}

func (stubService) CommitsDiff(ctx context.Context, req *pb.DiffRequest, resp *pb.DiffResponse) error {
	return nil
}
func (stubService) Shutdown(ctx context.Context) error { return nil }

// Code is decoded as a string because duh renders the reply's int64 code as a
// JSON string (protojson stringifies 64-bit integers). The numeric HTTP status is
// asserted separately via the recorder.
type reply struct {
	Code    string            ` + "`json:\"code\"`" + `
	Message string            ` + "`json:\"message\"`" + `
	Details map[string]string ` + "`json:\"details\"`" + `
}

func TestTeachingPathReplies404WithHint(t *testing.T) {
	h := NewHandler(stubService{})

	for _, method := range []string{http.MethodPost, http.MethodGet, http.MethodDelete} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/v1/diffs.get", nil)

		matched := h.ServeHTTP(rec, req)

		assert.True(t, matched)
		assert.Equal(t, 404, rec.Code)

		var r reply
		require.NoError(t, stdjson.Unmarshal(rec.Body.Bytes(), &r))
		assert.Equal(t, "404", r.Code)
		assert.Equal(t, "no such endpoint: /v1/diffs.get", r.Message)
		assert.Equal(t, map[string]string{"did_you_mean": "/v1/commits.diff"}, r.Details)
	}
}

func TestSecondTeachingPathReplies404WithHint(t *testing.T) {
	h := NewHandler(stubService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/revisions.compare", nil)

	matched := h.ServeHTTP(rec, req)

	assert.True(t, matched)
	assert.Equal(t, 404, rec.Code)

	var r reply
	require.NoError(t, stdjson.Unmarshal(rec.Body.Bytes(), &r))
	assert.Equal(t, "no such endpoint: /v1/revisions.compare", r.Message)
	assert.Equal(t, "/v1/commits.diff", r.Details["did_you_mean"])
}

func TestUnmatchedPathFallsThrough(t *testing.T) {
	h := NewHandler(stubService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/nope.missing", nil)

	matched := h.ServeHTTP(rec, req)

	assert.False(t, matched)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, 0, rec.Body.Len())
}

func TestCanonicalRouteKeepsMethodCheck(t *testing.T) {
	h := NewHandler(stubService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/commits.diff", nil)

	matched := h.ServeHTTP(rec, req)

	assert.True(t, matched)
	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "only POST")
}
`

// noExtensionSpec is a plain single-operation spec that uses no
// x-duh-did-you-mean extension; its generated server must carry no teaching arms.
const noExtensionSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /commits.diff:
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DiffRequest:
      type: object
      properties:
        id:
          type: string
    DiffResponse:
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

// TestDidYouMeanAbsentGeneratesNoTeachingArms pins the opt-in invariant: a spec
// without the extension generates a server with no teaching constructs.
func TestDidYouMeanAbsentGeneratesNoTeachingArms(t *testing.T) {
	specPath, stdout := setupTest(t, noExtensionSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	content := string(serverContent)

	assert.NotContains(t, content, "did_you_mean")
	assert.NotContains(t, content, "no such endpoint")
	assert.NotContains(t, content, "CodeNotFound")
}

// TestDidYouMeanCollidesWithCanonical proves generate refuses a teaching path that
// equals a real route, naming both the guess and its declaring operation.
func TestDidYouMeanCollidesWithCanonical(t *testing.T) {
	specPath, stdout := setupTest(t, collidesWithCanonicalSpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "/diffs.get")
	assert.Contains(t, output, "/commits.diff")
}

// TestDidYouMeanDuplicateAcrossPaths proves generate refuses the same teaching
// path declared on two different operations, naming the path and both operations.
func TestDidYouMeanDuplicateAcrossPaths(t *testing.T) {
	specPath, stdout := setupTest(t, duplicateAcrossPathsSpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "/diffs.get")
	assert.Contains(t, output, "/commits.diff")
	assert.Contains(t, output, "/revisions.compare")
}

// TestDidYouMeanDuplicateWithinList proves generate refuses the same teaching path
// declared twice inside one operation's list.
func TestDidYouMeanDuplicateWithinList(t *testing.T) {
	specPath, stdout := setupTest(t, duplicateWithinListSpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "/diffs.get")
	assert.Contains(t, output, "declared more than once")
}

// TestDidYouMeanMalformedNonString proves generate refuses a non-string entry,
// naming the offending path.
func TestDidYouMeanMalformedNonString(t *testing.T) {
	specPath, stdout := setupTest(t, malformedNonStringSpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "/commits.diff")
	assert.Contains(t, output, "malformed")
}

// TestDidYouMeanMalformedNoLeadingSlash proves generate refuses an entry that does
// not begin with '/', naming the offending path and entry.
func TestDidYouMeanMalformedNoLeadingSlash(t *testing.T) {
	specPath, stdout := setupTest(t, malformedNoSlashSpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "diffs.get")
	assert.Contains(t, output, "malformed")
}

const collidesWithCanonicalSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /diffs.get:
    post:
      summary: Get a diff
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /commits.diff:
    x-duh-did-you-mean:
      - /diffs.get
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CommitDiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CommitDiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DiffRequest:
      type: object
      properties:
        id:
          type: string
    DiffResponse:
      type: object
      properties:
        id:
          type: string
    CommitDiffRequest:
      type: object
      properties:
        id:
          type: string
    CommitDiffResponse:
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

const duplicateAcrossPathsSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /commits.diff:
    x-duh-did-you-mean:
      - /diffs.get
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CommitDiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CommitDiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /revisions.compare:
    x-duh-did-you-mean:
      - /diffs.get
    post:
      summary: Compare revisions
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CompareRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CompareResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CommitDiffRequest:
      type: object
      properties:
        id:
          type: string
    CommitDiffResponse:
      type: object
      properties:
        id:
          type: string
    CompareRequest:
      type: object
      properties:
        id:
          type: string
    CompareResponse:
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

const duplicateWithinListSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /commits.diff:
    x-duh-did-you-mean:
      - /diffs.get
      - /diffs.get
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DiffRequest:
      type: object
      properties:
        id:
          type: string
    DiffResponse:
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

const malformedNonStringSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /commits.diff:
    x-duh-did-you-mean:
      - 123
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DiffRequest:
      type: object
      properties:
        id:
          type: string
    DiffResponse:
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

const malformedNoSlashSpec = `openapi: 3.0.0
info:
  title: Commits API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /commits.diff:
    x-duh-did-you-mean:
      - diffs.get
    post:
      summary: Diff commits
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DiffRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DiffResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DiffRequest:
      type: object
      properties:
        id:
          type: string
    DiffResponse:
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
