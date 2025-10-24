package oapi

import (
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)

// NewConfig creates and validates a codegen configuration.
func NewConfig(packageName string, opts codegen.GenerateOptions) (codegen.Configuration, error) {
	config := codegen.Configuration{
		PackageName: packageName,
		Generate:    opts,
	}

	config = config.UpdateDefaults()

	if err := config.Validate(); err != nil {
		return config, err
	}

	return config, nil
}
