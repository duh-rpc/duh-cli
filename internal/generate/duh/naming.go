package duh

import (
	"fmt"
	"regexp"
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
	versionRegex := regexp.MustCompile(`^/v\d+/(.+)$`)
	matches := versionRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", "", fmt.Errorf("invalid path format: %s", path)
	}

	subjectMethod := matches[1]
	parts := strings.Split(subjectMethod, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid path format: must contain subject.method: %s", path)
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
