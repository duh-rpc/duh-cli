package rules

import (
	"fmt"

	"github.com/duh-rpc/duhrpc-lint/internal/types"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// HTTPMethodRule validates only POST is used
type HTTPMethodRule struct{}

func NewHTTPMethodRule() *HTTPMethodRule {
	return &HTTPMethodRule{}
}

func (r *HTTPMethodRule) Name() string {
	return "http-method"
}

func (r *HTTPMethodRule) Validate(doc *v3.Document) []types.Violation {
	var violations []types.Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		// Check each HTTP method
		methods := []struct {
			name      string
			operation *v3.Operation
		}{
			{"GET", pathItem.Get},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
			{"HEAD", pathItem.Head},
			{"OPTIONS", pathItem.Options},
			{"TRACE", pathItem.Trace},
		}

		for _, method := range methods {
			if method.operation != nil {
				violations = append(violations, types.Violation{
					RuleName:   r.Name(),
					Location:   fmt.Sprintf("%s %s", method.name, path),
					Message:    fmt.Sprintf("HTTP method %s is not allowed in DUH-RPC", method.name),
					Suggestion: "Use POST method for all DUH-RPC operations",
				})
			}
		}
	}

	return violations
}
