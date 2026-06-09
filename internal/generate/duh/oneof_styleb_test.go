package duh_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const styleBValidSpec = `openapi: "3.0.3"
info:
  title: Style B Union API
  version: 1.0.0
  description: A DUH-RPC spec exercising the wire-compatible style-B oneOf
servers:
  - url: https://api.example.com/v1
paths:
  /events.get:
    post:
      operationId: getEvent
      description: Get a single event, which is exactly one of the known variants
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    GetRequest:
      description: Request a single event by identifier
      type: object
      required: [id]
      properties:
        id:
          description: Unique identifier of the event to fetch
          type: string
    GetResponse:
      description: An event carrying exactly one variant payload
      type: object
      properties:
        cat_event:
          $ref: '#/components/schemas/Cat'
        dog_event:
          $ref: '#/components/schemas/Dog'
      oneOf:
        - required: [cat_event]
        - required: [dog_event]
    Cat:
      description: A cat event payload
      type: object
      required: [pet_name]
      properties:
        pet_name:
          description: Name of the cat
          type: string
    Dog:
      description: A dog event payload
      type: object
      required: [pet_name]
      properties:
        pet_name:
          description: Name of the dog
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          description: Human-readable error message
          type: string
`

// TestGenerateStyleBEmitsProtoOneof verifies that generating from a valid style-B
// spec emits a .proto whose message contains a oneof group with one field per
// variant, each annotated [json_name = "<property>"], and writes fieldmap.lock with a
// stable number per variant field. (Blueprint AC 2.)
func TestGenerateStyleBEmitsProtoOneof(t *testing.T) {
	specPath, stdout := setupTest(t, styleBValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)

	protoContent, err := os.ReadFile(filepath.Join(tempDir, "proto/v1/api.proto"))
	require.NoError(t, err)
	proto := string(protoContent)

	assert.Contains(t, proto, "oneof")
	assert.Contains(t, proto, `Cat cat_event = 1 [json_name = "cat_event"]`)
	assert.Contains(t, proto, `Dog dog_event = 2 [json_name = "dog_event"]`)

	// The variant fields live inside a oneof block within GetResponse.
	response := extractStyleBMessage(t, proto, "GetResponse")
	assert.Contains(t, response, "oneof")
	oneofBody := response[strings.Index(response, "oneof"):]
	assert.Contains(t, oneofBody, "cat_event")
	assert.Contains(t, oneofBody, "dog_event")

	lockContent, err := os.ReadFile(filepath.Join(tempDir, "fieldmap.lock"))
	require.NoError(t, err)
	lock := string(lockContent)
	assert.Contains(t, lock, "GetResponse")
	assert.Contains(t, lock, "cat_event")
	assert.Contains(t, lock, "dog_event")
}

// extractStyleBMessage returns the text of `message <name> { ... }` including nested braces.
func extractStyleBMessage(t *testing.T, proto, name string) string {
	t.Helper()
	marker := "message " + name + " {"
	start := strings.Index(proto, marker)
	require.GreaterOrEqual(t, start, 0)
	depth := 0
	for i := start; i < len(proto); i++ {
		switch proto[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return proto[start : i+1]
			}
		}
	}
	t.Fatalf("unbalanced braces for message %s", name)
	return ""
}
