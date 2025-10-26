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
paths:
  /v1/users.create:
    post:
      summary: Create a new user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /v1/users.get:
    post:
      summary: Get user by ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /v1/users.update:
    post:
      summary: Update a user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UpdateUserResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
    GetUserRequest:
      type: object
      properties:
        id:
          type: string
    UpdateUserRequest:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
    UserResponse:
      type: object
      properties:
        id:
          type: string
    UpdateUserResponse:
      type: object
      properties:
        id:
          type: string
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
        message:
          type: string
`

func TestGenerateDuhCreatesServerFile(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "server.go")

	_, err := os.Stat(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
}

func TestGeneratedServerCompiles(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	protoDir := filepath.Join(tempDir, "proto/v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))

	protoStub := `syntax = "proto3";

package duh.api.v1;

message CreateUserRequest {}
message UserResponse {}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.proto"), []byte(protoStub), 0644))

	goProtoStub := `package v1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type CreateUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateUserRequest) Reset() {}
func (x *CreateUserRequest) String() string { return "CreateUserRequest{}" }
func (x *CreateUserRequest) ProtoMessage() {}
func (x *CreateUserRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UserResponse) Reset() {}
func (x *UserResponse) String() string { return "UserResponse{}" }
func (x *UserResponse) ProtoMessage() {}
func (x *UserResponse) ProtoReflect() protoreflect.Message {
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

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

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

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})
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
