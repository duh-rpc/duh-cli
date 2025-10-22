package internal

import (
	"fmt"
	"io"
	"path/filepath"
)

// Print formats and outputs validation results
func Print(w io.Writer, result ValidationResult) {
	filename := filepath.Base(result.FilePath)

	if result.Valid() {
		_, _ = fmt.Fprintf(w, "✓ %s is DUH-RPC compliant\n", filename)
		return
	}

	_, _ = fmt.Fprintf(w, "Validating %s...\n", filename)
	_, _ = fmt.Fprintln(w, "ERRORS FOUND:")
	for _, violation := range result.Violations {
		_, _ = fmt.Fprintln(w, violation.String())
	}
	_, _ = fmt.Fprintf(w, "Summary: %d violations found in %s\n", len(result.Violations), filename)
}
