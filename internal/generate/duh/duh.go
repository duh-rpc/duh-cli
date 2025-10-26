package duh

import (
	"fmt"
	"io"

	"github.com/duh-rpc/duh-cli/internal/lint"
)

func Run(w io.Writer, specPath, packageName, outputDir, protoPath, protoImport, protoPackage string) error {
	spec, err := lint.Load(specPath)
	if err != nil {
		return err
	}

	result := lint.Validate(spec, specPath)
	if !result.Valid() {
		return fmt.Errorf("OpenAPI validation failed")
	}

	config, err := NewConfig(packageName, outputDir, protoPath, protoImport, protoPackage)
	if err != nil {
		return err
	}

	parser := NewParser(spec, config)
	data, err := parser.Parse()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "✓ Parsed specification successfully\n")
	_, _ = fmt.Fprintf(w, "  Module: %s\n", data.ModulePath)
	_, _ = fmt.Fprintf(w, "  Package: %s\n", data.Package)
	_, _ = fmt.Fprintf(w, "  Proto import: %s\n", data.ProtoImport)
	_, _ = fmt.Fprintf(w, "  Proto package: %s\n", data.ProtoPackage)
	_, _ = fmt.Fprintf(w, "  Timestamp: %s\n", data.Timestamp)
	_, _ = fmt.Fprintf(w, "  Operations: %d\n", len(data.Operations))

	for _, op := range data.Operations {
		_, _ = fmt.Fprintf(w, "    - %s (%s) %s -> %s\n", op.MethodName, op.ConstName, op.RequestType, op.ResponseType)
	}

	_, _ = fmt.Fprintf(w, "  List operations: %d\n", len(data.ListOps))

	for _, listOp := range data.ListOps {
		_, _ = fmt.Fprintf(w, "    - %s (iterator: %s, fetcher: %s)\n", listOp.MethodName, listOp.IteratorName, listOp.FetcherName)
	}

	return nil
}
