package duh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	duh "github.com/duh-rpc/duh-cli"
)

func getServerContentForParser(t *testing.T, specPath string) string {
	tempDir := filepath.Dir(specPath)
	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	return string(serverContent)
}

const multiOperationSpec = `openapi: 3.0.0
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
      summary: Update user
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
        name:
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

const inlineSchemaSpec = `openapi: 3.0.0
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
              type: object
              properties:
                name:
                  type: string
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
components:
  schemas:
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

const listOperationVariantsSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users.list:
    post:
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
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /users.list-active:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListActiveRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListActiveResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ListRequest:
      type: object
      properties:
        page:
          $ref: '#/components/schemas/PageRequest'
    PageRequest:
      type: object
      properties:
        first:
          type: integer
        after:
          type: string
    ListResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/User'
        page:
          $ref: '#/components/schemas/PageResponse'
    PageResponse:
      type: object
      properties:
        endCursor:
          type: string
        hasMore:
          type: boolean
    ListActiveRequest:
      type: object
      properties:
        page:
          $ref: '#/components/schemas/PageRequest'
    ListActiveResponse:
      type: object
      properties:
        active_users:
          type: array
          items:
            $ref: '#/components/schemas/User'
        page:
          $ref: '#/components/schemas/PageResponse'
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

const arrayOrderSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /data.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DataListRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DataListResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    DataListRequest:
      type: object
      properties:
        page:
          $ref: '#/components/schemas/DataPageRequest'
    DataPageRequest:
      type: object
      properties:
        first:
          type: integer
        after:
          type: string
    DataListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/DataItem'
        metadata:
          type: array
          items:
            $ref: '#/components/schemas/MetadataItem'
    DataItem:
      type: object
      properties:
        id:
          type: string
    MetadataItem:
      type: object
      properties:
        key:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

const notListNoPageSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users.list:
    post:
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
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    ListRequest:
      type: object
      properties:
        filter:
          type: string
    ListResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/User'
    User:
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

func TestParseOperationsExtractsMultiple(t *testing.T) {
	specPath, stdout := setupTest(t, multiOperationSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContentForParser(t, specPath)
	assert.Contains(t, content, "UsersCreate")
	assert.Contains(t, content, "UsersGet")
	assert.Contains(t, content, "UsersUpdate")
}

func TestParseOperationsExtractsPbPrefixedTypes(t *testing.T) {
	specPath, stdout := setupTest(t, multiOperationSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContentForParser(t, specPath)
	assert.Contains(t, content, "pb.CreateRequest")
	assert.Contains(t, content, "pb.CreateResponse")
}

func TestDetectListOperationsWith3Criteria(t *testing.T) {
	specPath, stdout := setupTest(t, specWithListOp)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
}

func TestDetectListOperationsMultipleVariants(t *testing.T) {
	specPath, stdout := setupTest(t, listOperationVariantsSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
}

func TestIsListOperationChecksMethodPortion(t *testing.T) {
	specPath, stdout := setupTest(t, listOperationVariantsSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
}

func TestIsListOperationWithoutPage(t *testing.T) {
	specPath, stdout := setupTest(t, notListNoPageSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
}

func TestFindFirstArrayFieldInYAMLOrder(t *testing.T) {
	specPath, stdout := setupTest(t, arrayOrderSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
}

func TestInlineSchemaReturnsError(t *testing.T) {
	specPath, stdout := setupTest(t, inlineSchemaSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "inline schema not supported")
}

func TestParseExtractsModulePathAndProtoImport(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContentForParser(t, specPath)
	assert.Contains(t, content, "github.com/example/test/proto/v1")
}

func TestParseGeneratesTimestampInCorrectFormat(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContentForParser(t, specPath)
	assert.Contains(t, content, "UTC")
}

func TestParseExtractsOperationSummary(t *testing.T) {
	specPath, stdout := setupTest(t, multiOperationSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContentForParser(t, specPath)
	assert.Contains(t, content, "// Create a new user")
}
