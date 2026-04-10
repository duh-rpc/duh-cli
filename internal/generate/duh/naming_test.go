package duh_test

import (
	"os"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const usersCreateSpec = `openapi: 3.0.0
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
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        name:
          type: string
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

const userProfilesGetByIdSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /user-profiles.get-by-id:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetByIdRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetByIdResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    GetByIdRequest:
      type: object
      properties:
        id:
          type: string
    GetByIdResponse:
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

const underscoreSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /user_profiles.get_by_id:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Get_by_idRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Get_by_idResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    Get_by_idRequest:
      type: object
      properties:
        id:
          type: string
    Get_by_idResponse:
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

const invalidPathNoVersionSpec = `openapi: 3.0.0
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
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        name:
          type: string
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

const invalidPathNoMethodSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users:
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
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
    UserResponse:
      type: object
      properties:
        id:
          type: string
    Error:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

func getServerContent(t *testing.T, specPath string) string {
	tempDir := filepath.Dir(specPath)
	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	return string(serverContent)
}

func TestGenerateOperationNameUsersCreate(t *testing.T) {
	specPath, stdout := setupTest(t, usersCreateSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "UsersCreate")
	assert.Contains(t, content, "RPCUsersCreate")
}

func TestGenerateOperationNameUserProfilesGetById(t *testing.T) {
	specPath, stdout := setupTest(t, userProfilesGetByIdSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "UserProfilesGetById")
	assert.Contains(t, content, "RPCUserProfilesGetById")
}

func TestGenerateOperationNameWithUnderscores(t *testing.T) {
	specPath, stdout := setupTest(t, underscoreSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "UserProfilesGetById")
}

func TestGenerateOperationNamePathWithoutVersionPrefix(t *testing.T) {
	specPath, stdout := setupTest(t, invalidPathNoVersionSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "UsersCreate")
}

func TestGenerateOperationNameInvalidPathNoMethod(t *testing.T) {
	specPath, stdout := setupTest(t, invalidPathNoMethodSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 2, exitCode)
}

func TestGenerateConstNamePrefixesRPC(t *testing.T) {
	specPath, stdout := setupTest(t, usersCreateSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "RPCUsersCreate")
}

func TestToCamelCaseWithHyphens(t *testing.T) {
	specPath, stdout := setupTest(t, userProfilesGetByIdSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "UserProfilesGetById")
}

func TestToCamelCaseWithUnderscores(t *testing.T) {
	specPath, stdout := setupTest(t, underscoreSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	content := getServerContent(t, specPath)
	assert.Contains(t, content, "UserProfilesGetById")
}
