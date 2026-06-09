package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExportConfig configures a one-shot export of a spec to a self-contained HTML file.
type ExportConfig struct {
	Writer   io.Writer
	SpecPath string
	Output   string
}

// Export renders the spec at cfg.SpecPath into a single self-contained HTML file
// at cfg.Output. The spec is parsed first; an unparseable spec is rejected before
// the output is touched (AC #3). The HTML is written to a temp file in the
// destination directory and atomically renamed into place, so a failed or
// interrupted write never leaves a truncated file and never clobbers a previously
// good output (INV-2, BC-3, AC #4).
func Export(cfg ExportConfig) error {
	data, err := os.ReadFile(cfg.SpecPath)
	if err != nil {
		return fmt.Errorf("failed to read spec: %w", err)
	}
	if err := parseSpec(data); err != nil {
		return err
	}

	html := buildExport(data)

	dir := filepath.Dir(cfg.Output)
	tmp, err := os.CreateTemp(dir, ".docs-*.html.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(html); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write export: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to write export: %w", err)
	}

	if err := os.Rename(tmpName, cfg.Output); err != nil {
		return fmt.Errorf("failed to finalize export: %w", err)
	}

	_, _ = fmt.Fprintf(cfg.Writer, "✓ Exported API documentation to %s\n", cfg.Output)
	return nil
}
