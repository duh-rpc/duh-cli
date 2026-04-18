package duh_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const multiOpSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
    post:
      summary: Create a new user
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
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /users.get:
    post:
      summary: Get user by ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /users.update:
    post:
      summary: Update a user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UpdateResponse'
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
        name:
          type: string
    GetRequest:
      type: object
      properties:
        id:
          type: string
    UpdateRequest:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    GetResponse:
      type: object
      properties:
        id:
          type: string
    UpdateResponse:
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

func TestGeneratedServerCompiles(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	protoDir := filepath.Join(tempDir, "proto/v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))

	protoStub := `syntax = "proto3";

package duh.api.v1;

message CreateRequest {}
message CreateResponse {}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.proto"), []byte(protoStub), 0644))

	goProtoStub := `package v1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type CreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateRequest) Reset() {}
func (x *CreateRequest) String() string { return "CreateRequest{}" }
func (x *CreateRequest) ProtoMessage() {}
func (x *CreateRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type CreateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateResponse) Reset() {}
func (x *CreateResponse) String() string { return "CreateResponse{}" }
func (x *CreateResponse) ProtoMessage() {}
func (x *CreateResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(goProtoStub), 0644))

	goMod := `module github.com/example/test

go 1.24

require github.com/duh-rpc/duh.go v0.0.0
`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644))

	cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go=github.com/duh-rpc/duh.go@v0.10.1")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd = exec.Command("go", "build", ".")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output:\n%s", string(output))
	}
	require.NoError(t, err)
}

func TestGeneratedServerStructure(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "server.go")

	_, err := os.Stat(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)

	content := string(serverContent)

	assert.Contains(t, content, "package api")
	assert.Contains(t, content, "const (")
	assert.Contains(t, content, "RPCUsersCreate")
	assert.Contains(t, content, "type ServiceInterface interface")
	assert.Contains(t, content, "type Handler struct")
	assert.Contains(t, content, "func (h *Handler) ServeHTTP")

	assert.NotContains(t, content, "//go:build")
	assert.NotContains(t, content, "// +build")

	assert.Contains(t, content, "Code generated by 'duh generate'")
	assert.Contains(t, content, "DO NOT EDIT")
}

func TestServerWithMultipleOperations(t *testing.T) {
	specPath, stdout := setupTest(t, multiOpSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)

	content := string(serverContent)

	assert.Contains(t, content, "RPCUsersCreate")
	assert.Contains(t, content, "RPCUsersGet")
	assert.Contains(t, content, "RPCUsersUpdate")

	assert.Contains(t, content, "UsersCreate(ctx context.Context")
	assert.Contains(t, content, "UsersGet(ctx context.Context")
	assert.Contains(t, content, "UsersUpdate(ctx context.Context")

	assert.Contains(t, content, "case RPCUsersCreate:")
	assert.Contains(t, content, "case RPCUsersGet:")
	assert.Contains(t, content, "case RPCUsersUpdate:")

	assert.Contains(t, content, "func (h *Handler) handleUsersCreate")
	assert.Contains(t, content, "func (h *Handler) handleUsersGet")
	assert.Contains(t, content, "func (h *Handler) handleUsersUpdate")

	summaryCount := strings.Count(content, "// Create a new user")
	assert.Equal(t, 1, summaryCount)
}
