package duh

import (
	"fmt"

	schema "github.com/duh-rpc/openapi-schema.go"
)

type ProtoConverter interface {
	Convert(openapi []byte, packageName, packagePath string, nums *schema.FieldNumbers) ([]byte, error)
}

func NewProtoConverter() ProtoConverter {
	return &realProtoConverter{}
}

type realProtoConverter struct{}

func (r *realProtoConverter) Convert(openapi []byte, packageName, packagePath string, nums *schema.FieldNumbers) ([]byte, error) {
	result, err := schema.Convert(openapi, schema.ConvertOptions{
		PackageName:  packageName,
		PackagePath:  packagePath,
		FieldNumbers: nums,
	})
	if err != nil {
		return nil, err
	}

	// DUH specs cannot contain oneOf/allOf unions (enforced by lint, which generate
	// requires to pass), so the library should never emit Go union code. A non-empty
	// Golang output means a union slipped past lint.
	if len(result.Golang) > 0 {
		return nil, fmt.Errorf("unexpected Go output from conversion: a union schema (oneOf/allOf) slipped past lint")
	}

	return result.Protobuf, nil
}
