package duh

import "io"

type RunConfig struct {
	Writer       io.Writer
	SpecPath     string
	PackageName  string
	OutputDir    string
	ProtoPath    string
	ProtoImport  string
	ProtoPackage string
	LockPath     string
	FullFlag     bool
	Converter    ProtoConverter
}

type TemplateData struct {
	PackageImport  string
	Package        string
	ModulePath     string
	ProtoImport    string
	ProtoPackage   string
	Operations     []Operation
	ListOps        []ListOperation
	HasListOps     bool
	NeedsIO        bool
	Timestamp      string
	IsFullTemplate bool
	GoModule       string
}

// StreamKind classifies an operation's response shape. The empty value is a
// normal unary (typed) response; the others are the three DUH-RPC streaming
// response content types. See docs (duh.go/docs/streaming.md).
const (
	StreamNone     = ""                // application/json | application/protobuf
	StreamBytes    = "bytes"           // application/octet-stream
	StreamJSON     = "stream-json"     // application/duh-stream+json
	StreamProtobuf = "stream-protobuf" // application/duh-stream+protobuf
)

type Operation struct {
	MethodName           string
	Path                 string
	RoutePath            string
	ConstName            string
	Summary              string
	RequestType          string
	ResponseType         string
	StreamKind           string
	IsInitTemplateMethod bool
}

// IsByteStream reports whether the operation streams unstructured bytes
// (application/octet-stream).
func (o Operation) IsByteStream() bool { return o.StreamKind == StreamBytes }

// IsStructStream reports whether the operation is a structured stream
// (application/duh-stream+json or application/duh-stream+protobuf).
func (o Operation) IsStructStream() bool {
	return o.StreamKind == StreamJSON || o.StreamKind == StreamProtobuf
}

// AcceptConst returns the duh runtime constant for the Accept header a client
// must send for a streaming response, or "" for a unary operation.
func (o Operation) AcceptConst() string {
	switch o.StreamKind {
	case StreamBytes:
		return "duh.ContentOctetStream"
	case StreamJSON:
		return "duh.ContentStreamJSON"
	case StreamProtobuf:
		return "duh.ContentStreamProtoBuf"
	default:
		return ""
	}
}

type ListOperation struct {
	Operation
	IteratorName  string
	ItemType      string
	ResponseField string
}
