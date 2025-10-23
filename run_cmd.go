package duhrpc

import (
	"fmt"
	"io"

	"github.com/duh-rpc/duhrpc/internal/lint"
	"github.com/spf13/cobra"
)

const Version = "1.0.0"

// RunCmd executes the CLI logic and returns exit code
func RunCmd(stdout io.Writer, args []string) int {
	exitCode := 0

	rootCmd := &cobra.Command{
		Use:   "duhrpc",
		Short: "DUH-RPC tooling",
		Long:  `duhrpc is a command-line tool for working with DUH-RPC specifications and code.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("duhrpc version {{.Version}}\n")

	lintCmd := &cobra.Command{
		Use:   "lint <openapi-file>",
		Short: "Validate OpenAPI specs for DUH-RPC compliance",
		Long: `Validate OpenAPI specs for DUH-RPC compliance.

The lint command checks OpenAPI 3.0 specifications against DUH-RPC requirements
and reports any violations found.

Exit Codes:
  0    Validation passed (spec is DUH-RPC compliant)
  1    Validation failed (violations found)
  2    Error (file not found, parse error, etc.)`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]

			doc, err := lint.Load(filePath)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}

			result := lint.Validate(doc, filePath)
			lint.Print(cmd.OutOrStdout(), result)

			if result.Valid() {
				exitCode = 0
			} else {
				exitCode = 1
			}
		},
	}

	rootCmd.AddCommand(lintCmd)
	rootCmd.SetOutput(stdout)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return 2
	}

	return exitCode
}
