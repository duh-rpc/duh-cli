package duh

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	yaml "go.yaml.in/yaml/v4"
)

// didYouMeanExtension declares, at the path-item level, the wrong-guess paths a
// caller is likely to try for this operation. Each becomes a teaching route in
// the generated switch that replies 404 naming the canonical path (ENG-133).
const didYouMeanExtension = "x-duh-did-you-mean"

type Parser struct {
	spec           *v3.Document
	config         *Config
	isFullTemplate bool
}

func NewParser(spec *v3.Document, config *Config, isFullTemplate bool) *Parser {
	return &Parser{
		spec:           spec,
		config:         config,
		isFullTemplate: isFullTemplate,
	}
}

func (p *Parser) Parse() (*TemplateData, error) {
	modulePath, err := p.config.DetectModulePath()
	if err != nil {
		return nil, err
	}

	operations, err := p.extractOperations()
	if err != nil {
		return nil, err
	}

	listOps, err := p.detectListOperations(operations)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	needsIO := false
	for _, op := range operations {
		if op.IsByteStream() {
			needsIO = true
			break
		}
	}

	return &TemplateData{
		PackageImport:  p.config.ConstructPackageImport(modulePath),
		Package:        p.config.PackageName,
		ModulePath:     modulePath,
		ProtoImport:    p.config.ConstructProtoImport(modulePath),
		ProtoPackage:   p.config.DeriveProtoPackage(),
		Operations:     operations,
		ListOps:        listOps,
		HasListOps:     len(listOps) > 0,
		NeedsIO:        needsIO,
		Timestamp:      timestamp,
		IsFullTemplate: p.isFullTemplate,
		GoModule:       modulePath,
	}, nil
}

// Standard DUH-RPC content types. The first two carry a structured (proto)
// message; the remaining three are streaming response content types. Anything
// else on a request/response body is a content endpoint (ENG-100).
const (
	contentTypeJSON          = "application/json"
	contentTypeProtobuf      = "application/protobuf"
	contentTypeOctetStream   = "application/octet-stream"
	contentTypeStreamJSON    = "application/duh-stream+json"
	contentTypeStreamProto   = "application/duh-stream+protobuf"
	errContentEndpointSuffix = "; see ENG-100"
)

func (p *Parser) extractOperations() ([]Operation, error) {
	var operations []Operation

	if p.spec.Paths == nil || p.spec.Paths.PathItems == nil {
		return operations, nil
	}

	base := p.serverBasePath()

	for pair := orderedmap.First(p.spec.Paths.PathItems); pair != nil; pair = pair.Next() {
		path := pair.Key()
		pathItem := pair.Value()

		if pathItem.Post == nil {
			continue
		}

		operation := pathItem.Post
		operationName, err := GenerateOperationName(path)
		if err != nil {
			continue
		}

		requestType, err := p.requestType(operation, path)
		if err != nil {
			return nil, err
		}
		if requestType == "" {
			continue
		}

		responseType, streamKind, err := p.responseShape(operation, path)
		if err != nil {
			return nil, err
		}
		if responseType == "" && streamKind == StreamNone {
			continue
		}

		summary := ""
		if operation.Summary != "" {
			summary = operation.Summary
		} else if operation.Description != "" {
			summary = operation.Description
		}

		operations = append(operations, Operation{
			IsInitTemplateMethod: p.isFullTemplate && isInitTemplateMethod(path),
			ConstName:            GenerateConstName(operationName),
			MethodName:           operationName,
			ResponseType:         responseType,
			RequestType:          requestType,
			StreamKind:           streamKind,
			DidYouMean:           resolveDidYouMean(base, pathItem),
			Summary:              summary,
			Path:                 path,
			RoutePath:            base + path,
		})
	}

	return operations, nil
}

// requestType returns the proto type for an operation's request body. DUH-RPC
// request bodies are always structured (application/json or application/protobuf
// referencing a named schema); streaming content types are response-only, and an
// opaque-content request body is a content endpoint (ENG-100). An operation with
// no request body returns "" and is skipped by the caller.
func (p *Parser) requestType(operation *v3.Operation, path string) (string, error) {
	if operation.RequestBody == nil || operation.RequestBody.Content == nil {
		return "", nil
	}

	// Scan every content type rather than returning on the first $ref so the request
	// body's structural type is derived independent of content-type declaration order.
	// application/json and application/protobuf co-declare the same message (the JSON
	// $ref and its application/protobuf wire twin {type:string,format:binary}); the
	// binary twin carries no message and is skipped for type derivation.
	unaryType := ""
	contentEndpoint := ""
	for pair := orderedmap.First(operation.RequestBody.Content); pair != nil; pair = pair.Next() {
		contentType := pair.Key()
		mediaType := pair.Value()
		switch contentType {
		case contentTypeJSON, contentTypeProtobuf:
			if mediaType.Schema != nil {
				if mediaType.Schema.IsReference() {
					if unaryType == "" {
						unaryType = "pb." + extractSchemaName(mediaType.Schema.GetReference())
					}
				} else if contentType == contentTypeJSON || isInlineStructured(mediaType.Schema) {
					return "", fmt.Errorf("inline schema not supported for request body in path %s", path)
				}
			}
		case contentTypeStreamJSON, contentTypeStreamProto:
			return "", fmt.Errorf("streaming content types are response-only; found %q on the request body in path %s", contentType, path)
		default:
			// octet-stream or a native MIME type: a content endpoint request body.
			contentEndpoint = contentType
		}
	}

	if unaryType != "" {
		return unaryType, nil
	}
	if contentEndpoint != "" {
		return "", fmt.Errorf("content endpoints not yet supported: request content type %q in path %s%s", contentEndpoint, path, errContentEndpointSuffix)
	}
	return "", nil
}

// isInlineStructured reports whether a media type's schema is an inline (non-$ref)
// object that declares properties — a genuine inline message the generator cannot
// name. A propertyless inline schema (such as the {type:string,format:binary} wire
// twin co-declared on application/protobuf) carries no message and is not structured,
// mirroring the SCHEMA_NO_INLINE_OBJECTS lint rule, which only flags inline schemas
// whose Properties.Len() > 0.
func isInlineStructured(schemaProxy *base.SchemaProxy) bool {
	if schemaProxy == nil || schemaProxy.IsReference() {
		return false
	}
	schema := schemaProxy.Schema()
	return schema != nil && schema.Properties != nil && schema.Properties.Len() > 0
}

// responseShape classifies an operation's success (2xx) response and returns the
// proto response type and the StreamKind. A streaming content type takes
// precedence over a co-declared structured type so a mixed response is treated as
// the stream it is. An opaque-content response is a content endpoint (ENG-100).
// An operation with no usable success response returns ("", StreamNone, nil) and
// is skipped by the caller.
func (p *Parser) responseShape(operation *v3.Operation, path string) (string, string, error) {
	if operation.Responses == nil || operation.Responses.Codes == nil {
		return "", StreamNone, nil
	}

	for pair := orderedmap.First(operation.Responses.Codes); pair != nil; pair = pair.Next() {
		if !isSuccessCode(pair.Key()) {
			continue
		}
		response := pair.Value()
		if response.Content == nil {
			continue
		}

		unaryType := ""
		contentEndpoint := ""
		for cp := orderedmap.First(response.Content); cp != nil; cp = cp.Next() {
			contentType := cp.Key()
			mediaType := cp.Value()
			switch contentType {
			case contentTypeOctetStream:
				return "", StreamBytes, nil
			case contentTypeStreamJSON, contentTypeStreamProto:
				if mediaType.Schema == nil || !mediaType.Schema.IsReference() {
					return "", StreamNone, fmt.Errorf("structured stream response in path %s must reference a named schema", path)
				}
				kind := StreamJSON
				if contentType == contentTypeStreamProto {
					kind = StreamProtobuf
				}
				return "pb." + extractSchemaName(mediaType.Schema.GetReference()), kind, nil
			case contentTypeJSON, contentTypeProtobuf:
				if mediaType.Schema != nil {
					if mediaType.Schema.IsReference() {
						if unaryType == "" {
							unaryType = "pb." + extractSchemaName(mediaType.Schema.GetReference())
						}
					} else if contentType == contentTypeJSON || isInlineStructured(mediaType.Schema) {
						// application/json is the required structural content type, so an
						// inline schema there is always an error. application/protobuf
						// co-declares the wire twin {type:string,format:binary}; a
						// propertyless binary schema carries no message and is skipped.
						return "", StreamNone, fmt.Errorf("inline schema not supported for response body in path %s", path)
					}
				}
			default:
				contentEndpoint = contentType
			}
		}

		if unaryType != "" {
			return unaryType, StreamNone, nil
		}
		if contentEndpoint != "" {
			return "", StreamNone, fmt.Errorf("content endpoints not yet supported: response content type %q in path %s%s", contentEndpoint, path, errContentEndpointSuffix)
		}
	}

	return "", StreamNone, nil
}

// isSuccessCode reports whether an OpenAPI response status code is a 2xx success.
// Streaming and content-type classification keys off the success response; error
// responses always carry a standard Reply and never define the endpoint shape.
func isSuccessCode(code string) bool {
	return len(code) == 3 && code[0] == '2'
}

// serverBasePath returns the path component of the first server URL, with any
// trailing slash trimmed. The DUH convention puts the version in
// servers[].url (e.g. https://api.example.com/v1) and keeps it out of the
// individual paths, so the served route is base + path (/v1/users.create).
// Specs without a server (or a server with no path) yield an empty base,
// leaving routes mounted at the root.
func (p *Parser) serverBasePath() string {
	if p.spec == nil || len(p.spec.Servers) == 0 || p.spec.Servers[0] == nil {
		return ""
	}
	u, err := url.Parse(p.spec.Servers[0].URL)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(u.Path, "/")
}

// didYouMeanNode returns the raw x-duh-did-you-mean sequence node declared on a
// path item, or (nil, false) when the extension is absent.
func didYouMeanNode(item *v3.PathItem) (*yaml.Node, bool) {
	if item == nil || item.Extensions == nil {
		return nil, false
	}
	node, ok := item.Extensions.Get(didYouMeanExtension)
	if !ok || node == nil {
		return nil, false
	}
	return node, true
}

// resolveDidYouMean returns the full teaching route paths (base + declared) for a
// path item. It trusts the entries have already passed ValidateDidYouMean and
// silently skips anything malformed, so a template never emits a broken case.
func resolveDidYouMean(base string, item *v3.PathItem) []string {
	node, ok := didYouMeanNode(item)
	if !ok || node.Kind != yaml.SequenceNode {
		return nil
	}
	var out []string
	for _, child := range node.Content {
		if child.Kind == yaml.ScalarNode && child.Tag == "!!str" && strings.HasPrefix(child.Value, "/") {
			out = append(out, base+child.Value)
		}
	}
	return out
}

// ValidateDidYouMean enforces the switch-path uniqueness invariant on the
// x-duh-did-you-mean extension before any code is generated: every entry must be
// a well-formed spec-relative path, must not equal a canonical route, and must be
// declared exactly once across the whole spec. All problems are collected so the
// author can fix the spec in one pass. Returns nil when no operation uses the
// extension, leaving generation byte-identical to a spec without it.
func ValidateDidYouMean(spec *v3.Document) error {
	if spec == nil || spec.Paths == nil || spec.Paths.PathItems == nil {
		return nil
	}

	canonical := make(map[string]bool)
	for pair := orderedmap.First(spec.Paths.PathItems); pair != nil; pair = pair.Next() {
		canonical[pair.Key()] = true
	}

	var problems []string
	declaredBy := make(map[string][]string)

	for pair := orderedmap.First(spec.Paths.PathItems); pair != nil; pair = pair.Next() {
		path := pair.Key()
		node, ok := didYouMeanNode(pair.Value())
		if !ok {
			continue
		}
		if node.Kind != yaml.SequenceNode {
			problems = append(problems, fmt.Sprintf("%s on path %q must be a list of strings, each beginning with '/'", didYouMeanExtension, path))
			continue
		}
		for _, child := range node.Content {
			if child.Kind != yaml.ScalarNode || child.Tag != "!!str" {
				problems = append(problems, fmt.Sprintf("%s entry on path %q is malformed: each entry must be a string beginning with '/'", didYouMeanExtension, path))
				continue
			}
			entry := child.Value
			if !strings.HasPrefix(entry, "/") {
				problems = append(problems, fmt.Sprintf("%s entry %q on path %q is malformed: must begin with '/'", didYouMeanExtension, entry, path))
				continue
			}
			if canonical[entry] {
				problems = append(problems, fmt.Sprintf("%s path %q on %q collides with the canonical route %q", didYouMeanExtension, entry, path, entry))
			}
			declaredBy[entry] = append(declaredBy[entry], path)
		}
	}

	for _, entry := range sortedKeys(declaredBy) {
		decls := declaredBy[entry]
		if len(decls) > 1 {
			// A within-list duplicate declares the same path item twice; name each
			// declaring path once so the message never reads "on X and X".
			seen := make(map[string]bool, len(decls))
			var quoted []string
			for _, d := range decls {
				if seen[d] {
					continue
				}
				seen[d] = true
				quoted = append(quoted, fmt.Sprintf("%q", d))
			}
			problems = append(problems, fmt.Sprintf("%s path %q is declared more than once (on %s)", didYouMeanExtension, entry, strings.Join(quoted, " and ")))
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return nil
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (p *Parser) detectListOperations(ops []Operation) ([]ListOperation, error) {
	var listOps []ListOperation

	for _, op := range ops {
		// Streaming endpoints carry a single payload type over many frames, not a
		// paginated collection; they are never list operations and their generated
		// methods do not use the (req, resp) iterator signature.
		if op.StreamKind != StreamNone {
			continue
		}

		requestSchema, responseSchema, err := p.getSchemas(op.Path)
		if err != nil {
			continue
		}

		if p.isListOperation(op.Path, requestSchema, responseSchema) {
			fieldName, itemType, found := p.findFirstArrayField(responseSchema)
			if !found {
				continue
			}

			listOps = append(listOps, ListOperation{
				Operation:     op,
				IteratorName:  op.MethodName + "Iter",
				ItemType:      "*pb." + itemType,
				ResponseField: fieldName,
			})
		}
	}

	return listOps, nil
}

func (p *Parser) isListOperation(path string, requestSchema, responseSchema *base.SchemaProxy) bool {
	_, method, err := parseSubjectMethod(path)
	if err != nil {
		return false
	}

	if !strings.Contains(strings.ToLower(method), "list") {
		return false
	}

	if requestSchema == nil || requestSchema.Schema() == nil {
		return false
	}

	hasPage := false
	if requestSchema.Schema().Properties != nil {
		for propPair := orderedmap.First(requestSchema.Schema().Properties); propPair != nil; propPair = propPair.Next() {
			if strings.ToLower(propPair.Key()) == "pagination" {
				hasPage = true
				break
			}
		}
	}

	if !hasPage {
		return false
	}

	_, _, found := p.findFirstArrayField(responseSchema)
	return found
}

func (p *Parser) findFirstArrayField(schema *base.SchemaProxy) (fieldName, itemType string, found bool) {
	if schema == nil || schema.Schema() == nil {
		return "", "", false
	}

	schemaObj := schema.Schema()
	if schemaObj.Properties == nil {
		return "", "", false
	}

	for propPair := orderedmap.First(schemaObj.Properties); propPair != nil; propPair = propPair.Next() {
		propName := propPair.Key()
		propSchema := propPair.Value()

		if propSchema.Schema() != nil && propSchema.Schema().Type != nil {
			if len(propSchema.Schema().Type) > 0 && propSchema.Schema().Type[0] == "array" {
				if propSchema.Schema().Items != nil && propSchema.Schema().Items.IsA() {
					itemSchema := propSchema.Schema().Items.A
					if itemSchema.IsReference() {
						ref := itemSchema.GetReference()
						itemType = extractSchemaName(ref)
						capitalizedFieldName := capitalizeFirst(propName)
						return capitalizedFieldName, itemType, true
					}
				}
			}
		}
	}

	return "", "", false
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (p *Parser) getSchemas(path string) (*base.SchemaProxy, *base.SchemaProxy, error) {
	if p.spec.Paths == nil || p.spec.Paths.PathItems == nil {
		return nil, nil, fmt.Errorf("no paths found")
	}

	pathItem := p.spec.Paths.PathItems.GetOrZero(path)
	if pathItem == nil || pathItem.Post == nil {
		return nil, nil, fmt.Errorf("path not found: %s", path)
	}

	operation := pathItem.Post

	var requestSchema *base.SchemaProxy
	if operation.RequestBody != nil && operation.RequestBody.Content != nil {
		for contentPair := orderedmap.First(operation.RequestBody.Content); contentPair != nil; contentPair = contentPair.Next() {
			mediaType := contentPair.Value()
			if mediaType.Schema != nil {
				if mediaType.Schema.IsReference() {
					ref := mediaType.Schema.GetReference()
					requestSchema = p.resolveSchemaRef(ref)
				}
				break
			}
		}
	}

	var responseSchema *base.SchemaProxy
	if operation.Responses != nil && operation.Responses.Codes != nil {
		for responsePair := orderedmap.First(operation.Responses.Codes); responsePair != nil; responsePair = responsePair.Next() {
			response := responsePair.Value()
			if response.Content != nil {
				for contentPair := orderedmap.First(response.Content); contentPair != nil; contentPair = contentPair.Next() {
					mediaType := contentPair.Value()
					if mediaType.Schema != nil {
						if mediaType.Schema.IsReference() {
							ref := mediaType.Schema.GetReference()
							responseSchema = p.resolveSchemaRef(ref)
						}
						break
					}
				}
			}
			if responseSchema != nil {
				break
			}
		}
	}

	return requestSchema, responseSchema, nil
}

func (p *Parser) resolveSchemaRef(ref string) *base.SchemaProxy {
	schemaName := extractSchemaName(ref)
	if p.spec.Components != nil && p.spec.Components.Schemas != nil {
		return p.spec.Components.Schemas.GetOrZero(schemaName)
	}
	return nil
}

func extractSchemaName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func isInitTemplateMethod(path string) bool {
	initTemplatePaths := []string{
		"/users.create",
		"/users.get",
		"/users.list",
		"/users.update",
	}

	for _, templatePath := range initTemplatePaths {
		if path == templatePath {
			return true
		}
	}
	return false
}
