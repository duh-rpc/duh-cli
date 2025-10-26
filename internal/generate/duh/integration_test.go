package duh_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fullSpec = `openapi: 3.0.3
info:
  title: DUH-RPC Example API
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
          description: User created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateUserResponse'
        '400':
          description: Bad request
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
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          description: Invalid user ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /v1/users.list:
    post:
      summary: List users with pagination
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListUsersRequest'
      responses:
        '200':
          description: Users retrieved
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'
        '400':
          description: Invalid pagination
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
          description: User updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UpdateUserResponse'
        '400':
          description: Invalid user data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
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
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
    CreateUserResponse:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    GetUserRequest:
      type: object
      properties:
        userId:
          type: string
    UserResponse:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    ListUsersRequest:
      type: object
      properties:
        offset:
          type: integer
        limit:
          type: integer
    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/UserResponse'
        total:
          type: integer
    UpdateUserRequest:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    UpdateUserResponse:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
`

const specWithoutListOps = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
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
        userId:
          type: string
    UserResponse:
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

const invalidSpec = `openapi: 3.0.0
info:
  title: Invalid API
  version: 1.0.0
paths:
  /v1/api/users:
    post:
      responses:
        '200':
          description: Success
`

func TestEndToEndGeneration(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "4 file(s)")

	_, err := os.Stat(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "iterator.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "proto/v1/api.proto"))
	require.NoError(t, err)
}

func TestGeneratedCodeCompiles(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n\ngo 1.24\n\nrequire github.com/duh-rpc/duh.go v0.0.0\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	protoDir := filepath.Join(tempDir, "proto/v1")
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

type CreateUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *CreateUserResponse) Reset() {}
func (x *CreateUserResponse) String() string { return "CreateUserResponse{}" }
func (x *CreateUserResponse) ProtoMessage() {}
func (x *CreateUserResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type GetUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *GetUserRequest) Reset() {}
func (x *GetUserRequest) String() string { return "GetUserRequest{}" }
func (x *GetUserRequest) ProtoMessage() {}
func (x *GetUserRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Users         []*UserResponse
	Total         int32
}
func (x *UserResponse) Reset() {}
func (x *UserResponse) String() string { return "UserResponse{}" }
func (x *UserResponse) ProtoMessage() {}
func (x *UserResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListUsersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Offset        int32
	Limit         int32
}
func (x *ListUsersRequest) Reset() {}
func (x *ListUsersRequest) String() string { return "ListUsersRequest{}" }
func (x *ListUsersRequest) ProtoMessage() {}
func (x *ListUsersRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListUsersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Users         []*UserResponse
	Total         int32
}
func (x *ListUsersResponse) Reset() {}
func (x *ListUsersResponse) String() string { return "ListUsersResponse{}" }
func (x *ListUsersResponse) ProtoMessage() {}
func (x *ListUsersResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateUserRequest) Reset() {}
func (x *UpdateUserRequest) String() string { return "UpdateUserRequest{}" }
func (x *UpdateUserRequest) ProtoMessage() {}
func (x *UpdateUserRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateUserResponse) Reset() {}
func (x *UpdateUserResponse) String() string { return "UpdateUserResponse{}" }
func (x *UpdateUserResponse) ProtoMessage() {}
func (x *UpdateUserResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(goProtoStub), 0644))

	replaceCmd := exec.Command("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go=github.com/duh-rpc/duh.go@v0.10.1")
	replaceCmd.Dir = tempDir
	output, err := replaceCmd.CombinedOutput()
	require.NoError(t, err, string(output))

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	output, err = tidyCmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

func TestGeneratedCodeStructure(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	serverStr := string(serverContent)

	assert.Contains(t, serverStr, "package api")
	assert.Contains(t, serverStr, "RPCUsersCreate")
	assert.Contains(t, serverStr, "RPCUsersGet")
	assert.Contains(t, serverStr, "RPCUsersList")
	assert.Contains(t, serverStr, "RPCUsersUpdate")
	assert.Contains(t, serverStr, "type ServiceInterface interface")
	assert.Contains(t, serverStr, "type Handler struct")
	assert.Contains(t, serverStr, "func (h *Handler) ServeHTTP")
	assert.NotContains(t, serverStr, "//go:build")

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	clientStr := string(clientContent)

	assert.Contains(t, clientStr, "package api")
	assert.Contains(t, clientStr, "type ClientInterface interface")
	assert.Contains(t, clientStr, "func NewClient")
	assert.Contains(t, clientStr, "func WithTLS")
	assert.Contains(t, clientStr, "func WithNoTLS")
	assert.NotContains(t, clientStr, "//go:build")

	iteratorContent, err := os.ReadFile(filepath.Join(tempDir, "iterator.go"))
	require.NoError(t, err)
	iteratorStr := string(iteratorContent)

	assert.Contains(t, iteratorStr, "package api")
	assert.Contains(t, iteratorStr, "type Page[T any]")
	assert.Contains(t, iteratorStr, "type Iterator[T any]")
	assert.Contains(t, iteratorStr, "type PageFetcher[T any]")
	assert.Contains(t, iteratorStr, "type GenericIterator[T any]")
	assert.NotContains(t, iteratorStr, "//go:build")
}

func TestListOperationIterator(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	clientStr := string(clientContent)

	assert.Contains(t, clientStr, "type UserPageFetcher struct")
	assert.Contains(t, clientStr, "func (c *Client) UsersListIter")
	assert.Contains(t, clientStr, "func (f *UserPageFetcher) FetchPage")
}

func TestMultipleOperations(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	serverStr := string(serverContent)

	assert.Contains(t, serverStr, "RPCUsersCreate")
	assert.Contains(t, serverStr, "RPCUsersGet")
	assert.Contains(t, serverStr, "RPCUsersList")
	assert.Contains(t, serverStr, "RPCUsersUpdate")

	assert.Contains(t, serverStr, "UsersCreate(")
	assert.Contains(t, serverStr, "UsersGet(")
	assert.Contains(t, serverStr, "UsersList(")
	assert.Contains(t, serverStr, "UsersUpdate(")

	assert.Contains(t, serverStr, "case RPCUsersCreate:")
	assert.Contains(t, serverStr, "case RPCUsersGet:")
	assert.Contains(t, serverStr, "case RPCUsersList:")
	assert.Contains(t, serverStr, "case RPCUsersUpdate:")
}

func TestProtoImportPaths(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/myorg/myproject\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(simpleValidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	assert.Contains(t, string(serverContent), "github.com/myorg/myproject/proto/v1")

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	assert.Contains(t, string(clientContent), "github.com/myorg/myproject/proto/v1")
}

func TestTimestampInHeaders(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})
	require.Equal(t, 0, exitCode)

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	serverStr := string(serverContent)
	assert.Contains(t, serverStr, "// Code generated by 'duh generate' on")
	assert.Contains(t, serverStr, "UTC. DO NOT EDIT.")

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	clientStr := string(clientContent)
	assert.Contains(t, clientStr, "// Code generated by 'duh generate' on")
	assert.Contains(t, clientStr, "UTC. DO NOT EDIT.")

	iteratorContent, err := os.ReadFile(filepath.Join(tempDir, "iterator.go"))
	require.NoError(t, err)
	iteratorStr := string(iteratorContent)
	assert.Contains(t, iteratorStr, "// Code generated by 'duh generate' on")
	assert.Contains(t, iteratorStr, "UTC. DO NOT EDIT.")

	lines := strings.Split(serverStr, "\n")
	require.Greater(t, len(lines), 0)
	serverTimestamp := lines[0]

	lines = strings.Split(clientStr, "\n")
	require.Greater(t, len(lines), 0)
	clientTimestamp := lines[0]

	lines = strings.Split(iteratorStr, "\n")
	require.Greater(t, len(lines), 0)
	iteratorTimestamp := lines[0]

	assert.Equal(t, serverTimestamp, clientTimestamp)
	assert.Equal(t, serverTimestamp, iteratorTimestamp)
}

func TestNonAtomicGeneration(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 2, exitCode)
}

func TestFullPipelineWithDependencies(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/integration\n\ngo 1.24\n\nrequire github.com/duh-rpc/duh.go v0.0.0\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath, "-p", "myapi", "--proto-path", "api/v1/service.proto"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "4 file(s)")

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	serverStr := string(serverContent)
	assert.Contains(t, serverStr, "package myapi")
	assert.Contains(t, serverStr, "github.com/example/integration/api/v1")

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	clientStr := string(clientContent)
	assert.Contains(t, clientStr, "package myapi")
	assert.Contains(t, clientStr, "github.com/example/integration/api/v1")

	iteratorContent, err := os.ReadFile(filepath.Join(tempDir, "iterator.go"))
	require.NoError(t, err)
	assert.Contains(t, string(iteratorContent), "package myapi")

	protoContent, err := os.ReadFile(filepath.Join(tempDir, "api/v1/service.proto"))
	require.NoError(t, err)
	protoStr := string(protoContent)
	assert.Contains(t, protoStr, `syntax = "proto3"`)
	assert.Contains(t, protoStr, "package duh.api.v1")

	protoDir := filepath.Join(tempDir, "api/v1")
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

type CreateUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *CreateUserResponse) Reset() {}
func (x *CreateUserResponse) String() string { return "CreateUserResponse{}" }
func (x *CreateUserResponse) ProtoMessage() {}
func (x *CreateUserResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type GetUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *GetUserRequest) Reset() {}
func (x *GetUserRequest) String() string { return "GetUserRequest{}" }
func (x *GetUserRequest) ProtoMessage() {}
func (x *GetUserRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Users         []*UserResponse
	Total         int32
}
func (x *UserResponse) Reset() {}
func (x *UserResponse) String() string { return "UserResponse{}" }
func (x *UserResponse) ProtoMessage() {}
func (x *UserResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListUsersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Offset        int32
	Limit         int32
}
func (x *ListUsersRequest) Reset() {}
func (x *ListUsersRequest) String() string { return "ListUsersRequest{}" }
func (x *ListUsersRequest) ProtoMessage() {}
func (x *ListUsersRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListUsersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Users         []*UserResponse
	Total         int32
}
func (x *ListUsersResponse) Reset() {}
func (x *ListUsersResponse) String() string { return "ListUsersResponse{}" }
func (x *ListUsersResponse) ProtoMessage() {}
func (x *ListUsersResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateUserRequest) Reset() {}
func (x *UpdateUserRequest) String() string { return "UpdateUserRequest{}" }
func (x *UpdateUserRequest) ProtoMessage() {}
func (x *UpdateUserRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateUserResponse) Reset() {}
func (x *UpdateUserResponse) String() string { return "UpdateUserResponse{}" }
func (x *UpdateUserResponse) ProtoMessage() {}
func (x *UpdateUserResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "service.pb.go"), []byte(goProtoStub), 0644))

	replaceCmd := exec.Command("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go=github.com/duh-rpc/duh.go@v0.10.1")
	replaceCmd.Dir = tempDir
	output, err := replaceCmd.CombinedOutput()
	require.NoError(t, err, string(output))

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	output, err = tidyCmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

func TestNoListOperations(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(specWithoutListOps), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "3 file(s)")

	_, err := os.Stat(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "proto/v1/api.proto"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tempDir, "iterator.go"))
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
