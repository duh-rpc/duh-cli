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
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
    post:
      operationId: createUser
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
  /users.get:
    post:
      operationId: getUserById
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
  /users.list:
    post:
      operationId: listUsers
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
  /users.update:
    post:
      operationId: updateUser
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
components:
  schemas:
    Error:
      type: object
      required: [message]
      properties:
        message: {type: string}
    CreateRequest:
      type: object
      properties:
        name: {type: string}
    CreateResponse:
      type: object
      properties:
        id: {type: string}
    GetRequest:
      type: object
      properties:
        id: {type: string}
    GetResponse:
      type: object
      properties:
        id: {type: string}
    ListRequest:
      type: object
      properties:
        pagination:
          $ref: '#/components/schemas/PaginationRequest'
    PaginationRequest:
      type: object
      properties:
        first: {type: integer, format: int32, minimum: 1, maximum: 100}
        after: {type: string}
    ListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/GetResponse'
        pagination:
          $ref: '#/components/schemas/PaginationResponse'
    PaginationResponse:
      type: object
      properties:
        endCursor: {type: string}
        hasMore: {type: boolean}
    UpdateRequest:
      type: object
      properties:
        id: {type: string}
    UpdateResponse:
      type: object
      properties:
        id: {type: string}
`

const partialSpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
    post:
      operationId: createUser
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
  /users.get:
    post:
      operationId: getUserById
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
components:
  schemas:
    Error:
      type: object
      required: [message]
      properties:
        message: {type: string}
    CreateRequest:
      type: object
      properties:
        name: {type: string}
    CreateResponse:
      type: object
      properties:
        id: {type: string}
    GetRequest:
      type: object
      properties:
        id: {type: string}
    GetResponse:
      type: object
      properties:
        id: {type: string}
`

const customSpec = `openapi: 3.0.3
info:
  title: Custom API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /products.create:
    post:
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ProductsCreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProductsCreateResponse'
components:
  schemas:
    Error:
      type: object
      required: [message]
      properties:
        message: {type: string}
    ProductsCreateRequest:
      type: object
      properties:
        name: {type: string}
    ProductsCreateResponse:
      type: object
      properties:
        id: {type: string}
`

const extraEndpointsSpec = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
    post:
      operationId: createUser
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
  /users.get:
    post:
      operationId: getUserById
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
  /users.list:
    post:
      operationId: listUsers
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
  /users.update:
    post:
      operationId: updateUser
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
  /products.create:
    post:
      operationId: createProduct
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ProductsCreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProductsCreateResponse'
components:
  schemas:
    Error:
      type: object
      required: [message]
      properties:
        message: {type: string}
    CreateRequest:
      type: object
      properties:
        name: {type: string}
    CreateResponse:
      type: object
      properties:
        id: {type: string}
    GetRequest:
      type: object
      properties:
        id: {type: string}
    GetResponse:
      type: object
      properties:
        id: {type: string}
    ListRequest:
      type: object
      properties:
        pagination:
          $ref: '#/components/schemas/PaginationRequest'
    PaginationRequest:
      type: object
      properties:
        first: {type: integer, format: int32, minimum: 1, maximum: 100}
        after: {type: string}
    ListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/GetResponse'
        pagination:
          $ref: '#/components/schemas/PaginationResponse'
    PaginationResponse:
      type: object
      properties:
        endCursor: {type: string}
        hasMore: {type: boolean}
    UpdateRequest:
      type: object
      properties:
        id: {type: string}
    UpdateResponse:
      type: object
      properties:
        id: {type: string}
    ProductsCreateRequest:
      type: object
      properties:
        name: {type: string}
    ProductsCreateResponse:
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

	t.Cleanup(func() { _ = os.Chdir(testStartDir) })
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

	t.Cleanup(func() { _ = os.Chdir(testStartDir) })
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

	t.Cleanup(func() { _ = os.Chdir(testStartDir) })
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

	t.Cleanup(func() { _ = os.Chdir(testStartDir) })
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

	t.Cleanup(func() { _ = os.Chdir(testStartDir) })
	require.NoError(t, os.Chdir(tempDir))

	var err error
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
