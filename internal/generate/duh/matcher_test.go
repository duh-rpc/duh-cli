package duh_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/require"
)

const initTemplateSpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      operationId: createUser
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
                $ref: '#/components/schemas/CreateUserResponse'
  /v1/users.get:
    post:
      operationId: getUserById
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserByIdRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetUserByIdResponse'
  /v1/users.list:
    post:
      operationId: listUsers
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListUsersRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'
  /v1/users.update:
    post:
      operationId: updateUser
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
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code: {type: integer}
        message: {type: string}
    CreateUserRequest:
      type: object
      properties:
        name: {type: string}
    CreateUserResponse:
      type: object
      properties:
        id: {type: string}
    GetUserByIdRequest:
      type: object
      properties:
        id: {type: string}
    GetUserByIdResponse:
      type: object
      properties:
        id: {type: string}
    ListUsersRequest:
      type: object
      properties:
        offset: {type: integer}
    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/GetUserByIdResponse'
    UpdateUserRequest:
      type: object
      properties:
        id: {type: string}
    UpdateUserResponse:
      type: object
      properties:
        id: {type: string}
`

const partialSpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      operationId: createUser
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
                $ref: '#/components/schemas/CreateUserResponse'
  /v1/users.get:
    post:
      operationId: getUserById
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserByIdRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetUserByIdResponse'
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code: {type: integer}
        message: {type: string}
    CreateUserRequest:
      type: object
      properties:
        name: {type: string}
    CreateUserResponse:
      type: object
      properties:
        id: {type: string}
    GetUserByIdRequest:
      type: object
      properties:
        id: {type: string}
    GetUserByIdResponse:
      type: object
      properties:
        id: {type: string}
`

const customSpec = `openapi: 3.0.3
info:
  title: Custom API
  version: 1.0.0
paths:
  /v1/products.create:
    post:
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateProductResponse'
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code: {type: integer}
        message: {type: string}
    CreateProductRequest:
      type: object
      properties:
        name: {type: string}
    CreateProductResponse:
      type: object
      properties:
        id: {type: string}
`

const extraEndpointsSpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      operationId: createUser
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
                $ref: '#/components/schemas/CreateUserResponse'
  /v1/users.get:
    post:
      operationId: getUserById
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserByIdRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetUserByIdResponse'
  /v1/users.list:
    post:
      operationId: listUsers
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListUsersRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'
  /v1/users.update:
    post:
      operationId: updateUser
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
  /v1/products.create:
    post:
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateProductResponse'
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code: {type: integer}
        message: {type: string}
    CreateUserRequest:
      type: object
      properties:
        name: {type: string}
    CreateUserResponse:
      type: object
      properties:
        id: {type: string}
    GetUserByIdRequest:
      type: object
      properties:
        id: {type: string}
    GetUserByIdResponse:
      type: object
      properties:
        id: {type: string}
    ListUsersRequest:
      type: object
      properties:
        offset: {type: integer}
    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/GetUserByIdResponse'
    UpdateUserRequest:
      type: object
      properties:
        id: {type: string}
    UpdateUserResponse:
      type: object
      properties:
        id: {type: string}
    CreateProductRequest:
      type: object
      properties:
        name: {type: string}
    CreateProductResponse:
      type: object
      properties:
        id: {type: string}
`

func TestIsInitTemplateSpecWithFullMatch(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	args := []string{"generate", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	if exitCode != 0 {
		t.Logf("Command output: %s", stdout.String())
	}

	require.Equal(t, 0, exitCode)
}

func TestIsInitTemplateSpecWithPartialMatch(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(partialSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	args := []string{"generate", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
}

func TestIsInitTemplateSpecWithNoMatch(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(customSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	args := []string{"generate", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
}

func TestIsInitTemplateSpecWithExtraEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(extraEndpointsSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	args := []string{"generate", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
}

func TestRunWithFullFlagFalse(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	args := []string{"generate", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)

	_, err = os.Stat("daemon.go")
	require.Error(t, err)

	_, err = os.Stat("service.go")
	require.Error(t, err)

	_, err = os.Stat("api_test.go")
	require.Error(t, err)

	_, err = os.Stat("Makefile")
	require.Error(t, err)
}
