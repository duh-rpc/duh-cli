package docs

import (
	"fmt"

	"github.com/pb33f/libopenapi"
)

// parseSpec validates that data parses as an OpenAPI document. It mirrors the
// loader used by lint so docs accepts exactly the documents lint can read,
// without imposing any DUH-RPC compliance check (ADR-0009).
func parseSpec(data []byte) error {
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}
	if _, err := doc.BuildV3Model(); err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}
	return nil
}
