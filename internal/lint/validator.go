package lint

import (
	rules2 "github.com/duh-rpc/duh-cli/internal/lint/rules"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Rule interface that all validation rules must implement
type Rule interface {
	Name() string
	Validate(doc *v3.Document) []Violation
}

// Validate runs all registered rules against the document
func Validate(doc *v3.Document, filePath string) ValidationResult {
	allRules := []Rule{
		rules2.NewPathFormatRule(),
		rules2.NewPathNoVersionPrefixRule(),
		rules2.NewServerURLVersioningRule(),
		rules2.NewHTTPMethodRule(),
		rules2.NewQueryParamsRule(),
		rules2.NewRequestBodyRule(),
		rules2.NewStatusCodeRule(),
		rules2.NewSuccessResponseRule(),
		rules2.NewContentTypeRule(),
		rules2.NewErrorResponseRule(),
		rules2.NewRPCPaginationParametersRule(),
		rules2.NewRPCPaginatedRequestStructureRule(),
		rules2.NewRPCRequestStandardNameRule(),
		rules2.NewRPCResponseStandardNameRule(),
		rules2.NewRPCRequestResponseUniqueRule(),
		rules2.NewRPCIntegerFormatRequiredRule(),
		rules2.NewRPCNoNullableRule(),
		rules2.NewRPCNoNestedArraysRule(),
		rules2.NewRPCTypedAdditionalPropertiesRule(),
		rules2.NewRPCProhibitedOneOfAndAllOfRule(),
		rules2.NewResponsePaginatedStructureRule(),
		rules2.NewPaginationParametersRule(),
		rules2.NewPathHyphenSeparatorRule(),
		rules2.NewPathPluralResourcesRule(),
		rules2.NewPathMultipleParametersRule(),
		rules2.NewSchemaNoInlineObjectsRule(),
		rules2.NewPropertyCamelCaseRule(),
		rules2.NewNullableSyntaxRule(),
		rules2.NewSchemaAdditionalPropertiesResponseRule(),
		rules2.NewNullableOptionalResponseRule(),
		rules2.NewNullableRequiredOnlyRule(),
	}

	var violations []Violation
	for _, rule := range allRules {
		ruleViolations := rule.Validate(doc)
		violations = append(violations, ruleViolations...)
	}

	return ValidationResult{
		Violations: violations,
		FilePath:   filePath,
	}
}
