package proof_test

import (
	stdjson "encoding/json"
	"reflect"
	"testing"

	"proof/gen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	json "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// canon parses a JSON document into a normalized value so comparisons ignore
// insignificant whitespace. This matters because protojson.Marshal deliberately
// injects random spaces to discourage byte-stability assumptions — so "wire
// compatible" means SEMANTIC JSON equality, which is exactly what any JSON
// parser (the actual wire consumer) sees.
func canon(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	require.NoError(t, stdjson.Unmarshal(b, &v))
	return v
}

// TestOneofAndOptionalProduceIdenticalWire proves the proto `oneof` form and the
// optional-fields form serialize to the SAME JSON on the wire, with snake_case
// preserved, in the nested `{"cat_event": {...}}` shape.
func TestOneofAndOptionalProduceIdenticalWire(t *testing.T) {
	// cat variant set, multi-word snake_case payload field
	oneof := &gen.OneofEvent{Kind: &gen.OneofEvent_CatEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}}}
	optional := &gen.OptionalEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}}

	oneofJSON, err := json.Marshal(oneof)
	require.NoError(t, err)
	optionalJSON, err := json.Marshal(optional)
	require.NoError(t, err)

	t.Logf("oneof form    -> %s", oneofJSON)
	t.Logf("optional form -> %s", optionalJSON)

	// Semantic equality: identical JSON object on the wire.
	assert.Equal(t, canon(t, oneofJSON), canon(t, optionalJSON))

	// Exact expected shape: NESTED key-tagged, snake_case preserved.
	assert.Equal(t, canon(t, []byte(`{"cat_event":{"pet_name":"Whiskers"}}`)), canon(t, oneofJSON))
	assert.Equal(t, canon(t, []byte(`{"cat_event":{"pet_name":"Whiskers"}}`)), canon(t, optionalJSON))
}

// TestDogVariantWire confirms the other variant behaves identically.
func TestDogVariantWire(t *testing.T) {
	oneof := &gen.OneofEvent{Kind: &gen.OneofEvent_DogEvent{DogEvent: &gen.Dog{PetName: "Rex"}}}
	optional := &gen.OptionalEvent{DogEvent: &gen.Dog{PetName: "Rex"}}

	oneofJSON, err := json.Marshal(oneof)
	require.NoError(t, err)
	optionalJSON, err := json.Marshal(optional)
	require.NoError(t, err)

	assert.Equal(t, canon(t, oneofJSON), canon(t, optionalJSON))
	assert.Equal(t, canon(t, []byte(`{"dog_event":{"pet_name":"Rex"}}`)), canon(t, oneofJSON))
}

// TestCrossUnmarshal proves true interop: bytes produced by a server using the
// `oneof` form are readable by a client using the optional-fields form, and
// vice versa. This is the operative definition of wire compatibility.
func TestCrossUnmarshal(t *testing.T) {
	// Server emits using the oneof form.
	serverBytes, err := json.Marshal(&gen.OneofEvent{
		Kind: &gen.OneofEvent_CatEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}},
	})
	require.NoError(t, err)

	// Client built from the optional-fields proto reads it.
	var asOptional gen.OptionalEvent
	require.NoError(t, json.Unmarshal(serverBytes, &asOptional))
	require.NotNil(t, asOptional.CatEvent)
	assert.Equal(t, "Whiskers", asOptional.CatEvent.PetName)
	assert.Nil(t, asOptional.DogEvent)

	// Reverse: optional-fields server -> oneof client.
	optionalBytes, err := json.Marshal(&gen.OptionalEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}})
	require.NoError(t, err)

	var asOneof gen.OneofEvent
	require.NoError(t, json.Unmarshal(optionalBytes, &asOneof))
	require.NotNil(t, asOneof.GetCatEvent())
	assert.Equal(t, "Whiskers", asOneof.GetCatEvent().PetName)
}

// TestGenericJSONConsumerInterop proves a NON-protobuf consumer (a plain Go
// struct using encoding/json, standing in for any OpenAPI-driven client) reads
// and writes the exact same shape. This is the OpenAPI side of the wire.
func TestGenericJSONConsumerInterop(t *testing.T) {
	type cat struct {
		PetName string `json:"pet_name"`
	}
	// Models the OpenAPI optional-properties object: one optional field per variant.
	type event struct {
		CatEvent *cat `json:"cat_event,omitempty"`
		DogEvent *cat `json:"dog_event,omitempty"`
	}

	// protobuf oneof server -> generic JSON client
	serverBytes, err := json.Marshal(&gen.OneofEvent{
		Kind: &gen.OneofEvent_CatEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}},
	})
	require.NoError(t, err)

	var got event
	require.NoError(t, stdjson.Unmarshal(serverBytes, &got))
	require.NotNil(t, got.CatEvent)
	assert.Equal(t, "Whiskers", got.CatEvent.PetName)
	assert.Nil(t, got.DogEvent)

	// generic JSON client -> protobuf oneof server
	clientBytes, err := stdjson.Marshal(event{CatEvent: &cat{PetName: "Whiskers"}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"cat_event":{"pet_name":"Whiskers"}}`, string(clientBytes))

	var back gen.OneofEvent
	require.NoError(t, json.Unmarshal(clientBytes, &back))
	require.NotNil(t, back.GetCatEvent())
	assert.Equal(t, "Whiskers", back.GetCatEvent().PetName)
}

// TestProtoOneofEnforcesExactlyOne shows the semantic difference that decides
// WHICH proto encoding to generate. A real proto `oneof` REJECTS JSON with two
// variants set (enforcing "exactly one" — matching OpenAPI oneOf semantics),
// while the optional-fields form accepts it. For all VALID (exactly-one) values
// the wire is identical; they only diverge on invalid input.
func TestProtoOneofEnforcesExactlyOne(t *testing.T) {
	bothSet := []byte(`{"cat_event":{"pet_name":"Whiskers"},"dog_event":{"pet_name":"Rex"}}`)

	// proto oneof: rejects multi-variant input (this is the exactly-one guard).
	var asOneof gen.OneofEvent
	err := json.Unmarshal(bothSet, &asOneof)
	require.Error(t, err)
	t.Logf("oneof form rejects two-variant input: %v", err)

	// optional fields: accepts both (no exactly-one enforcement at proto level).
	var asOptional gen.OptionalEvent
	require.NoError(t, json.Unmarshal(bothSet, &asOptional))
	assert.NotNil(t, asOptional.CatEvent)
	assert.NotNil(t, asOptional.DogEvent)
}

// TestNeitherSetWire confirms the empty case is also identical across forms.
func TestNeitherSetWire(t *testing.T) {
	oneofJSON, err := json.Marshal(&gen.OneofEvent{})
	require.NoError(t, err)
	optionalJSON, err := json.Marshal(&gen.OptionalEvent{})
	require.NoError(t, err)

	assert.Equal(t, canon(t, []byte(`{}`)), canon(t, oneofJSON))
	assert.Equal(t, canon(t, []byte(`{}`)), canon(t, optionalJSON))
}

// TestBinaryWireAlsoCompatible is a bonus: the protobuf BINARY wire (not just
// JSON) is identical too, because field numbers/types match between the oneof
// and optional forms. A oneof is binary-wire-identical to separate optional
// fields with the same numbers.
func TestBinaryWireAlsoCompatible(t *testing.T) {
	oneofBytes, err := proto.Marshal(&gen.OneofEvent{
		Kind: &gen.OneofEvent_CatEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}},
	})
	require.NoError(t, err)

	var asOptional gen.OptionalEvent
	require.NoError(t, proto.Unmarshal(oneofBytes, &asOptional))
	require.NotNil(t, asOptional.CatEvent)
	assert.Equal(t, "Whiskers", asOptional.CatEvent.PetName)

	optionalBytes, err := proto.Marshal(&gen.OptionalEvent{CatEvent: &gen.Cat{PetName: "Whiskers"}})
	require.NoError(t, err)
	assert.Equal(t, oneofBytes, optionalBytes)

	if reflect.DeepEqual(oneofBytes, optionalBytes) {
		t.Logf("binary wire identical: % x", oneofBytes)
	}
}
