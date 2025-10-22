package lint

import (
	"flag"
	"fmt"
	"io"

	"github.com/duh-rpc/duhrpc-lint/internal"
)

const Version = "1.0.0"

const helpText = `duhrpc-lint - Validate OpenAPI specs for DUH-RPC compliance

Usage:
  duhrpc-lint <openapi-file>
  duhrpc-lint --help
  duhrpc-lint --version

Arguments:
  <openapi-file>    Path to OpenAPI 3.0 YAML file

Options:
  --help            Show this help message
  --version         Show version information

Exit Codes:
  0    Validation passed (spec is DUH-RPC compliant)
  1    Validation failed (violations found)
  2    Error (file not found, parse error, etc.)
`

// RunCmd executes the CLI logic and returns exit code
func RunCmd(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("duhrpc-lint", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var showHelp bool
	var showVersion bool
	fs.BoolVar(&showHelp, "help", false, "Show help message")
	fs.BoolVar(&showVersion, "version", false, "Show version information")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if showHelp {
		_, _ = fmt.Fprint(stdout, helpText)
		return 0
	}

	if showVersion {
		_, _ = fmt.Fprintf(stdout, "duhrpc-lint version %s\n", Version)
		return 0
	}

	if fs.NArg() != 1 {
		_, _ = fmt.Fprintln(stderr, "Error: Exactly one OpenAPI file path is required")
		_, _ = fmt.Fprintln(stderr, "Use --help for usage information")
		return 2
	}

	filePath := fs.Arg(0)

	doc, err := internal.Load(filePath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 2
	}

	result := internal.Validate(doc, filePath)
	internal.Print(stdout, result)

	if result.Valid() {
		return 0
	}
	return 1
}
