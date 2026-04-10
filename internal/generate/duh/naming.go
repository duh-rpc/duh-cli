package duh

import (
	"fmt"
	"strings"
)

func GenerateOperationName(path string) (string, error) {
	subject, method, err := parseSubjectMethod(path)
	if err != nil {
		return "", err
	}

	subjectCamel := ToCamelCase(subject)
	methodCamel := ToCamelCase(method)

	return subjectCamel + methodCamel, nil
}

func GenerateConstName(operationName string) string {
	return "RPC" + operationName
}

func parseSubjectMethod(path string) (subject, method string, err error) {
	if !strings.HasPrefix(path, "/") {
		return "", "", fmt.Errorf("invalid path format: %s", path)
	}

	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.Split(trimmed, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid path format: must contain resource.method: %s", path)
	}

	return parts[0], parts[1], nil
}

func ToCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}

	return result.String()
}
