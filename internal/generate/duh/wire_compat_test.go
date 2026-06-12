package duh_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/require"
)

// TestStyleBWireCompatibility is the standing regression guard for the style-B
// oneOf wire contract (ENG-63). It re-points the one-off proof harness
// (docs/oneof-wire-proof/) at GENERATED output: it runs `duh generate` on a
// style-B spec, compiles the emitted .proto with the real protoc/protoc-gen-go
// toolchain, and runs a protojson/proto round-trip against the generated Go.
//
// The inner suite (styleBWireInnerTest) asserts the contract the blueprint
// pins (docs/features/wire-compatible-oneof/blueprint.md):
//   - the generated oneof marshals to the nested {"<variant>": {...}} shape with
//     snake_case keys preserved (json_name), not lowerCamelCase;
//   - those bytes are semantically equal to the optional-fields encoding of the
//     same value;
//   - a client built from the oneof form reads bytes from the optional-fields
//     form and vice versa, on both the JSON and binary wires;
//   - the real oneof rejects two-variant input (exactly-one enforcement).
//
// Marshaling regressions in duh.go (a change to protojson options, or dropping
// the json_name annotation) break this test. Requires the protobuf toolchain and
// network for `go mod tidy`; it skips when protoc or protoc-gen-go is absent.
func TestStyleBWireCompatibility(t *testing.T) {
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skip("protoc not installed")
	}
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		t.Skip("protoc-gen-go not installed")
	}

	specPath, stdout := setupTest(t, styleBValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})
	require.Equal(t, 0, exitCode)

	// The optional-fields twin of the generated oneof GetResponse: same variant
	// fields, numbers, and json_names, without a oneof group. It is the baseline
	// the round-trip proves the generated oneof is wire-equal to.
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "proto/v1/baseline.proto"), []byte(styleBBaselineProto), 0644))

	protoc := exec.Command("protoc", "-I", ".",
		"--go_out=.", "--go_opt=module=github.com/example/test",
		"proto/v1/api.proto", "proto/v1/baseline.proto")
	protoc.Dir = tempDir
	output, err := protoc.CombinedOutput()
	require.NoError(t, err, string(output))

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "wire"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "wire/wire_test.go"), []byte(styleBWireInnerTest), 0644))

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tempDir
	output, err = tidy.CombinedOutput()
	require.NoError(t, err, string(output))

	test := exec.Command("go", "test", "./wire/...")
	test.Dir = tempDir
	output, err = test.CombinedOutput()
	require.NoError(t, err, string(output))
}

const styleBBaselineProto = `syntax = "proto3";

package duh.api.v1;

option go_package = "github.com/example/test/proto/v1";

import "proto/v1/api.proto";

// OptionalGetResponse is the optional-fields twin of the generated oneof
// GetResponse: the same variant fields, numbers, and json_names, without a oneof
// group. The round-trip proves the generated oneof is wire-equal to this form.
message OptionalGetResponse {
  Cat cat_event = 1 [json_name = "cat_event"];
  Dog dog_event = 2 [json_name = "dog_event"];
}
`

// styleBWireInnerTest is compiled and run against the GENERATED protobuf code by
// the parent test. JSON literals use escaped double quotes so this file can be
// embedded as a Go raw-string constant.
const styleBWireInnerTest = `package wire_test

import (
	stdjson "encoding/json"
	"testing"

	apiv1 "github.com/example/test/proto/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	json "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// canon parses JSON into a normalized value so comparisons ignore the
// insignificant whitespace protojson deliberately injects; wire compatibility
// means SEMANTIC JSON equality, which is what any JSON consumer sees.
func canon(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	require.NoError(t, stdjson.Unmarshal(b, &v))
	return v
}

// TestStyleBOneofJSONWireShape proves the generated oneof marshals to the nested,
// key-tagged shape with snake_case preserved, identical to the optional form.
func TestStyleBOneofJSONWireShape(t *testing.T) {
	oneof := &apiv1.GetResponse{GetResponse: &apiv1.GetResponse_CatEvent{CatEvent: &apiv1.Cat{PetName: "Whiskers"}}}
	optional := &apiv1.OptionalGetResponse{CatEvent: &apiv1.Cat{PetName: "Whiskers"}}

	oneofJSON, err := json.Marshal(oneof)
	require.NoError(t, err)
	optionalJSON, err := json.Marshal(optional)
	require.NoError(t, err)

	assert.Equal(t, canon(t, oneofJSON), canon(t, optionalJSON))
	assert.Equal(t, canon(t, []byte("{\"cat_event\":{\"pet_name\":\"Whiskers\"}}")), canon(t, oneofJSON))
}

// TestStyleBCrossUnmarshalJSON proves interop on the JSON wire: bytes from the
// oneof form are readable by the optional form and vice versa.
func TestStyleBCrossUnmarshalJSON(t *testing.T) {
	serverBytes, err := json.Marshal(&apiv1.GetResponse{GetResponse: &apiv1.GetResponse_CatEvent{CatEvent: &apiv1.Cat{PetName: "Whiskers"}}})
	require.NoError(t, err)

	var asOptional apiv1.OptionalGetResponse
	require.NoError(t, json.Unmarshal(serverBytes, &asOptional))
	require.NotNil(t, asOptional.CatEvent)
	assert.Equal(t, "Whiskers", asOptional.CatEvent.PetName)
	assert.Nil(t, asOptional.DogEvent)

	optionalBytes, err := json.Marshal(&apiv1.OptionalGetResponse{CatEvent: &apiv1.Cat{PetName: "Whiskers"}})
	require.NoError(t, err)

	var asOneof apiv1.GetResponse
	require.NoError(t, json.Unmarshal(optionalBytes, &asOneof))
	require.NotNil(t, asOneof.GetCatEvent())
	assert.Equal(t, "Whiskers", asOneof.GetCatEvent().PetName)
}

// TestStyleBBinaryWireCompatible proves the protobuf BINARY wire is identical
// too: field numbers and types match between the oneof and optional forms.
func TestStyleBBinaryWireCompatible(t *testing.T) {
	oneofBytes, err := proto.Marshal(&apiv1.GetResponse{GetResponse: &apiv1.GetResponse_CatEvent{CatEvent: &apiv1.Cat{PetName: "Whiskers"}}})
	require.NoError(t, err)

	var asOptional apiv1.OptionalGetResponse
	require.NoError(t, proto.Unmarshal(oneofBytes, &asOptional))
	require.NotNil(t, asOptional.CatEvent)
	assert.Equal(t, "Whiskers", asOptional.CatEvent.PetName)

	optionalBytes, err := proto.Marshal(&apiv1.OptionalGetResponse{CatEvent: &apiv1.Cat{PetName: "Whiskers"}})
	require.NoError(t, err)
	assert.Equal(t, oneofBytes, optionalBytes)
}

// TestStyleBOneofRejectsTwoVariants proves the semantic difference that justifies
// generating a real oneof: it REJECTS two-variant JSON (exactly-one), while the
// optional form accepts it. Valid (exactly-one) values are wire-identical.
func TestStyleBOneofRejectsTwoVariants(t *testing.T) {
	bothSet := []byte("{\"cat_event\":{\"pet_name\":\"Whiskers\"},\"dog_event\":{\"pet_name\":\"Rex\"}}")

	var asOneof apiv1.GetResponse
	require.Error(t, json.Unmarshal(bothSet, &asOneof))

	var asOptional apiv1.OptionalGetResponse
	require.NoError(t, json.Unmarshal(bothSet, &asOptional))
	assert.NotNil(t, asOptional.CatEvent)
	assert.NotNil(t, asOptional.DogEvent)
}
`
