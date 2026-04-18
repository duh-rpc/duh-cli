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
          description: User created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad request
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
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetResponse'
        '400':
          description: Invalid user ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /users.list:
    post:
      summary: List users with pagination
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListRequest'
      responses:
        '200':
          description: Users retrieved
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
        '400':
          description: Invalid pagination
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
          description: User updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UpdateResponse'
        '400':
          description: Invalid user data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
    CreateRequest:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
    CreateResponse:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    GetRequest:
      type: object
      properties:
        userId:
          type: string
    GetResponse:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    ListRequest:
      type: object
      properties:
        pagination:
          $ref: '#/components/schemas/PaginationRequest'
    PaginationRequest:
      type: object
      properties:
        first:
          type: integer
          format: int32
          minimum: 1
          maximum: 100
        after:
          type: string
    ListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/User'
        pagination:
          $ref: '#/components/schemas/PaginationResponse'
    PaginationResponse:
      type: object
      properties:
        endCursor:
          type: string
        hasMore:
          type: boolean
    User:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    UpdateRequest:
      type: object
      properties:
        userId:
          type: string
        name:
          type: string
    UpdateResponse:
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
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
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
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /users.get:
    post:
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
        userId:
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
    ErrorDetails:
      type: object
      required:
        - message
      properties:
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "6 file(s)")

	_, err := os.Stat(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "iterator.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "proto/v1/api.proto"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "buf.yaml"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "buf.gen.yaml"))
	require.NoError(t, err)
}

func TestGeneratedCodeCompiles(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n\ngo 1.24\n\nrequire github.com/duh-rpc/duh.go v0.0.0\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	protoDir := filepath.Join(tempDir, "proto/v1")
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

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *GetRequest) Reset() {}
func (x *GetRequest) String() string { return "GetRequest{}" }
func (x *GetRequest) ProtoMessage() {}
func (x *GetRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *GetResponse) Reset() {}
func (x *GetResponse) String() string { return "GetResponse{}" }
func (x *GetResponse) ProtoMessage() {}
func (x *GetResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type PaginationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	First         int32
	After         string
}
func (x *PaginationRequest) Reset() {}
func (x *PaginationRequest) String() string { return "PaginationRequest{}" }
func (x *PaginationRequest) ProtoMessage() {}
func (x *PaginationRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type PaginationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	EndCursor     string
	HasMore       bool
}
func (x *PaginationResponse) Reset() {}
func (x *PaginationResponse) String() string { return "PaginationResponse{}" }
func (x *PaginationResponse) ProtoMessage() {}
func (x *PaginationResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Pagination    *PaginationRequest
}
func (x *ListRequest) Reset() {}
func (x *ListRequest) String() string { return "ListRequest{}" }
func (x *ListRequest) ProtoMessage() {}
func (x *ListRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Items         []*User
	Pagination    *PaginationResponse
}
func (x *ListResponse) Reset() {}
func (x *ListResponse) String() string { return "ListResponse{}" }
func (x *ListResponse) ProtoMessage() {}
func (x *ListResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *User) Reset() {}
func (x *User) String() string { return "User{}" }
func (x *User) ProtoMessage() {}
func (x *User) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateRequest) Reset() {}
func (x *UpdateRequest) String() string { return "UpdateRequest{}" }
func (x *UpdateRequest) ProtoMessage() {}
func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateResponse) Reset() {}
func (x *UpdateResponse) String() string { return "UpdateResponse{}" }
func (x *UpdateResponse) ProtoMessage() {}
func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})
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

func TestProtoImportPaths(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/myorg/myproject\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(simpleValidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})

	require.Equal(t, 2, exitCode)
}

func TestFullPipelineWithDependencies(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/integration\n\ngo 1.24\n\nrequire github.com/duh-rpc/duh.go v0.0.0\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(fullSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath, "-p", "myapi", "--proto-path", "api/v1/service.proto"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "6 file(s)")

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

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *GetRequest) Reset() {}
func (x *GetRequest) String() string { return "GetRequest{}" }
func (x *GetRequest) ProtoMessage() {}
func (x *GetRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *GetResponse) Reset() {}
func (x *GetResponse) String() string { return "GetResponse{}" }
func (x *GetResponse) ProtoMessage() {}
func (x *GetResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type PaginationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	First         int32
	After         string
}
func (x *PaginationRequest) Reset() {}
func (x *PaginationRequest) String() string { return "PaginationRequest{}" }
func (x *PaginationRequest) ProtoMessage() {}
func (x *PaginationRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type PaginationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	EndCursor     string
	HasMore       bool
}
func (x *PaginationResponse) Reset() {}
func (x *PaginationResponse) String() string { return "PaginationResponse{}" }
func (x *PaginationResponse) ProtoMessage() {}
func (x *PaginationResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Pagination    *PaginationRequest
}
func (x *ListRequest) Reset() {}
func (x *ListRequest) String() string { return "ListRequest{}" }
func (x *ListRequest) ProtoMessage() {}
func (x *ListRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Items         []*User
	Pagination    *PaginationResponse
}
func (x *ListResponse) Reset() {}
func (x *ListResponse) String() string { return "ListResponse{}" }
func (x *ListResponse) ProtoMessage() {}
func (x *ListResponse) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *User) Reset() {}
func (x *User) String() string { return "User{}" }
func (x *User) ProtoMessage() {}
func (x *User) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateRequest) Reset() {}
func (x *UpdateRequest) String() string { return "UpdateRequest{}" }
func (x *UpdateRequest) ProtoMessage() {}
func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	return (&protoimpl.MessageInfo{}).MessageOf(x)
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}
func (x *UpdateResponse) Reset() {}
func (x *UpdateResponse) String() string { return "UpdateResponse{}" }
func (x *UpdateResponse) ProtoMessage() {}
func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "5 file(s)")

	_, err := os.Stat(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "proto/v1/api.proto"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "buf.yaml"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(tempDir, "buf.gen.yaml"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tempDir, "iterator.go"))
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
