package duh_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDuhWithFullFlagAndInitSpec(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateWithExtraMethod), 0644))
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 10 file(s)")

	_, err = os.Stat("buf.yaml")
	require.NoError(t, err)

	_, err = os.Stat("daemon.go")
	require.NoError(t, err)

	_, err = os.Stat("service.go")
	require.NoError(t, err)

	_, err = os.Stat("api_test.go")
	require.NoError(t, err)

	_, err = os.Stat("Makefile")
	require.NoError(t, err)

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)
	assert.Contains(t, string(serviceContent), "map[string]*pb.UserResponse")
	assert.Contains(t, string(serviceContent), "func (s *Service) UsersCreate")
	assert.Contains(t, string(serviceContent), "func (s *Service) UsersGet")
	assert.Contains(t, string(serviceContent), "func (s *Service) UsersList")
	assert.Contains(t, string(serviceContent), "func (s *Service) UsersUpdate")

	apiTestContent, err := os.ReadFile("api_test.go")
	require.NoError(t, err)
	assert.Contains(t, string(apiTestContent), "func TestUsersCreate(t *testing.T)")
	assert.Contains(t, string(apiTestContent), "func TestUsersGet(t *testing.T)")
	assert.Contains(t, string(apiTestContent), "func TestUsersList(t *testing.T)")
	assert.Contains(t, string(apiTestContent), "func TestUsersUpdate(t *testing.T)")

	daemonContent, err := os.ReadFile("daemon.go")
	require.NoError(t, err)
	assert.Contains(t, string(daemonContent), "YOU CAN EDIT")

	makefileContent, err := os.ReadFile("Makefile")
	require.NoError(t, err)
	assert.Contains(t, string(makefileContent), "YOU CAN EDIT")
	assert.Contains(t, string(makefileContent), "proto:")
	assert.Contains(t, string(makefileContent), "test:")

	serverContent, err := os.ReadFile("server.go")
	require.NoError(t, err)
	serverStr := string(serverContent)
	assert.Contains(t, serverStr, "UsersCreate(ctx context.Context")
	assert.Contains(t, serverStr, "UsersGet(ctx context.Context")
	assert.Contains(t, serverStr, "UsersList(ctx context.Context")
	assert.Contains(t, serverStr, "UsersUpdate(ctx context.Context")

	serviceStr := string(serviceContent)
	assert.Contains(t, serviceStr, "UsersCreate(ctx context.Context")
	assert.Contains(t, serviceStr, "UsersGet(ctx context.Context")
	assert.Contains(t, serviceStr, "UsersList(ctx context.Context")
	assert.Contains(t, serviceStr, "UsersUpdate(ctx context.Context")
}

func TestGenerateDuhWithFullFlagAndCustomSpec(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 9 file(s)")

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)
	assert.Contains(t, string(serviceContent), "func (s *Service) ProductsCreate")
	assert.Contains(t, string(serviceContent), "CodeNotImplemented")

	apiTestContent, err := os.ReadFile("api_test.go")
	require.NoError(t, err)
	assert.Contains(t, string(apiTestContent), "TODO")
	assert.Contains(t, string(apiTestContent), "func TestProductsCreate(t *testing.T)")
}

func TestGenerateDuhWithoutFullFlag(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 6 file(s)")

	_, err = os.Stat("buf.yaml")
	require.NoError(t, err)

	_, err = os.Stat("daemon.go")
	require.Error(t, err)

	_, err = os.Stat("service.go")
	require.Error(t, err)

	_, err = os.Stat("api_test.go")
	require.Error(t, err)

	_, err = os.Stat("Makefile")
	require.Error(t, err)
}

func TestRegenerateWithFullFlagOverwrites(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)
	require.Equal(t, 0, exitCode)

	const customContent = "// MY CUSTOM EDIT"
	err = os.WriteFile("service.go", []byte(customContent), 0644)
	require.NoError(t, err)

	var stdout2 bytes.Buffer
	exitCode = duh.RunCmd(&stdout2, args)
	require.Equal(t, 0, exitCode)

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)
	assert.NotContains(t, string(serviceContent), customContent)
	assert.Contains(t, string(serviceContent), "func (s *Service) UsersCreate")
}

func TestMakefileGoesToProjectRoot(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputDir := filepath.Join(tempDir, "api")

	require.NoError(t, os.MkdirAll(outputDir, 0755))
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
	args := []string{"generate", "duh", "openapi.yaml", "--output-dir", "api", "--full"}
	exitCode := duh.RunCmd(&stdout, args)
	require.Equal(t, 0, exitCode)

	_, err = os.Stat(filepath.Join("api", "Makefile"))
	require.NoError(t, err)

	_, err = os.Stat("Makefile")
	require.Error(t, err)

	apiFiles := []string{"daemon.go", "service.go", "api_test.go", "server.go", "client.go"}
	for _, file := range apiFiles {
		filePath := filepath.Join("api", file)
		_, err = os.Stat(filePath)
		require.NoError(t, err, "file %s should exist in api directory", file)
	}
}

func TestFullGeneratedCodeFormat(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)
	require.Equal(t, 0, exitCode)

	goFiles := []string{"daemon.go", "service.go", "api_test.go", "server.go", "client.go"}
	for _, file := range goFiles {
		content, err := os.ReadFile(file)
		require.NoError(t, err)

		lines := strings.Split(string(content), "\n")
		require.NotEmpty(t, lines)

		firstLine := lines[0]
		assert.True(t, strings.HasPrefix(firstLine, "// Code generated"), "file %s should have generated header", file)

		if file == "daemon.go" || file == "service.go" || file == "api_test.go" {
			assert.Contains(t, firstLine, "YOU CAN EDIT", "file %s should have YOU CAN EDIT marker", file)
		} else {
			assert.Contains(t, firstLine, "DO NOT EDIT", "file %s should have DO NOT EDIT marker", file)
		}
	}
}

func TestBufFilesNotOverwrittenWhenExist(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	const customBufYaml = "# MY CUSTOM BUF.YAML\nversion: v2\n"
	const customBufGenYaml = "# MY CUSTOM BUF.GEN.YAML\nversion: v2\n"
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "buf.yaml"), []byte(customBufYaml), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "buf.gen.yaml"), []byte(customBufGenYaml), 0644))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	args := []string{"generate", "duh", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 4 file(s)")

	bufYamlContent, err := os.ReadFile("buf.yaml")
	require.NoError(t, err)
	assert.Equal(t, customBufYaml, string(bufYamlContent))

	bufGenYamlContent, err := os.ReadFile("buf.gen.yaml")
	require.NoError(t, err)
	assert.Equal(t, customBufGenYaml, string(bufGenYamlContent))
}

const initTemplateWithExtraMethod = `openapi: 3.0.0
info:
  title: Users API
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
                $ref: '#/components/schemas/CreateUserResponse'
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
  /v1/users.list:
    post:
      summary: List users
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
  /v1/users.delete:
    post:
      summary: Delete a user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DeleteUserRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeleteUserResponse'
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
        email:
          type: string
        age:
          type: integer
    CreateUserResponse:
      type: object
      properties:
        user_id:
          type: string
        name:
          type: string
        email:
          type: string
        created_at:
          type: string
          format: date-time
    GetUserRequest:
      type: object
      properties:
        user_id:
          type: string
    UserResponse:
      type: object
      properties:
        user_id:
          type: string
        name:
          type: string
        email:
          type: string
        age:
          type: integer
        created_at:
          type: string
          format: date-time
    ListUsersRequest:
      type: object
      properties:
        offset:
          type: integer
        limit:
          type: integer
        sort_by:
          type: string
    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/UserResponse'
        total:
          type: integer
        has_more:
          type: boolean
    UpdateUserRequest:
      type: object
      properties:
        user_id:
          type: string
        name:
          type: string
        email:
          type: string
        age:
          type: integer
        status:
          type: string
    UpdateUserResponse:
      type: object
      properties:
        user_id:
          type: string
        name:
          type: string
        email:
          type: string
        age:
          type: integer
        status:
          type: string
        updated_at:
          type: string
          format: date-time
        created_at:
          type: string
          format: date-time
    DeleteUserRequest:
      type: object
      properties:
        user_id:
          type: string
    DeleteUserResponse:
      type: object
      properties:
        user_id:
          type: string
        deleted_at:
          type: string
          format: date-time
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

func TestGenerateDuhWithFullFlagAndExtraEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateWithExtraMethod), 0644))
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)

	content := string(serviceContent)

	assert.Contains(t, content, "func (s *Service) UsersCreate")
	assert.Contains(t, content, "func (s *Service) UsersGet")
	assert.Contains(t, content, "func (s *Service) UsersList")
	assert.Contains(t, content, "func (s *Service) UsersUpdate")
	assert.Contains(t, content, "func (s *Service) UsersDelete")

	assert.Contains(t, content, `duh.NewServiceError(duh.CodeNotImplemented, "UsersDelete not implemented", nil, nil)`)

	serverContent, err := os.ReadFile("server.go")
	require.NoError(t, err)

	serverStr := string(serverContent)
	assert.Contains(t, serverStr, "UsersCreate(ctx context.Context")
	assert.Contains(t, serverStr, "UsersGet(ctx context.Context")
	assert.Contains(t, serverStr, "UsersList(ctx context.Context")
	assert.Contains(t, serverStr, "UsersUpdate(ctx context.Context")
	assert.Contains(t, serverStr, "UsersDelete(ctx context.Context")
}
