package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

const ignoreExtension = "x-duh-lint-ignore"

func isOperationIgnored(op *v3.Operation, ruleName string) bool {
	if op == nil || op.Extensions == nil {
		return false
	}

	node, ok := op.Extensions.Get(ignoreExtension)
	if !ok || node == nil {
		return false
	}

	var rules []string
	if err := node.Decode(&rules); err != nil {
		return false
	}

	for _, r := range rules {
		if r == ruleName {
			return true
		}
	}
	return false
}

func isSchemaIgnored(schema *base.Schema, ruleName string) bool {
	if schema == nil || schema.Extensions == nil {
		return false
	}

	node, ok := schema.Extensions.Get(ignoreExtension)
	if !ok || node == nil {
		return false
	}

	var rules []string
	if err := node.Decode(&rules); err != nil {
		return false
	}

	for _, r := range rules {
		if r == ruleName {
			return true
		}
	}
	return false
}
