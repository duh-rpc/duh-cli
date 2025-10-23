package init

import (
	"fmt"
	"io"
)

func Run(w io.Writer, outputPath string) error {
	if err := writeFile(outputPath, Template); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "✓ Created DUH-RPC compliant OpenAPI spec at %s\n", outputPath)
	return nil
}
