package duh_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedClientCompiles(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	protoDir := filepath.Join(tempDir, "proto/v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))

	protoStub := `syntax = "proto3";

package duh.api.v1;

message CreateRequest {}
message CreateResponse {}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.proto"), []byte(protoStub), 0644))

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
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(goProtoStub), 0644))

	goMod := `module github.com/example/test

go 1.24

require github.com/duh-rpc/duh.go/v2 v2.0.0
require github.com/kapetan-io/tackle v0.0.0
require google.golang.org/protobuf v0.0.0
`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644))

	cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go/v2=github.com/duh-rpc/duh.go/v2@v2.0.0")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd = exec.Command("go", "mod", "edit", "-replace", "github.com/kapetan-io/tackle=github.com/kapetan-io/tackle@v0.13.0")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
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

func TestClientIteratorIntegration(t *testing.T) {
	specPath, stdout := setupTest(t, specWithListOp)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)

	content := string(clientContent)
	assert.Contains(t, content, "duh.NewIterator")
	assert.Contains(t, content, "*duh.Iterator")
	assert.Contains(t, content, "UsersListIter")
	assert.NotContains(t, content, "UserPageFetcher")
	assert.NotContains(t, content, "FetchPage")

	_, err = os.Stat(filepath.Join(tempDir, "iterator.go"))
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	protoDir := filepath.Join(tempDir, "proto/v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))

	protoStub := `syntax = "proto3";

package duh.api.v1;

message ListRequest {}
message ListResponse {}
message User {}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.proto"), []byte(protoStub), 0644))

	goProtoStub := `package v1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

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
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(goProtoStub), 0644))

	goMod := `module github.com/example/test

go 1.24

require github.com/duh-rpc/duh.go/v2 v2.0.0
require github.com/kapetan-io/tackle v0.0.0
require google.golang.org/protobuf v0.0.0
`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644))

	cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go/v2=github.com/duh-rpc/duh.go/v2@v2.0.0")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd = exec.Command("go", "mod", "edit", "-replace", "github.com/kapetan-io/tackle=github.com/kapetan-io/tackle@v0.13.0")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
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

func TestClientWithoutIterator(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)

	content := string(clientContent)
	assert.NotContains(t, content, "PageFetcher")
	assert.NotContains(t, content, "Iter(")
	assert.NotContains(t, content, "FetchPage")

	_, err = os.Stat(filepath.Join(tempDir, "iterator.go"))
	require.True(t, os.IsNotExist(err))

	protoDir := filepath.Join(tempDir, "proto/v1")
	require.NoError(t, os.MkdirAll(protoDir, 0755))

	protoStub := `syntax = "proto3";

package duh.api.v1;

message CreateRequest {}
message CreateResponse {}
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.proto"), []byte(protoStub), 0644))

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
`
	require.NoError(t, os.WriteFile(filepath.Join(protoDir, "api.pb.go"), []byte(goProtoStub), 0644))

	goMod := `module github.com/example/test

go 1.24

require github.com/duh-rpc/duh.go/v2 v2.0.0
require github.com/kapetan-io/tackle v0.0.0
require google.golang.org/protobuf v0.0.0
`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644))

	cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/duh-rpc/duh.go/v2=github.com/duh-rpc/duh.go/v2@v2.0.0")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	cmd = exec.Command("go", "mod", "edit", "-replace", "github.com/kapetan-io/tackle=github.com/kapetan-io/tackle@v0.13.0")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
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

func TestClientStructure(t *testing.T) {
	specPath, stdout := setupTest(t, multiOpSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "client.go")

	_, err := os.Stat(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)

	content := string(clientContent)

	assert.Contains(t, content, "type ClientInterface interface")
	assert.Contains(t, content, "func NewClient")
	assert.Contains(t, content, "UsersCreate(ctx context.Context")
	assert.Contains(t, content, "UsersGet(ctx context.Context")
	assert.Contains(t, content, "UsersUpdate(ctx context.Context")
	assert.Contains(t, content, "Close(ctx context.Context) error")

	assert.Contains(t, content, "MaxConnsPerHost:     2_000")
	assert.Contains(t, content, "MaxIdleConns:        2_000")
	assert.Contains(t, content, "MaxIdleConnsPerHost: 2_000")
	assert.Contains(t, content, "MaxConnsPerHost:     5_000")
	assert.Contains(t, content, "MaxIdleConns:        5_000")
	assert.Contains(t, content, "MaxIdleConnsPerHost: 5_000")

	assert.Contains(t, content, "func WithTLS")
	assert.Contains(t, content, "func WithNoTLS")

	assert.Contains(t, content, "Code generated by 'duh generate'")
	assert.Contains(t, content, "DO NOT EDIT")
}
