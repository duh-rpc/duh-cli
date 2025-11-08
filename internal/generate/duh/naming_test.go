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
        - code
        - message
      properties:
        code:
          type: integer
        message:
          type: string
`

const userProfilesGetByIdSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/user-profiles.get-by-id:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserProfileRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserProfileResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    GetUserProfileRequest:
      type: object
      properties:
        id:
          type: string
    UserProfileResponse:
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

const underscoreSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/user_profiles.get_by_id:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserProfileRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserProfileResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    GetUserProfileRequest:
      type: object
      properties:
        id:
          type: string
    UserProfileResponse:
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

const invalidPathNoVersionSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users.create:
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
        - code
        - message
      properties:
        code:
          type: integer
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
        - code
        - message
      properties:
        code:
          type: integer
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

func TestGenerateOperationNameInvalidPathNoVersion(t *testing.T) {
	specPath, stdout := setupTest(t, invalidPathNoVersionSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 2, exitCode)
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
