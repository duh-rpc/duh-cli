package duh

import (
	"fmt"
	"io"

	"github.com/duh-rpc/duh-cli/internal/add"
	init_ "github.com/duh-rpc/duh-cli/internal/init"
	"github.com/duh-rpc/duh-cli/internal/lint"
	"github.com/spf13/cobra"
)

const Version = "1.0.0"

// RunCmd executes the CLI logic and returns exit code
func RunCmd(stdout io.Writer, args []string) int {
	exitCode := 0

	rootCmd := &cobra.Command{
		Use:   "duh",
		Short: "DUH-RPC tooling",
		Long:  `duh is a command-line tool for working with DUH-RPC specifications and code.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("duh version {{.Version}}\n")

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

	initCmd := &cobra.Command{
		Use:   "init [openapi-file]",
		Short: "Create a DUH-RPC compliant OpenAPI specification template",
		Long: `Create a DUH-RPC compliant OpenAPI specification template.

The init command generates a comprehensive example OpenAPI 3.0 specification
that demonstrates all DUH-RPC requirements and best practices.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.

Exit Codes:
  0    Template created successfully
  2    Error (file already exists, permission denied, etc.)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			const defaultOutput = "openapi.yaml"
			outputPath := defaultOutput
			if len(args) > 0 {
				outputPath = args[0]
			}

			if err := init_.Run(cmd.OutOrStdout(), outputPath); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <path> <name>",
		Short: "Add a new DUH-RPC endpoint to an OpenAPI specification",
		Long: `Add a new DUH-RPC endpoint to an OpenAPI specification.

The add command creates a new endpoint with the specified path and name,
generating request and response schemas with placeholder properties.

The path must follow DUH-RPC format: /v{N}/{subject}.{method}
For example: /v1/users.create

The name is used to generate schema names: {Name}Request and {Name}Response
For example: CreateUser generates CreateUserRequest and CreateUserResponse

Use the -f flag to specify a custom OpenAPI file (defaults to 'openapi.yaml').

Exit Codes:
  0    Endpoint added successfully
  2    Error (invalid path, file not found, path already exists, etc.)`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			name := args[1]
			filePath, _ := cmd.Flags().GetString("file")

			if err := add.Run(cmd.OutOrStdout(), filePath, path, name); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}
		},
	}
	addCmd.Flags().StringP("file", "f", "openapi.yaml", "OpenAPI specification file to modify")

	rootCmd.AddCommand(lintCmd, initCmd, addCmd)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stdout)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return 2
	}

	return exitCode
}
