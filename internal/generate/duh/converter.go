package duh

import (
	conv "github.com/duh-rpc/openapi-proto.go"
)

type ProtoConverter interface {
	Convert(openapi []byte, packageName string) ([]byte, error)
}

func NewProtoConverter() ProtoConverter {
	return &realProtoConverter{}
}

type realProtoConverter struct{}

func (r *realProtoConverter) Convert(openapi []byte, packageName string) ([]byte, error) {
	return conv.Convert(openapi, packageName)
}
