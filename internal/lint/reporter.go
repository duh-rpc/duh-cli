package lint

import (
	"fmt"
	"io"
	"path/filepath"
)

// Print formats and outputs validation results
func Print(w io.Writer, result ValidationResult) {
	filename := filepath.Base(result.FilePath)

	if len(result.Violations) == 0 {
		_, _ = fmt.Fprintf(w, "✓ %s is DUH-RPC compliant\n", filename)
		return
	}

	_, _ = fmt.Fprintf(w, "Validating %s...\n", filename)
	for _, violation := range result.Violations {
		_, _ = fmt.Fprintln(w, violation.String())
	}
	_, _ = fmt.Fprintf(w, "%d errors, %d warnings found in %s\n", result.ErrorCount(), result.WarningCount(), filename)
	if result.ErrorCount() == 0 {
		_, _ = fmt.Fprintf(w, "✓ %s is DUH-RPC compliant\n", filename)
	}
}
