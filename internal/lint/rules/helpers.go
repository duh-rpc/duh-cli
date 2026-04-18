package rules

import "strings"

func isPaginatedEndpoint(path string) bool {
	return strings.HasSuffix(path, ".list") ||
		strings.HasSuffix(path, ".search") ||
		strings.HasSuffix(path, ".query")
}
