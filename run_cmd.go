package duh

import (
	"fmt"
	"io"

	"github.com/duh-rpc/duh-cli/internal/add"
	"github.com/duh-rpc/duh-cli/internal/generate"
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
		Use:   "lint [openapi-file]",
		Short: "Validate OpenAPI specs for DUH-RPC compliance",
		Long: `Validate OpenAPI specs for DUH-RPC compliance.

The lint command checks OpenAPI 3.0 specifications against DUH-RPC requirements
and reports any violations found.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.

Exit Codes:
  0    Validation passed (spec is DUH-RPC compliant)
  1    Validation failed (violations found)
  2    Error (file not found, parse error, etc.)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			const defaultFile = "openapi.yaml"
			filePath := defaultFile
			if len(args) > 0 {
				filePath = args[0]
			}

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

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Go code from OpenAPI specifications",
		Long: `Generate Go code from OpenAPI specifications.

The generate command uses oapi-codegen to create HTTP clients, server stubs,
and type models from DUH-RPC compliant OpenAPI specifications.

Available subcommands:
  client    Generate HTTP client code
  server    Generate server stub code
  models    Generate type models
  all       Generate all components

Use "duh generate [command] --help" for more information about a command.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	clientCmd := &cobra.Command{
		Use:   "client [openapi-file]",
		Short: "Generate HTTP client code from OpenAPI specification",
		Long: `Generate HTTP client code from OpenAPI specification.

The client command generates a Go HTTP client for calling DUH-RPC endpoints
defined in the OpenAPI specification.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.
If no output is specified, defaults to 'client.go' in the current directory.
If no package is specified, defaults to 'api'.

Exit Codes:
  0    Client generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			const defaultFile = "openapi.yaml"
			const defaultOutput = "client.go"
			const defaultPackage = "api"

			filePath := defaultFile
			if len(args) > 0 {
				filePath = args[0]
			}

			outputPath, _ := cmd.Flags().GetString("output")
			if outputPath == "" {
				outputPath = defaultOutput
			}

			packageName, _ := cmd.Flags().GetString("package")
			if packageName == "" {
				packageName = defaultPackage
			}

			if err := generate.RunClient(cmd.OutOrStdout(), filePath, outputPath, packageName); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}
		},
	}
	clientCmd.Flags().StringP("output", "o", "", "Output file path (default: client.go)")
	clientCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

	serverCmd := &cobra.Command{
		Use:   "server [openapi-file]",
		Short: "Generate server stub code from OpenAPI specification",
		Long: `Generate server stub code from OpenAPI specification.

The server command generates Go HTTP server stubs using the standard library
net/http package for implementing DUH-RPC endpoints defined in the OpenAPI
specification.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.
If no output is specified, defaults to 'server.go' in the current directory.
If no package is specified, defaults to 'api'.

Exit Codes:
  0    Server generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			const defaultFile = "openapi.yaml"
			const defaultOutput = "server.go"
			const defaultPackage = "api"

			filePath := defaultFile
			if len(args) > 0 {
				filePath = args[0]
			}

			outputPath, _ := cmd.Flags().GetString("output")
			if outputPath == "" {
				outputPath = defaultOutput
			}

			packageName, _ := cmd.Flags().GetString("package")
			if packageName == "" {
				packageName = defaultPackage
			}

			if err := generate.RunServer(cmd.OutOrStdout(), filePath, outputPath, packageName); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}
		},
	}
	serverCmd.Flags().StringP("output", "o", "", "Output file path (default: server.go)")
	serverCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

	modelsCmd := &cobra.Command{
		Use:   "models [openapi-file]",
		Short: "Generate type models from OpenAPI specification",
		Long: `Generate type models from OpenAPI specification.

The models command generates Go type definitions for request and response
schemas defined in the OpenAPI specification, without generating client
or server code.

If no file path is provided, defaults to 'openapi.yaml' in the current directory.
If no output is specified, defaults to 'models.go' in the current directory.
If no package is specified, defaults to 'api'.

Exit Codes:
  0    Models generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			const defaultFile = "openapi.yaml"
			const defaultOutput = "models.go"
			const defaultPackage = "api"

			filePath := defaultFile
			if len(args) > 0 {
				filePath = args[0]
			}

			outputPath, _ := cmd.Flags().GetString("output")
			if outputPath == "" {
				outputPath = defaultOutput
			}

			packageName, _ := cmd.Flags().GetString("package")
			if packageName == "" {
				packageName = defaultPackage
			}

			if err := generate.RunModels(cmd.OutOrStdout(), filePath, outputPath, packageName); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}
		},
	}
	modelsCmd.Flags().StringP("output", "o", "", "Output file path (default: models.go)")
	modelsCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

	allCmd := &cobra.Command{
		Use:   "all [openapi-file]",
		Short: "Generate client, server, and models from OpenAPI specification",
		Long: `Generate client, server, and models from OpenAPI specification.

The all command generates all three components (HTTP client, server stubs,
and type models) from the OpenAPI specification in a single invocation.

By default, generates client.go, server.go, and models.go in the current
directory. Use --output-dir to specify a different directory.

All generated files will use the same package name (default: api).

If no file path is provided, defaults to 'openapi.yaml' in the current directory.

Exit Codes:
  0    All components generated successfully
  2    Error (file not found, parse error, generation failed, etc.)`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			const defaultFile = "openapi.yaml"
			const defaultOutputDir = "."
			const defaultPackage = "api"

			filePath := defaultFile
			if len(args) > 0 {
				filePath = args[0]
			}

			outputDir, _ := cmd.Flags().GetString("output-dir")
			if outputDir == "" {
				outputDir = defaultOutputDir
			}

			packageName, _ := cmd.Flags().GetString("package")
			if packageName == "" {
				packageName = defaultPackage
			}

			if err := generate.RunAll(cmd.OutOrStdout(), filePath, outputDir, packageName); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
				exitCode = 2
				return
			}
		},
	}
	allCmd.Flags().String("output-dir", "", "Output directory for generated files (default: current directory)")
	allCmd.Flags().StringP("package", "p", "", "Package name for generated code (default: api)")

	generateCmd.AddCommand(clientCmd, serverCmd, modelsCmd, allCmd)

	rootCmd.AddCommand(lintCmd, initCmd, addCmd, generateCmd)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stdout)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return 2
	}

	return exitCode
}
