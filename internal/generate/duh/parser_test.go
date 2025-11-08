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
      summary: Update user
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
        name:
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

const inlineSchemaSpec = `openapi: 3.0.0
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
                $ref: '#/components/schemas/UserResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
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

const listOperationVariantsSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.list:
    post:
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
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /v1/users.list-active:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListActiveUsersRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListActiveUsersResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
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
    ListActiveUsersRequest:
      type: object
      properties:
        offset:
          type: integer
        limit:
          type: integer
    ListActiveUsersResponse:
      type: object
      properties:
        active_users:
          type: array
          items:
            $ref: '#/components/schemas/UserResponse'
        total:
          type: integer
    UserResponse:
      type: object
      properties:
        id:
          type: string
        name:
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

const arrayOrderSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/data.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListDataRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListDataResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListDataRequest:
      type: object
      properties:
        offset:
          type: integer
        limit:
          type: integer
    ListDataResponse:
      type: object
      properties:
        total:
          type: integer
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

const notListNoOffsetSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.list:
    post:
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
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListUsersRequest:
      type: object
      properties:
        limit:
          type: integer
    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/UserResponse'
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
	assert.Contains(t, content, "pb.CreateUserRequest")
	assert.Contains(t, content, "pb.UserResponse")
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

func TestIsListOperationWithoutOffset(t *testing.T) {
	specPath, stdout := setupTest(t, notListNoOffsetSpec)

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
